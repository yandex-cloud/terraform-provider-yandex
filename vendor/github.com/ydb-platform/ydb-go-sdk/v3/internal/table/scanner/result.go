package scanner

import (
	"context"
	"errors"
	"io"
	"sync/atomic"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_TableStats"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stats"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
)

var errAlreadyClosed = xerrors.Wrap(errors.New("result closed early"))

type baseResult struct {
	valueScanner

	nextResultSetCounter atomic.Uint64
	statsMtx             xsync.RWMutex
	stats                *Ydb_TableStats.QueryStats

	closed atomic.Bool
}

type streamResult struct {
	baseResult

	recv  func(ctx context.Context) (*Ydb.ResultSet, *Ydb_TableStats.QueryStats, error)
	close func(error) error
}

// Err returns error caused Scanner to be broken.
func (r *streamResult) Err() error {
	if err := r.valueScanner.Err(); err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

type unaryResult struct {
	baseResult

	sets    []*Ydb.ResultSet
	nextSet int
}

// Err returns error caused Scanner to be broken.
func (r *unaryResult) Err() error {
	if err := r.valueScanner.Err(); err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

// Close closes the result, preventing further iteration.
func (r *unaryResult) Close() error {
	if r.closed.CompareAndSwap(false, true) {
		return nil
	}

	return xerrors.WithStackTrace(errAlreadyClosed)
}

func (r *unaryResult) ResultSetCount() int {
	return len(r.sets)
}

func (r *baseResult) isClosed() bool {
	return r.closed.Load()
}

type resultWithError interface {
	SetErr(err error)
}

type UnaryResult interface {
	result.Result
	resultWithError
}

type StreamResult interface {
	result.StreamResult
	resultWithError
}

type option func(r *baseResult)

func WithIgnoreTruncated(ignoreTruncated bool) option {
	return func(r *baseResult) {
		r.valueScanner.ignoreTruncated = ignoreTruncated
	}
}

func WithMarkTruncatedAsRetryable() option {
	return func(r *baseResult) {
		r.valueScanner.markTruncatedAsRetryable = true
	}
}

func NewStream(
	ctx context.Context,
	recv func(ctx context.Context) (*Ydb.ResultSet, *Ydb_TableStats.QueryStats, error),
	onClose func(error) error,
	opts ...option,
) (StreamResult, error) {
	r := &streamResult{
		recv:  recv,
		close: onClose,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&r.baseResult)
		}
	}
	if err := r.nextResultSetErr(ctx); err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return r, nil
}

func NewUnary(sets []*Ydb.ResultSet, stats *Ydb_TableStats.QueryStats, opts ...option) UnaryResult {
	r := &unaryResult{
		baseResult: baseResult{
			stats: stats,
		},
		sets: sets,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&r.baseResult)
		}
	}

	return r
}

func (r *baseResult) Reset(set *Ydb.ResultSet, columnNames ...string) {
	r.reset(set)
	if set != nil {
		r.setColumnIndexes(columnNames)
	}
}

func (r *unaryResult) NextResultSetErr(ctx context.Context, columns ...string) (err error) {
	if r.isClosed() {
		return xerrors.WithStackTrace(errAlreadyClosed)
	}
	if !r.HasNextResultSet() {
		return io.EOF
	}
	r.Reset(r.sets[r.nextSet], columns...)
	r.nextSet++

	return ctx.Err()
}

func (r *unaryResult) NextResultSet(ctx context.Context, columns ...string) bool {
	return r.NextResultSetErr(ctx, columns...) == nil
}

func (r *streamResult) nextResultSetErr(ctx context.Context, columns ...string) (err error) {
	// skipping second recv because first call of recv is from New Stream(), second call is from user
	if r.nextResultSetCounter.Add(1) == 2 { //nolint:mnd
		r.setColumnIndexes(columns)

		return ctx.Err()
	}
	s, stats, err := r.recv(ctx)
	if err != nil {
		r.Reset(nil)
		if xerrors.Is(err, io.EOF) {
			return err
		}

		return r.errorf(1, "streamResult.NextResultSetErr(): %w", err)
	}
	r.Reset(s, columns...)
	if stats != nil {
		r.statsMtx.WithLock(func() {
			r.stats = stats
		})
	}

	return ctx.Err()
}

func (r *streamResult) NextResultSetErr(ctx context.Context, columns ...string) (err error) {
	if r.isClosed() {
		return xerrors.WithStackTrace(errAlreadyClosed)
	}
	if err = r.Err(); err != nil {
		return xerrors.WithStackTrace(err)
	}
	if err := r.nextResultSetErr(ctx, columns...); err != nil {
		if xerrors.Is(err, io.EOF) {
			return io.EOF
		}

		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (r *streamResult) NextResultSet(ctx context.Context, columns ...string) bool {
	return r.NextResultSetErr(ctx, columns...) == nil
}

// CurrentResultSet get current result set
func (r *baseResult) CurrentResultSet() result.Set {
	return r
}

// Stats returns query execution queryStats.
func (r *baseResult) Stats() stats.QueryStats {
	r.statsMtx.RLock()
	defer r.statsMtx.RUnlock()

	return stats.FromQueryStats(r.stats)
}

// Close closes the result, preventing further iteration.
func (r *streamResult) Close() (err error) {
	if r.closed.CompareAndSwap(false, true) {
		return r.close(r.Err())
	}

	return xerrors.WithStackTrace(errAlreadyClosed)
}

func (r *baseResult) inactive() bool {
	return r.isClosed() || r.Err() != nil
}

// HasNextResultSet reports whether result set may be advanced.
// It may be useful to call HasNextResultSet() instead of NextResultSet() to look ahead
// without advancing the result set.
// Note that it does not work with sets from stream.
func (r *streamResult) HasNextResultSet() bool {
	return !r.inactive()
}

// HasNextResultSet reports whether result set may be advanced.
// It may be useful to call HasNextResultSet() instead of NextResultSet() to look ahead
// without advancing the result set.
// Note that it does not work with sets from stream.
func (r *unaryResult) HasNextResultSet() bool {
	if r.inactive() || r.nextSet >= len(r.sets) {
		return false
	}

	return true
}
