package ydb

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	grpcCodes "google.golang.org/grpc/codes"

	ratelimiterErrors "github.com/ydb-platform/ydb-go-sdk/v3/internal/ratelimiter/errors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/ratelimiter"
)

// IterateByIssues helps to iterate over internal issues of operation error.
func IterateByIssues(err error, it func(message string, code Ydb.StatusIds_StatusCode, severity uint32)) {
	xerrors.IterateByIssues(err, it)
}

// IsTimeoutError checks whether given err is a some timeout error (context, transport or operation).
func IsTimeoutError(err error) bool {
	return xerrors.IsTimeoutError(err)
}

// IsTransportError checks whether given err is a transport (grpc) error.
func IsTransportError(err error, codes ...grpcCodes.Code) bool {
	return xerrors.IsTransportError(err, codes...)
}

// Error is an interface of error which reports about error code and error name.
type Error interface {
	error

	// Code reports the error code
	Code() int32

	// Name reports the short name of error
	Name() string
}

// TransportError checks when given error is a transport error and returns description of transport error.
func TransportError(err error) Error {
	return xerrors.TransportError(err)
}

// IsYdbError reports when given error is and ydb error (transport, operation or internal driver error)
func IsYdbError(err error) bool {
	return xerrors.IsYdb(err)
}

// IsOperationError reports whether any error is an operation error with one of passed codes.
// If codes not defined IsOperationError returns true on error is an operation error.
func IsOperationError(err error, codes ...Ydb.StatusIds_StatusCode) bool {
	return xerrors.IsOperationError(err, codes...)
}

// OperationError returns operation error description.
// If given err is not an operation error - returns nil.
func OperationError(err error) Error {
	return xerrors.OperationError(err)
}

// IsOperationErrorOverloaded checks whether given err is an operation error with code Overloaded
func IsOperationErrorOverloaded(err error) bool {
	return IsOperationError(err, Ydb.StatusIds_OVERLOADED)
}

// IsOperationErrorUnavailable checks whether given err is an operation error with code Unavailable
func IsOperationErrorUnavailable(err error) bool {
	return IsOperationError(err, Ydb.StatusIds_UNAVAILABLE)
}

// IsOperationErrorAlreadyExistsError checks whether given err is an operation error with code AlreadyExistsError
func IsOperationErrorAlreadyExistsError(err error) bool {
	return IsOperationError(err, Ydb.StatusIds_ALREADY_EXISTS)
}

// IsOperationErrorNotFoundError checks whether given err is an operation error with code NotFoundError
func IsOperationErrorNotFoundError(err error) bool {
	return IsOperationError(err, Ydb.StatusIds_NOT_FOUND)
}

// IsOperationErrorSchemeError checks whether given err is an operation error with code SchemeError
func IsOperationErrorSchemeError(err error) bool {
	return IsOperationError(err, Ydb.StatusIds_SCHEME_ERROR)
}

// IsOperationErrorTransactionLocksInvalidated checks does err a TLI issue
//
//nolint:nonamedreturns
func IsOperationErrorTransactionLocksInvalidated(err error) (isTLI bool) {
	return xerrors.IsOperationErrorTransactionLocksInvalidated(err)
}

// IsRatelimiterAcquireError checks whether given err is an ratelimiter acquire error
func IsRatelimiterAcquireError(err error) bool {
	return ratelimiterErrors.IsAcquireError(err)
}

// ToRatelimiterAcquireError casts given err to ratelimiter.AcquireError.
// If given err is not ratelimiter acquire error - returns nil
func ToRatelimiterAcquireError(err error) ratelimiter.AcquireError {
	return ratelimiterErrors.ToAcquireError(err)
}
