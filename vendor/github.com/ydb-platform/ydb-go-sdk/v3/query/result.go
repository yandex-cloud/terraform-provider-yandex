package query

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/scanner"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
)

type (
	Result            = result.Result
	ResultSet         = result.Set
	ClosableResultSet = result.ClosableResultSet
	Row               = result.Row
	Type              = types.Type
	NamedDestination  = scanner.NamedDestination
	ScanStructOption  = scanner.ScanStructOption
)

func Named(columnName string, destinationValueReference interface{}) (dst NamedDestination) {
	return scanner.NamedRef(columnName, destinationValueReference)
}

func WithScanStructTagName(name string) ScanStructOption {
	return scanner.WithTagName(name)
}

func WithScanStructAllowMissingColumnsFromSelect() ScanStructOption {
	return scanner.WithAllowMissingColumnsFromSelect()
}

func WithScanStructAllowMissingFieldsInStruct() ScanStructOption {
	return scanner.WithAllowMissingFieldsInStruct()
}
