package retry

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

// retryMode reports whether operation is able retried and with which properties.
type retryMode struct {
	code    int64
	errType xerrors.Type
	backoff backoff.Type
}

func (m retryMode) MustRetry(isOperationIdempotent bool) bool {
	switch m.errType {
	case
		xerrors.TypeUndefined,
		xerrors.TypeNoError,
		xerrors.TypeNonRetryable:
		return false
	case xerrors.TypeConditionallyRetryable:
		return isOperationIdempotent
	default:
		return true
	}
}

func (m retryMode) StatusCode() int64 { return m.code }

func (m retryMode) MustBackoff() bool { return m.backoff&backoff.TypeAny != 0 }

func (m retryMode) BackoffType() backoff.Type { return m.backoff }

// MustDeleteSession
//
// Deprecated
func (m retryMode) MustDeleteSession() bool { return false }

// IsRetryObjectValid
//
// Deprecated
func (m retryMode) IsRetryObjectValid() bool { return true }
