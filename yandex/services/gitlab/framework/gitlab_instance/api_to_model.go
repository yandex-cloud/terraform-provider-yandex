package gitlab_instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/gitlab/v1"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
)

func InstanceToState(ctx context.Context, instance *gitlab.Instance, state *InstanceModel) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("instanceToState: Gitlab instance state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("instanceToState: Received Gitlab instance data: %+v", instance))

	state.FolderId = types.StringValue(instance.GetFolderId())
	state.CreatedAt = types.StringValue(timestamp.Get(instance.GetCreatedAt()))
	state.UpdatedAt = types.StringValue(timestamp.Get(instance.GetUpdatedAt()))
	state.Name = types.StringValue(instance.GetName())

	newDescription := types.StringValue(instance.GetDescription())
	if !stringsAreEqual(state.Description, newDescription) {
		state.Description = newDescription
	}

	labels, diags := types.MapValueFrom(ctx, types.StringType, instance.Labels)
	if diags.HasError() {
		return diags
	}
	if !labels.Equal(state.Labels) {
		state.Labels = labels
	}

	state.ResourcePresetId = types.StringValue(instance.GetResourcePresetId())
	state.DiskSize = types.Int64Value(datasize.ToGigabytes(instance.GetDiskSize()))
	state.Status = types.StringValue(instance.GetStatus().String())
	state.AdminEmail = types.StringValue(instance.GetAdminEmail())
	state.AdminLogin = types.StringValue(instance.GetAdminLogin())
	state.Domain = types.StringValue(instance.GetDomain())
	state.SubnetId = types.StringValue(instance.GetSubnetId())

	updatedBackupRetainPeriodDays := types.Int64Value(instance.GetBackupRetainPeriodDays())
	if !int64AreEqual(state.BackupRetainPeriodDays, updatedBackupRetainPeriodDays) {
		state.BackupRetainPeriodDays = updatedBackupRetainPeriodDays
	}
	state.MaintenanceDeleteUntagged = types.BoolValue(instance.GetMaintenanceDeleteUntagged())
	state.DeletionProtection = types.BoolValue(instance.GetDeletionProtection())

	updatedApprovalRulesId := types.StringValue(instance.GetApprovalRulesId())
	if !updatedApprovalRulesId.Equal(state.ApprovalRulesId) {
		state.ApprovalRulesId = updatedApprovalRulesId
	}

	state.GitlabVersion = types.StringValue(instance.GetGitlabVersion())
	return diags
}
