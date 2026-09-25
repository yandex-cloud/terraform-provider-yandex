package pool

import (
	"errors"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	grpcCodes "google.golang.org/grpc/codes"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var (
	errClosedPool     = errors.New("closed pool")
	errItemIsNotAlive = errors.New("item is not alive")
	errPoolIsOverflow = errors.New("pool is overflow")
	errNoProgress     = errors.New("no progress")
)

func isRetriable(err error) bool {
	if err == nil {
		panic("nil")
	}

	switch {
	case
		xerrors.Is(err, errPoolIsOverflow, errItemIsNotAlive, errNoProgress),
		xerrors.IsRetryableError(err),
		xerrors.IsOperationError(err, Ydb.StatusIds_OVERLOADED),
		xerrors.IsTransportError(
			err,
			grpcCodes.ResourceExhausted,
			grpcCodes.DeadlineExceeded,
			grpcCodes.Unavailable,
		):
		return true
	default:
		return false
	}
}
