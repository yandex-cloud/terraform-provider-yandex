package model

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type DashboardsSubConfig struct {
	NodeGroups types.List `tfsdk:"node_groups"`
}

type DashboardNode struct {
	Name           types.String `tfsdk:"name"`
	Resources      types.Object `tfsdk:"resources"`
	HostsCount     types.Int64  `tfsdk:"hosts_count"`
	ZoneIDs        types.Set    `tfsdk:"zone_ids"`
	SubnetIDs      types.List   `tfsdk:"subnet_ids"`
	AssignPublicIP types.Bool   `tfsdk:"assign_public_ip"`
}

func (n DashboardNode) GetResources() types.Object {
	return n.Resources
}

var DashboardsSubConfigAttrTypes = map[string]attr.Type{
	"node_groups": types.ListType{ElemType: DashboardNodeType},
}

var DashboardNodeType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":             types.StringType,
		"resources":        types.ObjectType{AttrTypes: NodeResourceAttrTypes},
		"hosts_count":      types.Int64Type,
		"zone_ids":         types.SetType{ElemType: types.StringType},
		"subnet_ids":       types.ListType{ElemType: types.StringType},
		"assign_public_ip": types.BoolType,
	},
}

func dashboardSubConfigToObject(ctx context.Context, cfg *opensearch.Dashboards, state *DashboardsSubConfig) (types.Object, diag.Diagnostics) {
	if cfg == nil || len(cfg.NodeGroups) == 0 {
		return types.ObjectNull(DashboardsSubConfigAttrTypes), diag.Diagnostics{}
	}

	var stateNodeGroups []DashboardNode

	if state != nil {
		// we have to keep node_groups order from config
		stateNodeGroups = make([]DashboardNode, 0, len(state.NodeGroups.Elements()))
		diags := state.NodeGroups.ElementsAs(ctx, &stateNodeGroups, false)
		if diags.HasError() {
			return types.ObjectNull(DashboardsSubConfigAttrTypes), diags
		}
	}

	nodeGroups, diags := dashboardsNodeGroupsToList(ctx, cfg.NodeGroups, stateNodeGroups)
	if diags.HasError() {
		return types.ObjectNull(DashboardsSubConfigAttrTypes), diags
	}

	return types.ObjectValueFrom(ctx, DashboardsSubConfigAttrTypes, DashboardsSubConfig{
		NodeGroups: nodeGroups,
	})
}

func dashboardsNodeGroupsToList(ctx context.Context, nodeGroups []*opensearch.Dashboards_NodeGroup, state []DashboardNode) (types.List, diag.Diagnostics) {
	groupsByName := GetGroupByName(nodeGroups)
	var ret = make([]DashboardNode, 0, len(nodeGroups))
	nodeGroupNames := getGroupNames(nodeGroups)
	stateGroupsByName := make(map[string]DashboardNode, len(state))

	if len(state) != 0 && len(state) == len(nodeGroups) {
		for i, s := range state {
			stateGroupsByName[s.Name.ValueString()] = s
			nodeGroupNames[i] = s.Name.ValueString()
		}
	}

	for _, groupName := range nodeGroupNames {
		v := groupsByName[groupName]
		stateGroup := stateGroupsByName[groupName]

		zoneIds, diags := nullableStringSliceToSet(ctx, v.GetZoneIds())
		if diags.HasError() {
			diags.AddError("Failed to parse dashboards.node_groups.zone_ids", fmt.Sprintf("Error while parsing zone_ids for group: %s", groupName))
			return types.ListUnknown(DashboardNodeType), diags
		}

		if setsAreEqual(stateGroup.ZoneIDs, zoneIds) {
			zoneIds = stateGroup.ZoneIDs
		}

		subnetIds, diags := nullableStringSliceToList(ctx, v.GetSubnetIds())
		if diags.HasError() {
			diags.AddError("Failed to parse dashboards.node_groups.subnet_ids", fmt.Sprintf("Error while parsing subnet_ids for group: %s", groupName))
			return types.ListUnknown(DashboardNodeType), diags
		}

		if sliceAndListAreEqual(ctx, stateGroup.SubnetIDs, v.GetSubnetIds()) {
			subnetIds = stateGroup.SubnetIDs
		}

		resources, diags := resourcesToObject(ctx, v.GetResources())
		if diags.HasError() {
			diags.AddError("Failed to parse dashboards.node_groups.resources", fmt.Sprintf("Error while parsing resources for group: %s", groupName))
			return types.ListUnknown(DashboardNodeType), diags
		}

		ret = append(ret, DashboardNode{
			Name:           types.StringValue(v.GetName()),
			Resources:      resources,
			HostsCount:     types.Int64Value(v.GetHostsCount()),
			ZoneIDs:        zoneIds,
			SubnetIDs:      subnetIds,
			AssignPublicIP: types.BoolValue(v.GetAssignPublicIp()),
		})
	}

	return types.ListValueFrom(ctx, DashboardNodeType, ret)
}

func ParseDashboardSubConfig(ctx context.Context, state *Config) (*DashboardsSubConfig, diag.Diagnostics) {
	if state == nil {
		return nil, diag.Diagnostics{}
	}

	res := &DashboardsSubConfig{}
	diags := state.Dashboards.As(ctx, &res, datasize.DefaultOpts) //NOTE: &res because state.Dashboards can be nil
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}
