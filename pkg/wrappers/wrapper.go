package wrappers

import (
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func Int64ToTF(v *wrapperspb.Int64Value) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(v.Value)
}

func BoolToTF(v *wrapperspb.BoolValue) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(v.Value)
}

func BoolFromTF(v types.Bool) *wrapperspb.BoolValue {
	if !IsPresent(v) {
		return nil
	}
	return &wrappers.BoolValue{Value: v.ValueBool()}
}

func Int64FromTF(v types.Int64) *wrapperspb.Int64Value {
	if !IsPresent(v) {
		return nil
	}
	return &wrappers.Int64Value{Value: v.ValueInt64()}
}

func StringFromTF(v types.String) string {
	if !IsPresent(v) {
		return ""
	}
	return v.ValueString()
}

func IsPresent[T attr.Value](v T) bool {
	return !v.IsNull() && !v.IsUnknown()
}
