package operation

import (
	"context"
	"time"
)

type (
	ctxOperationTimeoutKey     struct{}
	ctxOperationCancelAfterKey struct{}
)

// WithTimeout returns a copy of parent context in which YDB operation timeout
// parameter is set to d. If parent context timeout is smaller than d, parent context is returned.
func WithTimeout(ctx context.Context, operationTimeout time.Duration) context.Context {
	if d, ok := ctxTimeout(ctx); ok && operationTimeout >= d {
		// The current cancelation timeout is already smaller than the new one.
		return ctx
	}

	return context.WithValue(ctx, ctxOperationTimeoutKey{}, operationTimeout)
}

// WithCancelAfter returns a copy of parent context in which YDB operation
// cancel after parameter is set to d. If parent context cancellation timeout is smaller
// than d, parent context is returned.
func WithCancelAfter(ctx context.Context, operationCancelAfter time.Duration) context.Context {
	if d, ok := ctxCancelAfter(ctx); ok && operationCancelAfter >= d {
		// The current cancelation timeout is already smaller than the new one.
		return ctx
	}

	return context.WithValue(ctx, ctxOperationCancelAfterKey{}, operationCancelAfter)
}

// ctxTimeout returns the timeout within given context after which
// YDB should try to cancel operation and return result regardless of the cancelation.
func ctxTimeout(ctx context.Context) (d time.Duration, ok bool) {
	d, ok = ctx.Value(ctxOperationTimeoutKey{}).(time.Duration)

	return
}

// ctxCancelAfter returns the timeout within given context after which
// YDB should try to cancel operation and return result regardless of the cancellation.
func ctxCancelAfter(ctx context.Context) (d time.Duration, ok bool) {
	d, ok = ctx.Value(ctxOperationCancelAfterKey{}).(time.Duration)

	return
}

func ctxUntilDeadline(ctx context.Context) (time.Duration, bool) {
	deadline, ok := ctx.Deadline()
	if ok {
		return time.Until(deadline), true
	}

	return 0, false
}
