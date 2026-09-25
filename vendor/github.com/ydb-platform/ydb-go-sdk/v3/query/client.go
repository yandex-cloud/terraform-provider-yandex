package query

import (
	"context"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/closer"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry/budget"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type (
	// Executor is an interface for execute queries
	Executor interface {
		// Exec execute query without result
		//
		// Exec used by default:
		// - DefaultTxControl
		Exec(ctx context.Context, sql string, opts ...ExecuteOption) error

		// Query execute query with result
		//
		// Exec used by default:
		// - DefaultTxControl
		Query(ctx context.Context, sql string, opts ...ExecuteOption) (Result, error)

		// QueryResultSet execute query and take the exactly single materialized result set from result
		//
		// Exec used by default:
		// - DefaultTxControl
		QueryResultSet(ctx context.Context, sql string, opts ...ExecuteOption) (ClosableResultSet, error)

		// QueryRow execute query and take the exactly single row from exactly single result set from result
		//
		// Exec used by default:
		// - DefaultTxControl
		QueryRow(ctx context.Context, sql string, opts ...ExecuteOption) (Row, error)
	}
	// Client defines API of query client
	Client interface {
		Executor

		// Do provide the best effort for execute operation.
		//
		// Do implements internal busy loop until one of the following conditions is met:
		// - deadline was canceled or deadlined
		// - retry operation returned nil as error
		//
		// Warning: if context without deadline or cancellation func than Do can run indefinitely.
		Do(ctx context.Context, op Operation, opts ...DoOption) error

		// DoTx provide the best effort for execute transaction.
		//
		// DoTx implements internal busy loop until one of the following conditions is met:
		// - deadline was canceled or deadlined
		// - retry operation returned nil as error
		//
		// DoTx makes auto selector (with TransactionSettings, by default - SerializableReadWrite), commit and
		// rollback (on error) of transaction.
		//
		// If op TxOperation returns nil - transaction will be committed
		// If op TxOperation return non nil - transaction will be rollback
		// Warning: if context without deadline or cancellation func than DoTx can run indefinitely
		DoTx(ctx context.Context, op TxOperation, opts ...DoTxOption) error

		// Exec execute query without result
		//
		// Exec used by default:
		// - DefaultTxControl
		Exec(ctx context.Context, sql string, opts ...ExecuteOption) error

		// Query execute query with materialized result
		//
		// Warning: the large result from query will be materialized and can happened to "OOM killed" problem
		//
		// Exec used by default:
		// - DefaultTxControl
		Query(ctx context.Context, sql string, opts ...ExecuteOption) (Result, error)

		// QueryResultSet is a helper which read all rows from first result set in result
		//
		// Warning: the large result set from query will be materialized and can happened to "OOM killed" problem
		QueryResultSet(ctx context.Context, sql string, opts ...ExecuteOption) (ClosableResultSet, error)

		// QueryRow is a helper which read only one row from first result set in result
		//
		// ReadRow returns error if result contains more than one result set or more than one row
		QueryRow(ctx context.Context, sql string, opts ...ExecuteOption) (Row, error)

		// ExecuteScript starts long executing script with polling results later
		//
		// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
		ExecuteScript(
			ctx context.Context, sql string, ttl time.Duration, ops ...ExecuteOption,
		) (*options.ExecuteScriptOperation, error)

		// FetchScriptResults fetching the script results
		//
		// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
		FetchScriptResults(
			ctx context.Context, opID string, opts ...options.FetchScriptOption,
		) (*options.FetchScriptResult, error)
	}
)

func WithFetchToken(fetchToken string) options.FetchScriptOption {
	return options.WithFetchToken(fetchToken)
}

func WithResultSetIndex(resultSetIndex int64) options.FetchScriptOption {
	return options.WithResultSetIndex(resultSetIndex)
}

func WithRowsLimit(rowsLimit int64) options.FetchScriptOption {
	return options.WithRowsLimit(rowsLimit)
}

type (
	// Operation is the interface that holds an operation for retry.
	// if Operation returns not nil - operation will retry
	// if Operation returns nil - retry loop will break
	Operation func(ctx context.Context, s Session) error

	// TxOperation is the interface that holds an operation for retry.
	// if TxOperation returns not nil - operation will retry
	// if TxOperation returns nil - retry loop will break
	TxOperation func(ctx context.Context, tx TxActor) error

	ClosableSession interface {
		closer.Closer

		Session
	}
	DoOption   = options.DoOption
	DoTxOption = options.DoTxOption
)

func WithIdempotent() options.RetryOptionsOption {
	return options.WithIdempotent()
}

func WithTrace(t *trace.Query) options.TraceOption {
	return options.WithTrace(t)
}

func WithLabel(lbl string) options.LabelOption {
	return options.WithLabel(lbl)
}

// WithRetryBudget creates option with external budget
func WithRetryBudget(b budget.Budget) options.RetryOptionsOption {
	return options.WithRetryBudget(b)
}

// AllowImplicitSessions is an option to execute queries using an implicit session
// which allows the queries to be executed without explicitly creating a session.
// Please note that requests with this option use a separate session pool.
//
// Working with `query.Client.{Exec,Query,QueryResultSet,QueryRow}`.
func AllowImplicitSessions() config.Option {
	return config.AllowImplicitSessions()
}
