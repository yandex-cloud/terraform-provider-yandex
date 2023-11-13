package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"reflect"
	"strings"
)

var EmptyResources = &opensearch.Resources{
	DiskSize:         0,
	DiskTypeId:       "",
	ResourcePresetId: "",
}

func parseOpenSearchEnv(e string) (opensearch.Cluster_Environment, error) {
	v, ok := opensearch.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(opensearch.Cluster_Environment_value)), e)
	}
	return opensearch.Cluster_Environment(v), nil
}

func expandSchemaSet[T any](schemaSet *schema.Set, expander func(d interface{}) T) []T {
	retArray := make([]T, schemaSet.Len())
	for i, v := range schemaSet.List() {
		retArray[i] = expander(v)
	}
	return retArray
}

func openSearchGroupRolesExpander(data interface{}) opensearch.OpenSearch_GroupRole {
	roleString := data.(string)
	return opensearch.OpenSearch_GroupRole(opensearch.OpenSearch_GroupRole_value[strings.ToUpper(roleString)])
}

func expandResources(data *schema.Set) *opensearch.Resources {
	if data == nil || data.Len() == 0 {
		return EmptyResources
	}
	resources := data.List()[0].(map[string]interface{})
	return &opensearch.Resources{
		ResourcePresetId: resources["resource_preset_id"].(string),
		DiskSize:         int64(resources["disk_size"].(int)),
		DiskTypeId:       resources["disk_type_id"].(string),
	}
}

func openSearchCreateSpecNodeGroupExpander(data interface{}) *opensearch.OpenSearchCreateSpec_NodeGroup {
	d := data.(map[string]interface{})
	return &opensearch.OpenSearchCreateSpec_NodeGroup{
		Name:           d["name"].(string),
		Resources:      expandResources(d["resources"].(*schema.Set)),
		HostsCount:     int64(d["hosts_count"].(int)),
		ZoneIds:        convertStringSet(d["zone_ids"].(*schema.Set)),
		SubnetIds:      convertStringSet(d["subnet_ids"].(*schema.Set)),
		AssignPublicIp: d["assign_public_ip"].(bool),
		Roles: expandSchemaSet(
			d["roles"].(*schema.Set),
			openSearchGroupRolesExpander),
	}
}

func dashboardsCreateSpecNodeGroupExpander(data interface{}) *opensearch.DashboardsCreateSpec_NodeGroup {
	d := data.(map[string]interface{})
	return &opensearch.DashboardsCreateSpec_NodeGroup{
		Name:           d["name"].(string),
		Resources:      expandResources(d["resources"].(*schema.Set)),
		HostsCount:     int64(d["hosts_count"].(int)),
		ZoneIds:        convertStringSet(d["zone_ids"].(*schema.Set)),
		SubnetIds:      convertStringSet(d["subnet_ids"].(*schema.Set)),
		AssignPublicIp: d["assign_public_ip"].(bool),
	}
}

func expandOpenSearchConfigCreateSpec(raw interface{}) *opensearch.ConfigCreateSpec {
	rawList := raw.([]interface{})
	if len(rawList) == 0 {
		return nil
	}
	d := rawList[0].(map[string]interface{})
	config := &opensearch.ConfigCreateSpec{
		Version:       d["version"].(string),
		AdminPassword: d["admin_password"].(string),
	}

	openSearchList := d["opensearch"].([]interface{})
	if len(openSearchList) != 0 {
		openSearch := openSearchList[0].(map[string]interface{})

		config.OpensearchSpec = &opensearch.OpenSearchCreateSpec{
			NodeGroups: expandSchemaSet(
				openSearch["node_groups"].(*schema.Set),
				openSearchCreateSpecNodeGroupExpander),
			Plugins: expandStringSet(openSearch["plugins"]),
		}
	}

	dashboardsList := d["dashboards"].([]interface{})
	if len(openSearchList) != 0 {
		dashboards := dashboardsList[0].(map[string]interface{})

		config.DashboardsSpec = &opensearch.DashboardsCreateSpec{
			NodeGroups: expandSchemaSet(
				dashboards["node_groups"].(*schema.Set),
				dashboardsCreateSpecNodeGroupExpander),
		}
	}

	return config
}

func expandOpenSearchClusterUpdateSpec(d *schema.ResourceData, mask *fieldmaskpb.FieldMask) *opensearch.OpenSearchClusterUpdateSpec {
	config := &opensearch.OpenSearchClusterUpdateSpec{}
	if d.HasChange("config.0.opensearch.0.plugins") {
		config.Plugins = convertStringSet(d.Get("config.0.opensearch.0.plugins").(*schema.Set))
		mask.Paths = append(mask.Paths, "config_spec.opensearch_spec.plugins")
	}
	return config
}

func expandDashboardsClusterUpdateSpec(d *schema.ResourceData, mask *fieldmaskpb.FieldMask) *opensearch.DashboardsClusterUpdateSpec {
	config := &opensearch.DashboardsClusterUpdateSpec{}
	return config
}

func expandAccessUpdateSpec(d *schema.ResourceData, mask *fieldmaskpb.FieldMask) *opensearch.Access {
	config := &opensearch.Access{}
	if d.HasChange("config.0.opensearch.0.access.0.data_transfer") {
		config.DataTransfer = d.Get("config.0.opensearch.0.access.0.data_transfer").(bool)
		mask.Paths = append(mask.Paths, "config_spec.access")
	}
	return config
}

func expandOpenSearchConfigUpdateSpec(d *schema.ResourceData, mask *fieldmaskpb.FieldMask) *opensearch.ConfigUpdateSpec {
	config := &opensearch.ConfigUpdateSpec{}
	if d.HasChange("config.0.version") {
		config.Version = d.Get("config.0.version").(string)
		mask.Paths = append(mask.Paths, "config_spec.version")
	}

	if d.HasChange("config.0.admin_password") {
		config.AdminPassword = d.Get("config.0.admin_password").(string)
		mask.Paths = append(mask.Paths, "config_spec.admin_password")
	}

	if _, exist := d.GetOk("config.0.opensearch"); exist {
		config.OpensearchSpec = expandOpenSearchClusterUpdateSpec(d, mask)
	}

	if _, exist := d.GetOk("config.0.dashboards"); exist {
		config.DashboardsSpec = expandDashboardsClusterUpdateSpec(d, mask)
	}

	if _, exist := d.GetOk("config.0.access"); exist {
		config.Access = expandAccessUpdateSpec(d, mask)
	}

	return config
}

func flattenOpenSearchNodeGroupRoles(r []opensearch.OpenSearch_GroupRole) []string {
	roles := make([]string, len(r))
	for i, v := range r {
		roles[i] = v.String()
	}
	return roles
}

func flattenOpenSearchNodeGroups(nodeGroups []*opensearch.OpenSearch_NodeGroup) []interface{} {
	var ret = make([]interface{}, len(nodeGroups))
	for i, v := range nodeGroups {
		flattened := map[string]interface{}{
			"name":             v.Name,
			"roles":            flattenOpenSearchNodeGroupRoles(v.Roles),
			"assign_public_ip": v.AssignPublicIp,
			"hosts_count":      v.HostsCount,
			"subnet_ids":       v.SubnetIds,
			"zone_ids":         v.ZoneIds,
		}
		if v.Resources != EmptyResources {
			flattened["resources"] = []interface{}{map[string]interface{}{
				"resource_preset_id": v.Resources.ResourcePresetId,
				"disk_size":          v.Resources.DiskSize,
				"disk_type_id":       v.Resources.DiskTypeId,
			}}
		} else {
			flattened["resources"] = []interface{}{interface{}(nil)}
		}
		ret[i] = flattened
	}
	return ret
}

func flattenOpenSearchCreateSpecNodeGroups(nodeGroups []*opensearch.OpenSearchCreateSpec_NodeGroup) []interface{} {
	var ret = make([]interface{}, len(nodeGroups))
	//	for i, v := range config.Opensearch.NodeGroups {
	for i, v := range nodeGroups {
		flattened := map[string]interface{}{
			"name":             v.Name,
			"roles":            flattenOpenSearchNodeGroupRoles(v.Roles),
			"assign_public_ip": v.AssignPublicIp,
			"hosts_count":      v.HostsCount,
			"subnet_ids":       v.SubnetIds,
			"zone_ids":         v.ZoneIds,
		}
		if v.Resources != EmptyResources {
			flattened["resources"] = []interface{}{map[string]interface{}{
				"resource_preset_id": v.Resources.ResourcePresetId,
				"disk_size":          v.Resources.DiskSize,
				"disk_type_id":       v.Resources.DiskTypeId,
			}}
		} else {
			flattened["resources"] = []interface{}{interface{}(nil)}
		}
		ret[i] = flattened
	}
	return ret
}

func flattenDashboardsNodeGroups(config *opensearch.ClusterConfig) []interface{} {
	var ret = make([]interface{}, len(config.Dashboards.NodeGroups))
	for i, v := range config.Dashboards.NodeGroups {
		flattened := map[string]interface{}{
			"name":             v.Name,
			"assign_public_ip": v.AssignPublicIp,
			"hosts_count":      v.HostsCount,
			"subnet_ids":       v.SubnetIds,
			"zone_ids":         v.ZoneIds,
		}
		if v.Resources != EmptyResources {
			flattened["resources"] = []interface{}{map[string]interface{}{
				"resource_preset_id": v.Resources.ResourcePresetId,
				"disk_size":          v.Resources.DiskSize,
				"disk_type_id":       v.Resources.DiskTypeId,
			}}
		} else {
			flattened["resources"] = []interface{}{interface{}(nil)}
		}
		ret[i] = flattened
	}
	return ret
}

func flattenDashboardsCreateSpecNodeGroups(nodeGroups []*opensearch.DashboardsCreateSpec_NodeGroup) []interface{} {
	var ret = make([]interface{}, len(nodeGroups))
	for i, v := range nodeGroups {
		flattened := map[string]interface{}{
			"name":             v.Name,
			"assign_public_ip": v.AssignPublicIp,
			"hosts_count":      v.HostsCount,
			"subnet_ids":       v.SubnetIds,
			"zone_ids":         v.ZoneIds,
		}
		if v.Resources != EmptyResources {
			flattened["resources"] = []interface{}{map[string]interface{}{
				"resource_preset_id": v.Resources.ResourcePresetId,
				"disk_size":          v.Resources.DiskSize,
				"disk_type_id":       v.Resources.DiskTypeId,
			}}
		} else {
			flattened["resources"] = []interface{}{interface{}(nil)}
		}
		ret[i] = flattened
	}
	return ret
}

func flattenOpenSearchClusterConfig(config *opensearch.ClusterConfig, password string) []map[string]interface{} {
	res := map[string]interface{}{
		"version":        config.Version,
		"admin_password": password,
	}
	if config.Access != nil {
		res["access"] = []map[string]interface{}{{
			"data_transfer": config.Access.DataTransfer,
			"serverless":    config.Access.Serverless,
		}}
	}
	openSearchConfig := map[string]interface{}{}
	openSearchConfig["node_groups"] = flattenOpenSearchNodeGroups(config.Opensearch.NodeGroups)
	openSearchConfig["plugins"] = config.Opensearch.Plugins
	res["opensearch"] = []map[string]interface{}{openSearchConfig}

	dashboardsConfig := map[string]interface{}{}
	dashboardsConfig["node_groups"] = flattenDashboardsNodeGroups(config)
	res["dashboards"] = []map[string]interface{}{dashboardsConfig}

	return []map[string]interface{}{res}
}

func flattenOpenSearchConfigCreateSpec(config *opensearch.ConfigCreateSpec) []map[string]interface{} {
	res := map[string]interface{}{
		"version":        config.Version,
		"admin_password": config.AdminPassword,
	}
	if config.Access != nil {
		res["access"] = []map[string]interface{}{{
			"data_transfer": config.Access.DataTransfer,
			"serverless":    config.Access.Serverless,
		}}
	}
	openSearchConfig := map[string]interface{}{}
	openSearchConfig["node_groups"] = flattenOpenSearchCreateSpecNodeGroups(config.OpensearchSpec.NodeGroups)
	openSearchConfig["plugins"] = config.OpensearchSpec.Plugins
	res["opensearch"] = []map[string]interface{}{openSearchConfig}

	dashboardsConfig := map[string]interface{}{}
	dashboardsConfig["node_groups"] = flattenDashboardsCreateSpecNodeGroups(config.DashboardsSpec.NodeGroups)
	res["dashboards"] = []map[string]interface{}{dashboardsConfig}

	return []map[string]interface{}{res}
}

func openSearchRoleHash(v interface{}) int {
	if v == nil {
		return 0
	}
	return hashcode.String(strings.ToUpper(v.(string)))
}

func setHash(set *schema.Set) int {
	var hashCode int = 2166136261
	for _, v := range set.List() {
		hashCode = (hashCode * 16777619) ^ hashcode.String(v.(string))
	}
	return hashCode
}

func upperCaseStringSetHash(set *schema.Set) int {
	var hashCode int = 2166136261
	for _, v := range set.List() {
		hashCode = (hashCode * 16777619) ^ hashcode.String(strings.ToUpper(v.(string)))
	}
	return hashCode
}

func openSearchResourcesHash(v *schema.Set) int {
	if v == nil || v.Len() == 0 {
		return 0
	}
	resources := v.List()[0].(map[string]interface{})
	var hashCode int = 2166136261
	hashCode += (hashCode * 16777619) ^ hashcode.String(resources["resource_preset_id"].(string))
	hashCode += (hashCode * 16777619) ^ schema.HashInt(resources["disk_size"].(int))
	hashCode += (hashCode * 16777619) ^ hashcode.String(resources["disk_type_id"].(string))
	if hashCode < 0 {
		hashCode = -hashCode
	}
	return hashCode
}

func openSearchNodeGroupDeepHash(v interface{}) int {
	group := v.(map[string]interface{})

	var hashCode int = 2166136261
	hashCode = (hashCode * 16777619) ^ hashcode.String(group["name"].(string))
	resources := group["resources"]
	if resources != nil {
		hashCode = (hashCode * 16777619) ^ openSearchResourcesHash(resources.(*schema.Set))
	}
	hashCode = (hashCode * 16777619) ^ schema.HashInt(group["hosts_count"].(int))
	hashCode = (hashCode * 16777619) ^ setHash(group["zone_ids"].(*schema.Set))
	hashCode = (hashCode * 16777619) ^ setHash(group["subnet_ids"].(*schema.Set))
	if group["assign_public_ip"].(bool) {
		hashCode = (hashCode * 16777619) ^ 1
	}
	hashCode = (hashCode * 16777619) ^ upperCaseStringSetHash(group["roles"].(*schema.Set))
	if hashCode < 0 {
		hashCode = -hashCode
	}
	return hashCode
}

func parseOpenSearchWeekDay(wd string) (opensearch.WeeklyMaintenanceWindow_WeekDay, error) {
	val, ok := opensearch.WeeklyMaintenanceWindow_WeekDay_value[wd]
	// do not allow WEEK_DAY_UNSPECIFIED
	if !ok || val == 0 {
		return opensearch.WeeklyMaintenanceWindow_WEEK_DAY_UNSPECIFIED,
			fmt.Errorf("value for 'day' should be one of %s, not `%s`",
				getJoinedKeys(getEnumValueMapKeysExt(opensearch.WeeklyMaintenanceWindow_WeekDay_value, true)), wd)
	}

	return opensearch.WeeklyMaintenanceWindow_WeekDay(val), nil
}

func expandOpenSearchMaintenanceWindow(d *schema.ResourceData) (*opensearch.MaintenanceWindow, error) {
	mwType, ok := d.GetOk("maintenance_window.0.type")
	if !ok {
		return nil, nil
	}

	result := &opensearch.MaintenanceWindow{}

	switch mwType {
	case "ANYTIME":
		timeSet := false
		if _, ok := d.GetOk("maintenance_window.0.day"); ok {
			timeSet = true
		}
		if _, ok := d.GetOk("maintenance_window.0.hour"); ok {
			timeSet = true
		}
		if timeSet {
			return nil, fmt.Errorf("with ANYTIME type of maintenance window both DAY and HOUR should be omitted")
		}
		result.SetAnytime(&opensearch.AnytimeMaintenanceWindow{})

	case "WEEKLY":
		weekly := &opensearch.WeeklyMaintenanceWindow{}
		if val, ok := d.GetOk("maintenance_window.0.day"); ok {
			var err error
			weekly.Day, err = parseOpenSearchWeekDay(val.(string))
			if err != nil {
				return nil, err
			}
		}
		if v, ok := d.GetOk("maintenance_window.0.hour"); ok {
			weekly.Hour = int64(v.(int))
		}

		result.SetWeeklyMaintenanceWindow(weekly)
	}

	return result, nil
}

func flattenOpenSearchMaintenanceWindow(mw *opensearch.MaintenanceWindow) []map[string]interface{} {
	result := map[string]interface{}{}

	if val := mw.GetAnytime(); val != nil {
		result["type"] = "ANYTIME"
	}

	if val := mw.GetWeeklyMaintenanceWindow(); val != nil {
		result["type"] = "WEEKLY"
		result["day"] = val.Day.String()
		result["hour"] = val.Hour
	}

	return []map[string]interface{}{result}
}

func createAddOpenSearchNodeGroupRequest(clusterId string, group *opensearch.OpenSearchCreateSpec_NodeGroup) (*opensearch.AddOpenSearchNodeGroupRequest, error) {
	var request = &opensearch.AddOpenSearchNodeGroupRequest{
		ClusterId:     clusterId,
		NodeGroupSpec: group,
	}
	return request, nil
}

func createAddDashboardsNodeGroupRequest(clusterId string, group *opensearch.DashboardsCreateSpec_NodeGroup) (*opensearch.AddDashboardsNodeGroupRequest, error) {
	var request = &opensearch.AddDashboardsNodeGroupRequest{
		ClusterId:     clusterId,
		NodeGroupSpec: group,
	}
	return request, nil
}

func containsRole(roles []opensearch.OpenSearch_GroupRole, r opensearch.OpenSearch_GroupRole) bool {
	for _, a := range roles {
		if a == r {
			return true
		}
	}
	return false
}

func hasManagerRole(group *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return containsRole(group.Roles, opensearch.OpenSearch_MANAGER)
}

func dedicatedManagersGroup(group *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return len(group.Roles) == 1 && group.Roles[0] == opensearch.OpenSearch_MANAGER
}

func createUpdateOpenSearchNodeGroupRequest(clusterId string, oldGroup *opensearch.OpenSearchCreateSpec_NodeGroup, newGroup *opensearch.OpenSearchCreateSpec_NodeGroup) (*opensearch.UpdateOpenSearchNodeGroupRequest, error) {
	request := &opensearch.UpdateOpenSearchNodeGroupRequest{
		ClusterId: clusterId,
		Name:      newGroup.Name,
	}
	paths := make([]string, 0, 8)
	var nodeGroupSpec opensearch.OpenSearchNodeGroupUpdateSpec
	request.NodeGroupSpec = &nodeGroupSpec
	if oldGroup.Resources != newGroup.Resources {
		nodeGroupSpec.Resources = newGroup.Resources
		newResources := newGroup.Resources
		oldResources := oldGroup.Resources
		if newResources.ResourcePresetId != oldResources.ResourcePresetId {
			paths = append(paths, "resources.resource_preset_id")
		}
		if newResources.DiskTypeId != oldResources.DiskTypeId {
			paths = append(paths, "resources.disk_type_id")
		}
		if newResources.DiskSize != oldResources.DiskSize {
			paths = append(paths, "resources.disk_size")
		}
	}
	if oldGroup.HostsCount != newGroup.HostsCount {
		paths = append(paths, "hosts_count")
		nodeGroupSpec.HostsCount = newGroup.HostsCount
	}
	if !reflect.DeepEqual(oldGroup.Roles, newGroup.Roles) {
		paths = append(paths, "roles")
		if dedicatedManagersGroup(oldGroup) && !dedicatedManagersGroup(newGroup) {
			return nil, fmt.Errorf("can't change roles for dedicated managers group: %s", oldGroup.Name)
		}
		nodeGroupSpec.Roles = newGroup.Roles
	}
	request.UpdateMask = &field_mask.FieldMask{
		Paths: paths,
	}
	return request, nil
}

func createUpdateDashboardsNodeGroupRequest(clusterId string, oldGroup *opensearch.DashboardsCreateSpec_NodeGroup, newGroup *opensearch.DashboardsCreateSpec_NodeGroup) (*opensearch.UpdateDashboardsNodeGroupRequest, error) {
	request := &opensearch.UpdateDashboardsNodeGroupRequest{
		ClusterId: clusterId,
		Name:      newGroup.Name,
	}
	paths := make([]string, 0, 8)
	var nodeGroupSpec opensearch.DashboardsNodeGroupUpdateSpec
	request.NodeGroupSpec = &nodeGroupSpec
	if oldGroup.Resources != newGroup.Resources {
		nodeGroupSpec.Resources = newGroup.Resources
		newResources := newGroup.Resources
		oldResources := oldGroup.Resources
		if newResources.ResourcePresetId != oldResources.ResourcePresetId {
			paths = append(paths, "resources.resource_preset_id")
		}
		if newResources.DiskTypeId != oldResources.DiskTypeId {
			paths = append(paths, "resources.disk_type_id")
		}
		if newResources.DiskSize != oldResources.DiskSize {
			paths = append(paths, "resources.disk_size")
		}
	}
	if oldGroup.HostsCount != newGroup.HostsCount {
		paths = append(paths, "hosts_count")
		nodeGroupSpec.HostsCount = newGroup.HostsCount
	}
	request.UpdateMask = &field_mask.FieldMask{
		Paths: paths,
	}
	return request, nil
}

func createDeleteOpenSearchNodeGroupRequest(clusterId string, group *opensearch.OpenSearchCreateSpec_NodeGroup) *opensearch.DeleteOpenSearchNodeGroupRequest {
	var request = &opensearch.DeleteOpenSearchNodeGroupRequest{
		ClusterId: clusterId,
		Name:      group.Name,
	}
	return request
}

func createDeleteDashboardsNodeGroupRequest(clusterId string, group *opensearch.DashboardsCreateSpec_NodeGroup) *opensearch.DeleteDashboardsNodeGroupRequest {
	var request = &opensearch.DeleteDashboardsNodeGroupRequest{
		ClusterId: clusterId,
		Name:      group.Name,
	}
	return request
}

func copyOpenSearchNodeGroupsData(oldGroups []*opensearch.OpenSearchCreateSpec_NodeGroup, newGroups []*opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	var oldGroupsByName = map[string]*opensearch.OpenSearchCreateSpec_NodeGroup{}
	for _, g := range oldGroups {
		oldGroupsByName[g.Name] = g
	}

	modified := false
	for _, newGroup := range newGroups {
		if oldGroup, ok := oldGroupsByName[newGroup.Name]; ok {
			if newGroup.Roles == nil {
				newGroup.Roles = oldGroup.Roles
				modified = true
			}
			if newGroup.Resources == nil || newGroup.Resources == EmptyResources {
				newGroup.Resources = oldGroup.Resources
				modified = true
			}
			if newGroup.ZoneIds == nil {
				newGroup.ZoneIds = oldGroup.ZoneIds
				modified = true
			}
			if newGroup.SubnetIds == nil || len(newGroup.SubnetIds) == 0 {
				newGroup.SubnetIds = oldGroup.SubnetIds
				modified = true
			}
			if newGroup.HostsCount == 0 {
				newGroup.HostsCount = oldGroup.HostsCount
				modified = true
			}
		}
	}
	return modified
}

func copyDashboardsNodeGroupsData(oldGroups []*opensearch.DashboardsCreateSpec_NodeGroup, newGroups []*opensearch.DashboardsCreateSpec_NodeGroup) bool {
	var oldGroupsByName = map[string]*opensearch.DashboardsCreateSpec_NodeGroup{}
	for _, g := range oldGroups {
		oldGroupsByName[g.Name] = g
	}

	modified := false
	for _, newGroup := range newGroups {
		if oldGroup, ok := oldGroupsByName[newGroup.Name]; ok {
			if newGroup.Resources == nil || newGroup.Resources == EmptyResources {
				newGroup.Resources = oldGroup.Resources
				modified = true
			}
			if newGroup.ZoneIds == nil {
				newGroup.ZoneIds = oldGroup.ZoneIds
				modified = true
			}
			if newGroup.SubnetIds == nil || len(newGroup.SubnetIds) == 0 {
				newGroup.SubnetIds = oldGroup.SubnetIds
				modified = true
			}
			if newGroup.HostsCount == 0 {
				newGroup.HostsCount = oldGroup.HostsCount
				modified = true
			}
		}
	}
	return modified
}

func opensearchNodeGroupsDiffCustomize(ctx context.Context, rdiff *schema.ResourceDiff, _ interface{}) error {
	oc, nc := rdiff.GetChange("config")
	if oc == nil {
		if nc == nil {
			return fmt.Errorf("Missing required option: config")
		}
	}
	oldConfig := expandOpenSearchConfigCreateSpec(oc)
	newConfig := expandOpenSearchConfigCreateSpec(nc)

	modified := false
	if oldConfig != nil {
		if copyOpenSearchNodeGroupsData(oldConfig.OpensearchSpec.NodeGroups, newConfig.OpensearchSpec.NodeGroups) {
			modified = true
		}
		if copyDashboardsNodeGroupsData(oldConfig.DashboardsSpec.NodeGroups, newConfig.DashboardsSpec.NodeGroups) {
			modified = true
		}

		if newConfig.OpensearchSpec.Plugins == nil || len(oldConfig.OpensearchSpec.Plugins) == 0 {
			newConfig.OpensearchSpec.Plugins = oldConfig.OpensearchSpec.Plugins
			modified = true
		}
	}

	if modified {
		flattened := flattenOpenSearchConfigCreateSpec(newConfig)
		return rdiff.SetNew("config", flattened)
	}

	return nil
}
