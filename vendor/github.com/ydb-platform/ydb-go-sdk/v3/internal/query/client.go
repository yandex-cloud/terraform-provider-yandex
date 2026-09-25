package query

import (
	"context"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/closer"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/meta"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/pool"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

//go:generate mockgen -destination grpc_client_mock_test.go --typed -package query -write_package_comment=false github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1 QueryServiceClient,QueryService_AttachSessionClient,QueryService_ExecuteQueryClient

var (
	_ query.Client = (*Client)(nil)
	_ sessionPool  = (*pool.Pool[*Session, Session])(nil)
)

type (
	sessionPool interface {
		closer.Closer

		Stats() pool.Stats
		With(ctx context.Context, f func(ctx context.Context, s *Session) error, opts ...retry.Option) error
	}
	Client struct {
		config              *config.Config
		client              Ydb_Query_V1.QueryServiceClient
		explicitSessionPool sessionPool

		// implicitSessionPool is a pool of implicit sessions,
		// i.e. fake sessions created without CreateSession/AttachSession requests.
		implicitSessionPool sessionPool

		done chan struct{}
	}
)

func (c *Client) Config() *config.Config {
	return c.config
}

func fetchScriptResults(ctx context.Context,
	client Ydb_Query_V1.QueryServiceClient,
	opID string, opts ...options.FetchScriptOption,
) (*options.FetchScriptResult, error) {
	r, err := retry.RetryWithResult(ctx, func(ctx context.Context) (*options.FetchScriptResult, error) {
		request := &options.FetchScriptResultsRequest{
			FetchScriptResultsRequest: Ydb_Query.FetchScriptResultsRequest{
				OperationId: opID,
			},
		}
		for _, opt := range opts {
			if opt != nil {
				opt(request)
			}
		}

		response, err := client.FetchScriptResults(ctx, &request.FetchScriptResultsRequest)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		rs := response.GetResultSet()
		columns := rs.GetColumns()
		columnNames := make([]string, len(columns))
		columnTypes := make([]types.Type, len(columns))
		for i := range columns {
			columnNames[i] = columns[i].GetName()
			columnTypes[i] = types.TypeFromYDB(columns[i].GetType())
		}
		rows := make([]query.Row, len(rs.GetRows()))
		for i, r := range rs.GetRows() {
			rows[i] = NewRow(columns, r)
		}

		return &options.FetchScriptResult{
			ResultSetIndex: response.GetResultSetIndex(),
			ResultSet:      MaterializedResultSet(int(response.GetResultSetIndex()), columnNames, columnTypes, rows),
			NextToken:      response.GetNextFetchToken(),
		}, nil
	}, retry.WithIdempotent(true))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return r, nil
}

func (c *Client) FetchScriptResults(ctx context.Context,
	opID string, opts ...options.FetchScriptOption,
) (*options.FetchScriptResult, error) {
	r, err := retry.RetryWithResult(ctx, func(ctx context.Context) (*options.FetchScriptResult, error) {
		r, err := fetchScriptResults(ctx, c.client, opID,
			append(opts, func(request *options.FetchScriptResultsRequest) {
				request.Trace = c.config.Trace()
			})...,
		)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return r, nil
	}, retry.WithIdempotent(true))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return r, nil
}

type executeScriptSettings struct {
	executeSettings
	ttl             time.Duration
	operationParams *Ydb_Operations.OperationParams
}

func (s *executeScriptSettings) OperationParams() *Ydb_Operations.OperationParams {
	return s.operationParams
}

func (s *executeScriptSettings) ResultsTTL() time.Duration {
	return s.ttl
}

func executeScript(ctx context.Context,
	client Ydb_Query_V1.QueryServiceClient, request *Ydb_Query.ExecuteScriptRequest, grpcOpts ...grpc.CallOption,
) (*options.ExecuteScriptOperation, error) {
	op, err := retry.RetryWithResult(ctx, func(ctx context.Context) (*options.ExecuteScriptOperation, error) {
		response, err := client.ExecuteScript(ctx, request, grpcOpts...)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return &options.ExecuteScriptOperation{
			ID:            response.GetId(),
			ConsumedUnits: response.GetCostInfo().GetConsumedUnits(),
			Metadata:      options.ToMetadataExecuteQuery(response.GetMetadata()),
		}, nil
	}, retry.WithIdempotent(true))
	if err != nil {
		return op, xerrors.WithStackTrace(err)
	}

	return op, nil
}

func (c *Client) ExecuteScript(
	ctx context.Context, q string, ttl time.Duration, opts ...options.Execute,
) (
	op *options.ExecuteScriptOperation, err error,
) {
	settings := &executeScriptSettings{
		executeSettings: options.ExecuteSettings(opts...),
		ttl:             ttl,
		operationParams: operation.Params(
			ctx,
			c.config.OperationTimeout(),
			c.config.OperationCancelAfter(),
			operation.ModeSync,
		),
	}

	request, grpcOpts, err := executeQueryScriptRequest(q, settings)
	if err != nil {
		return op, xerrors.WithStackTrace(err)
	}

	op, err = executeScript(ctx, c.client, request, grpcOpts...)
	if err != nil {
		return op, xerrors.WithStackTrace(err)
	}

	return op, nil
}

func (c *Client) Close(ctx context.Context) error {
	if c == nil {
		return xerrors.WithStackTrace(errNilClient)
	}

	close(c.done)

	if err := c.explicitSessionPool.Close(ctx); err != nil {
		return xerrors.WithStackTrace(err)
	}

	if err := c.implicitSessionPool.Close(ctx); err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func do(
	ctx context.Context,
	pool sessionPool,
	op func(ctx context.Context, s *Session) error,
	opts ...retry.Option,
) (finalErr error) {
	err := pool.With(ctx, func(ctx context.Context, s *Session) error {
		s.SetStatus(StatusInUse)

		err := op(ctx, s)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		s.SetStatus(StatusIdle)

		return nil
	}, opts...)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (c *Client) Do(ctx context.Context, op query.Operation, opts ...options.DoOption) (finalErr error) {
	ctx, cancel := xcontext.WithDone(ctx, c.done)
	defer cancel()

	var (
		settings = options.ParseDoOpts(c.config.Trace(), opts...)
		onDone   = trace.QueryOnDo(settings.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.(*Client).Do"),
			settings.Label(),
		)
		attempts = 0
	)
	defer func() {
		onDone(attempts, finalErr)
	}()

	err := do(ctx, c.explicitSessionPool,
		func(ctx context.Context, s *Session) error {
			return op(ctx, s)
		},
		append([]retry.Option{
			retry.WithTrace(&trace.Retry{
				OnRetry: func(info trace.RetryLoopStartInfo) func(trace.RetryLoopDoneInfo) {
					return func(info trace.RetryLoopDoneInfo) {
						attempts = info.Attempts
					}
				},
			}),
		}, settings.RetryOpts()...)...,
	)

	return err
}

func doTx(
	ctx context.Context,
	pool sessionPool,
	op query.TxOperation,
	txSettings tx.Settings,
	opts ...retry.Option,
) (finalErr error) {
	err := do(ctx, pool, func(ctx context.Context, s *Session) (opErr error) {
		tx, err := s.Begin(ctx, txSettings)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		defer func() {
			_ = tx.Rollback(ctx)
		}()

		err = op(ctx, tx)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		err = tx.CommitTx(ctx)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		return nil
	}, opts...)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func clientQueryRow(
	ctx context.Context, pool sessionPool, q string, settings executeSettings, resultOpts ...resultOption,
) (row query.Row, finalErr error) {
	err := do(ctx, pool, func(ctx context.Context, s *Session) (err error) {
		row, err = s.queryRow(ctx, q, settings, resultOpts...)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		return nil
	}, settings.RetryOpts()...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return row, nil
}

// QueryRow is a helper which read only one row from first result set in result
func (c *Client) QueryRow(ctx context.Context, q string, opts ...options.Execute) (_ query.Row, finalErr error) {
	ctx, cancel := xcontext.WithDone(ctx, c.done)
	defer cancel()

	settings := options.ExecuteSettings(opts...)

	onDone := trace.QueryOnQueryRow(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.(*Client).QueryRow"),
		q, settings.Label(),
	)
	defer func() {
		onDone(finalErr)
	}()

	row, err := clientQueryRow(ctx, c.pool(), q, settings, withStreamResultTrace(c.config.Trace()))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return row, nil
}

func clientExec(ctx context.Context, pool sessionPool, q string, opts ...options.Execute) (finalErr error) {
	settings := options.ExecuteSettings(opts...)
	err := do(ctx, pool, func(ctx context.Context, s *Session) (err error) {
		streamResult, err := s.execute(ctx, q, settings, withStreamResultTrace(s.trace))
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		defer func() {
			_ = streamResult.Close(ctx)
		}()

		err = readAll(ctx, streamResult)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		return nil
	}, settings.RetryOpts()...)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (c *Client) Exec(ctx context.Context, q string, opts ...options.Execute) (finalErr error) {
	ctx, cancel := xcontext.WithDone(ctx, c.done)
	defer cancel()

	settings := options.ExecuteSettings(opts...)
	onDone := trace.QueryOnExec(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.(*Client).Exec"),
		q,
		settings.Label(),
	)
	defer func() {
		onDone(finalErr)
	}()

	err := clientExec(ctx, c.pool(), q, opts...)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func clientQuery(ctx context.Context, pool sessionPool, q string, opts ...options.Execute) (
	r query.Result, err error,
) {
	settings := options.ExecuteSettings(opts...)
	err = do(ctx, pool, func(ctx context.Context, s *Session) (err error) {
		streamResult, err := s.execute(ctx, q, options.ExecuteSettings(opts...), withStreamResultTrace(s.trace))
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		defer func() {
			_ = streamResult.Close(ctx)
		}()

		r, err = resultToMaterializedResult(ctx, streamResult)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		return nil
	}, settings.RetryOpts()...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return r, nil
}

func (c *Client) Query(ctx context.Context, q string, opts ...options.Execute) (r query.Result, err error) {
	ctx, cancel := xcontext.WithDone(ctx, c.done)
	defer cancel()

	settings := options.ExecuteSettings(opts...)
	onDone := trace.QueryOnQuery(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.(*Client).Query"),
		q, settings.Label(),
	)
	defer func() {
		onDone(err)
	}()

	r, err = clientQuery(ctx, c.pool(), q, opts...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return r, nil
}

func clientQueryResultSet(
	ctx context.Context, pool sessionPool, q string, settings executeSettings, resultOpts ...resultOption,
) (rs result.ClosableResultSet, rowsCount int, finalErr error) {
	err := do(ctx, pool, func(ctx context.Context, s *Session) error {
		streamResult, err := s.execute(ctx, q, settings, resultOpts...)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		defer func() {
			_ = streamResult.Close(ctx)
		}()

		rs, rowsCount, err = readMaterializedResultSet(ctx, streamResult)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		return nil
	}, settings.RetryOpts()...)
	if err != nil {
		return nil, 0, xerrors.WithStackTrace(err)
	}

	return rs, rowsCount, nil
}

// QueryResultSet is a helper which read all rows from first result set in result
func (c *Client) QueryResultSet(
	ctx context.Context, q string, opts ...options.Execute,
) (rs result.ClosableResultSet, finalErr error) {
	ctx, cancel := xcontext.WithDone(ctx, c.done)
	defer cancel()

	var (
		settings  = options.ExecuteSettings(opts...)
		rowsCount int
		err       error
	)

	onDone := trace.QueryOnQueryResultSet(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.(*Client).QueryResultSet"),
		q, settings.Label(),
	)
	defer func() {
		onDone(finalErr, rowsCount)
	}()

	rs, rowsCount, err = clientQueryResultSet(ctx, c.pool(), q, settings, withStreamResultTrace(c.config.Trace()))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return rs, nil
}

// pool returns the appropriate session pool based on the client configuration.
// If implicit sessions are enabled, it returns the implicit session pool;
// otherwise, it returns the explicit session pool.
func (c *Client) pool() sessionPool {
	if c.config.AllowImplicitSessions() {
		return c.implicitSessionPool
	}

	return c.explicitSessionPool
}

func (c *Client) DoTx(ctx context.Context, op query.TxOperation, opts ...options.DoTxOption) (finalErr error) {
	ctx, cancel := xcontext.WithDone(ctx, c.done)
	defer cancel()

	var (
		settings = options.ParseDoTxOpts(c.config.Trace(), opts...)
		onDone   = trace.QueryOnDoTx(settings.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.(*Client).DoTx"),
			settings.Label(),
		)
		attempts = 0
	)
	defer func() {
		onDone(attempts, finalErr)
	}()

	err := doTx(ctx, c.explicitSessionPool, op,
		settings.TxSettings(),
		append(
			[]retry.Option{
				retry.WithTrace(&trace.Retry{
					OnRetry: func(info trace.RetryLoopStartInfo) func(trace.RetryLoopDoneInfo) {
						return func(info trace.RetryLoopDoneInfo) {
							attempts = info.Attempts
						}
					},
				}),
			},
			settings.RetryOpts()...,
		)...,
	)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func CreateSession(ctx context.Context, client Ydb_Query_V1.QueryServiceClient, cfg *config.Config) (*Session, error) {
	s, err := retry.RetryWithResult(ctx, func(ctx context.Context) (*Session, error) {
		var (
			createCtx    context.Context
			cancelCreate context.CancelFunc
		)
		if d := cfg.SessionCreateTimeout(); d > 0 {
			createCtx, cancelCreate = xcontext.WithTimeout(ctx, d)
		} else {
			createCtx, cancelCreate = xcontext.WithCancel(ctx)
		}
		defer cancelCreate()

		s, err := createSession(createCtx, client,
			WithDeleteTimeout(cfg.SessionDeleteTimeout()),
			WithTrace(cfg.Trace()),
		)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		s.lazyTx = cfg.LazyTx()

		return s, nil
	})
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return s, nil
}

func New(ctx context.Context, cc grpc.ClientConnInterface, cfg *config.Config) *Client {
	onDone := trace.QueryOnNew(cfg.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/query.New"),
	)
	defer onDone()

	client := Ydb_Query_V1.NewQueryServiceClient(cc)

	return newWithQueryServiceClient(ctx, client, cc, cfg)
}

func newWithQueryServiceClient(ctx context.Context,
	client Ydb_Query_V1.QueryServiceClient,
	cc grpc.ClientConnInterface,
	cfg *config.Config,
) *Client {
	return &Client{
		config:              cfg,
		client:              client,
		done:                make(chan struct{}),
		implicitSessionPool: createImplicitSessionPool(ctx, cfg, client, cc),
		explicitSessionPool: pool.New(ctx,
			pool.WithLimit[*Session](cfg.PoolLimit()),
			pool.WithItemUsageLimit[*Session](cfg.PoolSessionUsageLimit()),
			pool.WithItemUsageTTL[*Session](cfg.PoolSessionUsageTTL()),
			pool.WithTrace[*Session](poolTrace(cfg.Trace())),
			pool.WithCreateItemTimeout[*Session](cfg.SessionCreateTimeout()),
			pool.WithCloseItemTimeout[*Session](cfg.SessionDeleteTimeout()),
			pool.WithMustDeleteItemFunc(func(s *Session, err error) bool {
				if !s.IsAlive() {
					return true
				}

				return err != nil && xerrors.MustDeleteTableOrQuerySession(err)
			}),
			pool.WithIdleTimeToLive[*Session](cfg.SessionIdleTimeToLive()),
			pool.WithCreateItemFunc(func(ctx context.Context) (_ *Session, err error) {
				var (
					createCtx    context.Context
					cancelCreate context.CancelFunc
				)
				if d := cfg.SessionCreateTimeout(); d > 0 {
					createCtx, cancelCreate = xcontext.WithTimeout(ctx, d)
				} else {
					createCtx, cancelCreate = xcontext.WithCancel(ctx)
				}
				defer cancelCreate()

				if !cfg.DisableSessionBalancer() {
					createCtx = meta.WithAllowFeatures(createCtx, meta.HintSessionBalancer)
				}

				s, err := createSession(createCtx, client,
					WithConn(cc),
					WithDeleteTimeout(cfg.SessionDeleteTimeout()),
					WithTrace(cfg.Trace()),
				)
				if err != nil {
					return nil, xerrors.WithStackTrace(err)
				}

				s.lazyTx = cfg.LazyTx()

				return s, nil
			}),
		),
	}
}

func poolTrace(t *trace.Query) *pool.Trace {
	return &pool.Trace{
		OnNew: func(ctx *context.Context, call stack.Caller) func(limit int) {
			onDone := trace.QueryOnPoolNew(t, ctx, call)

			return func(limit int) {
				onDone(limit)
			}
		},
		OnClose: func(ctx *context.Context, call stack.Caller) func(err error) {
			onDone := trace.QueryOnClose(t, ctx, call)

			return func(err error) {
				onDone(err)
			}
		},
		OnTry: func(ctx *context.Context, call stack.Caller) func(err error) {
			onDone := trace.QueryOnPoolTry(t, ctx, call)

			return func(err error) {
				onDone(err)
			}
		},
		OnWith: func(ctx *context.Context, call stack.Caller) func(attempts int, err error) {
			onDone := trace.QueryOnPoolWith(t, ctx, call)

			return func(attempts int, err error) {
				onDone(attempts, err)
			}
		},
		OnPut: func(ctx *context.Context, call stack.Caller, item any) func(err error) {
			onDone := trace.QueryOnPoolPut(t, ctx, call, item.(*Session)) //nolint:forcetypeassert

			return func(err error) {
				onDone(err)
			}
		},
		OnGet: func(ctx *context.Context, call stack.Caller) func(item any, attempts int, err error) {
			onDone := trace.QueryOnPoolGet(t, ctx, call)

			return func(item any, attempts int, err error) {
				onDone(item.(*Session), attempts, err) //nolint:forcetypeassert
			}
		},
		OnChange: func(stats pool.Stats) {
			trace.QueryOnPoolChange(t, stats.Limit, stats.Index, stats.Idle, stats.Wait, stats.CreateInProgress)
		},
	}
}

func createImplicitSessionPool(ctx context.Context,
	cfg *config.Config,
	c Ydb_Query_V1.QueryServiceClient,
	cc grpc.ClientConnInterface,
) sessionPool {
	return pool.New(ctx,
		pool.WithLimit[*Session](cfg.PoolLimit()),
		pool.WithTrace[*Session](poolTrace(cfg.Trace())),
		pool.WithCreateItemFunc(func(ctx context.Context) (_ *Session, err error) {
			core := &sessionCore{
				cc:     cc,
				Client: c,
				Trace:  cfg.Trace(),
				done:   make(chan struct{}),
			}

			return &Session{
				Core:   core,
				trace:  cfg.Trace(),
				client: c,
			}, nil
		}),
	)
}
