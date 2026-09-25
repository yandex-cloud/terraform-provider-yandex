package operation

import (
	"context"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"
)

func Params(
	ctx context.Context,
	timeout time.Duration,
	cancelAfter time.Duration,
	mode Mode,
) *Ydb_Operations.OperationParams {
	if d, ok := ctxTimeout(ctx); ok {
		timeout = d
	}
	if d, ok := ctxCancelAfter(ctx); ok {
		cancelAfter = d
	}
	if d, ok := ctxUntilDeadline(ctx); mode == ModeSync && ok && d < timeout {
		timeout = d
	}
	if timeout == 0 && cancelAfter == 0 && mode == 0 {
		return nil
	}

	return &Ydb_Operations.OperationParams{
		OperationMode:    mode.toYDB(),
		OperationTimeout: timeoutParam(timeout),
		CancelAfter:      timeoutParam(cancelAfter),
	}
}
