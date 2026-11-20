package mdb_mysql_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type mysqlConfigStateForUnknown struct{}

func (m mysqlConfigStateForUnknown) Description(context.Context) string {
	return "Use previous settings only when the MySQL version is unchanged."
}

func (m mysqlConfigStateForUnknown) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m mysqlConfigStateForUnknown) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if req.ConfigValue.IsNull() && req.PlanValue.IsUnknown() && !req.StateValue.IsNull() {
		var stateVersion, planVersion types.String
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("version"), &stateVersion)...)
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("version"), &planVersion)...)
		if resp.Diagnostics.HasError() || !planVersion.Equal(stateVersion) {
			// An upgrade may rename or remove settings. Read them after the operation.
			return
		}
	}
	mapplanmodifier.UseStateForUnknown().PlanModifyMap(ctx, req, resp)
}
