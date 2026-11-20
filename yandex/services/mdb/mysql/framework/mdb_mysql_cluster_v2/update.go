package mdb_mysql_cluster_v2

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	mysqlcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/common"
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
		if !pa.YandexQuery.Equal(sa.YandexQuery) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.access.yandex_query",
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

	if !plan.DiskSizeAutoscaling.Equal(state.DiskSizeAutoscaling) {
		updConf = true
		config.SetDiskSizeAutoscaling(expandDiskAutoScaling(ctx, plan.DiskSizeAutoscaling, &diags))

		var psda, sdsa DiskSizeAutoscaling
		diags.Append(state.DiskSizeAutoscaling.As(ctx, &sdsa, datasize.UnhandledOpts)...)
		diags.Append(plan.DiskSizeAutoscaling.As(ctx, &psda, datasize.UnhandledOpts)...)

		if !psda.DiskSizeLimit.Equal(sdsa.DiskSizeLimit) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.disk_size_autoscaling.disk_size_limit",
			)
		}
		if !psda.PlannedUsageThreshold.Equal(sdsa.PlannedUsageThreshold) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.disk_size_autoscaling.planned_usage_threshold",
			)
		}
		if !psda.EmergencyUsageThreshold.Equal(sdsa.EmergencyUsageThreshold) {
			request.UpdateMask.Paths = append(
				request.UpdateMask.Paths,
				"config_spec.disk_size_autoscaling.emergency_usage_threshold",
			)
		}
	}

	if !plan.BackupRetainPeriodDays.Equal(state.BackupRetainPeriodDays) {
		updConf = true
		config.SetBackupRetainPeriodDays(mdbcommon.ExpandInt64Wrapper(ctx, plan.BackupRetainPeriodDays, &diags))
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

	version := plan.Version.ValueString()
	versionChanged := !plan.Version.Equal(state.Version)
	settingsChanged := !plan.MySQLConfig.IsUnknown() && !plan.MySQLConfig.Equal(state.MySQLConfig)
	var mysqlConfig mysql.ConfigSpec_MysqlConfig
	if !plan.MySQLConfig.IsUnknown() && (settingsChanged || versionChanged) {
		mysqlConfig = expandMySQLConfig(ctx, version, plan.MySQLConfig, &diags)
		if diags.HasError() {
			return nil, diags
		}
	}
	if versionChanged && version == "9.7" {
		if !plan.MySQLConfig.IsUnknown() && !plan.MySQLConfig.IsNull() {
			updConf = true
			config.SetMysqlConfig(mysqlConfig)
			request.UpdateMask.Paths = append(request.UpdateMask.Paths, mysqlcommon.MySQL97UpgradeSettingsPaths()...)
		}
	} else if settingsChanged {
		updConf = true
		config.SetMysqlConfig(mysqlConfig)

		myFieldName, err := mysqlcommon.MySQLConfigFieldName(version)
		if err != nil {
			diags.AddError("Invalid version", fmt.Sprintf("Details: %v", err))
			return nil, diags
		}

		attrsState := mdbcommon.GetAttrNamesSetFromMap(state.MySQLConfig.MapValue, &diags)
		attrsPlan := mdbcommon.GetAttrNamesSetFromMap(plan.MySQLConfig.MapValue, &diags)

		if versionChanged {
			for attr := range attrsState {
				mask := &field_mask.FieldMask{Paths: []string{fmt.Sprintf("%s.%s", myFieldName, attr)}}
				if !mask.IsValid(config) {
					delete(attrsState, attr)
				}
			}
		}

		maps.Copy(attrsPlan, attrsState)
		for attr := range attrsPlan {
			request.UpdateMask.Paths = append(request.UpdateMask.Paths, fmt.Sprintf("config_spec.%s.%s", myFieldName, attr))
		}
	}

	if !plan.ConnectionManager.Equal(state.ConnectionManager) {
		updConf = true
		config.SetConnectionManager(mdbcommon.ExpandClusterConnectionManagerFramework(ctx, plan.ConnectionManager, &diags))
		cmPaths, d := mdbcommon.ClusterConnectionManagerUpdateMaskPaths(ctx, plan.ConnectionManager, state.ConnectionManager, "config_spec.connection_manager.")
		diags.Append(d...)
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, cmPaths...)
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
