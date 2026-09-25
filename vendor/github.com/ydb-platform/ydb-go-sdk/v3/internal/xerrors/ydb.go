package xerrors

import (
	"errors"
)

type isYdbError interface {
	isYdbError()
}

func IsYdb(err error) bool {
	var e isYdbError

	return errors.As(err, &e)
}

type ydbError struct {
	err error
}

func (e *ydbError) isYdbError() {}

func (e *ydbError) Error() string {
	return e.err.Error()
}

func (e *ydbError) Unwrap() error {
	return e.err
}

// Wrap makes internal ydb error
func Wrap(err error) error {
	return &ydbError{err: err}
}
