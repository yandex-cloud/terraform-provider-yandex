package types

import (
	"bytes"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/scanner"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
)

const (
	decimalPrecision uint32 = 22
	decimalScale     uint32 = 9
)

// Type describes YDB data types.
type Type = types.Type

// Equal checks for type equivalence
func Equal(lhs, rhs Type) bool {
	return types.Equal(lhs, rhs)
}

func List(t Type) Type {
	return types.NewList(t)
}

func Tuple(elems ...Type) Type {
	return types.NewTuple(elems...)
}

type tStructType struct {
	fields []types.StructField
}

type StructOption func(*tStructType)

func StructField(name string, t Type) StructOption {
	return func(s *tStructType) {
		s.fields = append(s.fields, types.StructField{
			Name: name,
			T:    t,
		})
	}
}

func Struct(opts ...StructOption) Type {
	var s tStructType
	for _, opt := range opts {
		if opt != nil {
			opt(&s)
		}
	}

	return types.NewStruct(s.fields...)
}

func Dict(k, v Type) Type {
	return types.NewDict(k, v)
}

func VariantStruct(opts ...StructOption) Type {
	var s tStructType
	for _, opt := range opts {
		if opt != nil {
			opt(&s)
		}
	}

	return types.NewVariantStruct(s.fields...)
}

func VariantTuple(elems ...Type) Type {
	return types.NewVariantTuple(elems...)
}

func Void() Type {
	return types.NewVoid()
}

func Optional(t Type) Type {
	return types.NewOptional(t)
}

var DefaultDecimal = DecimalType(decimalPrecision, decimalScale)

func DecimalType(precision, scale uint32) Type {
	return types.NewDecimal(precision, scale)
}

func DecimalTypeFromDecimal(d *Decimal) Type {
	return types.NewDecimal(d.Precision, d.Scale)
}

// Primitive types known by YDB.
const (
	TypeUnknown      = types.Unknown
	TypeBool         = types.Bool
	TypeInt8         = types.Int8
	TypeUint8        = types.Uint8
	TypeInt16        = types.Int16
	TypeUint16       = types.Uint16
	TypeInt32        = types.Int32
	TypeUint32       = types.Uint32
	TypeInt64        = types.Int64
	TypeUint64       = types.Uint64
	TypeFloat        = types.Float
	TypeDouble       = types.Double
	TypeDate         = types.Date
	TypeDatetime     = types.Datetime
	TypeTimestamp    = types.Timestamp
	TypeInterval     = types.Interval
	TypeTzDate       = types.TzDate
	TypeTzDatetime   = types.TzDatetime
	TypeTzTimestamp  = types.TzTimestamp
	TypeString       = types.Bytes
	TypeBytes        = types.Bytes
	TypeUTF8         = types.Text
	TypeText         = types.Text
	TypeYSON         = types.YSON
	TypeJSON         = types.JSON
	TypeUUID         = types.UUID
	TypeJSONDocument = types.JSONDocument
	TypeDyNumber     = types.DyNumber
)

// WriteTypeStringTo writes ydb type string representation into buffer
//
// Deprecated: use types.Type.Yql() instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WriteTypeStringTo(buf *bytes.Buffer, t Type) { //nolint: interfacer
	buf.WriteString(t.Yql())
}

type (
	RawValue = scanner.RawValue
	Scanner  = scanner.Scanner
)
