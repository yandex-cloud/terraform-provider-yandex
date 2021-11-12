package yandex

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

func TestDynamicSetRead(t *testing.T) {
	t.Parallel()

	us := &postgresql.UserSettings{}

	us.TempFileLimit = &wrappers.Int64Value{Value: 10}

	rv := reflect.ValueOf(us)
	rv = rv.Elem()

	for i := 0; i < rv.NumField(); i++ {
		f := rv.Type().Field(i)

		tg, okTg := findTag(f, "protobuf", "name")

		if okTg {
			if tg == "default_transaction_isolation" {
				v := 4
				err := setIntFieldToReflect(rv, f.Name, &v)
				if err != nil {
					t.Error(err)
				}
			}
			if tg == "lock_timeout" {
				v := 7
				err := setIntFieldToReflect(rv, f.Name, &v)
				if err != nil {
					t.Error(err)
				}
			}

			if tg == "temp_file_limit" {
				err := setIntFieldToReflect(rv, f.Name, nil)
				if err != nil {
					t.Error(err)
				}
			}

			if tg == "log_statement" {
				err := setIntFieldToReflect(rv, f.Name, nil)
				if err == nil {
					t.Error("setIntValueToReflect fail: Insert nil into not nil field")
				}
			}
		}

	}

	if us.LockTimeout == nil {
		t.Error("setIntValueToReflect fail: not set value")
	}

	if us.LockTimeout.GetValue() != 7 {
		t.Error("setIntValueToReflect fail: value set not correct in *wrappers.Int64Value")
	}

	if us.DefaultTransactionIsolation != 4 {
		t.Error("setIntValueToReflect fail: not set value in int")
	}

	if us.TempFileLimit != nil {
		t.Error("setIntValueToReflect fail: not set nil in *wrappers.Int64Value")
	}

	for i := 0; i < rv.NumField(); i++ {
		f := rv.Type().Field(i)

		tg, okTg := findTag(f, "protobuf", "name")

		if okTg {
			if tg == "default_transaction_isolation" {
				vl, err := getValueFromReflect(rv, f.Name)
				if err != nil {
					t.Error(err)
				}
				if vl.(int) != 4 {
					t.Error("getValueFromReflect fail: read not correct value from int")
				}
			}
			if tg == "lock_timeout" {
				vl, err := getValueFromReflect(rv, f.Name)
				if err != nil {
					t.Error(err)
				}
				if vl.(int) != 7 {
					t.Error("getValueFromReflect fail: read not correct value from *wrappers.Int64Value")
				}
			}

			if tg == "temp_file_limit" {
				vl, err := getValueFromReflect(rv, f.Name)
				if err != nil {
					t.Error(err)
				}
				if vl != nil {
					t.Error("getValueFromReflect read not corect nil value from *wrappers.Int64Value")
				}
			}

		}

	}

}

type TestStruct struct {
	A int32
	B int32 `tag_test:"varint,1,opt,name=b_name,h_name=bla,abc"`
	C int   `json:"default_transaction_isolation,omitempty"`

	D *wrappers.Int64Value `tag_test:"name=d_name"`

	E bool                `tag_test:"name=e_name"`
	F *wrappers.BoolValue `tag_test:"name=f_name"`

	G float64               `tag_test:"name=g_name"`
	H *wrappers.DoubleValue `tag_test:"name=h_name"`

	I string `tag_test:"name=i_name"`
}

func TestDynamicFindTagGetCorrect(t *testing.T) {
	t.Parallel()

	testStruct := TestStruct{}

	rv := reflect.ValueOf(testStruct)

	fl, ok := rv.Type().FieldByName("B")

	if !ok {
		t.Error("Not found field B")
	}

	vl, ok := findTag(fl, "tag_test", "name")

	if !ok {
		t.Error("Tag \"tag_test\"-\"name\" not found")
	}

	if vl != "b_name" {
		t.Errorf("Tag \"tag_test\"-\"name\" value shuld be \"b_name\", value %v is not correct", vl)
	}
}

func TestDynamicFindTagNotFound(t *testing.T) {
	t.Parallel()

	testStruct := TestStruct{}

	rv := reflect.ValueOf(testStruct)

	fl, ok := rv.Type().FieldByName("A")

	if !ok {
		t.Error("Not found field A")
	}

	_, ok = findTag(fl, "tag_test", "name")

	if ok {
		t.Error("Tag \"tag_test\"-\"name\" should be not found")
	}
}

func TestDynamicFindTagGetFlag(t *testing.T) {
	t.Parallel()

	testStruct := TestStruct{}

	rv := reflect.ValueOf(testStruct)

	fl, ok := rv.Type().FieldByName("C")

	if !ok {
		t.Error("Not found field C")
	}

	vl, ok := findTag(fl, "json", "omitempty")

	if !ok {
		t.Error("Tag \"json\"-\"omitempty\" should be found")
	}

	if vl != "" {
		t.Error("Tag value \"json\"-\"omitempty\" should be empty (\"\")")
	}
}

func TestDynamicSetIntValue(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}
	value := 6

	err := setIntField(testStruct, "A", &value)

	if err != nil {
		t.Errorf("setIntField: set value to A (int) fail, set should return nil error when pass value 6, but error: %v", err)
	}

	if testStruct.A != 6 {
		t.Error("setIntField: set value to A (int) fail, value is not setted")
	}

	err = setIntField(testStruct, "A", nil)

	if err == nil {
		t.Error("setIntField: set value to A (int) fail, value is setted nil into int")
	}

	err = setIntField(nil, "A", &value)

	if err == nil {
		t.Error("setIntField: set value to A (int) fail, value is setted into nil object")
	}
}

func TestDynamicSetWrappersInt64(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}
	value := 6

	err := setIntField(testStruct, "D", &value)

	if err != nil {
		t.Errorf("setIntField: set value to D (*wrappers.Int64Value) fail, set should return nil error when pass value 6, but error: %v", err)
	}

	if testStruct.D == nil {
		t.Errorf("setIntField: set value to D (*wrappers.Int64Value) fail, value (6) is not setted")
	}
	if testStruct.D.GetValue() != 6 {
		t.Errorf("setIntField: set value to D (*wrappers.Int64Value) fail, value setted not correct should be 6 but setted: %v", testStruct.D.GetValue())
	}

	err = setIntField(testStruct, "D", nil)

	if err != nil {
		t.Errorf("setIntField: set value to D (*wrappers.Int64Value) fail, set should return nil error when pass value nil, but error: %v", err)
	}

	if testStruct.D != nil {
		t.Errorf("setIntField: set value to D (*wrappers.Int64Value) fail, value nil is not setted")
	}

}

func TestDynamicSetBoolValue(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}
	value := true

	err := setBoolField(testStruct, "E", &value)

	if err != nil {
		t.Errorf("setBoolField: set value to E (bool) fail, set should return nil error when pass value true, but error: %v", err)
		t.Error(err)
	}

	if !testStruct.E {
		t.Error("setBoolField: set value to E (bool) fail, value is not setted")
	}

	err = setBoolField(testStruct, "E", nil)

	if err == nil {
		t.Error("setBoolField: set value to E (bool) fail, value is setted nil into bool")
	}

	err = setBoolField(nil, "E", &value)

	if err == nil {
		t.Error("setBoolField: set value to E (bool) fail, value is setted into nil object")
	}
}

func TestDynamicSetWrappersBool(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}
	value := true

	err := setBoolField(testStruct, "F", &value)

	if err != nil {
		t.Errorf("setIntField: set value to F (*wrappers.BoolValue) fail, set should return nil error when pass value true, but error: %v", err)
	}

	if testStruct.F == nil {
		t.Errorf("setIntField: set value to F (*wrappers.BoolValue) fail, value (true) is not setted")
	}
	if !testStruct.F.GetValue() {
		t.Errorf("setIntField: set value to F (*wrappers.BoolValue) fail, value setted not correct should be true but setted: %v", testStruct.F.GetValue())
	}

	err = setBoolField(testStruct, "F", nil)

	if err != nil {
		t.Errorf("setIntField: set value to F (*wrappers.BoolValue) fail, set should return nil error when pass value nil, but error: %v", err)
	}

	if testStruct.F != nil {
		t.Errorf("setIntField: set value to F (*wrappers.BoolValue) fail, value nil is not setted")
	}

}

func TestDynamicSetFloatValue(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}
	value := 7.6

	err := setFloatField(testStruct, "G", &value)

	if err != nil {
		t.Errorf("setFloatField: set value to G (float64) fail, set should return nil error when pass value 7.6, but error: %v", err)
	}

	if testStruct.G != 7.6 {
		t.Error("setFloatField: set value to G (float64) fail, value is not setted")
	}

	err = setFloatField(testStruct, "G", nil)

	if err == nil {
		t.Error("setFloatField: set value to G (float64) fail, value is setted nil into float")
	}

	err = setFloatField(nil, "G", &value)

	if err == nil {
		t.Error("setFloatField: set value to G (float64) fail, value is setted into nil object")
	}
}

func TestDynamicSetWrappersFloat(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}
	value := 7.6

	err := setFloatField(testStruct, "H", &value)

	if err != nil {
		t.Errorf("setFloatField: set value to H (*wrappers.DoubleValue) fail, set should return nil error when pass value 7.6, but error: %v", err)
	}

	if testStruct.H == nil {
		t.Errorf("setFloatField: set value to H (*wrappers.DoubleValue) fail, value (7.6) is not setted")
	}
	if testStruct.H.GetValue() != 7.6 {
		t.Errorf("setFloatField: set value to H (*wrappers.DoubleValue) fail, value setted not correct should be 6 but setted: %v", testStruct.H.GetValue())
	}

	err = setFloatField(testStruct, "H", nil)

	if err != nil {
		t.Errorf("setFloatField: set value to H (*wrappers.DoubleValue) fail, set should return nil error when pass value nil, but error: %v", err)
	}

	if testStruct.H != nil {
		t.Errorf("setFloatField: set value to H (*wrappers.DoubleValue) fail, value nil is not setted")
	}

}

func TestDynamicSetStringValue(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}
	value := "some text"

	err := setStringField(testStruct, "I", &value)

	if err != nil {
		t.Errorf("setStringField: set value to I (string) fail, set should return nil error when pass value \"some text\", but error: %v", err)
		t.Error(err)
	}

	if testStruct.I != "some text" {
		t.Error("setStringField: set value to I (string) fail, value is not setted")
	}

	err = setStringField(testStruct, "I", nil)

	if err == nil {
		t.Error("setStringField: set value to I (string) fail, value is setted nil into string")
	}

	err = setStringField(nil, "I", &value)

	if err == nil {
		t.Error("setStringField: set value to I (string) fail, value is setted into nil object")
	}
}

func TestDynamicGetValueFromInt(t *testing.T) {
	t.Parallel()

	testStruct := TestStruct{
		A: 6,
	}

	v, err := getValueFrom(testStruct, "A")
	if err != nil {
		t.Error(err)
	}

	if v == nil {
		t.Error("Geted value is nil int")
	}

	vi, ok := v.(int)
	if !ok {
		t.Error("Fail to covert geted value into int")
	}

	if vi != 6 {
		t.Errorf("Geted value has not correct value 6 != %v", vi)
	}
}

func TestDynamicGetValueFromWrappersInt64(t *testing.T) {
	t.Parallel()

	testStruct := TestStruct{}

	v, err := getValueFrom(testStruct, "D")
	if err != nil {
		t.Error(err)
	}
	if v != nil {
		t.Errorf("Geted value must be nil (wrappers.Int64Value) = \"%v\"", v)
	}

	testStruct.D = &wrappers.Int64Value{Value: 6}

	v, err = getValueFrom(testStruct, "D")
	if err != nil {
		t.Error(err)
	}

	if v == nil {
		t.Error("Geted value is nil *wrappers.Int64Value")
	}

	vi, ok := v.(int)
	if !ok {
		t.Error("Fail to covert geted value into int (*wrappers.Int64Value)")
	}

	if vi != 6 {
		t.Errorf("Geted value has not correct value 6 (*wrappers.Int64Value) != %v", vi)
	}
}

func TestDynamicGetValueFromBool(t *testing.T) {
	t.Parallel()

	testStruct := TestStruct{
		E: true,
	}

	v, err := getValueFrom(testStruct, "E")
	if err != nil {
		t.Error(err)
	}

	if v == nil {
		t.Error("Geted value is nil bool")
	}

	vb, ok := v.(bool)
	if !ok {
		t.Error("Fail to covert geted value into bool")
	}

	if !vb {
		t.Errorf("Geted value has not correct value true != %v", vb)
	}
}

func TestDynamicGetValueFromWrappersBool(t *testing.T) {
	t.Parallel()

	testStruct := TestStruct{}

	v, err := getValueFrom(testStruct, "F")
	if err != nil {
		t.Error(err)
	}
	if v != nil {
		t.Errorf("Geted value must be nil (wrappers.BoolValue) = \"%v\"", v)
	}

	testStruct.F = &wrappers.BoolValue{Value: true}

	v, err = getValueFrom(testStruct, "F")
	if err != nil {
		t.Error(err)
	}

	if v == nil {
		t.Error("Geted value is nil *wrappers.BoolValue")
	}

	vb, ok := v.(bool)
	if !ok {
		t.Error("Fail to covert geted value into bool (*wrappers.BoolValue)")
	}

	if !vb {
		t.Errorf("Geted value has not correct value true (*wrappers.BoolValue) != %v", vb)
	}

}

func TestDynamicGetValueFromFloat(t *testing.T) {
	t.Parallel()

	val := TestStruct{
		G: 6.5,
	}

	v, err := getValueFrom(val, "G")
	if err != nil {
		t.Error(err)
	}

	if v == nil {
		t.Error("Geted value is nil float64")
	}

	vf, ok := v.(float64)
	if !ok {
		t.Error("Fail to covert geted value into float64")
	}

	if vf != 6.5 {
		t.Errorf("Geted value has not correct value 6.5 != %v", vf)
	}
}

func TestDynamicGetValueFromWrappersFloat(t *testing.T) {
	t.Parallel()

	testStruct := TestStruct{}

	v, err := getValueFrom(testStruct, "H")
	if err != nil {
		t.Error(err)
	}
	if v != nil {
		t.Errorf("Geted value must be nil (wrappers.Float64Value) = \"%v\"", v)
	}

	testStruct.H = &wrappers.DoubleValue{Value: 6.5}

	v, err = getValueFrom(testStruct, "H")
	if err != nil {
		t.Error(err)
	}

	if v == nil {
		t.Error("Geted value is nil *wrappers.DoubleValue")
	}

	vf, ok := v.(float64)
	if !ok {
		t.Error("Fail to covert geted value into float64 (*wrappers.DoubleValue)")
	}

	if vf != 6.5 {
		t.Errorf("Geted value has not correct value 6.5 (*wrappers.DoubleValue) != %v", vf)
	}
}

func TestDynamicGetValueFromString(t *testing.T) {
	t.Parallel()

	val := TestStruct{
		I: "str",
	}

	v, err := getValueFrom(val, "I")
	if err != nil {
		t.Error(err)
	}

	if v == nil {
		t.Error("Geted value is nil string")
	}

	vs, ok := v.(string)
	if !ok {
		t.Error("Fail to covert geted value into string")
	}

	if vs != "str" {
		t.Errorf("Geted value has not correct value \"str\" != \"%v\"", vs)
	}
}

func TestDynamicGetStructType(t *testing.T) {
	t.Parallel()

	val := TestStruct{
		A: 6,
		E: true,
		G: 6.5,
		I: "str",
	}

	tp, err := getStructType(val)
	if err != nil {
		t.Error(err)
	}
	tpp, err := getStructType(&val)
	if err != nil {
		t.Error(err)
	}

	if tp != tpp {
		t.Errorf("Type get fail types must be equal %v != %v", tp, tpp)
	}
}

func TestDynamicGetStructValue(t *testing.T) {
	t.Parallel()

	val := TestStruct{
		A: 6,
		E: true,
		G: 6.5,
		I: "str",
	}

	sv, err := getStructValue(val)
	if err != nil {
		t.Error(err)
	}
	svp, err := getStructValue(&val)
	if err != nil {
		t.Error(err)
	}

	if sv.Type() != svp.Type() {
		t.Errorf("Value type get fail types must be equal %v != %v", sv.Type(), svp.Type())
	}

	if svp.Type() != reflect.TypeOf(val) {
		t.Errorf("Value type get fail types. Type must be not ptr. Must be equal %v != %v", sv.Type(), reflect.TypeOf(&val))
	}
}

func TestDynamicGetFieldsInfo(t *testing.T) {
	t.Parallel()

	testStruct := TestStruct{}

	m, err := getFieldsInfo(testStruct, "tag_test", "name")

	if err != nil {
		t.Error(err)
	}

	if len(m) != 7 {
		t.Errorf("must be found 7 fields not %v: %v", len(m), m)
	}

	if m["b_name"].valueType != schema.TypeInt {
		t.Errorf("field b_name must be int, not %v", m["b_name"].valueType)
	}
	if m["d_name"].valueType != schema.TypeInt {
		t.Errorf("field d_name must be int, not %v", m["d_name"].valueType)
	}
	if m["e_name"].valueType != schema.TypeBool {
		t.Errorf("field e_name must be bool, not %v", m["e_name"].valueType)
	}
	if m["f_name"].valueType != schema.TypeBool {
		t.Errorf("field f_name must be bool, not %v", m["f_name"].valueType)
	}
	if m["g_name"].valueType != schema.TypeFloat {
		t.Errorf("field g_name must be float, not %v", m["g_name"].valueType)
	}
	if m["h_name"].valueType != schema.TypeFloat {
		t.Errorf("field h_name must be float, not %v", m["h_name"].valueType)
	}
	if m["i_name"].valueType != schema.TypeString {
		t.Errorf("field i_name must be string, not %v", m["i_name"].valueType)
	}
}
