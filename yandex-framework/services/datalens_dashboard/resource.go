package datalens_dashboard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var (
	_ resource.Resource                = (*dashboardResource)(nil)
	_ resource.ResourceWithConfigure   = (*dashboardResource)(nil)
	_ resource.ResourceWithImportState = (*dashboardResource)(nil)
)

type dashboardResource struct {
	providerConfig *provider_config.Config
	client         *dashboardClient
}

func NewResource() resource.Resource {
	return &dashboardResource{}
}

func (r *dashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_dashboard"
}

func (r *dashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.AddWarning(
		"Experimental resource",
		"yandex_datalens_dashboard wraps DataLens dashboard endpoints that are "+
			"marked Experimental in the upstream API. The schema and behavior "+
			"may change in future provider versions.",
	)
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
	r.client = &dashboardClient{client: dlClient}
}

func (r *dashboardResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *dashboardResource) resolveOrgID(model *dashboardModel) string {
	if !model.OrganizationId.IsNull() && !model.OrganizationId.IsUnknown() {
		return model.OrganizationId.ValueString()
	}
	return r.providerConfig.ProviderState.OrganizationID.ValueString()
}

func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating DataLens dashboard")

	var plan dashboardModel
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

	body, err := marshalDashboard(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid plan", err.Error())
		return
	}

	createResp, err := r.client.CreateDashboard(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create DataLens dashboard", err.Error())
		return
	}

	if entry, ok := createResp["entry"].(map[string]any); ok {
		if v, ok := entry["entryId"].(string); ok && v != "" {
			plan.Id = types.StringValue(v)
		}
	}
	plan.OrganizationId = types.StringValue(orgID)

	apiResp, err := r.client.GetDashboard(ctx, orgID, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens dashboard after creation", err.Error())
		return
	}
	if err := unmarshalDashboardResponse(&plan, apiResp); err != nil {
		resp.Diagnostics.AddError("Failed to populate dashboard state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens dashboard")

	var state dashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	apiResp, err := r.client.GetDashboard(ctx, orgID, state.Id.ValueString())
	if err != nil {
		if datalens.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read DataLens dashboard", err.Error())
		return
	}

	if err := unmarshalDashboardResponse(&state, apiResp); err != nil {
		resp.Diagnostics.AddError("Failed to populate dashboard state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating DataLens dashboard")

	var plan, state dashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = state.Id
	plan.OrganizationId = state.OrganizationId

	body, err := marshalDashboard(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid plan", err.Error())
		return
	}
	body["dashboardId"] = plan.Id.ValueString()

	orgID := r.resolveOrgID(&plan)
	if err := r.client.UpdateDashboard(ctx, orgID, body); err != nil {
		resp.Diagnostics.AddError("Unable to update DataLens dashboard", err.Error())
		return
	}

	apiResp, err := r.client.GetDashboard(ctx, orgID, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens dashboard after update", err.Error())
		return
	}
	if err := unmarshalDashboardResponse(&plan, apiResp); err != nil {
		resp.Diagnostics.AddError("Failed to populate dashboard state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Deleting DataLens dashboard")

	var state dashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	if err := r.client.DeleteDashboard(ctx, orgID, state.Id.ValueString()); err != nil {
		if datalens.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete DataLens dashboard", err.Error())
		return
	}
}

func (r *dashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	orgID, dashboardID, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected format `organization_id:dashboard_id`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), dashboardID)...)
}
