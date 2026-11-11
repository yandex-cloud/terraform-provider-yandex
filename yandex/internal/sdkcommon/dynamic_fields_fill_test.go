package sdkcommon

import (
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

// Human-readable names are test input for the generic enum conversion.
var transactionIsolationHumanNames = map[int]string{
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_UNSPECIFIED):      "unspecified",
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_READ_UNCOMMITTED): "read uncommitted",
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_READ_COMMITTED):   "read committed",
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_REPEATABLE_READ):  "repeatable read",
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_SERIALIZABLE):     "serializable",
}

func TestExpandResourceGenerateOneFieldInt(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := NewObjectFieldsInfo().
		AddEnumHumanNames("b_name", transactionIsolationHumanNames,
			postgresql.UserSettings_TransactionIsolation_name)

	fiB := FieldReflectInfo{name: "B", valueType: schema.TypeInt}

	err := ExpandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, 6, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass value 6 (with flag skipNil=true), but error: %v", err)
	}
	err = ExpandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, 6, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass value 6 (with flag skipNil=false), but error: %v", err)
	}

	if testStruct.B != 6 {
		t.Errorf("Expand fail: expand into field B (int) should change value to 6, not to %v", testStruct.B)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, nil, false, "some.path.for.error")
	if !IsNilNotAllowedError(err) {
		t.Errorf("Expand fail: expand into field B (int) should return an error when pass value nil with flag skipNil=false")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, 6.7, true, "some.path.for.error")
	if !IsTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field B (int) should return an error when pass value float")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, "some text", true, "some.path.for.error")
	if !IsTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field B (int) should return an error when pass value random string")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, "read committed", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass string value from names enum (read committed), but error: %v", err)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, "TRANSACTION_ISOLATION_READ_UNCOMMITTED", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass string value from names enum (TRANSACTION_ISOLATION_READ_UNCOMMITTED), but error: %v", err)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, "6", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass string value convertable to int (\"6\"), but error: %v", err)
	}

}

func TestExpandResourceGenerateOneFieldWrappersInt64(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := NewObjectFieldsInfo()

	fiD := FieldReflectInfo{name: "D", valueType: schema.TypeInt}

	err := ExpandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, 7, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should return nil error when pass value 7, but error: %v", err)
	}

	if testStruct.D == nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should change value to 7, not nil")
	} else if testStruct.D.GetValue() != 7 {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should change value to 7, not to %v", testStruct.D.GetValue())
	}

	testStruct.D = &wrappers.Int64Value{Value: 7}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	if testStruct.D.GetValue() != 7 {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should skip changes when pass value nil with flag skipNil=true, but value: %v", testStruct.D.GetValue())
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, "", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should return nil error when pass value \"\" (empty string) with flag skipNil=true, but error: %v", err)
	}

	if testStruct.D.GetValue() != 7 {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should skip changes when pass value \"\" (empty string) with flag skipNil=true, but value: %v", testStruct.D.GetValue())
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, nil, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should return nil error when pass value nil with flag skipNil=false, but error: %v", err)
	}

	if testStruct.D != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should change value to nil when pass value nil with flag skipNil=false, but value: %v", testStruct.D.GetValue())
	}

	testStruct.D = &wrappers.Int64Value{Value: 7}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, "", false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should return nil error when pass value \"\" (empty string) with flag skipNil=false, but error: %v", err)
	}

	if testStruct.D != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should change value to nil when pass value \"\" (empty string) with flag skipNil=false, but value: %v", testStruct.D.GetValue())
	}

}

func TestExpandResourceGenerateOneFieldBool(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := NewObjectFieldsInfo()

	fiE := FieldReflectInfo{name: "E", valueType: schema.TypeBool}

	err := ExpandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, true, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field E (bool) should return nil error when pass value true (with flag skipNil=true), but error: %v", err)
	}
	err = ExpandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, true, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field E (bool) should return nil error when pass value true (with flag skipNil=false), but error: %v", err)
	}

	if !testStruct.E {
		t.Errorf("Expand fail: expand into field E (bool) should change value to true, not to %v", testStruct.E)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field E (bool) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, nil, false, "some.path.for.error")
	if !IsNilNotAllowedError(err) {
		t.Errorf("Expand fail: expand into field E (bool) should return an error when pass value nil with flag skipNil=false")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, 6.7, true, "some.path.for.error")
	if !IsTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field E (bool) should return an error when pass value float")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, "some text", true, "some.path.for.error")
	if !IsTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field E (bool) should return an error when pass value random string")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, "true", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field E (bool) should return nil error when pass string value convertable to bool (\"true\"), but error: %v", err)
	}

}

func TestExpandResourceGenerateOneFieldWrappersBool(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := NewObjectFieldsInfo()

	fiF := FieldReflectInfo{name: "F", valueType: schema.TypeBool}

	err := ExpandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, true, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should return nil error when pass value true, but error: %v", err)
	}

	if testStruct.F == nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should change value to true, not nil")
	} else if !testStruct.F.GetValue() {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should change value to true, not to %v", testStruct.F.GetValue())
	}

	testStruct.F = &wrappers.BoolValue{Value: true}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	if !testStruct.F.GetValue() {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should skip changes when pass value nil with flag skipNil=true, but value: %v", testStruct.F.GetValue())
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, "", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should return nil error when pass value \"\" (empty string) with flag skipNil=true, but error: %v", err)
	}

	if !testStruct.F.GetValue() {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should skip changes when pass value \"\" (empty string) with flag skipNil=true, but value: %v", testStruct.F.GetValue())
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, nil, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should return nil error when pass value nil with flag skipNil=false, but error: %v", err)
	}

	if testStruct.F != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should change value to nil when pass value nil with flag skipNil=false, but value: %v", testStruct.F.GetValue())
	}

	testStruct.F = &wrappers.BoolValue{Value: true}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, "", false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should return nil error when pass value \"\" (empty string) with flag skipNil=false, but error: %v", err)
	}

	if testStruct.F != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should change value to nil when pass value \"\" (empty string) with flag skipNil=false, but value: %v", testStruct.F.GetValue())
	}

}

func TestExpandResourceGenerateOneFieldFloat(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := NewObjectFieldsInfo()

	fiG := FieldReflectInfo{name: "G", valueType: schema.TypeFloat}

	err := ExpandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, 7.6, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field G (float64) should return nil error when pass value 7.6 (with flag skipNil=true), but error: %v", err)
	}
	err = ExpandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, 7.6, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field G (float64) should return nil error when pass value 7.6 (with flag skipNil=false), but error: %v", err)
	}

	if testStruct.G != 7.6 {
		t.Errorf("Expand fail: expand into field G (float64) should change value to 7.6, not to %v", testStruct.G)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field G (float64) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, nil, false, "some.path.for.error")
	if !IsNilNotAllowedError(err) {
		t.Errorf("Expand fail: expand into field G (float64) should return an error when pass value nil with flag skipNil=false")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, true, true, "some.path.for.error")
	if !IsTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field G (float64) should return an error when pass value bool")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, "some text", true, "some.path.for.error")
	if !IsTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field G (float64) should return an error when pass value random string")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, "7.6", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field G (float64) should return nil error when pass string value convertable to int (\"7.6\"), but error: %v", err)
	}

}

func TestExpandResourceGenerateOneFieldWrappersFloat(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := NewObjectFieldsInfo()

	fiH := FieldReflectInfo{name: "H", valueType: schema.TypeFloat}

	err := ExpandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, 7.8, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should return nil error when pass value 7.8, but error: %v", err)
	}

	if testStruct.H == nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should change value to 7.8, not nil")
	} else if testStruct.H.GetValue() != 7.8 {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should change value to 7.8, not to %v", testStruct.H.GetValue())
	}

	testStruct.H = &wrappers.DoubleValue{Value: 7.8}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	if testStruct.H.GetValue() != 7.8 {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should skip changes when pass value nil with flag skipNil=true, but value: %v", testStruct.H.GetValue())
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, "", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should return nil error when pass value \"\" (empty string) with flag skipNil=true, but error: %v", err)
	}

	if testStruct.H.GetValue() != 7.8 {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should skip changes when pass value \"\" (empty string) with flag skipNil=true, but value: %v", testStruct.H.GetValue())
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, nil, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should return nil error when pass value nil with flag skipNil=false, but error: %v", err)
	}

	if testStruct.H != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should change value to nil when pass value nil with flag skipNil=false, but value: %v", testStruct.H.GetValue())
	}

	testStruct.H = &wrappers.DoubleValue{Value: 7.8}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, "", false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should return nil error when pass value \"\" (empty string) with flag skipNil=false, but error: %v", err)
	}

	if testStruct.H != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should change value to nil when pass value \"\" (empty string) with flag skipNil=false, but value: %v", testStruct.H.GetValue())
	}

}

func TestExpandResourceGenerateOneFieldString(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := NewObjectFieldsInfo()

	fiI := FieldReflectInfo{name: "I", valueType: schema.TypeString}

	err := ExpandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, "some text", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field I (string) should return nil error when pass text value (\"some text\") (with flag skipNil=true), but error: %v", err)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, "some text", false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field I (string) should return nil error when pass text value (\"some text\") (with flag skipNil=false), but error: %v", err)
	}

	if testStruct.I != "some text" {
		t.Errorf("Expand fail: expand into field I (string) should change value to \"some text\", not to %v", testStruct.I)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field I (string) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, nil, false, "some.path.for.error")
	if !IsNilNotAllowedError(err) {
		t.Errorf("Expand fail: expand into field I (string) should return an error when pass value nil with flag skipNil=false")
	}

	err = ExpandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, true, true, "some.path.for.error")
	if !IsTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field I (string) should return an error when pass value bool")
	}

	testStruct.I = "some text"

	// "" is nil value for string
	err = ExpandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, "", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field I (string) should return nil error when pass text value (\"\") with flag skipNil=true, but error: %v", err)
	}

	if testStruct.I != "some text" {
		t.Errorf("Expand fail: expand into field I (string) should skip changes when pass text value (\"\") with flag skipNil=true and field value should be \"some text\", not \"%v\" empty string must be ignored", testStruct.I)
	}

}
