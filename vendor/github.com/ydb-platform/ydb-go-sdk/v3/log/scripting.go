package log

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/kv"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Scripting returns trace.Scripting with logging events from details
func Scripting(l Logger, d trace.Detailer, opts ...Option) (t trace.Scripting) {
	return internalScripting(wrapLogger(l, opts...), d)
}

//nolint:funlen
func internalScripting(l *wrapper, d trace.Detailer) (t trace.Scripting) {
	t.OnExecute = func(info trace.ScriptingExecuteStartInfo) func(trace.ScriptingExecuteDoneInfo) {
		if d.Details()&trace.ScriptingEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "scripting", "execute")
		l.Log(ctx, "starting script executing...")
		start := time.Now()

		return func(info trace.ScriptingExecuteDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "start script done",
					kv.Latency(start),
					kv.Int("resultSetCount", info.Result.ResultSetCount()),
					kv.NamedError("resultSetError", info.Result.Err()),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "start script failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnExplain = func(info trace.ScriptingExplainStartInfo) func(trace.ScriptingExplainDoneInfo) {
		if d.Details()&trace.ScriptingEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "scripting", "explain")
		l.Log(ctx, "starting script explain...")
		start := time.Now()

		return func(info trace.ScriptingExplainDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "script explain done",
					kv.Latency(start),
					kv.String("plan", info.Plan),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "script explain failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnStreamExecute = func(
		info trace.ScriptingStreamExecuteStartInfo,
	) func(
		trace.ScriptingStreamExecuteIntermediateInfo,
	) func(
		trace.ScriptingStreamExecuteDoneInfo,
	) {
		if d.Details()&trace.ScriptingEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "scripting", "stream", "execute")
		query := info.Query
		l.Log(ctx, "script stream execute starting...",
			appendFieldByCondition(l.logQuery,
				kv.String("query", query),
			)...,
		)
		start := time.Now()

		return func(
			info trace.ScriptingStreamExecuteIntermediateInfo,
		) func(
			trace.ScriptingStreamExecuteDoneInfo,
		) {
			if info.Error == nil {
				l.Log(ctx, "script stream execute intermediate success")
			} else {
				l.Log(WithLevel(ctx, WARN), "script stream execute intermediate failed",
					kv.Error(info.Error),
					kv.Version(),
				)
			}

			return func(info trace.ScriptingStreamExecuteDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "script stream execute done",
						appendFieldByCondition(l.logQuery,
							kv.String("query", query),
							kv.Latency(start),
						)...,
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "script stream execute failed",
						appendFieldByCondition(l.logQuery,
							kv.String("query", query),
							kv.Error(info.Error),
							kv.Latency(start),
							kv.Version(),
						)...,
					)
				}
			}
		}
	}
	t.OnClose = func(info trace.ScriptingCloseStartInfo) func(trace.ScriptingCloseDoneInfo) {
		if d.Details()&trace.ScriptingEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "scripting", "close")
		l.Log(ctx, "script close starting...")
		start := time.Now()

		return func(info trace.ScriptingCloseDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "script close done",
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "script close failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}

	return t
}
