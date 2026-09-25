package common

import (
	"context"
	"database/sql/driver"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
)

type Conn interface {
	driver.Validator
	driver.Pinger

	ID() string
	NodeID() uint32

	Exec(ctx context.Context, sql string, params *params.Params) (result driver.Result, err error)
	Query(ctx context.Context, sql string, params *params.Params) (result driver.RowsNextResultSet, err error)
	Explain(ctx context.Context, sql string, params *params.Params) (ast string, plan string, err error)
	BeginTx(ctx context.Context, opts driver.TxOptions) (Tx, error)

	Close() error
}
