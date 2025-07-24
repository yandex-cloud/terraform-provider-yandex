package planmodifiers

import (
	"context"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/converter"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func DurationPlanModifier() planmodifier.String {
	return &durationPlanModifier{}
}

type durationPlanModifier struct{}

func (d *durationPlanModifier) Description(ctx context.Context) string {
	return "Ensures that attribute_one and attribute_two attributes are kept synchronised."
}

func (d *durationPlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *durationPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	d1 := converter.ParseDuration(req.StateValue.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	d2 := converter.ParseDuration(req.PlanValue.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if d1 == nil && d2 == nil {
		return
	}

	if d1 != nil && d2 != nil && d1.Seconds == d2.Seconds && d1.Nanos == d2.Nanos {
		resp.PlanValue = req.StateValue
		return
	}
}
