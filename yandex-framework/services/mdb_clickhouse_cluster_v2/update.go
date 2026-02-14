package mdb_clickhouse_cluster_v2

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Folder id

func prepareFolderIdUpdateRequest(state, plan *models.Cluster) *clickhouse.MoveClusterRequest {
	if state.FolderId.Equal(plan.FolderId) {
		return nil
	}

	return &clickhouse.MoveClusterRequest{
		ClusterId:           state.Id.ValueString(),
		DestinationFolderId: plan.FolderId.ValueString(),
	}
}

// Version

func prepareVersionUpdateRequest(state, plan *models.Cluster) *clickhouse.UpdateClusterRequest {
	if state.Version.Equal(plan.Version) {
		return nil
	}

	return &clickhouse.UpdateClusterRequest{
		ClusterId: state.Id.ValueString(),
		ConfigSpec: &clickhouse.ConfigSpec{
			Version: plan.Version.ValueString(),
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"config_spec.version"},
		},
	}
}

// Cluster

func prepareClusterUpdateRequest(ctx context.Context, state, plan *models.Cluster, diags *diag.Diagnostics) *clickhouse.UpdateClusterRequest {
	request := &clickhouse.UpdateClusterRequest{
		ClusterId:  state.Id.ValueString(),
		UpdateMask: &field_mask.FieldMask{},
	}

	if !plan.Name.Equal(state.Name) {
		request.SetName(plan.Name.ValueString())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "name")
	}

	if !plan.Description.Equal(state.Description) {
		request.SetDescription(plan.Description.ValueString())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "description")
	}

	if !plan.Labels.Equal(state.Labels) {
		request.SetLabels(mdbcommon.ExpandLabels(ctx, plan.Labels, diags))
		if diags.HasError() {
			return nil
		}
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "labels")
	}

	config, updateMaskPaths := prepareClusterConfigSpec(ctx, plan, state, diags)
	if diags.HasError() {
		return nil
	}

	request.SetConfigSpec(config)
	request.UpdateMask.Paths = append(request.UpdateMask.Paths, updateMaskPaths...)

	if !plan.DeletionProtection.Equal(state.DeletionProtection) {
		request.SetDeletionProtection(plan.DeletionProtection.ValueBool())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "deletion_protection")
	}

	if !plan.SecurityGroupIds.Equal(state.SecurityGroupIds) {
		request.SetSecurityGroupIds(mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, diags))
		if diags.HasError() {
			return nil
		}
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "security_group_ids")
	}

	if !plan.MaintenanceWindow.Equal(state.MaintenanceWindow) {
		request.SetMaintenanceWindow(mdbcommon.ExpandClusterMaintenanceWindow[
			clickhouse.MaintenanceWindow,
			clickhouse.WeeklyMaintenanceWindow,
			clickhouse.AnytimeMaintenanceWindow,
			clickhouse.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, diags))
		if diags.HasError() {
			return nil
		}
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "maintenance_window")
	}

	if len(request.UpdateMask.Paths) == 0 {
		return nil
	} else {
		return request
	}
}

// Cluster config

func prepareClusterConfigSpec(ctx context.Context, plan, state *models.Cluster, diags *diag.Diagnostics) (*clickhouse.ConfigSpec, []string) {
	var updateMaskPaths []string
	config := &clickhouse.ConfigSpec{}

	if !plan.ClickHouse.Equal(state.ClickHouse) {
		config.SetClickhouse(models.ExpandClickHouse(ctx, plan.ClickHouse, diags))
		if diags.HasError() {
			return config, updateMaskPaths
		}

		var planClickHouse, stateClickHouse models.Clickhouse
		diags.Append(plan.ClickHouse.As(ctx, &planClickHouse, datasize.UnhandledOpts)...)
		diags.Append(state.ClickHouse.As(ctx, &stateClickHouse, datasize.UnhandledOpts)...)
		if diags.HasError() {
			return config, updateMaskPaths
		}

		if !planClickHouse.Resources.IsUnknown() && !planClickHouse.Resources.Equal(stateClickHouse.Resources) {
			updateMaskPaths = append(
				updateMaskPaths,
				"config_spec.clickhouse.resources",
			)
		}

		if !planClickHouse.DiskSizeAutoscaling.IsUnknown() && !planClickHouse.DiskSizeAutoscaling.Equal(stateClickHouse.DiskSizeAutoscaling) {
			updateMaskPaths = append(
				updateMaskPaths,
				"config_spec.clickhouse.disk_size_autoscaling",
			)
		}

		// Get update paths for clickhouse config
		updateMaskPaths = append(updateMaskPaths, getClickHouseConfigUpdatePaths(ctx, planClickHouse, stateClickHouse, diags)...)
		if diags.HasError() {
			return config, updateMaskPaths
		}
	}

	// We can update ZK only when it exists
	hasZooKeeper := false
	if !state.ZooKeeper.IsNull() && !state.ZooKeeper.IsUnknown() {
		var stateZooKeeper models.Zookeeper
		diags.Append(state.ZooKeeper.As(ctx, &stateZooKeeper, datasize.DefaultOpts)...)
		if diags.HasError() {
			return config, updateMaskPaths
		}

		hasZooKeeper = stateZooKeeper.IsConfigured(ctx, diags)
		if diags.HasError() {
			return config, updateMaskPaths
		}
	}

	if hasZooKeeper && !plan.ZooKeeper.Equal(state.ZooKeeper) {
		config.SetZookeeper(models.ExpandZooKeeper(ctx, plan.ZooKeeper, diags))
		if diags.HasError() {
			return config, updateMaskPaths
		}
		updateMaskPaths = append(
			updateMaskPaths,
			"config_spec.zookeeper",
		)
	}

	if !plan.BackupWindowStart.Equal(state.BackupWindowStart) {
		config.SetBackupWindowStart(mdbcommon.ExpandBackupWindow(ctx, plan.BackupWindowStart, diags))
		if diags.HasError() {
			return config, updateMaskPaths
		}
		updateMaskPaths = append(updateMaskPaths, "config_spec.backup_window_start")
	}

	if !plan.Access.Equal(state.Access) {
		config.SetAccess(models.ExpandAccess(ctx, plan.Access, diags))
		if diags.HasError() {
			return config, updateMaskPaths
		}

		updateMaskPaths = appendNestedConfigUpdatePaths(
			updateMaskPaths,
			plan.Access.Attributes(),
			state.Access.Attributes(),
			models.AccessAttrTypes,
			"config_spec.access.",
		)
	}

	if !plan.CloudStorage.Equal(state.CloudStorage) {
		config.SetCloudStorage(models.ExpandCloudStorage(ctx, plan.CloudStorage, diags))
		if diags.HasError() {
			return config, updateMaskPaths
		}

		updateMaskPaths = appendNestedConfigUpdatePaths(
			updateMaskPaths,
			plan.CloudStorage.Attributes(),
			state.CloudStorage.Attributes(),
			models.CloudStorageAttrTypes,
			"config_spec.cloud_storage.",
		)
	}

	if !plan.SqlDatabaseManagement.Equal(state.SqlDatabaseManagement) {
		config.SetSqlDatabaseManagement(&wrapperspb.BoolValue{Value: plan.SqlDatabaseManagement.ValueBool()})
		updateMaskPaths = append(updateMaskPaths, "config_spec.sql_database_management")
	}

	if !plan.SqlUserManagement.Equal(state.SqlUserManagement) {
		config.SetSqlUserManagement(&wrapperspb.BoolValue{Value: plan.SqlUserManagement.ValueBool()})
		updateMaskPaths = append(updateMaskPaths, "config_spec.sql_user_management")
	}

	if !plan.EmbeddedKeeper.Equal(state.EmbeddedKeeper) {
		config.SetEmbeddedKeeper(&wrapperspb.BoolValue{Value: plan.EmbeddedKeeper.ValueBool()})
		updateMaskPaths = append(updateMaskPaths, "config_spec.embedded_keeper")
	}

	if !plan.BackupRetainPeriodDays.Equal(state.BackupRetainPeriodDays) {
		config.SetBackupRetainPeriodDays(mdbcommon.ExpandInt64Wrapper(ctx, plan.BackupRetainPeriodDays, diags))
		if diags.HasError() {
			return config, updateMaskPaths
		}
		updateMaskPaths = append(updateMaskPaths, "config_spec.backup_retain_period_days")
	}

	return config, updateMaskPaths
}

func getClickHouseConfigUpdatePaths(ctx context.Context, planClickHouse, stateClickHouse models.Clickhouse, diags *diag.Diagnostics) []string {
	var updateMaskPaths []string

	if !planClickHouse.Config.Equal(stateClickHouse.Config) {
		var planClickHouseConfig, stateClickHouseConfig models.ClickhouseConfig
		diags.Append(planClickHouse.Config.As(ctx, &planClickHouseConfig, datasize.UnhandledOpts)...)
		diags.Append(stateClickHouse.Config.As(ctx, &stateClickHouseConfig, datasize.UnhandledOpts)...)
		if diags.HasError() {
			return updateMaskPaths
		}

		for setting := range models.ClickhouseConfigAttrTypes {
			switch setting {
			case "merge_tree":
				updateMaskPaths = appendNestedConfigUpdatePaths(
					updateMaskPaths,
					planClickHouseConfig.MergeTree.Attributes(),
					stateClickHouseConfig.MergeTree.Attributes(),
					models.MergeTreeConfigAttrTypes,
					"config_spec.clickhouse.config.merge_tree.",
				)
			case "access_control_improvements":
				updateMaskPaths = appendNestedConfigUpdatePaths(
					updateMaskPaths,
					planClickHouseConfig.AccessControlImprovements.Attributes(),
					stateClickHouseConfig.AccessControlImprovements.Attributes(),
					models.AccessControlImprovementsAttrTypes,
					"config_spec.clickhouse.config.access_control_improvements.",
				)
			case "kafka":
				updateMaskPaths = appendNestedConfigUpdatePaths(
					updateMaskPaths,
					planClickHouseConfig.Kafka.Attributes(),
					stateClickHouseConfig.Kafka.Attributes(),
					models.KafkaAttrTypes,
					"config_spec.clickhouse.config.kafka.",
				)
			case "rabbitmq":
				updateMaskPaths = appendNestedConfigUpdatePaths(
					updateMaskPaths,
					planClickHouseConfig.Rabbitmq.Attributes(),
					stateClickHouseConfig.Rabbitmq.Attributes(),
					models.RabbitmqAttrTypes,
					"config_spec.clickhouse.config.rabbitmq.",
				)
			case "query_cache":
				updateMaskPaths = appendNestedConfigUpdatePaths(
					updateMaskPaths,
					planClickHouseConfig.QueryCache.Attributes(),
					stateClickHouseConfig.QueryCache.Attributes(),
					models.QueryCacheAttrTypes,
					"config_spec.clickhouse.config.query_cache.",
				)
			case "jdbc_bridge":
				updateMaskPaths = appendNestedConfigUpdatePaths(
					updateMaskPaths,
					planClickHouseConfig.JdbcBridge.Attributes(),
					stateClickHouseConfig.JdbcBridge.Attributes(),
					models.JdbcBridgeAttrTypes,
					"config_spec.clickhouse.config.jdbc_bridge.",
				)
			default:
				planVal := planClickHouse.Config.Attributes()[setting]
				stateVal := stateClickHouse.Config.Attributes()[setting]
				if !planVal.Equal(stateVal) {
					updateMaskPaths = append(
						updateMaskPaths,
						"config_spec.clickhouse.config."+setting,
					)
				}
			}
		}
	}

	return updateMaskPaths
}

// Format schemas

func updateFormatSchemas(ctx context.Context, plan models.Cluster, sdk *ycsdk.SDK, diags *diag.Diagnostics) {
	cid := plan.Id.ValueString()
	currentFormatSchemas := clickhouseApi.ListFormatSchemas(ctx, sdk, diags, cid)
	if diags.HasError() {
		return
	}

	deleteFormatSchemaNames, updateFormatSchemasRequests, createFormatSchemaRequests := prepareFormatSchemaUpdateRequests(ctx, currentFormatSchemas, &plan, diags)
	if diags.HasError() {
		return
	}

	for _, name := range deleteFormatSchemaNames {
		clickhouseApi.DeleteFormatSchema(ctx, sdk, diags, cid, name)
		if diags.HasError() {
			return
		}
	}

	for _, request := range updateFormatSchemasRequests {
		clickhouseApi.UpdateFormatSchema(ctx, sdk, diags, request)
		if diags.HasError() {
			return
		}
	}

	for _, request := range createFormatSchemaRequests {
		clickhouseApi.CreateFormatSchema(ctx, sdk, diags, request)
		if diags.HasError() {
			return
		}
	}
}

func prepareFormatSchemaUpdateRequests(ctx context.Context, currentSchemas []*clickhouse.FormatSchema, plan *models.Cluster, diags *diag.Diagnostics) ([]string, []*clickhouse.UpdateFormatSchemaRequest, []*clickhouse.CreateFormatSchemaRequest) {
	targetSchemas := models.ExpandListFormatSchema(ctx, plan.FormatSchema, plan.Id.ValueString(), diags)
	if diags.HasError() {
		return nil, nil, nil
	}

	var toDelete []string
	var toUpdate []*clickhouse.FormatSchema

	mapTargetSchemaName := map[string]*clickhouse.FormatSchema{}
	for _, schema := range targetSchemas {
		mapTargetSchemaName[schema.Name] = schema
	}

	for _, currentSchema := range currentSchemas {
		if targetSchema, ok := mapTargetSchemaName[currentSchema.Name]; ok {
			if currentSchema.Type != targetSchema.Type {
				toDelete = append(toDelete, currentSchema.Name)
			} else {
				if currentSchema.Uri != targetSchema.Uri {
					toUpdate = append(toUpdate, targetSchema)
				}
				delete(mapTargetSchemaName, currentSchema.Name)
			}
		} else {
			toDelete = append(toDelete, currentSchema.Name)
		}
	}

	var updateRequests []*clickhouse.UpdateFormatSchemaRequest
	for _, schema := range toUpdate {
		updateRequests = append(updateRequests, &clickhouse.UpdateFormatSchemaRequest{
			ClusterId:        plan.Id.ValueString(),
			FormatSchemaName: schema.Name,
			Uri:              schema.Uri,
			UpdateMask:       &field_mask.FieldMask{Paths: []string{"uri"}},
		})
	}

	var createRequests []*clickhouse.CreateFormatSchemaRequest
	for _, schema := range mapTargetSchemaName {
		createRequests = append(createRequests, &clickhouse.CreateFormatSchemaRequest{
			ClusterId:        plan.Id.ValueString(),
			FormatSchemaName: schema.Name,
			Type:             schema.Type,
			Uri:              schema.Uri,
		})
	}

	return toDelete, updateRequests, createRequests
}

// Ml models

func updateMlModels(ctx context.Context, plan models.Cluster, sdk *ycsdk.SDK, diags *diag.Diagnostics) {
	cid := plan.Id.ValueString()
	currentMlModels := clickhouseApi.ListMlModels(ctx, sdk, diags, cid)
	if diags.HasError() {
		return
	}

	deleteMlModelNames, updateMlModelRequests, createMlModelRequests := prepareMlModelUpdateRequests(ctx, currentMlModels, &plan, diags)
	if diags.HasError() {
		return
	}

	for _, name := range deleteMlModelNames {
		clickhouseApi.DeleteMlModel(ctx, sdk, diags, cid, name)
		if diags.HasError() {
			return
		}
	}

	for _, request := range updateMlModelRequests {
		clickhouseApi.UpdateMlModel(ctx, sdk, diags, request)
		if diags.HasError() {
			return
		}
	}

	for _, request := range createMlModelRequests {
		clickhouseApi.CreateMlModel(ctx, sdk, diags, request)
		if diags.HasError() {
			return
		}
	}
}

func prepareMlModelUpdateRequests(ctx context.Context, currentModels []*clickhouse.MlModel, plan *models.Cluster, diags *diag.Diagnostics) ([]string, []*clickhouse.UpdateMlModelRequest, []*clickhouse.CreateMlModelRequest) {
	targetModels := models.ExpandListMLModel(ctx, plan.MLModel, plan.Id.ValueString(), diags)
	if diags.HasError() {
		return nil, nil, nil
	}

	var toDelete []string
	var toUpdate []*clickhouse.MlModel

	mapTargetModelName := map[string]*clickhouse.MlModel{}
	for _, model := range targetModels {
		mapTargetModelName[model.Name] = model
	}

	for _, currentModel := range currentModels {
		if targetModel, ok := mapTargetModelName[currentModel.Name]; ok {
			if currentModel.Type != targetModel.Type {
				toDelete = append(toDelete, currentModel.Name)
			} else {
				if currentModel.Uri != targetModel.Uri {
					toUpdate = append(toUpdate, targetModel)
				}
				delete(mapTargetModelName, currentModel.Name)
			}
		} else {
			toDelete = append(toDelete, currentModel.Name)
		}
	}

	var updateRequests []*clickhouse.UpdateMlModelRequest
	for _, model := range toUpdate {
		updateRequests = append(updateRequests, &clickhouse.UpdateMlModelRequest{
			ClusterId:   plan.Id.ValueString(),
			MlModelName: model.Name,
			Uri:         model.Uri,
			UpdateMask:  &field_mask.FieldMask{Paths: []string{"uri"}},
		})
	}

	var createRequests []*clickhouse.CreateMlModelRequest
	for _, model := range mapTargetModelName {
		createRequests = append(createRequests, &clickhouse.CreateMlModelRequest{
			ClusterId:   plan.Id.ValueString(),
			MlModelName: model.Name,
			Type:        model.Type,
			Uri:         model.Uri,
		})
	}

	return toDelete, updateRequests, createRequests
}

// Shard groups

func updateShardGroups(ctx context.Context, plan models.Cluster, sdk *ycsdk.SDK, diags *diag.Diagnostics) {
	cid := plan.Id.ValueString()
	currentShardGroups := clickhouseApi.ListShardGroups(ctx, sdk, diags, cid)
	if diags.HasError() {
		return
	}

	deleteShardGroupNames, updateShardGroupRequests, createShardGroupRequests := prepareShardGroupUpdateRequests(ctx, currentShardGroups, &plan, diags)
	if diags.HasError() {
		return
	}

	for _, name := range deleteShardGroupNames {
		clickhouseApi.DeleteShardGroup(ctx, sdk, diags, cid, name)
		if diags.HasError() {
			return
		}
	}

	for _, request := range updateShardGroupRequests {
		clickhouseApi.UpdateShardGroup(ctx, sdk, diags, request)
		if diags.HasError() {
			return
		}
	}

	for _, request := range createShardGroupRequests {
		clickhouseApi.CreateShardGroup(ctx, sdk, diags, request)
		if diags.HasError() {
			return
		}
	}
}

func prepareShardGroupUpdateRequests(ctx context.Context, currentShardGroups []*clickhouse.ShardGroup, plan *models.Cluster, diags *diag.Diagnostics) ([]string, []*clickhouse.UpdateClusterShardGroupRequest, []*clickhouse.CreateClusterShardGroupRequest) {
	targetShardGroups := models.ExpandListShardGroup(ctx, plan.ShardGroup, plan.Id.ValueString(), diags)
	if diags.HasError() {
		return nil, nil, nil
	}

	var toDelete []string
	var toUpdate []*clickhouse.ShardGroup

	mapTargetShardGroupName := map[string]*clickhouse.ShardGroup{}
	for _, group := range targetShardGroups {
		mapTargetShardGroupName[group.Name] = group
	}

	for _, currentShardGroup := range currentShardGroups {
		if targetShardGroup, ok := mapTargetShardGroupName[currentShardGroup.Name]; ok {
			if currentShardGroup.Description != targetShardGroup.Description || !reflect.DeepEqual(currentShardGroup.ShardNames, targetShardGroup.ShardNames) {
				toUpdate = append(toUpdate, targetShardGroup)
			}
			delete(mapTargetShardGroupName, currentShardGroup.Name)
		} else {
			toDelete = append(toDelete, currentShardGroup.Name)
		}
	}

	var updateRequests []*clickhouse.UpdateClusterShardGroupRequest
	for _, group := range toUpdate {
		updateRequests = append(updateRequests, &clickhouse.UpdateClusterShardGroupRequest{
			ClusterId:      plan.Id.ValueString(),
			ShardGroupName: group.Name,
			Description:    group.Description,
			ShardNames:     group.ShardNames,
			UpdateMask:     &field_mask.FieldMask{Paths: []string{"description", "shard_names"}},
		})
	}

	var createRequests []*clickhouse.CreateClusterShardGroupRequest
	for _, group := range mapTargetShardGroupName {
		createRequests = append(createRequests, &clickhouse.CreateClusterShardGroupRequest{
			ClusterId:      plan.Id.ValueString(),
			ShardGroupName: group.Name,
			Description:    group.Description,
			ShardNames:     group.ShardNames,
		})
	}

	return toDelete, updateRequests, createRequests
}

// Shards

func updateShards(ctx context.Context, plan models.Cluster, sdk *ycsdk.SDK, diags *diag.Diagnostics) {
	cid := plan.Id.ValueString()

	shards := models.ExpandListShard(ctx, plan.Shards, cid, diags)
	if diags.HasError() {
		return
	}

	for _, shard := range shards {
		updateShard(ctx, cid, shard, sdk, diags)
		if diags.HasError() {
			return
		}
	}
}

func updateShard(ctx context.Context, cid string, shardSpec *clickhouse.ShardSpec, sdk *ycsdk.SDK, diags *diag.Diagnostics) {
	var updateMaskPaths []string
	currentShard := clickhouseApi.GetShard(ctx, sdk, diags, cid, shardSpec.Name)
	if diags.HasError() {
		return
	}

	planClickHouse := shardSpec.GetConfigSpec().GetClickhouse()
	stateClickHouse := currentShard.GetConfig().GetClickhouse()

	if planClickHouse.GetWeight() != nil && planClickHouse.GetWeight().GetValue() != stateClickHouse.GetWeight().GetValue() {
		updateMaskPaths = append(updateMaskPaths, "config_spec.clickhouse.weight")
	}

	planResources := planClickHouse.GetResources()
	stateResources := stateClickHouse.GetResources()

	if planResources != nil {
		if planResources.GetDiskSize() != stateResources.GetDiskSize() {
			updateMaskPaths = append(updateMaskPaths, "config_spec.clickhouse.resources.disk_size")
		}

		if planResources.GetResourcePresetId() != stateResources.GetResourcePresetId() {
			updateMaskPaths = append(updateMaskPaths, "config_spec.clickhouse.resources.resource_preset_id")
		}

		if planResources.GetDiskTypeId() != stateResources.GetDiskTypeId() {
			updateMaskPaths = append(updateMaskPaths, "config_spec.clickhouse.resources.disk_type_id")
		}
	}

	planDsa := planClickHouse.GetDiskSizeAutoscaling()
	stateDsa := stateClickHouse.GetDiskSizeAutoscaling()

	if planDsa != nil {
		if planDsa.GetDiskSizeLimit().GetValue() != stateDsa.GetDiskSizeLimit().GetValue() {
			updateMaskPaths = append(updateMaskPaths, "config_spec.clickhouse.disk_size_autoscaling.disk_size_limit")
		}

		if planDsa.GetPlannedUsageThreshold().GetValue() != stateDsa.GetPlannedUsageThreshold().GetValue() {
			updateMaskPaths = append(updateMaskPaths, "config_spec.clickhouse.disk_size_autoscaling.planned_usage_threshold")
		}

		if planDsa.GetEmergencyUsageThreshold().GetValue() != stateDsa.GetEmergencyUsageThreshold().GetValue() {
			updateMaskPaths = append(updateMaskPaths, "config_spec.clickhouse.disk_size_autoscaling.emergency_usage_threshold")
		}
	}

	if len(updateMaskPaths) == 0 {
		return
	}

	clickhouseApi.UpdateShard(ctx, sdk, diags, &clickhouse.UpdateClusterShardRequest{
		ClusterId:  cid,
		ShardName:  shardSpec.Name,
		ConfigSpec: shardSpec.ConfigSpec,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMaskPaths,
		},
	})
	if diags.HasError() {
		return
	}
}

// Utils

func appendNestedConfigUpdatePaths(
	updateMaskPaths []string,
	planAttrs, stateAttrs map[string]attr.Value,
	attrTypes map[string]attr.Type,
	pathPrefix string,
) []string {
	for setting := range attrTypes {
		if !planAttrs[setting].Equal(stateAttrs[setting]) {
			updateMaskPaths = append(
				updateMaskPaths,
				pathPrefix+setting,
			)
		}
	}
	return updateMaskPaths
}
