package xcontext

import "context"

type (
	markRetryCallKey struct{}
)

func MarkRetryCall(ctx context.Context) context.Context {
	return context.WithValue(ctx, markRetryCallKey{}, true)
}

func IsNestedCall(ctx context.Context) bool {
	if _, has := ctx.Value(markRetryCallKey{}).(bool); has {
		return true
	}

	return false
}
