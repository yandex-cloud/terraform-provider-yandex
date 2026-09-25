package xerrors

import (
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xslices"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

func Join(errs ...error) error {
	errs = xslices.Filter(errs, func(err error) bool {
		return err != nil
	})

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return &joinError{
			errs: errs,
		}
	}
}

type joinError struct {
	errs []error
}

func (e *joinError) Error() string {
	b := xstring.Buffer()
	defer b.Free()
	b.WriteByte('[')
	for i, err := range e.errs {
		if i > 0 {
			_ = b.WriteByte(',')
		}
		_, _ = fmt.Fprintf(b, "%q", err.Error())
	}
	b.WriteByte(']')

	return b.String()
}

func (e *joinError) As(target interface{}) bool {
	for _, err := range e.errs {
		if As(err, target) {
			return true
		}
	}

	return false
}

func (e *joinError) Is(target error) bool {
	for _, err := range e.errs {
		if Is(err, target) {
			return true
		}
	}

	return false
}

func (e *joinError) Unwrap() []error {
	return e.errs
}
