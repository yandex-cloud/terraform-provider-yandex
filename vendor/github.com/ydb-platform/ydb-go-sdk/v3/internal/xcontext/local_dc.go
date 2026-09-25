package xcontext

import (
	"context"
	"fmt"
)

type localDcKey struct{}

func WithLocalDC(ctx context.Context, dc string) context.Context {
	return context.WithValue(ctx, localDcKey{}, dc)
}

func ExtractLocalDC(ctx context.Context) string {
	if val := ctx.Value(localDcKey{}); val != nil {
		s, ok := val.(string)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to string", s))
		}

		return s
	}

	return ""
}
