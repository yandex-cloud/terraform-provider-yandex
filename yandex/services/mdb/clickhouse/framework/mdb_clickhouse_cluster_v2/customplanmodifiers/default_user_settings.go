package customplanmodifiers

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DefaultUserSettingsPlanModifier() planmodifier.Object {
	return &defaultUserSettingsPlanModifierStruct{}
}

type defaultUserSettingsPlanModifierStruct struct{}

func (m *defaultUserSettingsPlanModifierStruct) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	configAttrs := map[string]attr.Value{}
	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		configAttrs = req.ConfigValue.Attributes()
	}

	stateAttrs := req.StateValue.Attributes()
	planAttrs := maps.Clone(req.PlanValue.Attributes())
	for setting := range planAttrs {
		if configValue, ok := configAttrs[setting]; ok && !configValue.IsNull() {
			continue
		}

		if stateValue, ok := stateAttrs[setting]; ok {
			planAttrs[setting] = stateValue
		}
	}

	newPlanValue, diags := types.ObjectValue(req.PlanValue.AttributeTypes(ctx), planAttrs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = newPlanValue
}

func (m *defaultUserSettingsPlanModifierStruct) Description(context.Context) string {
	return `
		Default user settings plan modifier.
		Keep the settings that are absent in the configuration as they are in the state,
		since the API returns only the settings that are set.
	`
}

func (m *defaultUserSettingsPlanModifierStruct) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
