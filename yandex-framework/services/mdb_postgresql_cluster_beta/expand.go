package mdb_postgresql_cluster_beta

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Set access to default if null
func expandAccess(ctx context.Context, cfgAccess types.Object, diags *diag.Diagnostics) *postgresql.Access {
	var access Access
	diags.Append(cfgAccess.As(ctx, &access, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})...)
	if diags.HasError() {
		return nil
	}
	return &postgresql.Access{
		WebSql:       access.WebSql.ValueBool(),
		DataLens:     access.DataLens.ValueBool(),
		DataTransfer: access.DataTransfer.ValueBool(),
		Serverless:   access.Serverless.ValueBool(),
	}
}

const (
	anytimeType = "ANYTIME"
	weeklyType  = "WEEKLY"
)

func expandClusterMaintenanceWindow(ctx context.Context, mw types.Object, diags *diag.Diagnostics) *postgresql.MaintenanceWindow {
	if mw.IsNull() || mw.IsUnknown() {
		return nil
	}

	out := &postgresql.MaintenanceWindow{}
	var mwConf MaintenanceWindow

	diags.Append(mw.As(ctx, &mwConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	if mwType := mwConf.Type.ValueString(); mwType == anytimeType {
		out.Policy = &postgresql.MaintenanceWindow_Anytime{
			Anytime: &postgresql.AnytimeMaintenanceWindow{},
		}
	} else if mwType == weeklyType {
		mwDay, mwHour := mwConf.Day.ValueString(), mwConf.Hour.ValueInt64()
		day := postgresql.WeeklyMaintenanceWindow_WeekDay_value[mwDay]

		out.Policy = &postgresql.MaintenanceWindow_WeeklyMaintenanceWindow{
			WeeklyMaintenanceWindow: &postgresql.WeeklyMaintenanceWindow{
				Hour: mwHour,
				Day:  postgresql.WeeklyMaintenanceWindow_WeekDay(day),
			},
		}
	} else {
		diags.AddError(
			"Failed to expand maintenance window.",
			fmt.Sprintf("maintenance_window.type should be %s or %s", anytimeType, weeklyType),
		)
		return nil
	}

	return out
}

func expandPerformanceDiagnostics(ctx context.Context, pd types.Object, diags *diag.Diagnostics) *postgresql.PerformanceDiagnostics {
	if pd.IsNull() || pd.IsUnknown() {
		return nil
	}
	var pdConf PerformanceDiagnostics

	diags.Append(pd.As(ctx, &pdConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &postgresql.PerformanceDiagnostics{
		Enabled:                    pdConf.Enabled.ValueBool(),
		SessionsSamplingInterval:   pdConf.SessionsSamplingInterval.ValueInt64(),
		StatementsSamplingInterval: pdConf.StatementsSamplingInterval.ValueInt64(),
	}
}

func expandBackupRetainPeriodDays(ctx context.Context, cfgBws types.Int64, diags *diag.Diagnostics) *wrapperspb.Int64Value {
	var pgBws *wrapperspb.Int64Value
	if !cfgBws.IsNull() && !cfgBws.IsUnknown() {
		pgBws = &wrapperspb.Int64Value{
			Value: cfgBws.ValueInt64(),
		}
	}

	return pgBws
}

func expandBackupWindowStart(ctx context.Context, cfgBws types.Object, diags *diag.Diagnostics) *timeofday.TimeOfDay {
	var backupWindowStart BackupWindowStart
	diags.Append(cfgBws.As(ctx, &backupWindowStart, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})...)
	if diags.HasError() {
		return nil
	}

	return &timeofday.TimeOfDay{
		Hours:   int32(backupWindowStart.Hours.ValueInt64()),
		Minutes: int32(backupWindowStart.Minutes.ValueInt64()),
	}
}
