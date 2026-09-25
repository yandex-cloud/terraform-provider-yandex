package xsql

import (
	"context"
	"database/sql/driver"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1"
	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/bind"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/meta"
	internalQuery "github.com/ydb-platform/ydb-go-sdk/v3/internal/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/xquery"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/xtable"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry/budget"
	"github.com/ydb-platform/ydb-go-sdk/v3/scheme"
	"github.com/ydb-platform/ydb-go-sdk/v3/scripting"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var (
	_ io.Closer        = (*Connector)(nil)
	_ driver.Connector = (*Connector)(nil)
)

type (
	Engine    uint8
	Connector struct {
		parent      ydbDriver
		balancer    grpc.ClientConnInterface
		queryConfig *config.Config

		processor Engine

		TableOpts             []xtable.Option
		QueryOpts             []xquery.Option
		disableServerBalancer bool
		onClose               []func(*Connector)

		clock          clockwork.Clock
		idleThreshold  time.Duration
		conns          xsync.Map[uuid.UUID, *Conn]
		done           chan struct{}
		trace          *trace.DatabaseSQL
		traceRetry     *trace.Retry
		retryBudget    budget.Budget
		pathNormalizer bind.TablePathPrefix
		bindings       bind.Bindings
	}
	ydbDriver interface {
		Name() string
		Table() table.Client
		Query() query.Client
		Scripting() scripting.Client
		Scheme() scheme.Client
	}
)

func (e Engine) String() string {
	switch e {
	case TABLE:
		return "TABLE"
	case QUERY:
		return "QUERY"
	default:
		return "UNKNOWN"
	}
}

func (c *Connector) Parent() ydbDriver {
	return c.parent
}

func (c *Connector) RetryBudget() budget.Budget {
	return c.retryBudget
}

func (c *Connector) Bindings() bind.Bindings {
	return c.bindings
}

func (c *Connector) Clock() clockwork.Clock {
	return c.clock
}

func (c *Connector) Trace() *trace.DatabaseSQL {
	return c.trace
}

func (c *Connector) TraceRetry() *trace.Retry {
	return c.traceRetry
}

const (
	QUERY = iota + 1
	TABLE
)

func (c *Connector) Open(name string) (driver.Conn, error) {
	return nil, xerrors.WithStackTrace(driver.ErrSkip)
}

func (c *Connector) Connect(ctx context.Context) (_ driver.Conn, finalErr error) { //nolint:funlen
	onDone := trace.DatabaseSQLOnConnectorConnect(c.Trace(), &ctx,
		stack.FunctionID("database/sql.(*Connector).Connect", stack.Package("database/sql")),
	)

	if !c.disableServerBalancer {
		ctx = meta.WithAllowFeatures(ctx, meta.HintSessionBalancer)
	}

	switch c.processor {
	case QUERY:
		s, err := internalQuery.CreateSession(ctx, Ydb_Query_V1.NewQueryServiceClient(c.balancer), c.queryConfig)
		defer func() {
			onDone(s, finalErr)
		}()
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		id := uuid.New()

		conn := &Conn{
			processor: QUERY,
			cc: xquery.New(ctx, s, append(
				c.QueryOpts,
				xquery.WithOnClose(func() {
					c.conns.Delete(id)
				}))...,
			),
			ctx:       ctx,
			connector: c,
			lastUsage: xsync.NewLastUsage(xsync.WithClock(c.Clock())),
		}

		c.conns.Set(id, conn)

		return conn, nil

	case TABLE:
		s, err := c.parent.Table().CreateSession(ctx) //nolint:staticcheck
		defer func() {
			onDone(s, finalErr)
		}()
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		id := uuid.New()

		conn := &Conn{
			processor: TABLE,
			cc: xtable.New(ctx, c.parent.Scripting(), s, append(c.TableOpts,
				xtable.WithOnClose(func() {
					c.conns.Delete(id)
				}))...,
			),
			ctx:       ctx,
			connector: c,
			lastUsage: xsync.NewLastUsage(xsync.WithClock(c.Clock())),
		}

		c.conns.Set(id, conn)

		return conn, nil
	default:
		return nil, xerrors.WithStackTrace(errWrongQueryProcessor)
	}
}

func (c *Connector) Driver() driver.Driver {
	return c
}

func (c *Connector) Close() error {
	select {
	case <-c.done:
		return nil
	default:
		close(c.done)

		for _, onClose := range c.onClose {
			onClose(c)
		}

		return nil
	}
}

func Open(
	parent ydbDriver,
	balancer grpc.ClientConnInterface,
	queryConfig *config.Config,
	opts ...Option,
) (_ *Connector, err error) {
	c := &Connector{
		parent:      parent,
		balancer:    balancer,
		queryConfig: queryConfig,
		processor: func() Engine {
			if overQueryService, _ := strconv.ParseBool(os.Getenv("YDB_DATABASE_SQL_OVER_QUERY_SERVICE")); overQueryService {
				return QUERY
			}

			return TABLE
		}(),
		clock:          clockwork.NewRealClock(),
		done:           make(chan struct{}),
		trace:          &trace.DatabaseSQL{},
		traceRetry:     &trace.Retry{},
		pathNormalizer: bind.TablePathPrefix(parent.Name()),
	}

	for _, opt := range opts {
		if opt != nil {
			if err = opt.Apply(c); err != nil {
				return nil, err
			}
		}
	}

	if c.idleThreshold > 0 {
		ctx, cancel := xcontext.WithDone(context.Background(), c.done)
		go func() {
			defer cancel()
			for {
				idleThresholdTimer := c.clock.NewTimer(c.idleThreshold)
				select {
				case <-ctx.Done():
					idleThresholdTimer.Stop()

					return
				case <-idleThresholdTimer.Chan():
					idleThresholdTimer.Stop() // no really need, stop for common style only
					c.conns.Range(func(_ uuid.UUID, cc *Conn) bool {
						if c.clock.Since(cc.LastUsage()) > c.idleThreshold {
							_ = cc.Close()
						}

						return true
					})
				}
			}
		}()
	}

	return c, nil
}
