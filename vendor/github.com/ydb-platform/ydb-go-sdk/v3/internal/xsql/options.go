package xsql

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/bind"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/xquery"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/xtable"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry/budget"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type (
	Option interface {
		Apply(c *Connector) error
	}
	QueryBindOption interface {
		Option
		bind.Bind
	}
	tablePathPrefixOption struct {
		bind.TablePathPrefix
	}
	processorOptionsOption struct {
		tableOpts []xtable.Option
		queryOpts []xquery.Option
	}
	traceDatabaseSQLOption struct {
		t    *trace.DatabaseSQL
		opts []trace.DatabaseSQLComposeOption
	}
	traceRetryOption struct {
		t    *trace.Retry
		opts []trace.RetryComposeOption
	}
	disableServerBalancerOption struct{}
	onCloseOption               func(*Connector)
	retryBudgetOption           struct {
		budget budget.Budget
	}
	BindOption struct {
		bind.Bind
	}
	queryProcessorOption Engine
)

func (t tablePathPrefixOption) Apply(c *Connector) error {
	c.pathNormalizer = t.TablePathPrefix
	c.bindings = append(c.bindings, t.TablePathPrefix)

	return nil
}

func (processor queryProcessorOption) Apply(c *Connector) error {
	c.processor = Engine(processor)

	return nil
}

func (opt BindOption) Apply(c *Connector) error {
	c.bindings = bind.Sort(append(c.bindings, opt.Bind))

	return nil
}

func (opt retryBudgetOption) Apply(c *Connector) error {
	c.retryBudget = opt.budget

	return nil
}

func (opt traceRetryOption) Apply(c *Connector) error {
	c.traceRetry = c.traceRetry.Compose(opt.t, opt.opts...)

	return nil
}

func (onClose onCloseOption) Apply(c *Connector) error {
	c.onClose = append(c.onClose, onClose)

	return nil
}

func (disableServerBalancerOption) Apply(c *Connector) error {
	c.disableServerBalancer = true

	return nil
}

func (opt traceDatabaseSQLOption) Apply(c *Connector) error {
	c.trace = c.trace.Compose(opt.t, opt.opts...)

	return nil
}

func (opt processorOptionsOption) Apply(c *Connector) error {
	c.QueryOpts = append(c.QueryOpts, opt.queryOpts...)
	c.TableOpts = append(c.TableOpts, opt.tableOpts...)

	return nil
}

func WithTrace(
	t *trace.DatabaseSQL, //nolint:gocritic
	opts ...trace.DatabaseSQLComposeOption,
) Option {
	return traceDatabaseSQLOption{
		t:    t,
		opts: opts,
	}
}

func WithDisableServerBalancer() Option {
	return disableServerBalancerOption{}
}

func WithOnClose(onClose func(*Connector)) Option {
	return onCloseOption(onClose)
}

func WithTraceRetry(
	t *trace.Retry,
	opts ...trace.RetryComposeOption,
) Option {
	return traceRetryOption{
		t:    t,
		opts: opts,
	}
}

func WithRetryBudget(budget budget.Budget) Option {
	return retryBudgetOption{
		budget: budget,
	}
}

func WithTablePathPrefix(tablePathPrefix string) QueryBindOption {
	return tablePathPrefixOption{
		bind.TablePathPrefix(tablePathPrefix),
	}
}

func WithQueryBind(bind bind.Bind) QueryBindOption {
	return BindOption{
		Bind: bind,
	}
}

func WithDefaultQueryMode(mode xtable.QueryMode) Option {
	return processorOptionsOption{
		tableOpts: []xtable.Option{
			xtable.WithDefaultQueryMode(mode),
		},
	}
}

func WithFakeTx(modes ...xtable.QueryMode) Option {
	return processorOptionsOption{
		tableOpts: []xtable.Option{
			xtable.WithFakeTxModes(modes...),
		},
	}
}

func WithIdleThreshold(idleThreshold time.Duration) Option {
	return processorOptionsOption{
		tableOpts: []xtable.Option{
			xtable.WithIdleThreshold(idleThreshold),
		},
	}
}

type mergedOptions []Option

func (opts mergedOptions) Apply(c *Connector) error {
	for _, opt := range opts {
		if err := opt.Apply(c); err != nil {
			return err
		}
	}

	return nil
}

func Merge(opts ...Option) Option {
	return mergedOptions(opts)
}

func WithTableOptions(opts ...xtable.Option) Option {
	return processorOptionsOption{
		tableOpts: opts,
	}
}

func WithQueryOptions(opts ...xquery.Option) Option {
	return processorOptionsOption{
		queryOpts: opts,
	}
}

func WithQueryService(b bool) Option {
	if b {
		return queryProcessorOption(QUERY)
	}

	return queryProcessorOption(TABLE)
}
