package xcontext

import (
	"context"
)

// ValueOnly helps to clear parent context from deadlines/cancels
func ValueOnly(ctx context.Context) context.Context {
	return context.WithoutCancel(ctx)
}
