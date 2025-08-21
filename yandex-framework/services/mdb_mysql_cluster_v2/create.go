package mdb_mysql_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func prepareCreateRequest(ctx context.Context, plan *Cluster, providerConfig *config.State) (*mysql.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	cfg := getConfigSpecFromState(ctx, plan, &diags)

	request := &mysql.CreateClusterRequest{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		FolderId:           mdbcommon.ExpandFolderId(ctx, plan.FolderId, providerConfig, &diags),
		NetworkId:          plan.NetworkId.ValueString(),
		Environment:        mdbcommon.ExpandEnvironment[mysql.Cluster_Environment](ctx, plan.Environment, &diags),
		Labels:             mdbcommon.ExpandLabels(ctx, plan.Labels, &diags),
		ConfigSpec:         expandConfig(ctx, cfg, &diags),
		DeletionProtection: plan.DeletionProtection.ValueBool(),
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

func getConfigSpecFromState(ctx context.Context, state *Cluster, diags *diag.Diagnostics) Config {
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
