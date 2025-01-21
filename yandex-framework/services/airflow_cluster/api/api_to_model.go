package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/airflow/v1"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
)

func ClusterToState(ctx context.Context, cluster *airflow.Cluster, state *ClusterModel) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("clusterToState: Airflow cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("clusterToState: Received Airflow cluster data: %+v", cluster))

	state.FolderId = types.StringValue(cluster.GetFolderId())
	state.CreatedAt = types.StringValue(timestamp.Get(cluster.GetCreatedAt()))
	state.Name = types.StringValue(cluster.GetName())
	state.Health = types.StringValue(cluster.GetHealth().String())
	state.Status = types.StringValue(cluster.GetStatus().String())
	state.DeletionProtection = types.BoolValue(cluster.GetDeletionProtection())
	state.ServiceAccountId = types.StringValue(cluster.ServiceAccountId)

	// Do not override state.Description if it is null and server returned description=""
	// otherwise we will get "Provider produced inconsistent result after apply"
	// Same principle applies to other attributes that can be null or empty.
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

	subnetIds, diags := nullableStringSliceToSet(ctx, cluster.GetNetwork().GetSubnetIds())
	if diags.HasError() {
		return diags
	}
	if !setsAreEqual(state.SubnetIds, subnetIds) {
		state.SubnetIds = subnetIds
	}

	securityGroupIds, diags := nullableStringSliceToSet(ctx, cluster.GetNetwork().GetSecurityGroupIds())
	if diags.HasError() {
		return diags
	}
	if !setsAreEqual(state.SecurityGroupIds, securityGroupIds) {
		state.SecurityGroupIds = securityGroupIds
	}

	pipPackages, diags := nullableStringSliceToSet(ctx, cluster.GetConfig().GetDependencies().GetPipPackages())
	if diags.HasError() {
		return diags
	}
	if !setsAreEqual(state.PipPackages, pipPackages) {
		state.PipPackages = pipPackages
	}

	debPackages, diags := nullableStringSliceToSet(ctx, cluster.GetConfig().GetDependencies().GetDebPackages())
	if diags.HasError() {
		return diags
	}
	if !setsAreEqual(state.DebPackages, debPackages) {
		state.DebPackages = debPackages
	}

	state.Webserver = webserverValueFromAPI(cluster.GetConfig().GetWebserver())
	state.Scheduler = schedulerValueFromAPI(cluster.GetConfig().GetScheduler())
	state.Worker = workerValueFromAPI(cluster.GetConfig().GetWorker())
	state.Triggerer = triggererValueFromAPI(cluster.GetConfig().GetTriggerer())

	codeSyncConfigObject, diags := codeSyncValueFromAPI(ctx, cluster.GetCodeSync())
	if diags.HasError() {
		return diags
	}
	state.CodeSync = codeSyncConfigObject

	loggingConfig, diags := loggingValueFromAPI(cluster.GetLogging())
	if diags.HasError() {
		return diags
	}
	if !state.Logging.IsNull() || loggingConfig.IsNull() || loggingConfig.Enabled.ValueBool() {
		state.Logging = loggingConfig
	}

	lockboxValue := lockboxValueFromAPI(cluster.GetConfig().GetLockbox())
	if !state.LockboxSecretsBackend.IsNull() || lockboxValue.IsNull() || lockboxValue.Enabled.ValueBool() {
		state.LockboxSecretsBackend = lockboxValue
	}

	airflowConfig, diags := airflowConfigFromAPI(ctx, cluster.GetConfig().GetAirflow())
	if diags.HasError() {
		return diags
	}
	if !state.AirflowConfig.IsNull() || airflowConfig.IsNull() || len(airflowConfig.Elements()) != 0 {
		state.AirflowConfig = airflowConfig
	}

	return diags
}

func webserverValueFromAPI(cfg *airflow.WebserverConfig) WebserverValue {
	if cfg == nil {
		return NewWebserverValueNull()
	}

	return WebserverValue{
		Count:            types.Int64Value(cfg.GetCount()),
		ResourcePresetId: types.StringValue(cfg.GetResources().GetResourcePresetId()),
		state:            attr.ValueStateKnown,
	}
}

func schedulerValueFromAPI(cfg *airflow.SchedulerConfig) SchedulerValue {
	if cfg == nil {
		return NewSchedulerValueNull()
	}

	return SchedulerValue{
		Count:            types.Int64Value(cfg.GetCount()),
		ResourcePresetId: types.StringValue(cfg.GetResources().GetResourcePresetId()),
		state:            attr.ValueStateKnown,
	}
}

func workerValueFromAPI(cfg *airflow.WorkerConfig) WorkerValue {
	if cfg == nil {
		return NewWorkerValueNull()
	}

	return WorkerValue{
		MinCount:         types.Int64Value(cfg.GetMinCount()),
		MaxCount:         types.Int64Value(cfg.GetMaxCount()),
		ResourcePresetId: types.StringValue(cfg.GetResources().GetResourcePresetId()),
		state:            attr.ValueStateKnown,
	}
}

func triggererValueFromAPI(cfg *airflow.TriggererConfig) TriggererValue {
	if cfg == nil {
		return NewTriggererValueNull()
	}

	return TriggererValue{
		Count:            types.Int64Value(cfg.GetCount()),
		ResourcePresetId: types.StringValue(cfg.GetResources().GetResourcePresetId()),
		state:            attr.ValueStateKnown,
	}
}

func codeSyncValueFromAPI(ctx context.Context, cfg *airflow.CodeSyncConfig) (CodeSyncValue, diag.Diagnostics) {
	if cfg == nil || cfg.GetSource() == nil {
		return NewCodeSyncValueNull(), diag.Diagnostics{}
	}

	s3Source, ok := cfg.GetSource().(*airflow.CodeSyncConfig_S3)
	if !ok {
		d := diag.NewErrorDiagnostic("Failed to parse Airflow cluster value received from Cloud API",
			"CodeSync source has unexpected type. Please update provider.")
		return NewCodeSyncValueUnknown(), diag.Diagnostics{d}
	}

	s3Value := S3Value{
		Bucket: types.StringValue(s3Source.S3.GetBucket()),
		state:  attr.ValueStateKnown,
	}

	s3AsObjectValue, diags := s3Value.ToObjectValue(ctx)
	if diags.HasError() {
		return NewCodeSyncValueUnknown(), diags
	}

	return CodeSyncValue{
		S3:    s3AsObjectValue,
		state: attr.ValueStateKnown,
	}, diag.Diagnostics{}
}

func nullableStringSliceToSet(ctx context.Context, s []string) (types.Set, diag.Diagnostics) {
	if s == nil {
		return types.SetNull(types.StringType), diag.Diagnostics{}
	}

	return types.SetValueFrom(ctx, types.StringType, s)
}

func loggingValueFromAPI(cfg *airflow.LoggingConfig) (LoggingValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if cfg == nil {
		return NewLoggingValueNull(), diags
	}

	minLevel := types.StringValue(cfg.GetMinLevel().String())
	if cfg.GetMinLevel() == 0 {
		minLevel = types.StringNull()
	}

	loggingValue := LoggingValue{
		Enabled:    types.BoolValue(cfg.GetEnabled()),
		FolderId:   types.StringNull(),
		LogGroupId: types.StringNull(),
		MinLevel:   minLevel,
		state:      attr.ValueStateKnown,
	}

	switch t := cfg.GetDestination().(type) {
	case *airflow.LoggingConfig_FolderId:
		loggingValue.FolderId = types.StringValue(t.FolderId)
	case *airflow.LoggingConfig_LogGroupId:
		loggingValue.LogGroupId = types.StringValue(t.LogGroupId)
	default:
		diags.AddError("Failed to parse Airflow cluster value received from Cloud API",
			"Logging destination has unexpected type. Please update provider.")
		return NewLoggingValueNull(), diags
	}

	return loggingValue, diags
}

func lockboxValueFromAPI(cfg *airflow.LockboxConfig) LockboxSecretsBackendValue {
	if cfg == nil {
		return NewLockboxSecretsBackendValueNull()
	}

	return LockboxSecretsBackendValue{
		Enabled: types.BoolValue(cfg.GetEnabled()),
		state:   attr.ValueStateKnown,
	}
}

func airflowConfigFromAPI(ctx context.Context, cfg *airflow.AirflowConfig) (basetypes.MapValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	elementType := types.MapType{ElemType: types.StringType}
	if cfg == nil {
		return types.MapNull(elementType), diags
	}

	plainConfig := cfg.Config
	configWithSections := make(map[string]map[string]string, 0)
	for fullName, propValue := range plainConfig {
		parts := strings.SplitN(fullName, ".", 2)
		if len(parts) != 2 {
			d := diag.NewErrorDiagnostic("Failed to parse Airflow config received from Cloud API",
				fmt.Sprintf("Config property is expected to have format \"<section_name>.<property_name>\" but got %q", fullName))
			return types.MapUnknown(elementType), diag.Diagnostics{d}
		}
		sectionName := parts[0]
		propName := parts[1]

		section, ok := configWithSections[sectionName]
		if !ok {
			section = make(map[string]string, 0)
			configWithSections[sectionName] = section
		}
		section[propName] = propValue
	}

	return types.MapValueFrom(ctx, elementType, configWithSections)
}
