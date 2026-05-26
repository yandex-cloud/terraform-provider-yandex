package planmodify

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

func DashboardsNodeGroupsChanges(ctx context.Context, planConfig, stateConfig *model.Config, diags *diag.Diagnostics) bool {
	if planConfig.Dashboards.Equal(stateConfig.Dashboards) {
		return false
	}

	tflog.Debug(ctx, "planConfig.Dashboards potentially have been changed")

	planDashboardsBlock, stateDashboardsBlock, d := model.ParseGenerics(ctx, planConfig, stateConfig, model.ParseDashboardSubConfig)
	diags.Append(d...)
	if diags.HasError() {
		return false
	}

	if stateDashboardsBlock != nil && planDashboardsBlock == nil {
		tflog.Debug(ctx, "Detected changes in config.dashboards.node_groups: state wasn't nil but plan is nil")
		return true
	}

	if stateDashboardsBlock == nil && planDashboardsBlock != nil {
		tflog.Debug(ctx, "Detected changes in config.dashboards.node_groups: state is nil but plan isn't")
		return true
	}

	if planDashboardsBlock.NodeGroups.Equal(stateDashboardsBlock.NodeGroups) {
		tflog.Debug(ctx, "config.dashboards.node_groups not changed")
		return false
	}

	tflog.Debug(ctx, "Detected changes in config.dashboards.node_groups")

	var nodeGroupsState []model.DashboardNode
	var nodeGroupsPlan []model.DashboardNode
	diags.Append(stateDashboardsBlock.NodeGroups.ElementsAs(ctx, &nodeGroupsState, false)...)
	diags.Append(planDashboardsBlock.NodeGroups.ElementsAs(ctx, &nodeGroupsPlan, false)...)
	if diags.HasError() {
		return false
	}

	isHostsChanged := false

	planGroupsByName := model.GetGroupByName(nodeGroupsPlan)
	// If some node group is not in the plan, it means that it was deleted
	for _, stateNodeGroup := range nodeGroupsState {
		if _, ok := planGroupsByName[stateNodeGroup.Name.ValueString()]; !ok {
			isHostsChanged = true
			break
		}
	}

	oldGroupsByName := model.GetGroupByName(nodeGroupsState)
	for i := range nodeGroupsPlan {
		// If some node group is not in the state, it means that it was added
		stateNodeGroup, ok := oldGroupsByName[nodeGroupsPlan[i].Name.ValueString()]
		if !ok {
			isHostsChanged = true
			continue
		}

		// sync disk_size with disk_size_gb for opensearch nodegroup
		nodeGroupsPlan[i].Resources = syncResourcesDiskSize(ctx, nodeGroupsPlan[i].Resources, diags)
		if diags.HasError() {
			return false
		}

		if nodeGroupsPlan[i].Equal(stateNodeGroup) {
			continue
		}

		tflog.Debug(ctx, fmt.Sprintf("Detected changes in '%s' node_group", nodeGroupsPlan[i].GetName()))
		isHostsChanged = true
	}

	planDashboardsBlock.NodeGroups, d = types.ListValueFrom(ctx, model.DashboardNodeType, nodeGroupsPlan)
	diags.Append(d...)
	if diags.HasError() {
		return false
	}

	planConfig.Dashboards, d = types.ObjectValueFrom(ctx, model.DashboardsSubConfigAttrTypes, planDashboardsBlock)
	diags.Append(d...)
	if diags.HasError() {
		return false
	}

	return isHostsChanged
}
