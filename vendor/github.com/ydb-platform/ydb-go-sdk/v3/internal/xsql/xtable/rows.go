package xtable

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/scanner"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/badconn"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/indexed"
)

var (
	_ driver.Rows                           = &rows{}
	_ driver.RowsNextResultSet              = &rows{}
	_ driver.RowsColumnTypeDatabaseTypeName = &rows{}
	_ driver.RowsColumnTypeNullable         = &rows{}

	_ scanner.Scanner = &valuer{}

	ignoreColumnPrefixName = "__discard_column_"
)

type rows struct {
	conn   *Conn
	result result.BaseResult

	// nextSet once need for get first result set as default.
	// Iterate over many result sets must be with rows.NextResultSet()
	nextSet sync.Once
}

func (r *rows) LastInsertId() (int64, error) { return 0, ErrUnsupported }
func (r *rows) RowsAffected() (int64, error) { return 0, ErrUnsupported }

func (r *rows) Columns() []string {
	r.nextSet.Do(func() {
		r.result.NextResultSet(context.Background())
	})
	cs := make([]string, 0, r.result.CurrentResultSet().ColumnCount())
	r.result.CurrentResultSet().Columns(func(m options.Column) {
		if !strings.HasPrefix(m.Name, ignoreColumnPrefixName) {
			cs = append(cs, m.Name)
		}
	})

	return cs
}

// TODO: Need to store column types to internal rows cache.
//
//nolint:godox
func (r *rows) ColumnTypeDatabaseTypeName(index int) string {
	r.nextSet.Do(func() {
		r.result.NextResultSet(context.Background())
	})

	var i int
	yqlTypes := make([]string, r.result.CurrentResultSet().ColumnCount())
	r.result.CurrentResultSet().Columns(func(m options.Column) {
		yqlTypes[i] = m.Type.Yql()
		i++
	})

	return yqlTypes[index]
}

// TODO: Need to store column nullables to internal rows cache.
//
//nolint:godox
func (r *rows) ColumnTypeNullable(index int) (nullable, ok bool) {
	r.nextSet.Do(func() {
		r.result.NextResultSet(context.Background())
	})

	var i int
	nullables := make([]bool, r.result.CurrentResultSet().ColumnCount())
	r.result.CurrentResultSet().Columns(func(m options.Column) {
		_, nullables[i] = m.Type.(interface {
			IsOptional()
		})
		i++
	})

	return nullables[index], true
}

func (r *rows) NextResultSet() (finalErr error) {
	r.nextSet.Do(func() {})
	err := r.result.NextResultSetErr(context.Background())
	if err != nil {
		return badconn.Map(xerrors.WithStackTrace(err))
	}

	return nil
}

func (r *rows) HasNextResultSet() bool {
	return r.result.HasNextResultSet()
}

func (r *rows) Next(dst []driver.Value) error {
	var err error
	r.nextSet.Do(func() {
		err = r.result.NextResultSetErr(context.Background())
	})
	if err != nil {
		return badconn.Map(xerrors.WithStackTrace(err))
	}
	if err = r.result.Err(); err != nil {
		return badconn.Map(xerrors.WithStackTrace(err))
	}
	if !r.result.NextRow() {
		return io.EOF
	}
	values := make([]indexed.RequiredOrOptional, len(dst))
	for i := range dst {
		values[i] = &valuer{}
	}
	if err = r.result.Scan(values...); err != nil {
		return badconn.Map(xerrors.WithStackTrace(err))
	}
	for i := range values {
		val, ok := values[i].(*valuer)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to *valuer", val))
		}

		dst[i] = val.Value()
	}
	if err = r.result.Err(); err != nil {
		return badconn.Map(xerrors.WithStackTrace(err))
	}

	return nil
}

func (r *rows) Close() error {
	return r.result.Close()
}
