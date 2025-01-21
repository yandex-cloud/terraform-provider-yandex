package cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
	"google.golang.org/genproto/protobuf/field_mask"
)

func PrepareUpdateParamsRequest(ctx context.Context, state, plan *model.OpenSearch) (*opensearch.UpdateClusterRequest, diag.Diagnostics) {
	clusterID := state.ID.ValueString()

	req := &opensearch.UpdateClusterRequest{
		ClusterId:  clusterID,
		UpdateMask: &field_mask.FieldMask{},
	}

	if !plan.Name.Equal(state.Name) {
		req.Name = plan.Name.ValueString()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if plan.Description.ValueString() != state.Description.ValueString() {
		req.Description = plan.Description.ValueString()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if !plan.Labels.Equal(state.Labels) {
		labels := make(map[string]string, len(plan.Labels.Elements()))
		diags := plan.Labels.ElementsAs(ctx, &labels, false)
		if diags.HasError() {
			return nil, diags
		}

		req.Labels = labels
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if !plan.Config.Equal(state.Config) {
		planConfig, stateConfig, diags := model.ParseGenerics(ctx, plan, state, model.ParseConfig)
		if diags.HasError() {
			return nil, diags
		}

		config, updateMaskPaths, diags := prepareConfigChange(ctx, planConfig, stateConfig)
		if diags.HasError() {
			return nil, diags
		}

		req.ConfigSpec = config
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, updateMaskPaths...)
	}

	if !plan.SecurityGroupIDs.Equal(state.SecurityGroupIDs) {
		securityGroupIDs := make([]string, 0, len(plan.SecurityGroupIDs.Elements()))
		diags := plan.SecurityGroupIDs.ElementsAs(ctx, &securityGroupIDs, false)
		if diags.HasError() {
			return nil, diags
		}

		req.SecurityGroupIds = securityGroupIDs
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "security_group_ids")
	}

	//this condition should be like this because of nil and "" are similar business values
	if plan.ServiceAccountID.ValueString() != state.ServiceAccountID.ValueString() {
		req.ServiceAccountId = plan.ServiceAccountID.ValueString()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "service_account_id")
	}

	if !plan.DeletionProtection.IsUnknown() && !plan.DeletionProtection.Equal(state.DeletionProtection) {
		req.DeletionProtection = plan.DeletionProtection.ValueBool()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "deletion_protection")
	}

	if !plan.MaintenanceWindow.Equal(state.MaintenanceWindow) {
		mw, diags := prepareMaintenanceWindow(ctx, plan)
		if diags.HasError() {
			return nil, diags
		}

		req.MaintenanceWindow = mw
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "maintenance_window")
	}

	if len(req.UpdateMask.Paths) == 0 {
		return nil, diag.Diagnostics{}
	}

	return req, diag.Diagnostics{}
}

func prepareConfigChange(ctx context.Context, plan, state *model.Config) (*opensearch.ConfigUpdateSpec, []string, diag.Diagnostics) {
	var updateMaskPaths []string
	config := &opensearch.ConfigUpdateSpec{}
	diags := diag.Diagnostics{}

	if !plan.Version.IsUnknown() && !plan.Version.Equal(state.Version) {
		config.Version = plan.Version.ValueString()
		updateMaskPaths = append(updateMaskPaths, "config_spec.version")
	}

	//do not check !AdminPassword.IsUnknown() because of planModifier useStateForUnknown
	if !plan.AdminPassword.Equal(state.AdminPassword) {
		config.AdminPassword = plan.AdminPassword.ValueString()
		updateMaskPaths = append(updateMaskPaths, "config_spec.admin_password")
	}

	//NOTE: all node_groups will be updated by different requests, so we skip it here and updates only plugins list
	planOpenSearchBlock, stateOpenSearchBlock, d := model.ParseGenerics(ctx, plan, state, model.ParseOpenSearchSubConfig)
	diags.Append(d...)
	if diags.HasError() {
		return nil, nil, diags
	}

	if !planOpenSearchBlock.Plugins.IsUnknown() && !planOpenSearchBlock.Plugins.Equal(stateOpenSearchBlock.Plugins) {
		plugins := make([]string, 0, len(planOpenSearchBlock.Plugins.Elements()))
		diags.Append(planOpenSearchBlock.Plugins.ElementsAs(ctx, &plugins, false)...)
		if diags.HasError() {
			return nil, nil, diags
		}

		config.OpensearchSpec = &opensearch.OpenSearchClusterUpdateSpec{
			Plugins: plugins,
		}
		updateMaskPaths = append(updateMaskPaths, "config_spec.opensearch_spec.plugins")
	}

	//NOTE: dashboards contains only node_groups so we can skip it here

	if !plan.Access.Equal(state.Access) {
		planAccess, diags := model.ParseAccess(ctx, plan)
		if diags.HasError() {
			return nil, nil, diags
		}

		stateAccess, diags := model.ParseAccess(ctx, state)
		if diags.HasError() {
			return nil, nil, diags
		}

		updateMaskPaths = append(updateMaskPaths, "config_spec.access")
		access := &opensearch.Access{}

		if !planAccess.DataTransfer.Equal(stateAccess.DataTransfer) {
			access.DataTransfer = planAccess.DataTransfer.ValueBool()
		}

		if !planAccess.Serverless.Equal(stateAccess.Serverless) {
			access.Serverless = planAccess.Serverless.ValueBool()
		}

		config.Access = access
	}

	return config, updateMaskPaths, diags
}
