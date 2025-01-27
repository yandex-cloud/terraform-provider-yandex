package mdb_postgresql_cluster_beta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
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
