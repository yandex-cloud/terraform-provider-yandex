package retry

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type retryableErrorOption xerrors.RetryableErrorOption

const (
	TypeNoBackoff   = backoff.TypeNoBackoff
	TypeFastBackoff = backoff.TypeFast
	TypeSlowBackoff = backoff.TypeSlow
)

// WithBackoff makes retryable error option with custom backoff type
func WithBackoff(t backoff.Type) retryableErrorOption {
	return retryableErrorOption(xerrors.WithBackoff(t))
}

// WithDeleteSession makes retryable error option with delete session flag
//
// Deprecated
func WithDeleteSession() retryableErrorOption {
	return nil
}

// RetryableError makes retryable error from options
// RetryableError provides retrying on custom errors
func RetryableError(err error, opts ...retryableErrorOption) error {
	return xerrors.Retryable(
		err,
		func() (retryableErrorOptions []xerrors.RetryableErrorOption) {
			for _, opt := range opts {
				if opt != nil {
					retryableErrorOptions = append(retryableErrorOptions, xerrors.RetryableErrorOption(opt))
				}
			}

			return retryableErrorOptions
		}()...,
	)
}
