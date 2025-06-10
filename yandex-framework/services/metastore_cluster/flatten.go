package metastore_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/metastore/v1"
)

func flattenStringSlice(ctx context.Context, s []string, diags *diag.Diagnostics) types.Set {
	if s == nil {
		return types.SetNull(types.StringType)
	}

	res, d := types.SetValueFrom(ctx, types.StringType, s)
	diags.Append(d...)
	return res
}

func flattenStringMap(ctx context.Context, m map[string]string, diags *diag.Diagnostics) types.Map {
	if m == nil {
		return types.MapNull(types.StringType)
	}

	res, d := types.MapValueFrom(ctx, types.StringType, m)
	diags.Append(d...)
	return res
}

func flattenLoggingConfig(cfg *metastore.LoggingConfig, diags *diag.Diagnostics) LoggingValue {
	if cfg == nil {
		return NewLoggingValueNull()
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
	case *metastore.LoggingConfig_FolderId:
		loggingValue.FolderId = types.StringValue(t.FolderId)
	case *metastore.LoggingConfig_LogGroupId:
		loggingValue.LogGroupId = types.StringValue(t.LogGroupId)
	default:
		diags.AddError("Failed to parse Metastore cluster value received from Cloud API",
			"Logging destination has unexpected type. Please update provider.")
		return NewLoggingValueNull()
	}

	return loggingValue
}

func flattenMaintenanceWindow(mw *metastore.MaintenanceWindow, diags *diag.Diagnostics) MaintenanceWindowValue {
	if mw == nil {
		return NewMaintenanceWindowValueNull()
	}

	var res MaintenanceWindowValue
	switch policy := mw.GetPolicy().(type) {
	case *metastore.MaintenanceWindow_Anytime:
		res = MaintenanceWindowValue{
			MaintenanceWindowType: types.StringValue("ANYTIME"),
			state:                 attr.ValueStateKnown,
		}
	case *metastore.MaintenanceWindow_WeeklyMaintenanceWindow:
		day := metastore.WeeklyMaintenanceWindow_WeekDay_name[int32(policy.WeeklyMaintenanceWindow.GetDay())]
		res = MaintenanceWindowValue{
			MaintenanceWindowType: types.StringValue("WEEKLY"),
			Day:                   types.StringValue(day),
			Hour:                  types.Int64Value(policy.WeeklyMaintenanceWindow.GetHour()),
			state:                 attr.ValueStateKnown,
		}
	default:
		diags.AddError(
			"Failed to parse Metastore maintenance window received from Cloud API",
			"Maintenance window has unexpected type",
		)
		return NewMaintenanceWindowValueNull()
	}

	return res
}
