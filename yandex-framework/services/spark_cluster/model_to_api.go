package spark_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func BuildCreateClusterRequest(ctx context.Context, clusterModel *ClusterModel, providerConfig *config.State) (*spark.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	folderID, d := validate.FolderID(clusterModel.FolderId, providerConfig)
	diags.Append(d)
	if diags.HasError() {
		return nil, diags
	}

	subnetIds := make([]string, 0, len(clusterModel.Network.SubnetIds.Elements()))
	diags.Append(clusterModel.Network.SubnetIds.ElementsAs(ctx, &subnetIds, false)...)
	if diags.HasError() {
		return nil, diags
	}

	common, _, dd := buildCommonForCreateAndUpdate(ctx, clusterModel, nil)
	diags.Append(dd...)
	if diags.HasError() {
		return nil, diags
	}

	clusterCreateRequest := &spark.CreateClusterRequest{
		FolderId:    folderID,
		Name:        common.Name,
		Description: common.Description,
		Labels:      common.Labels,
		Config: &spark.ClusterConfig{
			ResourcePools: common.Config.ResourcePools,
			HistoryServer: common.Config.HistoryServer,
			Dependencies:  common.Config.Dependencies,
			Metastore:     common.Config.Metastore,
			SparkVersion:  common.Config.SparkVersion,
		},
		Network: &spark.NetworkConfig{
			SubnetIds:        subnetIds,
			SecurityGroupIds: common.SecurityGroupIds,
		},
		DeletionProtection: common.DeletionProtection,
		ServiceAccountId:   common.ServiceAccountId,
		Logging:            common.Logging,
		MaintenanceWindow:  common.MaintenanceWindow,
	}

	return clusterCreateRequest, diags
}

type CommonForCreateAndUpdate struct {
	Name               string
	Description        string
	Labels             map[string]string
	Config             *spark.ClusterConfig
	SecurityGroupIds   []string
	DeletionProtection bool
	ServiceAccountId   string
	Logging            *spark.LoggingConfig
	MaintenanceWindow  *spark.MaintenanceWindow
}

func buildCommonForCreateAndUpdate(ctx context.Context, plan, state *ClusterModel) (*CommonForCreateAndUpdate, []string, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	updateMaskPaths := make([]string, 0)

	if state != nil {
		if !plan.Name.Equal(state.Name) {
			updateMaskPaths = append(updateMaskPaths, "name")
		}
		if !stringsAreEqual(plan.Description, state.Description) {
			updateMaskPaths = append(updateMaskPaths, "description")
		}
		if !plan.DeletionProtection.Equal(state.DeletionProtection) {
			updateMaskPaths = append(updateMaskPaths, "deletion_protection")
		}
		if !plan.ServiceAccountId.Equal(state.ServiceAccountId) {
			updateMaskPaths = append(updateMaskPaths, "service_account_id")
		}
	}

	labels := make(map[string]string, len(plan.Labels.Elements()))
	diags.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	if state != nil && !mapsAreEqual(plan.Labels, state.Labels) {
		updateMaskPaths = append(updateMaskPaths, "labels")
	}

	var clusterConfig *spark.ClusterConfig

	var updDriver, updExecutor, updDependencies, updHistoryServer, updMetastore bool
	var updDriverPoolPreset, updDriverPoolSize bool
	var updExecutorPoolPreset, updExecutorPoolSize bool
	var updPip, updDeb bool
	var updSparkVersion bool

	if !plan.Config.IsNull() {

		planDriverPool, planExecutorPool := extractPools(ctx, plan, &diags)
		if diags.HasError() {
			return nil, nil, diags
		}

		driverScalePolicy := &spark.ScalePolicy{}
		if planDriverPool.Size.ValueInt64() > 0 {
			driverScalePolicy.ScaleType = &spark.ScalePolicy_FixedScale_{
				FixedScale: &spark.ScalePolicy_FixedScale{
					Size: planDriverPool.Size.ValueInt64(),
				},
			}
		} else {
			driverScalePolicy.ScaleType = &spark.ScalePolicy_AutoScale_{
				AutoScale: &spark.ScalePolicy_AutoScale{
					MinSize: planDriverPool.MinSize.ValueInt64(),
					MaxSize: planDriverPool.MaxSize.ValueInt64(),
				},
			}
		}

		executorScalePolicy := &spark.ScalePolicy{}
		if planExecutorPool.Size.ValueInt64() > 0 {
			executorScalePolicy.ScaleType = &spark.ScalePolicy_FixedScale_{
				FixedScale: &spark.ScalePolicy_FixedScale{
					Size: planExecutorPool.Size.ValueInt64(),
				},
			}
		} else {
			executorScalePolicy.ScaleType = &spark.ScalePolicy_AutoScale_{
				AutoScale: &spark.ScalePolicy_AutoScale{
					MinSize: planExecutorPool.MinSize.ValueInt64(),
					MaxSize: planExecutorPool.MaxSize.ValueInt64(),
				},
			}
		}

		var planDependencies Dependencies
		var pipPackages []string
		var debPackages []string
		if !plan.Config.Dependencies.IsNull() {
			diags.Append(plan.Config.Dependencies.As(ctx, &planDependencies, datasize.DefaultOpts)...)
			if diags.HasError() {
				return nil, nil, diags
			}
			pipPackages = make([]string, 0, len(planDependencies.PipPackages.Elements()))
			diags.Append(planDependencies.PipPackages.ElementsAs(ctx, &pipPackages, false)...)
			if diags.HasError() {
				return nil, nil, diags
			}
			debPackages = make([]string, 0, len(planDependencies.DebPackages.Elements()))
			diags.Append(planDependencies.DebPackages.ElementsAs(ctx, &debPackages, false)...)
			if diags.HasError() {
				return nil, nil, diags
			}
		}

		planHistoryServer := HistoryServer{}
		if !plan.Config.HistoryServer.IsNull() {
			diags.Append(plan.Config.HistoryServer.As(ctx, &planHistoryServer, datasize.DefaultOpts)...)
			if diags.HasError() {
				return nil, nil, diags
			}
		}

		planMetastore := Metastore{}
		if !plan.Config.Metastore.IsNull() {
			diags.Append(plan.Config.Metastore.As(ctx, &planMetastore, datasize.DefaultOpts)...)
			if diags.HasError() {
				return nil, nil, diags
			}
		}

		if state != nil {
			stateDriverPool, stateExecutorPool := extractPools(ctx, state, &diags)
			if diags.HasError() {
				return nil, nil, diags
			}

			updDriverPoolPreset = !stringsAreEqual(planDriverPool.ResourcePresetId, stateDriverPool.ResourcePresetId)
			updDriverPoolSize = !planDriverPool.Size.Equal(stateDriverPool.Size) ||
				!planDriverPool.MinSize.Equal(stateDriverPool.MinSize) ||
				!planDriverPool.MaxSize.Equal(stateDriverPool.MaxSize)
			updDriver = updDriverPoolPreset && updDriverPoolSize

			updExecutorPoolPreset = !stringsAreEqual(planExecutorPool.ResourcePresetId, stateExecutorPool.ResourcePresetId)
			updExecutorPoolSize = !planExecutorPool.Size.Equal(stateExecutorPool.Size) ||
				!planExecutorPool.MinSize.Equal(stateExecutorPool.MinSize) ||
				!planExecutorPool.MaxSize.Equal(stateExecutorPool.MaxSize)
			updExecutor = updExecutorPoolPreset && updExecutorPoolSize

			var stateDependencies Dependencies
			diags.Append(state.Config.Dependencies.As(ctx, &stateDependencies, datasize.DefaultOpts)...)
			if diags.HasError() {
				return nil, nil, diags
			}
			updPip = !setsAreEqual(planDependencies.PipPackages, stateDependencies.PipPackages)
			updDeb = !setsAreEqual(planDependencies.DebPackages, stateDependencies.DebPackages)
			updDependencies = updPip && updDeb

			var stateHistoryServer HistoryServer
			diags.Append(state.Config.HistoryServer.As(ctx, &stateHistoryServer, datasize.DefaultOpts)...)
			if diags.HasError() {
				return nil, nil, diags
			}
			updHistoryServer = planHistoryServer.Enabled.ValueBool() != stateHistoryServer.Enabled.ValueBool()

			var stateMetastore Metastore
			diags.Append(state.Config.Metastore.As(ctx, &stateMetastore, datasize.DefaultOpts)...)
			if diags.HasError() {
				return nil, nil, diags
			}
			updMetastore = !stringsAreEqual(planMetastore.ClusterId, stateMetastore.ClusterId)

			updSparkVersion = !stringsAreEqual(plan.Config.SparkVersion, state.Config.SparkVersion)
		}

		clusterConfig = &spark.ClusterConfig{
			ResourcePools: &spark.ResourcePools{
				Driver: &spark.ResourcePool{
					ResourcePresetId: planDriverPool.ResourcePresetId.ValueString(),
					ScalePolicy:      driverScalePolicy,
				},
				Executor: &spark.ResourcePool{
					ResourcePresetId: planExecutorPool.ResourcePresetId.ValueString(),
					ScalePolicy:      executorScalePolicy,
				},
			},
			HistoryServer: &spark.HistoryServerConfig{
				Enabled: planHistoryServer.Enabled.ValueBool(),
			},
			Dependencies: &spark.Dependencies{
				PipPackages: pipPackages,
				DebPackages: debPackages,
			},
			Metastore: &spark.Metastore{
				ClusterId: planMetastore.ClusterId.ValueString(),
			},
			SparkVersion: plan.Config.SparkVersion.ValueString(),
		}
	}
	if state != nil && !clusterConfigsAreEqual(ctx, plan.Config, state.Config, &diags) {
		if updDriver && updExecutor && updDependencies && updHistoryServer && updMetastore && updSparkVersion {
			updateMaskPaths = append(updateMaskPaths, "config_spec")
		} else {
			if updDriver && updExecutor {
				updateMaskPaths = append(updateMaskPaths, "config_spec.resource_pools")
			} else {
				if updDriver {
					updateMaskPaths = append(updateMaskPaths, "config_spec.resource_pools.driver")
				} else {
					if updDriverPoolPreset {
						updateMaskPaths = append(updateMaskPaths, "config_spec.resource_pools.driver.resource_preset_id")
					}
					if updDriverPoolSize {
						updateMaskPaths = append(updateMaskPaths, "config_spec.resource_pools.driver.scale_policy")
					}
				}
				if updExecutor {
					updateMaskPaths = append(updateMaskPaths, "config_spec.resource_pools.executor")
				} else {
					if updExecutorPoolPreset {
						updateMaskPaths = append(updateMaskPaths, "config_spec.resource_pools.executor.resource_preset_id")
					}
					if updExecutorPoolSize {
						updateMaskPaths = append(updateMaskPaths, "config_spec.resource_pools.executor.scale_policy")
					}
				}
			}
			if updDependencies {
				updateMaskPaths = append(updateMaskPaths, "config_spec.dependencies")
			} else {
				if updPip {
					updateMaskPaths = append(updateMaskPaths, "config_spec.dependencies.pip_packages")
				}
				if updDeb {
					updateMaskPaths = append(updateMaskPaths, "config_spec.dependencies.deb_packages")
				}
			}
			if updHistoryServer {
				updateMaskPaths = append(updateMaskPaths, "config_spec.history_server")
			}
			if updMetastore {
				updateMaskPaths = append(updateMaskPaths, "config_spec.metastore")
			}
			if updSparkVersion {
				updateMaskPaths = append(updateMaskPaths, "config_spec.spark_version")
			}
		}
	}

	securityGroupIds := make([]string, 0, len(plan.Network.SecurityGroupIds.Elements()))
	diags.Append(plan.Network.SecurityGroupIds.ElementsAs(ctx, &securityGroupIds, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	if state != nil && !setsAreEqual(plan.Network.SecurityGroupIds, state.Network.SecurityGroupIds) {
		updateMaskPaths = append(updateMaskPaths, "network_spec.security_group_ids")
	}

	var loggingConfig *spark.LoggingConfig

	if !plan.Logging.IsNull() {
		loggingConfig = &spark.LoggingConfig{
			Enabled: plan.Logging.Enabled.ValueBool(),
		}

		// both folder_id and log_group_id are specified or both are not specified
		if plan.Logging.Enabled.ValueBool() && plan.Logging.FolderId.IsNull() == plan.Logging.LogGroupId.IsNull() {
			diags.AddError("Invalid Spark cluster logging configuration",
				"Exactly one of the attributes `folder_id` and `log_group_id` must be specified")
			return nil, nil, diags
		}

		if !plan.Logging.FolderId.IsNull() {
			loggingConfig.Destination = &spark.LoggingConfig_FolderId{
				FolderId: plan.Logging.FolderId.ValueString(),
			}
		} else if !plan.Logging.LogGroupId.IsNull() {
			loggingConfig.Destination = &spark.LoggingConfig_LogGroupId{
				LogGroupId: plan.Logging.LogGroupId.ValueString(),
			}
		}
	}
	if state != nil && !loggingValuesAreEqual(plan.Logging, state.Logging) {
		updateMaskPaths = append(updateMaskPaths, "logging")
	}

	maintenanceWindow := &spark.MaintenanceWindow{}

	if !plan.MaintenanceWindow.IsNull() {
		if mwType := plan.MaintenanceWindow.MaintenanceWindowType.ValueString(); mwType == "ANYTIME" {
			maintenanceWindow.Policy = &spark.MaintenanceWindow_Anytime{
				Anytime: &spark.AnytimeMaintenanceWindow{},
			}
		} else if mwType == "WEEKLY" {
			mwDay := plan.MaintenanceWindow.Day.ValueString()
			mwHour := plan.MaintenanceWindow.Hour.ValueInt64()

			day := spark.WeeklyMaintenanceWindow_WeekDay_value[mwDay]

			maintenanceWindow.Policy = &spark.MaintenanceWindow_WeeklyMaintenanceWindow{
				WeeklyMaintenanceWindow: &spark.WeeklyMaintenanceWindow{
					Hour: mwHour,
					Day:  spark.WeeklyMaintenanceWindow_WeekDay(day),
				},
			}
		} else {
			diags.AddError(
				"Invalid maintenance window configuration.",
				"maintenance_window.type should be ANYTIME or WEEKLY",
			)
			return nil, nil, diags
		}
	}
	if state != nil && !maintenanceWindowsAreEqual(plan.MaintenanceWindow, state.MaintenanceWindow) {
		updateMaskPaths = append(updateMaskPaths, "maintenance_window")
	}

	params := &CommonForCreateAndUpdate{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Labels:             labels,
		Config:             clusterConfig,
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		ServiceAccountId:   plan.ServiceAccountId.ValueString(),
		Logging:            loggingConfig,
		MaintenanceWindow:  maintenanceWindow,
	}

	return params, updateMaskPaths, diags
}

func extractPools(ctx context.Context, model *ClusterModel, diags *diag.Diagnostics) (ResourcePool, ResourcePool) {
	var resourcePools ResourcePools
	diags.Append(model.Config.ResourcePools.As(ctx, &resourcePools, datasize.DefaultOpts)...)
	if diags.HasError() {
		return ResourcePool{}, ResourcePool{}
	}

	var driverPool ResourcePool
	diags.Append(resourcePools.Driver.As(ctx, &driverPool, datasize.DefaultOpts)...)
	if diags.HasError() {
		return ResourcePool{}, ResourcePool{}
	}

	var executorPool ResourcePool
	diags.Append(resourcePools.Executor.As(ctx, &executorPool, datasize.DefaultOpts)...)
	if diags.HasError() {
		return ResourcePool{}, ResourcePool{}
	}

	return driverPool, executorPool
}

func BuildUpdateClusterRequest(ctx context.Context, state *ClusterModel, plan *ClusterModel) (*spark.UpdateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	common, updateMaskPaths, dd := buildCommonForCreateAndUpdate(ctx, plan, state)
	diags.Append(dd...)
	if diags.HasError() {
		return nil, diags
	}

	updateClusterRequest := &spark.UpdateClusterRequest{
		ClusterId:   state.Id.ValueString(),
		UpdateMask:  &field_mask.FieldMask{Paths: updateMaskPaths},
		Name:        common.Name,
		Description: common.Description,
		Labels:      common.Labels,
		ConfigSpec: &spark.UpdateClusterConfigSpec{
			ResourcePools: common.Config.ResourcePools,
			HistoryServer: common.Config.HistoryServer,
			Dependencies:  common.Config.Dependencies,
			Metastore:     common.Config.Metastore,
			SparkVersion:  common.Config.SparkVersion,
		},
		NetworkSpec: &spark.UpdateNetworkConfigSpec{
			SecurityGroupIds: common.SecurityGroupIds,
		},
		DeletionProtection: common.DeletionProtection,
		ServiceAccountId:   common.ServiceAccountId,
		Logging:            common.Logging,
		MaintenanceWindow:  common.MaintenanceWindow,
	}

	return updateClusterRequest, diags
}
