package mdb_postgresql_cluster_beta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func prepareUpdateAfterCreateRequest(ctx context.Context, plan *Cluster) (*postgresql.UpdateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var paths []string
	mw := expandClusterMaintenanceWindow(ctx, plan.MaintenanceWindow, &diags)
	if mw != nil {
		paths = append(paths, "maintenance_window")
	}

	if diags.HasError() || len(paths) == 0 {
		return nil, diags
	}

	return &postgresql.UpdateClusterRequest{
		ClusterId:         plan.Id.ValueString(),
		MaintenanceWindow: expandClusterMaintenanceWindow(ctx, plan.MaintenanceWindow, &diags),
		UpdateMask:        &field_mask.FieldMask{Paths: paths},
	}, nil
}

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
		updateMaskPaths = append(updateMaskPaths, "config_spec.access")
	}

	if !plan.PerformanceDiagnostics.Equal(state.PerformanceDiagnostics) {
		config.SetPerformanceDiagnostics(expandPerformanceDiagnostics(ctx, plan.PerformanceDiagnostics, &diags))
		updateMaskPaths = append(updateMaskPaths, "config_spec.performance_diagnostics")
	}

	if !plan.BackupRetainPeriodDays.Equal(state.BackupRetainPeriodDays) {
		config.SetBackupRetainPeriodDays(expandBackupRetainPeriodDays(ctx, plan.BackupRetainPeriodDays, &diags))
		updateMaskPaths = append(updateMaskPaths, "config_spec.backup_retain_period_days")
	}

	if !plan.BackupWindowStart.Equal(state.BackupWindowStart) {
		config.SetBackupWindowStart(expandBackupWindowStart(ctx, plan.BackupWindowStart, &diags))
		updateMaskPaths = append(updateMaskPaths, "config_spec.backup_window_start")
	}

	return config, updateMaskPaths, diags
}
