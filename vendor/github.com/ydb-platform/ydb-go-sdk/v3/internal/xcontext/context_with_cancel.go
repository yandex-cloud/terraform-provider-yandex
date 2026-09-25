package xcontext

import (
	"context"
	"sync"
	"time"
)

func WithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	childCtx := &cancelCtx{
		parentCtx: ctx,
	}
	childCtx.ctx, childCtx.ctxCancel = context.WithCancel(ctx)

	return childCtx, childCtx.cancel
}

type cancelCtx struct {
	parentCtx context.Context //nolint:containedctx
	ctx       context.Context //nolint:containedctx
	ctxCancel context.CancelFunc

	m   sync.Mutex
	err error
}

func (ctx *cancelCtx) Deadline() (deadline time.Time, ok bool) {
	return ctx.ctx.Deadline()
}

func (ctx *cancelCtx) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

func (ctx *cancelCtx) Err() error {
	ctx.m.Lock()
	defer ctx.m.Unlock()

	if ctx.err != nil {
		return ctx.err
	}

	ctx.err = ctx.ctx.Err()

	return ctx.err
}

func (ctx *cancelCtx) Value(key interface{}) interface{} {
	return ctx.ctx.Value(key)
}

func (ctx *cancelCtx) cancel() {
	ctx.m.Lock()
	defer ctx.m.Unlock()

	ctx.ctxCancel()

	if ctx.err != nil {
		return
	}

	if err := ctx.parentCtx.Err(); err != nil {
		ctx.err = err

		return
	}
	ctx.err = errAt(context.Canceled, 1)
}
