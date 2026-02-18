package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type OpenSearchSubConfig struct {
	NodeGroups types.List   `tfsdk:"node_groups"`
	Plugins    types.Set    `tfsdk:"plugins"`
	Config     types.Object `tfsdk:"config"`
}

type OpenSearchNode struct {
	Name                types.String `tfsdk:"name"`
	Resources           types.Object `tfsdk:"resources"`
	HostsCount          types.Int64  `tfsdk:"hosts_count"`
	ZoneIDs             types.Set    `tfsdk:"zone_ids"`
	SubnetIDs           types.List   `tfsdk:"subnet_ids"`
	AssignPublicIP      types.Bool   `tfsdk:"assign_public_ip"`
	Roles               types.Set    `tfsdk:"roles"`
	DiskSizeAutoscaling types.Object `tfsdk:"disk_size_autoscaling"`
}

func (o OpenSearchNode) Equal(other OpenSearchNode) bool {
	if !o.Name.Equal(other.Name) {
		return false
	}

	if !o.Resources.Equal(other.Resources) {
		return false
	}
	if !o.HostsCount.Equal(other.HostsCount) {
		return false
	}
	if !o.ZoneIDs.Equal(other.ZoneIDs) {
		return false
	}
	if !o.SubnetIDs.Equal(other.SubnetIDs) {
		return false
	}
	if !o.AssignPublicIP.Equal(other.AssignPublicIP) {
		return false
	}
	if !o.Roles.Equal(other.Roles) {
		return false
	}

	if (o.DiskSizeAutoscaling.IsUnknown() || o.DiskSizeAutoscaling.IsNull()) && (!other.DiskSizeAutoscaling.IsUnknown() && !other.DiskSizeAutoscaling.IsNull()) {
		return false
	}

	if (other.DiskSizeAutoscaling.IsUnknown() || other.DiskSizeAutoscaling.IsNull()) && (!o.DiskSizeAutoscaling.IsUnknown() && !o.DiskSizeAutoscaling.IsNull()) {
		return false
	}

	if !(o.DiskSizeAutoscaling.IsUnknown() || o.DiskSizeAutoscaling.IsNull()) && !(other.DiskSizeAutoscaling.IsUnknown() || other.DiskSizeAutoscaling.IsNull()) && !o.DiskSizeAutoscaling.Equal(other.DiskSizeAutoscaling) {
		return false
	}

	return true
}

func (n OpenSearchNode) GetName() string {
	return n.Name.ValueString()
}

func (n OpenSearchNode) GetResources() types.Object {
	return n.Resources
}

func (n OpenSearchNode) GetDiskSizeAutoscaling() types.Object {
	return n.DiskSizeAutoscaling
}

var OpenSearchSubConfigAttrTypes = map[string]attr.Type{
	"node_groups": types.ListType{ElemType: OpenSearchNodeType},
	"plugins":     types.SetType{ElemType: types.StringType},
	"config":      types.ObjectType{AttrTypes: openSearchConfig2Types},
}

var OpenSearchNodeType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":                  types.StringType,
		"resources":             types.ObjectType{AttrTypes: NodeResourceAttrTypes},
		"hosts_count":           types.Int64Type,
		"zone_ids":              types.SetType{ElemType: types.StringType},
		"subnet_ids":            types.ListType{ElemType: types.StringType},
		"assign_public_ip":      types.BoolType,
		"roles":                 types.SetType{ElemType: types.StringType},
		"disk_size_autoscaling": types.ObjectType{AttrTypes: DiskSizeAutoscalingAttrTypes},
	},
}

func openSearchSubConfigToObject(ctx context.Context, cfg *opensearch.OpenSearch, state OpenSearchSubConfig) (types.Object, diag.Diagnostics) {
	plugins, diags := nullableStringSliceToSet(ctx, cfg.GetPlugins())
	if diags.HasError() {
		return types.ObjectNull(OpenSearchSubConfigAttrTypes), diags
	}

	if setsAreEqual(state.Plugins, plugins) {
		//This condition needs to fix import in tests, because somehow acceptance tests will contain `types.SetType[!!! MISSING TYPE !!!] / underlying type: tftypes.Set[tftypes.DynamicPseudoType]` instead of `types.SetType[basetypes.StringType] / underlying type: tftypes.Set[tftypes.String]`
		switch state.Plugins.ElementType(ctx).(type) {
		case nil:
			tflog.Debug(ctx, "got nil element type for 'plugins', set received value")
		default:
			plugins = state.Plugins
		}
	}

	config := types.ObjectNull(openSearchConfig2Types)

	ocs2 := cfg.GetOpensearchConfigSet_2()
	if ocs2 != nil {
		config, diags = openSearchConfig2ToObject(ctx, ocs2.GetUserConfig())
		if diags.HasError() {
			return types.ObjectNull(OpenSearchSubConfigAttrTypes), diags
		}
	}

	// we have to keep node_groups order from config
	stateNodeGroups := make([]OpenSearchNode, 0, len(state.NodeGroups.Elements()))
	if len(state.NodeGroups.Elements()) != 0 {
		diags = state.NodeGroups.ElementsAs(ctx, &stateNodeGroups, false)
		if diags.HasError() {
			return types.ObjectNull(OpenSearchSubConfigAttrTypes), diags
		}
	}

	nodeGroups, diags := openSearchNodeGroupsToList(ctx, cfg.NodeGroups, stateNodeGroups)
	if diags.HasError() {
		return types.ObjectNull(OpenSearchSubConfigAttrTypes), diags
	}

	return types.ObjectValueFrom(ctx, OpenSearchSubConfigAttrTypes, OpenSearchSubConfig{
		NodeGroups: nodeGroups,
		Plugins:    plugins,
		Config:     config,
	})
}

func openSearchNodeGroupsToList(ctx context.Context, nodeGroups []*opensearch.OpenSearch_NodeGroup, state []OpenSearchNode) (types.List, diag.Diagnostics) {
	groupsByName := GetGroupByName(nodeGroups)
	var ret = make([]OpenSearchNode, 0, len(nodeGroups))
	nodeGroupNames := getGroupNames(nodeGroups)
	stateGroupsByName := make(map[string]OpenSearchNode, len(state))

	if len(state) != 0 && len(state) == len(nodeGroups) {
		for i, s := range state {
			stateGroupsByName[s.Name.ValueString()] = s
			nodeGroupNames[i] = s.Name.ValueString()
		}
	}

	for _, groupName := range nodeGroupNames {
		v := groupsByName[groupName]
		stateGroup := stateGroupsByName[groupName]

		roles, diags := rolesToSet(v.Roles)
		if diags.HasError() {
			diags.AddError("Failed to parse opensearch.node_groups.roles", fmt.Sprintf("Error while parsing roles for group: %s", groupName))
			return types.ListUnknown(OpenSearchNodeType), diags
		}

		if sameRoles(ctx, stateGroup.Roles, v.GetRoles()) {
			roles = stateGroup.Roles
		}

		zoneIds, diags := nullableStringSliceToSet(ctx, v.GetZoneIds())
		if diags.HasError() {
			diags.AddError("Failed to parse opensearch.node_groups.zone_ids", fmt.Sprintf("Error while parsing zone_ids for group: %s", groupName))
			return types.ListUnknown(OpenSearchNodeType), diags
		}

		if setsAreEqual(stateGroup.ZoneIDs, zoneIds) {
			zoneIds = stateGroup.ZoneIDs
		}

		subnetIds, diags := nullableStringSliceToList(ctx, v.GetSubnetIds())
		if diags.HasError() {
			diags.AddError("Failed to parse opensearch.node_groups.subnet_ids", fmt.Sprintf("Error while parsing subnet_ids for group: %s", groupName))
			return types.ListUnknown(OpenSearchNodeType), diags
		}

		if sliceAndListAreEqual(ctx, stateGroup.SubnetIDs, v.GetSubnetIds()) {
			subnetIds = stateGroup.SubnetIDs
		} else {
			tflog.Debug(ctx, fmt.Sprintf("slice %v is not 'same' with %v", v.GetSubnetIds(), stateGroup.SubnetIDs))
		}

		resources, diags := resourcesToObject(ctx, v.GetResources())
		if diags.HasError() {
			diags.AddError("Failed to parse opensearch.node_groups.resources", fmt.Sprintf("Error while parsing resources for group: %s", groupName))
			return types.ListUnknown(OpenSearchNodeType), diags
		}

		diskSizeAutoscaling, diags := diskSizeAutoscalingToObject(ctx, v.GetDiskSizeAutoscaling())
		if diags.HasError() {
			diags.AddError("Failed to parse opensearch.node_groups.disk_size_autoscaling", fmt.Sprintf("Error while parsing disk_size_autoscaling for group: %s", groupName))
			return types.ListUnknown(OpenSearchNodeType), diags
		}

		ret = append(ret, OpenSearchNode{
			Name:                types.StringValue(v.GetName()),
			Resources:           resources,
			HostsCount:          types.Int64Value(v.GetHostsCount()),
			ZoneIDs:             zoneIds,
			SubnetIDs:           subnetIds,
			AssignPublicIP:      types.BoolValue(v.GetAssignPublicIp()),
			Roles:               roles,
			DiskSizeAutoscaling: diskSizeAutoscaling,
		})
	}

	return types.ListValueFrom(ctx, OpenSearchNodeType, ret)
}

func sameRoles(ctx context.Context, state types.Set, res []opensearch.OpenSearch_GroupRole) bool {
	if state.IsUnknown() {
		return false
	}

	// ensure they are at the same length
	if len(state.Elements()) != len(res) {
		return false
	}

	if len(state.Elements()) == 0 {
		//both has 0 length so they equal
		return true
	}

	stringRoles := make([]string, 0, len(state.Elements()))
	_ = state.ElementsAs(ctx, &stringRoles, false)

	stateRoles := make(map[string]interface{}, len(stringRoles))
	for _, r := range stringRoles {
		stateRoles[strings.ToUpper(r)] = nil
	}

	for _, r := range res {
		if _, ok := stateRoles[strings.ToUpper(r.String())]; !ok {
			return false
		}
	}

	return true
}

func ParseOpenSearchSubConfig(ctx context.Context, state *Config) (OpenSearchSubConfig, diag.Diagnostics) {
	var res OpenSearchSubConfig
	if state == nil {
		return res, diag.Diagnostics{}
	}

	diags := state.OpenSearch.As(ctx, &res, datasize.DefaultOpts)
	if diags.HasError() {
		return res, diags
	}

	return res, diag.Diagnostics{}
}
