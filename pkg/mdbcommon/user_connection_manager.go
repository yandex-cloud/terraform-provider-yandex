package mdbcommon

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	mdbv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/v1"
)

// UserConnectionManagerSchema returns the schema for the user_connection_manager block in a resource.
// Folder IDs are Optional+Computed so omitting them from HCL keeps the effective value
// from state instead of forcing a "value -> null" diff; CustomizeDiffUserConnectionManager
// enforces the "cannot be changed after creation" rule.
func UserConnectionManagerSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Connection Manager settings for the user.",
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"connection_id": {
					Type:        schema.TypeString,
					Description: "ID of the connection manager connection for this user. Computed by the server.",
					Computed:    true,
				},
				"connection_folder_id": {
					Type:        schema.TypeString,
					Description: "ID of the folder where the connection is created. Defaults to the cluster's folder if not specified. Cannot be changed after user creation.",
					Optional:    true,
					Computed:    true,
				},
				"secret_folder_id": {
					Type:        schema.TypeString,
					Description: "ID of the folder where the secret is created. Defaults to the cluster's folder if not specified. Cannot be changed after user creation.",
					Optional:    true,
					Computed:    true,
				},
			},
		},
	}
}

// UserConnectionManagerDataSourceSchema returns the schema for the user_connection_manager
// block in a datasource (all fields Computed).
func UserConnectionManagerDataSourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Connection Manager settings for the user.",
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"connection_id": {
					Type:        schema.TypeString,
					Description: "ID of the connection manager connection for this user.",
					Computed:    true,
				},
				"connection_folder_id": {
					Type:        schema.TypeString,
					Description: "ID of the folder where the connection is created. Defaults to the cluster's folder if not specified.",
					Computed:    true,
				},
				"secret_folder_id": {
					Type:        schema.TypeString,
					Description: "ID of the folder where the secret is created. Defaults to the cluster's folder if not specified.",
					Computed:    true,
				},
			},
		},
	}
}

// FlattenUserConnectionManager converts a proto UserConnectionManager to a list of maps.
// Returns nil for a zero-value proto so very old clusters without Connection Manager
// integration support do not emit an empty block into state (which would cause permanent drift).
func FlattenUserConnectionManager(ucm *mdbv1.UserConnectionManager) []interface{} {
	if ucm == nil {
		return nil
	}
	if ucm.ConnectionId == "" && ucm.ConnectionFolderId == "" && ucm.SecretFolderId == "" {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"connection_id":        ucm.ConnectionId,
			"connection_folder_id": ucm.ConnectionFolderId,
			"secret_folder_id":     ucm.SecretFolderId,
		},
	}
}

// ExpandUserConnectionManager converts terraform state to a proto UserConnectionManager.
func ExpandUserConnectionManager(d *schema.ResourceData, userConnectionManagerPath string) *mdbv1.UserConnectionManager {
	v, ok := d.GetOk(userConnectionManagerPath)
	if !ok {
		return nil
	}

	l := v.([]interface{})
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})
	ucm := &mdbv1.UserConnectionManager{}

	if vf, ok := m["connection_folder_id"].(string); ok && vf != "" {
		ucm.ConnectionFolderId = vf
	}

	if vf, ok := m["secret_folder_id"].(string); ok && vf != "" {
		ucm.SecretFolderId = vf
	}

	return ucm
}

// CustomizeDiffUserConnectionManager rejects attempts to change folder IDs after creation.
// Fields omitted from HCL are not validated: that means "do not touch", not "clear the value".
func CustomizeDiffUserConnectionManager(_ context.Context, d *schema.ResourceDiff, userConnectionManagerPath string) error {
	if d.Id() == "" {
		return nil
	}

	for _, field := range []string{"connection_folder_id", "secret_folder_id"} {
		cfgVal, ok := LookupRawConfigPath(d, userConnectionManagerPath+".0."+field)
		if !ok {
			continue
		}
		if cfgVal.Type() != cty.String {
			return fmt.Errorf("CustomizeDiffUserConnectionManager: expected %s.0.%s to be string, got %s", userConnectionManagerPath, field, cfgVal.Type().FriendlyName())
		}
		stateVal, _ := d.GetChange(userConnectionManagerPath + ".0." + field)
		if cfgVal.AsString() != stateVal.(string) {
			return fmt.Errorf("%s.%s cannot be changed after user creation", userConnectionManagerPath, field)
		}
	}

	return nil
}
