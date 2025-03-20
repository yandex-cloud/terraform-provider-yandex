package mdb_mysql_cluster_beta

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/genproto/protobuf/field_mask"
)

func prepareVersionUpdateRequest(state, plan *Cluster) (*mysql.UpdateClusterRequest, diag.Diagnostics) {

	var diags diag.Diagnostics

	sv := state.Version
	pv := plan.Version

	if pv.Equal(sv) {
		return nil, diags
	}

	return &mysql.UpdateClusterRequest{
		ClusterId: state.Id.ValueString(),
		ConfigSpec: &mysql.ConfigSpec{
			Version: pv.ValueString(),
		},
		UpdateMask: &field_mask.FieldMask{Paths: []string{"config_spec.version"}},
	}, diags
}

func getMySQLConfigFieldName(version string) string {
	if version == "5.7" {
		return "mysql_config_5_7"
	}
	return "mysql_config_8_0"
}

func prepareUpdateRequest(ctx context.Context, state, plan *Cluster) (*mysql.UpdateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	request := &mysql.UpdateClusterRequest{
		ClusterId:  state.Id.ValueString(),
		UpdateMask: &field_mask.FieldMask{},
	}

	if !plan.Name.Equal(state.Name) {
		request.SetName(plan.Name.ValueString())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "name")
	}

	if !plan.Description.Equal(state.Description) {
		request.SetDescription(plan.Description.ValueString())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "description")
	}

	if !plan.Labels.Equal(state.Labels) {
		request.SetLabels(mdbcommon.ExpandLabels(ctx, plan.Labels, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "labels")
	}

	config := &mysql.ConfigSpec{}
	updConf := false

	if !plan.Resources.Equal(state.Resources) {
		updConf = true
		config.SetResources(mdbcommon.ExpandResources[mysql.Resources](ctx, plan.Resources, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "config_spec.resources")
	}

	if !plan.Access.Equal(state.Access) {
		updConf = true
		config.SetAccess(expandAccess(ctx, plan.Access, &diags))

		var pa, sa Access
		diags.Append(state.Access.As(ctx, &sa, datasize.UnhandledOpts)...)
		diags.Append(plan.Access.As(ctx, &pa, datasize.UnhandledOpts)...)

		if !pa.WebSql.Equal(sa.WebSql) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.access.web_sql",
			)
		}
		if !pa.DataTransfer.Equal(sa.DataTransfer) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.access.data_transfer",
			)
		}
		if !pa.DataLens.Equal(sa.DataLens) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.access.data_lens",
			)
		}
	}

	if !plan.PerformanceDiagnostics.Equal(state.PerformanceDiagnostics) {
		updConf = true
		config.SetPerformanceDiagnostics(expandPerformanceDiagnostics(ctx, plan.PerformanceDiagnostics, &diags))

		var ppd, spd PerformanceDiagnostics
		diags.Append(state.PerformanceDiagnostics.As(ctx, &spd, datasize.UnhandledOpts)...)
		diags.Append(plan.PerformanceDiagnostics.As(ctx, &ppd, datasize.UnhandledOpts)...)

		if !ppd.Enabled.Equal(spd.Enabled) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.performance_diagnostics.enabled",
			)
		}
		if !ppd.SessionsSamplingInterval.Equal(spd.SessionsSamplingInterval) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.performance_diagnostics.sessions_sampling_interval",
			)
		}
		if !ppd.StatementsSamplingInterval.Equal(spd.StatementsSamplingInterval) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.performance_diagnostics.statements_sampling_interval",
			)
		}
	}

	if !plan.BackupRetainPeriodDays.Equal(state.BackupRetainPeriodDays) {
		updConf = true
		config.SetBackupRetainPeriodDays(expandBackupRetainPeriodDays(ctx, plan.BackupRetainPeriodDays, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "config_spec.backup_retain_period_days")
	}

	if !plan.BackupWindowStart.Equal(state.BackupWindowStart) {
		updConf = true
		config.SetBackupWindowStart(mdbcommon.ExpandBackupWindow(ctx, plan.BackupWindowStart, &diags))

		var pbw, sbw BackupWindowStart
		diags.Append(state.BackupWindowStart.As(ctx, &sbw, datasize.UnhandledOpts)...)
		diags.Append(plan.BackupWindowStart.As(ctx, &pbw, datasize.UnhandledOpts)...)
		if !pbw.Hours.Equal(sbw.Hours) {
			request.UpdateMask.Paths = append(request.UpdateMask.Paths, "config_spec.backup_window_start.hours")
		}
		if !pbw.Minutes.Equal(sbw.Minutes) {
			request.UpdateMask.Paths = append(request.UpdateMask.Paths, "config_spec.backup_window_start.minutes")
		}
	}

	if !plan.MySQLConfig.Equal(state.MySQLConfig) {
		updConf = true
		config.SetMysqlConfig(expandMySQLConfig(ctx, plan.Version.ValueString(), plan.MySQLConfig, &diags))

		attrsState := mdbcommon.GetAttrNamesSetFromMap(state.MySQLConfig.MapValue, &diags)
		attrsPlan := mdbcommon.GetAttrNamesSetFromMap(plan.MySQLConfig.MapValue, &diags)

		maps.Copy(attrsPlan, attrsState)
		for attr := range attrsPlan {
			request.UpdateMask.Paths = append(request.UpdateMask.Paths, fmt.Sprintf("config_spec.%s.%s", getMySQLConfigFieldName(plan.Version.ValueString()), attr))
		}
	}

	if updConf {
		request.SetConfigSpec(config)
	}

	if !plan.DeletionProtection.Equal(state.DeletionProtection) {
		request.SetDeletionProtection(plan.DeletionProtection.ValueBool())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "deletion_protection")
	}

	if !plan.SecurityGroupIds.Equal(state.SecurityGroupIds) {
		request.SetSecurityGroupIds(mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "security_group_ids")
	}

	if !plan.MaintenanceWindow.Equal(state.MaintenanceWindow) {
		request.SetMaintenanceWindow(mdbcommon.ExpandClusterMaintenanceWindow[
			mysql.MaintenanceWindow,
			mysql.WeeklyMaintenanceWindow,
			mysql.AnytimeMaintenanceWindow,
			mysql.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "maintenance_window")
	}

	return request, diags
}
