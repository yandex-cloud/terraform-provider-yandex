package backoff

import (
	"time"
)

type (
	delayOptions struct {
		fast Backoff
		slow Backoff
	}
	delayOption func(o *delayOptions)
)

func WithFastBackoff(fast Backoff) delayOption {
	return func(o *delayOptions) {
		o.fast = fast
	}
}

func WithSlowBackoff(slow Backoff) delayOption {
	return func(o *delayOptions) {
		o.slow = slow
	}
}

func Delay(t Type, i int, opts ...delayOption) time.Duration {
	optionsHolder := delayOptions{
		fast: Fast,
		slow: Slow,
	}
	for _, opt := range opts {
		opt(&optionsHolder)
	}
	switch t {
	case TypeFast:
		return optionsHolder.fast.Delay(i)
	case TypeSlow:
		return optionsHolder.slow.Delay(i)
	default:
		return 0
	}
}
