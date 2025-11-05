package metastore_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	ycsdk "github.com/yandex-cloud/go-sdk"

	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &metastoreClusterResource{}
var _ resource.ResourceWithImportState = &metastoreClusterResource{}
var _ resource.ResourceWithValidateConfig = &metastoreClusterResource{}

func NewResource() resource.Resource {
	return &metastoreClusterResource{}
}

type metastoreClusterResource struct {
	providerConfig *provider_config.Config
}

// Metadata implements resource.Resource.
func (r *metastoreClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metastore_cluster"
}

// Configure implements resource.Resource.
func (r *metastoreClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState implements resource.ResourceWithImportState.
func (r *metastoreClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create implements resource.Resource.
func (r *metastoreClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClusterModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createClusterRequest, diags := BuildCreateClusterRequest(ctx, &plan, &r.providerConfig.ProviderState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Create Metastore cluster request: %+v", createClusterRequest))

	createTimeout, diags := plan.Timeouts.Create(ctx, YandexMetastoreClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	clusterID, d := CreateCluster(ctx, r.providerConfig.SDK, &resp.Diagnostics, createClusterRequest)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(clusterID)
	refreshState(ctx, r.providerConfig.SDK, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Finished creating Metastore cluster", clusterIDLogField(clusterID))
}

// Delete implements resource.Resource.
func (r *metastoreClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClusterModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting Metastore cluster", clusterIDLogField(clusterID))

	deleteTimeout, diags := state.Timeouts.Delete(ctx, YandexMetastoreClusterDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	d := DeleteCluster(ctx, r.providerConfig.SDK, clusterID)
	resp.Diagnostics.Append(d)

	tflog.Debug(ctx, "Finished deleting Metastore cluster", clusterIDLogField(clusterID))
}

// Read implements resource.Resource.
func (r *metastoreClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClusterModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.Id.ValueString()
	tflog.Debug(ctx, "Reading Metastore cluster", clusterIDLogField(clusterID))
	cluster, d := GetClusterByID(ctx, r.providerConfig.SDK, clusterID)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	if cluster == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	refreshState(ctx, r.providerConfig.SDK, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished reading Metastore cluster", clusterIDLogField(clusterID))
}

// Update implements resource.Resource.
func (r *metastoreClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ClusterModel
	var state ClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Metastore cluster", clusterIDLogField(state.Id.ValueString()))

	updateTimeout, diags := plan.Timeouts.Update(ctx, YandexMetastoreClusterUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Debug(ctx, fmt.Sprintf("Update Metastore cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update Metastore cluster plan: %+v", plan))

	updateReq, diags := BuildUpdateClusterRequest(ctx, &state, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Update Metastore cluster request: %+v", updateReq))

	d := UpdateCluster(ctx, r.providerConfig.SDK, updateReq)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshState(ctx, r.providerConfig.SDK, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "Finished updating Metastore cluster", clusterIDLogField(state.Id.ValueString()))
}

// Schema implements resource.Resource.
func (r *metastoreClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ClusterResourceSchema(ctx)
	resp.Schema.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})
}

func (r *metastoreClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cluster ClusterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cluster)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func refreshState(ctx context.Context, sdk *ycsdk.SDK, state *ClusterModel, diags *diag.Diagnostics) {
	clusterID := state.Id.ValueString()
	tflog.Debug(ctx, "Reading Metastore cluster", clusterIDLogField(clusterID))
	cluster, d := GetClusterByID(ctx, sdk, clusterID)
	diags.Append(d)
	if diags.HasError() {
		return
	}

	if cluster == nil {
		diags.AddError(
			"Metastore cluster not found",
			fmt.Sprintf("Metastore cluster with id %s not found", clusterID))
		return
	}

	state.ClusterConfig = ClusterConfigValue{
		ResourcePresetId: types.StringValue(cluster.GetClusterConfig().GetResources().GetResourcePresetId()),
		state:            attr.ValueStateKnown,
	}
	state.CreatedAt = types.StringValue(cluster.GetCreatedAt().String())
	state.DeletionProtection = types.BoolValue(cluster.GetDeletionProtection())

	description := types.StringValue(cluster.GetDescription())
	if !stringsAreEqual(state.Description, description) {
		state.Description = description
	}

	state.EndpointIp = types.StringValue(cluster.GetEndpointIp())
	state.FolderId = types.StringValue(cluster.GetFolderId())

	labels := flattenStringMap(ctx, cluster.GetLabels(), diags)
	if !mapsAreEqual(state.Labels, labels) {
		state.Labels = labels
	}

	logging := flattenLoggingConfig(cluster.GetLogging(), diags)
	if !loggingValuesAreEqual(state.Logging, logging) {
		state.Logging = logging
	}

	state.MaintenanceWindow = flattenMaintenanceWindow(cluster.GetMaintenanceWindow(), diags)
	state.Name = types.StringValue(cluster.GetName())
	state.NetworkId = types.StringValue(cluster.GetNetworkId())

	securityGroupIDs := flattenStringSlice(ctx, cluster.GetNetwork().GetSecurityGroupIds(), diags)
	if !setsAreEqual(state.SecurityGroupIds, securityGroupIDs) {
		state.SecurityGroupIds = securityGroupIDs
	}

	state.ServiceAccountId = types.StringValue(cluster.GetServiceAccountId())
	state.Status = types.StringValue(cluster.GetStatus().String())

	subnetIDs := flattenStringSlice(ctx, cluster.GetNetwork().GetSubnetIds(), diags)
	if !setsAreEqual(state.SubnetIds, subnetIDs) {
		state.SubnetIds = subnetIDs
	}

	state.Version = types.StringValue(cluster.GetVersion())
}

func clusterIDLogField(cid string) map[string]interface{} {
	return map[string]interface{}{
		"cluster_id": cid,
	}
}
