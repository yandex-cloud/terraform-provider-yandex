package scanner

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/decimal"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/scanner"
	internalTypes "github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/indexed"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

type valueScanner struct {
	set                      *Ydb.ResultSet
	row                      *Ydb.Value
	converter                *rawConverter
	stack                    scanStack
	nextRow                  int
	nextItem                 int
	ignoreTruncated          bool
	markTruncatedAsRetryable bool

	columnIndexes []int

	errMtx xsync.RWMutex
	err    error
}

// ColumnCount returns number of columns in the current result set.
func (s *valueScanner) ColumnCount() int {
	if s.set == nil {
		return 0
	}

	return len(s.set.GetColumns())
}

// Columns allows to iterate over all columns of the current result set.
func (s *valueScanner) Columns(it func(options.Column)) {
	if s.set == nil {
		return
	}
	for _, m := range s.set.GetColumns() {
		it(options.Column{
			Name: m.GetName(),
			Type: internalTypes.TypeFromYDB(m.GetType()),
		})
	}
}

// RowCount returns number of rows in the result set.
func (s *valueScanner) RowCount() int {
	if s.set == nil {
		return 0
	}

	return len(s.set.GetRows())
}

// ItemCount returns number of items in the current row.
func (s *valueScanner) ItemCount() int {
	if s.row == nil {
		return 0
	}

	return len(s.row.GetItems())
}

// HasNextRow reports whether result row may be advanced.
// It may be useful to call HasNextRow() instead of NextRow() to look ahead
// without advancing the result rows.
func (s *valueScanner) HasNextRow() bool {
	return s.err == nil && s.set != nil && s.nextRow < len(s.set.GetRows())
}

// NextRow selects next row in the current result set.
// It returns false if there are no more rows in the result set.
func (s *valueScanner) NextRow() bool {
	if !s.HasNextRow() {
		return false
	}
	s.row = s.set.GetRows()[s.nextRow]
	s.nextRow++
	s.nextItem = 0
	s.stack.reset()

	return true
}

func (s *valueScanner) preScanChecks(lenValues int) (err error) {
	if s.columnIndexes != nil {
		if len(s.columnIndexes) != lenValues {
			return s.errorf(
				1,
				"scan row failed: count of values and column are different (%d != %d)",
				len(s.columnIndexes),
				lenValues,
			)
		}
	}
	if s.ColumnCount() < lenValues {
		panic(fmt.Sprintf("scan row failed: count of columns less then values (%d < %d)", s.ColumnCount(), lenValues))
	}
	if s.nextItem != 0 {
		panic("scan row failed: double scan per row")
	}

	return s.Err()
}

func (s *valueScanner) ScanWithDefaults(values ...indexed.Required) (err error) {
	if err = s.preScanChecks(len(values)); err != nil {
		return
	}
	for i := range values {
		if _, ok := values[i].(named.Value); ok {
			panic("dont use NamedValue with ScanWithDefaults. Use ScanNamed instead")
		}
		if s.columnIndexes == nil {
			if err = s.seekItemByID(i); err != nil {
				return
			}
		} else {
			if err = s.seekItemByID(s.columnIndexes[i]); err != nil {
				return
			}
		}
		if s.isCurrentTypeOptional() {
			s.scanOptional(values[i], true)
		} else {
			s.scanRequired(values[i])
		}
	}
	s.nextItem += len(values)

	return s.Err()
}

func (s *valueScanner) Scan(values ...indexed.RequiredOrOptional) (err error) {
	if err = s.preScanChecks(len(values)); err != nil {
		return
	}
	for i := range values {
		if _, ok := values[i].(named.Value); ok {
			panic("dont use NamedValue with Scan. Use ScanNamed instead")
		}
		if s.columnIndexes == nil {
			if err = s.seekItemByID(i); err != nil {
				return
			}
		} else {
			if err = s.seekItemByID(s.columnIndexes[i]); err != nil {
				return
			}
		}
		if s.isCurrentTypeOptional() {
			s.scanOptional(values[i], false)
		} else {
			s.scanRequired(values[i])
		}
	}
	s.nextItem += len(values)

	return s.Err()
}

func (s *valueScanner) ScanNamed(namedValues ...named.Value) error {
	if err := s.Err(); err != nil {
		return err
	}
	if s.ColumnCount() < len(namedValues) {
		panic(fmt.Sprintf("scan row failed: count of columns less then values (%d < %d)", s.ColumnCount(), len(namedValues)))
	}
	if s.nextItem != 0 {
		panic("scan row failed: double scan per row")
	}
	for i := range namedValues {
		if err := s.seekItemByName(namedValues[i].Name); err != nil {
			return err
		}
		switch t := namedValues[i].Type; t {
		case named.TypeRequired:
			s.scanRequired(namedValues[i].Value)
		case named.TypeOptional:
			s.scanOptional(namedValues[i].Value, false)
		case named.TypeOptionalWithUseDefault:
			s.scanOptional(namedValues[i].Value, true)
		default:
			panic(fmt.Sprintf("unknown type of named.valueType: %d", t))
		}
	}
	s.nextItem += len(namedValues)

	return s.Err()
}

// Truncated returns true if current result set has been truncated by server
func (s *valueScanner) Truncated() bool {
	if s.set == nil {
		_ = s.errorf(0, "there are no sets in the scanner")

		return false
	}

	return s.set.GetTruncated()
}

// Truncated returns true if current result set has been truncated by server
func (s *valueScanner) truncated() bool {
	if s.set == nil {
		return false
	}

	return s.set.GetTruncated()
}

// Err returns error caused Scanner to be broken.
func (s *valueScanner) Err() error {
	s.errMtx.RLock()
	defer s.errMtx.RUnlock()
	if s.err != nil {
		return s.err
	}
	if !s.ignoreTruncated && s.truncated() {
		err := xerrors.Wrap(
			fmt.Errorf("more than %d rows: %w", len(s.set.GetRows()), result.ErrTruncated),
		)
		if s.markTruncatedAsRetryable {
			err = xerrors.Retryable(err)
		}

		return xerrors.WithStackTrace(err)
	}

	return nil
}

// Must not be exported.
func (s *valueScanner) reset(set *Ydb.ResultSet, columnNames ...string) {
	s.set = set
	s.row = nil
	s.nextRow = 0
	s.nextItem = 0
	s.columnIndexes = nil
	s.setColumnIndexes(columnNames)
	s.stack.reset()
	s.converter = &rawConverter{
		valueScanner: s,
	}
}

func (s *valueScanner) path() string {
	buf := xstring.Buffer()
	defer buf.Free()
	_, _ = s.writePathTo(buf)

	return buf.String()
}

func (s *valueScanner) writePathTo(w io.Writer) (n int64, err error) {
	x := s.stack.current()
	st := x.name
	m, err := io.WriteString(w, st)
	if err != nil {
		return n, xerrors.WithStackTrace(err)
	}
	n += int64(m)

	return n, nil
}

func (s *valueScanner) getType() internalTypes.Type {
	x := s.stack.current()
	if x.isEmpty() {
		return nil
	}

	return internalTypes.TypeFromYDB(x.t)
}

func (s *valueScanner) hasItems() bool {
	return s.err == nil && s.set != nil && s.row != nil
}

func (s *valueScanner) seekItemByID(id int) error {
	if !s.hasItems() || id >= len(s.set.GetColumns()) {
		return s.notFoundColumnByIndex(id)
	}
	col := s.set.GetColumns()[id]
	s.stack.scanItem.name = col.GetName()
	s.stack.scanItem.t = col.GetType()
	s.stack.scanItem.v = s.row.GetItems()[id]

	return nil
}

func (s *valueScanner) seekItemByName(name string) error {
	if !s.hasItems() {
		return s.notFoundColumnName(name)
	}
	for i, c := range s.set.GetColumns() {
		if name != c.GetName() {
			continue
		}
		s.stack.scanItem.name = c.GetName()
		s.stack.scanItem.t = c.GetType()
		s.stack.scanItem.v = s.row.GetItems()[i]

		return s.Err()
	}

	return s.notFoundColumnName(name)
}

func (s *valueScanner) setColumnIndexes(columns []string) {
	if columns == nil {
		s.columnIndexes = nil

		return
	}
	s.columnIndexes = make([]int, len(columns))
	for i, col := range columns {
		found := false
		for j, c := range s.set.GetColumns() {
			if c.GetName() == col {
				s.columnIndexes[i] = j
				found = true

				break
			}
		}
		if !found {
			_ = s.noColumnError(col)

			return
		}
	}
}

// Any returns any primitive or optional value.
// Currently, it may return one of these types:
//
//		bool
//		int8
//		uint8
//		int16
//		uint16
//		int32
//		uint32
//		int64
//		uint64
//		float32
//		float64
//		[]byte
//		string
//		[16]byte
//	    uuid
//
//nolint:gocyclo,funlen
func (s *valueScanner) any() interface{} {
	x := s.stack.current()
	if s.Err() != nil || x.isEmpty() {
		return nil
	}

	if s.isNull() {
		return nil
	}

	if s.isCurrentTypeOptional() {
		s.unwrap()
		x = s.stack.current()
	}

	t := internalTypes.TypeFromYDB(x.t)
	p, primitive := t.(internalTypes.Primitive)
	if !primitive {
		return s.value()
	}

	switch p {
	case internalTypes.Bool:
		return s.bool()
	case internalTypes.Int8:
		return s.int8()
	case internalTypes.Uint8:
		return s.uint8()
	case internalTypes.Int16:
		return s.int16()
	case internalTypes.Uint16:
		return s.uint16()
	case internalTypes.Int32:
		return s.int32()
	case internalTypes.Float:
		return s.float()
	case internalTypes.Double:
		return s.double()
	case internalTypes.Bytes:
		return s.bytes()
	case internalTypes.UUID:
		// replace to good uuid on migration
		return s.uuidBytesWithIssue1501()
	case internalTypes.Uint32:
		return s.uint32()
	case internalTypes.Date:
		return value.DateToTime(s.uint32())
	case internalTypes.Datetime:
		return value.DatetimeToTime(s.uint32())
	case internalTypes.Uint64:
		return s.uint64()
	case internalTypes.Timestamp:
		return value.TimestampToTime(s.uint64())
	case internalTypes.Int64:
		return s.int64()
	case internalTypes.Interval:
		return value.IntervalToDuration(s.int64())
	case internalTypes.TzDate:
		src, err := value.TzDateToTime(s.text())
		if err != nil {
			_ = s.errorf(0, "valueScanner.any(): %w", err)
		}

		return src
	case internalTypes.TzDatetime:
		src, err := value.TzDatetimeToTime(s.text())
		if err != nil {
			_ = s.errorf(0, "valueScanner.any(): %w", err)
		}

		return src
	case internalTypes.TzTimestamp:
		src, err := value.TzTimestampToTime(s.text())
		if err != nil {
			_ = s.errorf(0, "valueScanner.any(): %w", err)
		}

		return src
	case internalTypes.Text, internalTypes.DyNumber:
		return s.text()
	case
		internalTypes.YSON,
		internalTypes.JSON,
		internalTypes.JSONDocument:
		return xstring.ToBytes(s.text())
	default:
		_ = s.errorf(0, "unknown primitive types")

		return nil
	}
}

// valueType returns current item under scan as ydb.valueType types
func (s *valueScanner) value() value.Value {
	x := s.stack.current()

	return value.FromYDB(x.t, x.v)
}

func (s *valueScanner) isCurrentTypeOptional() bool {
	c := s.stack.current()

	return isOptional(c.t)
}

func (s *valueScanner) isNull() bool {
	_, yes := s.stack.currentValue().(*Ydb.Value_NullFlagValue)

	return yes
}

// unwrap current item under scan interpreting it as Optional<Type> types
// ignores if type is not optional
func (s *valueScanner) unwrap() {
	if s.Err() != nil {
		return
	}

	t, _ := s.stack.currentType().(*Ydb.Type_OptionalType)
	if t == nil {
		return
	}

	if isOptional(t.OptionalType.GetItem()) {
		s.stack.scanItem.v = s.unwrapValue()
	}
	s.stack.scanItem.t = t.OptionalType.GetItem()
}

func (s *valueScanner) unwrapValue() (v *Ydb.Value) {
	x, _ := s.stack.currentValue().(*Ydb.Value_NestedValue)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.NestedValue
}

func (s *valueScanner) unwrapDecimal() decimal.Decimal {
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

func (s *valueScanner) assertTypeDecimal(typ *Ydb.Type) (t *Ydb.Type_DecimalType) {
	if t, _ = typ.GetType().(*Ydb.Type_DecimalType); t == nil {
		s.typeError(typ.GetType(), t)
	}

	return
}

func (s *valueScanner) bool() (v bool) {
	x, _ := s.stack.currentValue().(*Ydb.Value_BoolValue)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.BoolValue
}

func (s *valueScanner) int8() (v int8) {
	d := s.int32()
	if d < math.MinInt8 || math.MaxInt8 < d {
		_ = s.overflowError(d, v)

		return
	}

	return int8(d)
}

func (s *valueScanner) uint8() (v uint8) {
	d := s.uint32()
	if d > math.MaxUint8 {
		_ = s.overflowError(d, v)

		return
	}

	return uint8(d)
}

func (s *valueScanner) int16() (v int16) {
	d := s.int32()
	if d < math.MinInt16 || math.MaxInt16 < d {
		_ = s.overflowError(d, v)

		return
	}

	return int16(d)
}

func (s *valueScanner) uint16() (v uint16) {
	d := s.uint32()
	if d > math.MaxUint16 {
		_ = s.overflowError(d, v)

		return
	}

	return uint16(d)
}

func (s *valueScanner) int32() (v int32) {
	x, _ := s.stack.currentValue().(*Ydb.Value_Int32Value)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.Int32Value
}

func (s *valueScanner) uint32() (v uint32) {
	x, _ := s.stack.currentValue().(*Ydb.Value_Uint32Value)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.Uint32Value
}

func (s *valueScanner) int64() (v int64) {
	x, _ := s.stack.currentValue().(*Ydb.Value_Int64Value)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.Int64Value
}

func (s *valueScanner) uint64() (v uint64) {
	x, _ := s.stack.currentValue().(*Ydb.Value_Uint64Value)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.Uint64Value
}

func (s *valueScanner) float() (v float32) {
	x, _ := s.stack.currentValue().(*Ydb.Value_FloatValue)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.FloatValue
}

func (s *valueScanner) double() (v float64) {
	x, _ := s.stack.currentValue().(*Ydb.Value_DoubleValue)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.DoubleValue
}

func (s *valueScanner) bytes() (v []byte) {
	x, _ := s.stack.currentValue().(*Ydb.Value_BytesValue)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.BytesValue
}

func (s *valueScanner) text() (v string) {
	x, _ := s.stack.currentValue().(*Ydb.Value_TextValue)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.TextValue
}

func (s *valueScanner) low128() (v uint64) {
	x, _ := s.stack.currentValue().(*Ydb.Value_Low_128)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)

		return
	}

	return x.Low_128
}

func (s *valueScanner) uint128() (v [16]byte) {
	if s.stack.current().t.GetTypeId() == Ydb.Type_UUID {
		_ = s.errorf(0, "ydb: failed to scan uuid: %w", value.ErrIssue1501BadUUID)
	}

	c := s.stack.current()
	if c.isEmpty() {
		_ = s.errorf(0, "not implemented convert to [16]byte")

		return
	}
	lo := s.low128()
	hi := c.v.GetHigh_128()

	return value.BigEndianUint128(hi, lo)
}

func (s *valueScanner) uuidBytesWithIssue1501() (v types.UUIDBytesWithIssue1501Type) {
	c := s.stack.current()
	if c.isEmpty() {
		_ = s.errorf(0, "not implemented convert to [16]byte")

		return
	}
	lo := s.low128()
	hi := c.v.GetHigh_128()

	return value.NewUUIDIssue1501FixedBytesWrapper(value.BigEndianUint128(hi, lo))
}

func (s *valueScanner) uuid() uuid.UUID {
	c := s.stack.current()
	if c.isEmpty() {
		_ = s.errorf(0, "not implemented convert to [16]byte")

		return uuid.UUID{}
	}
	lo := s.low128()
	hi := c.v.GetHigh_128()

	val := value.UUIDFromYDBPair(hi, lo)

	var uuidVal uuid.UUID
	err := value.CastTo(val, &uuidVal)
	if err != nil {
		_ = s.errorf(0, "failed to cast internal uuid type to uuid: %w", err)

		return uuid.UUID{}
	}

	return uuidVal
}

func (s *valueScanner) null() {
	x, _ := s.stack.currentValue().(*Ydb.Value_NullFlagValue)
	if x == nil {
		s.valueTypeError(s.stack.currentValue(), x)
	}
}

func (s *valueScanner) setTime(dst *time.Time) {
	switch t := s.stack.current().t.GetTypeId(); t {
	case Ydb.Type_DATE:
		*dst = value.DateToTime(s.uint32())
	case Ydb.Type_DATETIME:
		*dst = value.DatetimeToTime(s.uint32())
	case Ydb.Type_TIMESTAMP:
		*dst = value.TimestampToTime(s.uint64())
	case Ydb.Type_TZ_DATE:
		src, err := value.TzDateToTime(s.text())
		if err != nil {
			_ = s.errorf(0, "valueScanner.setTime(): %w", err)
		}
		*dst = src
	case Ydb.Type_TZ_DATETIME:
		src, err := value.TzDatetimeToTime(s.text())
		if err != nil {
			_ = s.errorf(0, "valueScanner.setTime(): %w", err)
		}
		*dst = src
	case Ydb.Type_TZ_TIMESTAMP:
		src, err := value.TzTimestampToTime(s.text())
		if err != nil {
			_ = s.errorf(0, "valueScanner.setTime(): %w", err)
		}
		*dst = src
	default:
		_ = s.errorf(0, "valueScanner.setTime(): incorrect source types %s", t)
	}
}

func (s *valueScanner) setString(dst *string) {
	switch t := s.stack.current().t.GetTypeId(); t {
	case Ydb.Type_UUID:
		_ = s.errorf(0, "ydb: failed scan uuid: %w", value.ErrIssue1501BadUUID)
	case Ydb.Type_UTF8, Ydb.Type_DYNUMBER, Ydb.Type_YSON, Ydb.Type_JSON, Ydb.Type_JSON_DOCUMENT:
		*dst = s.text()
	case Ydb.Type_STRING:
		*dst = xstring.FromBytes(s.bytes())
	default:
		_ = s.errorf(0, "scan row failed: incorrect source types %s", t)
	}
}

func (s *valueScanner) setByte(dst *[]byte) {
	switch t := s.stack.current().t.GetTypeId(); t {
	case Ydb.Type_UUID:
		_ = s.errorf(0, "ydb: failed to scan uuid: %w", value.ErrIssue1501BadUUID)
	case Ydb.Type_UTF8, Ydb.Type_DYNUMBER, Ydb.Type_YSON, Ydb.Type_JSON, Ydb.Type_JSON_DOCUMENT:
		*dst = xstring.ToBytes(s.text())
	case Ydb.Type_STRING:
		*dst = s.bytes()
	default:
		_ = s.errorf(0, "scan row failed: incorrect source types %s", t)
	}
}

func (s *valueScanner) trySetByteArray(v interface{}, optional, def bool) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Ptr {
		if !optional {
			return false
		}
		if s.isNull() {
			rv.Set(reflect.Zero(rv.Type()))

			return true
		}
		if rv.IsZero() {
			nv := reflect.New(rv.Type().Elem())
			rv.Set(nv)
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Array {
		return false
	}
	if rv.Type().Elem().Kind() != reflect.Uint8 {
		return false
	}
	if def {
		rv.Set(reflect.Zero(rv.Type()))

		return true
	}
	var dst []byte
	s.setByte(&dst)
	if rv.Len() != len(dst) {
		return false
	}
	reflect.Copy(rv, reflect.ValueOf(dst))

	return true
}

//nolint:gocyclo,funlen
func (s *valueScanner) scanRequired(v interface{}) {
	switch v := v.(type) {
	case *bool:
		*v = s.bool()
	case *int8:
		*v = s.int8()
	case *int16:
		*v = s.int16()
	case *int:
		*v = int(s.int32())
	case *int32:
		*v = s.int32()
	case *int64:
		*v = s.int64()
	case *uint8:
		*v = s.uint8()
	case *uint16:
		*v = s.uint16()
	case *uint32:
		*v = s.uint32()
	case *uint:
		*v = uint(s.uint32())
	case *uint64:
		*v = s.uint64()
	case *float32:
		*v = s.float()
	case *float64:
		*v = s.double()
	case *time.Time:
		s.setTime(v)
	case *time.Duration:
		*v = value.IntervalToDuration(s.int64())
	case *string:
		s.setString(v)
	case *[]byte:
		s.setByte(v)
	case *[16]byte:
		*v = s.uint128()
	case *types.UUIDBytesWithIssue1501Type:
		*v = s.uuidBytesWithIssue1501()
	case *uuid.UUID:
		*v = s.uuid()
	case *interface{}:
		*v = s.any()
	case *value.Value:
		*v = s.value()
	case *decimal.Decimal:
		*v = s.unwrapDecimal()
	case scanner.Scanner:
		err := v.UnmarshalYDB(s.converter)
		if err != nil {
			_ = s.errorf(0, "ydb.Scanner error: %w", err)
		}
	case sql.Scanner:
		err := v.Scan(s.any())
		if err != nil {
			_ = s.errorf(0, "sql.Scanner error: %w", err)
		}
	case json.Unmarshaler:
		var err error
		switch s.getType() {
		case internalTypes.JSON:
			err = v.UnmarshalJSON(s.converter.JSON())
		case internalTypes.JSONDocument:
			err = v.UnmarshalJSON(s.converter.JSONDocument())
		default:
			_ = s.errorf(0, "ydb required type %T not unsupported for applying to json.Unmarshaler", s.getType())
		}
		if err != nil {
			_ = s.errorf(0, "json.Unmarshaler error: %w", err)
		}
	default:
		ok := s.trySetByteArray(v, false, false)
		if !ok {
			_ = s.errorf(0, "scan row failed: type %T is unknown", v)
		}
	}
}

//nolint:gocyclo, funlen
func (s *valueScanner) scanOptional(v interface{}, defaultValueForOptional bool) {
	if defaultValueForOptional {
		if s.isNull() {
			s.setDefaultValue(v)
		} else {
			s.unwrap()
			s.scanRequired(v)
		}

		return
	}
	switch v := v.(type) {
	case **bool:
		if s.isNull() {
			*v = nil
		} else {
			src := s.bool()
			*v = &src
		}
	case **int8:
		if s.isNull() {
			*v = nil
		} else {
			src := s.int8()
			*v = &src
		}
	case **int16:
		if s.isNull() {
			*v = nil
		} else {
			src := s.int16()
			*v = &src
		}
	case **int32:
		if s.isNull() {
			*v = nil
		} else {
			src := s.int32()
			*v = &src
		}
	case **int:
		if s.isNull() {
			*v = nil
		} else {
			src := int(s.int32())
			*v = &src
		}
	case **int64:
		if s.isNull() {
			*v = nil
		} else {
			src := s.int64()
			*v = &src
		}
	case **uint8:
		if s.isNull() {
			*v = nil
		} else {
			src := s.uint8()
			*v = &src
		}
	case **uint16:
		if s.isNull() {
			*v = nil
		} else {
			src := s.uint16()
			*v = &src
		}
	case **uint32:
		if s.isNull() {
			*v = nil
		} else {
			src := s.uint32()
			*v = &src
		}
	case **uint:
		if s.isNull() {
			*v = nil
		} else {
			src := uint(s.uint32())
			*v = &src
		}
	case **uint64:
		if s.isNull() {
			*v = nil
		} else {
			src := s.uint64()
			*v = &src
		}
	case **float32:
		if s.isNull() {
			*v = nil
		} else {
			src := s.float()
			*v = &src
		}
	case **float64:
		if s.isNull() {
			*v = nil
		} else {
			src := s.double()
			*v = &src
		}
	case **time.Time:
		if s.isNull() {
			*v = nil
		} else {
			s.unwrap()
			var src time.Time
			s.setTime(&src)
			*v = &src
		}
	case **time.Duration:
		if s.isNull() {
			*v = nil
		} else {
			src := value.IntervalToDuration(s.int64())
			*v = &src
		}
	case **string:
		if s.isNull() {
			*v = nil
		} else {
			s.unwrap()
			var src string
			s.setString(&src)
			*v = &src
		}
	case **[]byte:
		if s.isNull() {
			*v = nil
		} else {
			s.unwrap()
			var src []byte
			s.setByte(&src)
			*v = &src
		}
	case **[16]byte:
		if s.isNull() {
			*v = nil
		} else {
			src := s.uint128()
			*v = &src
		}
	case **value.UUIDIssue1501FixedBytesWrapper:
		if s.isNull() {
			*v = nil
		} else {
			src := s.uint128()
			val := value.NewUUIDIssue1501FixedBytesWrapper(src)
			*v = &val
		}
	case **uuid.UUID:
		if s.isNull() {
			*v = nil
		} else {
			src := s.uuid()
			*v = &src
		}
	case **interface{}:
		if s.isNull() {
			*v = nil
		} else {
			src := s.any()
			*v = &src
		}
	case *value.Value:
		*v = s.value()
	case **decimal.Decimal:
		if s.isNull() {
			*v = nil
		} else {
			src := s.unwrapDecimal()
			*v = &src
		}
	case scanner.Scanner:
		err := v.UnmarshalYDB(s.converter)
		if err != nil {
			_ = s.errorf(0, "ydb.Scanner error: %w", err)
		}
	case sql.Scanner:
		err := v.Scan(s.any())
		if err != nil {
			_ = s.errorf(0, "sql.Scanner error: %w", err)
		}
	case json.Unmarshaler:
		s.unwrap()
		var err error
		switch s.getType() {
		case internalTypes.JSON:
			if s.isNull() {
				err = v.UnmarshalJSON(nil)
			} else {
				err = v.UnmarshalJSON(s.converter.JSON())
			}
		case internalTypes.JSONDocument:
			if s.isNull() {
				err = v.UnmarshalJSON(nil)
			} else {
				err = v.UnmarshalJSON(s.converter.JSONDocument())
			}
		default:
			_ = s.errorf(0, "ydb optional type %T not unsupported for applying to json.Unmarshaler", s.getType())
		}
		if err != nil {
			_ = s.errorf(0, "json.Unmarshaler error: %w", err)
		}
	default:
		s.unwrap()
		ok := s.trySetByteArray(v, true, false)
		if !ok {
			rv := reflect.TypeOf(v)
			if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Ptr {
				_ = s.errorf(0, "scan row failed: type %T is unknown", v)
			} else {
				_ = s.errorf(0, "scan row failed: type %T is not optional! use double pointer or sql.Scanner.", v)
			}
		}
	}
}

//nolint:funlen
func (s *valueScanner) setDefaultValue(dst interface{}) {
	switch v := dst.(type) {
	case *bool:
		*v = false
	case *int8:
		*v = 0
	case *int16:
		*v = 0
	case *int32:
		*v = 0
	case *int64:
		*v = 0
	case *uint8:
		*v = 0
	case *uint16:
		*v = 0
	case *uint32:
		*v = 0
	case *uint64:
		*v = 0
	case *float32:
		*v = 0
	case *float64:
		*v = 0
	case *time.Time:
		*v = time.Time{}
	case *time.Duration:
		*v = 0
	case *string:
		*v = ""
	case *[]byte:
		*v = nil
	case *[16]byte:
		*v = [16]byte{}
	case *interface{}:
		*v = nil
	case *value.Value:
		*v = s.value()
	case *decimal.Decimal:
		*v = decimal.Decimal{}
	case sql.Scanner:
		err := v.Scan(nil)
		if err != nil {
			_ = s.errorf(0, "sql.Scanner error: %w", err)
		}
	case scanner.Scanner:
		err := v.UnmarshalYDB(s.converter)
		if err != nil {
			_ = s.errorf(0, "ydb.Scanner error: %w", err)
		}
	case json.Unmarshaler:
		err := v.UnmarshalJSON(nil)
		if err != nil {
			_ = s.errorf(0, "json.Unmarshaler error: %w", err)
		}
	default:
		ok := s.trySetByteArray(v, false, true)
		if !ok {
			_ = s.errorf(0, "scan row failed: type %T is unknown", v)
		}
	}
}

func (r *baseResult) SetErr(err error) {
	r.errMtx.WithLock(func() {
		r.err = err
	})
}

func (s *valueScanner) errorf(depth int, f string, args ...interface{}) error {
	s.errMtx.Lock()
	defer s.errMtx.Unlock()
	if s.err != nil {
		return s.err
	}
	s.err = xerrors.WithStackTrace(fmt.Errorf(f, args...), xerrors.WithSkipDepth(depth+1))

	return s.err
}

func (s *valueScanner) typeError(act, exp interface{}) {
	_ = s.errorf(
		2, //nolint:mnd
		"unexpected types during scan at %q %s: %s; want %s",
		s.path(),
		s.getType(),
		nameIface(act),
		nameIface(exp),
	)
}

func (s *valueScanner) valueTypeError(act, exp interface{}) {
	// unexpected value during scan at \"migration_status\" Int64: NullFlag; want Int64
	_ = s.errorf(
		2, //nolint:mnd
		"unexpected value during scan at %q %s: %s; want %s",
		s.path(),
		s.getType(),
		nameIface(act),
		nameIface(exp),
	)
}

func (s *valueScanner) notFoundColumnByIndex(idx int) error {
	return s.errorf(
		2, //nolint:mnd
		"not found %d column",
		idx,
	)
}

func (s *valueScanner) notFoundColumnName(name string) error {
	return s.errorf(
		2, //nolint:mnd
		"not found column '%s'",
		name,
	)
}

func (s *valueScanner) noColumnError(name string) error {
	return s.errorf(
		2, //nolint:mnd
		"no column %q",
		name,
	)
}

func (s *valueScanner) overflowError(i, n interface{}) error {
	return s.errorf(
		2, //nolint:mnd
		"overflow error: %d overflows capacity of %t",
		i,
		n,
	)
}

var emptyItem item

type item struct {
	name string
	i    int // Index in listing types
	t    *Ydb.Type
	v    *Ydb.Value
}

func (x item) isEmpty() bool {
	return x.v == nil
}

type scanStack struct {
	v        []item
	p        int
	scanItem item
}

func (s *scanStack) size() int {
	if !s.scanItem.isEmpty() {
		s.set(s.scanItem)
	}

	return s.p + 1
}

func (s *scanStack) get(i int) item {
	return s.v[i]
}

func (s *scanStack) reset() {
	s.scanItem = emptyItem
	s.p = 0
}

func (s *scanStack) enter() {
	// support compatibility
	if !s.scanItem.isEmpty() {
		s.set(s.scanItem)
	}
	s.scanItem = emptyItem
	s.p++
}

func (s *scanStack) leave() {
	s.set(emptyItem)
	if s.p > 0 {
		s.p--
	}
}

func (s *scanStack) set(v item) {
	if s.p == len(s.v) {
		s.v = append(s.v, v)
	} else {
		s.v[s.p] = v
	}
}

func (s *scanStack) parent() item {
	if s.p == 0 {
		return emptyItem
	}

	return s.v[s.p-1]
}

func (s *scanStack) current() item {
	if !s.scanItem.isEmpty() {
		return s.scanItem
	}
	if s.v == nil {
		return emptyItem
	}

	return s.v[s.p]
}

func (s *scanStack) currentValue() interface{} {
	if v := s.current().v; v != nil {
		return v.GetValue()
	}

	return nil
}

func (s *scanStack) currentType() interface{} {
	if t := s.current().t; t != nil {
		return t.GetType()
	}

	return nil
}

func isOptional(typ *Ydb.Type) bool {
	if typ == nil {
		return false
	}
	_, yes := typ.GetType().(*Ydb.Type_OptionalType)

	return yes
}
