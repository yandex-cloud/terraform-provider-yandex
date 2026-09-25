package xsql

import (
	"context"
	"database/sql/driver"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/badconn"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/common"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type Conn struct {
	processor Engine

	cc        common.Conn
	currentTx *Tx
	ctx       context.Context //nolint:containedctx

	connector *Connector
	lastUsage xsync.LastUsage
}

func (c *Conn) ID() string {
	return c.cc.ID()
}

func (c *Conn) NodeID() uint32 {
	return c.cc.NodeID()
}

func (c *Conn) Ping(ctx context.Context) (finalErr error) {
	onDone := trace.DatabaseSQLOnConnPing(c.connector.trace, &c.ctx,
		stack.FunctionID("database/sql.(*Conn).Ping", stack.Package("database/sql")),
	)
	defer func() {
		onDone(finalErr)
	}()

	return c.cc.Ping(ctx)
}

func (c *Conn) CheckNamedValue(value *driver.NamedValue) (finalErr error) {
	onDone := trace.DatabaseSQLOnConnCheckNamedValue(c.connector.trace, &c.ctx,
		stack.FunctionID("database/sql.(*Conn).CheckNamedValue", stack.Package("database/sql")),
		value,
	)
	defer func() {
		onDone(finalErr)
	}()

	// on this stage allows all values
	return nil
}

func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (_ driver.Tx, finalErr error) {
	onDone := trace.DatabaseSQLOnConnBeginTx(c.connector.trace, &ctx,
		stack.FunctionID("database/sql.(*Conn).BeginTx", stack.Package("database/sql")),
	)
	defer func() {
		if c.currentTx != nil {
			onDone(c.currentTx, finalErr)
		} else {
			onDone(nil, finalErr)
		}
	}()

	if c.currentTx != nil {
		return nil, xerrors.WithStackTrace(xerrors.AlreadyHasTx(c.currentTx.ID()))
	}

	tx, err := c.cc.BeginTx(ctx, opts)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	c.currentTx = &Tx{
		conn: c,
		ctx:  ctx,
		tx:   tx,
	}

	return c.currentTx, nil
}

func (c *Conn) Close() (finalErr error) {
	onDone := trace.DatabaseSQLOnConnClose(c.connector.Trace(), &c.ctx,
		stack.FunctionID("database/sql.(*Conn).Close", stack.Package("database/sql")),
	)
	defer func() {
		onDone(finalErr)
	}()

	err := c.cc.Close()
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (c *Conn) Begin() (_ driver.Tx, finalErr error) {
	onDone := trace.DatabaseSQLOnConnBegin(c.connector.trace, &c.ctx,
		stack.FunctionID("database/sql.(*Conn).Begin", stack.Package("database/sql")),
	)
	defer func() {
		if c.currentTx != nil {
			onDone(c.currentTx, finalErr)
		} else {
			onDone(nil, finalErr)
		}
	}()

	if c.currentTx != nil {
		return nil, xerrors.WithStackTrace(xerrors.AlreadyHasTx(c.currentTx.ID()))
	}

	return nil, xerrors.WithStackTrace(errDeprecated)
}

func (c *Conn) Prepare(string) (driver.Stmt, error) {
	return nil, errDeprecated
}

func (c *Conn) PrepareContext(ctx context.Context, sql string) (_ driver.Stmt, finalErr error) {
	onDone := trace.DatabaseSQLOnConnPrepare(c.connector.Trace(), &ctx,
		stack.FunctionID("database/sql.(*Conn).PrepareContext", stack.Package("database/sql")),
		sql,
	)
	defer func() {
		onDone(finalErr)
	}()

	if !c.cc.IsValid() {
		return nil, xerrors.WithStackTrace(xerrors.Retryable(errNotReadyConn,
			xerrors.Invalid(c),
			xerrors.Invalid(c.cc),
		))
	}

	return &Stmt{
		conn:      c,
		processor: c.cc,
		ctx:       ctx,
		sql:       sql,
	}, nil
}

func (c *Conn) QueryContext(ctx context.Context, sql string, args []driver.NamedValue) (
	_ driver.Rows, finalErr error,
) {
	onDone := trace.DatabaseSQLOnConnQuery(c.connector.Trace(), &ctx,
		stack.FunctionID("database/sql.(*Conn).QueryContext", stack.Package("database/sql")),
		sql, c.connector.processor.String(), xcontext.IsIdempotent(ctx), c.connector.clock.Since(c.lastUsage.Get()),
	)
	defer func() {
		onDone(finalErr)
	}()

	done := c.lastUsage.Start()
	defer done()

	sql, params, err := c.toYdb(sql, args...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	if isExplain(ctx) {
		ast, plan, err := c.cc.Explain(ctx, sql, params)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return rowByAstPlan(ast, plan), nil
	}

	if c.currentTx != nil {
		return c.currentTx.tx.Query(ctx, sql, params)
	}

	result, err := c.cc.Query(ctx, sql, params)
	if err != nil {
		return nil, badconn.Map(err)
	}

	return result, nil
}

func (c *Conn) ExecContext(ctx context.Context, sql string, args []driver.NamedValue) (
	_ driver.Result, finalErr error,
) {
	onDone := trace.DatabaseSQLOnConnExec(c.connector.Trace(), &ctx,
		stack.FunctionID("database/sql.(*Conn).ExecContext", stack.Package("database/sql")),
		sql, c.connector.processor.String(), xcontext.IsIdempotent(ctx), c.connector.clock.Since(c.lastUsage.Get()),
	)
	defer func() {
		onDone(finalErr)
	}()

	done := c.lastUsage.Start()
	defer done()

	sql, params, err := c.toYdb(sql, args...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	if c.currentTx != nil {
		return c.currentTx.tx.Exec(ctx, sql, params)
	}

	result, err := c.cc.Exec(ctx, sql, params)
	if err != nil {
		return nil, badconn.Map(err)
	}

	return result, nil
}
