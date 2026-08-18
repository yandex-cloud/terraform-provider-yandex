package chcommon

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func MakeReversedMap(m map[int32]string, addMap map[string]int32) map[string]int32 {
	r := addMap
	for k, v := range m {
		r[v] = k
	}
	return r
}

func MakeEnumNamesValidator(m map[int32]string) []validator.String {
	res := make([]string, 0, len(m))
	for _, val := range m {
		res = append(res, val)
	}
	return []validator.String{stringvalidator.OneOf(res...)}
}

func IsProtoMessageEmpty(m protoreflect.Message) bool {
	if m == nil {
		return true
	}

	empty := true

	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		empty = false
		return false
	})

	return empty
}

func Int64FromWrapper(value *wrapperspb.Int64Value) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(value.GetValue())
}

func BoolFromWrapper(value *wrapperspb.BoolValue) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(value.Value)
}

func Float64FromWrapper(value *wrapperspb.DoubleValue) types.Float64 {
	if value == nil {
		return types.Float64Null()
	}
	return types.Float64Value(value.Value)
}

func NullableString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func WrapInt64(value types.Int64) *wrapperspb.Int64Value {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return &wrapperspb.Int64Value{Value: value.ValueInt64()}
}

func WrapBool(value types.Bool) *wrapperspb.BoolValue {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return &wrapperspb.BoolValue{Value: value.ValueBool()}
}

func WrapDouble(value types.Float64) *wrapperspb.DoubleValue {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return &wrapperspb.DoubleValue{Value: value.ValueFloat64()}
}

func WrapString(value types.String) string {
	return value.ValueString()
}
