package yandex

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type fieldAdditionalInfo struct {
	iDefaultValue *int
	iToString     func(*int) (string, error)
	iToInt        func(string) (*int, error)
	iMinVal       *int
	iMaxVal       *int
	isNotNullable bool
	skip          bool
	isIStringable bool
}

type fieldInfo struct {
	valueType schema.ValueType
}

type objectFieldsAdditionalInfo struct {
	fields           map[string]fieldAdditionalInfo
	valueTypes       map[string]fieldInfo
	disableBackToNil bool
}

func (of *objectFieldsAdditionalInfo) skip(field string) bool {
	if of == nil {
		return false
	}
	if of.fields == nil {
		return false
	}

	if fai, ok := of.fields[field]; ok {
		return fai.skip
	}
	return false
}
func (of *objectFieldsAdditionalInfo) isIStringable(field string) bool {
	if of == nil {
		return false
	}
	if of.fields == nil {
		return false
	}

	if fai, ok := of.fields[field]; ok {
		return fai.isIStringable
	}
	return false
}

func (of *objectFieldsAdditionalInfo) iGetDefault(field string) *int {
	if of == nil {
		return nil
	}
	if of.fields == nil {
		return nil
	}

	if fai, ok := of.fields[field]; ok {
		return fai.iDefaultValue
	}
	return nil
}

func (of *objectFieldsAdditionalInfo) backToNil(field string) bool {
	if of == nil {
		return true
	}

	if of.fields == nil {
		return false
	}
	if fai, ok := of.fields[field]; ok {
		return !fai.isNotNullable && !of.disableBackToNil
	}

	return !of.disableBackToNil
}

func (of *objectFieldsAdditionalInfo) iCheckSetValue(field string, v *int) error {
	if v == nil {
		if of.backToNil(field) {
			return fmt.Errorf("iCheckSetValue: you can't set nil %s", field)
		}

		return nil
	}

	if of.fields == nil {
		return nil
	}

	if fai, ok := of.fields[field]; ok {
		if fai.iMinVal != nil && *fai.iMinVal > *v {
			return fmt.Errorf("iCheckSetValue: min value for %s is %v value is %v", field, fai.iMinVal, v)
		}
		if fai.iMaxVal != nil && *fai.iMaxVal < *v {
			return fmt.Errorf("iCheckSetValue: max value for %s is %v value is %v", field, fai.iMaxVal, v)
		}
	}

	return nil

}

func (of *objectFieldsAdditionalInfo) iEqualDefault(field string, v interface{}) bool {

	if itm, ok := v.(int); ok {
		return of.iGetDefault(field) != nil && *of.iGetDefault(field) == itm
	}

	if itm, ok := v.(*int); ok {
		return of.iGetDefault(field) == nil && itm == nil || *of.iGetDefault(field) == *itm
	}

	return false
}

func (of *objectFieldsAdditionalInfo) iToString(field string, v *int) (string, error) {
	if of == nil {
		return defaultIToString(v)
	}
	if of.fields == nil {
		return defaultIToString(v)
	}

	if fai, ok := of.fields[field]; ok {
		if fai.iToString == nil {
			return defaultIToString(v)
		}
		return fai.iToString(v)
	}
	return defaultIToString(v)
}
func defaultIToString(v *int) (string, error) {
	if v == nil {
		return "", nil
	}
	return strconv.Itoa(*v), nil
}

func (of *objectFieldsAdditionalInfo) iToInt(field string, v string) (*int, error) {
	if of == nil {
		return defaultIToInt(v)
	}
	if of.fields == nil {
		return defaultIToInt(v)
	}

	if fai, ok := of.fields[field]; ok {
		if fai.iToString == nil {
			return defaultIToInt(v)
		}
		return fai.iToInt(v)
	}
	return defaultIToInt(v)
}
func defaultIToInt(v string) (*int, error) {

	if v == "" {
		return nil, nil
	}

	i, err := strconv.Atoi(v)
	return &i, err
}

func (of *objectFieldsAdditionalInfo) getType(field string) schema.ValueType {
	if of == nil {
		return schema.TypeInvalid
	}
	if of.valueTypes == nil {
		return schema.TypeInvalid
	}

	if vt, ok := of.valueTypes[field]; ok {
		return vt.valueType
	}
	return schema.TypeInvalid
}

func makeObjectFieldsAdditionalInfo() *objectFieldsAdditionalInfo {
	return &objectFieldsAdditionalInfo{
		fields:     make(map[string]fieldAdditionalInfo),
		valueTypes: make(map[string]fieldInfo),
	}
}
func (of *objectFieldsAdditionalInfo) updFromType(v interface{}) *objectFieldsAdditionalInfo {

	t := reflect.TypeOf(v)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tg, okTg := FindTag(f, "protobuf", "name")

		if okTg {
			if f.Type.Kind() == reflect.Int32 || f.Type == wrapperspbInt64Value() {
				of.valueTypes[tg] = fieldInfo{valueType: schema.TypeInt}
			}
		}
	}

	return of
}

func (of *objectFieldsAdditionalInfo) addField(field string, fai fieldAdditionalInfo) *objectFieldsAdditionalInfo {

	of.fields[field] = fai

	return of
}

func (of *objectFieldsAdditionalInfo) addSkip(field string) *objectFieldsAdditionalInfo {

	of.fields[field] = fieldAdditionalInfo{skip: true}

	return of
}
func (of *objectFieldsAdditionalInfo) addIDefault(field string, def int) *objectFieldsAdditionalInfo {

	of.fields[field] = fieldAdditionalInfo{iDefaultValue: &def}

	return of
}

// default value is 0
func (of *objectFieldsAdditionalInfo) addEnum(field string, values map[int]string) *objectFieldsAdditionalInfo {

	def := 0
	of.fields[field] = fieldAdditionalInfo{
		iDefaultValue: &def,
		iToString:     makeIToString(values, def),
		iToInt:        makeIToInt(values, &def),
		isIStringable: true,
		isNotNullable: true,
	}

	return of
}

// default value is 0
func (of *objectFieldsAdditionalInfo) addEnum2(field string, values map[int]string, values2 map[int32]string) *objectFieldsAdditionalInfo {

	def := 0
	of.fields[field] = fieldAdditionalInfo{
		iDefaultValue: &def,
		iToString:     makeIToString(values, def),
		iToInt:        makeIToInt2(values, convIValuesToI32(values2), &def),
		isIStringable: true,
		isNotNullable: true,
	}

	return of
}

func convIValuesToI32(values map[int32]string) map[int]string {
	valuesI := make(map[int]string)
	for k, v := range values {
		valuesI[int(k)] = v
	}
	return valuesI
}
func makeIToString(values map[int]string, defV int) func(*int) (string, error) {
	return func(i *int) (string, error) {
		if i == nil {
			i = &defV
		}
		out, ok := values[*i]
		if !ok {
			return "", fmt.Errorf("Value %v is not in enum", i)
		}
		return out, nil
	}
}
func makeIToInt(values map[int]string, defV *int) func(string) (*int, error) {

	valuesBack := make(map[string]int)
	for k, v := range values {
		valuesBack[v] = k
	}

	return func(s string) (*int, error) {
		if s == "" {
			return defV, nil
		}
		out, ok := valuesBack[s]
		if !ok {
			i, err := strconv.Atoi(s)
			if err == nil {
				return &i, nil
			}
			return nil, fmt.Errorf("Value %v is not in enum", s)
		}
		return &out, nil
	}
}
func makeIToInt2(values map[int]string, values2 map[int]string, defV *int) func(string) (*int, error) {

	valuesBack := make(map[string]int)
	for k, v := range values {
		valuesBack[v] = k
	}
	valuesBack2 := make(map[string]int)
	for k, v := range values2 {
		valuesBack2[v] = k
	}

	return func(s string) (*int, error) {
		if s == "" {
			return defV, nil
		}
		out, ok := valuesBack[s]
		if !ok {
			out, ok = valuesBack2[s]
			if !ok {
				i, err := strconv.Atoi(s)
				if err == nil {
					return &i, nil
				}
				return nil, fmt.Errorf("Value %v is not in enum", s)
			}
			return &out, nil
		}
		return &out, nil
	}
}
