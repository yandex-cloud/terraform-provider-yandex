package table

import (
	"context"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Table"
	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type statement struct {
	session *Session
	query   Query
	params  map[string]*Ydb.Type
}

// Execute executes prepared data query.
func (s *statement) Execute(
	ctx context.Context, txControl *table.TransactionControl,
	parameters *params.Params,
	opts ...options.ExecuteDataQueryOption,
) (txr table.Transaction, r result.Result, err error) {
	var (
		request = options.ExecuteDataQueryDesc{
			ExecuteDataQueryRequest: &Ydb_Table.ExecuteDataQueryRequest{
				SessionId: s.session.id,
				TxControl: txControl.ToYdbTableTransactionControl(),
				Parameters: func() map[string]*Ydb.TypedValue {
					p, _ := parameters.ToYDB()

					return p
				}(),
				Query: s.query.toYDB(),
			},
			IgnoreTruncated: s.session.config.IgnoreTruncated(),
		}
		callOptions []grpc.CallOption
	)

	request.QueryCachePolicy = &Ydb_Table.QueryCachePolicy{
		KeepInCache: len(request.Parameters) > 0,
	}

	request.OperationParams = operation.Params(ctx,
		s.session.config.OperationTimeout(),
		s.session.config.OperationCancelAfter(),
		operation.ModeSync,
	)

	for _, opt := range opts {
		if opt != nil {
			callOptions = append(callOptions, opt.ApplyExecuteDataQueryOption(&request)...)
		}
	}

	onDone := trace.TableOnSessionQueryExecute(
		s.session.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*statement).Execute"),
		s.session, s.query, parameters,
		request.QueryCachePolicy.GetKeepInCache(),
	)
	defer func() {
		onDone(txr, true, r, err)
	}()

	return s.execute(ctx, txControl, &request, callOptions...)
}

// execute executes prepared query without any tracing.
func (s *statement) execute(
	ctx context.Context,
	txControl *tx.Control,
	request *options.ExecuteDataQueryDesc,
	callOptions ...grpc.CallOption,
) (
	txr table.Transaction, r result.Result, err error,
) {
	t, r, err := s.session.dataQuery.execute(ctx, txControl, request.ExecuteDataQueryRequest, callOptions...)
	if err != nil {
		return nil, nil, xerrors.WithStackTrace(err)
	}

	if t != nil {
		t.s = s.session
	}

	return t, r, nil
}

func (s *statement) NumInput() int {
	return len(s.params)
}

func (s *statement) Text() string {
	return s.query.YQL()
}
