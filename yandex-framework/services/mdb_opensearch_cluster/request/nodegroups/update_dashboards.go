package nodegroups

import (
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
	"google.golang.org/genproto/protobuf/field_mask"
)

func PrepareAddDashboardsRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.DashboardsCreateSpec_NodeGroup) ([]*opensearch.AddDashboardsNodeGroupRequest, diag.Diagnostics) {
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

func PrepareUpdateDashboardsRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.DashboardsCreateSpec_NodeGroup) ([]*opensearch.UpdateDashboardsNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(stateNodeGroups)

	var requests []*opensearch.UpdateDashboardsNodeGroupRequest
	for _, g := range planNodeGroups {
		if _, ok := oldGroupsByName[g.Name]; !ok {
			continue
		}

		request := prepareUpdateDashboardsRequest(clusterID, g, oldGroupsByName[g.Name])
		if len(request.UpdateMask.Paths) == 0 {
			continue
		}

		requests = append(requests, request)
	}

	return requests, diag.Diagnostics{}
}

func PrepareUpdateDashboardsZoneAndSubnetIdsRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.DashboardsCreateSpec_NodeGroup) ([]*opensearch.UpdateDashboardsNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(stateNodeGroups)

	var requests []*opensearch.UpdateDashboardsNodeGroupRequest
	for _, g := range planNodeGroups {
		if _, ok := oldGroupsByName[g.Name]; !ok {
			continue
		}

		request := prepareUpdateDashboardsZoneAndSubnetIdsRequest(clusterID, g, oldGroupsByName[g.Name])
		if len(request.UpdateMask.Paths) == 0 {
			continue
		}

		requests = append(requests, request)
	}

	return requests, diag.Diagnostics{}
}

func prepareUpdateDashboardsRequest(clusterID string, planNodeGroup, stateNodeGroup *opensearch.DashboardsCreateSpec_NodeGroup) *opensearch.UpdateDashboardsNodeGroupRequest {
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

	if planNodeGroup.AssignPublicIp != stateNodeGroup.AssignPublicIp {
		nodeGroupSpec.AssignPublicIp = planNodeGroup.AssignPublicIp
		paths = append(paths, "assign_public_ip")
	}

	return &opensearch.UpdateDashboardsNodeGroupRequest{
		ClusterId:     clusterID,
		Name:          planNodeGroup.Name,
		UpdateMask:    &field_mask.FieldMask{Paths: paths},
		NodeGroupSpec: &nodeGroupSpec,
	}
}

func prepareUpdateDashboardsZoneAndSubnetIdsRequest(clusterID string, planNodeGroup, stateNodeGroup *opensearch.DashboardsCreateSpec_NodeGroup) *opensearch.UpdateDashboardsNodeGroupRequest {
	var paths []string
	nodeGroupSpec := opensearch.DashboardsNodeGroupUpdateSpec{}
	if !slices.Equal(planNodeGroup.ZoneIds, stateNodeGroup.ZoneIds) {
		nodeGroupSpec.ZoneIds = planNodeGroup.ZoneIds
		paths = append(paths, "zone_ids")
	}

	if !slices.Equal(planNodeGroup.SubnetIds, stateNodeGroup.SubnetIds) {
		nodeGroupSpec.SubnetIds = planNodeGroup.SubnetIds
		paths = append(paths, "subnet_ids")
	}

	return &opensearch.UpdateDashboardsNodeGroupRequest{
		ClusterId:     clusterID,
		Name:          planNodeGroup.Name,
		UpdateMask:    &field_mask.FieldMask{Paths: paths},
		NodeGroupSpec: &nodeGroupSpec,
	}
}

func PrepareDeleteDashboardsRequests(clusterID string, planNodeGroups, stateNodeGroups []*opensearch.DashboardsCreateSpec_NodeGroup) ([]*opensearch.DeleteDashboardsNodeGroupRequest, diag.Diagnostics) {
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
