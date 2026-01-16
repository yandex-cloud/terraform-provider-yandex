package cdn_resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseUnknownOnUpdate returns a plan modifier that sets the attribute to unknown
// during updates. This is useful for computed fields that change on every update
// (like updated_at timestamps) where the new value is determined by the API.
//
// Unlike UseStateForUnknown() which preserves the state value in the plan,
// this modifier always marks the value as unknown during updates, preventing
// "Provider produced inconsistent result" errors when the API returns a new value.
func UseUnknownOnUpdate() planmodifier.String {
	return useUnknownOnUpdateModifier{}
}

type useUnknownOnUpdateModifier struct{}

func (m useUnknownOnUpdateModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If there's no state (create operation), leave as unknown
	if req.State.Raw.IsNull() {
		return
	}

	// If the plan is being destroyed, no need to modify
	if req.Plan.Raw.IsNull() {
		return
	}

	// For updates: always mark as unknown since the value will change
	// This prevents "Provider produced inconsistent result" errors
	resp.PlanValue = types.StringUnknown()
}

func (m useUnknownOnUpdateModifier) Description(context.Context) string {
	return "Sets the attribute to unknown during updates, as the value will be computed by the API."
}

func (m useUnknownOnUpdateModifier) MarkdownDescription(context.Context) string {
	return "Sets the attribute to unknown during updates, as the value will be computed by the API."
}
