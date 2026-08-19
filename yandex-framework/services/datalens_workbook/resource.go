package datalens_workbook

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
	_ resource.Resource                = (*workbookResource)(nil)
	_ resource.ResourceWithConfigure   = (*workbookResource)(nil)
	_ resource.ResourceWithImportState = (*workbookResource)(nil)
)

type workbookResource struct {
	providerConfig *provider_config.Config
	client         *workbookClient
}

func NewResource() resource.Resource {
	return &workbookResource{}
}

func (r *workbookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_workbook"
}

func (r *workbookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = &workbookClient{client: dlClient}
}

func (r *workbookResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *workbookResource) resolveOrgID(model *workbookModel) string {
	if !model.OrganizationId.IsNull() && !model.OrganizationId.IsUnknown() {
		return model.OrganizationId.ValueString()
	}
	return r.providerConfig.ProviderState.OrganizationID.ValueString()
}

func (r *workbookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating DataLens workbook")

	var plan workbookModel
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
	delete(body, "workbookId") // not allowed at create

	createResp, err := r.client.CreateWorkbook(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create DataLens workbook", err.Error())
		return
	}

	if err := wire.Unmarshal(createResp, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}
	if plan.Id.IsNull() || plan.Id.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Unexpected create response",
			fmt.Sprintf("createWorkbook did not return workbookId. Response: %v", createResp),
		)
		return
	}
	plan.OrganizationId = types.StringValue(orgID)

	apiResp, err := r.client.GetWorkbook(ctx, orgID, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens workbook after creation", err.Error())
		return
	}
	if err := wire.Unmarshal(apiResp, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to parse get response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workbookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens workbook")

	var state workbookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	apiResp, err := r.client.GetWorkbook(ctx, orgID, state.Id.ValueString())
	if err != nil {
		if datalens.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read DataLens workbook", err.Error())
		return
	}

	if err := wire.Unmarshal(apiResp, &state); err != nil {
		resp.Diagnostics.AddError("Failed to parse get response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workbookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating DataLens workbook")

	var plan, state workbookModel
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

	orgID := r.resolveOrgID(&plan)
	if _, err := r.client.UpdateWorkbook(ctx, orgID, body); err != nil {
		resp.Diagnostics.AddError("Unable to update DataLens workbook", err.Error())
		return
	}

	apiResp, err := r.client.GetWorkbook(ctx, orgID, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens workbook after update", err.Error())
		return
	}
	if err := wire.Unmarshal(apiResp, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to parse get response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workbookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Deleting DataLens workbook")

	var state workbookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	if err := r.client.DeleteWorkbook(ctx, orgID, state.Id.ValueString()); err != nil {
		if datalens.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete DataLens workbook", err.Error())
		return
	}
}

func (r *workbookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	orgID, workbookID, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected format `organization_id:workbook_id`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), workbookID)...)
}
