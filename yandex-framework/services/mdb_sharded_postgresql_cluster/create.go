package mdb_sharded_postgresql_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func prepareCreateRequest(ctx context.Context, plan *Cluster, providerConfig *config.State) (*spqr.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	configSpec := Config{}
	diags.Append(plan.Config.As(ctx, &configSpec, datasize.DefaultOpts)...)

	request := &spqr.CreateClusterRequest{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		FolderId:           mdbcommon.ExpandFolderId(ctx, plan.FolderId, providerConfig, &diags),
		NetworkId:          plan.NetworkId.ValueString(),
		Environment:        mdbcommon.ExpandEnvironment[spqr.Cluster_Environment](ctx, plan.Environment, &diags),
		Labels:             mdbcommon.ExpandLabels(ctx, plan.Labels, &diags),
		ConfigSpec:         expandConfig(ctx, configSpec, &diags),
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		SecurityGroupIds:   mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags),
		/*MaintenanceWindow: mdbcommon.ExpandClusterMaintenanceWindow[
			sharded_postgresql.MaintenanceWindow,
			sharded_postgresql.WeeklyMaintenanceWindow,
			sharded_postgresql.AnytimeMaintenanceWindow,
			sharded_postgresql.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, &diags),*/
	}
	return request, diags
}
