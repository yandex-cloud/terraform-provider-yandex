package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
)

type OpenSearchSubConfig struct {
	NodeGroups types.List `tfsdk:"node_groups"`
	Plugins    types.Set  `tfsdk:"plugins"`
}

type OpenSearchNode struct {
	Name           types.String `tfsdk:"name"`
	Resources      types.Object `tfsdk:"resources"`
	HostsCount     types.Int64  `tfsdk:"hosts_count"`
	ZoneIDs        types.Set    `tfsdk:"zone_ids"`
	SubnetIDs      types.Set    `tfsdk:"subnet_ids"`
	AssignPublicIP types.Bool   `tfsdk:"assign_public_ip"`
	Roles          types.Set    `tfsdk:"roles"`
}

func (n OpenSearchNode) GetResources() types.Object {
	return n.Resources
}

var OpenSearchSubConfigAttrTypes = map[string]attr.Type{
	"node_groups": types.ListType{ElemType: OpenSearchNodeType},
	"plugins":     types.SetType{ElemType: types.StringType},
}

var OpenSearchNodeType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":             types.StringType,
		"resources":        types.ObjectType{AttrTypes: NodeResourceAttrTypes},
		"hosts_count":      types.Int64Type,
		"zone_ids":         types.SetType{ElemType: types.StringType},
		"subnet_ids":       types.SetType{ElemType: types.StringType},
		"assign_public_ip": types.BoolType,
		"roles":            types.SetType{ElemType: types.StringType},
	},
}

func openSearchSubConfigToObject(ctx context.Context, cfg *opensearch.OpenSearch, state *OpenSearchSubConfig) (types.Object, diag.Diagnostics) {
	plugins := types.SetNull(types.StringType)
	if state != nil && !state.Plugins.IsUnknown() {
		plugins = state.Plugins
	}

	if cfg.GetPlugins() != nil {
		p, diags := types.SetValueFrom(ctx, types.StringType, cfg.GetPlugins())
		if diags.HasError() {
			return types.ObjectNull(OpenSearchSubConfigAttrTypes), diags
		}

		plugins = p
	}

	var stateNodeGroups []OpenSearchNode

	if state != nil {
		// we have to keep node_groups order from config
		stateNodeGroups = make([]OpenSearchNode, 0, len(state.NodeGroups.Elements()))
		diags := state.NodeGroups.ElementsAs(ctx, &stateNodeGroups, false)
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
	})
}

func openSearchNodeGroupsToList(ctx context.Context, nodeGroups []*opensearch.OpenSearch_NodeGroup, state []OpenSearchNode) (types.List, diag.Diagnostics) {
	groupsByName := GetGroupByName(nodeGroups)
	var ret = make([]OpenSearchNode, 0, len(nodeGroups))
	nodeGroupNames := getGroupNames(nodeGroups)
	stateGroupsByName := make(map[string]OpenSearchNode, len(state))

	if len(state) != 0 && len(state) == len(nodeGroups) {
		stateNodeGroupNames := make([]string, 0, len(state))
		for _, s := range state {
			if ng, ok := groupsByName[s.Name.ValueString()]; !ok || !sameNodeGroup(ctx, ng, s) {
				continue
			}
			stateGroupsByName[s.Name.ValueString()] = s
			stateNodeGroupNames = append(stateNodeGroupNames, s.Name.ValueString())
		}

		//if in state and response have same names, we can use state names
		if len(stateNodeGroupNames) == len(nodeGroupNames) {
			nodeGroupNames = stateNodeGroupNames
		}
	}

	for _, groupName := range nodeGroupNames {
		v := groupsByName[groupName]

		var roles basetypes.SetValue
		var diags diag.Diagnostics
		if _, ok := stateGroupsByName[groupName]; !ok {
			roles, diags = rolesToSet(v.Roles)
			if diags.HasError() {
				diags.AddError("Failed to parse opensearch.node_groups.roles", fmt.Sprintf("Error while parsing roles for group: %s", groupName))
				return types.ListUnknown(OpenSearchNodeType), diags
			}
		} else {
			//to prevent change ordering and lower/upper case, to avoid terrafrom inconsistent error: "Provider produced inconsistent result after apply"
			roles = stateGroupsByName[groupName].Roles
		}

		zoneIds, diags := nullableStringSliceToSet(ctx, v.GetZoneIds())
		if diags.HasError() {
			diags.AddError("Failed to parse opensearch.node_groups.zone_ids", fmt.Sprintf("Error while parsing zone_ids for group: %s", groupName))
			return types.ListUnknown(OpenSearchNodeType), diags
		}

		subnetIds, diags := nullableStringSliceToSet(ctx, v.GetSubnetIds())
		if diags.HasError() {
			diags.AddError("Failed to parse opensearch.node_groups.subnet_ids", fmt.Sprintf("Error while parsing subnet_ids for group: %s", groupName))

			return types.ListUnknown(OpenSearchNodeType), diags
		}

		resources, diags := resourcesToObject(ctx, v.GetResources())
		if diags.HasError() {
			diags.AddError("Failed to parse opensearch.node_groups.resources", fmt.Sprintf("Error while parsing resources for group: %s", groupName))
			return types.ListUnknown(OpenSearchNodeType), diags
		}

		ret = append(ret, OpenSearchNode{
			Name:           types.StringValue(v.GetName()),
			Resources:      resources,
			HostsCount:     types.Int64Value(v.GetHostsCount()),
			ZoneIDs:        zoneIds,
			SubnetIDs:      subnetIds,
			AssignPublicIP: types.BoolValue(v.GetAssignPublicIp()),
			Roles:          roles,
		})
	}

	return types.ListValueFrom(ctx, OpenSearchNodeType, ret)
}

func sameNodeGroup(ctx context.Context, res *opensearch.OpenSearch_NodeGroup, state OpenSearchNode) bool {
	if res == nil {
		return false
	}

	stringRoles := make([]string, 0, len(state.Roles.Elements()))
	_ = state.Roles.ElementsAs(ctx, &stringRoles, false)

	stateRoles := make(map[string]interface{}, len(stringRoles))
	for _, r := range stringRoles {
		stateRoles[strings.ToUpper(r)] = nil
	}

	if len(stringRoles) != len(res.Roles) {
		return false
	}

	for _, r := range res.Roles {
		if _, ok := stateRoles[strings.ToUpper(r.String())]; !ok {
			return false
		}
	}

	zoneIds, _ := nullableStringSliceToSet(ctx, res.GetZoneIds())
	subnetIds, _ := nullableStringSliceToSet(ctx, res.GetSubnetIds())
	resources, _ := resourcesToObject(ctx, res.GetResources())

	return res.GetName() == state.Name.ValueString() &&
		res.GetHostsCount() == state.HostsCount.ValueInt64() &&
		res.GetAssignPublicIp() == state.AssignPublicIP.ValueBool() &&
		zoneIds.Equal(state.ZoneIDs) &&
		subnetIds.Equal(state.SubnetIDs) &&
		resources.Equal(state.Resources)
}

func ParseOpenSearchSubConfig(ctx context.Context, state *Config) (*OpenSearchSubConfig, diag.Diagnostics) {
	if state == nil {
		return nil, diag.Diagnostics{}
	}

	res := &OpenSearchSubConfig{}
	diags := state.OpenSearch.As(ctx, res, defaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}
