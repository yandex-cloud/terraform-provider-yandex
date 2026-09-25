package params

import (
	"time"

	"github.com/google/uuid"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
)

type (
	variantStruct struct {
		parent *variant

		fields []types.StructField
		name   string
		value  value.Value
	}

	variantStructField struct {
		name   string
		parent *variantStruct
	}

	variantStructItem struct {
		parent *variantStruct
	}

	variantStructBuilder struct {
		parent *variantStruct
	}
)

func (vs *variantStruct) Field(name string) *variantStructField {
	return &variantStructField{
		name:   name,
		parent: vs,
	}
}

func (vs *variantStruct) AddFields(args ...types.StructField) *variantStruct {
	vs.fields = append(vs.fields, args...)

	return vs
}

func (vsf *variantStructField) Text() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Text,
	})

	return vsf.parent
}

func (vsf *variantStructField) Bytes() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Bytes,
	})

	return vsf.parent
}

func (vsf *variantStructField) Bool() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Bool,
	})

	return vsf.parent
}

func (vsf *variantStructField) Uint64() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Uint64,
	})

	return vsf.parent
}

func (vsf *variantStructField) Int64() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Int64,
	})

	return vsf.parent
}

func (vsf *variantStructField) Uint32() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Uint32,
	})

	return vsf.parent
}

func (vsf *variantStructField) Int32() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Int32,
	})

	return vsf.parent
}

func (vsf *variantStructField) Uint16() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Uint16,
	})

	return vsf.parent
}

func (vsf *variantStructField) Int16() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Int16,
	})

	return vsf.parent
}

func (vsf *variantStructField) Uint8() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Uint8,
	})

	return vsf.parent
}

func (vsf *variantStructField) Int8() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Int8,
	})

	return vsf.parent
}

func (vsf *variantStructField) Float() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Float,
	})

	return vsf.parent
}

func (vsf *variantStructField) Double() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Double,
	})

	return vsf.parent
}

func (vsf *variantStructField) Decimal(precision, scale uint32) *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.NewDecimal(precision, scale),
	})

	return vsf.parent
}

func (vsf *variantStructField) Timestamp() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Timestamp,
	})

	return vsf.parent
}

func (vsf *variantStructField) Date() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Date,
	})

	return vsf.parent
}

func (vsf *variantStructField) Datetime() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Datetime,
	})

	return vsf.parent
}

func (vsf *variantStructField) Interval() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.Interval,
	})

	return vsf.parent
}

func (vsf *variantStructField) JSON() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.JSON,
	})

	return vsf.parent
}

func (vsf *variantStructField) JSONDocument() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.JSONDocument,
	})

	return vsf.parent
}

func (vsf *variantStructField) YSON() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.YSON,
	})

	return vsf.parent
}

//// UUID has data corruption bug and will be removed in next version.
////
//// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
//// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (vsf *variantStructField) UUID() *variantStruct {
//	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
//		Name: vsf.name,
//		T:    types.UUID,
//	})
//
//	return vsf.parent
//}

func (vsf *variantStructField) Uuid() *variantStruct { //nolint:revive
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.UUID,
	})

	return vsf.parent
}

func (vsf *variantStructField) UUIDWithIssue1501Value() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.UUID,
	})

	return vsf.parent
}

func (vsf *variantStructField) TzDate() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.TzDate,
	})

	return vsf.parent
}

func (vsf *variantStructField) TzTimestamp() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.TzTimestamp,
	})

	return vsf.parent
}

func (vsf *variantStructField) TzDatetime() *variantStruct {
	vsf.parent.fields = append(vsf.parent.fields, types.StructField{
		Name: vsf.name,
		T:    types.TzDatetime,
	})

	return vsf.parent
}

func (vs *variantStruct) Name(name string) *variantStructItem {
	vs.name = name

	return &variantStructItem{
		parent: vs,
	}
}

func (vsi *variantStructItem) Text(v string) *variantStructBuilder {
	vsi.parent.value = value.TextValue(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Bytes(v []byte) *variantStructBuilder {
	vsi.parent.value = value.BytesValue(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Bool(v bool) *variantStructBuilder {
	vsi.parent.value = value.BoolValue(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Uint64(v uint64) *variantStructBuilder {
	vsi.parent.value = value.Uint64Value(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Int64(v int64) *variantStructBuilder {
	vsi.parent.value = value.Int64Value(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Uint32(v uint32) *variantStructBuilder {
	vsi.parent.value = value.Uint32Value(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Int32(v int32) *variantStructBuilder {
	vsi.parent.value = value.Int32Value(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Uint16(v uint16) *variantStructBuilder {
	vsi.parent.value = value.Uint16Value(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Int16(v int16) *variantStructBuilder {
	vsi.parent.value = value.Int16Value(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Uint8(v uint8) *variantStructBuilder {
	vsi.parent.value = value.Uint8Value(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Int8(v int8) *variantStructBuilder {
	vsi.parent.value = value.Int8Value(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Float(v float32) *variantStructBuilder {
	vsi.parent.value = value.FloatValue(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Double(v float64) *variantStructBuilder {
	vsi.parent.value = value.DoubleValue(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Decimal(v [16]byte, precision, scale uint32) *variantStructBuilder {
	vsi.parent.value = value.DecimalValue(v, precision, scale)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Timestamp(v time.Time) *variantStructBuilder {
	vsi.parent.value = value.TimestampValueFromTime(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Date(v time.Time) *variantStructBuilder {
	vsi.parent.value = value.DateValueFromTime(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Datetime(v time.Time) *variantStructBuilder {
	vsi.parent.value = value.DatetimeValueFromTime(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) Interval(v time.Duration) *variantStructBuilder {
	vsi.parent.value = value.IntervalValueFromDuration(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) JSON(v string) *variantStructBuilder {
	vsi.parent.value = value.JSONValue(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) JSONDocument(v string) *variantStructBuilder {
	vsi.parent.value = value.JSONDocumentValue(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) YSON(v []byte) *variantStructBuilder {
	vsi.parent.value = value.YSONValue(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

//// UUID has data corruption bug and will be removed in next version.
////
//// Deprecated: Use Uuid (prefer) or UUIDWithIssue1501Value (for save old behavior) instead.
//// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (vsi *variantStructItem) UUID(v [16]byte) *variantStructBuilder {
//	vsi.parent.value = value.UUIDWithIssue1501Value(v)
//
//	return &variantStructBuilder{
//		parent: vsi.parent,
//	}
//}

func (vsi *variantStructItem) Uuid(v uuid.UUID) *variantStructBuilder { //nolint:revive
	vsi.parent.value = value.Uuid(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) UUIDWithIssue1501Value(v [16]byte) *variantStructBuilder {
	vsi.parent.value = value.UUIDWithIssue1501Value(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) TzDate(v time.Time) *variantStructBuilder {
	vsi.parent.value = value.TzDateValueFromTime(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) TzTimestamp(v time.Time) *variantStructBuilder {
	vsi.parent.value = value.TzTimestampValueFromTime(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsi *variantStructItem) TzDatetime(v time.Time) *variantStructBuilder {
	vsi.parent.value = value.TzDatetimeValueFromTime(v)

	return &variantStructBuilder{
		parent: vsi.parent,
	}
}

func (vsb *variantStructBuilder) EndStruct() *variantBuilder {
	vsb.parent.parent.value = value.VariantValueStruct(
		vsb.parent.value,
		vsb.parent.name,
		types.NewVariantStruct(vsb.parent.fields...),
	)

	return &variantBuilder{
		variant: vsb.parent.parent,
	}
}
