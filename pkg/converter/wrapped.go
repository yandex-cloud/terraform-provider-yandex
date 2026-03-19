package converter

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func WrappedBool(val types.Bool) *wrapperspb.BoolValue {
	if val.IsUnknown() || val.IsNull() {
		return nil
	}
	return wrapperspb.Bool(val.ValueBool())
}

func WrappedString(val types.String) *wrapperspb.StringValue {
	if val.IsUnknown() || val.IsNull() {
		return nil
	}
	return wrapperspb.String(val.ValueString())
}

func WrappedDouble(val types.Float64) *wrapperspb.DoubleValue {
	if val.IsUnknown() || val.IsNull() {
		return nil
	}
	return wrapperspb.Double(val.ValueFloat64())
}

func WrappedInt64(val types.Int64) *wrapperspb.Int64Value {
	if val.IsUnknown() || val.IsNull() {
		return nil
	}
	return wrapperspb.Int64(val.ValueInt64())
}
