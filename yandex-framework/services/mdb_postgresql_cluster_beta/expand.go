package mdb_postgresql_cluster_beta

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
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

func expandLabels(ctx context.Context, labels types.Map, diags *diag.Diagnostics) map[string]string {
	var lMap map[string]string
	if !(labels.IsUnknown() || labels.IsNull()) {
		diags.Append(labels.ElementsAs(ctx, &lMap, false)...)
		if diags.HasError() {
			return nil
		}
	}
	return lMap
}

func expandEnvironment(_ context.Context, e types.String, diags *diag.Diagnostics) postgresql.Cluster_Environment {

	if e.IsNull() || e.IsUnknown() {
		return 0
	}

	v, ok := postgresql.Cluster_Environment_value[e.ValueString()]
	if !ok || v == 0 {
		allowedEnvs := make([]string, 0, len(postgresql.Cluster_Environment_value))
		for k, v := range postgresql.Cluster_Environment_value {
			if v == 0 {
				continue
			}
			allowedEnvs = append(allowedEnvs, k)
		}

		diags.AddError(
			"Failed to parse PostgreSQL environment",
			fmt.Sprintf("Error while parsing value for 'environment'. Value must be one of `%s`, not `%s`", strings.Join(allowedEnvs, "`, `"), e),
		)

		return 0
	}
	return postgresql.Cluster_Environment(v)
}

func expandBoolWrapper(_ context.Context, b types.Bool, _ *diag.Diagnostics) *wrapperspb.BoolValue {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}

	return wrapperspb.Bool(b.ValueBool())
}

func expandSecurityGroupIds(ctx context.Context, sg types.Set, diags *diag.Diagnostics) []string {
	var securityGroupIds []string
	if !(sg.IsUnknown() || sg.IsNull()) {
		securityGroupIds = make([]string, len(sg.Elements()))
		diags.Append(sg.ElementsAs(ctx, &securityGroupIds, false)...)
		if diags.HasError() {
			return nil
		}
	}

	return securityGroupIds
}

func expandResources(ctx context.Context, r types.Object, diags *diag.Diagnostics) *postgresql.Resources {
	var resources Resources
	diags.Append(r.As(ctx, &resources, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &postgresql.Resources{
		ResourcePresetId: resources.ResourcePresetID.ValueString(),
		DiskTypeId:       resources.DiskTypeID.ValueString(),
		DiskSize:         datasize.ToBytes(resources.DiskSize.ValueInt64()),
	}
}

func expandConfig(ctx context.Context, c types.Object, diags *diag.Diagnostics) *postgresql.ConfigSpec {
	var configSpec Config
	diags.Append(c.As(ctx, &configSpec, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &postgresql.ConfigSpec{
		Version:                configSpec.Version.ValueString(),
		Resources:              expandResources(ctx, configSpec.Resources, diags),
		Autofailover:           expandBoolWrapper(ctx, configSpec.Autofailover, diags),
		Access:                 expandAccess(ctx, configSpec.Access, diags),
		PerformanceDiagnostics: expandPerformanceDiagnostics(ctx, configSpec.PerformanceDiagnostics, diags),
		BackupRetainPeriodDays: expandBackupRetainPeriodDays(ctx, configSpec.BackupRetainPeriodDays, diags),
		BackupWindowStart:      expandBackupWindowStart(ctx, configSpec.BackupWindowStart, diags),
	}
}

func expandFolderId(ctx context.Context, f types.String, providerConfig *config.State, diags *diag.Diagnostics) string {
	folderID, d := validate.FolderID(f, providerConfig)
	diags.Append(d)
	return folderID
}
