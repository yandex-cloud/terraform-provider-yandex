package types

import (
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/decimal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

type Value = value.Value

func BoolValue(v bool) Value { return value.BoolValue(v) }

func Int8Value(v int8) Value { return value.Int8Value(v) }

func Uint8Value(v uint8) Value { return value.Uint8Value(v) }

func Int16Value(v int16) Value { return value.Int16Value(v) }

func Uint16Value(v uint16) Value { return value.Uint16Value(v) }

func Int32Value(v int32) Value { return value.Int32Value(v) }

func Uint32Value(v uint32) Value { return value.Uint32Value(v) }

func Int64Value(v int64) Value { return value.Int64Value(v) }

func Uint64Value(v uint64) Value { return value.Uint64Value(v) }

func FloatValue(v float32) Value { return value.FloatValue(v) }

func DoubleValue(v float64) Value { return value.DoubleValue(v) }

// DateValue returns ydb date value by given days since Epoch
func DateValue(v uint32) Value { return value.DateValue(v) }

// DatetimeValue makes ydb datetime value from seconds since Epoch
func DatetimeValue(v uint32) Value { return value.DatetimeValue(v) }

// TimestampValue makes ydb timestamp value from microseconds since Epoch
func TimestampValue(v uint64) Value { return value.TimestampValue(v) }

// IntervalValueFromMicroseconds makes Value from given microseconds value
func IntervalValueFromMicroseconds(v int64) Value { return value.IntervalValue(v) }

// IntervalValue makes Value from given microseconds value
//
// Deprecated: use IntervalValueFromMicroseconds instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func IntervalValue(v int64) Value { return value.IntervalValue(v) }

// TzDateValue makes TzDate value from string
func TzDateValue(v string) Value { return value.TzDateValue(v) }

// TzDatetimeValue makes TzDatetime value from string
func TzDatetimeValue(v string) Value { return value.TzDatetimeValue(v) }

// TzTimestampValue makes TzTimestamp value from string
func TzTimestampValue(v string) Value { return value.TzTimestampValue(v) }

// DateValueFromTime makes Date value from time.Time
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func DateValueFromTime(t time.Time) Value {
	return value.DateValueFromTime(t)
}

// DatetimeValueFromTime makes Datetime value from time.Time
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func DatetimeValueFromTime(t time.Time) Value {
	return value.DatetimeValueFromTime(t)
}

// TimestampValueFromTime makes Timestamp value from time.Time
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func TimestampValueFromTime(t time.Time) Value {
	return value.TimestampValueFromTime(t)
}

// IntervalValueFromDuration makes Interval value from time.Duration
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func IntervalValueFromDuration(v time.Duration) Value {
	return value.IntervalValueFromDuration(v)
}

// TzDateValueFromTime makes TzDate value from time.Time
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func TzDateValueFromTime(t time.Time) Value {
	return value.TzDateValueFromTime(t)
}

// TzDatetimeValueFromTime makes TzDatetime value from time.Time
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func TzDatetimeValueFromTime(t time.Time) Value {
	return value.TzDatetimeValueFromTime(t)
}

// TzTimestampValueFromTime makes TzTimestamp value from time.Time
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func TzTimestampValueFromTime(t time.Time) Value {
	return value.TzTimestampValueFromTime(t)
}

// StringValue returns bytes value
//
// Deprecated: use BytesValue instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func StringValue(v []byte) Value { return value.BytesValue(v) }

func BytesValue(v []byte) Value { return value.BytesValue(v) }

func BytesValueFromString(v string) Value { return value.BytesValue(xstring.ToBytes(v)) }

// StringValueFromString makes String value from string
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func StringValueFromString(v string) Value { return value.BytesValue(xstring.ToBytes(v)) }

func UTF8Value(v string) Value { return value.TextValue(v) }

func TextValue(v string) Value { return value.TextValue(v) }

func YSONValue(v string) Value { return value.YSONValue(xstring.ToBytes(v)) }

// YSONValueFromBytes makes YSON value from bytes
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func YSONValueFromBytes(v []byte) Value { return value.YSONValue(v) }

func JSONValue(v string) Value { return value.JSONValue(v) }

// JSONValueFromBytes makes JSON value from bytes
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func JSONValueFromBytes(v []byte) Value { return value.JSONValue(xstring.FromBytes(v)) }

// removed for https://github.com/ydb-platform/ydb-go-sdk/issues/1501
// func UUIDValue(v [16]byte) Value { return UUIDWithIssue1501Value(v) }

// UUIDBytesWithIssue1501Type is type wrapper for scan expected values for values stored with bug
// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
type UUIDBytesWithIssue1501Type = value.UUIDIssue1501FixedBytesWrapper

func NewUUIDBytesWithIssue1501(val [16]byte) UUIDBytesWithIssue1501Type {
	return value.NewUUIDIssue1501FixedBytesWrapper(val)
}

// UUIDWithIssue1501Value is function for save uuid with old corrupted data format for save old behavior
// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//
// Use UuidValue for all new code
func UUIDWithIssue1501Value(v [16]byte) Value {
	return value.UUIDWithIssue1501Value(v)
}

func UuidValue(v uuid.UUID) Value { //nolint:revive
	return value.Uuid(v)
}

func JSONDocumentValue(v string) Value { return value.JSONDocumentValue(v) }

// JSONDocumentValueFromBytes makes JSONDocument value from bytes
//
// Warning: all *From* helpers will be removed at next major release
// (functional will be implements with go1.18 type lists)
func JSONDocumentValueFromBytes(v []byte) Value {
	return value.JSONDocumentValue(xstring.FromBytes(v))
}

func DyNumberValue(v string) Value { return value.DyNumberValue(v) }

func VoidValue() Value { return value.VoidValue() }

func NullValue(t Type) Value { return value.NullValue(t) }

func ZeroValue(t Type) Value { return value.ZeroValue(t) }

func OptionalValue(v Value) Value { return value.OptionalValue(v) }

// Decimal supported in scanner API
type Decimal = decimal.Decimal

// DecimalValue creates decimal value of given types t and value v.
// Note that Decimal.Bytes interpreted as big-endian int128.
func DecimalValue(v *Decimal) Value {
	return value.DecimalValue(v.Bytes, v.Precision, v.Scale)
}

func DecimalValueFromBigInt(v *big.Int, precision, scale uint32) Value {
	return value.DecimalValueFromBigInt(v, precision, scale)
}

func DecimalValueFromString(str string, precision, scale uint32) (Value, error) {
	return value.DecimalValueFromString(str, precision, scale)
}

func TupleValue(vs ...Value) Value {
	return value.TupleValue(vs...)
}

func ListValue(vs ...Value) Value {
	return value.ListValue(vs...)
}

func SetValue(vs ...Value) Value {
	return value.SetValue(vs...)
}

type structValueFields struct {
	fields []value.StructValueField
}

type StructValueOption func(*structValueFields)

func StructFieldValue(name string, v Value) StructValueOption {
	return func(t *structValueFields) {
		t.fields = append(t.fields, value.StructValueField{Name: name, V: v})
	}
}

func StructValue(opts ...StructValueOption) Value {
	var p structValueFields
	for _, opt := range opts {
		if opt != nil {
			opt(&p)
		}
	}

	return value.StructValue(p.fields...)
}

type dictValueFields struct {
	fields []value.DictValueField
}

type DictValueOption func(*dictValueFields)

func DictFieldValue(k, v Value) DictValueOption {
	return func(t *dictValueFields) {
		t.fields = append(t.fields, value.DictValueField{K: k, V: v})
	}
}

func DictValue(opts ...DictValueOption) Value {
	var p dictValueFields
	for _, opt := range opts {
		if opt != nil {
			opt(&p)
		}
	}

	return value.DictValue(p.fields...)
}

func VariantValueStruct(v Value, name string, variantT Type) Value {
	return value.VariantValueStruct(v, name, variantT)
}

func VariantValueTuple(v Value, i uint32, variantT Type) Value {
	return value.VariantValueTuple(v, i, variantT)
}

func NullableBoolValue(v *bool) Value {
	return value.NullableBoolValue(v)
}

func NullableInt8Value(v *int8) Value {
	return value.NullableInt8Value(v)
}

func NullableInt16Value(v *int16) Value {
	return value.NullableInt16Value(v)
}

func NullableInt32Value(v *int32) Value {
	return value.NullableInt32Value(v)
}

func NullableInt64Value(v *int64) Value {
	return value.NullableInt64Value(v)
}

func NullableUint8Value(v *uint8) Value {
	return value.NullableUint8Value(v)
}

func NullableUint16Value(v *uint16) Value {
	return value.NullableUint16Value(v)
}

func NullableUint32Value(v *uint32) Value {
	return value.NullableUint32Value(v)
}

func NullableUint64Value(v *uint64) Value {
	return value.NullableUint64Value(v)
}

func NullableFloatValue(v *float32) Value {
	return value.NullableFloatValue(v)
}

func NullableDoubleValue(v *float64) Value {
	return value.NullableDoubleValue(v)
}

func NullableDateValue(v *uint32) Value {
	return value.NullableDateValue(v)
}

func NullableDateValueFromTime(v *time.Time) Value {
	return value.NullableDateValueFromTime(v)
}

func NullableDecimalValue(v *[16]byte, precision, scale uint32) Value {
	return value.NullableDecimalValue(v, precision, scale)
}

func NullableDecimalValueFromBigInt(v *big.Int, precision, scale uint32) Value {
	return value.NullableDecimalValueFromBigInt(v, precision, scale)
}

func NullableDatetimeValue(v *uint32) Value {
	return value.NullableDatetimeValue(v)
}

func NullableDatetimeValueFromTime(v *time.Time) Value {
	return value.NullableDatetimeValueFromTime(v)
}

func NullableTzDateValue(v *string) Value {
	return value.NullableTzDateValue(v)
}

func NullableTzDateValueFromTime(v *time.Time) Value {
	return value.NullableTzDateValueFromTime(v)
}

func NullableTzDatetimeValue(v *string) Value {
	return value.NullableTzDatetimeValue(v)
}

func NullableTzDatetimeValueFromTime(v *time.Time) Value {
	return value.NullableTzDatetimeValueFromTime(v)
}

func NullableTimestampValue(v *uint64) Value {
	return value.NullableTimestampValue(v)
}

func NullableTimestampValueFromTime(v *time.Time) Value {
	return value.NullableTimestampValueFromTime(v)
}

func NullableTzTimestampValue(v *string) Value {
	return value.NullableTzTimestampValue(v)
}

func NullableTzTimestampValueFromTime(v *time.Time) Value {
	return value.NullableTzTimestampValueFromTime(v)
}

// NullableIntervalValue makes Value which maybe nil or valued
//
// Deprecated: use NullableIntervalValueFromMicroseconds instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func NullableIntervalValue(v *int64) Value {
	return value.NullableIntervalValueFromMicroseconds(v)
}

func NullableIntervalValueFromMicroseconds(v *int64) Value {
	return value.NullableIntervalValueFromMicroseconds(v)
}

func NullableIntervalValueFromDuration(v *time.Duration) Value {
	return value.NullableIntervalValueFromDuration(v)
}

// NullableStringValue
//
// Deprecated: use NullableBytesValue instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func NullableStringValue(v *[]byte) Value {
	return value.NullableBytesValue(v)
}

func NullableBytesValue(v *[]byte) Value {
	return value.NullableBytesValue(v)
}

func NullableStringValueFromString(v *string) Value {
	return value.NullableBytesValueFromString(v)
}

func NullableBytesValueFromString(v *string) Value {
	return value.NullableBytesValueFromString(v)
}

func NullableUTF8Value(v *string) Value {
	return value.NullableTextValue(v)
}

func NullableTextValue(v *string) Value {
	return value.NullableTextValue(v)
}

func NullableYSONValue(v *string) Value {
	return value.NullableYSONValue(v)
}

func NullableYSONValueFromBytes(v *[]byte) Value {
	return value.NullableYSONValueFromBytes(v)
}

func NullableJSONValue(v *string) Value {
	return value.NullableJSONValue(v)
}

func NullableJSONValueFromBytes(v *[]byte) Value {
	return value.NullableJSONValueFromBytes(v)
}

func NullableUUIDValue(v *[16]byte) Value {
	return value.NullableUUIDValue(v)
}

func NullableUUIDValueWithIssue1501(v *[16]byte) Value {
	return value.NullableUUIDValueWithIssue1501(v)
}

func NullableUUIDTypedValue(v *uuid.UUID) Value {
	return value.NullableUuidValue(v)
}

func NullableJSONDocumentValue(v *string) Value {
	return value.NullableJSONDocumentValue(v)
}

func NullableJSONDocumentValueFromBytes(v *[]byte) Value {
	return value.NullableJSONDocumentValueFromBytes(v)
}

func NullableDyNumberValue(v *string) Value {
	return value.NullableDyNumberValue(v)
}

func Nullable(t Type, v interface{}) Value {
	return value.Nullable(t, v)
}
