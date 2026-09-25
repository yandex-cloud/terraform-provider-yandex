package value

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/decimal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

const (
	decimalPrecision uint32 = 22
	decimalScale     uint32 = 9
)

type Value interface {
	Type() types.Type
	Yql() string

	castTo(dst any) error
	toYDB() *Ydb.Value
}

func ToYDB(v Value) *Ydb.TypedValue {
	if v == nil {
		return nil
	}

	return &Ydb.TypedValue{
		Type:  v.Type().ToYDB(),
		Value: v.toYDB(),
	}
}

// BigEndianUint128 builds a big-endian uint128 value.
func BigEndianUint128(hi, lo uint64) (v [16]byte) {
	binary.BigEndian.PutUint64(v[0:8], hi)
	binary.BigEndian.PutUint64(v[8:16], lo)

	return v
}

func FromYDB(t *Ydb.Type, v *Ydb.Value) Value {
	vv, err := fromYDB(t, v)
	if err != nil {
		panic(err)
	}

	return vv
}

func nullValueFromYDB(x *Ydb.Value, t types.Type) (_ Value, ok bool) {
	for {
		switch xx := x.GetValue().(type) {
		case *Ydb.Value_NestedValue:
			x = xx.NestedValue
		case *Ydb.Value_NullFlagValue:
			switch tt := t.(type) {
			case types.Optional:
				return NullValue(tt.InnerType()), true
			case types.Void:
				return VoidValue(), true
			default:
				return nil, false
			}
		default:
			return nil, false
		}
	}
}

//nolint:funlen,gocyclo
func primitiveValueFromYDB(t types.Primitive, v *Ydb.Value) (Value, error) {
	switch t {
	case types.Bool:
		return BoolValue(v.GetBoolValue()), nil

	case types.Int8:
		return Int8Value(int8(v.GetInt32Value())), nil

	case types.Int16:
		return Int16Value(int16(v.GetInt32Value())), nil

	case types.Int32:
		return Int32Value(v.GetInt32Value()), nil

	case types.Int64:
		return Int64Value(v.GetInt64Value()), nil

	case types.Uint8:
		return Uint8Value(uint8(v.GetUint32Value())), nil

	case types.Uint16:
		return Uint16Value(uint16(v.GetUint32Value())), nil

	case types.Uint32:
		return Uint32Value(v.GetUint32Value()), nil

	case types.Uint64:
		return Uint64Value(v.GetUint64Value()), nil

	case types.Date:
		return DateValue(v.GetUint32Value()), nil

	case types.Date32:
		return Date32Value(v.GetInt32Value()), nil

	case types.Datetime:
		return DatetimeValue(v.GetUint32Value()), nil

	case types.Datetime64:
		return Datetime64Value(v.GetInt64Value()), nil

	case types.Interval:
		return IntervalValue(v.GetInt64Value()), nil

	case types.Interval64:
		return Interval64Value(v.GetInt64Value()), nil

	case types.Timestamp:
		return TimestampValue(v.GetUint64Value()), nil

	case types.Timestamp64:
		return Timestamp64Value(v.GetInt64Value()), nil

	case types.Float:
		return FloatValue(v.GetFloatValue()), nil

	case types.Double:
		return DoubleValue(v.GetDoubleValue()), nil

	case types.Text:
		return TextValue(v.GetTextValue()), nil

	case types.YSON:
		switch vv := v.GetValue().(type) {
		case *Ydb.Value_TextValue:
			return YSONValue(xstring.ToBytes(vv.TextValue)), nil
		case *Ydb.Value_BytesValue:
			return YSONValue(vv.BytesValue), nil
		default:
			return nil, xerrors.WithStackTrace(fmt.Errorf("uncovered YSON internal type: %T", vv))
		}

	case types.JSON:
		return JSONValue(v.GetTextValue()), nil

	case types.JSONDocument:
		return JSONDocumentValue(v.GetTextValue()), nil

	case types.DyNumber:
		return DyNumberValue(v.GetTextValue()), nil

	case types.TzDate:
		return TzDateValue(v.GetTextValue()), nil

	case types.TzDatetime:
		return TzDatetimeValue(v.GetTextValue()), nil

	case types.TzTimestamp:
		return TzTimestampValue(v.GetTextValue()), nil

	case types.Bytes:
		return BytesValue(v.GetBytesValue()), nil

	case types.UUID:
		return UUIDFromYDBPair(v.GetHigh_128(), v.GetLow_128()), nil
	default:
		return nil, xerrors.WithStackTrace(fmt.Errorf("uncovered primitive type: %T", t))
	}
}

//nolint:funlen
func fromYDB(t *Ydb.Type, v *Ydb.Value) (Value, error) {
	tt := types.TypeFromYDB(t)

	if vv, ok := nullValueFromYDB(v, tt); ok {
		return vv, nil
	}

	switch ttt := tt.(type) {
	case types.Primitive:
		return primitiveValueFromYDB(ttt, v)

	case types.Void:
		return VoidValue(), nil

	case types.Null:
		return NullValue(tt), nil

	case *types.Decimal:
		return DecimalValue(BigEndianUint128(v.GetHigh_128(), v.GetLow_128()), ttt.Precision(), ttt.Scale()), nil

	case types.Optional:
		tt, ok := t.GetType().(*Ydb.Type_OptionalType)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to *Ydb.Type_OptionalType", tt))
		}
		t = tt.OptionalType.GetItem()
		if nestedValue, ok := v.GetValue().(*Ydb.Value_NestedValue); ok {
			return OptionalValue(FromYDB(t, nestedValue.NestedValue)), nil
		}

		return OptionalValue(FromYDB(t, v)), nil

	case *types.List:
		return ListValue(func() []Value {
			vv := make([]Value, len(v.GetItems()))
			for i, vvv := range v.GetItems() {
				vv[i] = FromYDB(ttt.ItemType().ToYDB(), vvv)
			}

			return vv
		}()...), nil

	case *types.Tuple:
		return TupleValue(func() []Value {
			vv := make([]Value, len(v.GetItems()))
			for i, vvv := range v.GetItems() {
				vv[i] = FromYDB(ttt.ItemType(i).ToYDB(), vvv)
			}

			return vv
		}()...), nil

	case *types.Struct:
		return StructValue(func() []StructValueField {
			vv := make([]StructValueField, len(v.GetItems()))
			for i, vvv := range v.GetItems() {
				vv[i] = StructValueField{
					Name: ttt.Field(i).Name,
					V:    FromYDB(ttt.Field(i).T.ToYDB(), vvv),
				}
			}

			return vv
		}()...), nil

	case *types.Dict:
		return DictValue(func() []DictValueField {
			vv := make([]DictValueField, len(v.GetPairs()))
			for i, vvv := range v.GetPairs() {
				vv[i] = DictValueField{
					K: FromYDB(ttt.KeyType().ToYDB(), vvv.GetKey()),
					V: FromYDB(ttt.ValueType().ToYDB(), vvv.GetPayload()),
				}
			}

			return vv
		}()...), nil

	case *types.Set:
		return SetValue(func() []Value {
			vv := make([]Value, len(v.GetPairs()))
			for i, vvv := range v.GetPairs() {
				vv[i] = FromYDB(ttt.ItemType().ToYDB(), vvv.GetKey())
			}

			return vv
		}()...), nil

	case *types.VariantStruct:

		val, ok := v.GetValue().(*Ydb.Value_NestedValue)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to *Ydb.Value_NestedValue", val))
		}

		return VariantValueStruct(
			FromYDB(
				ttt.Struct.Field(int(v.GetVariantIndex())).T.ToYDB(),
				val.NestedValue,
			),
			ttt.Struct.Field(int(v.GetVariantIndex())).Name,
			ttt.Struct,
		), nil

	case *types.VariantTuple:

		val, ok := v.GetValue().(*Ydb.Value_NestedValue)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to *Ydb.Value_NestedValue", val))
		}

		return VariantValueTuple(
			FromYDB(
				ttt.Tuple.ItemType(int(v.GetVariantIndex())).ToYDB(),
				val.NestedValue,
			),
			v.GetVariantIndex(),
			ttt.Tuple,
		), nil

	case *types.PgType:
		return &pgValue{
			t: types.PgType{
				OID: ttt.OID,
			},
			val: v.GetTextValue(),
		}, nil

	default:
		return nil, xerrors.WithStackTrace(fmt.Errorf("uncovered type: %T", ttt))
	}
}

type boolValue bool

func (v boolValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *bool:
		*vv = bool(v)

		return nil
	case *driver.Value:
		*vv = bool(v)

		return nil
	case *string:
		*vv = strconv.FormatBool(bool(v))

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v boolValue) Yql() string {
	return strconv.FormatBool(bool(v))
}

func (boolValue) Type() types.Type {
	return types.Bool
}

func (v boolValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_BoolValue{
			BoolValue: bool(v),
		},
	}
}

func BoolValue(v bool) boolValue {
	return boolValue(v)
}

type dateValue uint32

func (v dateValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *time.Time:
		*vv = DateToTime(uint32(v)).UTC()

		return nil
	case *driver.Value:
		*vv = DateToTime(uint32(v)).UTC()

		return nil
	case *uint64:
		*vv = uint64(v)

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	case *int32:
		*vv = int32(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v dateValue) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), DateToTime(uint32(v)).UTC().Format(LayoutDate))
}

func (dateValue) Type() types.Type {
	return types.Date
}

func (v dateValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Uint32Value{
			Uint32Value: uint32(v),
		},
	}
}

// DateValue returns ydb date value by given days since Epoch
func DateValue(v uint32) dateValue {
	return dateValue(v)
}

func DateValueFromTime(t time.Time) dateValue {
	return dateValue(uint64(t.Sub(epoch)/time.Second) / secondsPerDay)
}

type date32Value int32

func (v date32Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *time.Time:
		*vv = Date32ToTime(int32(v)).UTC()

		return nil
	case *driver.Value:
		*vv = Date32ToTime(int32(v)).UTC()

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	case *int32:
		*vv = int32(v)

		return nil
	case *int:
		*vv = int(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v date32Value) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), Date32ToTime(int32(v)).UTC().Format(LayoutDate))
}

func (date32Value) Type() types.Type {
	return types.Date32
}

func (v date32Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Int32Value{
			Int32Value: int32(v),
		},
	}
}

// Date32Value returns ydb date value from days around epoch time
func Date32Value(v int32) date32Value {
	return date32Value(v)
}

func Date32ValueFromTime(t time.Time) date32Value {
	return date32Value(uint64(t.Unix() / 24 / 60 / 60))
}

type datetimeValue uint32

func (v datetimeValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *time.Time:
		*vv = DatetimeToTime(uint32(v))

		return nil
	case *driver.Value:
		*vv = DatetimeToTime(uint32(v))

		return nil
	case *uint64:
		*vv = uint64(v)

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	case *uint32:
		*vv = uint32(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v datetimeValue) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), DatetimeToTime(uint32(v)).UTC().Format(LayoutDatetime))
}

func (datetimeValue) Type() types.Type {
	return types.Datetime
}

func (v datetimeValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Uint32Value{
			Uint32Value: uint32(v),
		},
	}
}

// DatetimeValue makes ydb datetime value from seconds since Epoch
func DatetimeValue(v uint32) datetimeValue {
	return datetimeValue(v)
}

func DatetimeValueFromTime(t time.Time) datetimeValue {
	return datetimeValue(t.Unix())
}

type datetime64Value int64

func (v datetime64Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *time.Time:
		*vv = Datetime64ToTime(int64(v))

		return nil
	case *driver.Value:
		*vv = Datetime64ToTime(int64(v))

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v datetime64Value) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), Datetime64ToTime(int64(v)).UTC().Format(LayoutDatetime))
}

func (datetime64Value) Type() types.Type {
	return types.Datetime64
}

func (v datetime64Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Int64Value{
			Int64Value: int64(v),
		},
	}
}

// Datetime64Value makes ydb datetime value from seconds around epoch time
func Datetime64Value(v int64) datetime64Value {
	return datetime64Value(v)
}

func Datetime64ValueFromTime(t time.Time) datetime64Value {
	return datetime64Value(t.Unix())
}

var _ DecimalValuer = (*decimalValue)(nil)

type decimalValue struct {
	value     [16]byte
	innerType *types.Decimal
}

func (v *decimalValue) Value() [16]byte {
	return v.value
}

func (v *decimalValue) Precision() uint32 {
	return v.innerType.Precision()
}

func (v *decimalValue) Scale() uint32 {
	return v.innerType.Scale()
}

type DecimalValuer interface {
	Value() [16]byte
	Precision() uint32
	Scale() uint32
}

func (v *decimalValue) castTo(dst any) error {
	switch dstValue := dst.(type) {
	case *driver.Value:
		*dstValue = v

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, dstValue,
		))
	}
}

func (v *decimalValue) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString(v.innerType.Name())
	buffer.WriteByte('(')
	buffer.WriteByte('"')
	s := decimal.FromBytes(v.value[:], v.innerType.Precision(), v.innerType.Scale()).String()
	if len(s) < int(v.innerType.Scale()) {
		s = strings.Repeat("0", int(v.innerType.Scale())-len(s)) + s
	}
	buffer.WriteString(s[:len(s)-int(v.innerType.Scale())] + "." + s[len(s)-int(v.innerType.Scale()):])
	buffer.WriteByte('"')
	buffer.WriteByte(',')
	buffer.WriteString(strconv.FormatUint(uint64(v.innerType.Precision()), 10))
	buffer.WriteByte(',')
	buffer.WriteString(strconv.FormatUint(uint64(v.innerType.Scale()), 10))
	buffer.WriteByte(')')

	return buffer.String()
}

func (v *decimalValue) Type() types.Type {
	return v.innerType
}

func (v *decimalValue) toYDB() *Ydb.Value {
	var bytes [16]byte
	if v != nil {
		bytes = v.value
	}

	return &Ydb.Value{
		High_128: binary.BigEndian.Uint64(bytes[0:8]),
		Value: &Ydb.Value_Low_128{
			Low_128: binary.BigEndian.Uint64(bytes[8:16]),
		},
	}
}

func DecimalValueFromBigInt(v *big.Int, precision, scale uint32) *decimalValue {
	b := decimal.BigIntToByte(v, precision, scale)

	return DecimalValue(b, precision, scale)
}

func DecimalValueFromString(str string, precision, scale uint32) (Value, error) {
	bigI, err := decimal.Parse(str, precision, scale)
	if err != nil {
		return nil, err
	}

	return DecimalValueFromBigInt(bigI, precision, scale), nil
}

func DecimalValue(v [16]byte, precision, scale uint32) *decimalValue {
	return &decimalValue{
		value: v,
		innerType: types.NewDecimal(
			precision,
			scale,
		),
	}
}

type (
	DictValueField struct {
		K Value
		V Value
	}
	dictValue struct {
		t      types.Type
		values []DictValueField
	}
)

func (v *dictValue) DictValues() map[Value]Value {
	values := make(map[Value]Value, len(v.values))
	for i := range v.values {
		values[v.values[i].K] = v.values[i].V
	}

	return values
}

func (v *dictValue) castTo(dst any) error {
	switch dstValue := dst.(type) {
	case *driver.Value:
		*dstValue = v

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, dstValue,
		))
	}
}

func (v *dictValue) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteByte('{')
	for i := range v.values {
		if i != 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteString(v.values[i].K.Yql())
		buffer.WriteByte(':')
		buffer.WriteString(v.values[i].V.Yql())
	}
	buffer.WriteByte('}')

	return buffer.String()
}

func (v *dictValue) Type() types.Type {
	return v.t
}

func (v *dictValue) toYDB() *Ydb.Value {
	var values []DictValueField
	if v != nil {
		values = v.values
	}
	pairs := make([]*Ydb.ValuePair, 0, len(values))
	for i := range values {
		pairs = append(pairs, &Ydb.ValuePair{
			Key:     values[i].K.toYDB(),
			Payload: values[i].V.toYDB(),
		})
	}

	return &Ydb.Value{
		Pairs: pairs,
	}
}

func DictValue(values ...DictValueField) *dictValue {
	sort.Slice(values, func(i, j int) bool {
		return values[i].K.Yql() < values[j].K.Yql()
	})
	var t types.Type
	switch {
	case len(values) > 0:
		t = types.NewDict(values[0].K.Type(), values[0].V.Type())
	default:
		t = types.NewEmptyDict()
	}

	return &dictValue{
		t:      t,
		values: values,
	}
}

type doubleValue struct {
	value float64
}

func (v *doubleValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *float64:
		*vv = v.value

		return nil
	case *driver.Value:
		*vv = v.value

		return nil
	case *string:
		*vv = strconv.FormatFloat(v.value, 'f', -1, 64)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatFloat(v.value, 'f', -1, 64))

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v *doubleValue) Yql() string {
	return fmt.Sprintf("%s(\"%v\")", v.Type().Yql(), v.value)
}

func (*doubleValue) Type() types.Type {
	return types.Double
}

func (v *doubleValue) toYDB() *Ydb.Value {
	var value float64
	if v != nil {
		value = v.value
	}

	return &Ydb.Value{
		Value: &Ydb.Value_DoubleValue{
			DoubleValue: value,
		},
	}
}

func DoubleValue(v float64) *doubleValue {
	return &doubleValue{value: v}
}

type dyNumberValue string

func (v dyNumberValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *string:
		*vv = string(v)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(string(v))

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v dyNumberValue) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), string(v))
}

func (dyNumberValue) Type() types.Type {
	return types.DyNumber
}

func (v dyNumberValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_TextValue{
			TextValue: string(v),
		},
	}
}

func DyNumberValue(v string) dyNumberValue {
	return dyNumberValue(v)
}

type floatValue struct {
	value float32
}

func (v *floatValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *float32:
		*vv = v.value

		return nil
	case *float64:
		*vv = float64(v.value)

		return nil
	case *driver.Value:
		*vv = v.value

		return nil
	case *string:
		*vv = strconv.FormatFloat(float64(v.value), 'f', -1, 32)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatFloat(float64(v.value), 'f', -1, 32))

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v *floatValue) Yql() string {
	return fmt.Sprintf("%s(\"%v\")", v.Type().Yql(), v.value)
}

func (*floatValue) Type() types.Type {
	return types.Float
}

func (v *floatValue) toYDB() *Ydb.Value {
	var value float32
	if v != nil {
		value = v.value
	}

	return &Ydb.Value{
		Value: &Ydb.Value_FloatValue{
			FloatValue: value,
		},
	}
}

func FloatValue(v float32) *floatValue {
	return &floatValue{value: v}
}

type int8Value int8

func (v int8Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *int8:
		*vv = int8(v)

		return nil
	case *driver.Value:
		*vv = int8(v)

		return nil
	case *string:
		*vv = strconv.FormatInt(int64(v), 10)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatInt(int64(v), 10))

		return nil

	case *int64:
		*vv = int64(v)

		return nil

	case *int32:
		*vv = int32(v)

		return nil

	case *int16:
		*vv = int16(v)

		return nil

	case *float64:
		*vv = float64(v)

		return nil
	case *float32:
		*vv = float32(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v int8Value) Yql() string {
	return strconv.FormatInt(int64(v), 10) + "t"
}

func (int8Value) Type() types.Type {
	return types.Int8
}

func (v int8Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Int32Value{
			Int32Value: int32(v),
		},
	}
}

func Int8Value(v int8) int8Value {
	return int8Value(v)
}

type int16Value int16

func (v int16Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *int16:
		*vv = int16(v)

		return nil
	case *driver.Value:
		*vv = int16(v)

		return nil
	case *string:
		*vv = strconv.FormatInt(int64(v), 10)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatInt(int64(v), 10))

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	case *int32:
		*vv = int32(v)

		return nil
	case *float64:
		*vv = float64(v)

		return nil
	case *float32:
		*vv = float32(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v int16Value) Yql() string {
	return strconv.FormatInt(int64(v), 10) + "s"
}

func (int16Value) Type() types.Type {
	return types.Int16
}

func (v int16Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Int32Value{
			Int32Value: int32(v),
		},
	}
}

func Int16Value(v int16) int16Value {
	return int16Value(v)
}

type int32Value int32

func (v int32Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *int32:
		*vv = int32(v)

		return nil
	case *driver.Value:
		*vv = int32(v)

		return nil
	case *string:
		*vv = strconv.FormatInt(int64(v), 10)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatInt(int64(v), 10))

		return nil

	case *int64:
		*vv = int64(v)

		return nil

	case *int:
		*vv = int(v)

		return nil

	case *float64:
		*vv = float64(v)

		return nil
	case *float32:
		*vv = float32(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v int32Value) Yql() string {
	return strconv.FormatInt(int64(v), 10)
}

func (int32Value) Type() types.Type {
	return types.Int32
}

func (v int32Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Int32Value{
			Int32Value: int32(v),
		},
	}
}

func Int32Value(v int32) int32Value {
	return int32Value(v)
}

type int64Value int64

func (v int64Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *int64:
		*vv = int64(v)

		return nil
	case *driver.Value:
		*vv = int64(v)

		return nil
	case *string:
		*vv = strconv.FormatInt(int64(v), 10)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatInt(int64(v), 10))

		return nil

	case *float64:
		*vv = float64(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v int64Value) Yql() string {
	return strconv.FormatInt(int64(v), 10) + "l"
}

func (int64Value) Type() types.Type {
	return types.Int64
}

func (v int64Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Int64Value{
			Int64Value: int64(v),
		},
	}
}

func Int64Value(v int64) int64Value {
	return int64Value(v)
}

type intervalValue int64

func (v intervalValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *time.Duration:
		*vv = IntervalToDuration(int64(v))

		return nil
	case *driver.Value:
		*vv = IntervalToDuration(int64(v))

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v intervalValue) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString(v.Type().Yql())
	buffer.WriteByte('(')
	buffer.WriteByte('"')
	d := IntervalToDuration(int64(v))
	if d < 0 {
		buffer.WriteByte('-')
		d = -d
	}
	buffer.WriteByte('P')
	//nolint:mnd
	if days := d / time.Hour / 24; days > 0 {
		d -= days * time.Hour * 24 //nolint:durationcheck
		buffer.WriteString(strconv.FormatInt(int64(days), 10))
		buffer.WriteByte('D')
	}
	if d > 0 {
		buffer.WriteByte('T')
	}
	if hours := d / time.Hour; hours > 0 {
		d -= hours * time.Hour //nolint:durationcheck
		buffer.WriteString(strconv.FormatInt(int64(hours), 10))
		buffer.WriteByte('H')
	}
	if minutes := d / time.Minute; minutes > 0 {
		d -= minutes * time.Minute //nolint:durationcheck
		buffer.WriteString(strconv.FormatInt(int64(minutes), 10))
		buffer.WriteByte('M')
	}
	if d > 0 {
		seconds := float64(d) / float64(time.Second)
		fmt.Fprintf(buffer, "%0.6f", seconds)
		buffer.WriteByte('S')
	}
	buffer.WriteByte('"')
	buffer.WriteByte(')')

	return buffer.String()
}

func (intervalValue) Type() types.Type {
	return types.Interval
}

func (v intervalValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Int64Value{
			Int64Value: int64(v),
		},
	}
}

// IntervalValue makes Value from given microseconds value
func IntervalValue(v int64) intervalValue {
	return intervalValue(v)
}

func IntervalValueFromDuration(v time.Duration) intervalValue {
	return intervalValue(durationToMicroseconds(v))
}

type interval64Value int64

func (v interval64Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *time.Duration:
		*vv = Interval64ToDuration(int64(v))

		return nil
	case *driver.Value:
		*vv = Interval64ToDuration(int64(v))

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v interval64Value) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString(v.Type().Yql())
	buffer.WriteByte('(')
	buffer.WriteByte('"')
	d := IntervalToDuration(int64(v))
	if d < 0 {
		buffer.WriteByte('-')
		d = -d
	}
	buffer.WriteByte('P')
	//nolint:mnd
	if days := d / time.Hour / 24; days > 0 {
		d -= days * time.Hour * 24 //nolint:durationcheck
		buffer.WriteString(strconv.FormatInt(int64(days), 10))
		buffer.WriteByte('D')
	}
	if d > 0 {
		buffer.WriteByte('T')
	}
	if hours := d / time.Hour; hours > 0 {
		d -= hours * time.Hour //nolint:durationcheck
		buffer.WriteString(strconv.FormatInt(int64(hours), 10))
		buffer.WriteByte('H')
	}
	if minutes := d / time.Minute; minutes > 0 {
		d -= minutes * time.Minute //nolint:durationcheck
		buffer.WriteString(strconv.FormatInt(int64(minutes), 10))
		buffer.WriteByte('M')
	}
	if d > 0 {
		seconds := float64(d) / float64(time.Second)
		fmt.Fprintf(buffer, "%0.6f", seconds)
		buffer.WriteByte('S')
	}
	buffer.WriteByte('"')
	buffer.WriteByte(')')

	return buffer.String()
}

func (interval64Value) Type() types.Type {
	return types.Interval64
}

func (v interval64Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Int64Value{
			Int64Value: int64(v),
		},
	}
}

// Interval64Value makes Value from given nanoseconds around epoch time
func Interval64Value(v int64) interval64Value {
	return interval64Value(v)
}

func Interval64ValueFromDuration(v time.Duration) interval64Value {
	return interval64Value(durationToNanoseconds(v))
}

type jsonValue string

func (v jsonValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *string:
		*vv = string(v)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(string(v))

		return nil
	case json.Unmarshaler:
		err := vv.UnmarshalJSON(xstring.ToBytes(string(v)))
		if err != nil {
			return xerrors.WithStackTrace(fmt.Errorf(
				"%w '%s(%+v)' to '%T' destination: %w",
				ErrCannotCast, v.Type().Yql(), v, vv, err,
			))
		}

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v jsonValue) Yql() string {
	return fmt.Sprintf("%s(@@%s@@)", v.Type().Yql(), string(v))
}

func (jsonValue) Type() types.Type {
	return types.JSON
}

func (v jsonValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_TextValue{
			TextValue: string(v),
		},
	}
}

func JSONValue(v string) jsonValue {
	return jsonValue(v)
}

type jsonDocumentValue string

func (v jsonDocumentValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *string:
		*vv = string(v)

		return nil
	case *driver.Value:
		*vv = string(v)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(string(v))

		return nil
	case json.Unmarshaler:
		err := vv.UnmarshalJSON(xstring.ToBytes(string(v)))
		if err != nil {
			return xerrors.WithStackTrace(fmt.Errorf(
				"%w '%s(%+v)' to '%T' destination: %w",
				ErrCannotCast, v.Type().Yql(), v, vv, err,
			))
		}

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v jsonDocumentValue) Yql() string {
	return fmt.Sprintf("%s(@@%s@@)", v.Type().Yql(), string(v))
}

func (jsonDocumentValue) Type() types.Type {
	return types.JSONDocument
}

func (v jsonDocumentValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_TextValue{
			TextValue: string(v),
		},
	}
}

func JSONDocumentValue(v string) jsonDocumentValue {
	return jsonDocumentValue(v)
}

type listValue struct {
	t     types.Type
	items []Value
}

func (v *listValue) ListItems() []Value {
	return v.items
}

func (v *listValue) castTo(dst any) error {
	switch dstValue := dst.(type) {
	case *driver.Value:
		*dstValue = v

		return nil
	case interface{}:
		ptr := reflect.ValueOf(dstValue)

		inner := reflect.Indirect(ptr)
		if inner.Kind() != reflect.Slice && inner.Kind() != reflect.Array {
			return xerrors.WithStackTrace(fmt.Errorf(
				"%w '%s(%+v)' to '%T' destination",
				ErrCannotCast, v.Type().Yql(), v, dstValue,
			))
		}

		targetType := inner.Type().Elem()
		valueInner := reflect.ValueOf(v.ListItems())

		newSlice := reflect.MakeSlice(reflect.SliceOf(targetType), valueInner.Len(), valueInner.Cap())
		inner.Set(newSlice)

		for i, item := range v.ListItems() {
			if err := item.castTo(inner.Index(i).Addr().Interface()); err != nil {
				return xerrors.WithStackTrace(fmt.Errorf(
					"%w '%s(%+v)' to '%T' destination: %w",
					ErrCannotCast, v.Type().Yql(), v, dstValue, err,
				))
			}
		}

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, dstValue,
		))
	}
}

func (v *listValue) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteByte('[')
	for i, item := range v.items {
		if i != 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteString(item.Yql())
	}
	buffer.WriteByte(']')

	return buffer.String()
}

func (v *listValue) Type() types.Type {
	return v.t
}

func (v *listValue) toYDB() *Ydb.Value {
	var items []Value
	if v != nil {
		items = v.items
	}
	result := &Ydb.Value{
		Items: make([]*Ydb.Value, 0, len(items)),
	}

	for _, vv := range items {
		result.Items = append(result.Items, vv.toYDB())
	}

	return result
}

func ListValue(items ...Value) *listValue {
	var t types.Type
	switch {
	case len(items) > 0:
		t = types.NewList(items[0].Type())
	default:
		t = types.NewEmptyList()
	}

	return &listValue{
		t:     t,
		items: items,
	}
}

type pgValue struct {
	t   types.PgType
	val string
}

func (v pgValue) castTo(dst any) error {
	switch dstValue := dst.(type) {
	case *driver.Value:
		*dstValue = v

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, dstValue,
		))
	}
}

func (v pgValue) Type() types.Type {
	return v.t
}

func (v pgValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_TextValue{
			TextValue: v.val,
		},
	}
}

func (v pgValue) Yql() string {
	return fmt.Sprintf(`PgConst("%v", PgType(%v))`, v.val, v.t.OID)
}

type setValue struct {
	t     types.Type
	items []Value
}

func (v *setValue) castTo(dst any) error {
	switch dstValue := dst.(type) {
	case *driver.Value:
		*dstValue = v

		return nil
	case interface{}:
		ptr := reflect.ValueOf(dstValue)

		inner := reflect.Indirect(ptr)
		if inner.Kind() != reflect.Slice && inner.Kind() != reflect.Array {
			return xerrors.WithStackTrace(fmt.Errorf(
				"%w '%s(%+v)' to '%T' destination",
				ErrCannotCast, v.Type().Yql(), v, dstValue,
			))
		}

		targetType := inner.Type().Elem()
		valueInner := reflect.ValueOf(v.items)

		newSlice := reflect.MakeSlice(reflect.SliceOf(targetType), valueInner.Len(), valueInner.Cap())
		inner.Set(newSlice)

		for i, item := range v.items {
			if err := item.castTo(inner.Index(i).Addr().Interface()); err != nil {
				return xerrors.WithStackTrace(fmt.Errorf(
					"%w '%s(%+v)' to '%T' destination: %w",
					ErrCannotCast, v.Type().Yql(), v, dstValue, err,
				))
			}
		}

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, dstValue,
		))
	}
}

func (v *setValue) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteByte('{')
	for i, item := range v.items {
		if i != 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteString(item.Yql())
	}
	buffer.WriteByte('}')

	return buffer.String()
}

func (v *setValue) Type() types.Type {
	return v.t
}

func (v *setValue) toYDB() *Ydb.Value {
	result := &Ydb.Value{
		Pairs: make([]*Ydb.ValuePair, 0, len(v.items)),
	}

	for _, vv := range v.items {
		result.Pairs = append(result.Pairs, &Ydb.ValuePair{
			Key:     vv.toYDB(),
			Payload: _voidValue,
		})
	}

	return result
}

func PgValue(oid uint32, val string) pgValue {
	return pgValue{
		t: types.PgType{
			OID: oid,
		},
		val: val,
	}
}

func SetValue(items ...Value) *setValue {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Yql() < items[j].Yql()
	})

	var t types.Type
	switch {
	case len(items) > 0:
		t = types.NewSet(items[0].Type())
	default:
		t = types.EmptySet()
	}

	return &setValue{
		t:     t,
		items: items,
	}
}

func NullValue(t types.Type) *optionalValue {
	return &optionalValue{
		innerType: types.NewOptional(t),
		value:     nil,
	}
}

type optionalValue struct {
	innerType types.Type
	value     Value
}

func (v *optionalValue) castTo(dst any) error {
	ptr := reflect.ValueOf(dst)
	if ptr.Kind() != reflect.Pointer {
		return xerrors.WithStackTrace(fmt.Errorf("%w: '%s'", errDestinationTypeIsNotAPointer, ptr.Kind().String()))
	}

	inner := reflect.Indirect(ptr)

	if inner.Kind() != reflect.Pointer {
		if v.value == nil {
			if ptr.CanAddr() {
				ptr.SetZero()
			}

			return nil
		}

		if err := v.value.castTo(ptr.Interface()); err != nil {
			return xerrors.WithStackTrace(err)
		}

		return nil
	}

	if v.value == nil {
		inner.SetZero()

		return nil
	}

	inner.Set(reflect.New(inner.Type().Elem()))

	if err := v.value.castTo(inner.Interface()); err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (v *optionalValue) Yql() string {
	if v.value == nil {
		return fmt.Sprintf("Nothing(%s)", v.Type().Yql())
	}

	return fmt.Sprintf("Just(%s)", v.value.Yql())
}

func (v *optionalValue) Type() types.Type {
	return v.innerType
}

func (v *optionalValue) toYDB() *Ydb.Value {
	if _, opt := v.value.(*optionalValue); opt {
		return &Ydb.Value{
			Value: &Ydb.Value_NestedValue{
				NestedValue: v.value.toYDB(),
			},
		}
	}
	if v.value != nil {
		return v.value.toYDB()
	}

	return &Ydb.Value{
		Value: &Ydb.Value_NullFlagValue{},
	}
}

func OptionalValue(v Value) *optionalValue {
	return &optionalValue{
		innerType: types.NewOptional(v.Type()),
		value:     v,
	}
}

type (
	StructValueField struct {
		Name string
		V    Value
	}
	structValue struct {
		t      types.Type
		fields []StructValueField
	}
)

func (v *structValue) StructFields() map[string]Value {
	fields := make(map[string]Value, len(v.fields))
	for i := range v.fields {
		fields[v.fields[i].Name] = v.fields[i].V
	}

	return fields
}

func (v *structValue) castTo(dst any) error {
	switch dstValue := dst.(type) {
	case *driver.Value:
		*dstValue = v

		return nil
	case interface{}:
		ptr := reflect.ValueOf(dst)

		inner := reflect.Indirect(ptr)
		if inner.Kind() != reflect.Struct {
			return xerrors.WithStackTrace(fmt.Errorf(
				"%w '%s(%+v)' to '%T' destination",
				ErrCannotCast, v.Type().Yql(), v, dstValue,
			))
		}

		for i, field := range v.fields {
			if err := field.V.castTo(inner.Field(i).Addr().Interface()); err != nil {
				return xerrors.WithStackTrace(fmt.Errorf(
					"scan error on struct field name '%s': %w",
					field.Name, err,
				))
			}
		}

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, dstValue,
		))
	}
}

func (v *structValue) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString("<|")
	for i := range v.fields {
		if i != 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteString("`" + v.fields[i].Name + "`:")
		buffer.WriteString(v.fields[i].V.Yql())
	}
	buffer.WriteString("|>")

	return buffer.String()
}

func (v *structValue) Type() types.Type {
	return v.t
}

func (v *structValue) toYDB() *Ydb.Value {
	result := &Ydb.Value{
		Items: make([]*Ydb.Value, 0, len(v.fields)),
	}

	for i := range v.fields {
		result.Items = append(result.Items, v.fields[i].V.toYDB())
	}

	return result
}

func StructValue(fields ...StructValueField) *structValue {
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})
	structFields := make([]types.StructField, 0, len(fields))
	for i := range fields {
		structFields = append(structFields, types.StructField{
			Name: fields[i].Name,
			T:    fields[i].V.Type(),
		})
	}

	return &structValue{
		t:      types.NewStruct(structFields...),
		fields: fields,
	}
}

type timestampValue uint64

func (v timestampValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *time.Time:
		*vv = TimestampToTime(uint64(v))

		return nil
	case *driver.Value:
		*vv = TimestampToTime(uint64(v))

		return nil
	case *uint64:
		*vv = uint64(v)

		return nil
	default:

		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v timestampValue) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), TimestampToTime(uint64(v)).UTC().Format(LayoutTimestamp))
}

func (timestampValue) Type() types.Type {
	return types.Timestamp
}

func (v timestampValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Uint64Value{
			Uint64Value: uint64(v),
		},
	}
}

// TimestampValue makes ydb timestamp value by given microseconds since Epoch
func TimestampValue(v uint64) timestampValue {
	return timestampValue(v)
}

func TimestampValueFromTime(t time.Time) timestampValue {
	return timestampValue(t.Sub(epoch) / time.Microsecond)
}

type timestamp64Value int64

func (v timestamp64Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *time.Time:
		*vv = Timestamp64ToTime(int64(v))

		return nil
	case *driver.Value:
		*vv = Timestamp64ToTime(int64(v))

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v timestamp64Value) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), Timestamp64ToTime(int64(v)).UTC().Format(LayoutTimestamp))
}

func (timestamp64Value) Type() types.Type {
	return types.Timestamp64
}

func (v timestamp64Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Int64Value{
			Int64Value: int64(v),
		},
	}
}

// Timestamp64Value makes ydb timestamp value by given signed microseconds around Epoch
func Timestamp64Value(v int64) timestamp64Value {
	return timestamp64Value(v)
}

func Timestamp64ValueFromTime(t time.Time) timestamp64Value {
	return timestamp64Value(t.UnixMicro())
}

type tupleValue struct {
	t     types.Type
	items []Value
}

func (v *tupleValue) TupleItems() []Value {
	return v.items
}

func (v *tupleValue) castTo(dst any) error {
	if len(v.items) == 1 {
		return v.items[0].castTo(dst)
	}

	switch dstValue := dst.(type) {
	case *driver.Value:
		*dstValue = v

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, dstValue,
		))
	}
}

func (v *tupleValue) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteByte('(')
	for i, item := range v.items {
		if i != 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteString(item.Yql())
	}
	buffer.WriteByte(')')

	return buffer.String()
}

func (v *tupleValue) Type() types.Type {
	return v.t
}

func (v *tupleValue) toYDB() *Ydb.Value {
	var items []Value
	if v != nil {
		items = v.items
	}
	vvv := &Ydb.Value{}

	for _, vv := range items {
		vvv.Items = append(vvv.GetItems(), vv.toYDB())
	}

	return vvv
}

func TupleValue(values ...Value) *tupleValue {
	tupleItems := make([]types.Type, 0, len(values))
	for _, v := range values {
		tupleItems = append(tupleItems, v.Type())
	}

	return &tupleValue{
		t:     types.NewTuple(tupleItems...),
		items: values,
	}
}

type tzDateValue string

func (v tzDateValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *string:
		*vv = string(v)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(string(v))

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v tzDateValue) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), string(v))
}

func (tzDateValue) Type() types.Type {
	return types.TzDate
}

func (v tzDateValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_TextValue{
			TextValue: string(v),
		},
	}
}

func TzDateValue(v string) tzDateValue {
	return tzDateValue(v)
}

func TzDateValueFromTime(t time.Time) tzDateValue {
	return tzDateValue(t.Format(LayoutTzDate) + "," + t.Location().String())
}

type tzDatetimeValue string

func (v tzDatetimeValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *string:
		*vv = string(v)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(string(v))

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v tzDatetimeValue) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), string(v))
}

func (tzDatetimeValue) Type() types.Type {
	return types.TzDatetime
}

func (v tzDatetimeValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_TextValue{
			TextValue: string(v),
		},
	}
}

func TzDatetimeValue(v string) tzDatetimeValue {
	return tzDatetimeValue(v)
}

func TzDatetimeValueFromTime(t time.Time) tzDatetimeValue {
	return tzDatetimeValue(t.Format(LayoutTzDatetime) + "," + t.Location().String())
}

type tzTimestampValue string

func (v tzTimestampValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *time.Time:
		t, err := TzTimestampToTime(string(v))
		if err != nil {
			return err
		}
		*vv = t

		return nil
	case *driver.Value:
		t, err := TzTimestampToTime(string(v))
		if err != nil {
			return err
		}
		*vv = t

		return nil
	case *string:
		*vv = string(v)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(string(v))

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v tzTimestampValue) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), string(v))
}

func (tzTimestampValue) Type() types.Type {
	return types.TzTimestamp
}

func (v tzTimestampValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_TextValue{
			TextValue: string(v),
		},
	}
}

func TzTimestampValue(v string) tzTimestampValue {
	return tzTimestampValue(v)
}

func TzTimestampValueFromTime(t time.Time) tzTimestampValue {
	return tzTimestampValue(t.Format(LayoutTzTimestamp) + "," + t.Location().String())
}

type uint8Value uint8

func (v uint8Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *uint8:
		*vv = uint8(v)

		return nil
	case *driver.Value:
		*vv = uint8(v)

		return nil
	case *string:
		*vv = strconv.FormatInt(int64(v), 10)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatInt(int64(v), 10))

		return nil
	case *uint64:
		*vv = uint64(v)

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	case *uint32:
		*vv = uint32(v)

		return nil
	case *int32:
		*vv = int32(v)

		return nil
	case *uint16:
		*vv = uint16(v)

		return nil
	case *int16:
		*vv = int16(v)

		return nil
	case *float64:
		*vv = float64(v)

		return nil
	case *float32:
		*vv = float32(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v uint8Value) Yql() string {
	return strconv.FormatUint(uint64(v), 10) + "ut"
}

func (uint8Value) Type() types.Type {
	return types.Uint8
}

func (v uint8Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Uint32Value{
			Uint32Value: uint32(v),
		},
	}
}

func Uint8Value(v uint8) uint8Value {
	return uint8Value(v)
}

type uint16Value uint16

func (v uint16Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *uint16:
		*vv = uint16(v)

		return nil
	case *driver.Value:
		*vv = uint16(v)

		return nil
	case *string:
		*vv = strconv.FormatUint(uint64(v), 10)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatUint(uint64(v), 10))

		return nil
	case *uint64:
		*vv = uint64(v)

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	case *uint32:
		*vv = uint32(v)

		return nil
	case *int32:
		*vv = int32(v)

		return nil
	case *float32:
		*vv = float32(v)

		return nil
	case *float64:
		*vv = float64(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v uint16Value) Yql() string {
	return strconv.FormatUint(uint64(v), 10) + "us"
}

func (uint16Value) Type() types.Type {
	return types.Uint16
}

func (v uint16Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Uint32Value{
			Uint32Value: uint32(v),
		},
	}
}

func Uint16Value(v uint16) uint16Value {
	return uint16Value(v)
}

type uint32Value uint32

func (v uint32Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *uint32:
		*vv = uint32(v)

		return nil
	case *driver.Value:
		*vv = uint32(v)

		return nil
	case *string:
		*vv = strconv.FormatUint(uint64(v), 10)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatUint(uint64(v), 10))

		return nil
	case *uint64:
		*vv = uint64(v)

		return nil
	case *int64:
		*vv = int64(v)

		return nil
	case *float64:
		*vv = float64(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v uint32Value) Yql() string {
	return strconv.FormatUint(uint64(v), 10) + "u"
}

func (uint32Value) Type() types.Type {
	return types.Uint32
}

func (v uint32Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Uint32Value{
			Uint32Value: uint32(v),
		},
	}
}

func Uint32Value(v uint32) uint32Value {
	return uint32Value(v)
}

type uint64Value uint64

func (v uint64Value) castTo(dst any) error {
	switch vv := dst.(type) {
	case *uint64:
		*vv = uint64(v)

		return nil
	case *driver.Value:
		*vv = uint64(v)

		return nil
	case *string:
		*vv = strconv.FormatUint(uint64(v), 10)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(strconv.FormatUint(uint64(v), 10))

		return nil

	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v uint64Value) Yql() string {
	return strconv.FormatUint(uint64(v), 10) + "ul"
}

func (uint64Value) Type() types.Type {
	return types.Uint64
}

func (v uint64Value) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_Uint64Value{
			Uint64Value: uint64(v),
		},
	}
}

func Uint64Value(v uint64) uint64Value {
	return uint64Value(v)
}

type textValue string

func (v textValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *string:
		*vv = string(v)

		return nil
	case *driver.Value:
		*vv = string(v)

		return nil
	case *[]byte:
		*vv = xstring.ToBytes(string(v))

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%q)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v textValue) Yql() string {
	return fmt.Sprintf("%qu", string(v))
}

func (textValue) Type() types.Type {
	return types.Text
}

func (v textValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_TextValue{
			TextValue: string(v),
		},
	}
}

func TextValue(v string) textValue {
	return textValue(v)
}

type UUIDIssue1501FixedBytesWrapper struct {
	val [16]byte
}

func NewUUIDIssue1501FixedBytesWrapper(val [16]byte) UUIDIssue1501FixedBytesWrapper {
	return UUIDIssue1501FixedBytesWrapper{val: val}
}

// PublicRevertReorderForIssue1501 needs for fix uuid when it was good stored in DB,
// but read as reordered. It may happen within migration period.
func (w UUIDIssue1501FixedBytesWrapper) PublicRevertReorderForIssue1501() uuid.UUID {
	return uuid.UUID(uuidFixBytesOrder(w.val))
}

func (w UUIDIssue1501FixedBytesWrapper) AsBytesArray() [16]byte {
	return w.val
}

func (w UUIDIssue1501FixedBytesWrapper) AsBytesSlice() []byte {
	return w.val[:]
}

func (w UUIDIssue1501FixedBytesWrapper) AsBrokenString() string {
	return string(w.val[:])
}

type uuidValue struct {
	value               uuid.UUID
	reproduceStorageBug bool
}

func (v *uuidValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *uuid.UUID:
		*vv = v.value

		return nil
	case *driver.Value:
		*vv = v.value

		return nil
	case *string:
		return ErrIssue1501BadUUID
	case *[]byte:
		return ErrIssue1501BadUUID
	case *[16]byte:
		return ErrIssue1501BadUUID
	case *UUIDIssue1501FixedBytesWrapper:
		*vv = NewUUIDIssue1501FixedBytesWrapper(uuidReorderBytesForReadWithBug(v.value))

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v *uuidValue) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString(v.Type().Yql())
	buffer.WriteByte('(')
	buffer.WriteByte('"')
	buffer.WriteString(v.value.String())
	buffer.WriteByte('"')
	buffer.WriteByte(')')

	return buffer.String()
}

func (*uuidValue) Type() types.Type {
	return types.UUID
}

func (v *uuidValue) toYDB() *Ydb.Value {
	if v.reproduceStorageBug {
		return v.toYDBWithBug()
	}

	var low, high uint64
	if v != nil {
		low, high = UUIDToHiLoPair(v.value)
	}

	return &Ydb.Value{
		High_128: high,
		Value: &Ydb.Value_Low_128{
			Low_128: low,
		},
	}
}

func UUIDToHiLoPair(id uuid.UUID) (low, high uint64) {
	bytes := uuidDirectBytesToLe(id)
	low = binary.LittleEndian.Uint64(bytes[0:8])
	high = binary.LittleEndian.Uint64(bytes[8:16])

	return low, high
}

func (v *uuidValue) toYDBWithBug() *Ydb.Value {
	var bytes [16]byte
	if v != nil {
		bytes = v.value
	}

	return &Ydb.Value{
		High_128: binary.BigEndian.Uint64(bytes[0:8]),
		Value: &Ydb.Value_Low_128{
			Low_128: binary.BigEndian.Uint64(bytes[8:16]),
		},
	}
}

func UUIDFromYDBPair(high uint64, low uint64) *uuidValue {
	var res uuid.UUID
	binary.LittleEndian.PutUint64(res[:], low)
	binary.LittleEndian.PutUint64(res[8:], high)
	res = uuidLeBytesToDirect(res)

	return &uuidValue{value: res}
}

func Uuid(val uuid.UUID) *uuidValue { //nolint:revive
	return &uuidValue{value: val}
}

func UUIDWithIssue1501Value(v [16]byte) *uuidValue {
	return &uuidValue{value: v, reproduceStorageBug: true}
}

func uuidDirectBytesToLe(direct [16]byte) [16]byte {
	// ordered as uuid bytes le in python
	// https://docs.python.org/3/library/uuid.html#uuid.UUID.bytes_le
	var le [16]byte
	le[0] = direct[3]
	le[1] = direct[2]
	le[2] = direct[1]
	le[3] = direct[0]
	le[4] = direct[5]
	le[5] = direct[4]
	le[6] = direct[7]
	le[7] = direct[6]
	le[8] = direct[8]
	le[9] = direct[9]
	le[10] = direct[10]
	le[11] = direct[11]
	le[12] = direct[12]
	le[13] = direct[13]
	le[14] = direct[14]
	le[15] = direct[15]

	return le
}

func uuidLeBytesToDirect(direct [16]byte) [16]byte {
	// ordered as uuid bytes le in python
	// https://docs.python.org/3/library/uuid.html#uuid.UUID.bytes_le
	var le [16]byte
	le[3] = direct[0]
	le[2] = direct[1]
	le[1] = direct[2]
	le[0] = direct[3]
	le[5] = direct[4]
	le[4] = direct[5]
	le[7] = direct[6]
	le[6] = direct[7]
	le[8] = direct[8]
	le[9] = direct[9]
	le[10] = direct[10]
	le[11] = direct[11]
	le[12] = direct[12]
	le[13] = direct[13]
	le[14] = direct[14]
	le[15] = direct[15]

	return le
}

func uuidReorderBytesForReadWithBug(val [16]byte) [16]byte {
	var res [16]byte
	res[0] = val[15]
	res[1] = val[14]
	res[2] = val[13]
	res[3] = val[12]
	res[4] = val[11]
	res[5] = val[10]
	res[6] = val[9]
	res[7] = val[8]
	res[8] = val[6]
	res[9] = val[7]
	res[10] = val[4]
	res[11] = val[5]
	res[12] = val[0]
	res[13] = val[1]
	res[14] = val[2]
	res[15] = val[3]

	return res
}

// uuidFixBytesOrder is reverse for uuidReorderBytesForReadWithBug
func uuidFixBytesOrder(val [16]byte) [16]byte {
	var res [16]byte
	res[0] = val[12]
	res[1] = val[13]
	res[2] = val[14]
	res[3] = val[15]
	res[4] = val[10]
	res[5] = val[11]
	res[6] = val[8]
	res[7] = val[9]
	res[8] = val[7]
	res[9] = val[6]
	res[10] = val[5]
	res[11] = val[4]
	res[12] = val[3]
	res[13] = val[2]
	res[14] = val[1]
	res[15] = val[0]

	return res
}

type variantValue struct {
	innerType types.Type
	value     Value
	idx       uint32
}

func (v *variantValue) Variant() (name string, index uint32) {
	switch t := v.innerType.(type) {
	case *types.VariantStruct:
		return t.Field(int(v.idx)).Name, v.idx
	default:
		return "", v.idx
	}
}

func (v *variantValue) Value() Value {
	return v.value
}

func (v *variantValue) castTo(dst any) error {
	switch dstValue := dst.(type) {
	case *driver.Value:
		*dstValue = v

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, dstValue,
		))
	}
}

func (v *variantValue) Yql() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString("Variant(")
	buffer.WriteString(v.value.Yql())
	buffer.WriteByte(',')
	switch t := v.innerType.(type) {
	case *types.VariantStruct:
		fmt.Fprintf(buffer, "%q", t.Field(int(v.idx)).Name)
	case *types.VariantTuple:
		fmt.Fprint(buffer, "\""+strconv.FormatUint(uint64(v.idx), 10)+"\"")
	}
	buffer.WriteByte(',')
	buffer.WriteString(v.Type().Yql())
	buffer.WriteByte(')')

	return buffer.String()
}

func (v *variantValue) Type() types.Type {
	return v.innerType
}

func (v *variantValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_NestedValue{
			NestedValue: v.value.toYDB(),
		},
		VariantIndex: v.idx,
	}
}

func VariantValueTuple(v Value, idx uint32, t types.Type) *variantValue {
	if tt, has := t.(*types.Tuple); has {
		t = types.NewVariantTuple(tt.InnerTypes()...)
	}

	return &variantValue{
		innerType: t,
		value:     v,
		idx:       idx,
	}
}

func VariantValueStruct(v Value, name string, t types.Type) *variantValue {
	var idx int
	switch tt := t.(type) {
	case *types.Struct:
		fields := tt.Fields()
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].Name < fields[j].Name
		})
		idx = sort.Search(len(fields), func(i int) bool {
			return fields[i].Name >= name
		})
		t = types.NewVariantStruct(fields...)
	case *types.VariantStruct:
		fields := tt.Fields()
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].Name < fields[j].Name
		})
		idx = sort.Search(len(fields), func(i int) bool {
			return fields[i].Name >= name
		})
	}

	return &variantValue{
		innerType: t,
		value:     v,
		idx:       uint32(idx),
	}
}

type voidValue struct{}

func (v voidValue) castTo(dst any) error {
	switch dstValue := dst.(type) {
	case *driver.Value:
		*dstValue = v

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, dstValue,
		))
	}
}

func (v voidValue) Yql() string {
	return v.Type().Yql() + "()"
}

var (
	_voidValueType = types.Void{}
	_voidValue     = &Ydb.Value{
		Value: new(Ydb.Value_NullFlagValue),
	}
)

func (voidValue) Type() types.Type {
	return _voidValueType
}

func (voidValue) toYDB() *Ydb.Value {
	return _voidValue
}

func VoidValue() voidValue {
	return voidValue{}
}

type ysonValue []byte

func (v ysonValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *string:
		*vv = xstring.FromBytes(v)

		return nil
	case *[]byte:
		*vv = v

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v ysonValue) Yql() string {
	return fmt.Sprintf("%s(%q)", v.Type().Yql(), string(v))
}

func (ysonValue) Type() types.Type {
	return types.YSON
}

func (v ysonValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_BytesValue{
			BytesValue: v,
		},
	}
}

func YSONValue(v []byte) ysonValue {
	return v
}

//nolint:funlen
func zeroPrimitiveValue(t types.Primitive) Value {
	switch t {
	case types.Bool:
		return BoolValue(false)

	case types.Int8:
		return Int8Value(0)

	case types.Uint8:
		return Uint8Value(0)

	case types.Int16:
		return Int16Value(0)

	case types.Uint16:
		return Uint16Value(0)

	case types.Int32:
		return Int32Value(0)

	case types.Uint32:
		return Uint32Value(0)

	case types.Int64:
		return Int64Value(0)

	case types.Uint64:
		return Uint64Value(0)

	case types.Float:
		return FloatValue(0)

	case types.Double:
		return DoubleValue(0)

	case types.Date:
		return DateValue(0)

	case types.Datetime:
		return DatetimeValue(0)

	case types.Timestamp:
		return TimestampValue(0)

	case types.Interval:
		return IntervalValue(0)

	case types.Text:
		return TextValue("")

	case types.YSON:
		return YSONValue([]byte(""))

	case types.JSON:
		return JSONValue("")

	case types.JSONDocument:
		return JSONDocumentValue("")

	case types.DyNumber:
		return DyNumberValue("0")

	case types.TzDate:
		return TzDateValue("")

	case types.TzDatetime:
		return TzDatetimeValue("")

	case types.TzTimestamp:
		return TzTimestampValue("")

	case types.Bytes:
		return BytesValue([]byte{})

	case types.UUID:
		return UUIDWithIssue1501Value([16]byte{})

	default:
		panic(fmt.Sprintf("uncovered primitive type '%T'", t))
	}
}

func ZeroValue(t types.Type) Value {
	switch t := t.(type) {
	case types.Primitive:
		return zeroPrimitiveValue(t)

	case types.Optional:
		return NullValue(t.InnerType())

	case *types.Void:
		return VoidValue()

	case *types.List, *types.EmptyList:
		return &listValue{
			t: t,
		}
	case *types.Set:
		return &setValue{
			t: t,
		}
	case *types.Dict:
		return &dictValue{
			t: t.ValueType(),
		}
	case *types.EmptyDict:
		return &dictValue{
			t: t,
		}
	case *types.Tuple:
		return TupleValue(func() []Value {
			innerTypes := t.InnerTypes()
			values := make([]Value, len(innerTypes))
			for i, tt := range innerTypes {
				values[i] = ZeroValue(tt)
			}

			return values
		}()...)
	case *types.Struct:
		return StructValue(func() []StructValueField {
			fields := t.Fields()
			values := make([]StructValueField, len(fields))
			for i := range fields {
				values[i] = StructValueField{
					Name: fields[i].Name,
					V:    ZeroValue(fields[i].T),
				}
			}

			return values
		}()...)
	case *types.Decimal:
		return DecimalValue([16]byte{}, decimalPrecision, decimalScale)

	default:
		panic(fmt.Sprintf("type '%T' have not a zero value", t))
	}
}

type bytesValue []byte

func (v bytesValue) castTo(dst any) error {
	switch vv := dst.(type) {
	case *[]byte:
		*vv = v

		return nil
	case *driver.Value:
		*vv = []byte(v)

		return nil
	case *string:
		*vv = xstring.FromBytes(v)

		return nil
	default:
		return xerrors.WithStackTrace(fmt.Errorf(
			"%w '%s(%+v)' to '%T' destination",
			ErrCannotCast, v.Type().Yql(), v, vv,
		))
	}
}

func (v bytesValue) Yql() string {
	return fmt.Sprintf("%q", string(v))
}

func (bytesValue) Type() types.Type {
	return types.Bytes
}

func (v bytesValue) toYDB() *Ydb.Value {
	return &Ydb.Value{
		Value: &Ydb.Value_BytesValue{
			BytesValue: v,
		},
	}
}

func BytesValue(v []byte) bytesValue {
	return v
}

var _ Value = (*protobufValue)(nil)

type protobufValue struct {
	pb *Ydb.TypedValue
}

func (v protobufValue) Type() types.Type {
	return types.FromProtobuf(v.pb.GetType())
}

func (v protobufValue) Yql() string {
	return FromYDB(v.pb.GetType(), v.pb.GetValue()).Yql()
}

func (v protobufValue) castTo(dst any) error {
	switch x := dst.(type) {
	case *Ydb.TypedValue:
		x.Type = v.pb.GetType()
		x.Value = v.pb.GetValue()

		return nil
	default:
		return xerrors.WithStackTrace(ErrCannotCast)
	}
}

func (v protobufValue) toYDB() *Ydb.Value {
	return v.pb.GetValue()
}

func FromProtobuf(pb *Ydb.TypedValue) protobufValue {
	return protobufValue{pb: pb}
}
