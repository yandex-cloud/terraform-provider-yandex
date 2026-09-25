package xerrors

import (
	"errors"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

type retryableError struct {
	name        string
	err         error
	backoffType backoff.Type
	code        int32
	traceID     string
	invalidObjs []any
}

func (re *retryableError) Code() int32 {
	return re.code
}

func (re *retryableError) Name() string {
	return "retryable/" + re.name
}

func (re *retryableError) Type() Type {
	return TypeRetryable
}

func (re *retryableError) BackoffType() backoff.Type {
	return re.backoffType
}

func (re *retryableError) Error() string {
	b := xstring.Buffer()
	defer b.Free()
	b.WriteString(re.Name())
	fmt.Fprintf(b, " (code = %d, source error = %q", re.code, re.err.Error())
	if len(re.traceID) > 0 {
		fmt.Fprintf(b, ", traceID: %q", re.traceID)
	}
	b.WriteString(")")

	return b.String()
}

func (re *retryableError) Unwrap() error {
	return re.err
}

type RetryableErrorOption interface {
	applyToRetryableError(re *retryableError)
}

var (
	_ RetryableErrorOption = backoffOption{}
	_ RetryableErrorOption = nameOption("")
	_ RetryableErrorOption = invalidObjectOption{}
)

type backoffOption struct {
	backoffType backoff.Type
}

func (t backoffOption) applyToRetryableError(re *retryableError) {
	re.backoffType = t.backoffType
}

func WithBackoff(t backoff.Type) backoffOption {
	return backoffOption{backoffType: t}
}

type nameOption string

func (name nameOption) applyToRetryableError(re *retryableError) {
	re.name = string(name)
}

func WithName(name string) nameOption {
	return nameOption(name)
}

type invalidObjectOption struct {
	obj any
}

func (opt invalidObjectOption) applyToRetryableError(re *retryableError) {
	re.invalidObjs = append(re.invalidObjs, opt.obj)
}

// InvalidObject deprecated option
//
// Deprecated: use Invalid instead
func InvalidObject() invalidObjectOption {
	return invalidObjectOption{}
}

func Invalid(obj any) invalidObjectOption {
	return invalidObjectOption{
		obj: obj,
	}
}

func Retryable(err error, opts ...RetryableErrorOption) error {
	if err == nil {
		return nil
	}
	var (
		e  Error
		re = &retryableError{
			err:  err,
			name: "CUSTOM",
			code: -1,
		}
	)
	if As(err, &e) {
		re.backoffType = e.BackoffType()
		re.code = e.Code()
		re.name = e.Name()
	}
	for _, opt := range opts {
		if opt != nil {
			opt.applyToRetryableError(re)
		}
	}

	return re
}

// RetryableError return Error if err is retriable error, else nil
func RetryableError(err error) Error {
	var unretriableErr unretryableError
	if errors.As(err, &unretriableErr) {
		return nil
	}

	var e *retryableError
	if errors.As(err, &e) {
		return e
	}

	return nil
}

func Unretryable(err error) unretryableError {
	return unretryableError{err}
}

type unretryableError struct {
	error
}

func (e unretryableError) Unwrap() error {
	return e.error
}

func IsRetryableError(err error) bool {
	if err == nil {
		return true
	}

	var e *retryableError
	if errors.As(err, &e) {
		return e != nil
	}

	return false
}

func IsValid[T comparable](err error, obj T) bool {
	if err == nil {
		return true
	}

	var e *retryableError
	if !errors.As(err, &e) {
		return true
	}

	for _, o := range e.invalidObjs {
		if invalidObj, has := o.(T); has && invalidObj == obj {
			return false
		}
	}

	return true
}
