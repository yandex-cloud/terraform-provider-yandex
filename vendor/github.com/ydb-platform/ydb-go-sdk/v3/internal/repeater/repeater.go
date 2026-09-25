package repeater

import (
	"context"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type Repeater interface {
	Stop()
	Force()
}

// repeater contains logic of repeating some task.
type repeater struct {
	// Interval contains an interval between task execution.
	// Interval must be greater than zero; if not, Repeater will panic.
	interval time.Duration

	name  string
	trace *trace.Driver

	// Task is a function that must be executed periodically.
	task func(context.Context) error

	cancel  context.CancelFunc
	stopped chan struct{}

	force chan struct{}
	clock clockwork.Clock
}

type option func(r *repeater)

func WithName(name string) option {
	return func(r *repeater) {
		r.name = name
	}
}

func WithTrace(trace *trace.Driver) option {
	return func(r *repeater) {
		r.trace = trace
	}
}

func WithInterval(interval time.Duration) option {
	return func(r *repeater) {
		r.interval = interval
	}
}

func WithClock(clock clockwork.Clock) option {
	return func(r *repeater) {
		r.clock = clock
	}
}

type Event = string

const (
	EventUnknown = Event("")
	EventInit    = Event("init")
	EventTick    = Event("tick")
	EventForce   = Event("force")
	EventCancel  = Event("cancel")
)

type ctxEventTypeKey struct{}

func EventType(ctx context.Context) Event {
	if eventType, ok := ctx.Value(ctxEventTypeKey{}).(Event); ok {
		return eventType
	}

	return EventUnknown
}

func WithEvent(ctx context.Context, event Event) context.Context {
	return context.WithValue(ctx,
		ctxEventTypeKey{},
		event,
	)
}

// New creates and begins to execute task periodically.
func New(
	ctx context.Context,
	interval time.Duration,
	task func(ctx context.Context) (err error),
	opts ...option,
) *repeater {
	ctx, cancel := xcontext.WithCancel(ctx)

	r := &repeater{
		interval: interval,
		task:     task,
		cancel:   cancel,
		stopped:  make(chan struct{}),
		force:    make(chan struct{}, 1),
		clock:    clockwork.NewRealClock(),
		trace:    &trace.Driver{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}

	go r.worker(ctx, r.clock.NewTicker(interval))

	return r
}

func (r *repeater) stop(onCancel func()) {
	r.cancel()
	if onCancel != nil {
		onCancel()
	}
	<-r.stopped
}

// Stop stops to execute its task.
func (r *repeater) Stop() {
	r.stop(nil)
}

func (r *repeater) Force() {
	select {
	case r.force <- struct{}{}:
	default:
	}
}

func (r *repeater) wakeUp(e Event) (err error) {
	ctx := WithEvent(context.Background(), e)

	onDone := trace.DriverOnRepeaterWakeUp(r.trace, &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/repeater.(*repeater).wakeUp"),
		r.name, e,
	)
	defer func() {
		onDone(err)

		if err != nil {
			r.Force()
		} else {
			select {
			case <-r.force:
			default:
			}
		}
	}()

	return r.task(ctx)
}

func (r *repeater) worker(ctx context.Context, tick clockwork.Ticker) {
	defer func() {
		close(r.stopped)
		tick.Stop()
	}()

	// force returns backoff with delays [500ms...32s]
	force := backoff.New(
		backoff.WithSlotDuration(500*time.Millisecond), //nolint:mnd
		backoff.WithCeiling(6),                         //nolint:mnd
		backoff.WithJitterLimit(1),
	)

	// forceIndex defines delay index for force backoff
	forceIndex := 0

	waitForceEvent := func() Event {
		if forceIndex == 0 {
			return EventForce
		}

		force := r.clock.NewTimer(force.Delay(forceIndex))
		defer force.Stop()

		select {
		case <-ctx.Done():
			return EventCancel
		case <-tick.Chan():
			return EventTick
		case <-force.Chan():
			return EventForce
		}
	}

	// processEvent func checks wakeup error and returns new force index
	processEvent := func(event Event) {
		if event == EventCancel {
			return
		}
		if err := r.wakeUp(event); err != nil {
			forceIndex++
		} else {
			forceIndex = 0
		}
	}

	for {
		select {
		case <-ctx.Done():
			return

		case <-tick.Chan():
			processEvent(EventTick)

		case <-r.force:
			processEvent(waitForceEvent())
		}
	}
}
