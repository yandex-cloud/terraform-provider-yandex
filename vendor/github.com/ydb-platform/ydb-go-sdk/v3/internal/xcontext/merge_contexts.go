package xcontext

import (
	"context"
	"time"
)

var _ context.Context = (*MergedContexts)(nil)

type MergedContexts struct {
	additionalValues context.Context //nolint:containedctx
	deadlineContext  context.Context //nolint:containedctx
}

func (ctx *MergedContexts) Deadline() (deadline time.Time, ok bool) {
	return ctx.deadlineContext.Deadline()
}

func (ctx *MergedContexts) Done() <-chan struct{} {
	return ctx.deadlineContext.Done()
}

func (ctx *MergedContexts) Err() error {
	return ctx.deadlineContext.Err()
}

func (ctx *MergedContexts) Value(key interface{}) interface{} {
	if ctx.deadlineContext.Value(key) != nil {
		return ctx.deadlineContext.Value(key)
	}

	return ctx.additionalValues.Value(key)
}

// MergeContexts returns a context with the deadline from the deadline context and values from both contexts
func MergeContexts(deadline context.Context, additionalValues context.Context) context.Context {
	return &MergedContexts{
		additionalValues: additionalValues,
		deadlineContext:  deadline,
	}
}
