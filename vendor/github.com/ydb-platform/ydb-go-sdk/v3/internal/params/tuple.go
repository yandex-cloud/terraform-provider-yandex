package params

import (
	"time"

	"github.com/google/uuid"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
)

type (
	tuple struct {
		parent Builder
		name   string
		values []value.Value
	}
	tupleItem struct {
		parent *tuple
	}
)

func (t *tuple) Add() *tupleItem {
	return &tupleItem{
		parent: t,
	}
}

func (t *tuple) AddItems(items ...value.Value) *tuple {
	t.values = append(t.values, items...)

	return t
}

func (t *tuple) EndTuple() Builder {
	t.parent.params = append(t.parent.params, &Parameter{
		parent: t.parent,
		name:   t.name,
		value:  value.TupleValue(t.values...),
	})

	return t.parent
}

func (t *tupleItem) Text(v string) *tuple {
	t.parent.values = append(t.parent.values, value.TextValue(v))

	return t.parent
}

func (t *tupleItem) Bytes(v []byte) *tuple {
	t.parent.values = append(t.parent.values, value.BytesValue(v))

	return t.parent
}

func (t *tupleItem) Bool(v bool) *tuple {
	t.parent.values = append(t.parent.values, value.BoolValue(v))

	return t.parent
}

func (t *tupleItem) Uint64(v uint64) *tuple {
	t.parent.values = append(t.parent.values, value.Uint64Value(v))

	return t.parent
}

func (t *tupleItem) Int64(v int64) *tuple {
	t.parent.values = append(t.parent.values, value.Int64Value(v))

	return t.parent
}

func (t *tupleItem) Uint32(v uint32) *tuple {
	t.parent.values = append(t.parent.values, value.Uint32Value(v))

	return t.parent
}

func (t *tupleItem) Int32(v int32) *tuple {
	t.parent.values = append(t.parent.values, value.Int32Value(v))

	return t.parent
}

func (t *tupleItem) Uint16(v uint16) *tuple {
	t.parent.values = append(t.parent.values, value.Uint16Value(v))

	return t.parent
}

func (t *tupleItem) Int16(v int16) *tuple {
	t.parent.values = append(t.parent.values, value.Int16Value(v))

	return t.parent
}

func (t *tupleItem) Uint8(v uint8) *tuple {
	t.parent.values = append(t.parent.values, value.Uint8Value(v))

	return t.parent
}

func (t *tupleItem) Int8(v int8) *tuple {
	t.parent.values = append(t.parent.values, value.Int8Value(v))

	return t.parent
}

func (t *tupleItem) Float(v float32) *tuple {
	t.parent.values = append(t.parent.values, value.FloatValue(v))

	return t.parent
}

func (t *tupleItem) Double(v float64) *tuple {
	t.parent.values = append(t.parent.values, value.DoubleValue(v))

	return t.parent
}

func (t *tupleItem) Decimal(v [16]byte, precision, scale uint32) *tuple {
	t.parent.values = append(t.parent.values, value.DecimalValue(v, precision, scale))

	return t.parent
}

func (t *tupleItem) Timestamp(v time.Time) *tuple {
	t.parent.values = append(t.parent.values, value.TimestampValueFromTime(v))

	return t.parent
}

func (t *tupleItem) Date(v time.Time) *tuple {
	t.parent.values = append(t.parent.values, value.DateValueFromTime(v))

	return t.parent
}

func (t *tupleItem) Datetime(v time.Time) *tuple {
	t.parent.values = append(t.parent.values, value.DatetimeValueFromTime(v))

	return t.parent
}

func (t *tupleItem) Interval(v time.Duration) *tuple {
	t.parent.values = append(t.parent.values, value.IntervalValueFromDuration(v))

	return t.parent
}

func (t *tupleItem) JSON(v string) *tuple {
	t.parent.values = append(t.parent.values, value.JSONValue(v))

	return t.parent
}

func (t *tupleItem) JSONDocument(v string) *tuple {
	t.parent.values = append(t.parent.values, value.JSONDocumentValue(v))

	return t.parent
}

func (t *tupleItem) YSON(v []byte) *tuple {
	t.parent.values = append(t.parent.values, value.YSONValue(v))

	return t.parent
}

//// UUID has data corruption bug and will be removed in next version.
////
//// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
//// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (t *tupleItem) UUID(v [16]byte) *tuple {
//	t.parent.values = append(t.parent.values, value.UUIDWithIssue1501Value(v))
//
//	return t.parent
//}

func (t *tupleItem) Uuid(v uuid.UUID) *tuple { //nolint:revive
	t.parent.values = append(t.parent.values, value.Uuid(v))

	return t.parent
}

func (t *tupleItem) UUIDWithIssue1501Value(v [16]byte) *tuple {
	t.parent.values = append(t.parent.values, value.UUIDWithIssue1501Value(v))

	return t.parent
}

func (t *tupleItem) TzDate(v time.Time) *tuple {
	t.parent.values = append(t.parent.values, value.TzDateValueFromTime(v))

	return t.parent
}

func (t *tupleItem) TzTimestamp(v time.Time) *tuple {
	t.parent.values = append(t.parent.values, value.TzTimestampValueFromTime(v))

	return t.parent
}

func (t *tupleItem) TzDatetime(v time.Time) *tuple {
	t.parent.values = append(t.parent.values, value.TzDatetimeValueFromTime(v))

	return t.parent
}
