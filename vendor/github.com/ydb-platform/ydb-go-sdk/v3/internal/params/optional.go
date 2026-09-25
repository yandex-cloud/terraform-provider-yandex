package params

import (
	"time"

	"github.com/google/uuid"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
)

type (
	optional struct {
		parent Builder
		name   string
		value  value.Value
	}
	optionalBuilder struct {
		opt *optional
	}
)

func (b *optionalBuilder) EndOptional() Builder {
	b.opt.parent.params = append(b.opt.parent.params, &Parameter{
		parent: b.opt.parent,
		name:   b.opt.name,
		value:  b.opt.value,
	})

	return b.opt.parent
}

func (p *optional) Text(v *string) *optionalBuilder {
	p.value = value.NullableTextValue(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Bytes(v *[]byte) *optionalBuilder {
	p.value = value.NullableBytesValue(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Bool(v *bool) *optionalBuilder {
	p.value = value.NullableBoolValue(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Uint64(v *uint64) *optionalBuilder {
	p.value = value.NullableUint64Value(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Int64(v *int64) *optionalBuilder {
	p.value = value.NullableInt64Value(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Uint32(v *uint32) *optionalBuilder {
	p.value = value.NullableUint32Value(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Int32(v *int32) *optionalBuilder {
	p.value = value.NullableInt32Value(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Uint16(v *uint16) *optionalBuilder {
	p.value = value.NullableUint16Value(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Int16(v *int16) *optionalBuilder {
	p.value = value.NullableInt16Value(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Uint8(v *uint8) *optionalBuilder {
	p.value = value.NullableUint8Value(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Int8(v *int8) *optionalBuilder {
	p.value = value.NullableInt8Value(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Float(v *float32) *optionalBuilder {
	p.value = value.NullableFloatValue(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Double(v *float64) *optionalBuilder {
	p.value = value.NullableDoubleValue(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Decimal(v *[16]byte, precision, scale uint32) *optionalBuilder {
	p.value = value.NullableDecimalValue(v, precision, scale)

	return &optionalBuilder{opt: p}
}

func (p *optional) Timestamp(v *time.Time) *optionalBuilder {
	p.value = value.NullableTimestampValueFromTime(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Date(v *time.Time) *optionalBuilder {
	p.value = value.NullableDateValueFromTime(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Datetime(v *time.Time) *optionalBuilder {
	p.value = value.NullableDatetimeValueFromTime(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) Interval(v *time.Duration) *optionalBuilder {
	p.value = value.NullableIntervalValueFromDuration(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) JSON(v *string) *optionalBuilder {
	p.value = value.NullableJSONValue(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) JSONDocument(v *string) *optionalBuilder {
	p.value = value.NullableJSONDocumentValue(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) YSON(v *[]byte) *optionalBuilder {
	p.value = value.NullableYSONValueFromBytes(v)

	return &optionalBuilder{opt: p}
}

//// UUID has data corruption bug and will be removed in next version.
////
//// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
//// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (p *optional) UUID(v *[16]byte) *optionalBuilder {
//	p.value = value.NullableUUIDValue(v)
//
//	return &optionalBuilder{opt: p}
//}

func (p *optional) Uuid(v *uuid.UUID) *optionalBuilder { //nolint:revive
	p.value = value.NullableUuidValue(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) UUIDWithIssue1501Value(v *[16]byte) *optionalBuilder {
	p.value = value.NullableUUIDValue(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) TzDate(v *time.Time) *optionalBuilder {
	p.value = value.NullableTzDateValueFromTime(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) TzTimestamp(v *time.Time) *optionalBuilder {
	p.value = value.NullableTzTimestampValueFromTime(v)

	return &optionalBuilder{opt: p}
}

func (p *optional) TzDatetime(v *time.Time) *optionalBuilder {
	p.value = value.NullableTzDatetimeValueFromTime(v)

	return &optionalBuilder{opt: p}
}
