package mdb_redis_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func configToState(ctx context.Context, apiConfig *redis.ClusterConfig, current configModel, diagnostics *diag.Diagnostics) configModel {
	config := FlattenConfig(apiConfig)
	config.Password = current.Password
	config.BackupWindowStart = mdbcommon.FlattenBackupWindowStart(ctx, apiConfig.GetBackupWindowStart(), diagnostics)
	config.BackupRetainPeriodDays = types.Int64Value(apiConfig.BackupRetainPeriodDays.GetValue())
	return config.configModel
}
