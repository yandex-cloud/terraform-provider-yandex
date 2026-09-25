package xtable

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/badconn"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/common"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
)

var _ common.Tx = (*transaction)(nil)

type transaction struct {
	conn *Conn
	tx   table.Transaction
}

func (tx *transaction) ID() string {
	if tx.tx == nil {
		return ""
	}

	return tx.tx.ID()
}

func (tx *transaction) Exec(ctx context.Context, sql string, params *params.Params) (driver.Result, error) {
	m := queryModeFromContext(ctx, tx.conn.defaultQueryMode)
	if m != DataQueryMode {
		return nil, xerrors.WithStackTrace(fmt.Errorf("%q: %w", m.String(), ErrWrongQueryMode))
	}
	_, err := tx.tx.Execute(ctx, sql, params, tx.conn.dataOpts...)
	if err != nil {
		return nil, badconn.Map(xerrors.WithStackTrace(err))
	}

	return resultNoRows{}, nil
}

func (tx *transaction) Query(ctx context.Context, sql string, params *params.Params) (driver.RowsNextResultSet, error) {
	m := queryModeFromContext(ctx, tx.conn.defaultQueryMode)
	if m != DataQueryMode {
		return nil, xerrors.WithStackTrace(
			fmt.Errorf("%s: %w", m.String(), ErrWrongQueryMode),
		)
	}
	res, err := tx.tx.Execute(ctx,
		sql, params, tx.conn.dataOpts...,
	)
	if err != nil {
		return nil, badconn.Map(xerrors.WithStackTrace(err))
	}

	if err = res.Err(); err != nil {
		return nil, badconn.Map(xerrors.WithStackTrace(err))
	}

	return &rows{
		conn:   tx.conn,
		result: res,
	}, nil
}

func (tx *transaction) Rollback(ctx context.Context) error {
	err := tx.tx.Rollback(ctx)
	if err != nil {
		return badconn.Map(xerrors.WithStackTrace(err))
	}

	return err
}

func beginTx(ctx context.Context, c *Conn, txOptions driver.TxOptions) (common.Tx, error) {
	txc, err := toYDB(txOptions)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}
	nativeTx, err := c.session.BeginTransaction(ctx, table.TxSettings(txc))
	if err != nil {
		return nil, badconn.Map(xerrors.WithStackTrace(err))
	}

	return &transaction{
		conn: c,
		tx:   nativeTx,
	}, nil
}

func (tx *transaction) Commit(ctx context.Context) (finalErr error) {
	if _, err := tx.tx.CommitTx(ctx); err != nil {
		return badconn.Map(xerrors.WithStackTrace(err))
	}

	return nil
}
