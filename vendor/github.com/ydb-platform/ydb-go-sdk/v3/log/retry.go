package log

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/kv"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Retry returns trace.Retry with logging events from details
func Retry(l Logger, d trace.Detailer, opts ...Option) (t trace.Retry) {
	return internalRetry(wrapLogger(l, opts...), d)
}

func internalRetry(l Logger, d trace.Detailer) (t trace.Retry) {
	t.OnRetry = func(info trace.RetryLoopStartInfo) func(trace.RetryLoopDoneInfo) {
		if d.Details()&trace.RetryEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "retrier starting work...", "retry")
		label := info.Label
		idempotent := info.Idempotent
		l.Log(ctx, "start",
			kv.String("label", label),
			kv.Bool("idempotent", idempotent),
		)
		start := time.Now()

		return func(info trace.RetryLoopDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "retries completed successfully",
					kv.String("label", label),
					kv.Latency(start),
					kv.Int("attempts", info.Attempts),
				)
			} else {
				lvl := ERROR
				if !xerrors.IsYdb(info.Error) {
					lvl = DEBUG
				}
				m := retry.Check(info.Error)
				l.Log(WithLevel(ctx, lvl), "retries failed",
					kv.Error(info.Error),
					kv.String("label", label),
					kv.Latency(start),
					kv.Int("attempts", info.Attempts),
					kv.Bool("retryable", m.MustRetry(idempotent)),
					kv.Int64("code", m.StatusCode()),
					kv.Bool("deleteSession", m.IsRetryObjectValid()),
					kv.Version(),
				)
			}
		}
	}

	return t
}
