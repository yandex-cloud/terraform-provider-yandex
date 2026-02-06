package iam_policy

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// parametersEqual compares two parameter maps
func parametersEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, valueA := range a {
		if valueB, exists := b[key]; !exists || valueA != valueB {
			return false
		}
	}
	return true
}

func IsStatusWithCode(err error, code codes.Code) bool {
	grpcStatus, ok := status.FromError(err)
	check := ok && grpcStatus.Code() == code

	nestedErr := errors.Unwrap(err)
	if nestedErr != nil {
		return IsStatusWithCode(nestedErr, code) || check
	}
	return check
}
