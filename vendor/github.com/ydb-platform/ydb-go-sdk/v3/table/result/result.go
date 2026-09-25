package result

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/indexed"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/stats"
)

// BaseResult is a result of a query.
//
// Use NextResultSet(), NextRow() and Scan() to advance through the result sets,
// its rows and row's items.
//
//	res, err := s.Execute(ctx, txc, "SELECT ...")
//	defer res.Close()
//	for res.NextResultSet(ctx) {
//	    for res.NextRow() {
//	        var id int64
//	        var name *string //optional value
//	        res.Scan(&id,&name)
//	    }
//	}
//	if err := res.err() { // get any error encountered during iteration
//	    // handle error
//	}
//
// If current value under scan
// is not requested types, then res.err() become non-nil.
// After that, NextResultSet(), NextRow() will return false.
type BaseResult interface {
	// HasNextResultSet reports whether result set may be advanced.
	// It may be useful to call HasNextResultSet() instead of NextResultSet() to look ahead
	// without advancing the result set.
	// Note that it does not work with sets from stream.
	HasNextResultSet() bool

	// NextResultSet selects next result set in the result.
	// columns - names of columns in the resultSet that will be scanned
	// It returns false if there are no more result sets.
	// Stream sets are supported.
	// After iterate over result sets should be checked Err()
	NextResultSet(ctx context.Context, columns ...string) bool

	// NextResultSetErr selects next result set in the result.
	// columns - names of columns in the result set that will be scanned
	// It returns:
	// - nil if select next result set successful
	// - io.EOF if no result sets
	// - some error if an error has occurred
	// NextResultSetErr func equal to sequential calls HasNextResultSet() and Err() after
	NextResultSetErr(ctx context.Context, columns ...string) error

	// CurrentResultSet get current result set to use ColumnCount(), RowCount() and other methods
	CurrentResultSet() Set

	// HasNextRow reports whether result row may be advanced.
	// It may be useful to call HasNextRow() instead of NextRow() to look ahead
	// without advancing the result rows.
	HasNextRow() bool

	// NextRow selects next row in the current result set.
	// It returns false if there are no more rows in the result set.
	NextRow() bool

	// ScanWithDefaults scan with default types values.
	// Nil values applied as default value types
	// Input params - pointers to types.
	// If some value implements ydb.table.types.Scanner then will be called
	// value.(ydb.table.types.Scanner).UnmarshalYDB(raw) where raw may be null.
	// In this case client-side implementation UnmarshalYDB must check raw.IsNull() and
	// applied default value or nothing to do
	ScanWithDefaults(values ...indexed.Required) error

	// Scan values.
	// Input params - pointers to types:
	//   bool
	//   int8
	//   uint8
	//   int16
	//   uint16
	//   int32
	//   uint32
	//   int64
	//   uint64
	//   float32
	//   float64
	//   []byte
	//   [16]byte
	//   string
	//   time.Time
	//   time.Duration
	//   ydb.valueType
	// For custom types implement sql.Scanner or json.Unmarshaler interface.
	// For optional types use double pointer construction.
	// For unknown types use interface types.
	// Supported scanning byte arrays of various length.
	// For complex yql types: Dict, List, Tuple and own specific scanning logic
	// implement ydb.table.types.Scanner with UnmarshalYDB method
	// See examples for more detailed information.
	// Output param - Scanner error
	Scan(values ...indexed.RequiredOrOptional) error

	// ScanNamed scans row with column names defined in namedValues
	ScanNamed(namedValues ...named.Value) error

	// Stats returns query execution QueryStats.
	//
	// If query result have no stats - returns nil
	Stats() (s stats.QueryStats)

	// Err return scanner error
	// To handle errors, do not need to check after scanning each row
	// It is enough to check after reading all Set
	Err() error

	// Close closes the Result, preventing further iteration.
	Close() error
}

type Result interface {
	BaseResult

	// ResultSetCount returns number of result sets.
	// Note that it does not work if r is the BaseResult of streaming operation.
	ResultSetCount() int
}

type StreamResult interface {
	BaseResult
}
