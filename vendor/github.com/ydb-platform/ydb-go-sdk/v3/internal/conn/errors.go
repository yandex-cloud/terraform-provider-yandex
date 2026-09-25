package conn

import (
	grpcCodes "google.golang.org/grpc/codes"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

func IsBadConn(err error, goodConnCodes ...grpcCodes.Code) bool {
	if !xerrors.IsTransportError(err) {
		return false
	}

	if xerrors.IsTransportError(err,
		append(
			goodConnCodes,
			grpcCodes.ResourceExhausted,
			grpcCodes.OutOfRange,
			grpcCodes.OK,
			// grpcCodes.Canceled,
			// grpcCodes.Unknown,
			// grpcCodes.InvalidArgument,
			// grpcCodes.DeadlineExceeded,
			// grpcCodes.NotFound,
			// grpcCodes.AlreadyExists,
			// grpcCodes.PermissionDenied,
			// grpcCodes.FailedPrecondition,
			// grpcCodes.Aborted,
			// grpcCodes.OutOfRange,
			// grpcCodes.Unimplemented,
			// grpcCodes.Internal,
			// grpcCodes.DataLoss,
			// grpcCodes.Unauthenticated,
		)...,
	) {
		return false
	}

	return true
}
