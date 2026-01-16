package mdb_clickhouse_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"
)

// Create cluster

func (r *clusterResource) createCluster(
	ctx context.Context,
	plan *models.Cluster,
	hostSpecsSlice []*clickhouse.HostSpec,
	diags *diag.Diagnostics,
) {
	tflog.Debug(ctx, "Creating ClickHouse cluster")

	request := prepareClusterCreateRequest(ctx, plan, &r.providerConfig.ProviderState, diags, hostSpecsSlice)
	if diags.HasError() {
		return
	}

	cid := clickhouseApi.CreateCluster(ctx, r.providerConfig.SDK, diags, request)
	if diags.HasError() {
		return
	}
	plan.Id = types.StringValue(cid)
}

func prepareClusterCreateRequest(
	ctx context.Context,
	plan *models.Cluster,
	providerConfig *config.State,
	diags *diag.Diagnostics,
	hostSpecsSlice []*clickhouse.HostSpec,
) *clickhouse.CreateClusterRequest {
	return &clickhouse.CreateClusterRequest{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		FolderId:         mdbcommon.ExpandFolderId(ctx, plan.FolderId, providerConfig, diags),
		NetworkId:        plan.NetworkId.ValueString(),
		Environment:      mdbcommon.ExpandEnvironment[clickhouse.Cluster_Environment](ctx, plan.Environment, diags),
		Labels:           mdbcommon.ExpandLabels(ctx, plan.Labels, diags),
		HostSpecs:        hostSpecsSlice,
		ShardSpecs:       models.ExpandListShard(ctx, plan.Shards, plan.Id.ValueString(), diags),
		ServiceAccountId: plan.ServiceAccountId.ValueString(),
		ConfigSpec: &clickhouse.ConfigSpec{
			Version:                plan.Version.ValueString(),
			Clickhouse:             models.ExpandClickHouse(ctx, plan.ClickHouse, diags),
			Zookeeper:              models.ExpandZooKeeper(ctx, plan.ZooKeeper, diags),
			BackupWindowStart:      mdbcommon.ExpandBackupWindow(ctx, plan.BackupWindowStart, diags),
			Access:                 models.ExpandAccess(ctx, plan.Access, diags),
			CloudStorage:           models.ExpandCloudStorage(ctx, plan.CloudStorage, diags),
			SqlDatabaseManagement:  mdbcommon.ExpandBoolWrapper(ctx, plan.SqlDatabaseManagement, diags),
			SqlUserManagement:      mdbcommon.ExpandBoolWrapper(ctx, plan.SqlUserManagement, diags),
			AdminPassword:          plan.AdminPassword.ValueString(),
			EmbeddedKeeper:         mdbcommon.ExpandBoolWrapper(ctx, plan.EmbeddedKeeper, diags),
			BackupRetainPeriodDays: mdbcommon.ExpandInt64Wrapper(ctx, plan.BackupRetainPeriodDays, diags),
		},
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		SecurityGroupIds:   mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, diags),
		MaintenanceWindow: mdbcommon.ExpandClusterMaintenanceWindow[
			clickhouse.MaintenanceWindow,
			clickhouse.WeeklyMaintenanceWindow,
			clickhouse.AnytimeMaintenanceWindow,
			clickhouse.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, diags),
		DiskEncryptionKeyId: mdbcommon.ExpandStringWrapper(ctx, plan.DiskEncryptionKeyId, diags),
	}
}

// Create format schemas

func (r *clusterResource) createFormatSchemas(ctx context.Context, plan models.Cluster, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "Creating ClickHouse format schemas")

	requests := prepareFormatSchemasCreateRequests(ctx, &plan, diags)
	if diags.HasError() {
		return
	}

	for _, request := range requests {
		clickhouseApi.CreateFormatSchema(ctx, r.providerConfig.SDK, diags, request)
		if diags.HasError() {
			return
		}
	}
}

func prepareFormatSchemasCreateRequests(ctx context.Context, plan *models.Cluster, diags *diag.Diagnostics) []*clickhouse.CreateFormatSchemaRequest {
	cid := plan.Id.ValueString()
	var requests []*clickhouse.CreateFormatSchemaRequest

	formatSchemas := models.ExpandListFormatSchema(ctx, plan.FormatSchema, cid, diags)
	if diags.HasError() {
		return requests
	}

	for _, formatSchema := range formatSchemas {
		typeValue := utils.ExpandEnum("type", formatSchema.Type.Enum().String(), clickhouse.FormatSchemaType_value, diags)
		if diags.HasError() {
			return requests
		}

		requests = append(requests, &clickhouse.CreateFormatSchemaRequest{
			ClusterId:        cid,
			FormatSchemaName: formatSchema.Name,
			Type:             clickhouse.FormatSchemaType(*typeValue),
			Uri:              formatSchema.Uri,
		})
	}

	return requests
}

// Create ML models

func (r *clusterResource) createMlModels(ctx context.Context, plan models.Cluster, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "Creating ClickHouse ml models")

	requests := prepareMlModelsCreateRequests(ctx, &plan, diags)
	if diags.HasError() {
		return
	}

	for _, request := range requests {
		clickhouseApi.CreateMlModel(ctx, r.providerConfig.SDK, diags, request)
		if diags.HasError() {
			return
		}
	}
}

func prepareMlModelsCreateRequests(ctx context.Context, plan *models.Cluster, diags *diag.Diagnostics) []*clickhouse.CreateMlModelRequest {
	cid := plan.Id.ValueString()
	var requests []*clickhouse.CreateMlModelRequest

	mlModels := models.ExpandListMLModel(ctx, plan.MLModel, cid, diags)
	if diags.HasError() {
		return requests
	}

	for _, mlModel := range mlModels {
		typeValue := utils.ExpandEnum("type", mlModel.Type.Enum().String(), clickhouse.MlModelType_value, diags)
		if diags.HasError() {
			return requests
		}

		requests = append(requests, &clickhouse.CreateMlModelRequest{
			ClusterId:   cid,
			MlModelName: mlModel.Name,
			Type:        clickhouse.MlModelType(*typeValue),
			Uri:         mlModel.Uri,
		})
	}

	return requests
}

// Create shard groups

func (r *clusterResource) createShardGroups(ctx context.Context, plan models.Cluster, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "Creating ClickHouse shard groups")

	requests := prepareShardGroupsCreateRequests(ctx, &plan, diags)
	if diags.HasError() {
		return
	}

	for _, request := range requests {
		clickhouseApi.CreateShardGroup(ctx, r.providerConfig.SDK, diags, request)
		if diags.HasError() {
			return
		}
	}
}

func prepareShardGroupsCreateRequests(ctx context.Context, plan *models.Cluster, diags *diag.Diagnostics) []*clickhouse.CreateClusterShardGroupRequest {
	cid := plan.Id.ValueString()
	var requests []*clickhouse.CreateClusterShardGroupRequest

	shardGroups := models.ExpandListShardGroup(ctx, plan.ShardGroup, cid, diags)
	if diags.HasError() {
		return requests
	}

	for _, shardGroup := range shardGroups {
		requests = append(requests, &clickhouse.CreateClusterShardGroupRequest{
			ClusterId:      cid,
			ShardGroupName: shardGroup.Name,
			Description:    shardGroup.Description,
			ShardNames:     shardGroup.ShardNames,
		})
	}

	return requests
}
