package log

import (
	"context"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/kv"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// DatabaseSQL makes trace.DatabaseSQL with logging events from details
func DatabaseSQL(l Logger, d trace.Detailer, opts ...Option) (t trace.DatabaseSQL) {
	return internalDatabaseSQL(wrapLogger(l, opts...), d)
}

//nolint:funlen
func internalDatabaseSQL(l *wrapper, d trace.Detailer) (t trace.DatabaseSQL) {
	t.OnConnectorConnect = func(
		info trace.DatabaseSQLConnectorConnectStartInfo,
	) func(
		trace.DatabaseSQLConnectorConnectDoneInfo,
	) {
		if d.Details()&trace.DatabaseSQLConnectorEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "connector", "connect")
		l.Log(ctx, "database/sql connect starting...")
		start := time.Now()

		return func(info trace.DatabaseSQLConnectorConnectDoneInfo) {
			if info.Error == nil {
				l.Log(WithLevel(ctx, DEBUG), "database/sql connect done",
					kv.Latency(start),
					kv.String("session_id", info.Session.ID()),
					kv.String("session_status", info.Session.Status()),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "database/sql connect failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}

	t.OnConnPing = func(info trace.DatabaseSQLConnPingStartInfo) func(trace.DatabaseSQLConnPingDoneInfo) {
		if d.Details()&trace.DatabaseSQLConnEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "conn", "ping")
		l.Log(ctx, "database/sql ping starting...")
		start := time.Now()

		return func(info trace.DatabaseSQLConnPingDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql ping done",
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "database/sql ping failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnConnClose = func(info trace.DatabaseSQLConnCloseStartInfo) func(trace.DatabaseSQLConnCloseDoneInfo) {
		if d.Details()&trace.DatabaseSQLConnEvents == 0 {
			return nil
		}
		ctx := with(context.Background(), TRACE, "ydb", "database", "sql", "conn", "close")
		l.Log(ctx, "sql connection close starting...")
		start := time.Now()

		return func(info trace.DatabaseSQLConnCloseDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql connection close done",
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "database/sql connection close failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnConnBegin = func(info trace.DatabaseSQLConnBeginStartInfo) func(trace.DatabaseSQLConnBeginDoneInfo) {
		if d.Details()&trace.DatabaseSQLConnEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "conn", "begin", "tx")
		l.Log(ctx, "database/sql transaction begin starting...")
		start := time.Now()

		return func(info trace.DatabaseSQLConnBeginDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql transaction begin done",
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "database/sql transaction begin failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnConnPrepare = func(info trace.DatabaseSQLConnPrepareStartInfo) func(trace.DatabaseSQLConnPrepareDoneInfo) {
		if d.Details()&trace.DatabaseSQLConnEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "conn", "prepare", "stmt")
		l.Log(ctx, "database/sql prepare starting...",
			appendFieldByCondition(l.logQuery,
				kv.String("query", info.Query),
			)...,
		)
		query := info.Query
		start := time.Now()

		return func(info trace.DatabaseSQLConnPrepareDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql prepare done",
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "database/sql prepare failed",
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
	t.OnConnExec = func(info trace.DatabaseSQLConnExecStartInfo) func(trace.DatabaseSQLConnExecDoneInfo) {
		if d.Details()&trace.DatabaseSQLConnEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "conn", "exec")
		l.Log(ctx, "database/sql exec starting...",
			appendFieldByCondition(l.logQuery,
				kv.String("query", info.Query),
			)...,
		)
		query := info.Query
		idempotent := info.Idempotent
		start := time.Now()

		return func(info trace.DatabaseSQLConnExecDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql exec done",
					kv.Latency(start),
				)
			} else {
				m := retry.Check(info.Error)
				l.Log(WithLevel(ctx, ERROR), "database/sql exec failed",
					appendFieldByCondition(l.logQuery,
						kv.String("query", query),
						kv.Bool("retryable", m.MustRetry(idempotent)),
						kv.Int64("code", m.StatusCode()),
						kv.Bool("deleteSession", m.IsRetryObjectValid()),
						kv.Error(info.Error),
						kv.Latency(start),
						kv.Version(),
					)...,
				)
			}
		}
	}
	t.OnConnQuery = func(info trace.DatabaseSQLConnQueryStartInfo) func(trace.DatabaseSQLConnQueryDoneInfo) {
		if d.Details()&trace.DatabaseSQLConnEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "conn", "query")
		l.Log(ctx, "database/sql query starting...",
			appendFieldByCondition(l.logQuery,
				kv.String("query", info.Query),
			)...,
		)
		query := info.Query
		idempotent := info.Idempotent
		start := time.Now()

		return func(info trace.DatabaseSQLConnQueryDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql query done",
					kv.Latency(start),
				)
			} else {
				m := retry.Check(info.Error)
				l.Log(WithLevel(ctx, ERROR), "database/sql query failed",
					appendFieldByCondition(l.logQuery,
						kv.String("query", query),
						kv.Bool("retryable", m.MustRetry(idempotent)),
						kv.Int64("code", m.StatusCode()),
						kv.Bool("deleteSession", m.IsRetryObjectValid()),
						kv.Error(info.Error),
						kv.Latency(start),
						kv.Version(),
					)...,
				)
			}
		}
	}
	t.OnTxCommit = func(info trace.DatabaseSQLTxCommitStartInfo) func(trace.DatabaseSQLTxCommitDoneInfo) {
		if d.Details()&trace.DatabaseSQLTxEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "tx", "commit")
		l.Log(ctx, "database/sql tx commit starting")
		start := time.Now()

		return func(info trace.DatabaseSQLTxCommitDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql tx commit done",
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "database/sql tx commit failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnTxRollback = func(info trace.DatabaseSQLTxRollbackStartInfo) func(trace.DatabaseSQLTxRollbackDoneInfo) {
		if d.Details()&trace.DatabaseSQLTxEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "tx", "rollback")
		l.Log(ctx, "database/sql tx rollback starting...")
		start := time.Now()

		return func(info trace.DatabaseSQLTxRollbackDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql tx rollback done",
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "database/sql tx rollback failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnStmtClose = func(info trace.DatabaseSQLStmtCloseStartInfo) func(trace.DatabaseSQLStmtCloseDoneInfo) {
		if d.Details()&trace.DatabaseSQLStmtEvents == 0 {
			return nil
		}
		ctx := with(context.Background(), TRACE, "ydb", "database", "sql", "stmt", "close")
		l.Log(ctx, "database/sql statement close starting...")
		start := time.Now()

		return func(info trace.DatabaseSQLStmtCloseDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql statement close done",
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "database/sql statement close failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnStmtExec = func(info trace.DatabaseSQLStmtExecStartInfo) func(trace.DatabaseSQLStmtExecDoneInfo) {
		if d.Details()&trace.DatabaseSQLStmtEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "stmt", "exec")
		l.Log(ctx, "database/sql statement exec starting...",
			appendFieldByCondition(l.logQuery,
				kv.String("query", info.Query),
			)...,
		)
		query := info.Query
		start := time.Now()

		return func(info trace.DatabaseSQLStmtExecDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql statement exec done",
					kv.Error(info.Error),
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "database/sql statement exec failed",
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
	t.OnStmtQuery = func(info trace.DatabaseSQLStmtQueryStartInfo) func(trace.DatabaseSQLStmtQueryDoneInfo) {
		if d.Details()&trace.DatabaseSQLStmtEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "database", "sql", "stmt", "query")
		l.Log(ctx, "database/sql statement query starting...",
			appendFieldByCondition(l.logQuery,
				kv.String("query", info.Query),
			)...,
		)
		query := info.Query
		start := time.Now()

		return func(info trace.DatabaseSQLStmtQueryDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "database/sql statement query done",
					kv.Latency(start),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "database/sql statement query failed",
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

	return t
}
