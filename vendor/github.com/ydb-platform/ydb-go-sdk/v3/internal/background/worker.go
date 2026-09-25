package background

import (
	"context"
	"errors"
	"fmt"
	"runtime/pprof"
	"strings"
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/empty"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
)

var (
	ErrAlreadyClosed       = xerrors.Wrap(errors.New("ydb: background worker already closed"))
	errClosedWithNilReason = xerrors.Wrap(errors.New("ydb: background worker closed with nil reason"))
)

// A Worker must not be copied after first use
type Worker struct {
	ctx            context.Context //nolint:containedctx
	name           string
	workers        sync.WaitGroup
	closeReason    error
	tasksCompleted empty.Chan
	tasks          chan backgroundTask
	stop           context.CancelFunc
	onceInit       sync.Once
	m              xsync.Mutex
	closed         bool
}

type CallbackFunc func(ctx context.Context)

func NewWorker(parent context.Context, name string) *Worker {
	w := Worker{
		name: name,
	}
	w.ctx, w.stop = xcontext.WithCancel(parent)

	return &w
}

func (b *Worker) Context() context.Context {
	b.init()

	return b.ctx
}

func (b *Worker) Start(name string, f CallbackFunc) {
	b.init()

	b.m.WithLock(func() {
		if b.closed {
			return
		}

		b.tasks <- backgroundTask{
			callback: f,
			name:     name,
		}
	})
}

func (b *Worker) Done() <-chan struct{} {
	b.init()

	return b.ctx.Done()
}

func (b *Worker) Close(ctx context.Context, err error) error {
	b.init()

	var resErr error
	b.m.WithLock(func() {
		if b.closed {
			// The error of Close is second close, close reason added for describe previous close only, for better debug
			//nolint:errorlint
			resErr = xerrors.WithStackTrace(fmt.Errorf("%w with reason: %+v", ErrAlreadyClosed, b.closeReason))

			return
		}

		b.closed = true

		close(b.tasks)
		b.closeReason = err
		if b.closeReason == nil {
			b.closeReason = errClosedWithNilReason
		}

		b.stop()
	})
	if resErr != nil {
		return resErr
	}

	<-b.tasksCompleted

	bgCompleted := make(empty.Chan)

	go func() {
		b.workers.Wait()
		close(bgCompleted)
	}()

	select {
	case <-bgCompleted:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *Worker) CloseReason() error {
	b.m.Lock()
	defer b.m.Unlock()

	return b.closeReason
}

func (b *Worker) StopDone() <-chan empty.Struct {
	return b.tasksCompleted
}

func (b *Worker) init() {
	b.onceInit.Do(func() {
		if b.ctx == nil {
			b.ctx, b.stop = xcontext.WithCancel(context.Background())
		}
		b.tasks = make(chan backgroundTask)
		b.tasksCompleted = make(empty.Chan)

		pprof.Do(b.ctx, pprof.Labels("worker-name", b.name), func(ctx context.Context) {
			go b.starterLoop(ctx)
		})
	})
}

func (b *Worker) starterLoop(ctx context.Context) {
	defer close(b.tasksCompleted)

	for bgTask := range b.tasks {
		b.workers.Add(1)

		go func(task backgroundTask) {
			defer b.workers.Done()

			safeLabel := strings.ReplaceAll(task.name, `"`, `'`)

			pprof.Do(ctx, pprof.Labels("background", safeLabel), task.callback)
		}(bgTask)
	}
}

type backgroundTask struct {
	callback CallbackFunc
	name     string
}
