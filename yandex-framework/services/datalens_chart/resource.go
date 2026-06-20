package datalens_chart

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
	_ resource.Resource                = (*chartResource)(nil)
	_ resource.ResourceWithConfigure   = (*chartResource)(nil)
	_ resource.ResourceWithImportState = (*chartResource)(nil)
)

type chartResource struct {
	providerConfig *provider_config.Config
	client         *chartClient
}

func NewResource() resource.Resource {
	return &chartResource{}
}

func (r *chartResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_chart"
}

func (r *chartResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.AddWarning(
		"Experimental resource",
		"yandex_datalens_chart wraps DataLens chart endpoints that are marked "+
			"Experimental in the upstream API. The schema and behavior may "+
			"change in future provider versions.",
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
	r.client = &chartClient{client: dlClient}
}

func (r *chartResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *chartResource) resolveOrgID(model *chartModel) string {
	if !model.OrganizationId.IsNull() && !model.OrganizationId.IsUnknown() {
		return model.OrganizationId.ValueString()
	}
	return r.providerConfig.ProviderState.OrganizationID.ValueString()
}

func (r *chartResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating DataLens chart")

	var plan chartModel
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

	body, err := marshalChart(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to serialize plan", err.Error())
		return
	}
	for _, k := range []string{"entryId", "createdAt", "updatedAt", "revId", "savedId", "publishedId"} {
		delete(body, k)
	}

	chartType := plan.Type.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Creating DataLens %s chart %q", chartType, plan.Name.ValueString()))

	createResp, err := r.client.CreateChart(ctx, orgID, chartType, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create DataLens chart", err.Error())
		return
	}

	if v, ok := createResp["entryId"].(string); ok && v != "" {
		plan.Id = types.StringValue(v)
	}
	plan.OrganizationId = types.StringValue(orgID)

	apiResp, err := r.client.GetChart(ctx, orgID, chartType, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens chart after creation", err.Error())
		return
	}
	if err := unmarshalChartResponse(&plan, apiResp); err != nil {
		resp.Diagnostics.AddError("Failed to populate chart state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *chartResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens chart")

	var state chartModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	chartType := state.Type.ValueString()
	apiResp, err := r.client.GetChart(ctx, orgID, chartType, state.Id.ValueString())
	if err != nil {
		if datalens.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read DataLens chart", err.Error())
		return
	}

	if err := unmarshalChartResponse(&state, apiResp); err != nil {
		resp.Diagnostics.AddError("Failed to populate chart state from response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *chartResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating DataLens chart")

	var plan, state chartModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = state.Id
	plan.OrganizationId = state.OrganizationId

	body, err := marshalChart(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to serialize plan", err.Error())
		return
	}
	// Update API expects { chartId, annotation, data, mode }.
	updateBody := map[string]any{
		"chartId":    plan.Id.ValueString(),
		"data":       body["data"],
		"annotation": body["annotation"],
	}

	orgID := r.resolveOrgID(&plan)
	if err := r.client.UpdateChart(ctx, orgID, plan.Type.ValueString(), updateBody); err != nil {
		resp.Diagnostics.AddError("Unable to update DataLens chart", err.Error())
		return
	}

	apiResp, err := r.client.GetChart(ctx, orgID, plan.Type.ValueString(), plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DataLens chart after update", err.Error())
		return
	}
	if err := unmarshalChartResponse(&plan, apiResp); err != nil {
		resp.Diagnostics.AddError("Failed to populate chart state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *chartResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Deleting DataLens chart")

	var state chartModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	if err := r.client.DeleteChart(ctx, orgID, state.Type.ValueString(), state.Id.ValueString()); err != nil {
		if datalens.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete DataLens chart", err.Error())
		return
	}
}

func (r *chartResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	orgID, rest, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected format `organization_id:type:chart_id` (type one of wizard|ql).")
		return
	}
	chartType, chartID, err := resourceid.Deconstruct(rest)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected format `organization_id:type:chart_id` (type one of wizard|ql).")
		return
	}
	if _, err := chartRPCSuffix(chartType); err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), chartType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), chartID)...)
}
