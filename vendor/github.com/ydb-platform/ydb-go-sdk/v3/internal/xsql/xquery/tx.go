package xquery

import (
	"context"
	"database/sql/driver"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/common"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

type transaction struct {
	conn *Conn
	tx   query.Transaction
}

func (t *transaction) ID() string {
	return t.tx.ID()
}

func (t *transaction) Exec(ctx context.Context, sql string, params *params.Params) (driver.Result, error) {
	opts := []query.ExecuteOption{
		options.WithParameters(params),
	}

	if txControl := tx.ControlFromContext(ctx, nil); txControl != nil {
		opts = append(opts, options.WithTxControl(txControl))
	}

	err := t.tx.Exec(ctx, sql, opts...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return resultNoRows{}, nil
}

func (t *transaction) Query(ctx context.Context, sql string, params *params.Params) (driver.RowsNextResultSet, error) {
	opts := []query.ExecuteOption{
		options.WithParameters(params),
	}

	if txControl := tx.ControlFromContext(ctx, nil); txControl != nil {
		opts = append(opts, options.WithTxControl(txControl))
	}

	res, err := t.tx.Query(ctx, sql, opts...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return &rows{
		conn:   t.conn,
		result: res,
	}, nil
}

func beginTx(ctx context.Context, c *Conn, txOptions driver.TxOptions) (common.Tx, error) {
	txc, err := toYDB(txOptions)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	nativeTx, err := c.session.Begin(ctx, query.TxSettings(txc))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return &transaction{
		conn: c,
		tx:   nativeTx,
	}, nil
}

func (t *transaction) Commit(ctx context.Context) (finalErr error) {
	if err := t.tx.CommitTx(ctx); err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (t *transaction) Rollback(ctx context.Context) (finalErr error) {
	err := t.tx.Rollback(ctx)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return err
}
