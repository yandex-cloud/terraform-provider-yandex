package datalens_connection

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
	_ resource.Resource                = (*connectionResource)(nil)
	_ resource.ResourceWithConfigure   = (*connectionResource)(nil)
	_ resource.ResourceWithImportState = (*connectionResource)(nil)
)

type connectionResource struct {
	providerConfig *provider_config.Config
	client         *connectionClient
}

func NewResource() resource.Resource {
	return &connectionResource{}
}

func (r *connectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalens_connection"
}

func (r *connectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = &connectionClient{client: dlClient}
}

func (r *connectionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *connectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating DataLens connection")

	var plan connectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&plan)
	if orgID == "" {
		resp.Diagnostics.AddError(
			"Missing Organization ID",
			"organization_id must be specified either in the resource configuration or in the provider configuration.",
		)
		return
	}


	body, err := marshalConnection(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to serialize plan", err.Error())
		return
	}
	for _, k := range []string{"id", "key", "created_at", "updated_at"} {
		delete(body, k)
	}

	connectionID, err := r.client.CreateConnection(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create DataLens Connection", err.Error())
		return
	}

	plan.Id = types.StringValue(connectionID)
	plan.OrganizationId = types.StringValue(orgID)

	apiResp, err := r.client.GetConnection(ctx, orgID, connectionID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read DataLens Connection after creation", err.Error())
		return
	}
	if err := unmarshalConnection(apiResp, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to parse get response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading DataLens connection")

	var state connectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	apiResponse, err := r.client.GetConnection(ctx, orgID, state.Id.ValueString())
	if err != nil {
		if datalens.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read DataLens Connection", err.Error())
		return
	}

	if err := unmarshalConnection(apiResponse, &state); err != nil {
		resp.Diagnostics.AddError("Failed to parse get response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *connectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating DataLens connection")

	var plan, state connectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&plan)
	plan.Id = state.Id


	data, err := marshalConnection(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to serialize plan", err.Error())
		return
	}
	// Update API expects only mutable fields under `data`.
	for _, k := range []string{"id", "type", "name", "key", "workbook_id", "dir_path", "created_at", "updated_at"} {
		delete(data, k)
	}

	if _, err := r.client.UpdateConnection(ctx, orgID, plan.Id.ValueString(), data); err != nil {
		resp.Diagnostics.AddError("Unable to Update DataLens Connection", err.Error())
		return
	}

	apiResp, err := r.client.GetConnection(ctx, orgID, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read DataLens Connection after update", err.Error())
		return
	}
	if err := unmarshalConnection(apiResp, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to parse get response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Deleting DataLens connection")

	var state connectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := r.resolveOrgID(&state)
	if err := r.client.DeleteConnection(ctx, orgID, state.Id.ValueString()); err != nil {
		if datalens.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to Delete DataLens Connection", err.Error())
		return
	}
}

func (r *connectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	orgID, connectionID, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Import ID must be in the format organization_id:connection_id. Error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), connectionID)...)
}

func (r *connectionResource) resolveOrgID(model *connectionModel) string {
	if !model.OrganizationId.IsNull() && !model.OrganizationId.IsUnknown() {
		return model.OrganizationId.ValueString()
	}
	return r.providerConfig.ProviderState.OrganizationID.ValueString()
}

// marshalConnection serializes the model into a wire body. Since variant
// blocks have `wire:"-"`, we explicitly merge their flat representation back
// into the parent body so DataLens sees the type-specific fields side-by-side
// with type/name/description.
func marshalConnection(m *connectionModel) (map[string]any, error) {
	body, err := wire.Marshal(m)
	if err != nil {
		return nil, err
	}
	if m.Ydb != nil {
		vb, err := wire.Marshal(m.Ydb)
		if err != nil {
			return nil, fmt.Errorf("ydb: %w", err)
		}
		for k, v := range vb {
			body[k] = v
		}
	}
	return body, nil
}

// unmarshalConnection fills the model from the API response. DataLens echoes
// the discriminator as `db_type` in get/update responses (vs `type` in create
// requests) — alias it. Then run wire.Unmarshal twice: first for the common
// fields on `connectionModel`, then for the variant struct from the same flat
// response.
func unmarshalConnection(resp map[string]any, m *connectionModel) error {
	if v, ok := resp["db_type"]; ok && v != nil {
		resp["type"] = v
	}
	if err := wire.Unmarshal(resp, m); err != nil {
		return err
	}
	switch m.Type.ValueString() {
	case "ydb":
		m.Ydb = &ydbConfigModel{}
		if err := wire.Unmarshal(resp, m.Ydb); err != nil {
			return fmt.Errorf("ydb: %w", err)
		}
	}
	return nil
}

func unmarshalConnectionDataSource(resp map[string]any, m *connectionDataSourceModel) error {
	if v, ok := resp["db_type"]; ok && v != nil {
		resp["type"] = v
	}
	if err := wire.Unmarshal(resp, m); err != nil {
		return err
	}
	switch m.Type.ValueString() {
	case "ydb":
		m.Ydb = &ydbDataSourceConfigModel{}
		if err := wire.Unmarshal(resp, m.Ydb); err != nil {
			return fmt.Errorf("ydb: %w", err)
		}
	}
	return nil
}
