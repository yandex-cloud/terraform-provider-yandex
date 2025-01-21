package nodegroups

import (
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
	"google.golang.org/genproto/protobuf/field_mask"
)

type openSearchNodeGroup = *opensearch.OpenSearchCreateSpec_NodeGroup
type openSearchNodeGroups = []*opensearch.OpenSearchCreateSpec_NodeGroup
type openSearchNodeGroupUpdate = *opensearch.UpdateOpenSearchNodeGroupRequest

func PrepareAddOpenSearchRequests(clusterID string, plan, state openSearchNodeGroups) ([]*opensearch.AddOpenSearchNodeGroupRequest, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(state)

	var groupsToCreate openSearchNodeGroups
	for _, g := range plan {
		if _, ok := oldGroupsByName[g.Name]; ok {
			continue
		}

		if managerOnlyGroup(g) {
			// add manager group to the beginning of the list
			groupsToCreate = append(openSearchNodeGroups{g}, groupsToCreate...)
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

func PrepareDeleteOpenSearchRequests(clusterID string, plan, state openSearchNodeGroups) ([]*opensearch.DeleteOpenSearchNodeGroupRequest, diag.Diagnostics) {
	newGroupsByName := model.GetGroupByName(plan)

	var requests []*opensearch.DeleteOpenSearchNodeGroupRequest
	for _, g := range state {
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

func PrepareManagersToIncreaseRequests(clusterID string, plan, state openSearchNodeGroups) ([]openSearchNodeGroupUpdate, diag.Diagnostics) {
	return prepareRequests(clusterID, plan, state, prepareUpdateOpenSearchRequest, func(new, old openSearchNodeGroup) bool {
		if !managerOnlyGroup(new) {
			return false
		}

		return new.HostsCount > old.HostsCount
	})
}

func PrepareDataManagersToDecreaseRequests(clusterID string, plan, state openSearchNodeGroups) ([]openSearchNodeGroupUpdate, diag.Diagnostics) {
	return prepareRequests(clusterID, plan, state, prepareUpdateOpenSearchRequest, func(new, old openSearchNodeGroup) bool {
		if managerOnlyGroup(new) {
			return false
		}

		return hostCountOnDataManagedGroupDecreased(new, old) || managerRoleRemoved(new, old)
	})
}

func PrepareOtherGroupsToUpdateRequests(clusterID string, plan, state openSearchNodeGroups) ([]openSearchNodeGroupUpdate, diag.Diagnostics) {
	return prepareRequests(clusterID, plan, state, prepareUpdateOpenSearchRequest, func(new, old openSearchNodeGroup) bool {
		// NOTE: first part is for managersToIncrease and managersToDecrease, second part is for dataManagersToDecrease
		return (!managerOnlyGroup(new) || new.HostsCount == old.HostsCount) &&
			(!hostCountOnDataManagedGroupDecreased(new, old) && !managerRoleRemoved(new, old))
	})
}

func PrepareManagersToDecreaseRequests(clusterID string, plan, state openSearchNodeGroups) ([]openSearchNodeGroupUpdate, diag.Diagnostics) {
	return prepareRequests(clusterID, plan, state, prepareUpdateOpenSearchRequest, func(new, old openSearchNodeGroup) bool {
		if !managerOnlyGroup(new) {
			return false
		}

		return new.HostsCount < old.HostsCount
	})
}

type condition func(new, old openSearchNodeGroup) bool
type maker func(clusterID string, plan, state openSearchNodeGroup) openSearchNodeGroupUpdate

func prepareRequests(clusterID string, plan, state openSearchNodeGroups, toRequest maker, c condition) ([]openSearchNodeGroupUpdate, diag.Diagnostics) {
	oldGroupsByName := model.GetGroupByName(state)

	var requests []openSearchNodeGroupUpdate
	for _, newGroup := range plan {
		oldGroup, ok := oldGroupsByName[newGroup.Name]
		if !ok {
			continue
		}

		if managerOnlyGroup(oldGroup) && !managerOnlyGroup(newGroup) {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Can't update manager-only node group to non-manager-only",
				fmt.Sprintf("Can't change roles for dedicated managers group: %s", oldGroup.Name),
			)}
		}

		if !c(newGroup, oldGroup) {
			continue
		}

		request := toRequest(clusterID, newGroup, oldGroup)

		if len(request.UpdateMask.Paths) == 0 {
			continue
		}

		requests = append(requests, request)
	}

	return requests, diag.Diagnostics{}
}

func managerRoleRemoved(newGroup, oldGroup *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return !hasManagerRole(newGroup) && hasManagerRole(oldGroup)
}

func hostCountOnDataManagedGroupDecreased(newGroup, oldGroup *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return hasManagerRole(newGroup) && newGroup.HostsCount < oldGroup.HostsCount
}

func hasManagerRole(group *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return slices.Contains(group.Roles, opensearch.OpenSearch_MANAGER)
}

func managerOnlyGroup(group *opensearch.OpenSearchCreateSpec_NodeGroup) bool {
	return len(group.GetRoles()) == 1 && group.GetRoles()[0] == opensearch.OpenSearch_MANAGER
}

func prepareUpdateOpenSearchRequest(clusterID string, plan, state openSearchNodeGroup) openSearchNodeGroupUpdate {
	var paths []string
	nodeGroupSpec := opensearch.OpenSearchNodeGroupUpdateSpec{}
	if plan.Resources != state.Resources {
		nodeGroupSpec.Resources = plan.Resources
		planResource := plan.Resources
		stateResource := state.Resources
		paths = appendIfNotEqual(paths, planResource.ResourcePresetId,
			stateResource.ResourcePresetId, "resources.resource_preset_id")
		paths = appendIfNotEqual(paths, planResource.DiskTypeId,
			stateResource.DiskTypeId, "resources.disk_type_id")
		paths = appendIfNotEqual(paths, planResource.DiskSize,
			stateResource.DiskSize, "resources.disk_size")
	}

	if plan.HostsCount != state.HostsCount {
		nodeGroupSpec.HostsCount = plan.HostsCount
		paths = append(paths, "hosts_count")
	}

	if !slices.Equal(plan.Roles, state.Roles) {
		nodeGroupSpec.Roles = plan.Roles
		paths = append(paths, "roles")
	}

	if !slices.Equal(plan.ZoneIds, state.ZoneIds) {
		paths = append(paths, "zone_ids")
		nodeGroupSpec.ZoneIds = plan.ZoneIds
	}

	if !slices.Equal(plan.SubnetIds, state.SubnetIds) {
		paths = append(paths, "subnet_ids")
		nodeGroupSpec.SubnetIds = plan.SubnetIds
	}

	if plan.AssignPublicIp != state.AssignPublicIp {
		paths = append(paths, "assign_public_ip")
		nodeGroupSpec.AssignPublicIp = plan.AssignPublicIp
	}

	return &opensearch.UpdateOpenSearchNodeGroupRequest{
		ClusterId:     clusterID,
		Name:          plan.Name,
		UpdateMask:    &field_mask.FieldMask{Paths: paths},
		NodeGroupSpec: &nodeGroupSpec,
	}
}
