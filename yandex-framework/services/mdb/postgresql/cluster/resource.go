package cluster

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/utils"
)

type clusterResource struct {
	providerConfig *provider_config.Config
}

func NewPostgreSQLClusterResourceBeta() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	// TODO we are open for better ideas
	resp.TypeName = req.ProviderTypeName + "_mdb_postgresql_cluster_beta"
}

func (r *clusterResource) Configure(_ context.Context,
	req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

func (r *clusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"folder_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the folder that the cluster belongs to.",
			},
			"network_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the network that the cluster belongs to.",
			},
			"environment": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Deployment environment of the cluster.",
			},
			"labels": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Custom labels for the cluster as key-value pairs.",
			},
			"hosts": schema.MapNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone": schema.StringAttribute{
							Required:    true,
							Description: "Availability zone where the host is located.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"subnet_id": schema.StringAttribute{
							Optional:    true,
							Description: "ID of the subnet where the host is located.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"assign_public_ip": schema.BoolAttribute{
							Optional:    true,
							Description: "Assign a public IP address to the host.",
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"fqdn": schema.StringAttribute{
							Computed:    true,
							Description: "Fully Qualified Domain Name. In other words, hostname.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"replication_source": schema.StringAttribute{
							Optional:    true,
							Description: "FQDN of the host that is used as a replication source.",
							Computed:    true,
							Default:     stringdefault.StaticString(""),
						},
					},
				},
			},
			"deletion_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Inhibits deletion of the cluster. Can be either true or false.",
			},
		},
		Blocks: map[string]schema.Block{
			"config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Required:    true,
						Description: "Version of the PostgreSQL cluster",
					},
					"autofailover": schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
				},
				Blocks: map[string]schema.Block{
					"resources": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"resource_preset_id": schema.StringAttribute{
								Required:    true,
								Description: "ID of the resource preset that determines the number of CPU cores and memory size for the host.",
							},
							"disk_type_id": schema.StringAttribute{
								Required:    true,
								Description: "ID of the disk type that determines the disk performance characteristics.",
							},
							"disk_size": schema.Int64Attribute{
								Required:    true,
								Description: "Size of the disk in bytes.",
							},
						},
					},
				},
			},
		},
	}
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Load the current state of the resource
	var state Cluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Here we retrieve map[terraform entity id] -> postgresql.HostSpec
	// We use only state here because in Read method there is not plan
	stateHostsMap, diags := hostsFromMapValue(ctx, state.HostSpecs)
	diags.Append(diags...)
	if diags.HasError() {
		return
	}

	// List API hosts
	cid := state.Id.ValueString()
	apiHosts, err := listHosts(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid)
	if err != nil {
		diags.AddError(
			"Failed to List PostgreSQL Hosts",
			"Error while requesting API to get PostgreSQL host:"+err.Error(),
		)
		return
	}
	fqdnToApiHost := make(map[string]*postgresql.Host)
	for _, host := range apiHosts {
		// We are always sure host.Name is not empty
		// since it is Read method
		fqdnToApiHost[host.Name] = host
	}

	// Construct a map of 'terraform entity id' -> API Host
	entityIdToApiHosts := make(map[string]Host)
	for entityID, host := range stateHostsMap {
		// We are always sure host.FQDN is not empty
		// since it is Read method
		apiHost, ok := fqdnToApiHost[host.FQDN.ValueString()]
		if !ok {
			// If you see this, this host exists in api and does not exist in state
			// Maybe it would be cool try to map this host to a not created terraform host. Maybe are they the same?
			// What if not to run any operation and just update the state?
			entityIdToApiHosts[fmt.Sprintf("host-to-drop-%d", rand.Intn(10000))] = Host{FQDN: host.FQDN}
		} else {
			entityIdToApiHosts[entityID] = Host{
				Zone:              types.StringValue(apiHost.ZoneId),
				SubnetId:          types.StringValue(apiHost.SubnetId),
				AssignPublicIp:    types.BoolValue(apiHost.AssignPublicIp),
				ReplicationSource: types.StringValue(apiHost.ReplicationSource),
				FQDN:              types.StringValue(apiHost.Name),
			}
		}
	}

	// Continue constructing a map of 'terraform entity id' -> API Host
	for _, apiHost := range apiHosts {
		apiHostExistInState := false
		for _, stateHost := range stateHostsMap {
			if apiHost.Name == stateHost.FQDN.ValueString() {
				// we found the host in the state
				apiHostExistInState = true
				continue
			}
		}
		if !apiHostExistInState {
			entityIdToApiHosts[fmt.Sprintf("host-to-drop-%s", apiHost.Name)] = Host{
				Zone:              types.StringValue(apiHost.ZoneId),
				SubnetId:          types.StringValue(apiHost.SubnetId),
				AssignPublicIp:    types.BoolValue(apiHost.AssignPublicIp),
				ReplicationSource: types.StringValue(apiHost.ReplicationSource),
				FQDN:              types.StringValue(apiHost.Name),
			}
		}
	}

	// Update Resource State
	r.refreshResourceState(ctx, &state, entityIdToApiHosts, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	d := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(d...)
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Cluster
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating PostgreSQL Cluster")

	hostsSquats := hostsSquats{}
	// Step 1 of the Hosts Squats. Build hosts for api request and save hosts mapping
	hostSpecsSlice, diags := hostsSquats.Step1(ctx, plan.HostSpecs)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Prepare Create Request
	request, diags := prepareCreateRequest(ctx, &plan, &r.providerConfig.ProviderState)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// Add Hosts to the request
	request.HostSpecs = hostSpecsSlice

	cid := createCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, request)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 2 of the Hosts Squats. Map hosts from the API response to the terraform entity id
	hosts, diags := hostsSquats.Step2(ctx, r.providerConfig.SDK, cid)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	plan.Id = types.StringValue(cid)
	r.refreshResourceState(ctx, &plan, hosts, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Load the current plan
	// We shouldnt read the state because we shouldn't use the state in the host update method
	// The plan and the Api response should be enough
	var plan Cluster
	var state Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating PostgreSQL Cluster", map[string]interface{}{"id": plan.Id.ValueString()})
	tflog.Debug(ctx, fmt.Sprintf("Update PostgreSQL Cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update PostgreSQL Cluster plan: %+v", plan))

	updateRequest, d := prepareUpdateRequest(ctx, &state, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateRequest)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map 'terraform entity id' -> host(fqdn)
	entityIdToPlanHost, diags := hostsFromMapValue(ctx, plan.HostSpecs)
	// panic(fmt.Sprintf("entityIdToPlanHost: %v\nplan.HostsSpecs: %v\n", entityIdToPlanHost, plan.HostSpecs))

	diags.Append(diags...)
	if diags.HasError() {
		return
	}

	// List API hosts
	cid := plan.Id.ValueString()
	apiHosts, err := listHosts(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid)
	if err != nil {
		diags.AddError(
			"Failed to List PostgreSQL Hosts",
			"Error while requesting API to get PostgreSQL host:"+err.Error(),
		)
		return
	}
	fqdnToApiHost := make(map[string]*postgresql.Host)
	for _, host := range apiHosts {
		fqdnToApiHost[host.Name] = host // like fqdn -> host
	}

	// Construct a map of 'terraform entity id' -> Existening API Host
	// First, iterate over the plan hosts
	entityIdToApiHosts := make(map[string]Host) // 'terraform entity id' -> postgresql.Host
	for entityID, host := range entityIdToPlanHost {
		// If it is a plan then it is possible there is no host.Name yet.
		// It is the new host that is going to be created.
		// Skip it.
		if host.FQDN.IsNull() || host.FQDN.IsUnknown() {
			continue
		}

		// host.Name is not empty. It may exist in api or not
		apiHost, exist := fqdnToApiHost[host.FQDN.ValueString()]
		if exist {
			// Host exists in both plan and api
			if apiHost == nil {
				panic("apiHost is not supposed to be nil")
			}
			entityIdToApiHosts[entityID] = Host{
				Zone:              types.StringValue(apiHost.ZoneId),
				SubnetId:          types.StringValue(apiHost.SubnetId),
				AssignPublicIp:    types.BoolValue(apiHost.AssignPublicIp),
				ReplicationSource: types.StringValue(apiHost.ReplicationSource),
				FQDN:              types.StringValue(apiHost.Name),
			}

		} else {
			// Host exists in plan but not in API
			continue
		}
	}

	// Continue constructing a map of 'terraform entity id' -> API Host
	for _, apiHost := range apiHosts {
		// Lets find all hosts that exists in api but not exist in plan
		apiHostExistInState := false
		for _, planHost := range entityIdToPlanHost {
			if apiHost.Name == planHost.FQDN.ValueString() {
				// we found the host in the state
				apiHostExistInState = true
				continue
			}
		}

		// Save this "extra" host to build proper host diff
		if !apiHostExistInState {
			entityIdToApiHosts[fmt.Sprintf("host-to-drop-%s", apiHost.Name)] = Host{
				Zone:              types.StringValue(apiHost.ZoneId),
				SubnetId:          types.StringValue(apiHost.SubnetId),
				AssignPublicIp:    types.BoolValue(apiHost.AssignPublicIp),
				ReplicationSource: types.StringValue(apiHost.ReplicationSource),
				FQDN:              types.StringValue(apiHost.Name),
			}
		}
	}

	// TODO here we should convert list toCreate toUpdate and toDelete to the list of operations
	// To minimize the amount of resources used at the moment

	// Lets update hosts
	toCreate, toUpdate, toDelete := hostsDiff(entityIdToPlanHost, entityIdToApiHosts)

	for tfid, hostSpec := range toCreate {
		metadata, diag := addHost(ctx, r.providerConfig.SDK, cid, hostSpec)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(metadata.HostNames) == 0 {
			panic("metadata.HostNames is not supposed to be empty")
		}

		host := entityIdToPlanHost[tfid]
		host.FQDN = types.StringValue(metadata.HostNames[0])
		entityIdToPlanHost[tfid] = host
	}

	for _, hostSpec := range toUpdate {
		updateHost(ctx, r.providerConfig.SDK, &diags, cid, hostSpec)
	}

	for _, hostname := range toDelete {
		deleteHost(ctx, r.providerConfig.SDK, &diags, cid, hostname)

		// Lets clean deleted host from future state
		for tfid, host := range entityIdToApiHosts {
			if host.FQDN.ValueString() == hostname {
				delete(entityIdToApiHosts, tfid)
			}
		}
	}

	r.refreshResourceState(ctx, &plan, entityIdToPlanHost, &resp.Diagnostics)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Cluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.Id.ValueString()
	deleteCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid)
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *clusterResource) refreshResourceState(ctx context.Context, state *Cluster, hosts map[string]Host, respDiagnostics *diag.Diagnostics) {
	cid := state.Id.ValueString()
	cluster := readCluster(ctx, r.providerConfig.SDK, respDiagnostics, cid)
	if respDiagnostics.HasError() {
		return
	}

	labels, diags := types.MapValueFrom(ctx, types.StringType, cluster.Labels)
	respDiagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	version := types.StringValue(cluster.Config.Version)
	resources, diags := types.ObjectValueFrom(ctx, ResourcesAttrTypes, Resources{
		ResourcePresetID: types.StringValue(cluster.Config.Resources.ResourcePresetId),
		DiskSize:         types.Int64Value(utils.ToGigabytes(cluster.Config.Resources.DiskSize)),
		DiskTypeID:       types.StringValue(cluster.Config.Resources.DiskTypeId),
	})
	autofailover := types.BoolValue(cluster.Config.GetAutofailover().GetValue())
	deletionProtection := types.BoolValue(cluster.GetDeletionProtection())

	respDiagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	config, diags := types.ObjectValueFrom(ctx, ConfigAttrTypes, Config{
		Version:      version,
		Resources:    resources,
		Autofailover: autofailover,
	})
	respDiagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	hostMapValue, diags := types.MapValueFrom(ctx, hostType, hosts)
	respDiagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	state.Id = types.StringValue(cluster.Id)
	state.FolderId = types.StringValue(cluster.FolderId)
	state.NetworkId = types.StringValue(cluster.NetworkId)
	state.Name = types.StringValue(cluster.Name)
	state.Description = types.StringValue(cluster.Description)
	state.Environment = types.StringValue(cluster.Environment.String())
	state.Labels = labels
	state.Config = config
	state.HostSpecs = hostMapValue
	state.DeletionProtection = deletionProtection
}
