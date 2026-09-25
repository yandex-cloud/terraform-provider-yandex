package errors

import (
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/ratelimiter"
)

type acquireError struct {
	err    error
	amount uint64
}

func (e *acquireError) Amount() uint64 {
	return e.amount
}

func (e *acquireError) Error() string {
	return fmt.Sprintf("acquire for amount (%d) failed: %v", e.amount, e.err)
}

func (e *acquireError) Unwrap() error {
	return e.err
}

func NewAcquire(amoount uint64, err error) ratelimiter.AcquireError {
	return &acquireError{
		err:    err,
		amount: amoount,
	}
}

func IsAcquireError(err error) bool {
	var ae *acquireError

	return xerrors.As(err, &ae)
}

func ToAcquireError(err error) ratelimiter.AcquireError {
	var ae *acquireError
	if xerrors.As(err, &ae) {
		return ae
	}

	return nil
}
