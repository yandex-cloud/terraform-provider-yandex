package mdb_postgresql_cluster_beta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func flattenAccess(ctx context.Context, pgAccess *postgresql.Access, diags *diag.Diagnostics) types.Object {
	if pgAccess == nil {
		return types.ObjectNull(AccessAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, AccessAttrTypes, Access{
			DataLens:     types.BoolValue(pgAccess.DataLens),
			DataTransfer: types.BoolValue(pgAccess.DataTransfer),
			Serverless:   types.BoolValue(pgAccess.Serverless),
			WebSql:       types.BoolValue(pgAccess.WebSql),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenMaintenanceWindow(ctx context.Context, mw *postgresql.MaintenanceWindow, diags *diag.Diagnostics) types.Object {

	var maintenanceWindow MaintenanceWindow
	if mw != nil {
		switch p := mw.GetPolicy().(type) {
		case *postgresql.MaintenanceWindow_Anytime:
			maintenanceWindow.Type = types.StringValue("ANYTIME")
			// do nothing
		case *postgresql.MaintenanceWindow_WeeklyMaintenanceWindow:
			maintenanceWindow.Type = types.StringValue("WEEKLY")
			maintenanceWindow.Day = types.StringValue(
				postgresql.WeeklyMaintenanceWindow_WeekDay_name[int32(p.WeeklyMaintenanceWindow.GetDay())],
			)
			maintenanceWindow.Hour = types.Int64Value(p.WeeklyMaintenanceWindow.Hour)
		default:
			diags.AddError("Failed to flatten maintenance window.", "Unsupported PostgreSQL maintenance policy type.")
			return types.ObjectNull(MaintenanceWindowAttrTypes)
		}
	} else {
		diags.AddError("Failed to flatten maintenance window.", "Unsupported nil PostgreSQL maintenance window type.")
		return types.ObjectNull(MaintenanceWindowAttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, MaintenanceWindowAttrTypes, maintenanceWindow)
	diags.Append(d...)

	return obj
}

func flattenPerformanceDiagnostics(ctx context.Context, pd *postgresql.PerformanceDiagnostics, diags *diag.Diagnostics) types.Object {
	if pd == nil {
		return types.ObjectNull(PerformanceDiagnosticsAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, PerformanceDiagnosticsAttrTypes, PerformanceDiagnostics{
			Enabled:                    types.BoolValue(pd.Enabled),
			SessionsSamplingInterval:   types.Int64Value(pd.SessionsSamplingInterval),
			StatementsSamplingInterval: types.Int64Value(pd.StatementsSamplingInterval),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenBackupRetainPeriodDays(ctx context.Context, pgBrpd *wrapperspb.Int64Value, diags *diag.Diagnostics) types.Int64 {
	if pgBrpd == nil {
		return types.Int64Null()
	}
	return types.Int64Value(pgBrpd.GetValue())
}

func flattenBackupWindowStart(ctx context.Context, pgBws *timeofday.TimeOfDay, diags *diag.Diagnostics) types.Object {
	if pgBws == nil {
		return types.ObjectNull(BackupWindowStartAttrTypes)
	}

	bwsObj, d := types.ObjectValueFrom(ctx, BackupWindowStartAttrTypes, BackupWindowStart{
		Hours:   types.Int64Value(int64(pgBws.GetHours())),
		Minutes: types.Int64Value(int64(pgBws.GetMinutes())),
	})
	diags.Append(d...)
	return bwsObj
}

func flattenMapString(ctx context.Context, ms map[string]string, diags *diag.Diagnostics) types.Map {
	obj, d := types.MapValueFrom(ctx, types.StringType, ms)
	diags.Append(d...)
	return obj
}

func flattenSetString(ctx context.Context, ss []string, diags *diag.Diagnostics) types.Set {
	if ss == nil {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}

	obj, d := types.SetValueFrom(ctx, types.StringType, ss)
	diags.Append(d...)
	return obj
}

func flattenBoolWrapper(ctx context.Context, wb *wrapperspb.BoolValue, diags *diag.Diagnostics) types.Bool {
	if wb == nil {
		return types.BoolNull()
	}
	return types.BoolValue(wb.GetValue())
}

func flattenResources(ctx context.Context, r *postgresql.Resources, diags *diag.Diagnostics) types.Object {
	if r == nil {
		diags.AddError("Failed to flatten resources.", "Resources of cluster can't be nil. It's error in provider")
		return types.ObjectNull(ResourcesAttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, ResourcesAttrTypes, Resources{
		ResourcePresetID: types.StringValue(r.ResourcePresetId),
		DiskSize:         types.Int64Value(datasize.ToGigabytes(r.DiskSize)),
		DiskTypeID:       types.StringValue(r.DiskTypeId),
	})

	diags.Append(d...)
	return obj
}

func flattenConfig(ctx context.Context, c *postgresql.ClusterConfig, diags *diag.Diagnostics) types.Object {
	if c == nil {
		diags.AddError("Failed to flatten config.", "Config of cluster can't be nil. It's error in provider")
		return types.ObjectNull(ConfigAttrTypes)
	}

	obj, d := types.ObjectValueFrom(ctx, ConfigAttrTypes, Config{
		Version:                types.StringValue(c.Version),
		Resources:              flattenResources(ctx, c.Resources, diags),
		Autofailover:           flattenBoolWrapper(ctx, c.GetAutofailover(), diags),
		Access:                 flattenAccess(ctx, c.Access, diags),
		PerformanceDiagnostics: flattenPerformanceDiagnostics(ctx, c.PerformanceDiagnostics, diags),
		BackupRetainPeriodDays: flattenBackupRetainPeriodDays(ctx, c.BackupRetainPeriodDays, diags),
		BackupWindowStart:      flattenBackupWindowStart(ctx, c.BackupWindowStart, diags),
	})
	diags.Append(d...)
	return obj
}
