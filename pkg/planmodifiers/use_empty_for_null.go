package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NilRelaxedMap() planmodifier.Map {
	return nilRelaxedModifier{}
}

func NilRelaxedSet() planmodifier.Set {
	return nilRelaxedModifier{}
}

func NilRelaxedList() planmodifier.List {
	return nilRelaxedModifier{}
}

func NilRelaxedString() planmodifier.String {
	return nilRelaxedModifier{}
}

func NilRelaxedInt64() planmodifier.Int64 {
	return nilRelaxedModifier{}
}

func NilRelaxedBool() planmodifier.Bool {
	return nilRelaxedModifier{}
}

func NilRelaxedFloat64() planmodifier.Float64 {
	return nilRelaxedModifier{}
}

type nilRelaxedModifier struct{}

func (_ nilRelaxedModifier) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.Float64Null()
	} else if req.PlanValue.IsNull() && req.StateValue.ValueFloat64() == float64(0) {
		resp.PlanValue = req.StateValue
	}
}

func (_ nilRelaxedModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.BoolNull()
	} else if req.PlanValue.IsNull() && req.StateValue.ValueBool() == false {
		resp.PlanValue = req.StateValue
	}
}

func (_ nilRelaxedModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.Int64Null()
	} else if req.PlanValue.IsNull() && req.StateValue.ValueInt64() == 0 {
		resp.PlanValue = req.StateValue
	}
}

func (_ nilRelaxedModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.StringNull()
	} else if req.PlanValue.IsNull() && req.StateValue.ValueString() == "" {
		resp.PlanValue = req.StateValue
	}
}

const desctiprion = "For compatibility with the states created by SDK provider, Terraform consider nil and zero values to be same."

func (_ nilRelaxedModifier) Description(context.Context) string {
	return desctiprion
}

func (_ nilRelaxedModifier) MarkdownDescription(context.Context) string {
	return desctiprion
}

func (_ nilRelaxedModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.MapNull(req.PlanValue.ElementType(ctx))
	} else if req.PlanValue.IsNull() && len(req.StateValue.Elements()) == 0 {
		resp.PlanValue = req.StateValue
	}
}

func (_ nilRelaxedModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.SetNull(req.PlanValue.ElementType(ctx))
	} else if req.PlanValue.IsNull() && len(req.StateValue.Elements()) == 0 {
		resp.PlanValue = req.StateValue
	}
}

func (_ nilRelaxedModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.PlanValue.IsUnknown() {
		resp.PlanValue = types.ListNull(req.PlanValue.ElementType(ctx))
	} else if req.PlanValue.IsNull() && len(req.StateValue.Elements()) == 0 {
		resp.PlanValue = req.StateValue
	}
}
