package mdbcommon

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	ycsdk "github.com/yandex-cloud/go-sdk"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
)

// T - terraform Host struct
// ProtoHost - protobuf Host
// ProtoHostSpec - protobuf HostSpec
// UpdateSpec - protobuf UpdateSpec

type Host interface {
	GetFQDN() types.String
}

type HostWithShard interface {
	Host
	GetShard() string
}

type ProtoHost interface {
	GetName() string
}
type ProtoHostWithShard interface {
	GetShardName() string
}

// HostApiService is an interface that defines methods for API operations involving hosts.
// It is parameterized with types `ProtoHost`, `ProtoHostSpec`, and `UpdateSpec`
type HostApiService[ProtoHost any, ProtoHostSpec any, UpdateSpec any] interface {
	ListHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) []ProtoHost
	CreateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []ProtoHostSpec)
	UpdateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*UpdateSpec)
	DeleteHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, fqdns []string)
}

// HostApiService is an interface that defines methods for API operations involving hosts with shards.
// It is parameterized with types `ProtoHost`, `ProtoHostSpec`, and `UpdateSpec`
type HostApiServiceWithShards[ProtoHost any, ProtoHostSpec any, UpdateSpec any] interface {
	HostApiService[ProtoHost, ProtoHostSpec, UpdateSpec]
	CreateShard(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, shardName string, hostSpecs []ProtoHostSpec)
	DeleteShard(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, shardName string)
}

// CmpHostService is an interface for implementing common parsing methods for host models.
// T - Represents the host model within Terraform, likely detailing configuration and state as defined in Terraform templates.
// PortoHost - Refers to the protobuf type returned by a listHosts API call. It represents how hosts are structured in API communications.
// ProtoHostSpec - Is the protobuf type used for creating a host. It specifies the properties and configurations needed for API host creation.
// UpdateSpec - The protobuf type used for updating hosts, detailing the necessary changes.
type CmpHostService[T Host, ProtoHost any, ProtoHostSpec any, UpdateSpec any] interface {
	// FullyMatch performs a complete comparison of all fields in two host models, similar to an equals method.
	// It also handles cases where fields may be unknown, ensuring robust comparison for complete state verification.
	FullyMatch(plan T, state T) bool
	// PartialMatch compares fields that cannot be changed by update methods.
	// It is designed to handle situations with unknown fields, focusing on attributes that must remain consistent.
	PartialMatch(plan T, state T) bool
	// GetChanges determines the differences between two host models.
	// If immutable fields have been changed, it returns an error.
	// If there are no changes, it returns nil; otherwise, it gathers an UpdateSpec to send to the API.
	GetChanges(plan T, state T) (*UpdateSpec, diag.Diagnostics)
	// ConvertToProto converts a host model to a specification suitable for host creation.
	// This method transforms the terraform model type to the `ProtoHostSpec` type used in API operations.
	ConvertToProto(T) ProtoHostSpec
	// ConvertFromProto converts an API-returned structure to the host model.
	// This method is essential for integrating responses into the terraform state model.
	ConvertFromProto(ProtoHost) T
}

// CreateClusterHosts is a method called during the host creation process.
// It converts hosts from their model representation to specifications suitable for use with the createHosts method.
// Finally, return the list of host specifications ready for creation and any diagnostic information collected.
func CreateClusterHosts[T Host, H any, HS any, U any](ctx context.Context,
	utilsHostService CmpHostService[T, H, HS, U],
	hostSpecs basetypes.MapValue) ([]HS, diag.Diagnostics) {
	var diags diag.Diagnostics
	var hostSpecsSlice []HS

	hostSpecsMap := make(map[string]T)
	diags.Append(hostSpecs.ElementsAs(ctx, &hostSpecsMap, false)...)
	if diags.HasError() {
		return nil, diags
	}

	for _, spec := range hostSpecsMap {
		hostSpecsSlice = append(hostSpecsSlice, utilsHostService.ConvertToProto(spec))
	}

	return hostSpecsSlice, diags
}

// UpdateClusterHostsWithShards Method to update hosts and shards within a cluster
// 1) Substitute labels using ModifyStateDependsPlan
//   - Utilize ModifyStateDependsPlan to adjust labels in the state to match those intended in the plan.
//   - This aims to prevent redundant host creation/deletion by better aligning the current state with the plan.
//
// 2) Calculate changes for Hosts
//   - Determine which hosts need to be created, updated, or deleted by comparing current state with the desired plan.
//
// 3) Calculate changes for Shards
//   - Analyze differences in shard configurations between the current state and desired plan to decide which shards require creation or deletion.
//
// 4) Clean up host changes based on shard changes
//   - Adjust host change operations to reflect shard changes accurately, avoiding unnecessary actions. For example, if a shard is being deleted, associated hosts will also be removed.
//
// 5) Create new shards
// 6) Create remaining hosts
// 7) Update existing hosts
// 8) Delete shards
// 9) Delete remaining hosts
func UpdateClusterHostsWithShards[T HostWithShard, H any, HS ProtoHostWithShard, U any](
	ctx context.Context,
	sdk *ycsdk.SDK,
	diagnostics *diag.Diagnostics,
	utilsHostService CmpHostService[T, H, HS, U],
	hostsApiService HostApiServiceWithShards[H, HS, U],
	cid string,
	plan, state types.Map,
) {
	entityIdToPlanHost := make(map[string]T)
	diagnostics.Append(plan.ElementsAs(ctx, &entityIdToPlanHost, false)...)
	if diagnostics.HasError() {
		return
	}
	entityIdToApiHosts := make(map[string]T)
	diagnostics.Append(state.ElementsAs(ctx, &entityIdToApiHosts, false)...)
	if diagnostics.HasError() {
		return
	}
	entityIdToApiHosts = ModifyStateDependsPlan(utilsHostService, entityIdToPlanHost, entityIdToApiHosts)

	toCreateShards, toDeleteShards, diags := shardsDiff(entityIdToPlanHost, entityIdToApiHosts)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return
	}

	toCreate, toUpdate, toDelete, diags := HostsDiff(utilsHostService, entityIdToPlanHost, entityIdToApiHosts)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return
	}
	toCreate, toDelete = deleteHostsDependsOnShards(toCreate, toDelete, toCreateShards, toDeleteShards)
	tflog.Debug(ctx, "shards operations will be processed", map[string]interface{}{
		"created": len(toCreateShards),
		"deleted": len(toDeleteShards),
	})

	tflog.Debug(ctx, "host operations will be processed", map[string]interface{}{
		"created": len(toCreate),
		"updated": len(toUpdate),
		"deleted": len(toDelete),
	})

	for shardName, hosts := range toCreateShards {
		var specs []HS
		for _, host := range hosts {
			specs = append(specs, utilsHostService.ConvertToProto(host))
		}
		hostsApiService.CreateShard(ctx, sdk, diagnostics, cid, shardName, specs)
		if diagnostics.HasError() {
			return
		}
	}

	hostsApiService.CreateHosts(ctx, sdk, diagnostics, cid, toCreate)
	if diagnostics.HasError() {
		return
	}

	hostsApiService.UpdateHosts(ctx, sdk, diagnostics, cid, toUpdate)
	if diagnostics.HasError() {
		return
	}

	for shardName := range toDeleteShards {
		hostsApiService.DeleteShard(ctx, sdk, diagnostics, cid, shardName)
		if diagnostics.HasError() {
			return
		}
	}
	hostsApiService.DeleteHosts(ctx, sdk, diagnostics, cid, toDelete)
	if diagnostics.HasError() {
		return
	}
}

// UpdateClusterHosts Method to update hosts within a cluster
// 1) Substitute labels using ModifyStateDependsPlan
//   - Utilize ModifyStateDependsPlan to adjust labels in the state to match those intended in the plan.
//   - This aims to prevent redundant host creation/deletion by better aligning the current state with the plan.
//
// 2) Calculate changes for Hosts
//   - Determine which hosts need to be created, updated, or deleted by comparing current state with the desired plan.
//
// 3) Create remaining hosts
// 4) Update existing hosts
// 5) Delete remaining hosts
func UpdateClusterHosts[T Host, H any, HS any, U any](
	ctx context.Context,
	sdk *ycsdk.SDK,
	diagnostics *diag.Diagnostics,
	utilsHostService CmpHostService[T, H, HS, U],
	hostsApiService HostApiService[H, HS, U],
	cid string,
	plan, state types.Map,
) {
	entityIdToPlanHost := make(map[string]T)
	diagnostics.Append(plan.ElementsAs(ctx, &entityIdToPlanHost, false)...)
	if diagnostics.HasError() {
		return
	}
	entityIdToApiHosts := make(map[string]T)
	diagnostics.Append(state.ElementsAs(ctx, &entityIdToApiHosts, false)...)
	if diagnostics.HasError() {
		return
	}
	entityIdToApiHosts = ModifyStateDependsPlan(utilsHostService, entityIdToPlanHost, entityIdToApiHosts)

	toCreate, toUpdate, toDelete, diags := HostsDiff(utilsHostService, entityIdToPlanHost, entityIdToApiHosts)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "host operations will be processed", map[string]interface{}{
		"created": len(toCreate),
		"updated": len(toUpdate),
		"deleted": len(toDelete),
	})

	hostsApiService.CreateHosts(ctx, sdk, diagnostics, cid, toCreate)
	if diagnostics.HasError() {
		return
	}

	hostsApiService.UpdateHosts(ctx, sdk, diagnostics, cid, toUpdate)
	if diagnostics.HasError() {
		return
	}

	hostsApiService.DeleteHosts(ctx, sdk, diagnostics, cid, toDelete)
	if diagnostics.HasError() {
		return
	}
}

// Processes the collected changes for hosts and shards.
// If a shard is being deleted, all hosts within it will be deleted automatically,
// so there is no need to delete them separately.
// Similarly, if a shard is being created, all hosts within it will be created automatically,
// so there is no need to create them separately.
// Returns a filtered list of hosts to create and hosts to delete.
func deleteHostsDependsOnShards[T HostWithShard, HS ProtoHostWithShard](
	toCreateHosts []HS,
	toDeleteHosts []string,
	toCreateShards map[string][]T,
	toDeleteShards map[string][]T,
) ([]HS, []string) {
	resD := slices.DeleteFunc(toDeleteHosts, func(fqdn string) bool {
		for _, hosts := range toDeleteShards {
			for _, host := range hosts {
				if host.GetFQDN().ValueString() == fqdn {
					return true
				}
			}
		}
		return false
	})

	resC := slices.DeleteFunc(toCreateHosts, func(spec HS) bool {
		for shardName := range toCreateShards {
			if spec.GetShardName() == shardName {
				return true
			}
		}
		return false
	})
	return resC, resD
}

// The shardsDiff function categorizes shards for further actions:
// create (a new shard needs to be created),
// skip (the shard has not changed and is skipped),
// delete (the shard is to be removed).
// The logic is quite simple and processes an empty shard only if there is just one shard in the cluster:
// 1) If shardName exists in both the plan and state: skip (no actions are needed for these shards).
// 2) If shardName is in the plan but not in the state, it indicates a new shard will be added.
// 3) If shardName is not in the plan but is in the state, it indicates the shard will be removed.
func shardsDiff[T HostWithShard](planHosts, stateHosts map[string]T) (map[string][]T, map[string][]T, diag.Diagnostics) {
	var diags diag.Diagnostics
	allShards := make(map[string]struct{})
	toCreateShards := make(map[string][]T)
	for _, p := range planHosts {
		toCreateShards[p.GetShard()] = append(toCreateShards[p.GetShard()], p)
		allShards[p.GetShard()] = struct{}{}
	}

	toDeleteShards := make(map[string][]T)
	for _, s := range stateHosts {
		toDeleteShards[s.GetShard()] = append(toDeleteShards[s.GetShard()], s)
		allShards[s.GetShard()] = struct{}{}
	}
	if _, ok := allShards[""]; ok {
		if len(allShards) <= 2 {
			// only one shard in this case
			return nil, nil, nil
		}
		if len(allShards) > 2 {
			diags.AddError(
				"Wrong Host Configuration",
				"Unexpected empty shard name for multisharded cluster",
			)
			return nil, nil, diags
		}
	}

	for k := range toCreateShards {
		if _, ok := toDeleteShards[k]; ok {
			delete(toDeleteShards, k)
			delete(toCreateShards, k)
		}
	}

	return toCreateShards, toDeleteShards, nil
}

// The HostsDiff function categorizes hosts into further actions:
// create (a new host needs to be created),
// skip (the host has not changed and is skipped),
// update (the host has modifiable parameters that have changed),
// delete (the host is to be removed).
// The logic is quite simple:
// 1) If a label exists in both the plan and state, the host either will be changed or remained the same, determined by calling getChanges.
// 2) If a label is in the plan but not in the state, a new host will be added.
// 3) If a label is not in the plan but is in the state, the host will be deleted.
// Returns slices for actions: create, update, delete, and diagnostics capturing potential errors or warnings.
func HostsDiff[T Host, H any, HS any, U any](
	hostService CmpHostService[T, H, HS, U], // A service utility to help with determining host state changes using comparison methods.
	planHosts map[string]T, // A map representing the planned host state, containing labels and their corresponding host objects.
	stateHosts map[string]T, // A map representing the current host state, with labels associated with their respective host objects.
) ([]HS, []*U, []string, diag.Diagnostics) {
	var toCreate []HS
	var toUpdate []*U
	var toDelete []string
	for label, planHost := range planHosts {
		if stateHost, exist := stateHosts[label]; !exist {
			toCreate = append(toCreate, hostService.ConvertToProto(planHost))
		} else {
			changes, diags := hostService.GetChanges(planHost, stateHost)
			if diags.HasError() {
				return nil, nil, nil, diags
			}
			if changes != nil {
				toUpdate = append(toUpdate, changes)
			}
		}
	}

	for label, s := range stateHosts {
		if _, ok := planHosts[label]; !ok {
			toDelete = append(toDelete, s.GetFQDN().ValueString())
		}
	}

	return toCreate, toUpdate, toDelete, nil
}

// The ModifyStateDependsPlan method collapses hosts with different labels in the `plan` and `state`
// to avoid unnecessary create/delete operations. The logic for forming `newState` involves the following actions:
//   - skip (the host has not changed and remains as is),
//   - update (the host has modifiable parameters that changed),
//   - delete (the host is to be removed).
//
// Sorting these actions will be handled in another method, `HostsDiff`; the current method only makes assumptions.
//
//  1. If a label exists in both the `plan` and `state`, it is left unchanged and copied to `newState`.
//     For such hosts, subsequent actions can be: skip, update.
//
//  2. For all other hosts in the `state`, find a new corresponding host in the `plan` that fully matches the fields (using the FullyMatch method).
//     For such a host, use the label from the `plan` in `newState`. As a result, the action will be: skip.
//
//  3. For the remaining hosts in the `state`, find a new corresponding host in the `plan` that partially matches the fields (using the PartialMatch method).
//     For such a host, use the label from the `plan` in `newState`. As a result, the action will be: update.
//
// 4. The remaining hosts from the `state` are moved to `newState` with their labels. For such hosts, the action will be: delete.
// Returns `newState`
func ModifyStateDependsPlan[T Host, H any, HS any, U any](
	hostService CmpHostService[T, H, HS, U],
	plan map[string]T,
	state map[string]T,
) map[string]T {
	fixedState := make(map[string]T)
	usedLabels := make(map[string]struct{})
	for label := range state {
		if _, ok := plan[label]; ok {
			fixedState[label] = state[label]
			usedLabels[label] = struct{}{}
		}
	}

	//fully match
	for label, stateHost := range state {
		for planLabel, planHost := range plan {
			_, okState := fixedState[planLabel]
			_, okPlan := usedLabels[label]
			if okState || okPlan {
				continue
			}
			if hostService.FullyMatch(planHost, stateHost) {
				fixedState[planLabel] = stateHost
				usedLabels[label] = struct{}{}
			}
		}
	}

	//partial match
	for label, stateHost := range state {
		for planLabel, planHost := range plan {
			_, okState := fixedState[planLabel]
			_, okPlan := usedLabels[label]
			if okState || okPlan {
				continue
			}
			if hostService.PartialMatch(planHost, stateHost) {
				fixedState[planLabel] = stateHost
				usedLabels[label] = struct{}{}
			}
		}
	}

	//to delete
	for label, stateHost := range state {
		if _, ok := usedLabels[label]; !ok {
			fixedState[label] = stateHost
		}
	}

	return fixedState
}

// The ReadHosts method is used to update the state of hosts.
//
// 1. Map hosts from api to hosts in state by fqdn
// 2. Map hosts from api to hosts in state without fqdn by equal attributes
// 3. Add hosts from api to state if not mapped
func ReadHosts[T Host, H ProtoHost, HS any, U any](
	ctx context.Context,
	sdk *ycsdk.SDK, // The SDK instance to interact with the relevant API.
	diags *diag.Diagnostics,
	utilsHostService CmpHostService[T, H, HS, U], // A service utility for comparative host operations involving generic types.
	hostsApiService HostApiService[H, HS, U], // The API service used to manage hosts, facilitating operations on host data.
	stateHosts basetypes.MapValue, // A map representing the state of hosts, used for validation and updates.
	cid string,
) map[string]T {
	planHostsMap := make(map[string]T)

	if utils.IsPresent(stateHosts) {
		diags.Append(stateHosts.ElementsAs(ctx, &planHostsMap, false)...)
		if diags.HasError() {
			return nil
		}
	}

	apiHosts := hostsApiService.ListHosts(ctx, sdk, diags, cid)
	if diags.HasError() {
		return nil
	}

	fqdnToApiHost := make(map[string]H)
	for _, host := range apiHosts {
		fqdnToApiHost[host.GetName()] = host
	}

	stateFqdns := make(map[string]struct{})
	for _, host := range planHostsMap {
		if v := host.GetFQDN().ValueString(); v != "" {
			stateFqdns[v] = struct{}{}
		}
	}

	entityIdToApiHosts := make(map[string]T)

	// For each host in the state that has an FQDN,
	// we correlate the host from the API. At the same time,
	// we ignore the inconsistent state in the case of Update/Create (tf framework already knows how to do this)
	for hostLabel, host := range planHostsMap {
		fqdn := host.GetFQDN().ValueString()
		if fqdn == "" {
			continue
		}
		if apiHost, ok := fqdnToApiHost[fqdn]; ok {
			// case when fqdn exist in API and STATE before apply
			// We need to add apiHost to our map under the appropriate label
			entityIdToApiHosts[hostLabel] = utilsHostService.ConvertFromProto(apiHost)
		}
	}

	// For each host without FQDN, we correlate the host from the API response with the same attribute values.
	// This situation can only occur with a Create/Update refresh.
	for hostLabel, host := range planHostsMap {
		fqdn := host.GetFQDN().ValueString()
		if fqdn != "" {
			continue
		}
		for _, apiHost := range apiHosts {
			if _, ok := stateFqdns[apiHost.GetName()]; ok {
				continue
			}
			if h := utilsHostService.ConvertFromProto(apiHost); utilsHostService.FullyMatch(host, h) {
				stateFqdns[apiHost.GetName()] = struct{}{}
				entityIdToApiHosts[hostLabel] = h
				break
			}
		}
		if _, ok := entityIdToApiHosts[hostLabel]; !ok {
			diags.AddError(
				"Hosts state is wrong after change",
				fmt.Sprintf("Expected host for label %q, but not was found. This is a problem with the provider", hostLabel),
			)
			return nil
		}
	}

	// The other hosts are simply entered into the state.
	for _, apiHost := range apiHosts {
		if _, ok := stateFqdns[apiHost.GetName()]; !ok {
			entityIdToApiHosts[apiHost.GetName()] = utilsHostService.ConvertFromProto(apiHost)
		}
	}

	return entityIdToApiHosts
}
