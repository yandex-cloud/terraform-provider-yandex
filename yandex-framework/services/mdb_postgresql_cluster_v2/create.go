package mdb_postgresql_cluster_v2

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func prepareCreateRequest(
	ctx context.Context,
	plan *Cluster,
	providerConfig *config.State,
	hostSpecsSlice []*postgresql.HostSpec,
) (*postgresql.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	request := &postgresql.CreateClusterRequest{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		FolderId:           expandFolderId(ctx, plan.FolderId, providerConfig, &diags),
		NetworkId:          plan.NetworkId.ValueString(),
		Environment:        mdbcommon.ExpandEnvironment[postgresql.Cluster_Environment](ctx, plan.Environment, &diags),
		Labels:             mdbcommon.ExpandLabels(ctx, plan.Labels, &diags),
		HostSpecs:          hostSpecsSlice,
		ConfigSpec:         expandConfig(ctx, plan.Config, &diags),
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		SecurityGroupIds:   mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags),
		MaintenanceWindow: mdbcommon.ExpandClusterMaintenanceWindow[
			postgresql.MaintenanceWindow,
			postgresql.WeeklyMaintenanceWindow,
			postgresql.AnytimeMaintenanceWindow,
			postgresql.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, &diags),
		DiskEncryptionKeyId: mdbcommon.ExpandStringWrapper(ctx, plan.DiskEncryptionKeyId, &diags),
	}
	return request, diags
}

func prepareRestoreRequest(
	ctx context.Context,
	plan *Cluster,
	providerConfig *config.State,
	hostSpecsSlice []*postgresql.HostSpec,
) (*postgresql.RestoreClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var restoreConf Restore

	diags.Append(plan.Restore.As(ctx, &restoreConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil, diags
	}

	var timeBackup *timestamp.Timestamp = nil
	timeInclusive := false

	if utils.IsPresent(restoreConf.Time) {
		time, err := mdbcommon.ParseStringToTime(restoreConf.Time.ValueString())
		if err != nil {

			diags.Append(
				diag.NewErrorDiagnostic(
					"Failed to create PostgreSQL cluster from backup",
					fmt.Sprintf(
						"Error while parsing restore time to create PostgreSQL Cluster from backup %v, value: %v error: %s",
						restoreConf.BackupId,
						restoreConf.Time,
						err.Error(),
					),
				),
			)
		}
		timeBackup = &timestamp.Timestamp{
			Seconds: time.Unix(),
		}
	}

	if utils.IsPresent(restoreConf.TimeInclusive) {
		timeInclusive = restoreConf.TimeInclusive.ValueBool()
	}

	request := &postgresql.RestoreClusterRequest{
		BackupId:           restoreConf.BackupId.ValueString(),
		Time:               timeBackup,
		TimeInclusive:      timeInclusive,
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		FolderId:           expandFolderId(ctx, plan.FolderId, providerConfig, &diags),
		NetworkId:          plan.NetworkId.ValueString(),
		Environment:        mdbcommon.ExpandEnvironment[postgresql.Cluster_Environment](ctx, plan.Environment, &diags),
		Labels:             mdbcommon.ExpandLabels(ctx, plan.Labels, &diags),
		HostSpecs:          hostSpecsSlice,
		ConfigSpec:         expandConfig(ctx, plan.Config, &diags),
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		SecurityGroupIds:   mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags),
		MaintenanceWindow: mdbcommon.ExpandClusterMaintenanceWindow[
			postgresql.MaintenanceWindow,
			postgresql.WeeklyMaintenanceWindow,
			postgresql.AnytimeMaintenanceWindow,
			postgresql.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, &diags),
		DiskEncryptionKeyId: mdbcommon.ExpandStringWrapper(ctx, plan.DiskEncryptionKeyId, &diags),
	}

	// Empty string will remove encryption when restoring
	if request.DiskEncryptionKeyId == nil {
		tflog.Warn(ctx, "Disk encryption key ID is not set. Encryption will be disabled if present in source cluster.")
		request.DiskEncryptionKeyId = wrapperspb.String("")
	}
	return request, diags
}
