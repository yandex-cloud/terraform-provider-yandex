package customplanmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func FullVersionPlanModifier() planmodifier.String {
	return &fullVersionModifier{}
}

type fullVersionModifier struct{}

func (m *fullVersionModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	var configVersion types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("version"), &configVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if configVersion.IsNull() || configVersion.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}

	var stateVersion types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("version"), &stateVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if configVersion.Equal(stateVersion) {
		resp.PlanValue = req.StateValue
	}
}

func (m *fullVersionModifier) Description(context.Context) string {
	return "Keeps prior state value when version is unchanged; marks unknown when version changes."
}

func (m *fullVersionModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
