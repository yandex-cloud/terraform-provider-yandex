package common

import (
	"context"
	"database/sql/driver"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
)

type Tx interface {
	ID() string

	Exec(ctx context.Context, sql string, params *params.Params) (driver.Result, error)
	Query(ctx context.Context, sql string, params *params.Params) (driver.RowsNextResultSet, error)

	Rollback(ctx context.Context) error
	Commit(ctx context.Context) error
}
