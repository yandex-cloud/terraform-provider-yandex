package spark_cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	ycsdk "github.com/yandex-cloud/go-sdk"

	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	YandexSparkClusterCreateTimeout = 30 * time.Minute
	YandexSparkClusterDeleteTimeout = 15 * time.Minute
	YandexSparkClusterUpdateTimeout = 60 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &sparkClusterResource{}
var _ resource.ResourceWithImportState = &sparkClusterResource{}
var _ resource.ResourceWithValidateConfig = &sparkClusterResource{}

func NewResource() resource.Resource {
	return &sparkClusterResource{}
}

type sparkClusterResource struct {
	providerConfig *provider_config.Config
}

// Metadata implements resource.Resource.
func (a *sparkClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spark_cluster"
}

// Configure implements resource.Resource.
func (a *sparkClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	a.providerConfig = providerConfig
}

// Create implements resource.Resource.
func (a *sparkClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClusterModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createClusterRequest, diags := BuildCreateClusterRequest(ctx, &plan, &a.providerConfig.ProviderState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Create Spark cluster request: %+v", createClusterRequest))

	createTimeout, diags := plan.Timeouts.Create(ctx, YandexSparkClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	clusterID, d := CreateCluster(ctx, a.providerConfig.SDK, &resp.Diagnostics, createClusterRequest)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(clusterID)
	diags = updateState(ctx, a.providerConfig.SDK, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Finished creating Spark cluster", clusterIDLogField(clusterID))
}

// Delete implements resource.Resource.
func (a *sparkClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClusterModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting Spark cluster", clusterIDLogField(clusterID))

	deleteTimeout, diags := state.Timeouts.Delete(ctx, YandexSparkClusterDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	d := DeleteCluster(ctx, a.providerConfig.SDK, clusterID)
	resp.Diagnostics.Append(d)

	tflog.Debug(ctx, "Finished deleting Spark cluster", clusterIDLogField(clusterID))
}

// Read implements resource.Resource.
func (a *sparkClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClusterModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.Id.ValueString()
	tflog.Debug(ctx, "Reading Spark cluster", clusterIDLogField(clusterID))
	cluster, d := GetClusterByID(ctx, a.providerConfig.SDK, clusterID)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	if cluster == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = ClusterToState(ctx, cluster, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished reading Spark cluster", clusterIDLogField(clusterID))
}

// Update implements resource.Resource.
func (a *sparkClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ClusterModel
	var state ClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Spark cluster", clusterIDLogField(state.Id.ValueString()))

	updateTimeout, diags := plan.Timeouts.Update(ctx, YandexSparkClusterUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Debug(ctx, fmt.Sprintf("Update Spark cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update Spark cluster plan: %+v", plan))

	updateReq, diags := BuildUpdateClusterRequest(ctx, &state, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Update Spark cluster request: %+v", updateReq))

	d := UpdateCluster(ctx, a.providerConfig.SDK, updateReq)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = updateState(ctx, a.providerConfig.SDK, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "Finished updating Spark cluster", clusterIDLogField(state.Id.ValueString()))
}

// Schema implements resource.Resource.
func (a *sparkClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ClusterResourceSchema(ctx)
	resp.Schema.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})
}

func (r *sparkClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *sparkClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cluster ClusterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cluster)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func updateState(ctx context.Context, sdk *ycsdk.SDK, state *ClusterModel) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterID := state.Id.ValueString()
	tflog.Debug(ctx, "Reading Spark cluster", clusterIDLogField(clusterID))
	cluster, d := GetClusterByID(ctx, sdk, clusterID)
	diags.Append(d)
	if diags.HasError() {
		return diags
	}

	if cluster == nil {
		diags.AddError(
			"Spark cluster not found",
			fmt.Sprintf("Spark cluster with id %s not found", clusterID))
		return diags
	}

	dd := ClusterToState(ctx, cluster, state)
	diags.Append(dd...)
	return diags
}

func clusterIDLogField(cid string) map[string]interface{} {
	return map[string]interface{}{
		"cluster_id": cid,
	}
}
