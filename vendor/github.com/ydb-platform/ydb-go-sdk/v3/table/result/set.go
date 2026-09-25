package result

import "github.com/ydb-platform/ydb-go-sdk/v3/table/options"

type Set interface {
	// ColumnCount returns number of columns in the current result set.
	ColumnCount() int

	// Columns allows to iterate over all columns of the current result set.
	Columns(it func(options.Column))

	// RowCount returns number of rows in the result set.
	RowCount() int

	// ItemCount returns number of items in the current row.
	ItemCount() int

	// Truncated returns true if current result set has been truncated by server
	Truncated() bool
}
