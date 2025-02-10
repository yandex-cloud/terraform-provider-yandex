package protobuf_adapter

import (
	"context"
	"fmt"
	"maps"
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestYandexProvider_MDBCommonProtobufFillFull(t *testing.T) {

	t.Parallel()
	f := NewProtobufMapDataAdapter()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        map[string]attr.Value
		expectedVal   TestMessage
		expectedError bool
	}{
		{
			testname: "CheckFullFillWithBasicAttrValues",
			reqVal: map[string]attr.Value{
				"string_field": types.StringValue("string_value"),
				"int32_field":  types.Int64Value(15),
				"int64_field":  types.Int64Value(30),
				"bool_field":   types.BoolValue(true),

				"repeated_string_field": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("string_value")}),
				"repeated_int32_field":  types.ListValueMust(types.Int64Type, []attr.Value{types.Int64Value(1), types.Int64Value(2)}),
				"repeated_int64_field":  types.ListValueMust(types.Int64Type, []attr.Value{types.Int64Value(1), types.Int64Value(2)}),
				"repeated_bool_field":   types.ListValueMust(types.BoolType, []attr.Value{types.BoolValue(true), types.BoolValue(false)}),

				"string_nested_field": types.StringValue("string_value_2"),
				"int32_nested_field":  types.Int64Value(1),

				"enum_field": types.Int64Value(2),

				"string_wrapper_field": types.StringValue("string_value_3"),
				"int32_wrapper_field":  types.Int64Value(2),
				"int64_wrapper_field":  types.Int64Value(3),
				"bool_wrapper_field":   types.BoolValue(true),
			},
			expectedVal: TestMessage{
				StringField: "string_value",
				Int32Field:  15,
				Int64Field:  30,
				BoolField:   true,

				RepeatedStringField: []string{"string_value"},
				RepeatedInt32Field:  []int32{1, 2},
				RepeatedInt64Field:  []int64{1, 2},
				RepeatedBoolField:   []bool{true, false},

				NestedMessageField: &TestMessage_NestedMessage{
					StringNestedField: "string_value_2",
					Int32NestedField:  1,
				},
				EnumField:          EnumType_SECOND_VALUE,
				StringWrapperField: wrapperspb.String("string_value_3"),
				Int32WrapperField:  wrapperspb.Int32(2),
				Int64WrapperField:  wrapperspb.Int64(3),
				BoolWrapperField:   wrapperspb.Bool(true),
			},
		},
		{
			testname: "CheckFullFillWithCommonAttrValues",
			reqVal: map[string]attr.Value{

				"int32_field": types.NumberValue(big.NewFloat(1)),
				"int64_field": types.NumberValue(big.NewFloat(2)),

				"repeated_int32_field": types.TupleValueMust(
					[]attr.Type{types.NumberType, types.NumberType},
					[]attr.Value{types.NumberValue(big.NewFloat(1)), types.NumberValue(big.NewFloat(2))},
				),
				"repeated_int64_field": types.TupleValueMust(
					[]attr.Type{types.NumberType, types.NumberType},
					[]attr.Value{types.NumberValue(big.NewFloat(1)), types.NumberValue(big.NewFloat(2))},
				),

				"int32_nested_field": types.NumberValue(big.NewFloat(3)),

				"enum_field": types.NumberValue(big.NewFloat(1)),

				"int32_wrapper_field": types.NumberValue(big.NewFloat(4)),
				"int64_wrapper_field": types.NumberValue(big.NewFloat(5)),
			},
			expectedVal: TestMessage{
				Int32Field: 1,
				Int64Field: 2,

				RepeatedInt32Field: []int32{1, 2},
				RepeatedInt64Field: []int64{1, 2},
				NestedMessageField: &TestMessage_NestedMessage{
					Int32NestedField: 3,
				},
				EnumField: EnumType_FIRST_VALUE,

				Int32WrapperField: wrapperspb.Int32(4),
				Int64WrapperField: wrapperspb.Int64(5),
			},
		},
		{
			testname: "CheckFullFillWithNullAttrValues",
			reqVal: map[string]attr.Value{

				"int32_field": types.NumberNull(),
				"int64_field": types.NumberNull(),

				"repeated_int32_field": types.TupleNull(
					[]attr.Type{types.NumberType, types.NumberType},
				),
				"repeated_int64_field": types.TupleNull(
					[]attr.Type{types.NumberType, types.NumberType},
				),

				"int32_nested_field": types.Int64Null(),

				"enum_field": types.Int64Null(),

				"int32_wrapper_field": types.Int64Null(),
				"int64_wrapper_field": types.Int64Null(),
			},
			expectedVal: TestMessage{
				NestedMessageField: &TestMessage_NestedMessage{},
			},
		},
		{
			testname: "CheckPartFillWithBasicAttrValues",
			reqVal: map[string]attr.Value{
				"int64_field": types.Int64Value(1),
				"repeated_int32_field": types.ListValueMust(
					types.Int64Type,
					[]attr.Value{types.Int64Value(1), types.Int64Value(2)},
				),
				"string_wrapper_field": types.StringValue("string"),
			},
			expectedVal: TestMessage{
				NestedMessageField: &TestMessage_NestedMessage{},
				Int64Field:         1,
				RepeatedInt32Field: []int32{1, 2},
				StringWrapperField: wrapperspb.String("string"),
			},
		},
	}

	for _, c := range cases {
		var diags diag.Diagnostics
		obj := TestMessage{}

		f.Fill(ctx, &obj, c.reqVal, &diags)
		if c.expectedError != diags.HasError() {
			if diags.HasError() {
				t.Errorf("Unexpected fill error: %v\n", diags.Errors())
			}
			t.Errorf("Unexpected fill error status: expected %v, actual %v", c.expectedError, diags.HasError())
			continue
		}
		if !c.expectedError && !reflect.DeepEqual(&obj, &c.expectedVal) {
			t.Errorf("Unexpected result: expected %+v, actual %+v", c.expectedVal, obj)
		}
	}
}

func TestYandexProvider_MDBCommonProtobufFillOneField(t *testing.T) {

	t.Parallel()
	f := NewProtobufMapDataAdapter()
	ctx := context.Background()

	basicValues := map[string]attr.Value{
		"string":      types.StringValue("string_value"),
		"int":         types.Int64Value(math.MaxInt64 - 100),
		"bool":        types.BoolValue(true),
		"string_list": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("string_value")}),
		"int_list":    types.ListValueMust(types.Int64Type, []attr.Value{types.Int64Value(1), types.Int64Value(2)}),
	}

	fieldCheckExcludeValueFromBasic := map[string]string{
		"string_field":          "string",
		"int32_field":           "int",
		"int64_field":           "int",
		"bool_field":            "bool",
		"repeated_string_field": "string_list",
		"repeated_int32_field":  "int_list",
		"repeated_int64_field":  "int_list",
		"string_nested_field":   "string",
		"string_wrapper_field":  "string",
		"int32_wrapper_field":   "int",
		"int64_wrapper_field":   "int",
		"bool_wrapper_field":    "bool",
	}

	type Case struct {
		testname      string
		reqVal        map[string]attr.Value
		expectedError bool
	}

	cases := []Case{}

	for field, excludedKey := range fieldCheckExcludeValueFromBasic {
		excludedValues := maps.Clone(basicValues)

		cases = append(cases, Case{
			testname: fmt.Sprintf("CheckValidField_%s", field),
			reqVal: map[string]attr.Value{
				field: excludedValues[excludedKey],
			},
		})

		delete(excludedValues, excludedKey)

		for _, v := range excludedValues {
			cases = append(cases, Case{
				testname: fmt.Sprintf("CheckInvalidFillFieldWithBasicAttrValuesExclude_%s_%s", field, v.Type(ctx).String()),
				reqVal: map[string]attr.Value{
					field: v,
				},
				expectedError: true,
			})
		}
	}

	cases = append(cases, Case{
		testname: "CheckInt32FillWithInt64",
		reqVal: map[string]attr.Value{
			"int32_field": types.Int64Value(math.MaxInt64 - 20),
		},
	})

	for _, c := range cases {
		var diags diag.Diagnostics
		obj := TestMessage{}

		f.Fill(ctx, &obj, c.reqVal, &diags)

		t.Logf("Run %s", c.testname)
		if c.expectedError != diags.HasError() {
			if diags.HasError() {
				t.Errorf("Unexpected fill error: %v\n", diags.Errors())
			}
			t.Errorf("Unexpected fill error status: expected %v, actual %v", c.expectedError, diags.HasError())
			continue
		}
	}
}

func TestYandexProvider_AdapterProtobufExtract(t *testing.T) {
	t.Parallel()
	f := NewProtobufMapDataAdapter()
	ctx := context.Background()

	fullReq := TestMessage{
		StringField: "string_value",
		Int32Field:  15,
		Int64Field:  30,
		BoolField:   true,

		RepeatedStringField: []string{"string_value_1", "string_value_2"},
		RepeatedInt32Field:  []int32{1, 2},
		RepeatedInt64Field:  []int64{3, 4},
		RepeatedBoolField:   []bool{true, false},

		NestedMessageField: &TestMessage_NestedMessage{
			StringNestedField: "string_value_2",
			Int32NestedField:  1,
		},
		EnumField:          EnumType_SECOND_VALUE,
		StringWrapperField: wrapperspb.String("string_value_3"),
		Int32WrapperField:  wrapperspb.Int32(2),
		Int64WrapperField:  wrapperspb.Int64(3),
		BoolWrapperField:   wrapperspb.Bool(true),
	}

	fullResult := map[string]attr.Value{
		"string_field": types.StringValue("string_value"),
		"int32_field":  types.NumberValue(big.NewFloat(15)),
		"int64_field":  types.NumberValue(big.NewFloat(30)),
		"bool_field":   types.BoolValue(true),
		"repeated_string_field": types.TupleValueMust(
			[]attr.Type{types.StringType, types.StringType},
			[]attr.Value{types.StringValue("string_value_1"), types.StringValue("string_value_2")},
		),
		"repeated_int32_field": types.TupleValueMust(
			[]attr.Type{types.NumberType, types.NumberType},
			[]attr.Value{types.NumberValue(big.NewFloat(1)), types.NumberValue(big.NewFloat(2))},
		),
		"repeated_int64_field": types.TupleValueMust(
			[]attr.Type{types.NumberType, types.NumberType},
			[]attr.Value{types.NumberValue(big.NewFloat(3)), types.NumberValue(big.NewFloat(4))},
		),
		"repeated_bool_field": types.TupleValueMust(
			[]attr.Type{types.BoolType, types.BoolType},
			[]attr.Value{types.BoolValue(true), types.BoolValue(false)},
		),
		"string_nested_field":  types.StringValue("string_value_2"),
		"int32_nested_field":   types.NumberValue(big.NewFloat(1)),
		"enum_field":           types.NumberValue(big.NewFloat(2)),
		"string_wrapper_field": types.StringValue("string_value_3"),
		"int32_wrapper_field":  types.NumberValue(big.NewFloat(2)),
		"int64_wrapper_field":  types.NumberValue(big.NewFloat(3)),
		"bool_wrapper_field":   types.BoolValue(true),
	}

	cases := []struct {
		testname      string
		reqVal        any
		expectedVal   map[string]attr.Value
		expectedError bool
	}{
		{
			testname:    "CheckFullExtract",
			reqVal:      fullReq,
			expectedVal: fullResult,
		},
		{
			testname:    "CheckFullPtrExtract",
			reqVal:      &fullReq,
			expectedVal: fullResult,
		},
		{
			testname: "CheckEmptyExtract",
			reqVal: TestMessage{
				NestedMessageField: &TestMessage_NestedMessage{},
			},
			expectedVal: map[string]attr.Value{
				"string_field": types.StringValue(""),
				"int32_field":  types.NumberValue(big.NewFloat(0)),
				"int64_field":  types.NumberValue(big.NewFloat(0)),
				"bool_field":   types.BoolValue(false),
				"repeated_string_field": types.TupleNull(
					[]attr.Type{},
				),
				"repeated_int32_field": types.TupleNull(
					[]attr.Type{},
				),
				"repeated_int64_field": types.TupleNull(
					[]attr.Type{},
				),
				"repeated_bool_field": types.TupleNull(
					[]attr.Type{},
				),
				"string_nested_field":  types.StringValue(""),
				"int32_nested_field":   types.NumberValue(big.NewFloat(0)),
				"enum_field":           types.NumberValue(big.NewFloat(0)),
				"string_wrapper_field": types.StringNull(),
				"int32_wrapper_field":  types.NumberNull(),
				"int64_wrapper_field":  types.NumberNull(),
				"bool_wrapper_field":   types.BoolNull(),
			},
			expectedError: false,
		},
		{
			testname: "CheckPartlyExtract",
			reqVal: TestMessage{
				StringField:        "string_value",
				RepeatedInt32Field: []int32{1, 2},
				NestedMessageField: &TestMessage_NestedMessage{
					StringNestedField: "string_value_2",
				},
				BoolWrapperField: wrapperspb.Bool(true),
			},
			expectedVal: map[string]attr.Value{
				"string_field": types.StringValue("string_value"),
				"int32_field":  types.NumberValue(big.NewFloat(0)),
				"int64_field":  types.NumberValue(big.NewFloat(0)),
				"bool_field":   types.BoolValue(false),
				"repeated_string_field": types.TupleNull(
					[]attr.Type{},
				),
				"repeated_int32_field": types.TupleValueMust(
					[]attr.Type{types.NumberType, types.NumberType},
					[]attr.Value{types.NumberValue(big.NewFloat(1)), types.NumberValue(big.NewFloat(2))},
				),
				"repeated_int64_field": types.TupleNull(
					[]attr.Type{},
				),
				"repeated_bool_field": types.TupleNull(
					[]attr.Type{},
				),
				"string_nested_field":  types.StringValue("string_value_2"),
				"int32_nested_field":   types.NumberValue(big.NewFloat(0)),
				"enum_field":           types.NumberValue(big.NewFloat(0)),
				"string_wrapper_field": types.StringNull(),
				"int32_wrapper_field":  types.NumberNull(),
				"int64_wrapper_field":  types.NumberNull(),
				"bool_wrapper_field":   types.BoolValue(true),
			},
		},
		{
			testname: "CheckEmptyCollectionsExtract",
			reqVal: TestMessage{

				RepeatedInt32Field:  []int32{},
				RepeatedInt64Field:  []int64{},
				RepeatedBoolField:   []bool{},
				RepeatedStringField: []string{},

				NestedMessageField: &TestMessage_NestedMessage{},
			},
			expectedVal: map[string]attr.Value{
				"string_field": types.StringValue(""),
				"int32_field":  types.NumberValue(big.NewFloat(0)),
				"int64_field":  types.NumberValue(big.NewFloat(0)),
				"bool_field":   types.BoolValue(false),
				"repeated_string_field": types.TupleValueMust(
					[]attr.Type{}, []attr.Value{},
				),
				"repeated_int32_field": types.TupleValueMust(
					[]attr.Type{}, []attr.Value{},
				),
				"repeated_int64_field": types.TupleValueMust(
					[]attr.Type{}, []attr.Value{},
				),
				"repeated_bool_field": types.TupleValueMust(
					[]attr.Type{}, []attr.Value{},
				),
				"string_nested_field":  types.StringValue(""),
				"int32_nested_field":   types.NumberValue(big.NewFloat(0)),
				"enum_field":           types.NumberValue(big.NewFloat(0)),
				"string_wrapper_field": types.StringNull(),
				"int32_wrapper_field":  types.NumberNull(),
				"int64_wrapper_field":  types.NumberNull(),
				"bool_wrapper_field":   types.BoolNull(),
			},
		},
	}

	for _, c := range cases {
		var diags diag.Diagnostics

		m := f.Extract(ctx, c.reqVal, &diags)

		t.Logf("Run %s", c.testname)
		if c.expectedError != diags.HasError() {
			if diags.HasError() {
				t.Errorf("Unexpected fill error: %v\n", diags.Errors())
			}
			t.Errorf("Unexpected fill error status: expected %v, actual %v", c.expectedError, diags.HasError())
			continue
		}

		if m == nil {
			t.Error("Unexpected result: nil")
			continue
		}

		if len(m) != len(c.expectedVal) {
			t.Errorf("Unexpected len result: expected %d attributes, actual %d", len(c.expectedVal), len(m))
		}

		for k, v := range c.expectedVal {
			if val, ok := m[k]; !ok || !val.Equal(v) {
				if !ok {
					t.Errorf("Unexpected result: expected attribute %s is not provided in result", k)
				} else {
					t.Errorf("Unexpected result for field %s: expected %s, actual %s", k, v.String(), val.String())
				}
			}
		}
	}
}
