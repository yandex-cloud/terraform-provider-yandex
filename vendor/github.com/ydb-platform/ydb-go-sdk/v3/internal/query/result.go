package query

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stats"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xiter"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var errReadNextResultSet = xerrors.Wrap(errors.New("ydb: stop read the result set because see part of next result set"))

var (
	_ result.Result = (*streamResult)(nil)
	_ result.Result = (*materializedResult)(nil)
)

type (
	materializedResult struct {
		resultSets []result.Set
		idx        int
	}
	streamResult struct {
		stream         Ydb_Query_V1.QueryService_ExecuteQueryClient
		closer         *ResultCloser
		lastPart       *Ydb_Query.ExecuteQueryResponsePart
		resultSetIndex int64
		trace          *trace.Query
		statsCallback  func(queryStats stats.QueryStats)
		onNextPartErr  []func(err error)
		onTxMeta       []func(txMeta *Ydb_Query.TransactionMeta)
		closeTimeout   time.Duration
	}
	resultOption func(s *streamResult)
)

func rangeResultSets(ctx context.Context, r result.Result) xiter.Seq2[result.Set, error] {
	return func(yield func(result.Set, error) bool) {
		for {
			rs, err := r.NextResultSet(ctx)
			if err != nil {
				if xerrors.Is(err, io.EOF) {
					return
				}
			}
			cont := yield(rs, err)
			if !cont || err != nil {
				return
			}
		}
	}
}

func (r *materializedResult) ResultSets(ctx context.Context) xiter.Seq2[result.Set, error] {
	return rangeResultSets(ctx, r)
}

func (r *streamResult) ResultSets(ctx context.Context) xiter.Seq2[result.Set, error] {
	return rangeResultSets(ctx, r)
}

func (r *materializedResult) Close(ctx context.Context) error {
	return nil
}

func (r *materializedResult) NextResultSet(ctx context.Context) (result.Set, error) {
	if r.idx == len(r.resultSets) {
		return nil, xerrors.WithStackTrace(io.EOF)
	}

	defer func() {
		r.idx++
	}()

	return r.resultSets[r.idx], nil
}

func withStreamResultTrace(t *trace.Query) resultOption {
	return func(s *streamResult) {
		s.trace = t
	}
}

func withStreamResultStatsCallback(callback func(queryStats stats.QueryStats)) resultOption {
	return func(s *streamResult) {
		s.statsCallback = callback
	}
}

func withStreamResultOnClose(onClose func()) resultOption {
	return func(s *streamResult) {
		s.closer.OnClose(onClose)
	}
}

func onNextPartErr(callback func(err error)) resultOption {
	return func(s *streamResult) {
		s.onNextPartErr = append(s.onNextPartErr, callback)
	}
}

func onTxMeta(callback func(txMeta *Ydb_Query.TransactionMeta)) resultOption {
	return func(s *streamResult) {
		s.onTxMeta = append(s.onTxMeta, callback)
	}
}

func withStreamResultCloseTimeout(timeout time.Duration) resultOption {
	return func(s *streamResult) {
		s.closeTimeout = timeout
	}
}

func newResult(
	ctx context.Context,
	stream Ydb_Query_V1.QueryService_ExecuteQueryClient,
	opts ...resultOption,
) (_ *streamResult, finalErr error) {
	r := streamResult{
		stream:         stream,
		closer:         NewResultCloser(),
		resultSetIndex: -1,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&r)
		}
	}

	if r.trace != nil {
		onDone := trace.QueryOnResultNew(r.trace, &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.newResult"),
		)
		defer func() {
			onDone(finalErr)
		}()
	}

	select {
	case <-ctx.Done():
		return nil, xerrors.WithStackTrace(ctx.Err())
	default:
		part, err := r.nextPart(ctx)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		r.lastPart = part

		if part.GetExecStats() != nil && r.statsCallback != nil {
			r.statsCallback(stats.FromQueryStats(part.GetExecStats()))
		}

		return &r, nil
	}
}

func (r *streamResult) nextPart(ctx context.Context) (
	part *Ydb_Query.ExecuteQueryResponsePart, err error,
) {
	if r.trace != nil {
		onDone := trace.QueryOnResultNextPart(r.trace, &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.(*streamResult).nextPart"),
		)
		defer func() {
			onDone(part.GetExecStats(), err)
		}()
	}

	select {
	case <-r.closer.Done():
		return nil, xerrors.WithStackTrace(r.closer.Err())
	case <-ctx.Done():
		return nil, xerrors.WithStackTrace(ctx.Err())
	default:
		stop := r.closer.CloseOnContextCancel(ctx)
		defer func() {
			stop()

			if err != nil {
				r.closer.Close(err)
			}

			err = r.closer.Err()
		}()

		part, err = nextPart(r.stream)
		if err != nil {
			for _, callback := range r.onNextPartErr {
				callback(err)
			}

			return nil, xerrors.WithStackTrace(err)
		}

		if txMeta := part.GetTxMeta(); txMeta != nil {
			for _, f := range r.onTxMeta {
				f(txMeta)
			}
		}

		return part, nil
	}
}

func nextPart(stream Ydb_Query_V1.QueryService_ExecuteQueryClient) (
	part *Ydb_Query.ExecuteQueryResponsePart, err error,
) {
	part, err = stream.Recv()
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return part, nil
}

func (r *streamResult) Close(ctx context.Context) (finalErr error) {
	defer func() {
		r.closer.Close(finalErr)
	}()

	if r.closeTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.closeTimeout)
		defer cancel()
	}

	if r.trace != nil {
		onDone := trace.QueryOnResultClose(r.trace, &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.(*streamResult).Close"),
		)
		defer func() {
			onDone(finalErr)
		}()
	}

	for {
		select {
		case <-ctx.Done():
			return xerrors.WithStackTrace(ctx.Err())
		case <-r.closer.Done():
			return nil
		default:
			_, err := r.nextPart(ctx)
			if err != nil {
				if xerrors.Is(err, io.EOF) {
					return nil
				}

				return xerrors.WithStackTrace(err)
			}
		}
	}
}

func (r *streamResult) nextResultSet(ctx context.Context) (_ *resultSet, err error) {
	nextResultSetIndex := r.resultSetIndex + 1
	for {
		select {
		case <-r.closer.Done():
			return nil, xerrors.WithStackTrace(r.closer.Err())
		case <-ctx.Done():
			return nil, xerrors.WithStackTrace(ctx.Err())
		default:
			if resultSetIndex := r.lastPart.GetResultSetIndex(); resultSetIndex >= nextResultSetIndex {
				r.resultSetIndex = resultSetIndex

				return newResultSet(r.nextPartFunc(ctx, nextResultSetIndex), r.lastPart), nil
			}
			if r.stream == nil {
				return nil, xerrors.WithStackTrace(io.EOF)
			}
			part, err := r.nextPart(ctx)
			if err != nil {
				return nil, xerrors.WithStackTrace(err)
			}
			if part.GetExecStats() != nil && r.statsCallback != nil {
				r.statsCallback(stats.FromQueryStats(part.GetExecStats()))
			}
			if part.GetResultSetIndex() < r.resultSetIndex {
				r.closer.Close(nil)

				if part.GetResultSetIndex() <= 0 && r.resultSetIndex > 0 {
					return nil, xerrors.WithStackTrace(io.EOF)
				}

				return nil, xerrors.WithStackTrace(fmt.Errorf(
					"next result set rowIndex %d less than last result set index %d: %w",
					part.GetResultSetIndex(), r.resultSetIndex, errWrongNextResultSetIndex,
				))
			}
			r.lastPart = part
			r.resultSetIndex = part.GetResultSetIndex()
		}
	}
}

func (r *streamResult) nextPartFunc(
	ctx context.Context,
	nextResultSetIndex int64,
) func() (_ *Ydb_Query.ExecuteQueryResponsePart, err error) {
	return func() (_ *Ydb_Query.ExecuteQueryResponsePart, err error) {
		select {
		case <-ctx.Done():
			return nil, xerrors.WithStackTrace(ctx.Err())
		case <-r.closer.Done():
			return nil, xerrors.WithStackTrace(r.closer.Err())
		default:
			if r.stream == nil {
				return nil, xerrors.WithStackTrace(io.EOF)
			}
			part, err := r.nextPart(ctx)
			if err != nil {
				return nil, xerrors.WithStackTrace(err)
			}
			r.lastPart = part
			if part.GetExecStats() != nil && r.statsCallback != nil {
				r.statsCallback(stats.FromQueryStats(part.GetExecStats()))
			}
			if part.GetResultSetIndex() > nextResultSetIndex {
				return nil, xerrors.WithStackTrace(fmt.Errorf(
					"result set (index=%d) receive part (index=%d) for next result set: %w (%w)",
					nextResultSetIndex, part.GetResultSetIndex(), io.EOF, errReadNextResultSet,
				))
			}

			return part, nil
		}
	}
}

func (r *streamResult) NextResultSet(ctx context.Context) (_ result.Set, err error) {
	if r.trace != nil {
		onDone := trace.QueryOnResultNextResultSet(r.trace, &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.(*streamResult).NextResultSet"),
		)
		defer func() {
			onDone(err)
		}()
	}

	return r.nextResultSet(ctx)
}

func exactlyOneRowFromResult(ctx context.Context, r result.Result) (row result.Row, err error) {
	rs, err := r.NextResultSet(ctx)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}
	row, err = rs.NextRow(ctx)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	_, err = rs.NextRow(ctx)
	switch {
	case err == nil:
		return nil, xerrors.WithStackTrace(errMoreThanOneRow)
	case errors.Is(err, io.EOF):
		// pass
	default:
		return nil, xerrors.WithStackTrace(err)
	}

	_, err = r.NextResultSet(ctx)
	switch {
	case err == nil:
		return nil, xerrors.WithStackTrace(errMoreThanOneRow)
	case errors.Is(err, io.EOF):
		// pass
	default:
		return nil, xerrors.WithStackTrace(err)
	}

	return row, nil
}

func exactlyOneResultSetFromResult(ctx context.Context, r result.Result) (rs result.Set, err error) {
	var rows []query.Row
	rs, err = r.NextResultSet(ctx)
	if err != nil {
		if xerrors.Is(err, io.EOF) {
			return nil, xerrors.WithStackTrace(errNoResultSets)
		}

		return nil, xerrors.WithStackTrace(err)
	}

	var row query.Row
	for {
		row, err = rs.NextRow(ctx)
		if err != nil {
			if xerrors.Is(err, io.EOF) {
				break
			}

			return nil, xerrors.WithStackTrace(err)
		}

		rows = append(rows, row)
	}

	_, err = r.NextResultSet(ctx)
	switch {
	case err == nil:
		return nil, xerrors.WithStackTrace(errMoreThanOneResultSet)
	case errors.Is(err, io.EOF):
		// pass
	default:
		return nil, xerrors.WithStackTrace(err)
	}

	return MaterializedResultSet(rs.Index(), rs.Columns(), rs.ColumnTypes(), rows), nil
}

func resultToMaterializedResult(ctx context.Context, r result.Result) (result.Result, error) {
	var resultSets []result.Set

	for {
		rs, err := r.NextResultSet(ctx)
		if err != nil {
			if xerrors.Is(err, io.EOF) {
				break
			}

			return nil, xerrors.WithStackTrace(err)
		}

		var rows []query.Row
		for {
			row, err := rs.NextRow(ctx)
			if err != nil {
				if xerrors.Is(err, io.EOF) {
					break
				}

				return nil, xerrors.WithStackTrace(err)
			}

			rows = append(rows, row)
		}

		resultSets = append(resultSets, MaterializedResultSet(rs.Index(), rs.Columns(), rs.ColumnTypes(), rows))
	}

	return &materializedResult{
		resultSets: resultSets,
	}, nil
}
