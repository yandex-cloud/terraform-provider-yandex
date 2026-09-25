package log

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Ratelimiter returns trace.Ratelimiter with logging events from details
func Ratelimiter(l Logger, d trace.Detailer, opts ...Option) (t trace.Ratelimiter) {
	return t
}
