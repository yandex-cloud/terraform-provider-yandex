package xsql

import (
	"database/sql"
	"database/sql/driver"
	"io"
)

type singleRow struct {
	values  []sql.NamedArg
	readAll bool
}

func rowByAstPlan(ast, plan string) *singleRow {
	return &singleRow{
		values: []sql.NamedArg{
			{
				Name:  "Ast",
				Value: ast,
			},
			{
				Name:  "Plan",
				Value: plan,
			},
		},
	}
}

func (r *singleRow) Columns() (columns []string) {
	for i := range r.values {
		columns = append(columns, r.values[i].Name)
	}

	return columns
}

func (r *singleRow) Close() error {
	return nil
}

func (r *singleRow) Next(dst []driver.Value) error {
	if r.values == nil || r.readAll {
		return io.EOF
	}
	for i := range r.values {
		dst[i] = r.values[i].Value
	}
	r.readAll = true

	return nil
}
