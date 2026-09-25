package log

import (
	"context"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/kv"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Table makes trace.Table with logging events from details
func Table(l Logger, d trace.Detailer, opts ...Option) (t trace.Table) {
	return internalTable(wrapLogger(l, opts...), d)
}

//nolint:gocyclo,funlen
func internalTable(l *wrapper, d trace.Detailer) (t trace.Table) {
	return trace.Table{
		OnInit: func(info trace.TableInitStartInfo) func(trace.TableInitDoneInfo) {
			if d.Details()&trace.TableEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "init")
			l.Log(ctx, "table init starting...")
			start := time.Now()

			return func(info trace.TableInitDoneInfo) {
				l.Log(WithLevel(ctx, INFO), "table init done",
					kv.Latency(start),
					kv.Int("size_max", info.Limit),
				)
			}
		},
		OnClose: func(info trace.TableCloseStartInfo) func(trace.TableCloseDoneInfo) {
			if d.Details()&trace.TableEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "close")
			l.Log(ctx, "table close starting...")
			start := time.Now()

			return func(info trace.TableCloseDoneInfo) {
				if info.Error == nil {
					l.Log(WithLevel(ctx, INFO), "table close done",
						kv.Latency(start),
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "table close failed",
						kv.Latency(start),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnDo: func(
			info trace.TableDoStartInfo,
		) func(
			trace.TableDoDoneInfo,
		) {
			if d.Details()&trace.TablePoolAPIEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "do")
			idempotent := info.Idempotent
			label := info.Label
			l.Log(ctx, "table do starting...",
				kv.Bool("idempotent", idempotent),
				kv.String("label", label),
			)
			start := time.Now()

			return func(info trace.TableDoDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table do done",
						kv.Latency(start),
						kv.Bool("idempotent", idempotent),
						kv.String("label", label),
						kv.Int("attempts", info.Attempts),
					)
				} else {
					lvl := ERROR
					if !xerrors.IsYdb(info.Error) {
						lvl = DEBUG
					}
					m := retry.Check(info.Error)
					l.Log(WithLevel(ctx, lvl), "table do failed",
						kv.Latency(start),
						kv.Bool("idempotent", idempotent),
						kv.String("label", label),
						kv.Int("attempts", info.Attempts),
						kv.Error(info.Error),
						kv.Bool("retryable", m.MustRetry(idempotent)),
						kv.Int64("code", m.StatusCode()),
						kv.Bool("deleteSession", m.IsRetryObjectValid()),
						kv.Version(),
					)
				}
			}
		},
		OnDoTx: func(
			info trace.TableDoTxStartInfo,
		) func(
			trace.TableDoTxDoneInfo,
		) {
			if d.Details()&trace.TablePoolAPIEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "do", "tx")
			idempotent := info.Idempotent
			label := info.Label
			l.Log(ctx, "table dotx starting...",
				kv.Bool("idempotent", idempotent),
				kv.String("label", label),
			)
			start := time.Now()

			return func(info trace.TableDoTxDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table dotx done",
						kv.Latency(start),
						kv.Bool("idempotent", idempotent),
						kv.String("label", label),
						kv.Int("attempts", info.Attempts),
					)
				} else {
					lvl := WARN
					if !xerrors.IsYdb(info.Error) {
						lvl = DEBUG
					}
					m := retry.Check(info.Error)
					l.Log(WithLevel(ctx, lvl), "table dotx failed",
						kv.Latency(start),
						kv.Bool("idempotent", idempotent),
						kv.String("label", label),
						kv.Int("attempts", info.Attempts),
						kv.Error(info.Error),
						kv.Bool("retryable", m.MustRetry(idempotent)),
						kv.Int64("code", m.StatusCode()),
						kv.Bool("deleteSession", m.IsRetryObjectValid()),
						kv.Version(),
					)
				}
			}
		},
		OnCreateSession: func(
			info trace.TableCreateSessionStartInfo,
		) func(
			trace.TableCreateSessionDoneInfo,
		) {
			if d.Details()&trace.TablePoolAPIEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "create", "session")
			l.Log(ctx, "table create session starting...")
			start := time.Now()

			return func(info trace.TableCreateSessionDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table create session done",
						kv.Latency(start),
						kv.Int("attempts", info.Attempts),
						kv.String("session_id", info.Session.ID()),
						kv.String("session_status", info.Session.Status()),
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "table create session failed",
						kv.Latency(start),
						kv.Int("attempts", info.Attempts),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnSessionNew: func(info trace.TableSessionNewStartInfo) func(trace.TableSessionNewDoneInfo) {
			if d.Details()&trace.TableSessionEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "new")
			l.Log(ctx, "table new session starting")
			start := time.Now()

			return func(info trace.TableSessionNewDoneInfo) {
				if info.Error == nil {
					if info.Session != nil {
						l.Log(ctx, "table new session done",
							kv.Latency(start),
							kv.String("id", info.Session.ID()),
						)
					} else {
						l.Log(WithLevel(ctx, WARN), "table new session failed without error",
							kv.Latency(start),
							kv.Version(),
						)
					}
				} else {
					l.Log(WithLevel(ctx, WARN), "table new session failed",
						kv.Latency(start),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnSessionDelete: func(info trace.TableSessionDeleteStartInfo) func(trace.TableSessionDeleteDoneInfo) {
			if d.Details()&trace.TableSessionEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "delete")
			session := info.Session
			l.Log(ctx, "table session delete starting...",
				kv.String("id", info.Session.ID()),
				kv.String("status", info.Session.Status()),
			)
			start := time.Now()

			return func(info trace.TableSessionDeleteDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table session delete done",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
					)
				} else {
					l.Log(WithLevel(ctx, WARN), "table session delete failed",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnSessionKeepAlive: func(info trace.TableKeepAliveStartInfo) func(trace.TableKeepAliveDoneInfo) {
			if d.Details()&trace.TableSessionEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "keep", "alive")
			session := info.Session
			l.Log(ctx, "table session keepalive starting...",
				kv.String("id", session.ID()),
				kv.String("status", session.Status()),
			)
			start := time.Now()

			return func(info trace.TableKeepAliveDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table session keepalive done",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
					)
				} else {
					l.Log(WithLevel(ctx, WARN), "table session keepalive failed",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnSessionQueryPrepare: func(
			info trace.TablePrepareDataQueryStartInfo,
		) func(
			trace.TablePrepareDataQueryDoneInfo,
		) {
			if d.Details()&trace.TableSessionQueryInvokeEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "query", "prepare")
			session := info.Session
			query := info.Query
			l.Log(ctx, "table session query prepare starting...",
				appendFieldByCondition(l.logQuery,
					kv.String("query", info.Query),
					kv.String("id", session.ID()),
					kv.String("status", session.Status()),
				)...,
			)
			start := time.Now()

			return func(info trace.TablePrepareDataQueryDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table session query prepare done",
						appendFieldByCondition(l.logQuery,
							kv.Stringer("result", info.Result),
							appendFieldByCondition(l.logQuery,
								kv.String("query", query),
								kv.String("id", session.ID()),
								kv.String("status", session.Status()),
								kv.Latency(start),
							)...,
						)...,
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "table session query prepare failed",
						appendFieldByCondition(l.logQuery,
							kv.String("query", query),
							kv.Error(info.Error),
							kv.String("id", session.ID()),
							kv.String("status", session.Status()),
							kv.Latency(start),
							kv.Version(),
						)...,
					)
				}
			}
		},
		OnSessionQueryExecute: func(
			info trace.TableExecuteDataQueryStartInfo,
		) func(
			trace.TableExecuteDataQueryDoneInfo,
		) {
			if d.Details()&trace.TableSessionQueryInvokeEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "query", "execute")
			session := info.Session
			query := info.Query
			l.Log(ctx, "table session query execute starting...",
				appendFieldByCondition(l.logQuery,
					kv.Stringer("query", info.Query),
					kv.String("id", session.ID()),
					kv.String("status", session.Status()),
				)...,
			)
			start := time.Now()

			return func(info trace.TableExecuteDataQueryDoneInfo) {
				if info.Error == nil {
					tx := info.Tx
					l.Log(ctx, "table session query execute done",
						appendFieldByCondition(l.logQuery,
							kv.Stringer("query", query),
							kv.String("id", session.ID()),
							kv.String("tx", tx.ID()),
							kv.String("status", session.Status()),
							kv.Bool("prepared", info.Prepared),
							kv.NamedError("result_err", info.Result.Err()),
							kv.Latency(start),
						)...,
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "table session query execute failed",
						appendFieldByCondition(l.logQuery,
							kv.Stringer("query", query),
							kv.Error(info.Error),
							kv.String("id", session.ID()),
							kv.String("status", session.Status()),
							kv.Bool("prepared", info.Prepared),
							kv.Latency(start),
							kv.Version(),
						)...,
					)
				}
			}
		},
		OnSessionQueryStreamExecute: func(
			info trace.TableSessionQueryStreamExecuteStartInfo,
		) func(
			trace.TableSessionQueryStreamExecuteDoneInfo,
		) {
			if d.Details()&trace.TableSessionQueryStreamEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "query", "stream", "execute")
			session := info.Session
			query := info.Query
			l.Log(ctx, "table session query stream execute starting...",
				appendFieldByCondition(l.logQuery,
					kv.Stringer("query", info.Query),
					kv.String("id", session.ID()),
					kv.String("status", session.Status()),
				)...,
			)
			start := time.Now()

			return func(info trace.TableSessionQueryStreamExecuteDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table session query stream execute done",
						appendFieldByCondition(l.logQuery,
							kv.Stringer("query", query),
							kv.Error(info.Error),
							kv.String("id", session.ID()),
							kv.String("status", session.Status()),
							kv.Latency(start),
						)...,
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "table session query stream execute failed",
						appendFieldByCondition(l.logQuery,
							kv.Stringer("query", query),
							kv.Error(info.Error),
							kv.String("id", session.ID()),
							kv.String("status", session.Status()),
							kv.Latency(start),
							kv.Version(),
						)...,
					)
				}
			}
		},
		OnSessionQueryStreamRead: func(
			info trace.TableSessionQueryStreamReadStartInfo,
		) func(
			trace.TableSessionQueryStreamReadDoneInfo,
		) {
			if d.Details()&trace.TableSessionQueryStreamEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "query", "stream", "read")
			session := info.Session
			l.Log(ctx, "table session query stream read starting...",
				kv.String("id", session.ID()),
				kv.String("status", session.Status()),
			)
			start := time.Now()

			return func(info trace.TableSessionQueryStreamReadDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table session query stream read done",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "table session query stream read failed",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnTxBegin: func(
			info trace.TableTxBeginStartInfo,
		) func(
			trace.TableTxBeginDoneInfo,
		) {
			if d.Details()&trace.TableSessionTransactionEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "tx", "begin")
			session := info.Session
			l.Log(ctx, "table tx begin starting...",
				kv.String("id", session.ID()),
				kv.String("status", session.Status()),
			)
			start := time.Now()

			return func(info trace.TableTxBeginDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table tx begin done",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.String("tx", info.Tx.ID()),
					)
				} else {
					l.Log(WithLevel(ctx, WARN), "table tx begin failed",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnTxCommit: func(info trace.TableTxCommitStartInfo) func(trace.TableTxCommitDoneInfo) {
			if d.Details()&trace.TableSessionTransactionEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "tx", "commit")
			session := info.Session
			tx := info.Tx
			l.Log(ctx, "table tx commit starting...",
				kv.String("id", session.ID()),
				kv.String("status", session.Status()),
				kv.String("tx", info.Tx.ID()),
			)
			start := time.Now()

			return func(info trace.TableTxCommitDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table tx commit done",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.String("tx", tx.ID()),
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "table tx commit failed",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.String("tx", tx.ID()),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnTxRollback: func(
			info trace.TableTxRollbackStartInfo,
		) func(
			trace.TableTxRollbackDoneInfo,
		) {
			if d.Details()&trace.TableSessionTransactionEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "session", "tx", "rollback")
			session := info.Session
			tx := info.Tx
			l.Log(ctx, "table tx rollback starting...",
				kv.String("id", session.ID()),
				kv.String("status", session.Status()),
				kv.String("tx", tx.ID()),
			)
			start := time.Now()

			return func(info trace.TableTxRollbackDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table tx rollback done",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.String("tx", tx.ID()),
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "table tx rollback failed",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.String("tx", tx.ID()),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnPoolPut: func(info trace.TablePoolPutStartInfo) func(trace.TablePoolPutDoneInfo) {
			if d.Details()&trace.TablePoolAPIEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "pool", "put")
			session := info.Session
			l.Log(ctx, "table pool put starting...",
				kv.String("id", session.ID()),
				kv.String("status", session.Status()),
			)
			start := time.Now()

			return func(info trace.TablePoolPutDoneInfo) {
				if info.Error == nil {
					l.Log(ctx, "table pool put done",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
					)
				} else {
					l.Log(WithLevel(ctx, ERROR), "table pool put failed",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnPoolGet: func(info trace.TablePoolGetStartInfo) func(trace.TablePoolGetDoneInfo) {
			if d.Details()&trace.TablePoolAPIEvents == 0 {
				return nil
			}
			ctx := with(*info.Context, TRACE, "ydb", "table", "pool", "get")
			l.Log(ctx, "table pool get starting...")
			start := time.Now()

			return func(info trace.TablePoolGetDoneInfo) {
				if info.Error == nil {
					session := info.Session
					l.Log(ctx, "done",
						kv.Latency(start),
						kv.String("id", session.ID()),
						kv.String("status", session.Status()),
						kv.Int("attempts", info.Attempts),
					)
				} else {
					l.Log(WithLevel(ctx, WARN), "failed",
						kv.Latency(start),
						kv.Int("attempts", info.Attempts),
						kv.Error(info.Error),
						kv.Version(),
					)
				}
			}
		},
		OnPoolStateChange: func(info trace.TablePoolStateChangeInfo) {
			if d.Details()&trace.TablePoolLifeCycleEvents == 0 {
				return
			}
			ctx := with(context.Background(), TRACE, "ydb", "table", "pool", "state", "change")
			l.Log(WithLevel(ctx, DEBUG), "table session pool state changed",
				kv.Int("limit", info.Limit),
				kv.Int("index", info.Index),
				kv.Int("idle", info.Idle),
				kv.Int("wait", info.Wait),
				kv.Int("create_in_progress", info.CreateInProgress),
			)
		},
		OnSessionBulkUpsert:   nil,
		OnSessionQueryExplain: nil,
		OnTxExecute:           nil,
		OnTxExecuteStatement:  nil,
		OnPoolWith:            nil,
	}
}
