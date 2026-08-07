package mdb_redis_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	redisproto "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
)

type clusterReadModel interface {
	commonCluster() *clusterModel
	commonConfig() *configModel
}

func clusterRead(ctx context.Context, sdk *ycsdk.SDK, diagnostics *diag.Diagnostics, state clusterReadModel) {
	clusterState := state.commonCluster()
	configState := state.commonConfig()
	cid := clusterState.ID.ValueString()
	cluster := redisAPI.GetCluster(ctx, sdk, diagnostics, cid)
	if diagnostics.HasError() {
		return
	}

	clusterState.ClusterID = clusterState.ID
	clusterState.Name = types.StringValue(cluster.Name)
	clusterState.NetworkID = types.StringValue(cluster.NetworkId)
	clusterState.Environment = types.StringValue(cluster.GetEnvironment().String())
	clusterState.Description = types.StringValue(cluster.Description)
	clusterState.Sharded = types.BoolValue(cluster.Sharded)
	clusterState.TlsEnabled = types.BoolValue(cluster.TlsEnabled)
	clusterState.PersistenceMode = types.StringValue(cluster.GetPersistenceMode().String())
	clusterState.AnnounceHostnames = types.BoolValue(cluster.AnnounceHostnames)
	clusterState.FolderID = types.StringValue(cluster.FolderId)
	clusterState.CreatedAt = types.StringValue(timestamp.Get(cluster.CreatedAt))
	clusterState.DeletionProtection = types.BoolValue(cluster.DeletionProtection)
	clusterState.AuthSentinel = types.BoolValue(cluster.AuthSentinel)
	clusterState.DiskEncryptionKeyId = mdbcommon.FlattenStringWrapper(ctx, cluster.DiskEncryptionKeyId, diagnostics)

	clusterLabels := cluster.Labels
	if clusterLabels == nil {
		clusterLabels = make(map[string]string)
	}

	labels, diags := types.MapValueFrom(ctx, types.StringType, clusterLabels)
	clusterState.Labels = labels
	diagnostics.Append(diags...)

	clusterState.SecurityGroupIDs = mdbcommon.FlattenSetString(ctx, cluster.SecurityGroupIds, diagnostics)

	clusterState.Resources = mdbcommon.FlattenResources[redisproto.Resources](ctx, cluster.GetConfig().GetResources(), diagnostics)

	*configState = configToState(ctx, cluster.Config, *configState, diagnostics)

	clusterState.DiskSizeAutoscaling, diags = flattenAutoscaling(ctx, cluster.GetConfig().GetDiskSizeAutoscaling())
	diagnostics.Append(diags...)

	clusterState.Modules, diags = flattenModules(ctx, cluster.GetConfig().GetModules())
	diagnostics.Append(diags...)

	clusterState.MaintenanceWindow = mdbcommon.FlattenMaintenanceWindow[
		redis.MaintenanceWindow,
		redis.WeeklyMaintenanceWindow,
		redis.AnytimeMaintenanceWindow,
		redis.WeeklyMaintenanceWindow_WeekDay,
	](ctx, cluster.MaintenanceWindow, diagnostics)

	clusterState.Access, diags = flattenAccess(ctx, cluster.GetConfig().GetAccess())
	diagnostics.Append(diags...)

	var entityIdToApiHosts map[string]Host = mdbcommon.ReadHosts[Host, *redisproto.Host, *redisproto.HostSpec, redisproto.UpdateHostSpec](ctx, sdk, diagnostics, redisHostService, &redisAPI, clusterState.HostSpecs, cid)

	clusterState.HostSpecs, diags = types.MapValueFrom(ctx, HostType, entityIdToApiHosts)

	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}
}
