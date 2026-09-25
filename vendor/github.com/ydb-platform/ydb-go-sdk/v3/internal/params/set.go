package params

import (
	"time"

	"github.com/google/uuid"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
)

type (
	set struct {
		parent Builder
		name   string
		values []value.Value
	}

	setItem struct {
		parent *set
	}
)

func (s *set) Add() *setItem {
	return &setItem{
		parent: s,
	}
}

func (s *set) AddItems(items ...value.Value) *set {
	s.values = append(s.values, items...)

	return s
}

func (s *set) EndSet() Builder {
	s.parent.params = append(s.parent.params, &Parameter{
		parent: s.parent,
		name:   s.name,
		value:  value.SetValue(s.values...),
	})

	return s.parent
}

func (s *setItem) Text(v string) *set {
	s.parent.values = append(s.parent.values, value.TextValue(v))

	return s.parent
}

func (s *setItem) Bytes(v []byte) *set {
	s.parent.values = append(s.parent.values, value.BytesValue(v))

	return s.parent
}

func (s *setItem) Bool(v bool) *set {
	s.parent.values = append(s.parent.values, value.BoolValue(v))

	return s.parent
}

func (s *setItem) Uint64(v uint64) *set {
	s.parent.values = append(s.parent.values, value.Uint64Value(v))

	return s.parent
}

func (s *setItem) Int64(v int64) *set {
	s.parent.values = append(s.parent.values, value.Int64Value(v))

	return s.parent
}

func (s *setItem) Uint32(v uint32) *set {
	s.parent.values = append(s.parent.values, value.Uint32Value(v))

	return s.parent
}

func (s *setItem) Int32(v int32) *set {
	s.parent.values = append(s.parent.values, value.Int32Value(v))

	return s.parent
}

func (s *setItem) Uint16(v uint16) *set {
	s.parent.values = append(s.parent.values, value.Uint16Value(v))

	return s.parent
}

func (s *setItem) Int16(v int16) *set {
	s.parent.values = append(s.parent.values, value.Int16Value(v))

	return s.parent
}

func (s *setItem) Uint8(v uint8) *set {
	s.parent.values = append(s.parent.values, value.Uint8Value(v))

	return s.parent
}

func (s *setItem) Int8(v int8) *set {
	s.parent.values = append(s.parent.values, value.Int8Value(v))

	return s.parent
}

func (s *setItem) Float(v float32) *set {
	s.parent.values = append(s.parent.values, value.FloatValue(v))

	return s.parent
}

func (s *setItem) Double(v float64) *set {
	s.parent.values = append(s.parent.values, value.DoubleValue(v))

	return s.parent
}

func (s *setItem) Decimal(v [16]byte, precision, scale uint32) *set {
	s.parent.values = append(s.parent.values, value.DecimalValue(v, precision, scale))

	return s.parent
}

func (s *setItem) Timestamp(v time.Time) *set {
	s.parent.values = append(s.parent.values, value.TimestampValueFromTime(v))

	return s.parent
}

func (s *setItem) Date(v time.Time) *set {
	s.parent.values = append(s.parent.values, value.DateValueFromTime(v))

	return s.parent
}

func (s *setItem) Datetime(v time.Time) *set {
	s.parent.values = append(s.parent.values, value.DatetimeValueFromTime(v))

	return s.parent
}

func (s *setItem) Interval(v time.Duration) *set {
	s.parent.values = append(s.parent.values, value.IntervalValueFromDuration(v))

	return s.parent
}

func (s *setItem) JSON(v string) *set {
	s.parent.values = append(s.parent.values, value.JSONValue(v))

	return s.parent
}

func (s *setItem) JSONDocument(v string) *set {
	s.parent.values = append(s.parent.values, value.JSONDocumentValue(v))

	return s.parent
}

func (s *setItem) YSON(v []byte) *set {
	s.parent.values = append(s.parent.values, value.YSONValue(v))

	return s.parent
}

//// UUID has data corruption bug and will be removed in next version.
////
//// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
//// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (s *setItem) UUID(v [16]byte) *set {
//	s.parent.values = append(s.parent.values, value.UUIDWithIssue1501Value(v))
//
//	return s.parent
//}

func (s *setItem) Uuid(v uuid.UUID) *set { //nolint:revive
	s.parent.values = append(s.parent.values, value.Uuid(v))

	return s.parent
}

func (s *setItem) UUIDWithIssue1501Value(v [16]byte) *set {
	s.parent.values = append(s.parent.values, value.UUIDWithIssue1501Value(v))

	return s.parent
}

func (s *setItem) TzDate(v time.Time) *set {
	s.parent.values = append(s.parent.values, value.TzDateValueFromTime(v))

	return s.parent
}

func (s *setItem) TzTimestamp(v time.Time) *set {
	s.parent.values = append(s.parent.values, value.TzTimestampValueFromTime(v))

	return s.parent
}

func (s *setItem) TzDatetime(v time.Time) *set {
	s.parent.values = append(s.parent.values, value.TzDatetimeValueFromTime(v))

	return s.parent
}
