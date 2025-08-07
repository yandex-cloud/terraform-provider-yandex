package mdb_postgresql_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func prepareCreateRequest(ctx context.Context, plan *Cluster, providerConfig *config.State) (*postgresql.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	request := &postgresql.CreateClusterRequest{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		FolderId:           expandFolderId(ctx, plan.FolderId, providerConfig, &diags),
		NetworkId:          plan.NetworkId.ValueString(),
		Environment:        expandEnvironment(ctx, plan.Environment, &diags),
		Labels:             mdbcommon.ExpandLabels(ctx, plan.Labels, &diags),
		ConfigSpec:         expandConfig(ctx, plan.Config, &diags),
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		SecurityGroupIds:   expandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags),
		MaintenanceWindow: mdbcommon.ExpandClusterMaintenanceWindow[
			postgresql.MaintenanceWindow,
			postgresql.WeeklyMaintenanceWindow,
			postgresql.AnytimeMaintenanceWindow,
			postgresql.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, &diags),
		DiskEncryptionKeyId: expandStringWrapper(ctx, plan.DiskEncryptionKeyId, &diags),
	}
	return request, diags
}
