package params

import (
	"time"

	"github.com/google/uuid"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
)

type (
	structure struct {
		parent Builder
		name   string
		values []value.StructValueField
	}

	structValue struct {
		parent *structure
		name   string
	}
)

func (s *structure) AddItems(items ...value.StructValueField) *structure {
	s.values = append(s.values, items...)

	return s
}

func (s *structure) Field(v string) *structValue {
	return &structValue{
		parent: s,
		name:   v,
	}
}

func (s *structure) EndStruct() Builder {
	s.parent.params = append(s.parent.params, &Parameter{
		parent: s.parent,
		name:   s.name,
		value:  value.StructValue(s.values...),
	})

	return s.parent
}

func (s *structValue) Text(v string) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.TextValue(v),
	})

	return s.parent
}

func (s *structValue) Bytes(v []byte) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.BytesValue(v),
	})

	return s.parent
}

func (s *structValue) Bool(v bool) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.BoolValue(v),
	})

	return s.parent
}

func (s *structValue) Uint64(v uint64) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.Uint64Value(v),
	})

	return s.parent
}

func (s *structValue) Int64(v int64) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.Int64Value(v),
	})

	return s.parent
}

func (s *structValue) Uint32(v uint32) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.Uint32Value(v),
	})

	return s.parent
}

func (s *structValue) Int32(v int32) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.Int32Value(v),
	})

	return s.parent
}

func (s *structValue) Uint16(v uint16) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.Uint16Value(v),
	})

	return s.parent
}

func (s *structValue) Int16(v int16) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.Int16Value(v),
	})

	return s.parent
}

func (s *structValue) Uint8(v uint8) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.Uint8Value(v),
	})

	return s.parent
}

func (s *structValue) Int8(v int8) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.Int8Value(v),
	})

	return s.parent
}

func (s *structValue) Float(v float32) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.FloatValue(v),
	})

	return s.parent
}

func (s *structValue) Double(v float64) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.DoubleValue(v),
	})

	return s.parent
}

func (s *structValue) Decimal(v [16]byte, precision, scale uint32) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.DecimalValue(v, precision, scale),
	})

	return s.parent
}

func (s *structValue) Timestamp(v time.Time) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.TimestampValueFromTime(v),
	})

	return s.parent
}

func (s *structValue) Date(v time.Time) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.DateValueFromTime(v),
	})

	return s.parent
}

func (s *structValue) Datetime(v time.Time) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.DatetimeValueFromTime(v),
	})

	return s.parent
}

func (s *structValue) Interval(v time.Duration) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.IntervalValueFromDuration(v),
	})

	return s.parent
}

func (s *structValue) JSON(v string) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.JSONValue(v),
	})

	return s.parent
}

func (s *structValue) JSONDocument(v string) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.JSONDocumentValue(v),
	})

	return s.parent
}

func (s *structValue) YSON(v []byte) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.YSONValue(v),
	})

	return s.parent
}

//// UUID has data corruption bug and will be removed in next version.
////
//// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
//// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (s *structValue) UUID(v [16]byte) *structure {
//	s.parent.values = append(s.parent.values, value.StructValueField{
//		Name: s.name,
//		V:    value.UUIDWithIssue1501Value(v),
//	})
//
//	return s.parent
//}

func (s *structValue) Uuid(v uuid.UUID) *structure { //nolint:revive
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.Uuid(v),
	})

	return s.parent
}

func (s *structValue) UUIDWithIssue1501Value(v [16]byte) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.UUIDWithIssue1501Value(v),
	})

	return s.parent
}

func (s *structValue) TzDatetime(v time.Time) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.TzDatetimeValueFromTime(v),
	})

	return s.parent
}

func (s *structValue) TzTimestamp(v time.Time) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.TzTimestampValueFromTime(v),
	})

	return s.parent
}

func (s *structValue) TzDate(v time.Time) *structure {
	s.parent.values = append(s.parent.values, value.StructValueField{
		Name: s.name,
		V:    value.TzDateValueFromTime(v),
	})

	return s.parent
}
