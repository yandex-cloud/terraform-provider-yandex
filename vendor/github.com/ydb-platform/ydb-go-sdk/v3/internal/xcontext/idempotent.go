package xcontext

import "context"

type ctxIdempotentKey struct{}

func WithIdempotent(ctx context.Context, idempotent bool) context.Context {
	return context.WithValue(ctx,
		ctxIdempotentKey{},
		idempotent,
	)
}

func IsIdempotent(ctx context.Context) bool {
	if idempotent, ok := ctx.Value(ctxIdempotentKey{}).(bool); ok {
		return idempotent
	}

	return false
}
