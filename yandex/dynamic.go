package yandex

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// findTag find tag in field
func findTag(f reflect.StructField, tagName, name string) (string, bool) {
	tagF := f.Tag.Get(tagName)
	if tagF == "" {
		return "", false
	}

	tags := strings.Split(tagF, ",")
	for _, tag := range tags {
		kvs := strings.SplitN(tag, "=", 2)
		if kvs[0] != name {
			continue
		}
		val := ""
		if len(kvs) == 2 {
			val = kvs[1]
		}
		return val, true
	}
	return "", false
}

type nilNotAllowedError struct {
	text string
}

func (err *nilNotAllowedError) Error() string {
	return err.text
}

func isNilNotAllowedError(err error) bool {

	if err == nil {
		return false
	}

	_, ok := err.(*nilNotAllowedError)

	return ok
}

type typeMismatchError struct {
	text string
}

func (err *typeMismatchError) Error() string {
	return err.text
}

func isTypeMismatchError(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(*typeMismatchError)

	return ok
}

// setIntField set int value into named field
// v must be ptr
func setIntField(v interface{}, name string, iv *int) error {
	rv, err := getStructValue(v)
	if err != nil {
		return err
	}
	return setIntFieldToReflect(rv, name, iv)
}
func setIntFieldToReflect(rv reflect.Value, name string, v *int) error {

	f := rv.FieldByName(name)

	switch f.Type().Kind() {
	case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
		if v == nil {
			return &nilNotAllowedError{text: fmt.Sprintf("setIntValueToReflect: field %s is not nilable", name)}
		}
		f.SetInt(int64(*v))
		return nil
	case reflect.Ptr:
		if f.Type() == wrapperspbInt64Value() {
			if v == nil {
				var w *wrappers.Int64Value
				f.Set(reflect.ValueOf(w))
				return nil
			}

			w := wrappers.Int64Value{
				Value: int64(*v),
			}
			f.Set(reflect.ValueOf(&w))
			return nil

		}
		return &typeMismatchError{text: "setIntValueToReflect: type ptr not implement"}
	default:
		return &typeMismatchError{text: "setIntValueToReflect: type not implement"}
	}
}

// setBoolField set bool value into named field
// v must be ptr
func setBoolField(v interface{}, name string, bv *bool) error {
	rv, err := getStructValue(v)
	if err != nil {
		return err
	}
	return setBoolFieldToReflect(rv, name, bv)
}
func setBoolFieldToReflect(rv reflect.Value, name string, v *bool) error {

	f := rv.FieldByName(name)

	switch f.Type().Kind() {
	case reflect.Bool:
		if v == nil {
			return &nilNotAllowedError{text: fmt.Sprintf("setBoolValueToReflect: field %s is not nilable", name)}
		}
		f.SetBool(*v)
		return nil
	case reflect.Ptr:
		if f.Type() == wrapperspbBoolValue() {
			if v == nil {
				var w *wrappers.BoolValue
				f.Set(reflect.ValueOf(w))
				return nil
			}

			w := wrappers.BoolValue{
				Value: *v,
			}
			f.Set(reflect.ValueOf(&w))
			return nil

		}
		return &typeMismatchError{text: "setBoolValueToReflect: type ptr not implement"}
	default:
		return &typeMismatchError{text: "setBoolValueToReflect: type not implement"}
	}
}

// setFloatField set float64 value into named field
// v must be ptr
func setFloatField(v interface{}, name string, fv *float64) error {
	rv, err := getStructValue(v)
	if err != nil {
		return err
	}
	return setFloatFieldToReflect(rv, name, fv)
}
func setFloatFieldToReflect(rv reflect.Value, name string, v *float64) error {

	f := rv.FieldByName(name)

	switch f.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		if v == nil {
			return &nilNotAllowedError{text: fmt.Sprintf("setFloatValueToReflect: field %s is not nilable", name)}
		}
		f.SetFloat(*v)
		return nil
	case reflect.Ptr:
		if f.Type() == wrapperspbDoubleValue() {
			if v == nil {
				var w *wrappers.DoubleValue
				f.Set(reflect.ValueOf(w))
				return nil
			}

			w := wrappers.DoubleValue{
				Value: *v,
			}
			f.Set(reflect.ValueOf(&w))
			return nil

		}
		return &typeMismatchError{text: "setFloatValueToReflect: type ptr not implement"}
	default:
		return &typeMismatchError{text: "setFloatValueToReflect: type not implement"}
	}
}

// setStringField set string value into named field
// v must be ptr
func setStringField(v interface{}, name string, fv *string) error {
	rv, err := getStructValue(v)
	if err != nil {
		return err
	}
	return setStringFieldToReflect(rv, name, fv)
}
func setStringFieldToReflect(rv reflect.Value, name string, v *string) error {

	f := rv.FieldByName(name)

	switch f.Type().Kind() {
	case reflect.String:
		if v == nil {
			return &nilNotAllowedError{text: fmt.Sprintf("setFloatValueToReflect: field %s is not nilable", name)}
		}
		f.SetString(*v)
		return nil
	default:
		return &typeMismatchError{text: "setFloatValueToReflect: type not implement"}
	}
}

// getValueFrom get named value from v
func getValueFrom(v interface{}, name string) (interface{}, error) {
	rv, err := getStructValue(v)
	if err != nil {
		return nil, err
	}

	val, err := getValueFromReflect(rv, name)
	if err != nil {
		return nil, err
	}

	return val, err
}
func getValueFromReflect(rv reflect.Value, name string) (interface{}, error) {
	f := rv.FieldByName(name)

	if f.Kind() == reflect.Ptr && f.IsNil() {
		return nil, nil
	}

	switch f.Kind() {
	case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
		return int(f.Int()), nil
	case reflect.Float32, reflect.Float64:
		return f.Float(), nil
	case reflect.Bool:
		return f.Bool(), nil
	case reflect.String:
		return f.String(), nil
	case reflect.Ptr:
		if f.Type() == wrapperspbInt64Value() {
			if f.IsNil() {
				return nil, nil
			}
			v, ok := f.Interface().(*wrappers.Int64Value)
			if ok {
				return int(v.GetValue()), nil
			}
			return nil, &typeMismatchError{text: "getValueFromReflect: type mismatch should be *wrappers.Int64Value"}
		}
		if f.Type() == wrapperspbDoubleValue() {
			if f.IsNil() {
				return nil, nil
			}
			v, ok := f.Interface().(*wrappers.DoubleValue)
			if ok {
				return v.GetValue(), nil
			}
			return nil, &typeMismatchError{text: "getValueFromReflect: type mismatch should be *wrappers.DoubleValue"}
		}
		if f.Type() == wrapperspbBoolValue() {
			if f.IsNil() {
				return nil, nil
			}
			v, ok := f.Interface().(*wrappers.BoolValue)
			if ok {
				return v.GetValue(), nil
			}
			return nil, &typeMismatchError{text: "getValueFromReflect: type mismatch should be *wrappers.BoolValue"}
		}
		return nil, &typeMismatchError{text: "getValueFromReflect: type ptr not implement"}
	default:
		return nil, &typeMismatchError{text: fmt.Sprintf("getValueFromReflect: type not implement %s, %v, %s", name, f.Kind(), rv.Type().Name())}
	}
}

// getStructType returns if v is nil then nil error if v is not struct and not v is not ptr on struct
func getStructType(v interface{}) (reflect.Type, error) {

	if v == nil {
		return nil, nil
	}

	return getStructTypeReflect(reflect.ValueOf(v))

}
func getStructTypeReflect(rv reflect.Value) (reflect.Type, error) {

	t := rv.Type()

	if t.Kind() == reflect.Ptr && rv.IsNil() {
		return nil, nil
	}
	if t.Kind() == reflect.Ptr {
		return getStructTypeReflect(rv.Elem())
	}

	if t.Kind() == reflect.Struct {
		return t, nil
	}

	return nil, fmt.Errorf("getStructTypeReflect: type %v is not struct", t.Kind())

}

// getStructValue returns error if v is nil or v is not struct and not v is not ptr on struct
func getStructValue(v interface{}) (rv reflect.Value, err error) {

	if v == nil {
		return rv, fmt.Errorf("getStructValue: value is nil")
	}

	return getStructValueReflect(reflect.ValueOf(v))

}
func getStructValueReflect(rv reflect.Value) (reflect.Value, error) {

	t := rv.Type()

	if t.Kind() == reflect.Ptr && rv.IsNil() {
		return rv, fmt.Errorf("getStructValueReflect: value is nil")
	}
	if t.Kind() == reflect.Ptr {
		return getStructValueReflect(rv.Elem())
	}

	if t.Kind() == reflect.Struct {
		return rv, nil
	}

	return rv, fmt.Errorf("getStructValueReflect: type %v is not struct", t.Kind())

}

func wrapperspbInt64Value() reflect.Type {
	return reflect.TypeOf(&wrappers.Int64Value{})
}

func wrapperspbBoolValue() reflect.Type {
	return reflect.TypeOf(&wrappers.BoolValue{})
}

func wrapperspbDoubleValue() reflect.Type {
	return reflect.TypeOf(&wrappers.DoubleValue{})
}

func getFieldsInfo(v interface{}, tagName string, tagValue string) (map[string]fieldReflectInfo, error) {

	fis := make(map[string]fieldReflectInfo)
	t, err := getStructType(v)
	if err != nil {
		return nil, err
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tg, okTg := findTag(f, tagName, tagValue)

		if okTg {
			if f.Type.Kind() == reflect.Int32 || f.Type.Kind() == reflect.Int64 || f.Type == wrapperspbInt64Value() {
				fis[tg] = fieldReflectInfo{
					valueType: schema.TypeInt,
					name:      f.Name,
				}
			}
			if f.Type.Kind() == reflect.Bool || f.Type == wrapperspbBoolValue() {
				fis[tg] = fieldReflectInfo{
					valueType: schema.TypeBool,
					name:      f.Name,
				}
			}
			if f.Type.Kind() == reflect.Float32 || f.Type.Kind() == reflect.Float64 || f.Type == wrapperspbDoubleValue() {
				fis[tg] = fieldReflectInfo{
					valueType: schema.TypeFloat,
					name:      f.Name,
				}
			}
			if f.Type.Kind() == reflect.String {
				fis[tg] = fieldReflectInfo{
					valueType: schema.TypeString,
					name:      f.Name,
				}
			}
		}
	}

	return fis, nil
}
