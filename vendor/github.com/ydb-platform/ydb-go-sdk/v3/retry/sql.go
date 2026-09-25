package retry

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/badconn"
	budget "github.com/ydb-platform/ydb-go-sdk/v3/retry/budget"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type doOptions struct {
	retryOptions []Option
}

// doTxOption defines option for redefine default Retry behavior
type doOption interface {
	ApplyDoOption(opts *doOptions)
}

var (
	_ doOption = doRetryOptionsOption(nil)
	_ doOption = labelOption("")
)

type doRetryOptionsOption []Option

func (retryOptions doRetryOptionsOption) ApplyDoOption(opts *doOptions) {
	opts.retryOptions = append(opts.retryOptions, retryOptions...)
}

// WithDoRetryOptions specified retry options
// Deprecated: use explicit options instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithDoRetryOptions(opts ...Option) doRetryOptionsOption {
	return opts
}

// Do is a retryer of database/sql conn with fallbacks on errors
func Do(ctx context.Context, db *sql.DB, op func(ctx context.Context, cc *sql.Conn) error, opts ...doOption) error {
	_, err := DoWithResult(ctx, db, func(ctx context.Context, cc *sql.Conn) (*struct{}, error) {
		err := op(ctx, cc)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return nil, nil //nolint:nilnil
	}, opts...)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

// DoWithResult is a retryer of database/sql conn with fallbacks on errors
func DoWithResult[T any](ctx context.Context, db *sql.DB,
	op func(ctx context.Context, cc *sql.Conn) (T, error),
	opts ...doOption,
) (T, error) {
	var (
		zeroValue T
		options   = doOptions{
			retryOptions: []Option{
				withCaller(stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/retry.DoWithResult")),
			},
		}
		attempts = 0
	)
	if tracer, has := db.Driver().(interface {
		TraceRetry() *trace.Retry
	}); has {
		options.retryOptions = append(options.retryOptions, nil)
		copy(options.retryOptions[1:], options.retryOptions)
		options.retryOptions[0] = WithTrace(tracer.TraceRetry())
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyDoOption(&options)
		}
	}
	v, err := RetryWithResult(ctx, func(ctx context.Context) (_ T, finalErr error) {
		attempts++
		cc, err := db.Conn(ctx)
		if err != nil {
			return zeroValue, xerrors.WithStackTrace(err)
		}
		defer func() {
			if finalErr != nil && mustDeleteConn(finalErr, cc) {
				_ = cc.Raw(func(driverConn any) error {
					return xerrors.WithStackTrace(badconn.Errorf("close connection because: %w", finalErr))
				})
			}

			_ = cc.Close()
		}()
		v, err := op(xcontext.MarkRetryCall(ctx), cc)
		if err != nil {
			return zeroValue, xerrors.WithStackTrace(err)
		}

		return v, nil
	}, options.retryOptions...)
	if err != nil {
		return zeroValue, xerrors.WithStackTrace(
			fmt.Errorf("operation failed with %d attempts: %w", attempts, err),
		)
	}

	return v, nil
}

type doTxOptions struct {
	txOptions    *sql.TxOptions
	retryOptions []Option
}

// doTxOption defines option for redefine default Retry behavior
type doTxOption interface {
	ApplyDoTxOption(o *doTxOptions)
}

var _ doTxOption = doTxRetryOptionsOption(nil)

type doTxRetryOptionsOption []Option

func (doTxRetryOptions doTxRetryOptionsOption) ApplyDoTxOption(o *doTxOptions) {
	o.retryOptions = append(o.retryOptions, doTxRetryOptions...)
}

// WithDoTxRetryOptions specified retry options
// Deprecated: use explicit options instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithDoTxRetryOptions(opts ...Option) doTxRetryOptionsOption {
	return opts
}

var _ doTxOption = txOptionsOption{}

type txOptionsOption struct {
	txOptions *sql.TxOptions
}

func (txOptions txOptionsOption) ApplyDoTxOption(o *doTxOptions) {
	o.txOptions = txOptions.txOptions
}

// WithTxOptions specified transaction options
func WithTxOptions(txOptions *sql.TxOptions) txOptionsOption {
	return txOptionsOption{
		txOptions: txOptions,
	}
}

// DoTx is a retryer of database/sql transactions with fallbacks on errors
func DoTx(ctx context.Context, db *sql.DB, op func(context.Context, *sql.Tx) error, opts ...doTxOption) error {
	_, err := DoTxWithResult(ctx, db, func(ctx context.Context, tx *sql.Tx) (*struct{}, error) {
		err := op(ctx, tx)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return nil, nil //nolint:nilnil
	}, opts...)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

// DoTxWithResult is a retryer of database/sql transactions with fallbacks on errors
func DoTxWithResult[T any](ctx context.Context, db *sql.DB,
	op func(context.Context, *sql.Tx) (T, error),
	opts ...doTxOption,
) (T, error) {
	var (
		zeroValue T
		options   = doTxOptions{
			retryOptions: []Option{
				withCaller(stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/retry.DoTxWithResult")),
			},
			txOptions: &sql.TxOptions{
				Isolation: sql.LevelDefault,
				ReadOnly:  false,
			},
		}
		attempts = 0
	)
	if d, has := db.Driver().(interface {
		TraceRetry() *trace.Retry
		RetryBudget() budget.Budget
	}); has {
		options.retryOptions = append(options.retryOptions, nil, nil)
		copy(options.retryOptions[2:], options.retryOptions)
		options.retryOptions[0] = WithTrace(d.TraceRetry())
		options.retryOptions[1] = WithBudget(d.RetryBudget())
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyDoTxOption(&options)
		}
	}
	v, err := RetryWithResult(ctx, func(ctx context.Context) (_ T, finalErr error) {
		attempts++
		tx, err := db.BeginTx(ctx, options.txOptions)
		if err != nil {
			return zeroValue, xerrors.WithStackTrace(err)
		}
		defer func() {
			_ = tx.Rollback()
		}()
		v, err := op(xcontext.MarkRetryCall(ctx), tx)
		if err != nil {
			return zeroValue, xerrors.WithStackTrace(err)
		}
		if err = tx.Commit(); err != nil {
			// We create and use tx in this method, so if we catch this error, it means context cancellation
			return zeroValue, xerrors.WithStackTrace(transformCommitError(ctx, err))
		}

		return v, nil
	}, options.retryOptions...)
	if err != nil {
		return zeroValue, xerrors.WithStackTrace(
			fmt.Errorf("tx operation failed with %d attempts: %w", attempts, err),
		)
	}

	return v, nil
}

func transformCommitError(ctx context.Context, err error) error {
	if xerrors.Is(err, sql.ErrTxDone) {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
	}

	return err
}

func mustDeleteConn[T interface {
	*sql.Conn
}](err error, conn T) bool {
	if xerrors.Is(err, driver.ErrBadConn) {
		return true
	}

	return !xerrors.IsValid(err, conn)
}
