package spark_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
)

func ClusterToState(ctx context.Context, cluster *spark.Cluster, state *ClusterModel) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("clusterToState: Spark cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("clusterToState: Received Spark cluster data: %+v", cluster))

	state.FolderId = types.StringValue(cluster.GetFolderId())
	state.CreatedAt = types.StringValue(timestamp.Get(cluster.GetCreatedAt()))
	state.Name = types.StringValue(cluster.GetName())

	newDescription := types.StringValue(cluster.GetDescription())
	if !stringsAreEqual(state.Description, newDescription) {
		state.Description = newDescription
	}

	labels, diags := types.MapValueFrom(ctx, types.StringType, cluster.Labels)
	if diags.HasError() {
		return diags
	}
	if !mapsAreEqual(state.Labels, labels) {
		state.Labels = labels
	}

	clusterConfig, diags := clusterConfigFromAPI(ctx, cluster.GetConfig())
	if diags.HasError() {
		return diags
	}
	if !clusterConfigsAreEqual(ctx, state.Config, clusterConfig, &diags) {
		state.Config = clusterConfig
	}

	state.Status = types.StringValue(cluster.GetStatus().String())

	networkConfig, diags := networkConfigFromAPI(ctx, cluster.GetNetwork())
	if diags.HasError() {
		return diags
	}
	if !networkConfigsAreEqual(state.Network, networkConfig) {
		state.Network = networkConfig
	}

	state.DeletionProtection = types.BoolValue(cluster.GetDeletionProtection())
	state.ServiceAccountId = types.StringValue(cluster.ServiceAccountId)

	loggingConfig, diags := loggingConfigFromAPI(cluster.GetLogging())
	if diags.HasError() {
		return diags
	}
	if !loggingValuesAreEqual(state.Logging, loggingConfig) {
		state.Logging = loggingConfig
	}

	maintenanceWindow, diags := maintenanceWindowFromAPI(cluster.GetMaintenanceWindow())
	if !maintenanceWindowsAreEqual(state.MaintenanceWindow, maintenanceWindow) {
		state.MaintenanceWindow = maintenanceWindow
	}

	return diags
}

func clusterConfigFromAPI(ctx context.Context, cfg *spark.ClusterConfig) (ConfigValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if cfg == nil {
		return NewConfigValueNull(), diags
	}

	resourcePoolsObj, d := resourcePoolsFromAPI(ctx, cfg.GetResourcePools())
	diags.Append(d...)

	dependenciesObj, d := dependenciesFromAPI(ctx, cfg.GetDependencies())
	diags.Append(d...)

	historyServerObj, d := historyServerFromAPI(ctx, cfg.GetHistoryServer())
	diags.Append(d...)

	metastoreObj, d := metastoreFromAPI(ctx, cfg.GetMetastore())
	diags.Append(d...)

	return ConfigValue{
		ResourcePools: resourcePoolsObj,
		Dependencies:  dependenciesObj,
		HistoryServer: historyServerObj,
		Metastore:     metastoreObj,
		state:         attr.ValueStateKnown,
	}, diags
}

func networkConfigFromAPI(ctx context.Context, cfg *spark.NetworkConfig) (NetworkValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if cfg == nil {
		return NewNetworkValueNull(), diags
	}

	subnetIds, diags := nullableStringSliceToSet(ctx, cfg.GetSubnetIds())
	if diags.HasError() {
		return NewNetworkValueNull(), diags
	}

	securityGroupIds, diags := nullableStringSliceToSet(ctx, cfg.GetSecurityGroupIds())
	if diags.HasError() {
		return NewNetworkValueNull(), diags
	}

	return NetworkValue{
		SubnetIds:        subnetIds,
		SecurityGroupIds: securityGroupIds,
		state:            attr.ValueStateKnown,
	}, diags
}

func loggingConfigFromAPI(cfg *spark.LoggingConfig) (LoggingValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if cfg == nil {
		return NewLoggingValueNull(), diags
	}

	loggingValue := LoggingValue{
		Enabled:    types.BoolValue(cfg.GetEnabled()),
		FolderId:   types.StringNull(),
		LogGroupId: types.StringNull(),
		state:      attr.ValueStateKnown,
	}

	if cfg.GetEnabled() {
		switch t := cfg.GetDestination().(type) {
		case *spark.LoggingConfig_FolderId:
			loggingValue.FolderId = types.StringValue(t.FolderId)
		case *spark.LoggingConfig_LogGroupId:
			loggingValue.LogGroupId = types.StringValue(t.LogGroupId)
		default:
			diags.AddError("Failed to parse Spark cluster value received from Cloud API",
				"Logging destination has unexpected type. Please update provider.")
			return NewLoggingValueNull(), diags
		}
	}

	return loggingValue, diags
}

func maintenanceWindowFromAPI(mw *spark.MaintenanceWindow) (MaintenanceWindowValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if mw == nil {
		return NewMaintenanceWindowValueNull(), diags
	}

	var maintenanceWindow MaintenanceWindowValue
	switch policy := mw.GetPolicy().(type) {
	case *spark.MaintenanceWindow_Anytime:
		maintenanceWindow = MaintenanceWindowValue{
			MaintenanceWindowType: types.StringValue("ANYTIME"),
			state:                 attr.ValueStateKnown,
		}
	case *spark.MaintenanceWindow_WeeklyMaintenanceWindow:
		day := spark.WeeklyMaintenanceWindow_WeekDay_name[int32(policy.WeeklyMaintenanceWindow.GetDay())]
		maintenanceWindow = MaintenanceWindowValue{
			MaintenanceWindowType: types.StringValue("WEEKLY"),
			Day:                   types.StringValue(day),
			Hour:                  types.Int64Value(policy.WeeklyMaintenanceWindow.GetHour()),
			state:                 attr.ValueStateKnown,
		}
	default:
		diags.AddError(
			"Failed to parse Spark maintenance window received from Cloud API",
			"Maintenance window has unexpected type",
		)
		return NewMaintenanceWindowValueNull(), diags
	}

	return maintenanceWindow, diags
}

func resourcePoolsFromAPI(ctx context.Context, cfg *spark.ResourcePools) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if cfg == nil {
		return types.ObjectNull(ResourcePoosAttrTypes), diags
	}
	driverResourcePool := ResourcePool{
		ResourcePresetId: types.StringValue(cfg.GetDriver().GetResourcePresetId()),
	}
	driverScalePolicy := cfg.GetDriver().GetScalePolicy()
	if driverScalePolicy.GetFixedScale() != nil {
		driverResourcePool.Size = types.Int64Value(driverScalePolicy.GetFixedScale().GetSize())
	} else if driverScalePolicy.GetAutoScale() != nil {
		driverResourcePool.MinSize = types.Int64Value(driverScalePolicy.GetAutoScale().GetMinSize())
		driverResourcePool.MaxSize = types.Int64Value(driverScalePolicy.GetAutoScale().GetMaxSize())
	}
	driverObj, d := types.ObjectValueFrom(
		ctx,
		ResourcePoolAttrTypes,
		driverResourcePool,
	)
	diags.Append(d...)

	executorResourcePool := ResourcePool{
		ResourcePresetId: types.StringValue(cfg.GetExecutor().GetResourcePresetId()),
	}
	executorScalePolicy := cfg.GetExecutor().GetScalePolicy()
	if executorScalePolicy.GetFixedScale() != nil {
		executorResourcePool.Size = types.Int64Value(executorScalePolicy.GetFixedScale().GetSize())
	} else if executorScalePolicy.GetAutoScale() != nil {
		executorResourcePool.MinSize = types.Int64Value(executorScalePolicy.GetAutoScale().GetMinSize())
		executorResourcePool.MaxSize = types.Int64Value(executorScalePolicy.GetAutoScale().GetMaxSize())
	}
	executorObj, d := types.ObjectValueFrom(
		ctx,
		ResourcePoolAttrTypes,
		executorResourcePool,
	)
	diags.Append(d...)

	resourcePoolsObj, d := types.ObjectValueFrom(
		ctx,
		ResourcePoosAttrTypes,
		ResourcePools{
			Driver:   driverObj,
			Executor: executorObj,
		},
	)
	diags.Append(d...)
	return resourcePoolsObj, diags
}

func dependenciesFromAPI(ctx context.Context, cfg *spark.Dependencies) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	pipPackages := types.SetValueMust(types.StringType, []attr.Value{})
	debPackages := types.SetValueMust(types.StringType, []attr.Value{})
	if cfg != nil {
		if cfg.GetPipPackages() != nil {
			pipPackages, diags = types.SetValueFrom(ctx, types.StringType, cfg.GetPipPackages())
			if diags.HasError() {
				return types.ObjectNull(DependenciesAttrTypes), diags
			}
		}
		if cfg.GetDebPackages() != nil {
			debPackages, diags = types.SetValueFrom(ctx, types.StringType, cfg.GetDebPackages())
			if diags.HasError() {
				return types.ObjectNull(DependenciesAttrTypes), diags
			}
		}
	}
	dependenciesObj, d := types.ObjectValueFrom(
		ctx,
		DependenciesAttrTypes,
		Dependencies{
			PipPackages: pipPackages,
			DebPackages: debPackages,
		},
	)
	diags.Append(d...)
	return dependenciesObj, diags
}

func historyServerFromAPI(ctx context.Context, cfg *spark.HistoryServerConfig) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	var enabled bool
	if cfg != nil {
		enabled = cfg.GetEnabled()
	}
	historyServerObj, d := types.ObjectValueFrom(
		ctx,
		HistoryServerAttrTypes,
		HistoryServer{
			Enabled: types.BoolValue(enabled),
		},
	)
	diags.Append(d...)
	return historyServerObj, diags
}

func metastoreFromAPI(ctx context.Context, cfg *spark.Metastore) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	var metastoreClusterID string
	if cfg != nil {
		metastoreClusterID = cfg.GetClusterId()
	}
	metastoreObj, d := types.ObjectValueFrom(
		ctx,
		MetastoreAttrTypes,
		Metastore{
			ClusterId: types.StringValue(metastoreClusterID),
		},
	)
	diags.Append(d...)
	return metastoreObj, diags
}
