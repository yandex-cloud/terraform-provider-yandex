package cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/utils"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func prepareUpdateRequest(ctx context.Context, state, plan *Cluster) (*postgresql.UpdateClusterRequest, diag.Diagnostics) {
	request := &postgresql.UpdateClusterRequest{
		ClusterId:  state.Id.ValueString(),
		UpdateMask: &field_mask.FieldMask{},
	}

	if !plan.Name.Equal(state.Name) {
		request.Name = plan.Name.ValueString()
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "name")
	}

	if !plan.Description.Equal(state.Description) {
		request.Description = plan.Description.ValueString()
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "description")
	}

	if !plan.Labels.Equal(state.Labels) {
		labels := make(map[string]string, len(plan.Labels.Elements()))
		diags := plan.Labels.ElementsAs(ctx, &labels, false)
		if diags.HasError() {
			return nil, diags
		}

		request.Labels = labels
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, "labels")
	}

	if !plan.Config.Equal(state.Config) {
		var planConfig Config
		diags := plan.Config.As(ctx, &planConfig, utils.DefaultOpts)
		if diags.HasError() {
			return nil, diags
		}
		var stateConfig Config
		diags = state.Config.As(ctx, &stateConfig, utils.DefaultOpts)
		if diags.HasError() {
			return nil, diags
		}

		config, updateMaskPaths, diags := prepareConfigChange(ctx, &planConfig, &stateConfig)
		if diags.HasError() {
			return nil, diags
		}

		request.ConfigSpec = config
		request.UpdateMask.Paths = append(request.UpdateMask.Paths, updateMaskPaths...)
	}

	return request, diag.Diagnostics{}
}

func prepareConfigChange(ctx context.Context, plan, state *Config) (*postgresql.ConfigSpec, []string, diag.Diagnostics) {
	var updateMaskPaths []string
	config := &postgresql.ConfigSpec{}
	diags := diag.Diagnostics{}

	if !plan.Version.IsUnknown() && !plan.Version.IsNull() && !plan.Version.Equal(state.Version) {
		config.Version = plan.Version.ValueString()
		updateMaskPaths = append(updateMaskPaths, "config_spec.version")
	}

	if !plan.Resources.IsUnknown() && !plan.Resources.IsNull() && !plan.Resources.Equal(state.Resources) {
		var resources Resources
		diags := plan.Resources.As(ctx, &resources, utils.DefaultOpts)
		if diags.HasError() {
			return nil, nil, diags
		}
		config.Resources = &postgresql.Resources{
			ResourcePresetId: resources.ResourcePresetID.ValueString(),
			DiskSize:         utils.ToBytes(resources.DiskSize.ValueInt64()),
			DiskTypeId:       resources.DiskTypeID.ValueString(),
		}
		updateMaskPaths = append(updateMaskPaths, "config_spec.resources")
	}

	if !plan.Resources.IsUnknown() && !plan.Resources.IsNull() && !plan.Resources.Equal(state.Autofailover) {
		config.SetAutofailover(
			&wrapperspb.BoolValue{
				Value: plan.Autofailover.ValueBool(),
			},
		)
		updateMaskPaths = append(updateMaskPaths, "config_spec.autofailover")
	}

	return config, updateMaskPaths, diags
}
