package xerrors

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	grpcCodes "google.golang.org/grpc/codes"
)

func MustDeleteTableOrQuerySession(err error) bool {
	if IsOperationError(err,
		Ydb.StatusIds_BAD_SESSION,
		Ydb.StatusIds_SESSION_BUSY,
		Ydb.StatusIds_SESSION_EXPIRED,
	) {
		return true
	}

	if IsTransportError(err,
		grpcCodes.Canceled,
		grpcCodes.Unknown,
		grpcCodes.InvalidArgument,
		grpcCodes.DeadlineExceeded,
		grpcCodes.NotFound,
		grpcCodes.AlreadyExists,
		grpcCodes.PermissionDenied,
		grpcCodes.FailedPrecondition,
		grpcCodes.Aborted,
		grpcCodes.Unimplemented,
		grpcCodes.Internal,
		grpcCodes.Unavailable,
		grpcCodes.DataLoss,
		grpcCodes.Unauthenticated,
	) {
		return true
	}

	return false
}
