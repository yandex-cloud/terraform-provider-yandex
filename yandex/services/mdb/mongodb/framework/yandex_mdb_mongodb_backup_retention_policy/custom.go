package yandex_mdb_mongodb_backup_retention_policy

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// customBackupRetentionPolicyImporter implements custom import logic for composite ID format: cluster_id:policy_id
func customBackupRetentionPolicyImporter(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: cluster_id:policy_id, got: %s", req.ID),
		)
		return
	}

	clusterId := strings.TrimSpace(parts[0])
	policyId := strings.TrimSpace(parts[1])

	if clusterId == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"cluster_id cannot be empty",
		)
		return
	}

	if policyId == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"policy_id cannot be empty",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), clusterId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), policyId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), policyId)...)
}
