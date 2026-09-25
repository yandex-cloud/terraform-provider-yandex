package types

import (
	"errors"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var errNilValue = errors.New("nil value")

// CastTo try cast value to destination type value
func CastTo(v Value, dst interface{}) error {
	if v == nil {
		return xerrors.WithStackTrace(errNilValue)
	}

	return value.CastTo(v, dst)
}

// IsOptional checks if type is optional and returns innerType if it is.
func IsOptional(t Type) (isOptional bool, innerType Type) {
	if optionalType, isOptional := t.(interface {
		IsOptional()
		InnerType() Type
	}); isOptional {
		return isOptional, optionalType.InnerType()
	}

	return false, nil
}

// IsNull checks if optional value is NULL.
// If value is not optional value - returns false
func IsNull(v Value) bool {
	return value.IsNull(v)
}

// Unwrap returns inner optional value
// If value is not optional value - returns self
func Unwrap(v Value) Value {
	return value.Unwrap(v)
}

// ToDecimal returns Decimal struct from abstract Value
func ToDecimal(v Value) (*Decimal, error) {
	if valuer, isDecimalValuer := v.(value.DecimalValuer); isDecimalValuer {
		return &Decimal{
			Bytes:     valuer.Value(),
			Precision: valuer.Precision(),
			Scale:     valuer.Scale(),
		}, nil
	}

	return nil, xerrors.WithStackTrace(fmt.Errorf("value type '%s' is not decimal type", v.Type().Yql()))
}

// ListItems returns list items from abstract Value
func ListItems(v Value) ([]Value, error) {
	if vv, has := v.(interface {
		ListItems() []Value
	}); has {
		return vv.ListItems(), nil
	}

	return nil, xerrors.WithStackTrace(fmt.Errorf("cannot get list items from '%s'", v.Type().Yql()))
}

// TupleItems returns tuple items from abstract Value
func TupleItems(v Value) ([]Value, error) {
	if vv, has := v.(interface {
		TupleItems() []Value
	}); has {
		return vv.TupleItems(), nil
	}

	return nil, xerrors.WithStackTrace(fmt.Errorf("cannot get tuple items from '%s'", v.Type().Yql()))
}

// StructFields returns struct fields from abstract Value
func StructFields(v Value) (map[string]Value, error) {
	if vv, has := v.(interface {
		StructFields() map[string]Value
	}); has {
		return vv.StructFields(), nil
	}

	return nil, xerrors.WithStackTrace(fmt.Errorf("cannot get struct fields from '%s'", v.Type().Yql()))
}

// VariantValue returns variant value from abstract Value
func VariantValue(v Value) (name string, idx uint32, _ Value, _ error) {
	if vv, has := v.(interface {
		Variant() (name string, index uint32)
		Value() Value
	}); has {
		name, idx := vv.Variant()

		return name, idx, vv.Value(), nil
	}

	return "", 0, nil, xerrors.WithStackTrace(fmt.Errorf("cannot get variant value from '%s'", v.Type().Yql()))
}

// DictFields returns dict values from abstract Value
//
// Deprecated: use DictValues instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func DictFields(v Value) (map[Value]Value, error) {
	return DictValues(v)
}

// DictValues returns dict values from abstract Value
func DictValues(v Value) (map[Value]Value, error) {
	if vv, has := v.(interface {
		DictValues() map[Value]Value
	}); has {
		return vv.DictValues(), nil
	}

	return nil, xerrors.WithStackTrace(fmt.Errorf("cannot get dict values from '%s'", v.Type().Yql()))
}
