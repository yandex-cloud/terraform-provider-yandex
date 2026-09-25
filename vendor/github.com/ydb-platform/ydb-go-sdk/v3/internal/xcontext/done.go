package xcontext

import (
	"context"
	"time"
)

type doneCtx <-chan struct{}

func (done doneCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (done doneCtx) Done() <-chan struct{} {
	return done
}

func (done doneCtx) Err() error {
	select {
	case <-done:
		return context.Canceled
	default:
		return nil
	}
}

func (done doneCtx) Value(key any) any {
	return nil
}

func WithDone(parent context.Context, done <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	select {
	case <-done:
		cancel()

		return ctx, cancel
	default:
	}

	stop := context.AfterFunc(doneCtx(done), func() {
		cancel()
	})

	return ctx, func() {
		stop()
		cancel()
	}
}
