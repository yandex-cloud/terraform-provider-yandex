package airflow_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	ycsdk "github.com/yandex-cloud/go-sdk"

	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	af_api "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/airflow_cluster/api"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &airflowClusterResource{}
var _ resource.ResourceWithImportState = &airflowClusterResource{}
var _ resource.ResourceWithValidateConfig = &airflowClusterResource{}

func NewResource() resource.Resource {
	return &airflowClusterResource{}
}

type airflowClusterResource struct {
	providerConfig *provider_config.Config
}

// Metadata implements resource.Resource.
func (a *airflowClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_airflow_cluster"
}

// Configure implements resource.Resource.
func (a *airflowClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState implements resource.ResourceWithImportState.
func (a *airflowClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	adminPassword := path.Root("admin_password")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, adminPassword, af_api.AdminPasswordStubOnImport)...)
}

// Create implements resource.Resource.
func (a *airflowClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan af_api.ClusterModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createClusterRequest, diags := af_api.BuildCreateClusterRequest(ctx, &plan, &a.providerConfig.ProviderState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Create Airflow cluster request: %+v", createClusterRequest))

	createTimeout, diags := plan.Timeouts.Create(ctx, af_api.YandexAirflowClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	clusterID, d := af_api.CreateCluster(ctx, a.providerConfig.SDK, &resp.Diagnostics, createClusterRequest)
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

	tflog.Debug(ctx, "Finished creating Airflow cluster", clusterIDLogField(clusterID))
}

// Delete implements resource.Resource.
func (a *airflowClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state af_api.ClusterModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting Airflow cluster", clusterIDLogField(clusterID))

	deleteTimeout, diags := state.Timeouts.Delete(ctx, af_api.YandexAirflowClusterDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	d := af_api.DeleteCluster(ctx, a.providerConfig.SDK, clusterID)
	resp.Diagnostics.Append(d)

	tflog.Debug(ctx, "Finished deleting Airflow cluster", clusterIDLogField(clusterID))
}

// Read implements resource.Resource.
func (a *airflowClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state af_api.ClusterModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.Id.ValueString()
	tflog.Debug(ctx, "Reading Airflow cluster", clusterIDLogField(clusterID))
	cluster, d := af_api.GetClusterByID(ctx, a.providerConfig.SDK, clusterID)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	if cluster == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = af_api.ClusterToState(ctx, cluster, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished reading Airflow cluster", clusterIDLogField(clusterID))
}

// Update implements resource.Resource.
func (a *airflowClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan af_api.ClusterModel
	var state af_api.ClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Airflow cluster", clusterIDLogField(state.Id.ValueString()))

	updateTimeout, diags := plan.Timeouts.Update(ctx, af_api.YandexAirflowClusterUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Debug(ctx, fmt.Sprintf("Update Airflow cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update Airflow cluster plan: %+v", plan))

	updateReq, diags := af_api.BuildUpdateClusterRequest(ctx, &state, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Update Airflow cluster request: %+v", updateReq))

	d := af_api.UpdateCluster(ctx, a.providerConfig.SDK, updateReq)
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
	tflog.Debug(ctx, "Finished updating Airflow cluster", clusterIDLogField(state.Id.ValueString()))
}

// Schema implements resource.Resource.
func (a *airflowClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = af_api.ClusterResourceSchema(ctx)
	resp.Schema.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})
}

func (r *airflowClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cluster af_api.ClusterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cluster)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !cluster.Logging.IsNull() {
		// both folder_id and log_group_id are specified or both are not specified
		if cluster.Logging.FolderId.IsNull() == cluster.Logging.LogGroupId.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("logging"),
				"Invalid Airflow cluster logging configuration",
				"Exactly one of the attributes `folder_id` and `log_group_id` must be specified",
			)
			return
		}
	}
}

func updateState(ctx context.Context, sdk *ycsdk.SDK, state *af_api.ClusterModel) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterID := state.Id.ValueString()
	tflog.Debug(ctx, "Reading Airflow cluster", clusterIDLogField(clusterID))
	cluster, d := af_api.GetClusterByID(ctx, sdk, clusterID)
	diags.Append(d)
	if diags.HasError() {
		return diags
	}

	if cluster == nil {
		diags.AddError(
			"Airflow cluster not found",
			fmt.Sprintf("Airflow cluster with id %s not found", clusterID))
		return diags
	}

	dd := af_api.ClusterToState(ctx, cluster, state)
	diags.Append(dd...)
	return diags
}

func clusterIDLogField(cid string) map[string]interface{} {
	return map[string]interface{}{
		"cluster_id": cid,
	}
}
