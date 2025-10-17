package mdb_redis_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// modulesUseStateForUnknownWhenNotConfigured is a custom plan modifier for the modules attribute
// that uses the state value when the attribute is not configured (null or unknown) in the plan
func modulesUseStateForUnknownWhenNotConfigured() planmodifier.Object {
	return modulesUseStateForUnknownWhenNotConfiguredModifier{}
}

// modulesUseStateForUnknownWhenNotConfiguredModifier implements the plan modifier
type modulesUseStateForUnknownWhenNotConfiguredModifier struct{}

// Description returns a human-readable description of the plan modifier
func (m modulesUseStateForUnknownWhenNotConfiguredModifier) Description(_ context.Context) string {
	return "Use the state value for the modules attribute when it is not configured in the plan"
}

// MarkdownDescription returns a markdown description of the plan modifier
func (m modulesUseStateForUnknownWhenNotConfiguredModifier) MarkdownDescription(_ context.Context) string {
	return "Use the state value for the modules attribute when it is not configured in the plan"
}

// PlanModifyObject implements the plan modification logic
func (m modulesUseStateForUnknownWhenNotConfiguredModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// If the plan value is known and not null, use it as is
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	// If the plan value is null and the state value is also null, keep it as null
	if req.PlanValue.IsNull() && (req.StateValue.IsNull() || req.StateValue.IsUnknown()) {
		return
	}

	// If the plan value is null but the state value is not null/unknown, use the state value
	// This handles the case where modules are not specified in config but API returns default modules
	if req.PlanValue.IsNull() && !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}

	// If the plan value is unknown, use the state value if it's known
	if req.PlanValue.IsUnknown() && !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}
}
