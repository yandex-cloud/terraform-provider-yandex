package trino_catalog

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

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &trinoCatalogResource{}
	_ resource.ResourceWithImportState    = &trinoCatalogResource{}
	_ resource.ResourceWithValidateConfig = &trinoCatalogResource{}
)

func NewResource() resource.Resource {
	return &trinoCatalogResource{}
}

type trinoCatalogResource struct {
	providerConfig *provider_config.Config
}

// Metadata implements resource.Resource.
func (t *trinoCatalogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trino_catalog"
}

// Configure implements resource.Resource.
func (t *trinoCatalogResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState implements resource.ResourceWithImportState.
func (r *trinoCatalogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterID, catalogID, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), catalogID)...)
}

// Create implements resource.Resource.
func (t *trinoCatalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CatalogModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createCatalogRequest, diags := BuildCreateCatalogRequest(ctx, &plan, &t.providerConfig.ProviderState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Create Trino catalog request: %+v", createCatalogRequest))

	createTimeout, diags := plan.Timeouts.Create(ctx, YandexTrinoCatalogCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	catalogID, d := CreateCatalog(ctx, t.providerConfig.SDK, &resp.Diagnostics, createCatalogRequest)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(catalogID)
	diags = updateState(ctx, t.providerConfig.SDK, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Finished creating Trino catalog", catalogIDLogField(catalogID))
}

// Delete implements resource.Resource.
func (t *trinoCatalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CatalogModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := state.Id.ValueString()
	clusterID := state.ClusterId.ValueString()
	tflog.Debug(ctx, "Deleting Trino catalog", catalogIDLogField(catalogID))

	deleteTimeout, diags := state.Timeouts.Delete(ctx, YandexTrinoCatalogDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	d := DeleteCatalog(ctx, t.providerConfig.SDK, catalogID, clusterID)
	resp.Diagnostics.Append(d)

	tflog.Debug(ctx, "Finished deleting Trino catalog", catalogIDLogField(catalogID))
}

// Read implements resource.Resource.
func (t *trinoCatalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CatalogModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := state.Id.ValueString()
	clusterID := state.ClusterId.ValueString()
	tflog.Debug(ctx, "Reading Trino catalog", catalogIDLogField(catalogID))
	catalog, d := GetCatalogByID(ctx, t.providerConfig.SDK, catalogID, clusterID)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	if catalog == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = CatalogToState(ctx, catalog, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished reading Trino catalog", catalogIDLogField(catalogID))
}

// Update implements resource.Resource.
func (t *trinoCatalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CatalogModel
	var state CatalogModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updating Trino catalog", catalogIDLogField(state.Id.ValueString()))

	updateTimeout, diags := plan.Timeouts.Update(ctx, YandexTrinoCatalogUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Debug(ctx, fmt.Sprintf("Update Trino catalog state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("Update Trino catalog plan: %+v", plan))

	updateReq, diags := BuildUpdateCatalogRequest(ctx, &state, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Update Trino catalog request: %+v", updateReq))

	d := UpdateCatalog(ctx, t.providerConfig.SDK, updateReq)
	resp.Diagnostics.Append(d)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = updateState(ctx, t.providerConfig.SDK, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "Finished updating Trino catalog", catalogIDLogField(state.Id.ValueString()))
}

// Schema implements resource.Resource.
func (t *trinoCatalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CatalogResourceSchema(ctx)
	resp.Schema.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})
}

func (t *trinoCatalogResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var catalog CatalogModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &catalog)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func updateState(ctx context.Context, sdk *ycsdk.SDK, state *CatalogModel) diag.Diagnostics {
	var diags diag.Diagnostics
	catalogID := state.Id.ValueString()
	clusterID := state.ClusterId.ValueString()

	tflog.Debug(ctx, "Reading Trino catalog", catalogIDLogField(catalogID))
	catalog, d := GetCatalogByID(ctx, sdk, catalogID, clusterID)
	diags.Append(d)
	if diags.HasError() {
		return diags
	}

	if catalog == nil {
		diags.AddError(
			"Trino catalog not found",
			fmt.Sprintf("Trino catalog with id %s not found", catalogID))
		return diags
	}

	dd := CatalogToState(ctx, catalog, state)
	diags.Append(dd...)
	return diags
}

func catalogIDLogField(cid string) map[string]interface{} {
	return map[string]interface{}{
		"catalog_id": cid,
	}
}
