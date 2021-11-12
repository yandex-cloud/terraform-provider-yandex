package yandex

import (
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

func TestExpandResourceGenerateOneFieldInt(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := newObjectFieldsInfo().
		addEnumHumanNames("b_name", mdbPGUserSettingsTransactionIsolationName,
			postgresql.UserSettings_TransactionIsolation_name)

	fiB := fieldReflectInfo{name: "B", valueType: schema.TypeInt}

	err := expandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, 6, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass value 6 (with flag skipNil=true), but error: %v", err)
	}
	err = expandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, 6, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass value 6 (with flag skipNil=false), but error: %v", err)
	}

	if testStruct.B != 6 {
		t.Errorf("Expand fail: expand into field B (int) should change value to 6, not to %v", testStruct.B)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, nil, false, "some.path.for.error")
	if !isNilNotAllowedError(err) {
		t.Errorf("Expand fail: expand into field B (int) should return an error when pass value nil with flag skipNil=false")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, 6.7, true, "some.path.for.error")
	if !isTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field B (int) should return an error when pass value float")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, "some text", true, "some.path.for.error")
	if !isTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field B (int) should return an error when pass value random string")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, "read committed", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass string value from names enum (read committed), but error: %v", err)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, "TRANSACTION_ISOLATION_READ_UNCOMMITTED", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass string value from names enum (TRANSACTION_ISOLATION_READ_UNCOMMITTED), but error: %v", err)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiB, "b_name", testStruct, "6", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field B (int) should return nil error when pass string value convertable to int (\"6\"), but error: %v", err)
	}

}

func TestExpandResourceGenerateOneFieldWrappersInt64(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := newObjectFieldsInfo()

	fiD := fieldReflectInfo{name: "D", valueType: schema.TypeInt}

	err := expandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, 7, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should return nil error when pass value 7, but error: %v", err)
	}

	if testStruct.D == nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should change value to 7, not nil")
	} else if testStruct.D.GetValue() != 7 {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should change value to 7, not to %v", testStruct.D.GetValue())
	}

	testStruct.D = &wrappers.Int64Value{Value: 7}

	err = expandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	if testStruct.D.GetValue() != 7 {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should skip changes when pass value nil with flag skipNil=true, but value: %v", testStruct.D.GetValue())
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, "", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should return nil error when pass value \"\" (empty string) with flag skipNil=true, but error: %v", err)
	}

	if testStruct.D.GetValue() != 7 {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should skip changes when pass value \"\" (empty string) with flag skipNil=true, but value: %v", testStruct.D.GetValue())
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, nil, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should return nil error when pass value nil with flag skipNil=false, but error: %v", err)
	}

	if testStruct.D != nil {
		t.Errorf("Expand fail: expand into field D (*wrappers.Int64Value) should change value to nil when pass value nil with flag skipNil=false, but value: %v", testStruct.D.GetValue())
	}

	testStruct.D = &wrappers.Int64Value{Value: 7}

	err = expandResourceGenerateOneField(fieldsInfo, fiD, "d_name", testStruct, "", false, "some.path.for.error")
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

	fieldsInfo := newObjectFieldsInfo()

	fiE := fieldReflectInfo{name: "E", valueType: schema.TypeBool}

	err := expandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, true, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field E (bool) should return nil error when pass value true (with flag skipNil=true), but error: %v", err)
	}
	err = expandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, true, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field E (bool) should return nil error when pass value true (with flag skipNil=false), but error: %v", err)
	}

	if !testStruct.E {
		t.Errorf("Expand fail: expand into field E (bool) should change value to true, not to %v", testStruct.E)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field E (bool) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, nil, false, "some.path.for.error")
	if !isNilNotAllowedError(err) {
		t.Errorf("Expand fail: expand into field E (bool) should return an error when pass value nil with flag skipNil=false")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, 6.7, true, "some.path.for.error")
	if !isTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field E (bool) should return an error when pass value float")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, "some text", true, "some.path.for.error")
	if !isTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field E (bool) should return an error when pass value random string")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiE, "e_name", testStruct, "true", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field E (bool) should return nil error when pass string value convertable to bool (\"true\"), but error: %v", err)
	}

}

func TestExpandResourceGenerateOneFieldWrappersBool(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := newObjectFieldsInfo()

	fiF := fieldReflectInfo{name: "F", valueType: schema.TypeBool}

	err := expandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, true, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should return nil error when pass value true, but error: %v", err)
	}

	if testStruct.F == nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should change value to true, not nil")
	} else if !testStruct.F.GetValue() {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should change value to true, not to %v", testStruct.F.GetValue())
	}

	testStruct.F = &wrappers.BoolValue{Value: true}

	err = expandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	if !testStruct.F.GetValue() {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should skip changes when pass value nil with flag skipNil=true, but value: %v", testStruct.F.GetValue())
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, "", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should return nil error when pass value \"\" (empty string) with flag skipNil=true, but error: %v", err)
	}

	if !testStruct.F.GetValue() {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should skip changes when pass value \"\" (empty string) with flag skipNil=true, but value: %v", testStruct.F.GetValue())
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, nil, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should return nil error when pass value nil with flag skipNil=false, but error: %v", err)
	}

	if testStruct.F != nil {
		t.Errorf("Expand fail: expand into field F (*wrappers.BoolValue) should change value to nil when pass value nil with flag skipNil=false, but value: %v", testStruct.F.GetValue())
	}

	testStruct.F = &wrappers.BoolValue{Value: true}

	err = expandResourceGenerateOneField(fieldsInfo, fiF, "f_name", testStruct, "", false, "some.path.for.error")
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

	fieldsInfo := newObjectFieldsInfo()

	fiG := fieldReflectInfo{name: "G", valueType: schema.TypeFloat}

	err := expandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, 7.6, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field G (float64) should return nil error when pass value 7.6 (with flag skipNil=true), but error: %v", err)
	}
	err = expandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, 7.6, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field G (float64) should return nil error when pass value 7.6 (with flag skipNil=false), but error: %v", err)
	}

	if testStruct.G != 7.6 {
		t.Errorf("Expand fail: expand into field G (float64) should change value to 7.6, not to %v", testStruct.G)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field G (float64) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, nil, false, "some.path.for.error")
	if !isNilNotAllowedError(err) {
		t.Errorf("Expand fail: expand into field G (float64) should return an error when pass value nil with flag skipNil=false")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, true, true, "some.path.for.error")
	if !isTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field G (float64) should return an error when pass value bool")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, "some text", true, "some.path.for.error")
	if !isTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field G (float64) should return an error when pass value random string")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiG, "g_name", testStruct, "7.6", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field G (float64) should return nil error when pass string value convertable to int (\"7.6\"), but error: %v", err)
	}

}

func TestExpandResourceGenerateOneFieldWrappersFloat(t *testing.T) {
	t.Parallel()

	testStruct := &TestStruct{}

	fieldsInfo := newObjectFieldsInfo()

	fiH := fieldReflectInfo{name: "H", valueType: schema.TypeFloat}

	err := expandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, 7.8, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should return nil error when pass value 7.8, but error: %v", err)
	}

	if testStruct.H == nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should change value to 7.8, not nil")
	} else if testStruct.H.GetValue() != 7.8 {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should change value to 7.8, not to %v", testStruct.H.GetValue())
	}

	testStruct.H = &wrappers.DoubleValue{Value: 7.8}

	err = expandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	if testStruct.H.GetValue() != 7.8 {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should skip changes when pass value nil with flag skipNil=true, but value: %v", testStruct.H.GetValue())
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, "", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should return nil error when pass value \"\" (empty string) with flag skipNil=true, but error: %v", err)
	}

	if testStruct.H.GetValue() != 7.8 {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should skip changes when pass value \"\" (empty string) with flag skipNil=true, but value: %v", testStruct.H.GetValue())
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, nil, false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should return nil error when pass value nil with flag skipNil=false, but error: %v", err)
	}

	if testStruct.H != nil {
		t.Errorf("Expand fail: expand into field H (*wrappers.DoubleValue) should change value to nil when pass value nil with flag skipNil=false, but value: %v", testStruct.H.GetValue())
	}

	testStruct.H = &wrappers.DoubleValue{Value: 7.8}

	err = expandResourceGenerateOneField(fieldsInfo, fiH, "h_name", testStruct, "", false, "some.path.for.error")
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

	fieldsInfo := newObjectFieldsInfo()

	fiI := fieldReflectInfo{name: "I", valueType: schema.TypeString}

	err := expandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, "some text", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field I (string) should return nil error when pass text value (\"some text\") (with flag skipNil=true), but error: %v", err)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, "some text", false, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field I (string) should return nil error when pass text value (\"some text\") (with flag skipNil=false), but error: %v", err)
	}

	if testStruct.I != "some text" {
		t.Errorf("Expand fail: expand into field I (string) should change value to \"some text\", not to %v", testStruct.I)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, nil, true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field I (string) should return nil error when pass value nil with flag skipNil=true, but error: %v", err)
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, nil, false, "some.path.for.error")
	if !isNilNotAllowedError(err) {
		t.Errorf("Expand fail: expand into field I (string) should return an error when pass value nil with flag skipNil=false")
	}

	err = expandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, true, true, "some.path.for.error")
	if !isTypeMismatchError(err) {
		t.Errorf("Expand fail: expand into field I (string) should return an error when pass value bool")
	}

	testStruct.I = "some text"

	// "" is nil value for string
	err = expandResourceGenerateOneField(fieldsInfo, fiI, "i_name", testStruct, "", true, "some.path.for.error")
	if err != nil {
		t.Errorf("Expand fail: expand into field I (string) should return nil error when pass text value (\"\") with flag skipNil=true, but error: %v", err)
	}

	if testStruct.I != "some text" {
		t.Errorf("Expand fail: expand into field I (string) should skip changes when pass text value (\"\") with flag skipNil=true and field value should be \"some text\", not \"%v\" empty string must be ignored", testStruct.I)
	}

}

func TestFieldsDynamicGenerateMapSchemaValidateFuncCorrect(t *testing.T) {
	t.Parallel()

	validateFunc := generateMapSchemaValidateFunc(mdbPGSettingsFieldsInfo)

	value := map[string]interface{}{
		"autovacuum_vacuum_scale_factor":    "0.32",
		"default_transaction_isolation":     "TRANSACTION_ISOLATION_READ_UNCOMMITTED",
		"enable_parallel_hash":              "true",
		"max_connections":                   "395",
		"vacuum_cleanup_index_scale_factor": "0.2",
	}

	_, errors := validateFunc(value, "")

	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("generateMapSchemaValidateFunc: Error on correct example: %v", err)
		}
	}

}

func TestFieldsDynamicGenerateMapSchemaValidateFuncFail(t *testing.T) {
	t.Parallel()

	validateFunc := generateMapSchemaValidateFunc(mdbPGSettingsFieldsInfo)

	value := map[string]interface{}{
		// type is not correct string but should be int
		"autovacuum_vacuum_scale_factor": "a0.32",
		// Value not in enum
		"default_transaction_isolation": "TRANSACTION_OLATION_READ_UNCOMMITTED",
		// type is not correct int but should be bool
		"enable_parallel_hash": "5",
		// type is not correct bool but should be int
		"max_connections": "true",
		// type is not correct string but should be float
		"vacuum_cleanup_index_scale_factor": "p0.2",
	}

	_, errors := validateFunc(value, "")

	if len(errors) != 5 {
		t.Errorf("generateMapSchemaValidateFunc: Error on not correct example should be 5 errors, but out only: %v", len(errors))
		for _, err := range errors {
			t.Errorf("generateMapSchemaValidateFunc: Error on not correct example was: %v", err)
		}
	}

}

func TestFieldsDynamicGenerateMapSchemaDiffSuppressFuncFloat(t *testing.T) {
	t.Parallel()

	suppressFunc := generateMapSchemaDiffSuppressFunc(mdbPGSettingsFieldsInfo)

	if suppressFunc("ttt.autovacuum_vacuum_scale_factor", "0.32", "0.33", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: float values should be not equal (0.32 != 0.33)")
	}

	if !suppressFunc("ttt.autovacuum_vacuum_scale_factor", "0.32", "0.32", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: float values should be equal (0.32 == 0.32)")
	}

	if !suppressFunc("ttt.autovacuum_vacuum_scale_factor", "0.32", "", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: float values should be equal when new value is empty")
	}
}

func TestFieldsDynamicGenerateMapSchemaDiffSuppressFuncInt(t *testing.T) {
	t.Parallel()

	suppressFunc := generateMapSchemaDiffSuppressFunc(mdbPGSettingsFieldsInfo)

	if suppressFunc("ttt.max_connections", "32", "33", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be not equal (32 != 33)")
	}

	if !suppressFunc("ttt.max_connections", "32", "32", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be equal (32 == 32)")
	}

	if !suppressFunc("ttt.max_connections", "32", "", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be equal when new value is empty")
	}
}

func TestFieldsDynamicGenerateMapSchemaDiffSuppressFuncBool(t *testing.T) {
	t.Parallel()

	suppressFunc := generateMapSchemaDiffSuppressFunc(mdbPGSettingsFieldsInfo)

	if suppressFunc("ttt.enable_parallel_hash", "true", "false", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be not equal (true != false)")
	}

	if !suppressFunc("ttt.enable_parallel_hash", "true", "true", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be equal (true == true)")
	}

	if !suppressFunc("ttt.enable_parallel_hash", "true", "", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be equal when new value is empty")
	}
}
func TestFieldsDynamicGenerateMapSchemaDiffSuppressFuncEnum(t *testing.T) {
	t.Parallel()

	suppressFunc := generateMapSchemaDiffSuppressFunc(mdbPGUserSettingsFieldsInfo)

	if suppressFunc("ttt.default_transaction_isolation", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", "read committed", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: enum values should be not equal (TRANSACTION_ISOLATION_READ_UNCOMMITTED != read committed)")
	}

	if !suppressFunc("ttt.default_transaction_isolation", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", "read uncommitted", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: enum values should be equal (TRANSACTION_ISOLATION_READ_UNCOMMITTED == read uncommitted)")
	}

	if !suppressFunc("ttt.default_transaction_isolation", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: enum values should be equal (TRANSACTION_ISOLATION_READ_UNCOMMITTED == TRANSACTION_ISOLATION_READ_UNCOMMITTED)")
	}

	if !suppressFunc("ttt.default_transaction_isolation", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", "", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: enum values should be equal when new value is empty")
	}
}
