package mdb_postgresql_cluster_beta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Map `terraformEntityId` -> postgresql.HostSpec
func hostsFromMapValue(ctx context.Context, stateHosts types.Map) (map[string]Host, diag.Diagnostics) {
	hostsType := make(map[string]Host)
	diags := stateHosts.ElementsAs(ctx, &hostsType, false)
	return hostsType, diags
}

// Returns lists to create, update and delete hosts.
// stateHosts:  terraform id -> host
// apiHosts: terraform id -> host
func hostsDiff(planHosts map[string]Host, apiHosts map[string]Host) (toCreate map[string]*postgresql.HostSpec, toUpdate []*postgresql.UpdateHostSpec, toDelete []string) {
	toCreate = make(map[string]*postgresql.HostSpec)
	toUpdate = []*postgresql.UpdateHostSpec{}
	toDelete = []string{}

	// Lets run on planned hosts first
	for tfid, p := range planHosts {
		if apiHost, exist := apiHosts[tfid]; !exist {
			// If there is no such host in the API, we need to create it
			toCreate[tfid] = &postgresql.HostSpec{
				ZoneId:            p.Zone.ValueString(),
				SubnetId:          p.SubnetId.ValueString(),
				AssignPublicIp:    p.AssignPublicIp.ValueBool(),
				ReplicationSource: p.ReplicationSource.ValueString(),
			}
		} else {
			if p.AssignPublicIp != apiHost.AssignPublicIp || p.ReplicationSource != apiHost.ReplicationSource {
				// Host is different, we need to update it
				if p.FQDN.IsNull() || p.FQDN.IsUnknown() {
					panic("host name is not supposed to be empty")
				}
				toUpdate = append(toUpdate, &postgresql.UpdateHostSpec{
					HostName: p.FQDN.ValueString(),
					UpdateMask: &fieldmaskpb.FieldMask{
						Paths: []string{"assign_public_ip", "replication_source"},
					},
					AssignPublicIp:    p.AssignPublicIp.ValueBool(),
					ReplicationSource: p.ReplicationSource.ValueString(),
				})
			}
		}
	}

	// Lets iterate over api hosts
	// And find non-existence plan hosts
	// We need to delete them
	for tfid, s := range apiHosts {
		if s.FQDN.IsNull() || s.FQDN.IsUnknown() {
			panic("api host name is not supposed to be empty")
		}

		// If there is no such host in the plan, we need to delete it
		if _, ok := planHosts[tfid]; !ok {
			toDelete = append(toDelete, s.FQDN.ValueString())
		}
	}

	return
}

// hostsSquats is a helper struct to work with hosts
// prisedaniia s hostami
type hostsSquats struct {
	mapping      map[hostEntity][]string
	requestHosts []*postgresql.HostSpec
}

type hostEntity struct {
	Zone           string
	SubnetId       string
	assingPublicIp bool
}

// Step1 of the Hosts Squats. Build hosts for api request and save hosts mapping
func (h *hostsSquats) Step1(ctx context.Context, hostSpecs basetypes.MapValue) ([]*postgresql.HostSpec, diag.Diagnostics) {
	var diag diag.Diagnostics
	hostSpecsSlice := []*postgresql.HostSpec{}
	mapping := map[hostEntity][]string{} // We will use this slice later to initialize the state

	hostSpecsMap, diags := hostsFromMapValue(ctx, hostSpecs) // Map `terraformEntityId` -> postgresql.HostSpec
	diag.Append(diags...)
	if diags.HasError() {
		return nil, diags
	}

	for tfid, spec := range hostSpecsMap {
		hostSpecsSlice = append(hostSpecsSlice, &postgresql.HostSpec{
			ZoneId:            spec.Zone.ValueString(),
			SubnetId:          spec.SubnetId.ValueString(),
			AssignPublicIp:    spec.AssignPublicIp.ValueBool(),
			ReplicationSource: spec.ReplicationSource.ValueString(),
		})
		key := hostEntity{spec.Zone.ValueString(), spec.SubnetId.ValueString(), spec.AssignPublicIp.ValueBool()}
		mapping[key] = append(mapping[key], tfid)
	}
	h.mapping = mapping
	h.requestHosts = hostSpecsSlice

	return hostSpecsSlice, diag
}

// Step2 of the Hosts Squats. Map hosts from the API response to the terraform entity id
func (h *hostsSquats) Step2(ctx context.Context, sdk *ycsdk.SDK, cid string) (map[string]Host, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Let's list hosts
	apiHosts, err := listHosts(ctx, sdk, &diags, cid)
	if err != nil {
		diags.AddError(
			"Failed to List PostgreSQL Hosts",
			"Error while requesting API to get PostgreSQL host:"+err.Error(),
		)
		return map[string]Host{}, diags
	}

	if len(apiHosts) != len(h.requestHosts) {
		diags.AddError(
			"Failed to Create resource",
			"Number of hosts in the state and in the API are different",
		)
		return map[string]Host{}, diags
	}

	// The order of the hosts does not remains the same
	hosts := make(map[string]Host)
	for _, apiHost := range apiHosts {
		key := hostEntity{apiHost.ZoneId, apiHost.SubnetId, apiHost.AssignPublicIp}
		tfids, ok := h.mapping[key]
		if !ok {
			diags.AddError(
				"Failed to Create resource",
				"Host mapping(tfids) is empty. This is a problem with the provider",
			)
			return map[string]Host{}, diags
		}

		v := tfids[len(tfids)-1]
		h.mapping[key] = tfids[:len(tfids)-1]
		hosts[v] = Host{
			Zone:              types.StringValue(apiHost.ZoneId),
			SubnetId:          types.StringValue(apiHost.SubnetId),
			AssignPublicIp:    types.BoolValue(apiHost.AssignPublicIp),
			ReplicationSource: types.StringValue(apiHost.ReplicationSource),
			FQDN:              types.StringValue(apiHost.Name),
		}
	}

	return hosts, diags
}
