package customplanmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

const requiresReplaceOnDisableDescription = "The cluster is recreated when the attribute is disabled, " +
	"since the API supports enabling it only."

func RequiresReplaceOnDisable() planmodifier.Bool {
	return boolplanmodifier.RequiresReplaceIf(
		func(ctx context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
			if req.StateValue.IsNull() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
				return
			}

			resp.RequiresReplace = req.StateValue.ValueBool() && !req.PlanValue.ValueBool()
		},
		requiresReplaceOnDisableDescription,
		requiresReplaceOnDisableDescription,
	)
}
