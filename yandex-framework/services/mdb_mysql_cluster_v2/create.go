package mdb_mysql_cluster_v2

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
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
	hostSpecsSlice []*mysql.HostSpec,
) (*mysql.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	cfg := getConfigSpecFromState(plan)

	request := &mysql.CreateClusterRequest{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		FolderId:           mdbcommon.ExpandFolderId(ctx, plan.FolderId, providerConfig, &diags),
		NetworkId:          plan.NetworkId.ValueString(),
		Environment:        mdbcommon.ExpandEnvironment[mysql.Cluster_Environment](ctx, plan.Environment, &diags),
		Labels:             mdbcommon.ExpandLabels(ctx, plan.Labels, &diags),
		ConfigSpec:         expandConfig(ctx, cfg, &diags),
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		HostSpecs:          hostSpecsSlice,
		MaintenanceWindow: mdbcommon.ExpandClusterMaintenanceWindow[
			mysql.MaintenanceWindow,
			mysql.WeeklyMaintenanceWindow,
			mysql.AnytimeMaintenanceWindow,
			mysql.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, &diags),
		SecurityGroupIds:    mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags),
		DiskEncryptionKeyId: mdbcommon.ExpandStringWrapper(ctx, plan.DiskEncryptionKeyId, &diags),
	}
	return request, diags
}

func prepareRestoreRequest(
	ctx context.Context,
	plan *Cluster,
	providerConfig *config.State,
	hostSpecsSlice []*mysql.HostSpec,
) (*mysql.RestoreClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var restoreConf Restore
	diags.Append(plan.Restore.As(ctx, &restoreConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil, diags
	}

	var timeBackup *timestamp.Timestamp = nil

	if utils.IsPresent(restoreConf.Time) {
		time, err := mdbcommon.ParseStringToTime(restoreConf.Time.ValueString())
		if err != nil {

			diags.Append(
				diag.NewErrorDiagnostic(
					"Failed to create MySQL cluster from backup",
					fmt.Sprintf(
						"Error while parsing restore time to create MySQL Cluster from backup %v, value: %v error: %s",
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

	cfg := getConfigSpecFromState(plan)

	request := &mysql.RestoreClusterRequest{
		BackupId:           restoreConf.BackupId.ValueString(),
		Time:               timeBackup,
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		FolderId:           mdbcommon.ExpandFolderId(ctx, plan.FolderId, providerConfig, &diags),
		NetworkId:          plan.NetworkId.ValueString(),
		Environment:        mdbcommon.ExpandEnvironment[mysql.Cluster_Environment](ctx, plan.Environment, &diags),
		Labels:             mdbcommon.ExpandLabels(ctx, plan.Labels, &diags),
		ConfigSpec:         expandConfig(ctx, cfg, &diags),
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		HostSpecs:          hostSpecsSlice,
		MaintenanceWindow: mdbcommon.ExpandClusterMaintenanceWindow[
			mysql.MaintenanceWindow,
			mysql.WeeklyMaintenanceWindow,
			mysql.AnytimeMaintenanceWindow,
			mysql.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, &diags),
		SecurityGroupIds:    mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags),
		DiskEncryptionKeyId: mdbcommon.ExpandStringWrapper(ctx, plan.DiskEncryptionKeyId, &diags),
	}

	// Empty string will remove encryption when restoring
	if request.DiskEncryptionKeyId == nil {
		tflog.Warn(ctx, "Disk encryption key ID is not set. Encryption will be disabled if present in source cluster.")
		request.DiskEncryptionKeyId = wrapperspb.String("")
	}

	return request, diags
}

func getConfigSpecFromState(state *Cluster) Config {
	return Config{
		Version:                state.Version,
		Resources:              state.Resources,
		Access:                 state.Access,
		PerformanceDiagnostics: state.PerformanceDiagnostics,
		BackupRetainPeriodDays: state.BackupRetainPeriodDays,
		BackupWindowStart:      state.BackupWindowStart,
		MySQLConfig:            state.MySQLConfig,
	}
}
