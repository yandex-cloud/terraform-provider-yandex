package params

import (
	"time"

	"github.com/google/uuid"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
)

type (
	variantTuple struct {
		parent *variant

		types []types.Type
		index uint32
		value value.Value
	}

	variantTupleTypes struct {
		tuple *variantTuple
	}

	variantTupleItem struct {
		tuple *variantTuple
	}

	variantTupleBuilder struct {
		tuple *variantTuple
	}
)

func (vt *variantTuple) Types() *variantTupleTypes {
	return &variantTupleTypes{
		tuple: vt,
	}
}

func (vtt *variantTupleTypes) AddTypes(args ...types.Type) *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, args...)

	return vtt
}

func (vtt *variantTupleTypes) Text() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Text)

	return vtt
}

func (vtt *variantTupleTypes) Bytes() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Bytes)

	return vtt
}

func (vtt *variantTupleTypes) Bool() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Bool)

	return vtt
}

func (vtt *variantTupleTypes) Uint64() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Uint64)

	return vtt
}

func (vtt *variantTupleTypes) Int64() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Int64)

	return vtt
}

func (vtt *variantTupleTypes) Uint32() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Uint32)

	return vtt
}

func (vtt *variantTupleTypes) Int32() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Int32)

	return vtt
}

func (vtt *variantTupleTypes) Uint16() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Uint16)

	return vtt
}

func (vtt *variantTupleTypes) Int16() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Int16)

	return vtt
}

func (vtt *variantTupleTypes) Uint8() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Uint8)

	return vtt
}

func (vtt *variantTupleTypes) Int8() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Int8)

	return vtt
}

func (vtt *variantTupleTypes) Float() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Float)

	return vtt
}

func (vtt *variantTupleTypes) Double() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Double)

	return vtt
}

func (vtt *variantTupleTypes) Decimal(precision, scale uint32) *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.NewDecimal(precision, scale))

	return vtt
}

func (vtt *variantTupleTypes) Timestamp() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Timestamp)

	return vtt
}

func (vtt *variantTupleTypes) Date() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Date)

	return vtt
}

func (vtt *variantTupleTypes) Datetime() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Datetime)

	return vtt
}

func (vtt *variantTupleTypes) Interval() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.Interval)

	return vtt
}

func (vtt *variantTupleTypes) JSON() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.JSON)

	return vtt
}

func (vtt *variantTupleTypes) JSONDocument() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.JSONDocument)

	return vtt
}

func (vtt *variantTupleTypes) YSON() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.YSON)

	return vtt
}

//// UUID has data corruption bug and will be removed in next version.
////
//// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
//// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (vtt *variantTupleTypes) UUID() *variantTupleTypes {
//	vtt.tuple.types = append(vtt.tuple.types, types.UUID)
//
//	return vtt
//}

func (vtt *variantTupleTypes) Uuid() *variantTupleTypes { //nolint:revive
	vtt.tuple.types = append(vtt.tuple.types, types.UUID)

	return vtt
}

func (vtt *variantTupleTypes) UUIDWithIssue1501Value() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.UUID)

	return vtt
}

func (vtt *variantTupleTypes) TzDate() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.TzDate)

	return vtt
}

func (vtt *variantTupleTypes) TzTimestamp() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.TzTimestamp)

	return vtt
}

func (vtt *variantTupleTypes) TzDatetime() *variantTupleTypes {
	vtt.tuple.types = append(vtt.tuple.types, types.TzDatetime)

	return vtt
}

func (vtt *variantTupleTypes) Index(i uint32) *variantTupleItem {
	vtt.tuple.index = i

	return &variantTupleItem{
		tuple: vtt.tuple,
	}
}

func (vti *variantTupleItem) Text(v string) *variantTupleBuilder {
	vti.tuple.value = value.TextValue(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Bytes(v []byte) *variantTupleBuilder {
	vti.tuple.value = value.BytesValue(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Bool(v bool) *variantTupleBuilder {
	vti.tuple.value = value.BoolValue(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Uint64(v uint64) *variantTupleBuilder {
	vti.tuple.value = value.Uint64Value(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Int64(v int64) *variantTupleBuilder {
	vti.tuple.value = value.Int64Value(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Uint32(v uint32) *variantTupleBuilder {
	vti.tuple.value = value.Uint32Value(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Int32(v int32) *variantTupleBuilder {
	vti.tuple.value = value.Int32Value(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Uint16(v uint16) *variantTupleBuilder {
	vti.tuple.value = value.Uint16Value(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Int16(v int16) *variantTupleBuilder {
	vti.tuple.value = value.Int16Value(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Uint8(v uint8) *variantTupleBuilder {
	vti.tuple.value = value.Uint8Value(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Int8(v int8) *variantTupleBuilder {
	vti.tuple.value = value.Int8Value(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Float(v float32) *variantTupleBuilder {
	vti.tuple.value = value.FloatValue(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Double(v float64) *variantTupleBuilder {
	vti.tuple.value = value.DoubleValue(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Decimal(v [16]byte, precision, scale uint32) *variantTupleBuilder {
	vti.tuple.value = value.DecimalValue(v, precision, scale)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Timestamp(v time.Time) *variantTupleBuilder {
	vti.tuple.value = value.TimestampValueFromTime(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Date(v time.Time) *variantTupleBuilder {
	vti.tuple.value = value.DateValueFromTime(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Datetime(v time.Time) *variantTupleBuilder {
	vti.tuple.value = value.DatetimeValueFromTime(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) Interval(v time.Duration) *variantTupleBuilder {
	vti.tuple.value = value.IntervalValueFromDuration(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) JSON(v string) *variantTupleBuilder {
	vti.tuple.value = value.JSONValue(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) JSONDocument(v string) *variantTupleBuilder {
	vti.tuple.value = value.JSONDocumentValue(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) YSON(v []byte) *variantTupleBuilder {
	vti.tuple.value = value.YSONValue(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

//// UUID has data corruption bug and will be removed in next version.
////
//// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
//// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (vti *variantTupleItem) UUID(v [16]byte) *variantTupleBuilder {
//	vti.tuple.value = value.UUIDWithIssue1501Value(v)
//
//	return &variantTupleBuilder{
//		tuple: vti.tuple,
//	}
//}

func (vti *variantTupleItem) Uuid(v uuid.UUID) *variantTupleBuilder { //nolint:revive
	vti.tuple.value = value.Uuid(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) UUIDWithIssue1501Value(v [16]byte) *variantTupleBuilder {
	vti.tuple.value = value.UUIDWithIssue1501Value(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) TzDate(v time.Time) *variantTupleBuilder {
	vti.tuple.value = value.TzDateValueFromTime(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) TzTimestamp(v time.Time) *variantTupleBuilder {
	vti.tuple.value = value.TzTimestampValueFromTime(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vti *variantTupleItem) TzDatetime(v time.Time) *variantTupleBuilder {
	vti.tuple.value = value.TzDatetimeValueFromTime(v)

	return &variantTupleBuilder{
		tuple: vti.tuple,
	}
}

func (vtb *variantTupleBuilder) EndTuple() *variantBuilder {
	vtb.tuple.parent.value = value.VariantValueTuple(
		vtb.tuple.value,
		vtb.tuple.index,
		types.NewVariantTuple(vtb.tuple.types...),
	)

	return &variantBuilder{
		variant: vtb.tuple.parent,
	}
}
