package xsql

import (
	"context"
	"database/sql/driver"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/badconn"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/common"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type Tx struct {
	conn *Conn
	tx   common.Tx
	ctx  context.Context //nolint:containedctx
}

func (tx *Tx) ID() string {
	if tx.tx == nil {
		return ""
	}

	return tx.tx.ID()
}

var (
	_ driver.Tx             = &Tx{}
	_ driver.ExecerContext  = &Tx{}
	_ driver.QueryerContext = &Tx{}
)

func (tx *Tx) Commit() (finalErr error) {
	defer func() {
		tx.conn.currentTx = nil
	}()

	var (
		ctx    = tx.ctx
		onDone = trace.DatabaseSQLOnTxCommit(tx.conn.connector.Trace(), &ctx,
			stack.FunctionID("database/sql.(*Tx).Commit", stack.Package("database/sql")),
			tx,
		)
	)
	defer func() {
		onDone(finalErr)
	}()

	if err := tx.tx.Commit(tx.ctx); err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (tx *Tx) Rollback() (finalErr error) {
	defer func() {
		tx.conn.currentTx = nil
	}()

	var (
		ctx    = tx.ctx
		onDone = trace.DatabaseSQLOnTxRollback(tx.conn.connector.Trace(), &ctx,
			stack.FunctionID("database/sql.(*Tx).Rollback", stack.Package("database/sql")),
			tx,
		)
	)
	defer func() {
		onDone(finalErr)
	}()

	err := tx.tx.Rollback(tx.ctx)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return err
}

func (tx *Tx) QueryContext(ctx context.Context, sql string, args []driver.NamedValue) (
	_ driver.Rows, finalErr error,
) {
	onDone := trace.DatabaseSQLOnTxQuery(tx.conn.connector.Trace(), &ctx,
		stack.FunctionID("database/sql.(*Tx).QueryContext", stack.Package("database/sql")),
		tx.ctx, tx, sql,
	)
	defer func() {
		onDone(finalErr)
	}()

	sql, params, err := tx.conn.toYdb(sql, args...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	if isExplain(ctx) {
		ast, plan, err := tx.conn.cc.Explain(ctx, sql, params)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return rowByAstPlan(ast, plan), nil
	}

	rows, err := tx.tx.Query(ctx, sql, params)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return rows, nil
}

func (tx *Tx) ExecContext(ctx context.Context, sql string, args []driver.NamedValue) (
	_ driver.Result, finalErr error,
) {
	onDone := trace.DatabaseSQLOnTxExec(tx.conn.connector.Trace(), &ctx,
		stack.FunctionID("database/sql.(*Tx).ExecContext", stack.Package("database/sql")),
		tx.ctx, tx, sql,
	)
	defer func() {
		onDone(finalErr)
	}()

	sql, params, err := tx.conn.toYdb(sql, args...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	result, err := tx.tx.Exec(ctx, sql, params)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return result, nil
}

func (tx *Tx) PrepareContext(ctx context.Context, sql string) (_ driver.Stmt, finalErr error) {
	onDone := trace.DatabaseSQLOnTxPrepare(tx.conn.connector.Trace(), &ctx,
		stack.FunctionID("database/sql.(*Tx).PrepareContext", stack.Package("database/sql")),
		tx.ctx, tx, sql,
	)
	defer func() {
		onDone(finalErr)
	}()
	if !tx.conn.cc.IsValid() {
		return nil, badconn.Map(xerrors.WithStackTrace(xerrors.Retryable(errNotReadyConn,
			xerrors.Invalid(tx),
			xerrors.Invalid(tx.conn),
			xerrors.Invalid(tx.conn.cc),
		)))
	}

	return &Stmt{
		conn:      tx.conn,
		processor: tx.tx,
		ctx:       ctx,
		sql:       sql,
	}, nil
}
