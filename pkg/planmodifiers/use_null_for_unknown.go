package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NullWriteOnlyMap() planmodifier.Map {
	return nullWriteOnlyModifier{}
}

func NullWriteOnlySet() planmodifier.Set {
	return nullWriteOnlyModifier{}
}

func NullWriteOnlyList() planmodifier.List {
	return nullWriteOnlyModifier{}
}

func NullWriteOnlyString() planmodifier.String {
	return nullWriteOnlyModifier{}
}

func NullWriteOnlyInt64() planmodifier.Int64 {
	return nullWriteOnlyModifier{}
}

func NullWriteOnlyBool() planmodifier.Bool {
	return nullWriteOnlyModifier{}
}

func NullWriteOnlyFloat64() planmodifier.Float64 {
	return nullWriteOnlyModifier{}
}

func NullWriteOnlyObject() planmodifier.Object {
	return nullWriteOnlyModifier{}
}

type nullWriteOnlyModifier struct{}

func (_ nullWriteOnlyModifier) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.Float64Null()
	}
}

func (_ nullWriteOnlyModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.BoolNull()
	}
}

func (_ nullWriteOnlyModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.Int64Null()
	}
}

func (_ nullWriteOnlyModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.StringNull()
	}
}

func (_ nullWriteOnlyModifier) Description(context.Context) string {
	return "For setting unknown attributes that api doesn't return."
}

func (_ nullWriteOnlyModifier) MarkdownDescription(context.Context) string {
	return "For setting unknown attributes that api doesn't return."
}

func (_ nullWriteOnlyModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.MapNull(req.PlanValue.ElementType(ctx))
	}
}

func (_ nullWriteOnlyModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.SetNull(req.PlanValue.ElementType(ctx))
	}
}

func (_ nullWriteOnlyModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.ListNull(req.PlanValue.ElementType(ctx))
	}
}

func (_ nullWriteOnlyModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.ObjectNull(req.PlanValue.AttributeTypes(ctx))
	}
}
