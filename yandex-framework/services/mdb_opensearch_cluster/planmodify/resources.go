package planmodify

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

// FixResourcesDiskSizeForAutoscaling aligns planned disk fields with state when autoscaling increased disk size in the cloud (same semantics as mdbcommon.FixDiskSizeOnAutoscalingChanges).
func fixResourcesDiskSizeForAutoscaling(ctx context.Context, plan, state types.Object, autoscalingEnabled bool, diags *diag.Diagnostics) types.Object {
	if state.IsNull() || !autoscalingEnabled {
		return plan
	}

	var planR, stateR model.NodeResource
	diags.Append(plan.As(ctx, &planR, datasize.DefaultOpts)...)
	diags.Append(state.As(ctx, &stateR, datasize.DefaultOpts)...)
	if diags.HasError() {
		return plan
	}

	if stateR.DiskSize.ValueInt64() > planR.DiskSize.ValueInt64() {
		planR.DiskSize = stateR.DiskSize
		planR.DiskSizeGb = types.Int64Value(datasize.ToGigabytes(stateR.DiskSize.ValueInt64()))
		obj, d := types.ObjectValueFrom(ctx, model.NodeResourceAttrTypes, planR)
		diags.Append(d...)
		return obj
	}
	return plan
}

// syncDiskSizeAutoscaling keeps disk_size_limit (bytes) and disk_size_gb_limit (GiB) in
// sync inside a planned DiskSizeAutoscaling object.
func syncDiskSizeAutoscaling(ctx context.Context, planObj types.Object, diags *diag.Diagnostics) types.Object {
	if planObj.IsNull() || planObj.IsUnknown() {
		return planObj
	}

	var p model.DiskSizeAutoscaling
	diags.Append(planObj.As(ctx, &p, datasize.DefaultOpts)...)
	if diags.HasError() {
		return planObj
	}

	switch {
	case p.DiskSizeGbLimit.IsUnknown():
		p.DiskSizeGbLimit = types.Int64Value(datasize.ToGigabytes(p.DiskSizeLimit.ValueInt64()))
	case p.DiskSizeLimit.IsUnknown():
		p.DiskSizeLimit = types.Int64Value(datasize.ToBytes(p.DiskSizeGbLimit.ValueInt64()))
	default:
		return planObj
	}

	obj, d := types.ObjectValueFrom(ctx, model.DiskSizeAutoscalingAttrTypes, p)
	diags.Append(d...)
	if diags.HasError() {
		return planObj
	}
	return obj
}

// syncResourcesDiskSize keeps disk_size (bytes) and disk_size_gb (GiB) in sync
// inside a planned NodeResource object.
func syncResourcesDiskSize(ctx context.Context, planRes types.Object, diags *diag.Diagnostics) types.Object {
	if planRes.IsNull() || planRes.IsUnknown() {
		return planRes
	}

	var p model.NodeResource
	diags.Append(planRes.As(ctx, &p, datasize.DefaultOpts)...)
	if diags.HasError() {
		return planRes
	}

	switch {
	case p.DiskSizeGb.IsUnknown():
		p.DiskSizeGb = types.Int64Value(datasize.ToGigabytes(p.DiskSize.ValueInt64()))
	case p.DiskSize.IsUnknown():
		p.DiskSize = types.Int64Value(datasize.ToBytes(p.DiskSizeGb.ValueInt64()))
	default:
		return planRes
	}

	obj, d := types.ObjectValueFrom(ctx, model.NodeResourceAttrTypes, p)
	diags.Append(d...)
	if diags.HasError() {
		return planRes
	}
	return obj
}
