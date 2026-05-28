package mdbcommon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	mdbv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ClusterConnectionManagerModel is the framework state model for ClusterConnectionManager.
type ClusterConnectionManagerModel struct {
	Enabled             types.Bool   `tfsdk:"enabled"`
	ConnectionsFolderId types.String `tfsdk:"connections_folder_id"`
	SecretsFolderId     types.String `tfsdk:"secrets_folder_id"`
}

var ClusterConnectionManagerAttrTypes = map[string]attr.Type{
	"enabled":               types.BoolType,
	"connections_folder_id": types.StringType,
	"secrets_folder_id":     types.StringType,
}

// ClusterConnectionManagerFrameworkSchema returns the SingleNestedAttribute for the
// connection_manager block shared by framework-based cluster resources.
// enabled is Optional+Computed because the API chooses the default for new clusters;
// folder IDs are plain Optional and follow standard Terraform Optional semantics
// (omitted in HCL == default value == empty string on the server).
func ClusterConnectionManagerFrameworkSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Connection Manager integration settings. If the block is omitted, the API enables the integration by default for newly created clusters. Disabling the integration after the cluster is created is not supported.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicates whether Connection Manager integration is enabled. Set to `true` to enable the integration. If omitted, the API enables the integration by default for newly created clusters. Disabling the integration after the cluster is created is not supported.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"connections_folder_id": schema.StringAttribute{
				Description: "ID of the folder where connections for the cluster are created. Defaults to the cluster's folder if not specified.",
				Optional:    true,
			},
			"secrets_folder_id": schema.StringAttribute{
				Description: "ID of the folder where connection secrets are created. Defaults to the cluster's folder if not specified.",
				Optional:    true,
			},
		},
	}
}

// ValidateClusterConnectionManagerFromConfig rejects an explicit `enabled = false` in HCL.
// Computed / missing values are ignored: the API may legitimately have enabled=false for
// clusters configured outside of Terraform.
func ValidateClusterConnectionManagerFromConfig(ctx context.Context, cfg tfsdk.Config, configPath path.Path, diags *diag.Diagnostics) {
	var cmConfig types.Object
	diags.Append(cfg.GetAttribute(ctx, configPath, &cmConfig)...)
	if diags.HasError() || cmConfig.IsNull() || cmConfig.IsUnknown() {
		return
	}
	var cm ClusterConnectionManagerModel
	diags.Append(cmConfig.As(ctx, &cm, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}
	if !cm.Enabled.IsNull() && !cm.Enabled.IsUnknown() && !cm.Enabled.ValueBool() {
		diags.AddError(
			"connection_manager.enabled cannot be set to false, disabling Connection Manager integration is not supported",
			"Remove the `enabled = false` line or set `enabled = true` to keep the integration enabled.",
		)
	}
}

// ClusterConnectionManagerUpdateMaskPaths returns update_mask paths for connection_manager
// fields whose plan value differs from state. enabled is treated as Optional+Computed and is
// only included when the user set it in HCL (a value coming from state via Computed is ignored).
// Folder IDs are plain Optional and follow standard Terraform semantics: any plan-vs-state
// difference is forwarded to the API, including "value -> null" transitions.
func ClusterConnectionManagerUpdateMaskPaths(ctx context.Context, plan, state types.Object, updateMaskPrefix string) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var planCM, stateCM ClusterConnectionManagerModel
	if !plan.IsNull() && !plan.IsUnknown() {
		diags.Append(plan.As(ctx, &planCM, basetypes.ObjectAsOptions{})...)
	}
	if !state.IsNull() && !state.IsUnknown() {
		diags.Append(state.As(ctx, &stateCM, basetypes.ObjectAsOptions{})...)
	}
	if diags.HasError() {
		return nil, diags
	}

	var paths []string
	// enabled: include only when explicitly set in HCL (computed values from state are ignored).
	if !planCM.Enabled.IsNull() && !planCM.Enabled.IsUnknown() && !planCM.Enabled.Equal(stateCM.Enabled) {
		paths = append(paths, updateMaskPrefix+"enabled")
	}
	// Folder IDs: any difference is a change (incl. null<->value).
	if !planCM.ConnectionsFolderId.IsUnknown() && !planCM.ConnectionsFolderId.Equal(stateCM.ConnectionsFolderId) {
		paths = append(paths, updateMaskPrefix+"connections_folder_id")
	}
	if !planCM.SecretsFolderId.IsUnknown() && !planCM.SecretsFolderId.Equal(stateCM.SecretsFolderId) {
		paths = append(paths, updateMaskPrefix+"secrets_folder_id")
	}
	return paths, diags
}

func ExpandClusterConnectionManagerFramework(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mdbv1.ClusterConnectionManager {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}

	var model ClusterConnectionManagerModel
	diags.Append(o.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	cm := &mdbv1.ClusterConnectionManager{}
	if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
		cm.Enabled = wrapperspb.Bool(model.Enabled.ValueBool())
	}
	if !model.ConnectionsFolderId.IsNull() && !model.ConnectionsFolderId.IsUnknown() {
		cm.ConnectionsFolderId = model.ConnectionsFolderId.ValueString()
	}
	if !model.SecretsFolderId.IsNull() && !model.SecretsFolderId.IsUnknown() {
		cm.SecretsFolderId = model.SecretsFolderId.ValueString()
	}
	return cm
}

// FlattenClusterConnectionManagerFramework converts a proto ClusterConnectionManager to a framework types.Object.
// Empty folder IDs from the API are flattened to Null so they line up with the standard Terraform
// Optional semantics (omitted in HCL => Null in plan, Null in state, no apply-time inconsistency).
func FlattenClusterConnectionManagerFramework(ctx context.Context, cm *mdbv1.ClusterConnectionManager, diags *diag.Diagnostics) types.Object {
	if cm == nil {
		return types.ObjectNull(ClusterConnectionManagerAttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, ClusterConnectionManagerAttrTypes, ClusterConnectionManagerModel{
		Enabled:             FlattenClusterConnectionManagerEnabled(cm.Enabled),
		ConnectionsFolderId: FlattenStringOrNull(cm.ConnectionsFolderId),
		SecretsFolderId:     FlattenStringOrNull(cm.SecretsFolderId),
	})
	diags.Append(d...)
	return obj
}

func FlattenClusterConnectionManagerEnabled(enabled *wrapperspb.BoolValue) types.Bool {
	if enabled == nil {
		return types.BoolNull()
	}
	return types.BoolValue(enabled.GetValue())
}
