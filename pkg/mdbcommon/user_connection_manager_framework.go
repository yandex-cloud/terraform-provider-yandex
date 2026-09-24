package mdbcommon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	mdbv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/v1"
)

// UserConnectionManagerModel is the framework state model for UserConnectionManager.
type UserConnectionManagerModel struct {
	ConnectionId       types.String `tfsdk:"connection_id"`
	ConnectionFolderId types.String `tfsdk:"connection_folder_id"`
	SecretFolderId     types.String `tfsdk:"secret_folder_id"`
}

var UserConnectionManagerAttrTypes = map[string]attr.Type{
	"connection_id":        types.StringType,
	"connection_folder_id": types.StringType,
	"secret_folder_id":     types.StringType,
}

func UserConnectionManagerFrameworkSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Connection Manager settings for the user.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{
				Description: "ID of the Connection Manager connection for this user. Computed by the server.",
				Computed:    true,
			},
			"connection_folder_id": schema.StringAttribute{
				Description: "ID of the folder where the connection is created. Defaults to the cluster's folder if not specified. Cannot be changed after user creation.",
				Optional:    true,
				Computed:    true,
			},
			"secret_folder_id": schema.StringAttribute{
				Description: "ID of the folder where the secret is created. Defaults to the cluster's folder if not specified. Cannot be changed after user creation.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// UserConnectionManagerFrameworkDataSourceSchema returns the schema for the
// user_connection_manager block in a datasource (all fields Computed).
func UserConnectionManagerFrameworkDataSourceSchema() datasourceschema.SingleNestedAttribute {
	return datasourceschema.SingleNestedAttribute{
		Description: "Connection Manager settings for the user.",
		Computed:    true,
		Attributes: map[string]datasourceschema.Attribute{
			"connection_id": datasourceschema.StringAttribute{
				Description: "ID of the Connection Manager connection for this user.",
				Computed:    true,
			},
			"connection_folder_id": datasourceschema.StringAttribute{
				Description: "ID of the folder where the connection is created.",
				Computed:    true,
			},
			"secret_folder_id": datasourceschema.StringAttribute{
				Description: "ID of the folder where the secret is created.",
				Computed:    true,
			},
		},
	}
}

// FlattenUserConnectionManagerFramework converts a proto UserConnectionManager to a framework types.Object.
// Returns a null object for a zero-value proto so clusters without Connection Manager integration
// do not get an empty block in state.
func FlattenUserConnectionManagerFramework(ctx context.Context, ucm *mdbv1.UserConnectionManager, diags *diag.Diagnostics) types.Object {
	if ucm == nil || (ucm.ConnectionId == "" && ucm.ConnectionFolderId == "" && ucm.SecretFolderId == "") {
		return types.ObjectNull(UserConnectionManagerAttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, UserConnectionManagerAttrTypes, UserConnectionManagerModel{
		ConnectionId:       FlattenStringOrNull(ucm.ConnectionId),
		ConnectionFolderId: FlattenStringOrNull(ucm.ConnectionFolderId),
		SecretFolderId:     FlattenStringOrNull(ucm.SecretFolderId),
	})
	diags.Append(d...)
	return obj
}

// ExpandUserConnectionManagerFramework converts a framework types.Object to a proto UserConnectionManager.
// Only folder IDs are sent, connection_id is assigned by the server.
func ExpandUserConnectionManagerFramework(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mdbv1.UserConnectionManager {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}

	var model UserConnectionManagerModel
	diags.Append(o.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	ucm := &mdbv1.UserConnectionManager{}
	if !model.ConnectionFolderId.IsNull() && !model.ConnectionFolderId.IsUnknown() {
		ucm.ConnectionFolderId = model.ConnectionFolderId.ValueString()
	}
	if !model.SecretFolderId.IsNull() && !model.SecretFolderId.IsUnknown() {
		ucm.SecretFolderId = model.SecretFolderId.ValueString()
	}
	return ucm
}

// ValidateUserConnectionManagerNotChanged rejects attempts to change folder IDs after creation.
// Fields omitted from HCL are not validated: that means "do not touch", not "clear the value".
func ValidateUserConnectionManagerNotChanged(ctx context.Context, config, state types.Object, configPath path.Path, diags *diag.Diagnostics) {
	if state.IsNull() || state.IsUnknown() || config.IsNull() || config.IsUnknown() {
		return
	}

	var configModel, stateModel UserConnectionManagerModel
	diags.Append(config.As(ctx, &configModel, basetypes.ObjectAsOptions{})...)
	diags.Append(state.As(ctx, &stateModel, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}

	validateFolder := func(name string, config, prev types.String) {
		if config.IsNull() || config.IsUnknown() || config.Equal(prev) {
			return
		}

		diags.AddAttributeError(
			configPath.AtName(name),
			"Connection Manager folder cannot be changed after user creation",
			"Recreate the user to move its connection or secret to another folder.",
		)
	}

	validateFolder("connection_folder_id", configModel.ConnectionFolderId, stateModel.ConnectionFolderId)
	validateFolder("secret_folder_id", configModel.SecretFolderId, stateModel.SecretFolderId)
}
