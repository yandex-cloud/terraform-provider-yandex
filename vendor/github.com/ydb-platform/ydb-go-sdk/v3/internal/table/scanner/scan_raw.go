package scanner

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/decimal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

type rawConverter struct {
	*valueScanner
}

func (s *rawConverter) String() (v []byte) {
	s.unwrap()

	return s.bytes()
}

func (s *rawConverter) HasItems() bool {
	return s.hasItems()
}

func (s *rawConverter) HasNextItem() bool {
	return s.hasItems() && s.nextItem < len(s.row.GetItems())
}

func (s *rawConverter) Path() string {
	var buf bytes.Buffer
	_, _ = s.WritePathTo(&buf)

	return buf.String()
}

func (s *rawConverter) WritePathTo(w io.Writer) (n int64, err error) {
	for sp := 0; sp < s.stack.size(); sp++ {
		if sp > 0 {
			var m int
			m, err = io.WriteString(w, ".")
			if err != nil {
				return n, xerrors.WithStackTrace(err)
			}
			n += int64(m)
		}
		x := s.stack.get(sp)
		s := x.name
		if s == "" {
			s = strconv.Itoa(x.i)
		}
		var m int
		m, err = io.WriteString(w, s)
		if err != nil {
			return n, xerrors.WithStackTrace(err)
		}
		n += int64(m)
	}

	return n, nil
}

func (s *rawConverter) Type() types.Type {
	return s.getType()
}

func (s *rawConverter) Bool() (v bool) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.bool()
}

func (s *rawConverter) Int8() (v int8) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.int8()
}

func (s *rawConverter) Uint8() (v uint8) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.uint8()
}

func (s *rawConverter) Int16() (v int16) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.int16()
}

func (s *rawConverter) Uint16() (v uint16) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.uint16()
}

func (s *rawConverter) Int32() (v int32) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.int32()
}

func (s *rawConverter) Uint32() (v uint32) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.uint32()
}

func (s *rawConverter) Int64() (v int64) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.int64()
}

func (s *rawConverter) Uint64() (v uint64) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.uint64()
}

func (s *rawConverter) Float() (v float32) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.float()
}

func (s *rawConverter) Double() (v float64) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.double()
}

func (s *rawConverter) Date() (v time.Time) {
	s.unwrap()

	return value.DateToTime(s.uint32())
}

func (s *rawConverter) Datetime() (v time.Time) {
	s.unwrap()

	return value.DatetimeToTime(s.uint32())
}

func (s *rawConverter) Timestamp() (v time.Time) {
	s.unwrap()

	return value.TimestampToTime(s.uint64())
}

func (s *rawConverter) Interval() (v time.Duration) {
	s.unwrap()

	return value.IntervalToDuration(s.int64())
}

func (s *rawConverter) TzDate() (v time.Time) {
	s.unwrap()
	if s.isNull() {
		return
	}
	src, err := value.TzDateToTime(s.text())
	if err != nil {
		_ = s.errorf(0, "rawConverter.TzDate(): %w", err)
	}

	return src
}

func (s *rawConverter) TzDatetime() (v time.Time) {
	s.unwrap()
	if s.isNull() {
		return
	}
	src, err := value.TzDatetimeToTime(s.text())
	if err != nil {
		_ = s.errorf(0, "rawConverter.TzDatetime(): %w", err)
	}

	return src
}

func (s *rawConverter) TzTimestamp() (v time.Time) {
	s.unwrap()
	if s.isNull() {
		return
	}
	src, err := value.TzTimestampToTime(s.text())
	if err != nil {
		_ = s.errorf(0, "rawConverter.TzTimestamp(): %w", err)
	}

	return src
}

func (s *rawConverter) UTF8() (v string) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.text()
}

func (s *rawConverter) YSON() (v []byte) {
	s.unwrap()

	return s.bytes()
}

func (s *rawConverter) JSON() (v []byte) {
	s.unwrap()

	return xstring.ToBytes(s.text())
}

func (s *rawConverter) JSONDocument() (v []byte) {
	s.unwrap()

	return xstring.ToBytes(s.text())
}

// removed for https://github.com/ydb-platform/ydb-go-sdk/issues/1501
//func (s *rawConverter) UUID() (v [16]byte) {
//	return s.uuidBytesWithIssue1501().AsBytesArray()
//}

func (s *rawConverter) UUIDWithIssue1501() (v [16]byte) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.uint128()
}

func (s *rawConverter) UUIDTyped() (v uuid.UUID) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.uuid()
}

func (s *rawConverter) DyNumber() (v string) {
	if s.Err() != nil {
		return
	}
	s.unwrap()

	return s.text()
}

func (s *rawConverter) Any() interface{} {
	return s.any()
}

// Value returns current item under scan as value
func (s *rawConverter) Value() value.Value {
	if s.Err() != nil {
		return nil
	}
	s.unwrap()

	return s.value()
}

func (s *rawConverter) AssertType(t types.Type) bool {
	return s.assertCurrentTypeIs(t)
}

func (s *rawConverter) Null() {
	if s.Err() != nil || !s.assertCurrentTypeNullable() {
		return
	}
	s.null()
}

func (s *rawConverter) IsNull() bool {
	if s.Err() != nil {
		return false
	}

	return s.isNull()
}

func (s *rawConverter) IsOptional() bool {
	if s.Err() != nil {
		return false
	}

	return s.isCurrentTypeOptional()
}

// --------non-primitive---------

func (s *rawConverter) ListIn() (size int) {
	if s.Err() != nil {
		return 0
	}
	x := s.stack.current()
	if s.assertTypeList(x.t) != nil {
		return s.itemsIn()
	}

	return 0
}

func (s *rawConverter) ListItem(i int) {
	if s.Err() != nil {
		return
	}
	p := s.stack.parent()
	if !s.itemsBoundsCheck(p.v.GetItems(), i) {
		return
	}
	if t := s.assertTypeList(p.t); t != nil {
		s.stack.set(item{
			i: i,
			t: t.ListType.GetItem(),
			v: p.v.GetItems()[i],
		})
	}
}

func (s *rawConverter) ListOut() {
	if s.Err() != nil {
		return
	}
	p := s.stack.parent()
	if t := s.assertTypeList(p.t); t != nil {
		s.itemsOut()
	}
}

func (s *rawConverter) TupleIn() (size int) {
	if s.Err() != nil {
		return 0
	}
	x := s.stack.current()
	if s.assertTypeTuple(x.t) != nil {
		return s.itemsIn()
	}

	return 0
}

func (s *rawConverter) TupleItem(i int) {
	if s.Err() != nil {
		return
	}
	p := s.stack.parent()
	if !s.itemsBoundsCheck(p.v.GetItems(), i) {
		return
	}
	if t := s.assertTypeTuple(p.t); t != nil {
		s.stack.set(item{
			i: i,
			t: t.TupleType.GetElements()[i],
			v: p.v.GetItems()[i],
		})
	}
}

func (s *rawConverter) TupleOut() {
	if s.Err() != nil {
		return
	}
	p := s.stack.parent()
	if t := s.assertTypeTuple(p.t); t != nil {
		s.itemsOut()
	}
}

func (s *rawConverter) StructIn() (size int) {
	if s.Err() != nil {
		return 0
	}
	x := s.stack.current()
	if s.assertTypeStruct(x.t) != nil {
		return s.itemsIn()
	}

	return 0
}

func (s *rawConverter) StructField(i int) (name string) {
	if s.Err() != nil {
		return
	}
	p := s.stack.parent()
	if !s.itemsBoundsCheck(p.v.GetItems(), i) {
		return
	}
	if t := s.assertTypeStruct(p.t); t != nil {
		m := t.StructType.GetMembers()[i]
		name = m.GetName()
		s.stack.set(item{
			name: m.GetName(),
			i:    i,
			t:    m.GetType(),
			v:    p.v.GetItems()[i],
		})
	}

	return
}

func (s *rawConverter) StructOut() {
	if s.Err() != nil {
		return
	}
	p := s.stack.parent()
	if t := s.assertTypeStruct(p.t); t != nil {
		s.itemsOut()
	}
}

func (s *rawConverter) DictIn() (size int) {
	if s.Err() != nil {
		return 0
	}
	x := s.stack.current()
	if s.assertTypeDict(x.t) != nil {
		return s.pairsIn()
	}

	return 0
}

func (s *rawConverter) DictKey(i int) {
	if s.Err() != nil {
		return
	}
	p := s.stack.parent()
	if !s.pairsBoundsCheck(p.v.GetPairs(), i) {
		return
	}
	if t := s.assertTypeDict(p.t); t != nil {
		s.stack.set(item{
			i: i,
			t: t.DictType.GetKey(),
			v: p.v.GetPairs()[i].GetKey(),
		})
	}
}

func (s *rawConverter) DictPayload(i int) {
	if s.Err() != nil {
		return
	}
	p := s.stack.parent()
	if !s.pairsBoundsCheck(p.v.GetPairs(), i) {
		return
	}
	if t := s.assertTypeDict(p.t); t != nil {
		s.stack.set(item{
			i: i,
			t: t.DictType.GetPayload(),
			v: p.v.GetPairs()[i].GetPayload(),
		})
	}
}

func (s *rawConverter) DictOut() {
	if s.Err() != nil {
		return
	}
	p := s.stack.parent()
	if t := s.assertTypeDict(p.t); t != nil {
		s.pairsOut()
	}
}

func (s *rawConverter) Variant() (name string, index uint32) {
	if s.Err() != nil {
		return
	}
	x := s.stack.current()
	t := s.assertTypeVariant(x.t)
	if t == nil {
		return
	}
	v, index := s.variant()
	if v == nil {
		return
	}
	name, typ := s.unwrapVariantType(t, index)
	s.stack.scanItem.v = nil
	s.stack.set(item{
		name: name,
		i:    int(index),
		t:    typ,
		v:    v,
	})

	return name, index
}

func (s *rawConverter) Unwrap() {
	if s.Err() != nil {
		return
	}
	x := s.stack.current()
	t := s.assertTypeOptional(x.t)
	if t == nil {
		return
	}
	v := x.v
	if isOptional(t.OptionalType.GetItem()) {
		v = s.unwrapValue()
	}
	s.stack.enter()
	s.stack.set(item{
		name: "*",
		t:    t.OptionalType.GetItem(),
		v:    v,
	})
}

func (s *rawConverter) Decimal(t types.Type) (v [16]byte) {
	if s.Err() != nil {
		return
	}
	s.unwrap()
	if !s.assertCurrentTypeDecimal(t) {
		return
	}

	return s.uint128()
}

func (s *rawConverter) UnwrapDecimal() decimal.Decimal {
	if s.Err() != nil {
		return decimal.Decimal{}
	}
	s.unwrap()
	d := s.assertTypeDecimal(s.stack.current().t)
	if d == nil {
		return decimal.Decimal{}
	}

	return decimal.Decimal{
		Bytes:     s.uint128(),
		Precision: d.DecimalType.GetPrecision(),
		Scale:     d.DecimalType.GetScale(),
	}
}

func (s *rawConverter) IsDecimal() bool {
	if s.Err() != nil {
		return false
	}

	return s.isCurrentTypeDecimal()
}

func isEqualDecimal(d *Ydb.DecimalType, t types.Type) bool {
	w, ok := t.(*types.Decimal)
	if !ok {
		panic(fmt.Sprintf("unsupported type conversion from %T to *types.Decimal", w))
	}

	return d.GetPrecision() == w.Precision() && d.GetScale() == w.Scale()
}

func (s *rawConverter) isCurrentTypeDecimal() bool {
	c := s.stack.current()
	_, ok := c.t.GetType().(*Ydb.Type_DecimalType)

	return ok
}

func (s *rawConverter) unwrapVariantType(typ *Ydb.Type_VariantType, index uint32) (name string, t *Ydb.Type) {
	i := int(index)
	switch x := typ.VariantType.GetType().(type) {
	case *Ydb.VariantType_TupleItems:
		if i >= len(x.TupleItems.GetElements()) {
			_ = s.errorf(0, "unimplemented")

			return
		}

		return "", x.TupleItems.GetElements()[i]

	case *Ydb.VariantType_StructItems:
		if i >= len(x.StructItems.GetMembers()) {
			_ = s.errorf(0, "unimplemented")

			return
		}
		m := x.StructItems.GetMembers()[i]

		return m.GetName(), m.GetType()

	default:
		panic("unexpected variant items types")
	}
}

func (s *rawConverter) variant() (v *Ydb.Value, index uint32) {
	v = s.unwrapValue()
	if v == nil {
		return
	}
	x := s.stack.current() // Is not nil if unwrapValue succeeded.
	index = x.v.GetVariantIndex()

	return
}

func (s *rawConverter) itemsIn() int {
	x := s.stack.current()
	if x.isEmpty() {
		return -1
	}
	s.stack.enter()

	return len(x.v.GetItems())
}

func (s *rawConverter) itemsOut() {
	s.stack.leave()
}

func (s *rawConverter) itemsBoundsCheck(xs []*Ydb.Value, i int) bool {
	return s.boundsCheck(len(xs), i)
}

func (s *rawConverter) pairsIn() int {
	x := s.stack.current()
	if x.isEmpty() {
		return -1
	}
	s.stack.enter()

	return len(x.v.GetPairs())
}

func (s *rawConverter) pairsOut() {
	s.stack.leave()
}

func (s *rawConverter) pairsBoundsCheck(xs []*Ydb.ValuePair, i int) bool {
	return s.boundsCheck(len(xs), i)
}

func (s *rawConverter) boundsCheck(n, i int) bool {
	if i < 0 || n <= i {
		s.boundsError(n, i)

		return false
	}

	return true
}

func (s *valueScanner) assertTypeOptional(typ *Ydb.Type) (t *Ydb.Type_OptionalType) {
	if t, _ = typ.GetType().(*Ydb.Type_OptionalType); t == nil {
		s.typeError(typ.GetType(), t)
	}

	return
}

func (s *rawConverter) assertCurrentTypeNullable() bool {
	c := s.stack.current()
	if isOptional(c.t) {
		return true
	}
	p := s.stack.parent()
	if isOptional(p.t) {
		return true
	}
	_ = s.errorf(
		1,
		"not nullable types at %q: %s (%d %s %s)",
		s.Path(),
		s.Type(),
		s.stack.size(),
		c.t,
		p.t,
	)

	return false
}

func (s *rawConverter) assertCurrentTypeIs(t types.Type) bool {
	c := s.stack.current()
	act := types.TypeFromYDB(c.t)
	if !types.Equal(act, t) {
		_ = s.errorf(
			1,
			"unexpected types at %q %s: %s; want %s",
			s.Path(),
			s.Type(),
			act,
			t,
		)

		return false
	}

	return true
}

func (s *rawConverter) assertCurrentTypeDecimal(t types.Type) bool {
	d := s.assertTypeDecimal(s.stack.current().t)
	if d == nil {
		return false
	}
	if !isEqualDecimal(d.DecimalType, t) {
		s.decimalTypeError(t)

		return false
	}

	return true
}

func (s *rawConverter) assertTypeList(typ *Ydb.Type) (t *Ydb.Type_ListType) {
	if t, _ = typ.GetType().(*Ydb.Type_ListType); t == nil {
		s.typeError(typ.GetType(), t)
	}

	return
}

func (s *rawConverter) assertTypeTuple(typ *Ydb.Type) (t *Ydb.Type_TupleType) {
	if t, _ = typ.GetType().(*Ydb.Type_TupleType); t == nil {
		s.typeError(typ.GetType(), t)
	}

	return
}

func (s *rawConverter) assertTypeStruct(typ *Ydb.Type) (t *Ydb.Type_StructType) {
	if t, _ = typ.GetType().(*Ydb.Type_StructType); t == nil {
		s.typeError(typ.GetType(), t)
	}

	return
}

func (s *rawConverter) assertTypeDict(typ *Ydb.Type) (t *Ydb.Type_DictType) {
	if t, _ = typ.GetType().(*Ydb.Type_DictType); t == nil {
		s.typeError(typ.GetType(), t)
	}

	return
}

func (s *rawConverter) assertTypeDecimal(typ *Ydb.Type) (t *Ydb.Type_DecimalType) {
	if t, _ = typ.GetType().(*Ydb.Type_DecimalType); t == nil {
		s.typeError(typ.GetType(), t)
	}

	return
}

func (s *rawConverter) assertTypeVariant(typ *Ydb.Type) (t *Ydb.Type_VariantType) {
	if t, _ = typ.GetType().(*Ydb.Type_VariantType); t == nil {
		s.typeError(typ.GetType(), t)
	}

	return
}

func (s *rawConverter) boundsError(n, i int) {
	_ = s.errorf(
		1, "index out of range: %d; have %d",
		i, n,
	)
}

func (s *rawConverter) decimalTypeError(t types.Type) {
	_ = s.errorf(
		1, "unexpected decimal types at %q %s: want %s",
		s.Path(), s.getType(), t,
	)
}

func nameIface(v interface{}) string {
	if v == nil {
		return "nil"
	}
	t := reflect.TypeOf(v)
	s := t.String()
	s = strings.TrimPrefix(s, "*Ydb.Value_")
	s = strings.TrimSuffix(s, "valueType")
	s = strings.TrimPrefix(s, "*Ydb.Type_")
	s = strings.TrimSuffix(s, "Type")

	return s
}
