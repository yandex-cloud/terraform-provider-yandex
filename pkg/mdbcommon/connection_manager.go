package mdbcommon

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mdbv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/v1"
)

// ClusterConnectionManagerSchema returns the full connection_manager block for a resource:
// a TypeList with MaxItems=1 wrapping the connection_manager attributes. The block is
// Optional+Computed because the API chooses defaults for new clusters when it is omitted.
// Inside the block enabled is Optional+Computed; folder ID fields are plain Optional and
// omitting them sends an empty string to the API, which defaults the folder to the cluster's.
func ClusterConnectionManagerSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Connection Manager integration configuration for the cluster. If the block is omitted, the API enables the integration by default for newly created clusters. Disabling the integration after the cluster is created is not supported.",
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Description: "Indicates whether Connection Manager integration is enabled for the cluster. Set to `true` to enable the integration. If the block is omitted, the API enables the integration by default for newly created clusters. Disabling the integration after the cluster is created is not supported.",
					Optional:    true,
					Computed:    true,
				},
				"connections_folder_id": {
					Type:        schema.TypeString,
					Description: "ID of the folder where connections for the cluster are created. Defaults to the cluster's folder if not specified.",
					Optional:    true,
				},
				"secrets_folder_id": {
					Type:        schema.TypeString,
					Description: "ID of the folder where connection secrets are created. Defaults to the cluster's folder if not specified.",
					Optional:    true,
				},
			},
		},
	}
}

// ClusterConnectionManagerDataSourceSchema returns the full connection_manager block for a
// datasource: a Computed TypeList with all inner fields Computed.
func ClusterConnectionManagerDataSourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Connection Manager integration configuration for the cluster.",
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Description: "Indicates whether Connection Manager integration is enabled for the cluster.",
					Computed:    true,
				},
				"connections_folder_id": {
					Type:        schema.TypeString,
					Description: "ID of the folder where connections for the cluster are created.",
					Computed:    true,
				},
				"secrets_folder_id": {
					Type:        schema.TypeString,
					Description: "ID of the folder where connection secrets are created.",
					Computed:    true,
				},
			},
		},
	}
}

// FlattenClusterConnectionManager converts a proto ClusterConnectionManager to terraform state.
func FlattenClusterConnectionManager(cm *mdbv1.ClusterConnectionManager) []interface{} {
	if cm == nil {
		return nil
	}

	out := map[string]interface{}{}
	if cm.Enabled != nil {
		out["enabled"] = cm.Enabled.GetValue()
	}
	out["connections_folder_id"] = cm.ConnectionsFolderId
	out["secrets_folder_id"] = cm.SecretsFolderId

	return []interface{}{out}
}

// ExpandClusterConnectionManager converts the user's configuration into a proto ClusterConnectionManager.
// enabled is only written when explicitly set in HCL (it is Optional+Computed and we must not
// echo back state-derived values); folder IDs follow plain Optional semantics and are passed as-is.
func ExpandClusterConnectionManager(d *schema.ResourceData, connectionManagerPath string) *mdbv1.ClusterConnectionManager {
	if _, ok := LookupRawConfigPath(d, connectionManagerPath+".0"); !ok {
		return nil
	}

	cm := &mdbv1.ClusterConnectionManager{
		ConnectionsFolderId: d.Get(connectionManagerPath + ".0.connections_folder_id").(string),
		SecretsFolderId:     d.Get(connectionManagerPath + ".0.secrets_folder_id").(string),
	}

	if v, ok := LookupRawConfigPath(d, connectionManagerPath+".0.enabled"); ok && v.Type() == cty.Bool {
		cm.Enabled = wrapperspb.Bool(v.True())
	}

	return cm
}

// ClusterConnectionManagerEnabledChangedPath returns the update_mask path for
// connection_manager.enabled if the user explicitly set it in HCL and the value differs from state.
// Folder ID fields are not handled here: as plain Optional strings they are picked up
// by the resource's generic d.HasChange-based update mask, see ClusterConnectionManagerUpdateFields.
func ClusterConnectionManagerEnabledChangedPath(d *schema.ResourceData, connectionManagerPath, updateMaskPrefix string) []string {
	enabledPath := connectionManagerPath + ".0.enabled"
	cfgVal, ok := LookupRawConfigPath(d, enabledPath)
	if !ok || cfgVal.Type() != cty.Bool {
		return nil
	}
	oldVal, newVal := d.GetChange(enabledPath)
	if oldVal.(bool) == newVal.(bool) {
		return nil
	}
	return []string{updateMaskPrefix + "enabled"}
}

// ClusterConnectionManagerUpdateFields returns the schema-path -> update-mask-path
// pairs for connection_manager folder ID fields. It is meant to be merged into the
// resource's d.HasChange-based update mask map so callers do not need to know the
// internal structure of the connection_manager block.
func ClusterConnectionManagerUpdateFields(connectionManagerPath, updateMaskPrefix string) map[string]string {
	return map[string]string{
		connectionManagerPath + ".0.connections_folder_id": updateMaskPrefix + "connections_folder_id",
		connectionManagerPath + ".0.secrets_folder_id":     updateMaskPrefix + "secrets_folder_id",
	}
}

// CustomizeDiffClusterConnectionManager rejects an explicit `enabled = false` in HCL.
// Computed / missing values are ignored: the API may legitimately have enabled=false
// for clusters configured outside of Terraform.
func CustomizeDiffClusterConnectionManager(_ context.Context, d *schema.ResourceDiff, connectionManagerPath string) error {
	enabledPath := connectionManagerPath + ".0.enabled"
	enabled, ok := LookupRawConfigPath(d, enabledPath)
	if !ok {
		return nil
	}
	if enabled.Type() != cty.Bool {
		return fmt.Errorf("CustomizeDiffClusterConnectionManager: expected %s to be bool, got %s", enabledPath, enabled.Type().FriendlyName())
	}
	if enabled.True() {
		return nil
	}
	return fmt.Errorf("connection_manager.enabled cannot be set to false, disabling Connection Manager integration is not supported")
}
