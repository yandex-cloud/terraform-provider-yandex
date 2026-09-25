package params

import (
	"time"

	"github.com/google/uuid"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
)

type (
	dict struct {
		parent Builder
		name   string
		values []value.DictValueField
	}
	dictPair struct {
		parent   *dict
		keyValue value.Value
	}
	dictValue struct {
		pair *dictPair
	}
)

func (d *dict) Add() *dictPair {
	return &dictPair{
		parent: d,
	}
}

func (d *dict) AddPairs(pairs ...value.DictValueField) *dict {
	d.values = append(d.values, pairs...)

	return d
}

func (d *dictPair) Text(v string) *dictValue {
	d.keyValue = value.TextValue(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Bytes(v []byte) *dictValue {
	d.keyValue = value.BytesValue(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Bool(v bool) *dictValue {
	d.keyValue = value.BoolValue(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Uint64(v uint64) *dictValue {
	d.keyValue = value.Uint64Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Int64(v int64) *dictValue {
	d.keyValue = value.Int64Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Uint32(v uint32) *dictValue {
	d.keyValue = value.Uint32Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Int32(v int32) *dictValue {
	d.keyValue = value.Int32Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Uint16(v uint16) *dictValue {
	d.keyValue = value.Uint16Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Int16(v int16) *dictValue {
	d.keyValue = value.Int16Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Uint8(v uint8) *dictValue {
	d.keyValue = value.Uint8Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Int8(v int8) *dictValue {
	d.keyValue = value.Int8Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Float(v float32) *dictValue {
	d.keyValue = value.FloatValue(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Double(v float64) *dictValue {
	d.keyValue = value.DoubleValue(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Decimal(v [16]byte, precision, scale uint32) *dictValue {
	d.keyValue = value.DecimalValue(v, precision, scale)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Timestamp(v time.Time) *dictValue {
	d.keyValue = value.TimestampValueFromTime(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Date(v time.Time) *dictValue {
	d.keyValue = value.DateValueFromTime(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Datetime(v time.Time) *dictValue {
	d.keyValue = value.DatetimeValueFromTime(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Interval(v time.Duration) *dictValue {
	d.keyValue = value.IntervalValueFromDuration(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) JSON(v string) *dictValue {
	d.keyValue = value.JSONValue(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) JSONDocument(v string) *dictValue {
	d.keyValue = value.JSONDocumentValue(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) YSON(v []byte) *dictValue {
	d.keyValue = value.YSONValue(v)

	return &dictValue{
		pair: d,
	}
}

// UUID has data corruption bug and will be removed in next version.
//
// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
func (d *dictPair) UUID(v [16]byte) *dictValue {
	d.keyValue = value.UUIDWithIssue1501Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) Uuid(v uuid.UUID) *dictValue { //nolint:revive
	d.keyValue = value.Uuid(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) UUIDWithIssue1501Value(v [16]byte) *dictValue {
	d.keyValue = value.UUIDWithIssue1501Value(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) TzDate(v time.Time) *dictValue {
	d.keyValue = value.TzDateValueFromTime(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) TzTimestamp(v time.Time) *dictValue {
	d.keyValue = value.TzTimestampValueFromTime(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictPair) TzDatetime(v time.Time) *dictValue {
	d.keyValue = value.TzDatetimeValueFromTime(v)

	return &dictValue{
		pair: d,
	}
}

func (d *dictValue) Text(v string) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.TextValue(v),
	})

	return d.pair.parent
}

func (d *dictValue) Bytes(v []byte) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.BytesValue(v),
	})

	return d.pair.parent
}

func (d *dictValue) Bool(v bool) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.BoolValue(v),
	})

	return d.pair.parent
}

func (d *dictValue) Uint64(v uint64) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.Uint64Value(v),
	})

	return d.pair.parent
}

func (d *dictValue) Int64(v int64) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.Int64Value(v),
	})

	return d.pair.parent
}

func (d *dictValue) Uint32(v uint32) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.Uint32Value(v),
	})

	return d.pair.parent
}

func (d *dictValue) Int32(v int32) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.Int32Value(v),
	})

	return d.pair.parent
}

func (d *dictValue) Uint16(v uint16) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.Uint16Value(v),
	})

	return d.pair.parent
}

func (d *dictValue) Int16(v int16) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.Int16Value(v),
	})

	return d.pair.parent
}

func (d *dictValue) Uint8(v uint8) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.Uint8Value(v),
	})

	return d.pair.parent
}

func (d *dictValue) Int8(v int8) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.Int8Value(v),
	})

	return d.pair.parent
}

func (d *dictValue) Float(v float32) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.FloatValue(v),
	})

	return d.pair.parent
}

func (d *dictValue) Double(v float64) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.DoubleValue(v),
	})

	return d.pair.parent
}

func (d *dictValue) Decimal(v [16]byte, precision, scale uint32) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.DecimalValue(v, precision, scale),
	})

	return d.pair.parent
}

func (d *dictValue) Timestamp(v time.Time) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.TimestampValueFromTime(v),
	})

	return d.pair.parent
}

func (d *dictValue) Date(v time.Time) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.DateValueFromTime(v),
	})

	return d.pair.parent
}

func (d *dictValue) Datetime(v time.Time) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.DatetimeValueFromTime(v),
	})

	return d.pair.parent
}

func (d *dictValue) Interval(v time.Duration) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.IntervalValueFromDuration(v),
	})

	return d.pair.parent
}

func (d *dictValue) JSON(v string) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.JSONValue(v),
	})

	return d.pair.parent
}

func (d *dictValue) JSONDocument(v string) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.JSONDocumentValue(v),
	})

	return d.pair.parent
}

func (d *dictValue) YSON(v []byte) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.YSONValue(v),
	})

	return d.pair.parent
}

//// UUID has data corruption bug and will be removed in next version.
////
//// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
//// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (d *dictValue) UUID(v [16]byte) *dict {
//	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
//		K: d.pair.keyValue,
//		V: value.UUIDWithIssue1501Value(v),
//	})
//
//	return d.pair.parent
//}

func (d *dictValue) Uuid(v uuid.UUID) *dict { //nolint:revive
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.Uuid(v),
	})

	return d.pair.parent
}

func (d *dictValue) UUIDWithIssue1501Value(v [16]byte) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.UUIDWithIssue1501Value(v),
	})

	return d.pair.parent
}

func (d *dictValue) TzDate(v time.Time) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.TzDateValueFromTime(v),
	})

	return d.pair.parent
}

func (d *dictValue) TzTimestamp(v time.Time) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.TzTimestampValueFromTime(v),
	})

	return d.pair.parent
}

func (d *dictValue) TzDatetime(v time.Time) *dict {
	d.pair.parent.values = append(d.pair.parent.values, value.DictValueField{
		K: d.pair.keyValue,
		V: value.TzDatetimeValueFromTime(v),
	})

	return d.pair.parent
}

func (d *dict) EndDict() Builder {
	d.parent.params = append(d.parent.params, &Parameter{
		parent: d.parent,
		name:   d.name,
		value:  value.DictValue(d.values...),
	})

	return d.parent
}
