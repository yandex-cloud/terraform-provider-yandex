package xcontext

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
)

func WithTimeout(ctx context.Context, t time.Duration) (context.Context, context.CancelFunc) {
	childCtx := &timeoutCtx{
		parentCtx: ctx,
		from:      stack.Record(1),
	}
	childCtx.ctx, childCtx.ctxCancel = context.WithTimeout(ctx, t)

	return childCtx, childCtx.cancel
}

type timeoutCtx struct {
	parentCtx context.Context //nolint:containedctx
	ctx       context.Context //nolint:containedctx
	ctxCancel context.CancelFunc
	from      string

	m   sync.Mutex
	err error
}

func (ctx *timeoutCtx) Deadline() (deadline time.Time, ok bool) {
	return ctx.ctx.Deadline()
}

func (ctx *timeoutCtx) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

func (ctx *timeoutCtx) Err() error {
	ctx.m.Lock()
	defer ctx.m.Unlock()

	if ctx.err != nil {
		return ctx.err
	}

	if errors.Is(ctx.ctx.Err(), context.DeadlineExceeded) && ctx.parentCtx.Err() == nil {
		ctx.err = errFrom(context.DeadlineExceeded, ctx.from)

		return ctx.err
	}

	if err := ctx.parentCtx.Err(); err != nil {
		ctx.err = err

		return ctx.err
	}

	return nil
}

func (ctx *timeoutCtx) Value(key interface{}) interface{} {
	return ctx.ctx.Value(key)
}

func (ctx *timeoutCtx) cancel() {
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
