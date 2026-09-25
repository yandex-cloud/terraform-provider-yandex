package query

import (
	"context"
	"io"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/stats"
)

type executeSettings interface {
	ExecMode() options.ExecMode
	StatsMode() options.StatsMode
	StatsCallback() func(stats stats.QueryStats)
	TxControl() options.TxControl
	Syntax() options.Syntax
	Params() params.Parameters
	CallOptions() []grpc.CallOption
	RetryOpts() []retry.Option
	ResourcePool() string
	ResponsePartLimitSizeBytes() int64
	Label() string
}

type executeScriptConfig interface {
	executeSettings

	ResultsTTL() time.Duration
	OperationParams() *Ydb_Operations.OperationParams
}

func executeQueryScriptRequest(q string, cfg executeScriptConfig) (
	*Ydb_Query.ExecuteScriptRequest,
	[]grpc.CallOption,
	error,
) {
	params, err := cfg.Params().ToYDB()
	if err != nil {
		return nil, nil, xerrors.WithStackTrace(err)
	}

	request := &Ydb_Query.ExecuteScriptRequest{
		OperationParams: &Ydb_Operations.OperationParams{
			OperationMode:    0,
			OperationTimeout: nil,
			CancelAfter:      nil,
			Labels:           nil,
			ReportCostInfo:   0,
		},
		ExecMode:      Ydb_Query.ExecMode(cfg.ExecMode()),
		ScriptContent: queryQueryContent(Ydb_Query.Syntax(cfg.Syntax()), q),
		Parameters:    params,
		StatsMode:     Ydb_Query.StatsMode(cfg.StatsMode()),
		ResultsTtl:    durationpb.New(cfg.ResultsTTL()),
		PoolId:        cfg.ResourcePool(),
	}

	return request, cfg.CallOptions(), nil
}

func executeQueryRequest(sessionID, q string, cfg executeSettings) (
	*Ydb_Query.ExecuteQueryRequest,
	[]grpc.CallOption,
	error,
) {
	params, err := cfg.Params().ToYDB()
	if err != nil {
		return nil, nil, xerrors.WithStackTrace(err)
	}

	request := &Ydb_Query.ExecuteQueryRequest{
		SessionId: sessionID,
		ExecMode:  Ydb_Query.ExecMode(cfg.ExecMode()),
		TxControl: cfg.TxControl().ToYdbQueryTransactionControl(),
		Query: &Ydb_Query.ExecuteQueryRequest_QueryContent{
			QueryContent: &Ydb_Query.QueryContent{
				Syntax: Ydb_Query.Syntax(cfg.Syntax()),
				Text:   q,
			},
		},
		Parameters:             params,
		StatsMode:              Ydb_Query.StatsMode(cfg.StatsMode()),
		ConcurrentResultSets:   false,
		PoolId:                 cfg.ResourcePool(),
		ResponsePartLimitBytes: cfg.ResponsePartLimitSizeBytes(),
	}

	return request, cfg.CallOptions(), nil
}

func queryQueryContent(syntax Ydb_Query.Syntax, q string) *Ydb_Query.QueryContent {
	return &Ydb_Query.QueryContent{
		Syntax: syntax,
		Text:   q,
	}
}

func execute(
	ctx context.Context, sessionID string, c Ydb_Query_V1.QueryServiceClient,
	q string, settings executeSettings, opts ...resultOption,
) (
	_ *streamResult, finalErr error,
) {
	request, callOptions, err := executeQueryRequest(sessionID, q, settings)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	executeCtx, executeCancel := xcontext.WithCancel(xcontext.ValueOnly(ctx))

	stop := context.AfterFunc(ctx, executeCancel)
	defer stop()

	defer func() {
		if finalErr != nil {
			executeCancel()
		}
	}()

	stream, err := c.ExecuteQuery(executeCtx, request, callOptions...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	r, err := newResult(ctx, stream, append(opts,
		withStreamResultStatsCallback(settings.StatsCallback()),
		withStreamResultOnClose(executeCancel),
	)...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return r, nil
}

func readAll(ctx context.Context, r *streamResult) error {
	defer func() {
		_ = r.Close(ctx)
	}()

	for {
		_, err := r.nextResultSet(ctx)
		if err != nil {
			if xerrors.Is(err, io.EOF) {
				return nil
			}

			return xerrors.WithStackTrace(err)
		}
	}
}

func readResultSet(ctx context.Context, r *streamResult) (_ *resultSetWithClose, finalErr error) {
	rs, err := r.nextResultSet(ctx)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}
	rs.mustBeLastResultSet = true

	return &resultSetWithClose{
		resultSet: rs,
		close:     r.Close,
	}, nil
}

func readMaterializedResultSet(ctx context.Context, r *streamResult) (
	_ *materializedResultSet, rowsCount int, finalErr error,
) {
	defer func() {
		_ = r.Close(ctx)
	}()

	rs, err := r.nextResultSet(ctx)
	if err != nil {
		return nil, 0, xerrors.WithStackTrace(err)
	}

	var rows []query.Row
	for {
		row, err := rs.nextRow(ctx)
		if err != nil {
			if xerrors.Is(err, io.EOF) {
				break
			}

			return nil, 0, xerrors.WithStackTrace(err)
		}

		rows = append(rows, row)
	}

	_, err = r.nextResultSet(ctx)
	if err == nil {
		return nil, 0, xerrors.WithStackTrace(errMoreThanOneResultSet)
	}
	if !xerrors.Is(err, io.EOF) {
		return nil, 0, xerrors.WithStackTrace(err)
	}

	return MaterializedResultSet(rs.Index(), rs.Columns(), rs.ColumnTypes(), rows), len(rows), nil
}

func readRow(ctx context.Context, r *streamResult) (_ *Row, finalErr error) {
	defer func() {
		_ = r.Close(ctx)
	}()

	rs, err := r.nextResultSet(ctx)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	row, err := rs.nextRow(ctx)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	_, err = rs.nextRow(ctx)
	if err == nil {
		return nil, xerrors.WithStackTrace(errMoreThanOneRow)
	}
	if !xerrors.Is(err, io.EOF) {
		return nil, xerrors.WithStackTrace(err)
	}

	_, err = r.NextResultSet(ctx)
	if err == nil {
		return nil, xerrors.WithStackTrace(errMoreThanOneResultSet)
	}
	if !xerrors.Is(err, io.EOF) {
		return nil, xerrors.WithStackTrace(err)
	}

	return row, nil
}
