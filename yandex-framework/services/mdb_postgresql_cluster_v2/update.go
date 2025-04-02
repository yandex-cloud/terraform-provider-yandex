package mdb_postgresql_cluster_v2

import (
	"context"
	"fmt"

	"maps"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func prepareVersionUpdateRequest(state, plan *Cluster) (*postgresql.UpdateClusterRequest, diag.Diagnostics) {

	const versionAttr = "version"

	var diags diag.Diagnostics

	sv := state.Config.Attributes()[versionAttr]
	pv := plan.Config.Attributes()[versionAttr]

	if pv.Equal(sv) {
		return nil, diags
	}

	pvVal, ok := pv.(types.String)
	if !ok {
		diags.AddError("Invalid version", "Version must be a string")
		return nil, diags
	}

	return &postgresql.UpdateClusterRequest{
		ClusterId: state.Id.ValueString(),
		ConfigSpec: &postgresql.ConfigSpec{
			Version: pvVal.ValueString(),
		},
		UpdateMask: &field_mask.FieldMask{Paths: []string{"config_spec.version"}},
	}, diags
}

func prepareUpdateRequest(ctx context.Context, state, plan *Cluster) (*postgresql.UpdateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	request := &postgresql.UpdateClusterRequest{
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
		request.SetLabels(expandLabels(ctx, plan.Labels, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "labels")
	}

	if !plan.Config.Equal(state.Config) {
		var planConfig Config
		diags := plan.Config.As(ctx, &planConfig, datasize.DefaultOpts)
		if diags.HasError() {
			return nil, diags
		}
		var stateConfig Config
		diags = state.Config.As(ctx, &stateConfig, datasize.DefaultOpts)
		if diags.HasError() {
			return nil, diags
		}

		config, updateMaskPaths, diags := prepareConfigChange(ctx, &planConfig, &stateConfig)
		if diags.HasError() {
			return nil, diags
		}

		request.SetConfigSpec(config)
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, updateMaskPaths...)
	}

	if !plan.DeletionProtection.Equal(state.DeletionProtection) {
		request.SetDeletionProtection(plan.DeletionProtection.ValueBool())
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "deletion_protection")
	}

	if !plan.SecurityGroupIds.Equal(state.SecurityGroupIds) {
		request.SetSecurityGroupIds(expandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "security_group_ids")
	}

	if !plan.MaintenanceWindow.Equal(state.MaintenanceWindow) {
		request.SetMaintenanceWindow(expandClusterMaintenanceWindow(ctx, plan.MaintenanceWindow, &diags))
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "maintenance_window")
	}

	return request, diags
}

func getPostgreSQLConfigFieldName(version string) string {
	return "postgresql_config_" + strings.Replace(version, "-", "_", -1)
}

func getAttrNamesFromConfigPGConfig(state *Config, diags *diag.Diagnostics) map[string]struct{} {

	attrs := make(map[string]struct{})
	if state.PostgtgreSQLConfig.IsNull() || state.PostgtgreSQLConfig.IsUnknown() {
		return attrs
	}

	for attr := range state.PostgtgreSQLConfig.Elements() {
		attrs[attr] = struct{}{}
	}

	return attrs
}

func prepareConfigChange(ctx context.Context, plan, state *Config) (*postgresql.ConfigSpec, []string, diag.Diagnostics) {
	var updateMaskPaths []string
	config := &postgresql.ConfigSpec{}
	diags := diag.Diagnostics{}

	if !plan.Resources.Equal(state.Resources) {
		config.SetResources(expandResources(ctx, plan.Resources, &diags))
		updateMaskPaths = append(updateMaskPaths, "config_spec.resources")
	}

	if !plan.Autofailover.Equal(state.Autofailover) {
		config.SetAutofailover(
			&wrapperspb.BoolValue{
				Value: plan.Autofailover.ValueBool(),
			},
		)
		updateMaskPaths = append(updateMaskPaths, "config_spec.autofailover")
	}

	if !plan.Access.Equal(state.Access) {
		config.SetAccess(expandAccess(ctx, plan.Access, &diags))
		updateMaskPaths = append(
			updateMaskPaths,
			"config_spec.access.web_sql",
			"config_spec.access.data_lens",
			"config_spec.access.data_transfer",
			"config_spec.access.serverless",
		)
	}

	if !plan.PerformanceDiagnostics.Equal(state.PerformanceDiagnostics) {
		config.SetPerformanceDiagnostics(expandPerformanceDiagnostics(ctx, plan.PerformanceDiagnostics, &diags))
		updateMaskPaths = append(
			updateMaskPaths,
			"config_spec.performance_diagnostics.enabled",
			"config_spec.performance_diagnostics.sessions_sampling_interval",
			"config_spec.performance_diagnostics.statements_sampling_interval",
		)
	}

	if !plan.BackupRetainPeriodDays.Equal(state.BackupRetainPeriodDays) {
		config.SetBackupRetainPeriodDays(expandBackupRetainPeriodDays(ctx, plan.BackupRetainPeriodDays, &diags))
		updateMaskPaths = append(updateMaskPaths, "config_spec.backup_retain_period_days")
	}

	if !plan.BackupWindowStart.Equal(state.BackupWindowStart) {
		config.SetBackupWindowStart(expandBackupWindowStart(ctx, plan.BackupWindowStart, &diags))
		updateMaskPaths = append(updateMaskPaths, "config_spec.backup_window_start")
	}

	if !plan.PoolerConfig.Equal(state.PoolerConfig) {
		config.SetPoolerConfig(expandPoolerConfig(ctx, plan.PoolerConfig, &diags))
		updateMaskPaths = append(
			updateMaskPaths,
			"config_spec.pooler_config.pooling_mode",
			"config_spec.pooler_config.pool_discard",
		)
	}

	if !plan.DiskSizeAutoscaling.Equal(state.DiskSizeAutoscaling) {
		config.SetDiskSizeAutoscaling(expandDiskSizeAutoscaling(ctx, plan.DiskSizeAutoscaling, &diags))
		updateMaskPaths = append(
			updateMaskPaths,
			"config_spec.disk_size_autoscaling.disk_size_limit",
			"config_spec.disk_size_autoscaling.planned_usage_threshold",
			"config_spec.disk_size_autoscaling.emergency_usage_threshold",
		)
	}

	if !plan.PostgtgreSQLConfig.Equal(state.PostgtgreSQLConfig) {
		config.SetPostgresqlConfig(expandPostgresqlConfig(ctx, plan.Version.ValueString(), plan.PostgtgreSQLConfig, &diags))

		attrsState := getAttrNamesFromConfigPGConfig(state, &diags)
		attrsPlan := getAttrNamesFromConfigPGConfig(plan, &diags)

		maps.Copy(attrsPlan, attrsState)
		for attr := range attrsPlan {
			updateMaskPaths = append(updateMaskPaths, fmt.Sprintf("config_spec.%s.%s", getPostgreSQLConfigFieldName(plan.Version.ValueString()), attr))
		}
	}

	return config, updateMaskPaths, diags
}
