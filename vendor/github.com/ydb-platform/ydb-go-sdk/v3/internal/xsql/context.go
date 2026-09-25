package xsql

import "context"

type ctxExplainQueryModeKey struct{}

func WithExplain(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxExplainQueryModeKey{}, true)
}

func isExplain(ctx context.Context) bool {
	v, has := ctx.Value(ctxExplainQueryModeKey{}).(bool)

	return has && v
}
