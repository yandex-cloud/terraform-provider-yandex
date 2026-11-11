package mdb_redis_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// labelsUnknownToEmptyModifier is a plan modifier for the labels map attribute.
// When the plan value is unknown (labels not specified in config), it replaces
// the unknown value with an empty map. This prevents a perpetual diff between
// the empty map returned by the API (stored in state) and the unknown plan value.
// Labels are still cleared when explicitly removed from config because the plan
// value becomes {} (empty map) rather than the previous state value.
type labelsUnknownToEmptyModifier struct{}

func (m labelsUnknownToEmptyModifier) Description(_ context.Context) string {
	return "Replaces unknown labels plan value with an empty map to prevent perpetual diffs."
}

func (m labelsUnknownToEmptyModifier) MarkdownDescription(_ context.Context) string {
	return "Replaces unknown labels plan value with an empty map to prevent perpetual diffs."
}

func (m labelsUnknownToEmptyModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// Only act when the plan value is unknown (attribute not set in config).
	if !req.PlanValue.IsUnknown() {
		return
	}
	// Replace unknown with an empty map so the plan is stable.
	emptyMap, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.PlanValue = emptyMap
}
