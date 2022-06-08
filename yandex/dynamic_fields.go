package yandex

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type fieldManualInfo struct {
	defaultIntValue    *int
	defaultStringValue string
	isDefaultSet       bool
	intToString        func(*int) (string, error)
	stringToInt        func(string) (*int, error)
	minIntVal          *int
	maxMaxVal          *int
	isNotNullable      bool
	skip               bool
	isStringable       bool

	emptySliceValue string

	checkValueFunc   func(fieldsInfo *objectFieldsInfo, v interface{}) error
	compareValueFunc func(fieldsInfo *objectFieldsInfo, old, new string) bool
}

type fieldReflectInfo struct {
	name      string
	valueType schema.ValueType
}

type objectFieldsInfo struct {
	fieldsManual     map[string]fieldManualInfo
	fieldsReflect    map[reflect.Type]map[string]fieldReflectInfo
	nameFieldsType   map[string][]reflect.Type
	disableBackToNil bool
}

func (fieldsInfo *objectFieldsInfo) skip(field string) bool {
	if fieldsInfo == nil {
		return false
	}
	if fieldsInfo.fieldsManual == nil {
		return false
	}

	if fieldInfo, ok := fieldsInfo.fieldsManual[field]; ok {
		return fieldInfo.skip
	}
	return false
}
func (fieldsInfo *objectFieldsInfo) isStringable(field string) bool {
	if fieldsInfo == nil {
		return false
	}
	if fieldsInfo.fieldsManual == nil {
		return false
	}

	if fieldInfo, ok := fieldsInfo.fieldsManual[field]; ok {
		return fieldInfo.isStringable
	}
	return false
}

func (fieldsInfo *objectFieldsInfo) getDefault(field string) *int {
	if fieldsInfo == nil {
		return nil
	}
	if fieldsInfo.fieldsManual == nil {
		return nil
	}

	if fieldInfo, ok := fieldsInfo.fieldsManual[field]; ok && fieldInfo.isDefaultSet {
		return fieldInfo.defaultIntValue
	}
	return nil
}

func (fieldsInfo *objectFieldsInfo) backToNil(field string) bool {
	if fieldsInfo == nil {
		return true
	}

	if fieldsInfo.fieldsManual == nil {
		return !fieldsInfo.disableBackToNil
	}
	if fieldInfo, ok := fieldsInfo.fieldsManual[field]; ok {
		return !fieldInfo.isNotNullable && !fieldsInfo.disableBackToNil
	}

	return !fieldsInfo.disableBackToNil
}

func (fieldsInfo *objectFieldsInfo) intCheckSetValue(field string, v *int) error {
	if v == nil {
		if !fieldsInfo.backToNil(field) {
			return &nilNotAllowedError{text: fmt.Sprintf("intCheckSetValue: you can't set nil %s", field)}
		}

		return nil
	}

	if fieldsInfo.fieldsManual == nil {
		return nil
	}

	if fieldInfo, ok := fieldsInfo.fieldsManual[field]; ok {
		if fieldInfo.minIntVal != nil && *fieldInfo.minIntVal > *v {
			return fmt.Errorf("intCheckSetValue: min value for %s is %v value is %v", field, fieldInfo.minIntVal, v)
		}
		if fieldInfo.maxMaxVal != nil && *fieldInfo.maxMaxVal < *v {
			return fmt.Errorf("intCheckSetValue: max value for %s is %v value is %v", field, fieldInfo.maxMaxVal, v)
		}
	}

	return nil

}

func (fieldsInfo *objectFieldsInfo) floatCheckSetValue(field string, v *float64) error {
	if v == nil && !fieldsInfo.backToNil(field) {
		return &nilNotAllowedError{text: fmt.Sprintf("floatCheckSetValue: you can't set nil %s", field)}

	}

	return nil
}

func (fieldsInfo *objectFieldsInfo) boolCheckSetValue(field string, v *bool) error {
	if v == nil && !fieldsInfo.backToNil(field) {
		return &nilNotAllowedError{text: fmt.Sprintf("boolCheckSetValue: you can't set nil %s", field)}

	}
	return nil
}

func (fieldsInfo *objectFieldsInfo) stringCheckSetValue(field string, v *string) error {
	if v == nil && !fieldsInfo.backToNil(field) {
		return &nilNotAllowedError{text: fmt.Sprintf("stringCheckSetValue: you can't set nil %s", field)}

	}
	return nil
}

func (fieldsInfo *objectFieldsInfo) intEqualDefault(field string, v interface{}) bool {

	if itm, ok := v.(int); ok {
		return fieldsInfo.getDefault(field) != nil && *fieldsInfo.getDefault(field) == itm
	}

	if itm, ok := v.(*int); ok {
		return fieldsInfo.getDefault(field) == nil && itm == nil || *fieldsInfo.getDefault(field) == *itm
	}

	return false
}

func (fieldsInfo *objectFieldsInfo) intToString(field string, v *int) (string, error) {
	if fieldsInfo == nil {
		return defaultIntToString(v)
	}
	if fieldsInfo.fieldsManual == nil {
		return defaultIntToString(v)
	}

	if fieldInfo, ok := fieldsInfo.fieldsManual[field]; ok {
		if fieldInfo.intToString == nil {
			return defaultIntToString(v)
		}
		return fieldInfo.intToString(v)
	}
	return defaultIntToString(v)
}
func defaultIntToString(v *int) (string, error) {
	if v == nil {
		return "", nil
	}
	return strconv.Itoa(*v), nil
}

func (fieldsInfo *objectFieldsInfo) stringToInt(field string, v string) (*int, error) {
	if fieldsInfo == nil {
		return defaultStringToInt(v)
	}
	if fieldsInfo.fieldsManual == nil {
		return defaultStringToInt(v)
	}

	if fieldInfo, ok := fieldsInfo.fieldsManual[field]; ok {
		if fieldInfo.intToString == nil {
			return defaultStringToInt(v)
		}
		return fieldInfo.stringToInt(v)
	}
	return defaultStringToInt(v)
}
func defaultStringToInt(v string) (*int, error) {

	if v == "" {
		return nil, nil
	}

	i, err := strconv.Atoi(v)
	if err == nil {
		return &i, nil
	}
	return nil, &typeMismatchError{text: err.Error()}
}

func (fieldsInfo *objectFieldsInfo) stringToFloat(field string, v string) (*float64, error) {
	return defaultStringToFloat(v)
}
func defaultStringToFloat(v string) (*float64, error) {

	if v == "" {
		return nil, nil
	}

	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return &f, nil
	}
	return nil, &typeMismatchError{text: err.Error()}
}

func (fieldsInfo *objectFieldsInfo) stringToBool(field string, v string) (*bool, error) {
	return defaultStringToBool(v)
}
func defaultStringToBool(v string) (*bool, error) {

	if v == "" {
		return nil, nil
	}

	b, err := strconv.ParseBool(v)
	if err == nil {
		return &b, nil
	}
	return nil, &typeMismatchError{text: err.Error()}
}

func (fieldsInfo *objectFieldsInfo) checkValueFunc(field string) func(v interface{}) error {
	if fieldsInfo == nil {
		return nil
	}
	if fieldsInfo.fieldsManual == nil {
		return nil
	}

	if fieldInfo, ok := fieldsInfo.fieldsManual[field]; ok {
		if fieldInfo.checkValueFunc != nil {
			return func(v interface{}) error {
				return fieldInfo.checkValueFunc(fieldsInfo, v)
			}
		}
		return nil
	}
	return nil
}
func (fieldsInfo *objectFieldsInfo) compareValueFunc(field string) func(old string, new string) bool {
	if fieldsInfo == nil {
		return nil
	}
	if fieldsInfo.fieldsManual == nil {
		return nil
	}

	if fieldInfo, ok := fieldsInfo.fieldsManual[field]; ok {
		if fieldInfo.compareValueFunc != nil {
			return func(old string, new string) bool {
				return fieldInfo.compareValueFunc(fieldsInfo, old, new)
			}
		}
		return nil
	}
	return nil
}

func (fieldsInfo *objectFieldsInfo) getFields(t reflect.Type) map[string]fieldReflectInfo {
	if fieldsInfo == nil {
		return make(map[string]fieldReflectInfo)
	}
	if fieldsInfo.fieldsReflect == nil {
		return make(map[string]fieldReflectInfo)
	}

	if fis, ok := fieldsInfo.fieldsReflect[t]; ok && fis != nil {
		return fis
	}

	return make(map[string]fieldReflectInfo)
}

func (fieldsInfo *objectFieldsInfo) getType(t reflect.Type, field string) fieldReflectInfo {
	fis := fieldsInfo.getFields(t)

	if fi, ok := fis[field]; ok {
		return fi
	}
	return fieldReflectInfo{}
}

func newObjectFieldsInfo() *objectFieldsInfo {
	return &objectFieldsInfo{
		fieldsManual:   make(map[string]fieldManualInfo),
		fieldsReflect:  make(map[reflect.Type]map[string]fieldReflectInfo),
		nameFieldsType: make(map[string][]reflect.Type),
	}
}

// addType panics on get type errors and any other errors
func (fieldsInfo *objectFieldsInfo) addType(v interface{}) *objectFieldsInfo {

	fis, err := getFieldsInfo(v, "protobuf", "name")
	if err != nil {
		panic(err.Error())
	}

	t, err := getStructType(v)
	if err != nil {
		panic(err.Error())
	}

	if t == nil {
		panic("addType: type must be not nil")
	}

	fieldsInfo.fieldsReflect[t] = fis

	for k := range fis {
		if l, ok := fieldsInfo.nameFieldsType[k]; ok {
			fieldsInfo.nameFieldsType[k] = append(l, t)
		} else {
			fieldsInfo.nameFieldsType[k] = []reflect.Type{t}
		}
	}

	return fieldsInfo
}

func (fieldsInfo *objectFieldsInfo) addSkipEnumGeneratedNames(field string, values map[int32]string,
	checkValueFunc func(fieldsInfo *objectFieldsInfo, v interface{}) error, compareValueFunc func(fieldsInfo *objectFieldsInfo, old, new string) bool) *objectFieldsInfo {

	def := 0
	fieldsInfo.fieldsManual[field] = fieldManualInfo{
		defaultIntValue:  &def,
		isDefaultSet:     true,
		intToString:      makeIntToString(convIValuesToI32(values), def),
		stringToInt:      makeStringToInt(convIValuesToI32(values), &def),
		isStringable:     true,
		isNotNullable:    true,
		skip:             true,
		checkValueFunc:   checkValueFunc,
		compareValueFunc: compareValueFunc,
	}

	return fieldsInfo
}

func (fieldsInfo *objectFieldsInfo) addIDefault(field string, def int) *objectFieldsInfo {

	fieldsInfo.fieldsManual[field] = fieldManualInfo{defaultIntValue: &def, isDefaultSet: true}

	return fieldsInfo
}

// default value is 0
func (fieldsInfo *objectFieldsInfo) addEnumGeneratedNames(field string, values map[int32]string) *objectFieldsInfo {

	def := 0
	fieldsInfo.fieldsManual[field] = fieldManualInfo{
		defaultIntValue: &def,
		isDefaultSet:    true,
		intToString:     makeIntToString(convIValuesToI32(values), def),
		stringToInt:     makeStringToInt(convIValuesToI32(values), &def),
		isStringable:    true,
		isNotNullable:   true,
	}

	return fieldsInfo
}

// default value is 0
func (fieldsInfo *objectFieldsInfo) addEnumHumanNames(field string, values map[int]string, values2 map[int32]string) *objectFieldsInfo {

	def := 0
	fieldsInfo.fieldsManual[field] = fieldManualInfo{
		defaultIntValue: &def,
		isDefaultSet:    true,
		intToString:     makeIntToString(values, def),
		stringToInt:     makeStringToInt2(values, convIValuesToI32(values2), &def),
		isStringable:    true,
		isNotNullable:   true,
	}

	return fieldsInfo
}

func convIValuesToI32(values map[int32]string) map[int]string {
	valuesI := make(map[int]string)
	for k, v := range values {
		valuesI[int(k)] = v
	}
	return valuesI
}
func makeIntToString(values map[int]string, defV int) func(*int) (string, error) {
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

func makeStringToInt(values map[int]string, defV *int) func(string) (*int, error) {

	reversedMap := make(map[string]int)
	for k, v := range values {
		reversedMap[v] = k
	}

	return func(s string) (*int, error) {
		if s == "" {
			return defV, nil
		}
		out, ok := reversedMap[s]
		if !ok {
			i, err := strconv.Atoi(s)
			if err == nil {
				return &i, nil
			}

			return nil, &typeMismatchError{text: fmt.Sprintf("Value %v is not in enum", s)}
		}
		return &out, nil
	}
}
func makeStringToInt2(values map[int]string, values2 map[int]string, defV *int) func(string) (*int, error) {

	reversedMap := make(map[string]int)
	for k, v := range values {
		reversedMap[v] = k
	}
	reversedMap2 := make(map[string]int)
	for k, v := range values2 {
		reversedMap2[v] = k
	}

	return func(s string) (*int, error) {
		if s == "" {
			return defV, nil
		}
		out, ok := reversedMap[s]
		if !ok {
			out, ok = reversedMap2[s]
			if !ok {
				i, err := strconv.Atoi(s)
				if err == nil {
					return &i, nil
				}
				return nil, &typeMismatchError{text: fmt.Sprintf("Value %v is not in enum", s)}
			}
			return &out, nil
		}
		return &out, nil
	}
}

func (fieldsInfo *objectFieldsInfo) stringToIntSlice(field string, value string) (result []int32, err error) {

	fieldInfo, ok := fieldsInfo.fieldsManual[field]

	if !ok {
		return nil, fmt.Errorf("Field should be define")
	}

	result = make([]int32, 0)

	if value == "" && fieldInfo.isDefaultSet {
		value = fieldInfo.defaultStringValue
	}

	if value == fieldInfo.emptySliceValue {
		return result, nil
	}

	separator := ","
	for _, val := range strings.Split(value, separator) {
		v, err := fieldsInfo.stringToInt(field, val)
		if err != nil {
			return nil, err
		}
		result = append(result, int32(*v))
	}

	return result, nil
}

func (fieldsInfo *objectFieldsInfo) intSliceToString(field string, values []int32) (result string, err error) {

	fieldInfo, ok := fieldsInfo.fieldsManual[field]

	if !ok {
		return "", fmt.Errorf("Field should be define")
	}

	if len(values) == 0 {
		return fieldInfo.emptySliceValue, nil
	}

	separator := ","
	for i, value := range values {
		if i > 0 {
			result += separator
		}
		v := int(value)
		val, err := fieldsInfo.intToString(field, &v)
		if err != nil {
			return "", err
		}
		result += val
	}

	return result, nil
}

func defaultStringOfEnumsCheck(fieldname string) func(*objectFieldsInfo, interface{}) error {
	return func(fieldsInfo *objectFieldsInfo, v interface{}) error {
		s, ok := v.(string)
		if ok {
			if s == "" {
				return nil
			}

			for _, sv := range strings.Split(s, ",") {

				i, err := fieldsInfo.stringToInt(fieldname, sv)
				if err != nil {
					return err
				}
				err = fieldsInfo.intCheckSetValue(fieldname, i)
				if err != nil {
					return err
				}
			}
			return nil
		}
		return fmt.Errorf("defaultStringOfEnumsCheck: Unsupported type for value %v", v)
	}
}

func defaultStringCompare(fieldsInfo *objectFieldsInfo, old, new string) bool {
	return old == new
}
