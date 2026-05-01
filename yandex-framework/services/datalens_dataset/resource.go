package datalens_dataset

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens/wire"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var (
	_ resource.Resource                = (*datasetResource)(nil)
	_ resource.ResourceWithConfigure   = (*datasetResource)(nil)
	_ resource.ResourceWithImportState = (*datasetResource)(nil)
)

type datasetResource struct {
	providerConfig *provider_config.Config
	client         *datasetClient
}

func NewResource() resource.Resource {
	return &datasetResource{}
}

func (r *datasetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_dataset"
}

func (r *datasetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T.", req.ProviderData),
		)
		return
	}
	r.providerConfig = providerConfig

	dlClient, err := datalens.NewClient(datalens.Config{
		Endpoint: providerConfig.ProviderState.DatalensEndpoint.ValueString(),
		TokenProvider: func(ctx context.Context) (string, error) {
			t, err := providerConfig.SDK.CreateIAMToken(ctx)
			if err != nil {
				return "", fmt.Errorf("failed to get IAM token: %w", err)
			}
			return t.IamToken, nil
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create DataLens client", err.Error())
		return
	}
	r.client = &datasetClient{client: dlClient}
}

func (r *datasetResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *datasetResource) resolveOrgID(model *datasetModel) string {
	if !model.OrganizationId.IsNull() && !model.OrganizationId.IsUnknown() {
		return model.OrganizationId.ValueString()
	}
	return r.providerConfig.ProviderState.OrganizationID.ValueString()
}

func (r *datasetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating DataLens dataset")

	var plan datasetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&plan)
	if orgID == "" {
		resp.Diagnostics.AddError(
			"Missing Organization ID",
			"organization_id must be specified either on the resource or at the provider level.",
		)
		return
	}

	body, err := wire.Marshal(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to serialize plan", err.Error())
		return
	}
	for _, k := range []string{"id", "key", "is_favorite"} {
		delete(body, k)
	}

	createResp, err := r.client.CreateDataset(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create DataLens dataset", err.Error())
		return
	}

	if id, ok := createResp["id"].(string); ok {
		plan.Id = types.StringValue(id)
	}
	plan.OrganizationId = types.StringValue(orgID)
	// created_via is Optional+Computed but DataLens does not echo it back in
	// the response, so we have to materialize it here. Default to "api" when
	// the user did not pin a value.
	if plan.CreatedVia.IsNull() || plan.CreatedVia.IsUnknown() {
		plan.CreatedVia = types.StringValue("api")
	}

	apiResp, err := r.client.GetDataset(ctx, orgID, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens dataset after creation", err.Error())
		return
	}
	if err := wire.Unmarshal(apiResp, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to populate dataset state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *datasetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens dataset")

	var state datasetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	apiResp, err := r.client.GetDataset(ctx, orgID, state.Id.ValueString())
	if err != nil {
		if datalens.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read DataLens dataset", err.Error())
		return
	}

	if err := wire.Unmarshal(apiResp, &state); err != nil {
		resp.Diagnostics.AddError("Failed to populate dataset state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datasetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating DataLens dataset")

	var plan, state datasetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = state.Id
	plan.OrganizationId = state.OrganizationId

	body, err := wire.Marshal(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to serialize plan", err.Error())
		return
	}
	updateMap, _ := body["dataset"].(map[string]any)
	if updateMap == nil {
		resp.Diagnostics.AddError("Invalid plan", "`dataset` block is required")
		return
	}
	orgID := r.resolveOrgID(&plan)
	if err := r.client.UpdateDataset(ctx, orgID, plan.Id.ValueString(), updateMap); err != nil {
		resp.Diagnostics.AddError("Unable to update DataLens dataset", err.Error())
		return
	}

	apiResp, err := r.client.GetDataset(ctx, orgID, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens dataset after update", err.Error())
		return
	}
	if err := wire.Unmarshal(apiResp, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to populate dataset state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *datasetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Deleting DataLens dataset")

	var state datasetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	if err := r.client.DeleteDataset(ctx, orgID, state.Id.ValueString()); err != nil {
		if datalens.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete DataLens dataset", err.Error())
		return
	}
}

func (r *datasetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	orgID, datasetID, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected format `organization_id:dataset_id`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), datasetID)...)
}
