package trino_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
)

func ClusterToState(ctx context.Context, cluster *trino.Cluster, state *ClusterModel) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("clusterToState: Trino cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("clusterToState: Received Trino cluster data: %+v", cluster))

	state.FolderId = types.StringValue(cluster.GetFolderId())
	state.CreatedAt = types.StringValue(timestamp.Get(cluster.GetCreatedAt()))
	state.Name = types.StringValue(cluster.GetName())
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

	state.Coordinator = coordinatorValueFromAPI(cluster.GetTrino().GetCoordinatorConfig())

	worker, diags := workerValueFromAPI(ctx, cluster.GetTrino().GetWorkerConfig())
	if diags.HasError() {
		return diags
	}
	state.Worker = worker

	retryPolicy, diags := retryPolicyValueFromAPI(ctx, cluster.GetTrino().GetRetryPolicy())
	if diags.HasError() {
		return diags
	}
	state.RetryPolicy = retryPolicy

	loggingConfig, diags := loggingValueFromAPI(cluster.GetLogging())
	if diags.HasError() {
		return diags
	}
	if !state.Logging.IsNull() || loggingConfig.IsNull() || loggingConfig.Enabled.ValueBool() {
		state.Logging = loggingConfig
	}

	maintenanceWindow, diags := maintenanceWindowFromAPI(cluster.GetMaintenanceWindow())
	if diags.HasError() {
		return diags
	}
	state.MaintenanceWindow = maintenanceWindow

	return diags
}

func coordinatorValueFromAPI(cfg *trino.CoordinatorConfig) CoordinatorValue {
	if cfg == nil {
		return NewCoordinatorValueNull()
	}

	return CoordinatorValue{
		ResourcePresetId: types.StringValue(cfg.GetResources().GetResourcePresetId()),
		state:            attr.ValueStateKnown,
	}
}

func workerValueFromAPI(ctx context.Context, cfg *trino.WorkerConfig) (WorkerValue, diag.Diagnostics) {
	if cfg == nil {
		return NewWorkerValueNull(), diag.Diagnostics{}
	}

	value := WorkerValue{
		FixedScale:       basetypes.NewObjectNull(FixedScaleValue{}.AttributeTypes(ctx)),
		AutoScale:        basetypes.NewObjectNull(AutoScaleValue{}.AttributeTypes(ctx)),
		ResourcePresetId: types.StringValue(cfg.GetResources().GetResourcePresetId()),
		state:            attr.ValueStateKnown,
	}

	switch scale := cfg.GetScalePolicy().GetScaleType().(type) {
	case *trino.WorkerConfig_WorkerScalePolicy_FixedScale:
		object, diags := FixedScaleValue{
			Count: types.Int64Value(scale.FixedScale.GetCount()),
			state: attr.ValueStateKnown,
		}.ToObjectValue(ctx)
		if diags.HasError() {
			return NewWorkerValueNull(), diags
		}
		value.FixedScale = object
	case *trino.WorkerConfig_WorkerScalePolicy_AutoScale:
		object, diags := AutoScaleValue{
			MaxCount: types.Int64Value(scale.AutoScale.GetMaxCount()),
			MinCount: types.Int64Value(scale.AutoScale.GetMinCount()),
			state:    attr.ValueStateKnown,
		}.ToObjectValue(ctx)
		if diags.HasError() {
			return NewWorkerValueNull(), diags
		}
		value.AutoScale = object
	}

	return value, diag.Diagnostics{}
}

func retryPolicyValueFromAPI(ctx context.Context, cfg *trino.RetryPolicyConfig) (RetryPolicyValue, diag.Diagnostics) {
	if cfg == nil {
		return NewRetryPolicyValueNull(), diag.Diagnostics{}
	}

	additionalProperties, diags := types.MapValueFrom(ctx, types.StringType, cfg.AdditionalProperties)
	if diags.HasError() {
		return NewRetryPolicyValueUnknown(), diags
	}

	_, ok := cfg.GetExchangeManager().GetStorage().GetType().(*trino.ExchangeManagerStorage_ServiceS3_)
	if !ok {
		d := diag.NewErrorDiagnostic("Failed to parse Trino cluster value received from Cloud API",
			"ExchangeManager storage has unexpected type. Please update provider.")
		return NewRetryPolicyValueUnknown(), diag.Diagnostics{d}
	}

	serviceS3, diags := ServiceS3Value{state: attr.ValueStateKnown}.ToObjectValue(ctx)
	if diags.HasError() {
		return NewRetryPolicyValueUnknown(), diags
	}

	exchangeManagerAdditionalProperties, diags := types.MapValueFrom(ctx, types.StringType, cfg.ExchangeManager.AdditionalProperties)
	if diags.HasError() {
		return NewRetryPolicyValueUnknown(), diags
	}

	exchangeManagerValue := ExchangeManagerValue{
		AdditionalProperties: exchangeManagerAdditionalProperties,
		ServiceS3:            serviceS3,
		state:                attr.ValueStateKnown,
	}

	exchangeManagerObject, diags := exchangeManagerValue.ToObjectValue(ctx)
	if diags.HasError() {
		return NewRetryPolicyValueUnknown(), diags
	}

	value := RetryPolicyValue{
		AdditionalProperties: additionalProperties,
		ExchangeManager:      exchangeManagerObject,
		Policy:               types.StringValue(trino.RetryPolicyConfig_RetryPolicy_name[int32(cfg.Policy)]),
		state:                attr.ValueStateKnown,
	}

	return value, diag.Diagnostics{}
}

func nullableStringSliceToSet(ctx context.Context, s []string) (types.Set, diag.Diagnostics) {
	if s == nil {
		return types.SetNull(types.StringType), diag.Diagnostics{}
	}

	return types.SetValueFrom(ctx, types.StringType, s)
}

func loggingValueFromAPI(cfg *trino.LoggingConfig) (LoggingValue, diag.Diagnostics) {
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
	case *trino.LoggingConfig_FolderId:
		loggingValue.FolderId = types.StringValue(t.FolderId)
	case *trino.LoggingConfig_LogGroupId:
		loggingValue.LogGroupId = types.StringValue(t.LogGroupId)
	default:
		diags.AddError("Failed to parse Trino cluster value received from Cloud API",
			"Logging destination has unexpected type. Please update provider.")
		return NewLoggingValueNull(), diags
	}

	return loggingValue, diags
}

func maintenanceWindowFromAPI(mw *trino.MaintenanceWindow) (MaintenanceWindowValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if mw == nil {
		return NewMaintenanceWindowValueNull(), diags
	}

	var res MaintenanceWindowValue
	switch policy := mw.GetPolicy().(type) {
	case *trino.MaintenanceWindow_Anytime:
		res = MaintenanceWindowValue{
			MaintenanceWindowType: types.StringValue("ANYTIME"),
			state:                 attr.ValueStateKnown,
		}
	case *trino.MaintenanceWindow_WeeklyMaintenanceWindow:
		day := trino.WeeklyMaintenanceWindow_WeekDay_name[int32(policy.WeeklyMaintenanceWindow.GetDay())]
		res = MaintenanceWindowValue{
			MaintenanceWindowType: types.StringValue("WEEKLY"),
			Day:                   types.StringValue(day),
			Hour:                  types.Int64Value(policy.WeeklyMaintenanceWindow.GetHour()),
			state:                 attr.ValueStateKnown,
		}
	default:
		diags.AddError(
			"Failed to parse Trino maintenance window received from Cloud API",
			"Maintenance window has unexpected type",
		)
		return NewMaintenanceWindowValueNull(), diags
	}

	return res, diags
}
