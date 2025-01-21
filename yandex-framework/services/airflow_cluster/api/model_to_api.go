package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/airflow/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func BuildCreateClusterRequest(ctx context.Context, clusterModel *ClusterModel, providerConfig *config.State) (*airflow.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	folderID, d := validate.FolderID(clusterModel.FolderId, providerConfig)
	diags.Append(d)
	if diags.HasError() {
		return nil, diags
	}

	subnetIds := make([]string, 0, len(clusterModel.SubnetIds.Elements()))
	diags.Append(clusterModel.SubnetIds.ElementsAs(ctx, &subnetIds, false)...)
	if diags.HasError() {
		return nil, diags
	}

	common, _, dd := buildCommonForCreateAndUpdate(ctx, clusterModel, nil)
	diags.Append(dd...)
	if diags.HasError() {
		return nil, diags
	}

	clusterCreateRequest := &airflow.CreateClusterRequest{
		FolderId:    folderID,
		Name:        common.Name,
		Description: common.Description,
		Labels:      common.Labels,
		Config: &airflow.ClusterConfig{
			Airflow:      common.AirflowConfig,
			Webserver:    common.Webserver,
			Scheduler:    common.Scheduler,
			Triggerer:    common.Triggerer,
			Worker:       common.Worker,
			Dependencies: common.Dependencies,
			Lockbox:      common.Lockbox,
		},
		Network: &airflow.NetworkConfig{
			SubnetIds:        subnetIds,
			SecurityGroupIds: common.SecurityGroupIds,
		},
		CodeSync:           common.CodeSync,
		DeletionProtection: common.DeletionProtection, // todo set default to false
		ServiceAccountId:   common.ServiceAccountId,
		Logging:            common.Logging,
		AdminPassword:      clusterModel.AdminPassword.ValueString(),
	}

	return clusterCreateRequest, diags
}

type CommonForCreateAndUpdate struct {
	Name               string
	Description        string
	Labels             map[string]string
	CodeSync           *airflow.CodeSyncConfig
	SecurityGroupIds   []string
	DeletionProtection bool
	ServiceAccountId   string
	Logging            *airflow.LoggingConfig

	AirflowConfig *airflow.AirflowConfig
	Webserver     *airflow.WebserverConfig
	Scheduler     *airflow.SchedulerConfig
	Worker        *airflow.WorkerConfig
	Triggerer     *airflow.TriggererConfig
	Dependencies  *airflow.Dependencies
	Lockbox       *airflow.LockboxConfig
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

	pipPackages := make([]string, 0, len(plan.PipPackages.Elements()))
	diags.Append(plan.PipPackages.ElementsAs(ctx, &pipPackages, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	if state != nil && !setsAreEqual(plan.PipPackages, state.PipPackages) {
		updateMaskPaths = append(updateMaskPaths, "config_spec.dependencies.pip_packages")
	}

	debPackages := make([]string, 0, len(plan.DebPackages.Elements()))
	diags.Append(plan.DebPackages.ElementsAs(ctx, &debPackages, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	if state != nil && !setsAreEqual(plan.DebPackages, state.DebPackages) {
		updateMaskPaths = append(updateMaskPaths, "config_spec.dependencies.deb_packages")
	}

	securityGroupIds := make([]string, 0, len(plan.SecurityGroupIds.Elements()))
	diags.Append(plan.SecurityGroupIds.ElementsAs(ctx, &securityGroupIds, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	if state != nil && !setsAreEqual(plan.SecurityGroupIds, state.SecurityGroupIds) {
		updateMaskPaths = append(updateMaskPaths, "network_spec.security_group_ids")
	}

	objectValuable, dd := S3Type{}.ValueFromObject(ctx, plan.CodeSync.S3)
	diags.Append(dd...)
	if diags.HasError() {
		return nil, nil, diags
	}
	s3Value := objectValuable.(S3Value)
	codeSyncConfig := &airflow.CodeSyncConfig{
		Source: &airflow.CodeSyncConfig_S3{
			S3: &airflow.S3Config{Bucket: s3Value.Bucket.ValueString()},
		},
	}
	if state != nil && !plan.CodeSync.Equal(state.CodeSync) {
		updateMaskPaths = append(updateMaskPaths, "code_sync")
	}

	var loggingConfig *airflow.LoggingConfig
	if !plan.Logging.IsNull() {
		minLevel, d := logLevelToAPI(plan.Logging.MinLevel)
		diags.Append(d)
		if diags.HasError() {
			return nil, nil, diags
		}

		loggingConfig = &airflow.LoggingConfig{
			Enabled:  plan.Logging.Enabled.ValueBool(),
			MinLevel: minLevel,
		}

		// both folder_id and log_group_id are specified or both are not specified
		if plan.Logging.FolderId.IsNull() == plan.Logging.LogGroupId.IsNull() {
			diags.AddError("Invalid Airflow cluster logging configuration",
				"Exactly one of the attributes `folder_id` and `log_group_id` must be specified")
			return nil, nil, diags
		}

		if !plan.Logging.FolderId.IsNull() {
			loggingConfig.Destination = &airflow.LoggingConfig_FolderId{
				FolderId: plan.Logging.FolderId.ValueString(),
			}
		} else {
			loggingConfig.Destination = &airflow.LoggingConfig_LogGroupId{
				LogGroupId: plan.Logging.LogGroupId.ValueString(),
			}
		}
	}
	if state != nil && !loggingValuesAreEqual(plan.Logging, state.Logging) {
		updateMaskPaths = append(updateMaskPaths, "logging")
	}

	var lockboxConfig *airflow.LockboxConfig
	if !plan.LockboxSecretsBackend.IsNull() {
		lockboxConfig = &airflow.LockboxConfig{Enabled: plan.LockboxSecretsBackend.Enabled.ValueBool()}
	}
	if state != nil && !lockboxSecretsBackendValuesAreEqual(plan.LockboxSecretsBackend, state.LockboxSecretsBackend) {
		updateMaskPaths = append(updateMaskPaths, "config_spec.lockbox")
	}

	var airflowConfig *airflow.AirflowConfig
	if !plan.AirflowConfig.IsNull() {
		configWithSections := make(map[string]map[string]string, len(plan.AirflowConfig.Elements()))
		diags.Append(plan.AirflowConfig.ElementsAs(ctx, &configWithSections, false)...)
		if diags.HasError() {
			return nil, nil, diags
		}

		plainConfig := make(map[string]string, 0)
		for sectionName, section := range configWithSections {
			for propName, propValue := range section {
				fullName := fmt.Sprintf("%s.%s", sectionName, propName)
				plainConfig[fullName] = propValue
			}
		}

		airflowConfig = &airflow.AirflowConfig{Config: plainConfig}
	}
	if state != nil && !mapsAreEqual(plan.AirflowConfig, state.AirflowConfig) {
		updateMaskPaths = append(updateMaskPaths, "config_spec.airflow")
	}

	webserverConfig := &airflow.WebserverConfig{
		Count:     plan.Webserver.Count.ValueInt64(),
		Resources: &airflow.Resources{ResourcePresetId: plan.Webserver.ResourcePresetId.ValueString()},
	}
	if state != nil && !plan.Webserver.Equal(state.Webserver) {
		updateMaskPaths = append(updateMaskPaths, "config_spec.webserver")
	}

	schedulerConfig := &airflow.SchedulerConfig{
		Count:     plan.Scheduler.Count.ValueInt64(),
		Resources: &airflow.Resources{ResourcePresetId: plan.Scheduler.ResourcePresetId.ValueString()},
	}
	if state != nil && !plan.Scheduler.Equal(state.Scheduler) {
		updateMaskPaths = append(updateMaskPaths, "config_spec.scheduler")
	}

	workerConfig := &airflow.WorkerConfig{
		MinCount:  plan.Worker.MinCount.ValueInt64(),
		MaxCount:  plan.Worker.MaxCount.ValueInt64(),
		Resources: &airflow.Resources{ResourcePresetId: plan.Worker.ResourcePresetId.ValueString()},
	}
	if state != nil && !plan.Worker.Equal(state.Worker) {
		updateMaskPaths = append(updateMaskPaths, "config_spec.worker")
	}

	var triggererConfig *airflow.TriggererConfig
	if !plan.Triggerer.IsNull() {
		triggererConfig = &airflow.TriggererConfig{
			Count:     plan.Triggerer.Count.ValueInt64(),
			Resources: &airflow.Resources{ResourcePresetId: plan.Triggerer.ResourcePresetId.ValueString()},
		}
	}
	if state != nil && !plan.Triggerer.Equal(state.Triggerer) {
		updateMaskPaths = append(updateMaskPaths, "config_spec.triggerer")
	}

	params := &CommonForCreateAndUpdate{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Labels:             labels,
		CodeSync:           codeSyncConfig,
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		ServiceAccountId:   plan.ServiceAccountId.ValueString(),
		Logging:            loggingConfig,
		AirflowConfig:      airflowConfig,
		Webserver:          webserverConfig,
		Scheduler:          schedulerConfig,
		Worker:             workerConfig,
		Triggerer:          triggererConfig,
		Dependencies: &airflow.Dependencies{
			PipPackages: pipPackages,
			DebPackages: debPackages,
		},
		Lockbox: lockboxConfig,
	}

	return params, updateMaskPaths, diags
}

func logLevelToAPI(minLevelValue types.String) (logging.LogLevel_Level, diag.Diagnostic) {
	if minLevelValue.IsNull() {
		return logging.LogLevel_LEVEL_UNSPECIFIED, nil
	}

	minLevel, ok := logging.LogLevel_Level_value[minLevelValue.ValueString()]
	if !ok || minLevel == 0 {
		return 0, diag.NewErrorDiagnostic("Invalid Airflow cluster logging configuration",
			fmt.Sprintf("Unsupported value for `min_level` attribute provided. It must be one of `%s`", strings.Join(allowedLogLevels(), "`, `")))
	}
	return logging.LogLevel_Level(minLevel), nil
}

func BuildUpdateClusterRequest(ctx context.Context, state *ClusterModel, plan *ClusterModel) (*airflow.UpdateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	common, updateMaskPaths, dd := buildCommonForCreateAndUpdate(ctx, plan, state)
	diags.Append(dd...)
	if diags.HasError() {
		return nil, diags
	}

	updateClusterRequest := &airflow.UpdateClusterRequest{
		ClusterId:   state.Id.ValueString(),
		UpdateMask:  &field_mask.FieldMask{Paths: updateMaskPaths},
		Name:        common.Name,
		Description: common.Description,
		Labels:      common.Labels,
		ConfigSpec: &airflow.UpdateClusterConfigSpec{
			Airflow:      common.AirflowConfig,
			Webserver:    common.Webserver,
			Scheduler:    common.Scheduler,
			Triggerer:    common.Triggerer,
			Worker:       common.Worker,
			Dependencies: common.Dependencies,
			Lockbox:      common.Lockbox,
		},
		CodeSync: common.CodeSync,
		NetworkSpec: &airflow.UpdateNetworkConfigSpec{
			SecurityGroupIds: common.SecurityGroupIds,
		},
		DeletionProtection: common.DeletionProtection,
		ServiceAccountId:   common.ServiceAccountId,
		Logging:            common.Logging,
	}

	return updateClusterRequest, diags
}

func setsAreEqual(set1, set2 types.Set) bool {
	if set1.Equal(set2) {
		return true
	}
	// if one of sets is null and the other is empty then we assume that they are equal
	if len(set1.Elements()) == 0 && len(set2.Elements()) == 0 {
		return true
	}
	return false
}

func mapsAreEqual(map1, map2 types.Map) bool {
	if map1.Equal(map2) {
		return true
	}
	// if one of map is null and the other is empty then we assume that they are equal
	if len(map1.Elements()) == 0 && len(map2.Elements()) == 0 {
		return true
	}
	return false
}

func stringsAreEqual(str1, str2 types.String) bool {
	if str1.Equal(str2) {
		return true
	}
	// if one of strings is null and the other is empty then we assume that they are equal
	if str1.ValueString() == "" && str2.ValueString() == "" {
		return true
	}
	return false
}

func (v LockboxSecretsBackendValue) IsExplicitlyDisabled() bool {
	return !v.IsNull() && !v.Enabled.ValueBool()
}

func lockboxSecretsBackendValuesAreEqual(val1, val2 LockboxSecretsBackendValue) bool {
	if val1.Equal(val2) {
		return true
	}
	// if one of values is null and the other is empty then we assume that they are equal
	if (val1.IsExplicitlyDisabled() && val2.IsNull()) || (val1.IsNull() && val2.IsExplicitlyDisabled()) {
		return true
	}

	return false
}

func (v LoggingValue) IsExplicitlyDisabled() bool {
	return !v.IsNull() && !v.Enabled.ValueBool()
}

func loggingValuesAreEqual(val1, val2 LoggingValue) bool {
	if val1.Equal(val2) {
		return true
	}
	// if one of values is null and the other is empty then we assume that they are equal
	if (val1.IsExplicitlyDisabled() && val2.IsNull()) || (val1.IsNull() && val2.IsExplicitlyDisabled()) {
		return true
	}

	return false
}

func allowedLogLevels() []string {
	allowedLevels := make([]string, 0, len(logging.LogLevel_Level_value))
	for levelName, val := range logging.LogLevel_Level_value {
		if val == 0 {
			continue
		}
		allowedLevels = append(allowedLevels, levelName)
	}
	return allowedLevels
}

func logLevelValidator() validator.String {
	return stringvalidator.OneOf(allowedLogLevels()...)
}
