package utils

import "google.golang.org/grpc/status"

func ErrorMessage(err error) string {
	grpcStatus, _ := status.FromError(err)
	return grpcStatus.Message()
}
