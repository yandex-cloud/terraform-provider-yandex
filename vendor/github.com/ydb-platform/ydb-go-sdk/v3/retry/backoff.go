package retry

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
)

// Backoff makes backoff object with custom params
func Backoff(slotDuration time.Duration, ceiling uint, jitterLimit float64) backoff.Backoff {
	return backoff.New(
		backoff.WithSlotDuration(slotDuration),
		backoff.WithCeiling(ceiling),
		backoff.WithJitterLimit(jitterLimit),
	)
}
