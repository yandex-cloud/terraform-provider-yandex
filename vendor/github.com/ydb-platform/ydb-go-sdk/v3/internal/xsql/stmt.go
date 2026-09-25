package xsql

import (
	"context"
	"database/sql/driver"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type Stmt struct {
	conn      *Conn
	processor interface {
		Exec(ctx context.Context, sql string, params *params.Params) (driver.Result, error)
		Query(ctx context.Context, sql string, params *params.Params) (driver.RowsNextResultSet, error)
	}
	sql string
	ctx context.Context //nolint:containedctx
}

var (
	_ driver.Stmt             = &Stmt{}
	_ driver.StmtQueryContext = &Stmt{}
	_ driver.StmtExecContext  = &Stmt{}
)

func (stmt *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (_ driver.Rows, finalErr error) {
	onDone := trace.DatabaseSQLOnStmtQuery(stmt.conn.connector.Trace(), &ctx,
		stack.FunctionID("database/sql.(*Stmt).QueryContext", stack.Package("database/sql")),
		stmt.ctx, stmt.sql,
	)
	defer func() {
		onDone(finalErr)
	}()

	if !stmt.conn.cc.IsValid() {
		return nil, xerrors.WithStackTrace(xerrors.Retryable(errNotReadyConn,
			xerrors.Invalid(stmt),
			xerrors.Invalid(stmt.conn),
			xerrors.Invalid(stmt.conn.cc),
		))
	}

	sql, params, err := stmt.conn.toYdb(stmt.sql, args...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return stmt.processor.Query(ctx, sql, params)
}

func (stmt *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (_ driver.Result, finalErr error) {
	onDone := trace.DatabaseSQLOnStmtExec(stmt.conn.connector.Trace(), &ctx,
		stack.FunctionID("database/sql.(*Stmt).ExecContext", stack.Package("database/sql")),
		stmt.ctx, stmt.sql,
	)
	defer func() {
		onDone(finalErr)
	}()

	if !stmt.conn.cc.IsValid() {
		return nil, xerrors.WithStackTrace(xerrors.Retryable(errNotReadyConn,
			xerrors.Invalid(stmt),
			xerrors.Invalid(stmt.conn),
			xerrors.Invalid(stmt.conn.cc),
		))
	}

	sql, params, err := stmt.conn.toYdb(stmt.sql, args...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return stmt.processor.Exec(ctx, sql, params)
}

func (stmt *Stmt) NumInput() int {
	return -1
}

func (stmt *Stmt) Close() (finalErr error) {
	var (
		ctx    = stmt.ctx
		onDone = trace.DatabaseSQLOnStmtClose(stmt.conn.connector.Trace(), &ctx,
			stack.FunctionID("database/sql.(*Stmt).Close", stack.Package("database/sql")),
		)
	)
	defer func() {
		onDone(finalErr)
	}()

	return nil
}

func (stmt *Stmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, errDeprecated
}

func (stmt *Stmt) Query([]driver.Value) (driver.Rows, error) {
	return nil, errDeprecated
}
