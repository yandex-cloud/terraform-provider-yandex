package planmodify

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

func OpenSearchNodeGroupsChanges(ctx context.Context, planConfig, stateConfig *model.Config, diags *diag.Diagnostics) bool {
	if planConfig.OpenSearch.Equal(stateConfig.OpenSearch) {
		return false
	}

	tflog.Debug(ctx, "planConfig.OpenSearch potentially have been changed")

	planOpenSearchBlock, stateOpenSearchBlock, d := model.ParseGenerics(ctx, planConfig, stateConfig, model.ParseOpenSearchSubConfig)
	diags.Append(d...)
	if diags.HasError() {
		return false
	}

	if planOpenSearchBlock.NodeGroups.Equal(stateOpenSearchBlock.NodeGroups) {
		tflog.Debug(ctx, "config.opensearch.node_groups not changed")
		return false
	}

	tflog.Debug(ctx, "Detected changes in config.opensearch.node_groups")

	var nodeGroupsState []model.OpenSearchNode
	var nodeGroupsPlan []model.OpenSearchNode
	diags.Append(planOpenSearchBlock.NodeGroups.ElementsAs(ctx, &nodeGroupsPlan, false)...)
	diags.Append(stateOpenSearchBlock.NodeGroups.ElementsAs(ctx, &nodeGroupsState, false)...)
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

		// sync disk_size_limit with disk_size_gb_limit for opensearch nodegroup autoscaling
		nodeGroupsPlan[i].DiskSizeAutoscaling = syncDiskSizeAutoscaling(ctx, nodeGroupsPlan[i].DiskSizeAutoscaling, diags)
		if diags.HasError() {
			return false
		}

		planNodeGroup := nodeGroupsPlan[i]

		if planNodeGroup.Equal(stateNodeGroup) {
			continue
		}

		tflog.Debug(ctx, fmt.Sprintf("Detected changes in '%s' node_group", planNodeGroup.GetName()))
		isHostsChanged = true

		autoscalingOn := utils.IsPresent(attr.Value(stateNodeGroup.DiskSizeAutoscaling))

		if !autoscalingOn {
			continue
		}

		modifiedResources := fixResourcesDiskSizeForAutoscaling(ctx, planNodeGroup.Resources, stateNodeGroup.Resources, autoscalingOn, diags)
		if diags.HasError() {
			return false
		}

		if !planNodeGroup.Resources.Equal(modifiedResources) {
			tflog.Warn(ctx, fmt.Sprintf("Detected changes in '%s' node_group.resources.disk_size but ignore them due to enabled autoscaling", planNodeGroup.GetName()))
		}

		nodeGroupsPlan[i].Resources = modifiedResources
	}

	planOpenSearchBlock.NodeGroups, d = types.ListValueFrom(ctx, model.OpenSearchNodeType, nodeGroupsPlan)
	diags.Append(d...)
	if diags.HasError() {
		return false
	}

	planConfig.OpenSearch, d = types.ObjectValueFrom(ctx, model.OpenSearchSubConfigAttrTypes, planOpenSearchBlock)
	diags.Append(d...)
	if diags.HasError() {
		return false
	}

	return isHostsChanged
}
