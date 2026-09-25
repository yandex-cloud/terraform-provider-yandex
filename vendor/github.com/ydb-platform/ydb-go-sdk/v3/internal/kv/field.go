package kv

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/version"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

const nilPtr = "<nil>"

// KeyValue represents typed log field (a key-value pair). Adapters should determine
// KeyValue's type based on Type and use the corresponding getter method to retrieve
// the value:
//
//	switch f.Type() {
//	case logs.IntType:
//	    var i int = f.Int()
//	    // handle int value
//	case logs.StringType:
//	    var s string = f.String()
//	    // handle string value
//	//...
//	}
//
// Getter methods must not be called on fields with wrong Type (e.g. calling String()
// on fields with Type != StringType).
// KeyValue must not be initialized directly as a struct literal.
type KeyValue struct {
	ftype FieldType
	key   string

	vint int64
	vstr string
	vany interface{}
}

func (f KeyValue) Type() FieldType {
	return f.ftype
}

func (f KeyValue) Key() string {
	return f.key
}

// StringValue is a value getter for fields with StringType type
func (f KeyValue) StringValue() string {
	f.checkType(StringType)

	return f.vstr
}

// IntValue is a value getter for fields with IntType type
func (f KeyValue) IntValue() int {
	f.checkType(IntType)

	return int(f.vint)
}

// Int64Value is a value getter for fields with Int64Type type
func (f KeyValue) Int64Value() int64 {
	f.checkType(Int64Type)

	return f.vint
}

// BoolValue is a value getter for fields with BoolType type
func (f KeyValue) BoolValue() bool {
	f.checkType(BoolType)

	return f.vint != 0
}

// DurationValue is a value getter for fields with DurationType type
func (f KeyValue) DurationValue() time.Duration {
	f.checkType(DurationType)

	return time.Nanosecond * time.Duration(f.vint)
}

// StringsValue is a value getter for fields with StringsType type
func (f KeyValue) StringsValue() []string {
	f.checkType(StringsType)
	if f.vany == nil {
		return nil
	}
	val, ok := f.vany.([]string)
	if !ok {
		panic(fmt.Sprintf("unsupported type conversion from %T to []string", val))
	}

	return val
}

// ErrorValue is a value getter for fields with ErrorType type
func (f KeyValue) ErrorValue() error {
	f.checkType(ErrorType)
	if f.vany == nil {
		return nil
	}
	val, ok := f.vany.(error)
	if !ok {
		panic(fmt.Sprintf("unsupported type conversion from %T to error", val))
	}

	return val
}

// AnyValue is a value getter for fields with AnyType type
func (f KeyValue) AnyValue() interface{} {
	switch f.ftype {
	case AnyType:
		return f.vany
	case IntType:
		return f.IntValue()
	case Int64Type:
		return f.Int64Value()
	case StringType:
		return f.StringValue()
	case BoolType:
		return f.BoolValue()
	case DurationType:
		return f.DurationValue()
	case StringsType:
		return f.StringsValue()
	case ErrorType:
		return f.ErrorValue()
	case StringerType:
		return f.Stringer()
	default:
		panic(fmt.Sprintf("unknown FieldType %d", f.ftype))
	}
}

// Stringer is a value getter for fields with StringerType type
func (f KeyValue) Stringer() fmt.Stringer {
	f.checkType(StringerType)
	if f.vany == nil {
		return nil
	}
	val, ok := f.vany.(fmt.Stringer)
	if !ok {
		panic(fmt.Sprintf("unsupported type conversion from %T to fmt.Stringer", val))
	}

	return val
}

// Panics on type mismatch
func (f KeyValue) checkType(want FieldType) {
	if f.ftype != want {
		panic(fmt.Sprintf("bad type. have: %s, want: %s", f.ftype, want))
	}
}

// Returns default string representation of KeyValue value.
// It should be used by adapters that don't support f.Type directly.
func (f KeyValue) String() string {
	switch f.ftype {
	case IntType, Int64Type:
		return strconv.FormatInt(f.vint, 10)
	case StringType:
		return f.vstr
	case BoolType:
		return strconv.FormatBool(f.BoolValue())
	case DurationType:
		return f.DurationValue().String()
	case StringsType:
		return fmt.Sprintf("%v", f.StringsValue())
	case ErrorType:
		if f.vany == nil {
			return nilPtr
		}

		val, ok := f.vany.(error)
		if !ok {
			panic(fmt.Sprintf("unsupported type conversion from %T to fmt.Stringer", val))
		}

		if val == nil {
			return "<nil>"
		}

		return f.ErrorValue().Error()
	case AnyType:
		if f.vany == nil {
			return nilPtr
		}
		if v := reflect.ValueOf(f.vany); v.Type().Kind() == reflect.Ptr {
			if v.IsNil() {
				return nilPtr
			}

			return v.Type().String() + "(" + fmt.Sprint(v.Elem()) + ")"
		}

		return fmt.Sprint(f.vany)
	case StringerType:
		return f.Stringer().String()
	default:
		panic(fmt.Sprintf("unknown FieldType %d", f.ftype))
	}
}

// String constructs KeyValue with StringType
func String(k, v string) KeyValue {
	return KeyValue{
		ftype: StringType,
		key:   k,
		vstr:  v,
	}
}

// Int constructs KeyValue with IntType
func Int(k string, v int) KeyValue {
	return KeyValue{
		ftype: IntType,
		key:   k,
		vint:  int64(v),
	}
}

func Int64(k string, v int64) KeyValue {
	return KeyValue{
		ftype: Int64Type,
		key:   k,
		vint:  v,
	}
}

// Bool constructs KeyValue with BoolType
func Bool(key string, value bool) KeyValue {
	var vint int64
	if value {
		vint = 1
	} else {
		vint = 0
	}

	return KeyValue{
		ftype: BoolType,
		key:   key,
		vint:  vint,
	}
}

// Duration constructs KeyValue with DurationType
func Duration(key string, value time.Duration) KeyValue {
	return KeyValue{
		ftype: DurationType,
		key:   key,
		vint:  value.Nanoseconds(),
	}
}

// Strings constructs KeyValue with StringsType
func Strings(key string, value []string) KeyValue {
	return KeyValue{
		ftype: StringsType,
		key:   key,
		vany:  value,
	}
}

// NamedError constructs KeyValue with ErrorType
func NamedError(key string, value error) KeyValue {
	return KeyValue{
		ftype: ErrorType,
		key:   key,
		vany:  value,
	}
}

// Error is the same as NamedError("error", value)
func Error(value error) KeyValue {
	return NamedError("error", value)
}

// Any constructs untyped KeyValue.
func Any(key string, value interface{}) KeyValue {
	return KeyValue{
		ftype: AnyType,
		key:   key,
		vany:  value,
	}
}

// Stringer constructs KeyValue with StringerType. If value is nil,
// resulting KeyValue will be of AnyType instead of StringerType.
func Stringer(key string, value fmt.Stringer) KeyValue {
	if value == nil {
		return Any(key, nil)
	}

	return KeyValue{
		ftype: StringerType,
		key:   key,
		vany:  value,
	}
}

// FieldType indicates type info about the KeyValue. This enum might be extended in future releases.
// Adapters that don't support some FieldType value should use KeyValue.Fallback() for marshaling.
type FieldType int

const (
	// InvalidType indicates that KeyValue was not initialized correctly. Adapters
	// should either ignore such field or issue an error. No value getters should
	// be called on field with such type.
	InvalidType FieldType = iota

	IntType
	Int64Type
	StringType
	BoolType
	DurationType

	// StringsType corresponds to []string
	StringsType

	ErrorType
	// AnyType indicates that the KeyValue is untyped. Adapters should use
	// reflection-based approached to marshal this field.
	AnyType

	// StringerType corresponds to fmt.Stringer
	StringerType

	endType
)

func (ft FieldType) String() (typeName string) {
	switch ft {
	case InvalidType:
		typeName = "invalid"
	case IntType:
		typeName = "int"
	case Int64Type:
		typeName = "int64"
	case StringType:
		typeName = "string"
	case BoolType:
		typeName = "bool"
	case DurationType:
		typeName = "time.Duration"
	case StringsType:
		typeName = "[]string"
	case ErrorType:
		typeName = "error"
	case AnyType:
		typeName = "any"
	case StringerType:
		typeName = "stringer"
	case endType:
		typeName = "endtype"
	default:
		panic("not implemented")
	}

	return typeName
}

// Latency creates KeyValue "latency": time.Since(start)
func Latency(start time.Time) KeyValue {
	return Duration("latency", time.Since(start))
}

// Version creates KeyValue "version": version.Version
func Version() KeyValue {
	return String("version", version.Version)
}

type Endpoints []trace.EndpointInfo

func (ee Endpoints) String() string {
	b := xstring.Buffer()
	defer b.Free()
	b.WriteByte('[')
	for i, e := range ee {
		if i != 0 {
			b.WriteByte(',')
		}
		b.WriteString(e.String())
	}
	b.WriteByte(']')

	return b.String()
}

type Metadata map[string][]string

func (m Metadata) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("error:%s", err)
	}

	return xstring.FromBytes(b)
}
