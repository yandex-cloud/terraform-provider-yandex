package value

import (
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

func IsNull(v Value) bool {
	if v == nil {
		return true
	}

	if opt, has := v.(*optionalValue); has {
		return opt.value == nil
	}

	return false
}

func Unwrap(v Value) Value {
	if v == nil {
		return nil
	}

	if opt, has := v.(*optionalValue); has {
		if opt.value == nil {
			return nil
		}

		return opt.value
	}

	return v
}

func NullableBoolValue(v *bool) Value {
	if v == nil {
		return NullValue(types.Bool)
	}

	return OptionalValue(BoolValue(*v))
}

func NullableInt8Value(v *int8) Value {
	if v == nil {
		return NullValue(types.Int8)
	}

	return OptionalValue(Int8Value(*v))
}

func NullableInt16Value(v *int16) Value {
	if v == nil {
		return NullValue(types.Int16)
	}

	return OptionalValue(Int16Value(*v))
}

func NullableInt32Value(v *int32) Value {
	if v == nil {
		return NullValue(types.Int32)
	}

	return OptionalValue(Int32Value(*v))
}

func NullableInt64Value(v *int64) Value {
	if v == nil {
		return NullValue(types.Int64)
	}

	return OptionalValue(Int64Value(*v))
}

func NullableUint8Value(v *uint8) Value {
	if v == nil {
		return NullValue(types.Uint8)
	}

	return OptionalValue(Uint8Value(*v))
}

func NullableUint16Value(v *uint16) Value {
	if v == nil {
		return NullValue(types.Uint16)
	}

	return OptionalValue(Uint16Value(*v))
}

func NullableUint32Value(v *uint32) Value {
	if v == nil {
		return NullValue(types.Uint32)
	}

	return OptionalValue(Uint32Value(*v))
}

func NullableUint64Value(v *uint64) Value {
	if v == nil {
		return NullValue(types.Uint64)
	}

	return OptionalValue(Uint64Value(*v))
}

func NullableFloatValue(v *float32) Value {
	if v == nil {
		return NullValue(types.Float)
	}

	return OptionalValue(FloatValue(*v))
}

func NullableDoubleValue(v *float64) Value {
	if v == nil {
		return NullValue(types.Double)
	}

	return OptionalValue(DoubleValue(*v))
}

func NullableDateValue(v *uint32) Value {
	if v == nil {
		return NullValue(types.Date)
	}

	return OptionalValue(DateValue(*v))
}

func NullableDateValueFromTime(v *time.Time) Value {
	if v == nil {
		return NullValue(types.Date)
	}

	return OptionalValue(DateValueFromTime(*v))
}

func NullableDatetimeValue(v *uint32) Value {
	if v == nil {
		return NullValue(types.Datetime)
	}

	return OptionalValue(DatetimeValue(*v))
}

func NullableDatetimeValueFromTime(v *time.Time) Value {
	if v == nil {
		return NullValue(types.Datetime)
	}

	return OptionalValue(DatetimeValueFromTime(*v))
}

func NullableDecimalValue(v *[16]byte, precision, scale uint32) Value {
	if v == nil {
		return NullValue(types.NewDecimal(precision, scale))
	}

	return OptionalValue(DecimalValue(*v, precision, scale))
}

func NullableDecimalValueFromBigInt(v *big.Int, precision, scale uint32) Value {
	if v == nil {
		return NullValue(types.NewDecimal(precision, scale))
	}

	return OptionalValue(DecimalValueFromBigInt(v, precision, scale))
}

func NullableTzDateValue(v *string) Value {
	if v == nil {
		return NullValue(types.TzDate)
	}

	return OptionalValue(TzDateValue(*v))
}

func NullableTzDateValueFromTime(v *time.Time) Value {
	if v == nil {
		return NullValue(types.TzDate)
	}

	return OptionalValue(TzDateValueFromTime(*v))
}

func NullableTzDatetimeValue(v *string) Value {
	if v == nil {
		return NullValue(types.TzDatetime)
	}

	return OptionalValue(TzDatetimeValue(*v))
}

func NullableTzDatetimeValueFromTime(v *time.Time) Value {
	if v == nil {
		return NullValue(types.TzDatetime)
	}

	return OptionalValue(TzDatetimeValueFromTime(*v))
}

func NullableTimestampValue(v *uint64) Value {
	if v == nil {
		return NullValue(types.Timestamp)
	}

	return OptionalValue(TimestampValue(*v))
}

func NullableTimestampValueFromTime(v *time.Time) Value {
	if v == nil {
		return NullValue(types.Timestamp)
	}

	return OptionalValue(TimestampValueFromTime(*v))
}

func NullableTzTimestampValue(v *string) Value {
	if v == nil {
		return NullValue(types.TzTimestamp)
	}

	return OptionalValue(TzTimestampValue(*v))
}

func NullableTzTimestampValueFromTime(v *time.Time) Value {
	if v == nil {
		return NullValue(types.TzTimestamp)
	}

	return OptionalValue(TzTimestampValueFromTime(*v))
}

func NullableIntervalValueFromMicroseconds(v *int64) Value {
	if v == nil {
		return NullValue(types.Interval)
	}

	return OptionalValue(IntervalValue(*v))
}

func NullableIntervalValueFromDuration(v *time.Duration) Value {
	if v == nil {
		return NullValue(types.Interval)
	}

	return OptionalValue(IntervalValueFromDuration(*v))
}

func NullableBytesValue(v *[]byte) Value {
	if v == nil {
		return NullValue(types.Bytes)
	}

	return OptionalValue(BytesValue(*v))
}

func NullableBytesValueFromString(v *string) Value {
	if v == nil {
		return NullValue(types.Bytes)
	}

	return OptionalValue(BytesValue(xstring.ToBytes(*v)))
}

func NullableTextValue(v *string) Value {
	if v == nil {
		return NullValue(types.Text)
	}

	return OptionalValue(TextValue(*v))
}

func NullableYSONValue(v *string) Value {
	if v == nil {
		return NullValue(types.YSON)
	}

	return OptionalValue(YSONValue(xstring.ToBytes(*v)))
}

func NullableYSONValueFromBytes(v *[]byte) Value {
	if v == nil {
		return NullValue(types.YSON)
	}

	return OptionalValue(YSONValue(*v))
}

func NullableJSONValue(v *string) Value {
	if v == nil {
		return NullValue(types.JSON)
	}

	return OptionalValue(JSONValue(*v))
}

func NullableJSONValueFromBytes(v *[]byte) Value {
	if v == nil {
		return NullValue(types.JSON)
	}

	return OptionalValue(JSONValue(xstring.FromBytes(*v)))
}

func NullableUUIDValue(v *[16]byte) Value {
	if v == nil {
		return NullValue(types.UUID)
	}

	return OptionalValue(UUIDWithIssue1501Value(*v))
}

func NullableUUIDValueWithIssue1501(v *[16]byte) Value {
	if v == nil {
		return NullValue(types.UUID)
	}

	return OptionalValue(UUIDWithIssue1501Value(*v))
}

func NullableUuidValue(v *uuid.UUID) Value { //nolint:revive
	if v == nil {
		return NullValue(types.UUID)
	}

	return OptionalValue(Uuid(*v))
}

func NullableJSONDocumentValue(v *string) Value {
	if v == nil {
		return NullValue(types.JSONDocument)
	}

	return OptionalValue(JSONDocumentValue(*v))
}

func NullableJSONDocumentValueFromBytes(v *[]byte) Value {
	if v == nil {
		return NullValue(types.JSONDocument)
	}

	return OptionalValue(JSONDocumentValue(xstring.FromBytes(*v)))
}

func NullableDyNumberValue(v *string) Value {
	if v == nil {
		return NullValue(types.DyNumber)
	}

	return OptionalValue(DyNumberValue(*v))
}

// Nullable makes optional value from nullable type
// Warning: type interface will be replaced in the future with typed parameters pattern from go1.18
//
//nolint:gocyclo, funlen
func Nullable(t types.Type, v interface{}) Value {
	switch t {
	case types.Bool:
		tt, ok := v.(*bool)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeBool", tt))
		}

		return NullableBoolValue(tt)
	case types.Int8:
		tt, ok := v.(*int8)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeBool", tt))
		}

		return NullableInt8Value(tt)
	case types.Uint8:
		tt, ok := v.(*uint8)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeUint8", tt))
		}

		return NullableUint8Value(tt)
	case types.Int16:
		tt, ok := v.(*int16)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeInt16", tt))
		}

		return NullableInt16Value(tt)
	case types.Uint16:
		tt, ok := v.(*uint16)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeUint16", tt))
		}

		return NullableUint16Value(tt)
	case types.Int32:
		tt, ok := v.(*int32)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeInt32", tt))
		}

		return NullableInt32Value(tt)
	case types.Uint32:
		tt, ok := v.(*uint32)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to Uint32", tt))
		}

		return NullableUint32Value(tt)
	case types.Int64:
		tt, ok := v.(*int64)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeInt64", tt))
		}

		return NullableInt64Value(tt)
	case types.Uint64:
		tt, ok := v.(*uint64)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeUint64", tt))
		}

		return NullableUint64Value(tt)
	case types.Float:
		tt, ok := v.(*float32)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeFloat", tt))
		}

		return NullableFloatValue(tt)
	case types.Double:
		tt, ok := v.(*float64)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeDouble", tt))
		}

		return NullableDoubleValue(tt)
	case types.Date:
		switch tt := v.(type) {
		case *uint32:
			return NullableDateValue(tt)
		case *time.Time:
			return NullableDateValueFromTime(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeDate", tt))
		}
	case types.Datetime:
		switch tt := v.(type) {
		case *uint32:
			return NullableDatetimeValue(tt)
		case *time.Time:
			return NullableDatetimeValueFromTime(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeDatetime", tt))
		}
	case types.Timestamp:
		switch tt := v.(type) {
		case *uint64:
			return NullableTimestampValue(tt)
		case *time.Time:
			return NullableTimestampValueFromTime(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeTimestamp", tt))
		}
	case types.Interval:
		switch tt := v.(type) {
		case *int64:
			return NullableIntervalValueFromMicroseconds(tt)
		case *time.Duration:
			return NullableIntervalValueFromDuration(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeInterval", tt))
		}
	case types.TzDate:
		switch tt := v.(type) {
		case *string:
			return NullableTzDateValue(tt)
		case *time.Time:
			return NullableTzDateValueFromTime(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeTzDate", tt))
		}
	case types.TzDatetime:
		switch tt := v.(type) {
		case *string:
			return NullableTzDatetimeValue(tt)
		case *time.Time:
			return NullableTzDatetimeValueFromTime(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeTzDatetime", tt))
		}
	case types.TzTimestamp:
		switch tt := v.(type) {
		case *string:
			return NullableTzTimestampValue(tt)
		case *time.Time:
			return NullableTzTimestampValueFromTime(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeTzTimestamp", tt))
		}
	case types.Bytes:
		switch tt := v.(type) {
		case *[]byte:
			return NullableBytesValue(tt)
		case *string:
			return NullableBytesValueFromString(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeBytes", tt))
		}
	case types.Text:
		switch tt := v.(type) {
		case *string:
			return NullableTextValue(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeText", tt))
		}
	case types.YSON:
		switch tt := v.(type) {
		case *string:
			return NullableYSONValue(tt)
		case *[]byte:
			return NullableYSONValueFromBytes(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeYSON", tt))
		}
	case types.JSON:
		switch tt := v.(type) {
		case *string:
			return NullableJSONValue(tt)
		case *[]byte:
			return NullableJSONValueFromBytes(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeJSON", tt))
		}
	case types.UUID:
		switch tt := v.(type) {
		case *[16]byte:
			return NullableUUIDValue(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeUUID", tt))
		}
	case types.JSONDocument:
		switch tt := v.(type) {
		case *string:
			return NullableJSONDocumentValue(tt)
		case *[]byte:
			return NullableJSONDocumentValueFromBytes(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeJSONDocument", tt))
		}
	case types.DyNumber:
		switch tt := v.(type) {
		case *string:
			return NullableDyNumberValue(tt)
		default:
			panic(fmt.Sprintf("unsupported type conversion from %T to TypeDyNumber", tt))
		}
	default:
		panic(fmt.Sprintf("unsupported type: %T", t))
	}
}
