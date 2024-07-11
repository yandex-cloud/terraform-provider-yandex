package request

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/model"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/validate"
	"google.golang.org/genproto/protobuf/field_mask"
)

func PrepareCreateOpenSearchRequest(ctx context.Context, plan *model.OpenSearch, providerConfig *config.State) (*opensearch.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	folderID, d := validate.FolderID(plan.FolderID, providerConfig)
	diags.Append(d)
	if diags.HasError() {
		return nil, diags
	}

	labels := make(map[string]string, len(plan.Labels.Elements()))
	diags.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	env, d := parseOpenSearchEnv(plan.Environment)
	diags.Append(d)
	if diags.HasError() {
		return nil, diags
	}

	config, diags := parseOpenSearchConfigCreateSpec(ctx, plan)
	if diags.HasError() {
		return nil, diags
	}

	networkID, d := validate.NetworkId(plan.NetworkID, providerConfig)
	diags.Append(d)
	if diags.HasError() {
		return nil, diags
	}

	securityGroupIds := make([]string, 0, len(plan.SecurityGroupIDs.Elements()))
	diags.Append(plan.SecurityGroupIDs.ElementsAs(ctx, &securityGroupIds, false)...)
	if diags.HasError() {
		return nil, diags
	}

	mw, diags := parseOpenSearchMaintenanceWindow(ctx, plan)
	if diags.HasError() {
		return nil, diags
	}

	req := &opensearch.CreateClusterRequest{
		FolderId:           folderID,
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Labels:             labels,
		Environment:        env,
		ConfigSpec:         config,
		NetworkId:          networkID,
		SecurityGroupIds:   securityGroupIds,
		ServiceAccountId:   plan.ServiceAccountID.ValueString(),
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		MaintenanceWindow:  mw,
	}

	return req, diag.Diagnostics{}
}

func parseOpenSearchConfigCreateSpec(ctx context.Context, c *model.OpenSearch) (*opensearch.ConfigCreateSpec, diag.Diagnostics) {
	config, diags := model.ParseConfig(ctx, c)
	if diags.HasError() {
		return nil, diags
	}

	access, diags := tryParseAccess(ctx, config)
	if diags.HasError() {
		return nil, diags
	}

	if config.OpenSearch.IsNull() || config.OpenSearch.IsUnknown() {
		diags.AddError("config.opensearch is required", "")
		return nil, diags
	}

	openSearchBlock, diags := model.ParseOpenSearchSubConfig(ctx, config)
	if diags.HasError() {
		return nil, diags
	}

	var plugins []string
	if !(openSearchBlock.Plugins.IsUnknown() || openSearchBlock.Plugins.IsNull()) {
		plugins = make([]string, 0, len(openSearchBlock.Plugins.Elements()))
		diags.Append(openSearchBlock.Plugins.ElementsAs(ctx, &plugins, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	nodeGroups, diags := ParseOpenSearchCreateSpecNodeGroups(ctx, openSearchBlock)
	if diags.HasError() {
		return nil, diags
	}

	opensearchSpec := &opensearch.OpenSearchCreateSpec{
		NodeGroups: nodeGroups,
		Plugins:    plugins,
	}

	if config.Dashboards.IsNull() || config.Dashboards.IsUnknown() {
		return &opensearch.ConfigCreateSpec{
			Access:         access,
			AdminPassword:  config.AdminPassword.ValueString(),
			Version:        config.Version.ValueString(),
			OpensearchSpec: opensearchSpec,
		}, diags
	}

	dashboardsBlock, diags := model.ParseDashboardSubConfig(ctx, config)
	if diags.HasError() {
		return nil, diags
	}

	dashboardsNodeGroups, diags := ParseDashboardsCreateSpecNodeGroups(ctx, dashboardsBlock)
	if diags.HasError() {
		return nil, diags
	}

	dashboardsSpec := &opensearch.DashboardsCreateSpec{
		NodeGroups: dashboardsNodeGroups,
	}

	return &opensearch.ConfigCreateSpec{
		Access:         access,
		AdminPassword:  config.AdminPassword.ValueString(),
		Version:        config.Version.ValueString(),
		OpensearchSpec: opensearchSpec,
		DashboardsSpec: dashboardsSpec,
	}, diags
}

func ParseOpenSearchCreateSpecNodeGroups(ctx context.Context, cfg *model.OpenSearchSubConfig) ([]*opensearch.OpenSearchCreateSpec_NodeGroup, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	nodeGroups := make([]model.OpenSearchNode, 0, len(cfg.NodeGroups.Elements()))
	diags.Append(cfg.NodeGroups.ElementsAs(ctx, &nodeGroups, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]*opensearch.OpenSearchCreateSpec_NodeGroup, 0, len(nodeGroups))
	for _, ng := range nodeGroups {

		resources, d := parseResources(ctx, ng)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		zoneIDs := make([]string, 0, len(ng.ZoneIDs.Elements()))
		diags.Append(ng.ZoneIDs.ElementsAs(ctx, &zoneIDs, false)...)
		if diags.HasError() {
			return nil, diags
		}

		subnetIDs := make([]string, 0, len(ng.SubnetIDs.Elements()))
		diags.Append(ng.SubnetIDs.ElementsAs(ctx, &subnetIDs, false)...)
		if diags.HasError() {
			return nil, diags
		}

		roles := make([]opensearch.OpenSearch_GroupRole, 0, len(ng.Roles.Elements()))

		stringRoles := make([]string, 0, len(ng.Roles.Elements()))
		diags.Append(ng.Roles.ElementsAs(ctx, &stringRoles, false)...)
		if diags.HasError() {
			return nil, diags
		}

		for _, role := range stringRoles {
			roleId := opensearch.OpenSearch_GroupRole_value[strings.ToUpper(role)]
			roles = append(roles, opensearch.OpenSearch_GroupRole(roleId))
		}

		result = append(result, &opensearch.OpenSearchCreateSpec_NodeGroup{
			Name:           ng.Name.ValueString(),
			Resources:      resources,
			HostsCount:     ng.HostsCount.ValueInt64(),
			ZoneIds:        zoneIDs,
			SubnetIds:      subnetIDs,
			AssignPublicIp: ng.AssignPublicIP.ValueBool(),
			Roles:          roles,
		})
	}

	return result, diags
}

func ParseDashboardsCreateSpecNodeGroups(ctx context.Context, cfg *model.DashboardsSubConfig) ([]*opensearch.DashboardsCreateSpec_NodeGroup, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	nodeGroups := make([]model.DashboardNode, 0, len(cfg.NodeGroups.Elements()))
	diags.Append(cfg.NodeGroups.ElementsAs(ctx, &nodeGroups, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]*opensearch.DashboardsCreateSpec_NodeGroup, 0, len(nodeGroups))
	for _, ng := range nodeGroups {

		resources, diags := parseResources(ctx, ng)
		if diags.HasError() {
			return nil, diags
		}

		zoneIDs := make([]string, 0, len(ng.ZoneIDs.Elements()))
		diags.Append(ng.ZoneIDs.ElementsAs(ctx, &zoneIDs, false)...)
		if diags.HasError() {
			return nil, diags
		}

		subnetIDs := make([]string, 0, len(ng.SubnetIDs.Elements()))
		diags.Append(ng.SubnetIDs.ElementsAs(ctx, &subnetIDs, false)...)
		if diags.HasError() {
			return nil, diags
		}

		result = append(result, &opensearch.DashboardsCreateSpec_NodeGroup{
			Name:           ng.Name.ValueString(),
			Resources:      resources,
			HostsCount:     ng.HostsCount.ValueInt64(),
			ZoneIds:        zoneIDs,
			SubnetIds:      subnetIDs,
			AssignPublicIp: ng.AssignPublicIP.ValueBool(),
		})
	}

	return result, diags
}

func parseOpenSearchEnv(e basetypes.StringValue) (opensearch.Cluster_Environment, diag.Diagnostic) {
	v, ok := opensearch.Cluster_Environment_value[e.ValueString()]
	if !ok || v == 0 {
		allowedEnvs := make([]string, 0, len(opensearch.Cluster_Environment_value))
		for k, v := range opensearch.Cluster_Environment_value {
			if v == 0 {
				continue
			}
			allowedEnvs = append(allowedEnvs, k)
		}

		return 0, diag.NewErrorDiagnostic(
			"Failed to parse OpenSearch environment",
			fmt.Sprintf("Error while parsing value for 'environment'. Value must be one of `%s`, not `%s`", strings.Join(allowedEnvs, "`, `"), e),
		)
	}
	return opensearch.Cluster_Environment(v), nil
}

func parseResources(ctx context.Context, ng model.WithResources) (*opensearch.Resources, diag.Diagnostics) {
	r := ng.GetResources()
	if r.IsUnknown() || r.IsNull() {
		return &opensearch.Resources{}, diag.Diagnostics{}
	}

	resource, diags := model.ParseNodeResource(ctx, ng)
	if diags.HasError() {
		return nil, diags
	}

	return &opensearch.Resources{
		ResourcePresetId: resource.ResourcePresetID.ValueString(),
		DiskSize:         resource.DiskSize.ValueInt64(),
		DiskTypeId:       resource.DiskTypeID.ValueString(),
	}, diag.Diagnostics{}
}

func tryParseAccess(ctx context.Context, cfg *model.Config) (*opensearch.Access, diag.Diagnostics) {
	if cfg.Access.IsUnknown() || cfg.Access.IsNull() {
		return nil, diag.Diagnostics{}
	}

	access, diags := model.ParseAccess(ctx, cfg)
	if diags.HasError() {
		return nil, diags
	}

	return &opensearch.Access{
		DataTransfer: access.DataTransfer.ValueBool(),
		Serverless:   access.Serverless.ValueBool(),
	}, diag.Diagnostics{}
}

func parseOpenSearchMaintenanceWindow(ctx context.Context, m *model.OpenSearch) (*opensearch.MaintenanceWindow, diag.Diagnostics) {
	mw, diags := model.ParseMaintenanceWindow(ctx, m)
	if diags.HasError() {
		return nil, diags
	}

	result := &opensearch.MaintenanceWindow{}

	switch mw.Type.ValueString() {
	case "ANYTIME":
		if mw.Day.ValueString() != "" || !(mw.Hour.IsUnknown() || mw.Hour.IsNull()) {
			diags.Append(diag.NewErrorDiagnostic(
				"Failed to parse OpenSearch maintenance window",
				"Error while parsing value for 'maintenance_window'. With ANYTIME type of maintenance window both DAY and HOUR should be omitted"))
			return nil, diags
		}

		result.SetAnytime(&opensearch.AnytimeMaintenanceWindow{})
	case "WEEKLY":
		weekly := &opensearch.WeeklyMaintenanceWindow{}
		if mw.Day.ValueString() != "" {
			day, d := parseOpenSearchWeekDay(mw.Day)
			diags.Append(d)
			if diags.HasError() {
				return nil, diags
			}
			weekly.Day = day
		}

		if !(mw.Hour.IsUnknown() || mw.Hour.IsNull()) {
			weekly.Hour = mw.Hour.ValueInt64()
		}

		result.SetWeeklyMaintenanceWindow(weekly)
	default:
		diags.Append(diag.NewErrorDiagnostic(
			"Failed to parse OpenSearch maintenance window",
			fmt.Sprintf("Error while parsing value for 'maintenance_window'. Unknown type '%s'", mw.Type.ValueString())))
		return nil, diags
	}

	return result, diags
}

func parseOpenSearchWeekDay(e basetypes.StringValue) (opensearch.WeeklyMaintenanceWindow_WeekDay, diag.Diagnostic) {
	v, ok := opensearch.WeeklyMaintenanceWindow_WeekDay_value[e.ValueString()]
	if !ok || v == 0 {
		allowedDays := make([]string, 0, len(opensearch.WeeklyMaintenanceWindow_WeekDay_value))
		for k, v := range opensearch.WeeklyMaintenanceWindow_WeekDay_value {
			if v == 0 {
				continue
			}
			allowedDays = append(allowedDays, k)
		}

		return 0, diag.NewErrorDiagnostic(
			"Failed to parse OpenSearch maintenance window",
			fmt.Sprintf("Error while parsing value for 'maintenance_window'. Value for 'day' should be one of `%s`, not `%s`", strings.Join(allowedDays, "`, `"), e),
		)
	}
	return opensearch.WeeklyMaintenanceWindow_WeekDay(v), nil
}

// TODO: refactor this func
func PrepareUpdateClusterParamsRequest(ctx context.Context, state, plan *model.OpenSearch) (*opensearch.UpdateClusterRequest, diag.Diagnostics) {
	clusterID := state.ID.ValueString()

	req := &opensearch.UpdateClusterRequest{
		ClusterId:  clusterID,
		UpdateMask: &field_mask.FieldMask{},
	}

	if !plan.Name.Equal(state.Name) {
		req.Name = plan.Name.ValueString()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if !plan.Description.Equal(state.Description) {
		req.Description = plan.Description.ValueString()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if !plan.Labels.Equal(state.Labels) {
		labels := make(map[string]string, len(plan.Labels.Elements()))
		diags := plan.Labels.ElementsAs(ctx, &labels, false)
		if diags.HasError() {
			return nil, diags
		}

		req.Labels = labels
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	//TODO: what about environment change?

	if !plan.Config.Equal(state.Config) {
		planConfig, stateConfig, diags := model.ParseGenerics(ctx, plan, state, model.ParseConfig)
		if diags.HasError() {
			return nil, diags
		}

		config := &opensearch.ConfigUpdateSpec{}

		if !planConfig.Version.IsUnknown() && !planConfig.Version.Equal(stateConfig.Version) {
			config.Version = planConfig.Version.ValueString()
			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.version")
		}

		//do not check !AdminPassword.IsUnknown() because of planModifier useStateForUnknown
		if !planConfig.AdminPassword.Equal(stateConfig.AdminPassword) {
			config.AdminPassword = planConfig.AdminPassword.ValueString()
			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.admin_password")
		}

		//NOTE: all node_groups will be updated by different requests, so we skip it here and updates only plugins list
		planOpenSearchBlock, stateOpenSearchBlock, d := model.ParseGenerics(ctx, planConfig, stateConfig, model.ParseOpenSearchSubConfig)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		if !planOpenSearchBlock.Plugins.Equal(stateOpenSearchBlock.Plugins) {
			plugins := make([]string, 0, len(planOpenSearchBlock.Plugins.Elements()))
			diags.Append(planOpenSearchBlock.Plugins.ElementsAs(ctx, &plugins, false)...)
			if diags.HasError() {
				return nil, diags
			}

			config.OpensearchSpec = &opensearch.OpenSearchClusterUpdateSpec{
				Plugins: plugins,
			}
			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.opensearch_spec.plugins")
		}

		//NOTE: dashboards contains only node_groups so we skip it

		if !planConfig.Access.Equal(stateConfig.Access) {
			planAccess, diags := model.ParseAccess(ctx, planConfig)
			if diags.HasError() {
				return nil, diags
			}

			stateAccess, diags := model.ParseAccess(ctx, stateConfig)
			if diags.HasError() {
				return nil, diags
			}

			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.access")
			access := &opensearch.Access{}

			if !planAccess.DataTransfer.Equal(stateAccess.DataTransfer) {
				access.DataTransfer = planAccess.DataTransfer.ValueBool()
			}

			if !planAccess.Serverless.Equal(stateAccess.Serverless) {
				access.Serverless = planAccess.Serverless.ValueBool()
			}

			config.Access = access
		}

		req.ConfigSpec = config
	}

	//TODO: what about network_id change?

	if !plan.SecurityGroupIDs.Equal(state.SecurityGroupIDs) {
		securityGroupIDs := make([]string, 0, len(plan.SecurityGroupIDs.Elements()))
		diags := plan.SecurityGroupIDs.ElementsAs(ctx, &securityGroupIDs, false)
		if diags.HasError() {
			return nil, diags
		}

		req.SecurityGroupIds = securityGroupIDs
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "security_group_ids")
	}

	//this condition should be like this because of nil and "" are similar business values
	if plan.ServiceAccountID.ValueString() != state.ServiceAccountID.ValueString() {
		req.ServiceAccountId = plan.ServiceAccountID.ValueString()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "service_account_id")
	}

	if !plan.DeletionProtection.IsUnknown() && !plan.DeletionProtection.Equal(state.DeletionProtection) {
		req.DeletionProtection = plan.DeletionProtection.ValueBool()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "deletion_protection")
	}

	if !plan.MaintenanceWindow.Equal(state.MaintenanceWindow) {
		mw, diags := parseOpenSearchMaintenanceWindow(ctx, plan)
		if diags.HasError() {
			return nil, diags
		}

		req.MaintenanceWindow = mw
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "maintenance_window")
	}

	if len(req.UpdateMask.Paths) == 0 {
		return nil, diag.Diagnostics{}
	}

	return req, diag.Diagnostics{}
}

func PrepareAddOpenSearchNodeGroupRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.OpenSearchCreateSpec_NodeGroup) ([]*opensearch.AddOpenSearchNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(stateNodeGroups)

	var groupsToCreate []*opensearch.OpenSearchCreateSpec_NodeGroup
	for _, g := range planNodeGroups {
		if _, ok := oldGroupsByName[g.Name]; ok {
			continue
		}

		if isManagerOnlyGroup(g) {
			// add manager group to the beginning of the list
			groupsToCreate = append([]*opensearch.OpenSearchCreateSpec_NodeGroup{g}, groupsToCreate...)
		} else {
			groupsToCreate = append(groupsToCreate, g)
		}
	}

	requests := make([]*opensearch.AddOpenSearchNodeGroupRequest, 0, len(groupsToCreate))
	for _, g := range groupsToCreate {
		requests = append(requests, &opensearch.AddOpenSearchNodeGroupRequest{
			ClusterId:     clusterID,
			NodeGroupSpec: g,
		})
	}

	return requests, diag.Diagnostics{}
}

func PrepareDeleteOpenSearchNodeGroupRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.OpenSearchCreateSpec_NodeGroup) ([]*opensearch.DeleteOpenSearchNodeGroupRequest, diag.Diagnostics) {
	newGroupsByName := model.GetGroupByName(planNodeGroups)

	var requests []*opensearch.DeleteOpenSearchNodeGroupRequest
	for _, g := range stateNodeGroups {
		if _, ok := newGroupsByName[g.Name]; ok {
			continue
		}

		requests = append(requests, &opensearch.DeleteOpenSearchNodeGroupRequest{
			ClusterId: clusterID,
			Name:      g.Name,
		})
	}

	return requests, diag.Diagnostics{}
}

func PrepareManagersToIncreaseRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.OpenSearchCreateSpec_NodeGroup) ([]*opensearch.UpdateOpenSearchNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(stateNodeGroups)

	var requests []*opensearch.UpdateOpenSearchNodeGroupRequest
	for _, newGroup := range planNodeGroups {
		oldGroup, ok := oldGroupsByName[newGroup.Name]
		if !ok {
			continue
		}

		if !isManagerOnlyGroup(newGroup) {
			continue
		}

		if newGroup.HostsCount <= oldGroup.HostsCount {
			continue
		}

		request, d := prepareUpdateOpenSearchNodeGroupRequest(clusterID, newGroup, oldGroup)
		if d != nil {
			return nil, diag.Diagnostics{d}
		}

		if len(request.UpdateMask.Paths) == 0 {
			continue
		}

		requests = append(requests, request)
	}

	return requests, diag.Diagnostics{}
}

func PrepareDataManagersToDecreaseRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.OpenSearchCreateSpec_NodeGroup) ([]*opensearch.UpdateOpenSearchNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(stateNodeGroups)

	var requests []*opensearch.UpdateOpenSearchNodeGroupRequest
	for _, newGroup := range planNodeGroups {
		oldGroup, ok := oldGroupsByName[newGroup.Name]
		if !ok {
			continue
		}

		if isManagerOnlyGroup(newGroup) {
			continue
		}

		if !isDecreaseHostsCountOnDataManagedGroup(newGroup, oldGroup) && !isRemovedManagerRole(newGroup, oldGroup) {
			continue
		}

		request, d := prepareUpdateOpenSearchNodeGroupRequest(clusterID, newGroup, oldGroup)
		if d != nil {
			return nil, diag.Diagnostics{d}
		}

		if len(request.UpdateMask.Paths) == 0 {
			continue
		}

		requests = append(requests, request)
	}

	return requests, diag.Diagnostics{}
}

func PrepareOtherGroupsToUpdateRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.OpenSearchCreateSpec_NodeGroup) ([]*opensearch.UpdateOpenSearchNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(stateNodeGroups)

	var requests []*opensearch.UpdateOpenSearchNodeGroupRequest
	for _, newGroup := range planNodeGroups {
		oldGroup, ok := oldGroupsByName[newGroup.Name]
		if !ok {
			continue
		}

		// NOTE: first part is for managersToIncrease and managersToDecrease, second part is for dataManagersToDecrease
		if (isManagerOnlyGroup(newGroup) && newGroup.HostsCount != oldGroup.HostsCount) ||
			(isDecreaseHostsCountOnDataManagedGroup(newGroup, oldGroup) || isRemovedManagerRole(newGroup, oldGroup)) {
			continue
		}

		request, d := prepareUpdateOpenSearchNodeGroupRequest(clusterID, newGroup, oldGroup)
		if d != nil {
			return nil, diag.Diagnostics{d}
		}

		if len(request.UpdateMask.Paths) == 0 {
			continue
		}

		requests = append(requests, request)
	}

	return requests, diag.Diagnostics{}
}

func PrepareManagersToDecreaseRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.OpenSearchCreateSpec_NodeGroup) ([]*opensearch.UpdateOpenSearchNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(stateNodeGroups)

	var requests []*opensearch.UpdateOpenSearchNodeGroupRequest
	for _, newGroup := range planNodeGroups {
		oldGroup, ok := oldGroupsByName[newGroup.Name]
		if !ok {
			continue
		}

		if !isManagerOnlyGroup(newGroup) {
			continue
		}

		if newGroup.HostsCount >= oldGroup.HostsCount {
			continue
		}

		request, d := prepareUpdateOpenSearchNodeGroupRequest(clusterID, newGroup, oldGroup)
		if d != nil {
			return nil, diag.Diagnostics{d}
		}

		if len(request.UpdateMask.Paths) == 0 {
			continue
		}

		requests = append(requests, request)
	}

	return requests, diag.Diagnostics{}
}

func isRemovedManagerRole(newGroup, oldGroup *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return !hasManagerRole(newGroup) && hasManagerRole(oldGroup)
}

func isDecreaseHostsCountOnDataManagedGroup(newGroup, oldGroup *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return hasManagerRole(newGroup) && newGroup.HostsCount < oldGroup.HostsCount
}

func hasManagerRole(group *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return slices.Contains(group.Roles, opensearch.OpenSearch_MANAGER)
}

func prepareUpdateOpenSearchNodeGroupRequest(clusterID string, planNodeGroup, stateNodeGroup *opensearch.OpenSearchCreateSpec_NodeGroup) (*opensearch.UpdateOpenSearchNodeGroupRequest, diag.Diagnostic) {
	if isManagerOnlyGroup(stateNodeGroup) && !isManagerOnlyGroup(planNodeGroup) {
		return nil, diag.NewErrorDiagnostic(
			"Can't update manager-only node group to non-manager-only",
			fmt.Sprintf("Can't change roles for dedicated managers group: %s", stateNodeGroup.Name),
		)
	}

	var paths []string
	nodeGroupSpec := opensearch.OpenSearchNodeGroupUpdateSpec{}
	if planNodeGroup.Resources != stateNodeGroup.Resources {
		nodeGroupSpec.Resources = planNodeGroup.Resources
		planResource := planNodeGroup.Resources
		stateResource := stateNodeGroup.Resources
		paths = appendIfNotEqual(paths, planResource.ResourcePresetId,
			stateResource.ResourcePresetId, "resources.resource_preset_id")
		paths = appendIfNotEqual(paths, planResource.DiskTypeId,
			stateResource.DiskTypeId, "resources.disk_type_id")
		paths = appendIfNotEqual(paths, planResource.DiskSize,
			stateResource.DiskSize, "resources.disk_size")
	}

	if planNodeGroup.HostsCount != stateNodeGroup.HostsCount {
		nodeGroupSpec.HostsCount = planNodeGroup.HostsCount
		paths = append(paths, "hosts_count")
	}

	if !slices.Equal(planNodeGroup.Roles, stateNodeGroup.Roles) {
		nodeGroupSpec.Roles = planNodeGroup.Roles
		paths = append(paths, "roles")
	}

	//TODO: move changing zone_ids and subnet_ids to separate requests
	if !slices.Equal(planNodeGroup.ZoneIds, stateNodeGroup.ZoneIds) {
		paths = append(paths, "zone_ids")
		nodeGroupSpec.ZoneIds = planNodeGroup.ZoneIds
	}

	if !slices.Equal(planNodeGroup.SubnetIds, stateNodeGroup.SubnetIds) {
		paths = append(paths, "subnet_ids")
		nodeGroupSpec.SubnetIds = planNodeGroup.SubnetIds
	}

	//TODO: what about assign_public_ip

	return &opensearch.UpdateOpenSearchNodeGroupRequest{
		ClusterId:     clusterID,
		Name:          planNodeGroup.Name,
		UpdateMask:    &field_mask.FieldMask{Paths: paths},
		NodeGroupSpec: &nodeGroupSpec,
	}, nil
}

func PrepareAddDashboardsNodeGroupRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.DashboardsCreateSpec_NodeGroup) ([]*opensearch.AddDashboardsNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(stateNodeGroups)

	var requests []*opensearch.AddDashboardsNodeGroupRequest
	for _, g := range planNodeGroups {
		if _, ok := oldGroupsByName[g.Name]; ok {
			continue
		}

		requests = append(requests, &opensearch.AddDashboardsNodeGroupRequest{
			ClusterId:     clusterID,
			NodeGroupSpec: g,
		})
	}

	return requests, diag.Diagnostics{}
}

func PrepareUpdateDashboardsNodeGroupRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.DashboardsCreateSpec_NodeGroup) ([]*opensearch.UpdateDashboardsNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(stateNodeGroups)

	var requests []*opensearch.UpdateDashboardsNodeGroupRequest
	for _, g := range planNodeGroups {
		if _, ok := oldGroupsByName[g.Name]; !ok {
			continue
		}

		request := PrepareUpdateDashboardsNodeGroupRequest(clusterID, g, oldGroupsByName[g.Name])
		if len(request.UpdateMask.Paths) == 0 {
			continue
		}

		requests = append(requests, request)
	}

	return requests, diag.Diagnostics{}
}

func PrepareUpdateDashboardsNodeGroupRequest(clusterID string, planNodeGroup, stateNodeGroup *opensearch.DashboardsCreateSpec_NodeGroup) *opensearch.UpdateDashboardsNodeGroupRequest {
	var paths []string
	nodeGroupSpec := opensearch.DashboardsNodeGroupUpdateSpec{}

	if planNodeGroup.Resources != stateNodeGroup.Resources {
		nodeGroupSpec.Resources = planNodeGroup.Resources
		planResource := planNodeGroup.Resources
		stateResource := stateNodeGroup.Resources
		paths = appendIfNotEqual(paths, planResource.ResourcePresetId,
			stateResource.ResourcePresetId, "resources.resource_preset_id")
		paths = appendIfNotEqual(paths, planResource.DiskTypeId,
			stateResource.DiskTypeId, "resources.disk_type_id")
		paths = appendIfNotEqual(paths, planResource.DiskSize,
			stateResource.DiskSize, "resources.disk_size")
	}

	if planNodeGroup.HostsCount != stateNodeGroup.HostsCount {
		nodeGroupSpec.HostsCount = planNodeGroup.HostsCount
		paths = append(paths, "hosts_count")
	}

	//TODO: move changing zone_ids and subnet_ids to separate requests
	if !slices.Equal(planNodeGroup.ZoneIds, stateNodeGroup.ZoneIds) {
		nodeGroupSpec.ZoneIds = planNodeGroup.ZoneIds
		paths = append(paths, "zone_ids")
	}

	if !slices.Equal(planNodeGroup.SubnetIds, stateNodeGroup.SubnetIds) {
		nodeGroupSpec.SubnetIds = planNodeGroup.SubnetIds
		paths = append(paths, "subnet_ids")
	}

	//TODO: what about assign_public_ip

	return &opensearch.UpdateDashboardsNodeGroupRequest{
		ClusterId:     clusterID,
		Name:          planNodeGroup.Name,
		UpdateMask:    &field_mask.FieldMask{Paths: paths},
		NodeGroupSpec: &nodeGroupSpec,
	}
}

func PrepareDeleteDashboardsNodeGroupRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.DashboardsCreateSpec_NodeGroup) ([]*opensearch.DeleteDashboardsNodeGroupRequest, diag.Diagnostics) {
	newGroupsByName := model.GetGroupByName(planNodeGroups)

	var requests []*opensearch.DeleteDashboardsNodeGroupRequest
	for _, g := range stateNodeGroups {
		if _, ok := newGroupsByName[g.Name]; ok {
			continue
		}

		requests = append(requests, &opensearch.DeleteDashboardsNodeGroupRequest{
			ClusterId: clusterID,
			Name:      g.Name,
		})
	}

	return requests, diag.Diagnostics{}
}

func appendIfNotEqual[T, V comparable](slice []V, v1, v2 T, value V) []V {
	if v1 != v2 {
		return append(slice, value)
	}

	return slice
}

func isManagerOnlyGroup(group *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return len(group.GetRoles()) == 1 && group.GetRoles()[0] == opensearch.OpenSearch_MANAGER
}
