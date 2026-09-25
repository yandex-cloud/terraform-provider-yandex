package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xrand"
)

type (
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	Budget interface {
		// Acquire will called on second and subsequent retry attempts
		Acquire(ctx context.Context) error
	}
	fixedBudget struct {
		clock  clockwork.Clock
		ticker clockwork.Ticker
		quota  chan struct{}
		done   chan struct{}
	}
	fixedBudgetOption func(q *fixedBudget)
	percentBudget     struct {
		percent int
		rand    xrand.Rand
	}
)

func withFixedBudgetClock(clock clockwork.Clock) fixedBudgetOption {
	return func(q *fixedBudget) {
		q.clock = clock
	}
}

// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func Limited(attemptsPerSecond int, opts ...fixedBudgetOption) *fixedBudget {
	q := &fixedBudget{
		clock: clockwork.NewRealClock(),
		done:  make(chan struct{}),
	}
	for _, opt := range opts {
		opt(q)
	}
	if attemptsPerSecond <= 0 {
		q.quota = make(chan struct{})
		close(q.quota)
	} else {
		q.quota = make(chan struct{}, attemptsPerSecond)
		for range make([]struct{}, attemptsPerSecond) {
			q.quota <- struct{}{}
		}
		q.ticker = q.clock.NewTicker(time.Second / time.Duration(attemptsPerSecond))
		go func() {
			defer close(q.quota)
			for {
				select {
				case <-q.ticker.Chan():
					select {
					case q.quota <- struct{}{}:
					case <-q.done:
						return
					}
				case <-q.done:
					return
				}
			}
		}()
	}

	return q
}

// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (q *fixedBudget) Stop() {
	if q.ticker != nil {
		q.ticker.Stop()
	}
	close(q.done)
}

// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (q *fixedBudget) Acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return xerrors.WithStackTrace(err)
	}
	select {
	case <-q.done:
		return xerrors.WithStackTrace(errClosedBudget)
	case <-q.quota:
		return nil
	case <-ctx.Done():
		return xerrors.WithStackTrace(ctx.Err())
	}
}

// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func Percent(percent int) *percentBudget {
	if percent > 100 || percent < 0 {
		panic(fmt.Sprintf("wrong percent value: %d", percent))
	}

	return &percentBudget{
		percent: percent,
		rand:    xrand.New(xrand.WithLock()),
	}
}

func (b *percentBudget) Acquire(ctx context.Context) error {
	if b.rand.Int(100) < b.percent { //nolint:mnd
		return nil
	}

	return ErrNoQuota
}
