package xerrors

import (
	"context"
	"errors"
	"io"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	grpcCodes "google.golang.org/grpc/codes"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/backoff"
)

// Error is an interface of error which reports about error code and error name.
type Error interface {
	error

	Code() int32
	Name() string
	Type() Type
	BackoffType() backoff.Type
}

func IsTimeoutError(err error) bool {
	switch {
	case
		IsOperationError(
			err,
			Ydb.StatusIds_TIMEOUT,
			Ydb.StatusIds_CANCELLED,
		),
		IsTransportError(err, grpcCodes.Canceled, grpcCodes.DeadlineExceeded),
		Is(
			err,
			context.DeadlineExceeded,
			context.Canceled,
		):
		return true
	default:
		return false
	}
}

func ErrIf(cond bool, err error) error {
	if cond {
		return err
	}

	return nil
}

func HideEOF(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, io.EOF) {
		return nil
	}

	return err
}

// As is a proxy to errors.As
// This need to single import errors
func As(err error, targets ...interface{}) bool {
	if err == nil {
		panic("nil err")
	}
	if len(targets) == 0 {
		panic("empty targets")
	}
	for _, t := range targets {
		if errors.As(err, t) {
			return true
		}
	}

	return false
}

// IsErrorFromServer return true if err returned from server
// (opposite to raised internally in sdk)
func IsErrorFromServer(err error) bool {
	return IsTransportError(err) || IsOperationError(err)
}

// Is is a improved proxy to errors.Is
// This need to single import errors
func Is(err error, targets ...error) bool {
	if len(targets) == 0 {
		panic("empty targets")
	}
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

func IsContextError(err error) bool {
	return Is(err, context.Canceled, context.DeadlineExceeded)
}
