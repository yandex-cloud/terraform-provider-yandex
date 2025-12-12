package mdb_greenplum_cluster_v2

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func flattenBoolWrapper(wb *wrapperspb.BoolValue) types.Bool {
	if wb == nil {
		return types.BoolNull()
	}
	return types.BoolValue(wb.GetValue())
}

func flattenInt64Wrapper(wi *wrapperspb.Int64Value) types.Int64 {
	if wi == nil {
		return types.Int64Null()
	}
	return types.Int64Value(wi.GetValue())
}

func flattenFloat64Wrapper(wf *wrapperspb.DoubleValue) types.Float64 {
	if wf == nil {
		return types.Float64Null()
	}
	return types.Float64Value(wf.GetValue())
}

func flattenEnum[T interface {
	Number() protoreflect.EnumNumber
	String() string
}](e T) types.String {
	if e.Number() == 0 {
		return types.StringNull()
	}
	return types.StringValue(e.String())
}

func makeDefaultEmptyObjectAttrs(attrTypes map[string]attr.Type) map[string]attr.Value {
	result := make(map[string]attr.Value, len(attrTypes))
	for name, ttype := range attrTypes {
		switch ttype {
		case types.StringType:
			result[name] = types.StringNull()
		case types.Int64Type:
			result[name] = types.Int64Null()
		case types.BoolType:
			result[name] = types.BoolNull()
		case types.Float64Type:
			result[name] = types.Float64Null()
		default:
			panic("unknown type")
		}
	}
	return result
}

func expandBoolWrapper(b types.Bool) *wrapperspb.BoolValue {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}
	return wrapperspb.Bool(b.ValueBool())
}

func expandStringWrapper(b types.String) *wrapperspb.StringValue {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}
	return wrapperspb.String(b.ValueString())
}

func expandInt64Wrapper(b types.Int64) *wrapperspb.Int64Value {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}
	return wrapperspb.Int64(b.ValueInt64())
}

func expandFloat64Wrapper(b types.Float64) *wrapperspb.DoubleValue {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}
	return wrapperspb.Double(b.ValueFloat64())
}
