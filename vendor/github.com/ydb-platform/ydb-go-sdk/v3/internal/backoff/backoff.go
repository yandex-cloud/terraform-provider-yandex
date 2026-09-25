package backoff

import (
	"math"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xrand"
)

// Backoff is the interface that contains logic of delaying operation retry.
type Backoff interface {
	// Delay returns mapping of i to Delay.
	Delay(i int) time.Duration
}

// Default parameters used by Retry() functions within different sub packages.
const (
	fastSlot = 5 * time.Millisecond
	slowSlot = 1 * time.Second
)

var (
	Fast = New(
		WithSlotDuration(fastSlot),
		WithCeiling(6), //nolint:mnd
	)
	Slow = New(
		WithSlotDuration(slowSlot),
		WithCeiling(6), //nolint:mnd
	)
)

var _ Backoff = (*logBackoff)(nil)

// logBackoff contains logarithmic Backoff policy.
type logBackoff struct {
	// slotDuration is a size of a single time slot used in Backoff Delay
	// calculation.
	// If slotDuration is less or equal to zero, then the time.Second value is
	// used.
	slotDuration time.Duration

	// ceiling is a maximum degree of Backoff Delay growth.
	// If ceiling is less or equal to zero, then the default ceiling of 1 is
	// used.
	ceiling uint

	// jitterLimit controls fixed and random portions of Backoff Delay.
	// Its value can be in range [0, 1].
	// If jitterLimit is non zero, then the Backoff Delay will be equal to (F + R),
	// where F is a result of multiplication of this value and calculated Delay
	// duration D; and R is a random sized part from [0,(D - F)].
	jitterLimit float64

	// generator of jitter
	r xrand.Rand
}

type option func(b *logBackoff)

func WithSlotDuration(slotDuration time.Duration) option {
	return func(b *logBackoff) {
		b.slotDuration = slotDuration
	}
}

func WithCeiling(ceiling uint) option {
	return func(b *logBackoff) {
		b.ceiling = ceiling
	}
}

func WithJitterLimit(jitterLimit float64) option {
	return func(b *logBackoff) {
		b.jitterLimit = jitterLimit
	}
}

func WithSeed(seed int64) option {
	return func(b *logBackoff) {
		b.r = xrand.New(xrand.WithLock(), xrand.WithSeed(seed))
	}
}

func New(opts ...option) logBackoff {
	b := logBackoff{
		r: xrand.New(xrand.WithLock()),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&b)
		}
	}

	return b
}

// Delay returns mapping of i to Delay.
func (b logBackoff) Delay(i int) time.Duration {
	s := b.slotDuration
	if s <= 0 {
		s = time.Second
	}
	n := 1 << min(uint(i), max(1, b.ceiling))
	d := s * time.Duration(n)
	f := time.Duration(math.Min(1, math.Abs(b.jitterLimit)) * float64(d))
	if f == d {
		return f
	}

	return f + time.Duration(b.r.Int64(int64(d-f)+1))
}

func min(a, b uint) uint {
	if a < b {
		return a
	}

	return b
}

func max(a, b uint) uint {
	if a > b {
		return a
	}

	return b
}
