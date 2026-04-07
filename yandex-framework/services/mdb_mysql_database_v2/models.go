package mdb_mysql_database_v2

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
)

type Database struct {
	Id                     types.String   `tfsdk:"id"`
	ClusterID              types.String   `tfsdk:"cluster_id"`
	Name                   types.String   `tfsdk:"name"`
	DeletionProtectionMode types.String   `tfsdk:"deletion_protection_mode"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
}

func specToState(spec *mysql.Database, state *Database) {
	state.Id = types.StringValue(resourceid.Construct(spec.ClusterId, spec.Name))
	state.ClusterID = types.StringValue(spec.ClusterId)
	state.Name = types.StringValue(spec.Name)
	state.DeletionProtectionMode = types.StringValue(spec.DeletionProtectionMode.String())
}

func stateToSpec(state *Database, spec *mysql.DatabaseSpec) {
	spec.Name = state.Name.ValueString()
	spec.DeletionProtectionMode = getDeletionProtectionModeValue(state.DeletionProtectionMode)
}

func getDeletionProtectionModeValue(mode types.String) mysql.DeletionProtectionMode {
	if mode.IsNull() || mode.IsUnknown() {
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_DISABLED
	}

	switch mode.ValueString() {
	case "DELETION_PROTECTION_MODE_ENABLED":
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_ENABLED
	case "DELETION_PROTECTION_MODE_DISABLED":
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_DISABLED
	case "DELETION_PROTECTION_MODE_INHERITED":
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_INHERITED
	default:
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_DISABLED
	}
}
