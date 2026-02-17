package datalens_connection

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

// Ensure provider defined types fully satisfy framework interfaces.
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
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig

	dlClient, err := datalens.NewClient(datalens.Config{
		Endpoint: providerConfig.ProviderState.DatalensEndpoint.ValueString(),
		TokenProvider: func(ctx context.Context) (string, error) {
			resp, err := providerConfig.SDK.CreateIAMToken(ctx)
			if err != nil {
				return "", fmt.Errorf("failed to get IAM token: %w", err)
			}
			return resp.IamToken, nil
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create DataLens client",
			fmt.Sprintf("Error creating the DataLens API client: %s", err),
		)
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

	body, err := r.buildCreateRequest(&plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to build create request",
			fmt.Sprintf("Error building the create connection request: %s", err),
		)
		return
	}

	connName := r.resolveConnectionName(&plan)
	tflog.Debug(ctx, fmt.Sprintf("Creating DataLens connection of type %s with name %s",
		plan.Type.ValueString(), connName))

	connectionID, err := r.client.CreateConnection(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create DataLens Connection",
			fmt.Sprintf("An unexpected error occurred while creating the connection. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	plan.Id = types.StringValue(connectionID)
	r.setOrgID(&plan, orgID)
	r.readAndPopulateState(ctx, &plan, resp)
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
		resp.Diagnostics.AddError(
			"Unable to Read DataLens Connection",
			fmt.Sprintf("An unexpected error occurred while reading the connection. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	r.populateModelFromResponse(&state, apiResponse)

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

	connectionID := plan.Id.ValueString()

	updateData, err := r.buildUpdateRequest(&plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to build update request",
			fmt.Sprintf("Error building the update connection request: %s", err),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating DataLens connection %s", connectionID))

	_, err = r.client.UpdateConnection(ctx, orgID, connectionID, updateData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update DataLens Connection",
			fmt.Sprintf("An unexpected error occurred while updating the connection. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	// The update API returns an empty response, so re-read to get the full state.
	apiResponse, err := r.client.GetConnection(ctx, orgID, connectionID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read DataLens Connection After Update",
			fmt.Sprintf("The connection was updated successfully, but an error occurred while "+
				"reading the updated state.\n\nError: %s", err),
		)
		return
	}

	r.populateModelFromResponse(&plan, apiResponse)

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
	connectionID := state.Id.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Deleting DataLens connection %s", connectionID))

	err := r.client.DeleteConnection(ctx, orgID, connectionID)
	if err != nil {
		if datalens.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete DataLens Connection",
			fmt.Sprintf("An unexpected error occurred while deleting the connection. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("DataLens connection %s deleted", connectionID))
}

func (r *connectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: organization_id:connection_id
	orgID, connectionID, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be in the format organization_id:connection_id. Error: %s", err),
		)
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

func (r *connectionResource) setOrgID(model *connectionModel, orgID string) {
	model.OrganizationId = types.StringValue(orgID)
}

func (r *connectionResource) resolveConnectionName(model *connectionModel) string {
	return model.Name.ValueString()
}

func flattenCommonFieldsToMap(model *connectionModel, m map[string]interface{}) {
	m["type"] = model.Type.ValueString()
	m["name"] = model.Name.ValueString()

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		m["description"] = model.Description.ValueString()
	}
}

func (r *connectionResource) buildCreateRequest(plan *connectionModel) (createConnectionRequest, error) {
	body := make(createConnectionRequest)

	flattenCommonFieldsToMap(plan, body)

	switch plan.Type.ValueString() {
	case "ydb":
		if plan.Ydb == nil {
			return nil, fmt.Errorf("ydb configuration block is required when type is \"ydb\"")
		}
		if (plan.Ydb.WorkbookId.IsNull() || plan.Ydb.WorkbookId.IsUnknown()) &&
			(plan.Ydb.DirPath.IsNull() || plan.Ydb.DirPath.IsUnknown()) {
			return nil, fmt.Errorf("either workbook_id or dir_path must be specified for ydb connection")
		}
		flattenYdbToMap(plan.Ydb, body)
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", plan.Type.ValueString())
	}

	return body, nil
}

// buildUpdateRequest builds the data map for the update connection API call.
// It includes name and description (mutable common fields) plus connection-type-specific
// fields. Immutable fields (type, workbook_id) are excluded.
func (r *connectionResource) buildUpdateRequest(plan *connectionModel) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// description is a mutable common field; name and type are immutable
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		data["description"] = plan.Description.ValueString()
	}

	switch plan.Type.ValueString() {
	case "ydb":
		if plan.Ydb == nil {
			return nil, fmt.Errorf("ydb configuration block is required when type is \"ydb\"")
		}
		flattenYdbToUpdateMap(plan.Ydb, data)
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", plan.Type.ValueString())
	}

	return data, nil
}

func setStringIfKnown(m map[string]interface{}, key string, val types.String) {
	if !val.IsNull() && !val.IsUnknown() {
		m[key] = val.ValueString()
	}
}

func (r *connectionResource) readAndPopulateState(ctx context.Context, model *connectionModel, resp *resource.CreateResponse) {
	orgID := r.resolveOrgID(model)

	apiResponse, err := r.client.GetConnection(ctx, orgID, model.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read DataLens Connection after creation",
			fmt.Sprintf("An unexpected error occurred while reading the connection after creation. "+
				"The connection was created successfully (ID: %s), but reading it failed.\n\n"+
				"Error: %s", model.Id.ValueString(), err),
		)
		return
	}

	r.populateModelFromResponse(model, apiResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *connectionResource) populateModelFromResponse(model *connectionModel, apiResponse map[string]interface{}) {
	if v, ok := apiResponse["id"].(string); ok {
		model.Id = types.StringValue(v)
	}
	// The API accepts "type" on create but returns "db_type" on read.
	if v, ok := apiResponse["db_type"].(string); ok {
		model.Type = types.StringValue(v)
	} else if v, ok := apiResponse["type"].(string); ok {
		model.Type = types.StringValue(v)
	}
	if v, ok := apiResponse["name"].(string); ok {
		model.Name = types.StringValue(v)
	}
	if v, ok := apiResponse["description"]; ok {
		if v == nil {
			model.Description = types.StringNull()
		} else if s, ok := v.(string); ok {
			if s == "" && model.Description.IsNull() {
				// API returns "" for unset description; preserve null to avoid plan diff.
				model.Description = types.StringNull()
			} else {
				model.Description = types.StringValue(s)
			}
		}
	}
	if v, ok := apiResponse["created_at"].(string); ok {
		model.CreatedAt = types.StringValue(v)
	}
	if v, ok := apiResponse["updated_at"].(string); ok {
		model.UpdatedAt = types.StringValue(v)
	}

	connType := model.Type.ValueString()
	switch connType {
	case "ydb":
		if model.Ydb == nil {
			model.Ydb = &ydbConfigModel{}
		}
		populateYdbFromResponse(model.Ydb, apiResponse)
	}
}
