package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/timestamp"
)

var defaultOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}

type OpenSearch struct {
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
	ID                 types.String   `tfsdk:"id"`
	ClusterID          types.String   `tfsdk:"cluster_id"`
	FolderID           types.String   `tfsdk:"folder_id"`
	CreatedAt          types.String   `tfsdk:"created_at"`
	Name               types.String   `tfsdk:"name"`
	Description        types.String   `tfsdk:"description"`
	Labels             types.Map      `tfsdk:"labels"`
	Environment        types.String   `tfsdk:"environment"`
	Config             types.Object   `tfsdk:"config"`
	Hosts              types.List     `tfsdk:"hosts"`
	NetworkID          types.String   `tfsdk:"network_id"`
	Health             types.String   `tfsdk:"health"`
	Status             types.String   `tfsdk:"status"`
	SecurityGroupIDs   types.Set      `tfsdk:"security_group_ids"`
	ServiceAccountID   types.String   `tfsdk:"service_account_id"`
	DeletionProtection types.Bool     `tfsdk:"deletion_protection"`
	MaintenanceWindow  types.Object   `tfsdk:"maintenance_window"`
}

type Config struct {
	Version       types.String `tfsdk:"version"`
	AdminPassword types.String `tfsdk:"admin_password"`
	OpenSearch    types.Object `tfsdk:"opensearch"`
	Dashboards    types.Object `tfsdk:"dashboards"`
	Access        types.Object `tfsdk:"access"`
}

type Access struct {
	DataTransfer types.Bool `tfsdk:"data_transfer"`
	Serverless   types.Bool `tfsdk:"serverless"`
}

type OpenSearchSubConfig struct {
	NodeGroups types.List `tfsdk:"node_groups"`
	Plugins    types.Set  `tfsdk:"plugins"`
}

type DashboardsSubConfig struct {
	NodeGroups types.List `tfsdk:"node_groups"`
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

type DashboardNode struct {
	Name           types.String `tfsdk:"name"`
	Resources      types.Object `tfsdk:"resources"`
	HostsCount     types.Int64  `tfsdk:"hosts_count"`
	ZoneIDs        types.Set    `tfsdk:"zone_ids"`
	SubnetIDs      types.Set    `tfsdk:"subnet_ids"`
	AssignPublicIP types.Bool   `tfsdk:"assign_public_ip"`
}

func (n DashboardNode) GetResources() types.Object {
	return n.Resources
}

type NodeResource struct {
	ResourcePresetID types.String `tfsdk:"resource_preset_id"`
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	DiskTypeID       types.String `tfsdk:"disk_type_id"`
}

type Host struct {
	FQDN           types.String `tfsdk:"fqdn"`
	Type           types.String `tfsdk:"type"`
	Roles          types.Set    `tfsdk:"roles"`
	AssignPublicIP types.Bool   `tfsdk:"assign_public_ip"`
	Zone           types.String `tfsdk:"zone"`
	SubnetID       types.String `tfsdk:"subnet_id"`
	NodeGroup      types.String `tfsdk:"node_group"`
}

type MaintenanceWindow struct {
	Type types.String `tfsdk:"type"`
	Day  types.String `tfsdk:"day"`
	Hour types.Int64  `tfsdk:"hour"`
}

var ConfigAttrTypes = map[string]attr.Type{
	"version":        types.StringType,
	"admin_password": types.StringType,
	"opensearch":     types.ObjectType{AttrTypes: OpenSearchSubConfigAttrTypes},
	"dashboards":     types.ObjectType{AttrTypes: DashboardsSubConfigAttrTypes},
	"access":         types.ObjectType{AttrTypes: accessAttrTypes},
}

var accessAttrTypes = map[string]attr.Type{
	"data_transfer": types.BoolType,
	"serverless":    types.BoolType,
}

var OpenSearchSubConfigAttrTypes = map[string]attr.Type{
	"node_groups": types.ListType{ElemType: OpenSearchNodeType},
	"plugins":     types.SetType{ElemType: types.StringType},
}

var DashboardsSubConfigAttrTypes = map[string]attr.Type{
	"node_groups": types.ListType{ElemType: DashboardNodeType},
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

var DashboardNodeType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":             types.StringType,
		"resources":        types.ObjectType{AttrTypes: NodeResourceAttrTypes},
		"hosts_count":      types.Int64Type,
		"zone_ids":         types.SetType{ElemType: types.StringType},
		"subnet_ids":       types.SetType{ElemType: types.StringType},
		"assign_public_ip": types.BoolType,
	},
}

var NodeResourceAttrTypes = map[string]attr.Type{
	"resource_preset_id": types.StringType,
	"disk_size":          types.Int64Type,
	"disk_type_id":       types.StringType,
}

var HostType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"fqdn":             types.StringType,
		"type":             types.StringType,
		"roles":            types.SetType{ElemType: types.StringType},
		"assign_public_ip": types.BoolType,
		"zone":             types.StringType,
		"subnet_id":        types.StringType,
		"node_group":       types.StringType,
	},
}

var MaintenanceWindowAttrTypes = map[string]attr.Type{
	"type": types.StringType,
	"day":  types.StringType,
	"hour": types.Int64Type,
}

func ClusterToState(ctx context.Context, cluster *opensearch.Cluster, state *OpenSearch) diag.Diagnostics {
	state.FolderID = types.StringValue(cluster.GetFolderId())
	state.CreatedAt = types.StringValue(timestamp.Get(cluster.GetCreatedAt()))
	state.Name = types.StringValue(cluster.GetName())
	if !state.Description.IsNull() || cluster.GetDescription() != "" {
		state.Description = types.StringValue(cluster.GetDescription())
	}

	if !state.Labels.IsNull() || cluster.Labels != nil {
		labels, diags := types.MapValueFrom(ctx, types.StringType, cluster.Labels)
		if diags.HasError() {
			return diags
		}
		state.Labels = labels
	}

	state.Environment = types.StringValue(cluster.GetEnvironment().String())

	var diags diag.Diagnostics
	state.Config, diags = configToState(ctx, cluster.Config, state)
	if diags.HasError() {
		return diags
	}

	state.NetworkID = types.StringValue(cluster.GetNetworkId())
	state.Health = types.StringValue(cluster.GetHealth().String())
	state.Status = types.StringValue(cluster.GetStatus().String())

	state.SecurityGroupIDs, diags = nullableStringSliceToSet(ctx, cluster.SecurityGroupIds)
	if diags.HasError() {
		return diags
	}

	if !state.ServiceAccountID.IsNull() || cluster.ServiceAccountId != "" {
		state.ServiceAccountID = types.StringValue(cluster.ServiceAccountId)
	}

	state.DeletionProtection = types.BoolValue(cluster.GetDeletionProtection())
	state.MaintenanceWindow, diags = maintenanceWindowToObject(ctx, cluster.MaintenanceWindow)
	return diags
}

func configToState(ctx context.Context, cfg *opensearch.ClusterConfig, state *OpenSearch) (types.Object, diag.Diagnostics) {
	stateCfg, diags := ParseConfig(ctx, state)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	adminPassword := types.StringValue("")
	if !(stateCfg == nil || stateCfg.AdminPassword.IsNull() || stateCfg.AdminPassword.IsUnknown()) {
		adminPassword, diags = stateCfg.AdminPassword.ToStringValue(ctx)
		if diags.HasError() {
			return types.ObjectUnknown(ConfigAttrTypes), diags
		}
	}

	//It is required to have a config.opensearch block, so we can skip checking it
	stateOpenSearch, diags := ParseOpenSearchSubConfig(ctx, stateCfg)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	opensearchSubConfig, diags := openSearchSubConfigToObject(ctx, cfg.Opensearch, stateOpenSearch)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	stateDashboards, diags := ParseDashboardSubConfig(ctx, stateCfg)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	dashboardSubConfig, diags := dashboardSubConfigToObject(ctx, cfg.Dashboards, stateDashboards)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	access, diags := accessToObject(ctx, cfg.Access)
	if diags.HasError() {
		return types.ObjectUnknown(ConfigAttrTypes), diags
	}

	return types.ObjectValueFrom(ctx, ConfigAttrTypes, Config{
		Version:       types.StringValue(cfg.GetVersion()),
		AdminPassword: adminPassword,
		OpenSearch:    opensearchSubConfig,
		Dashboards:    dashboardSubConfig,
		Access:        access,
	})
}

func accessToObject(ctx context.Context, cfg *opensearch.Access) (types.Object, diag.Diagnostics) {
	if cfg == nil {
		return types.ObjectNull(accessAttrTypes), nil
	}

	return types.ObjectValueFrom(ctx, accessAttrTypes, Access{
		DataTransfer: types.BoolValue(cfg.GetDataTransfer()),
		Serverless:   types.BoolValue(cfg.GetServerless()),
	})
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

func dashboardsNodeGroupsToList(ctx context.Context, nodeGroups []*opensearch.Dashboards_NodeGroup, state []DashboardNode) (types.List, diag.Diagnostics) {
	groupsByName := GetGroupByName(nodeGroups)
	var ret = make([]DashboardNode, 0, len(nodeGroups))
	nodeGroupNames := getGroupNames(nodeGroups)
	if len(state) != 0 {
		nodeGroupNames = make([]string, 0, len(state))
		for _, s := range state {
			nodeGroupNames = append(nodeGroupNames, s.Name.ValueString())
		}
	}

	for _, groupName := range nodeGroupNames {
		v := groupsByName[groupName]
		zoneIds, diags := nullableStringSliceToSet(ctx, v.GetZoneIds())
		if diags.HasError() {
			diags.AddError("Failed to parse dashboards.node_groups.zone_ids", fmt.Sprintf("Error while parsing zone_ids for group: %s", groupName))
			return types.ListUnknown(DashboardNodeType), diags
		}

		subnetIds, diags := nullableStringSliceToSet(ctx, v.GetSubnetIds())
		if diags.HasError() {
			diags.AddError("Failed to parse dashboards.node_groups.subnet_ids", fmt.Sprintf("Error while parsing subnet_ids for group: %s", groupName))
			return types.ListUnknown(DashboardNodeType), diags
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

func resourcesToObject(ctx context.Context, r *opensearch.Resources) (types.Object, diag.Diagnostics) {
	if isEmptyResources(r) {
		return types.ObjectNull(NodeResourceAttrTypes), diag.Diagnostics{}
	}

	return types.ObjectValueFrom(ctx, NodeResourceAttrTypes, NodeResource{
		ResourcePresetID: types.StringValue(r.GetResourcePresetId()),
		DiskSize:         types.Int64Value(r.GetDiskSize()),
		DiskTypeID:       types.StringValue(r.GetDiskTypeId()),
	})
}

func isEmptyResources(r *opensearch.Resources) bool {
	return r == nil ||
		(r.DiskSize == 0 && r.DiskTypeId == "" && r.ResourcePresetId == "")
}

func rolesToSet(roles []opensearch.OpenSearch_GroupRole) (types.Set, diag.Diagnostics) {
	if roles == nil {
		return types.SetNull(types.StringType), diag.Diagnostics{}
	}

	res := make([]attr.Value, 0, len(roles))
	for _, v := range roles {
		res = append(res, types.StringValue(v.String()))
	}

	return types.SetValue(types.StringType, res)
}

func maintenanceWindowToObject(ctx context.Context, mw *opensearch.MaintenanceWindow) (types.Object, diag.Diagnostics) {
	var res basetypes.ObjectValue
	var diags diag.Diagnostics
	if val := mw.GetAnytime(); val != nil {
		res, diags = types.ObjectValueFrom(ctx, MaintenanceWindowAttrTypes, MaintenanceWindow{
			Type: types.StringValue("ANYTIME"),
		})
	}

	if val := mw.GetWeeklyMaintenanceWindow(); val != nil {
		res, diags = types.ObjectValueFrom(ctx, MaintenanceWindowAttrTypes, MaintenanceWindow{
			Type: types.StringValue("WEEKLY"),
			Day:  types.StringValue(val.GetDay().String()),
			Hour: types.Int64Value(val.GetHour()),
		})
	}

	if diags.HasError() {
		return types.ObjectUnknown(MaintenanceWindowAttrTypes), diags
	}

	return res, diags
}

func nullableStringSliceToSet(ctx context.Context, s []string) (types.Set, diag.Diagnostics) {
	if s == nil {
		return types.SetNull(types.StringType), diag.Diagnostics{}
	}

	return types.SetValueFrom(ctx, types.StringType, s)
}

func HostsToState(ctx context.Context, hosts []*opensearch.Host) (types.List, diag.Diagnostics) {
	res := make([]Host, 0, len(hosts))

	for _, h := range hosts {
		roles, diags := rolesToSet(h.GetRoles())
		if diags.HasError() {
			diags.AddError("Failed to parse hosts.roles", fmt.Sprintf("Error while parsing roles for host: %s", h.GetName()))
			return types.ListUnknown(HostType), diags
		}

		res = append(res, Host{
			FQDN:           types.StringValue(h.GetName()),
			Type:           types.StringValue(h.GetType().String()),
			Roles:          roles,
			AssignPublicIP: types.BoolValue(h.GetAssignPublicIp()),
			Zone:           types.StringValue(h.GetZoneId()),
			SubnetID:       types.StringValue(h.GetSubnetId()),
			NodeGroup:      types.StringValue(h.GetNodeGroup()),
		})
	}

	return types.ListValueFrom(ctx, HostType, res)
}

func ParseConfig(ctx context.Context, state *OpenSearch) (*Config, diag.Diagnostics) {
	planConfig := &Config{}
	diags := state.Config.As(ctx, &planConfig, defaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return planConfig, diag.Diagnostics{}
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

func ParseDashboardSubConfig(ctx context.Context, state *Config) (*DashboardsSubConfig, diag.Diagnostics) {
	if state == nil {
		return nil, diag.Diagnostics{}
	}

	res := &DashboardsSubConfig{}
	diags := state.Dashboards.As(ctx, &res, defaultOpts) //NOTE: &res because state.Dashboards can be nil
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}

type WithResources interface {
	GetResources() types.Object
}

func ParseNodeResource(ctx context.Context, ng WithResources) (*NodeResource, diag.Diagnostics) {
	res := &NodeResource{}
	diags := ng.GetResources().As(ctx, res, defaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}

func ParseAccess(ctx context.Context, state *Config) (*Access, diag.Diagnostics) {
	res := &Access{}
	diags := state.Access.As(ctx, res, defaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}

func ParseMaintenanceWindow(ctx context.Context, model *OpenSearch) (*MaintenanceWindow, diag.Diagnostics) {
	res := &MaintenanceWindow{}
	diags := model.MaintenanceWindow.As(ctx, res, defaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}

func ParseGenerics[T any, V any](ctx context.Context, plan, state T, parse func(context.Context, T) (V, diag.Diagnostics)) (V, V, diag.Diagnostics) {
	planConfig, diags := parse(ctx, plan)
	if diags.HasError() {
		//NOTE: can't create an empty value result, so just dublicate planConfig
		return planConfig, planConfig, diags
	}

	stateConfig, diags := parse(ctx, state)
	if diags.HasError() {
		return planConfig, stateConfig, diags
	}

	return planConfig, stateConfig, diag.Diagnostics{}
}

type withName interface {
	GetName() string
}

func GetGroupByName[T withName](groups []T) map[string]T {
	groupsByName := make(map[string]T, len(groups))
	for _, g := range groups {
		groupsByName[g.GetName()] = g
	}

	return groupsByName
}

func getGroupNames[T withName](groups []T) []string {
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.GetName()
	}

	return names
}
