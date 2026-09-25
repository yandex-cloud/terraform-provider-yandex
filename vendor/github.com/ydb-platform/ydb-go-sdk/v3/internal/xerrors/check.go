package xerrors

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
)

// Check returns retry mode for err.
func Check(err error) (
	code int64,
	errType Type,
	backoffType backoff.Type,
) {
	if err == nil {
		return -1,
			TypeNoError,
			backoff.TypeNoBackoff
	}
	var e Error
	if As(err, &e) {
		return int64(e.Code()), e.Type(), e.BackoffType()
	}

	return -1,
		TypeNonRetryable, // unknown errors are not retryable
		backoff.TypeNoBackoff
}
