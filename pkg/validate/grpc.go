package validate

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// isStatusWithCode checks if any nested error matches provided code
func IsStatusWithCode(err error, code codes.Code) bool {
	grpcStatus, ok := status.FromError(err)
	check := ok && grpcStatus.Code() == code

	if check {
		return true
	}

	if nestedErr := errors.Unwrap(err); nestedErr != nil {
		return IsStatusWithCode(nestedErr, code)
	}

	return check
}
