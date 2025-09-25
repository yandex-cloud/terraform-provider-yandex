package trino_access_control

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/trino_access_control/models"
)

func NewResource() resource.Resource {
	return &trinoAccessControlResource{}
}

var (
	_ resource.Resource                   = &trinoAccessControlResource{}
	_ resource.ResourceWithImportState    = &trinoAccessControlResource{}
	_ resource.ResourceWithValidateConfig = &trinoAccessControlResource{}
)

type trinoAccessControlResource struct {
	providerConfig *provider_config.Config
}

// Configure implements resource.Resource.
func (t *trinoAccessControlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	t.providerConfig = providerConfig
}

// Create implements resource.Resource.
func (t *trinoAccessControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.AccessControlModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := plan.ClusterId.ValueString()
	accessControl, dd := models.ToAPI(ctx, &plan)
	resp.Diagnostics.Append(dd...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Create(ctx, YandexTrinoAccessControlCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	resp.Diagnostics.Append(UpdateClusterAccessControl(ctx, t.providerConfig.SDK, clusterID, accessControl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	tflog.Debug(ctx, "Finished updating Trino access control", clusterIDLogField(clusterID))
}

// Delete implements resource.Resource.
func (t *trinoAccessControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan models.AccessControlModel
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := plan.ClusterId.ValueString()
	updateTimeout, diags := plan.Timeouts.Create(ctx, YandexTrinoAccessControlDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	resp.Diagnostics.Append(DeleteClusterAccessControl(ctx, t.providerConfig.SDK, clusterID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished deleting Trino access control", clusterIDLogField(clusterID))
}

// Metadata implements resource.Resource.
func (t *trinoAccessControlResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trino_access_control"
}

// Read implements resource.Resource.
func (t *trinoAccessControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.AccessControlModel
	diags := req.State.Get(ctx, &state)
	tflog.Debug(ctx, fmt.Sprintf("Current access control state: %+v", state))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterId.ValueString()
	tflog.Debug(ctx, "Reading Trino access control", clusterIDLogField(clusterID))
	accessControl, diags := GetClusterAccessControl(ctx, t.providerConfig.SDK, clusterID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if accessControl == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	readState, diags := models.FromAPI(ctx, clusterID, accessControl)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// NOTE: Instead of setting readState to response apply changes to current state
	// in order to preserve current state field that are not changed semantically.
	// Otherwise, Terraform may report outside resource changes
	// right after successful `terraform apply` with config having empty strings or containers.

	resp.Diagnostics.Append(state.ApplyChanges(readState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Read access control state: %+v", state))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished reading Trino access control", clusterIDLogField(clusterID))
}

// Schema implements resource.Resource.
func (t *trinoAccessControlResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AccessControlResourceSchema(ctx)
	resp.Schema.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})
}

// Update implements resource.Resource.
func (t *trinoAccessControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.AccessControlModel
	var state models.AccessControlModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := plan.ClusterId.ValueString()
	if models.EqualSemantically(state, plan) {
		tflog.Warn(ctx, "No effective changes detected in access control. State will be updated without actual API call.")
	} else {
		resp.Diagnostics.Append(t.update(ctx, plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Cluster access control state after update: %+v", plan))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	tflog.Debug(ctx, "Finished updating Trino access control", clusterIDLogField(clusterID))
}

func (t *trinoAccessControlResource) update(ctx context.Context, plan models.AccessControlModel) diag.Diagnostics {
	var diags diag.Diagnostics
	accessControl, dd := models.ToAPI(ctx, &plan)
	diags.Append(dd...)
	if diags.HasError() {
		return diags
	}

	updateTimeout, dd := plan.Timeouts.Create(ctx, YandexTrinoAccessControlUpdateTimeout)
	diags.Append(dd...)
	if diags.HasError() {
		return diags
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	diags.Append(UpdateClusterAccessControl(ctx, t.providerConfig.SDK, plan.ClusterId.ValueString(), accessControl)...)
	return diags
}

// ValidateConfig implements resource.ResourceWithValidateConfig.
func (t *trinoAccessControlResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var model models.AccessControlModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(model.Validate()...)
}

// ImportState implements resource.ResourceWithImportState.
func (t *trinoAccessControlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("cluster_id"), req, resp)
}

func clusterIDLogField(cid string) map[string]interface{} {
	return map[string]interface{}{
		"cluster_id": cid,
	}
}
