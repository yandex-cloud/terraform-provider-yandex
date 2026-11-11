package model

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type DiskSizeAutoscaling struct {
	DiskSizeLimit           types.Int64 `tfsdk:"disk_size_limit"`
	DiskSizeGbLimit         types.Int64 `tfsdk:"disk_size_gb_limit"`
	PlannedUsageThreshold   types.Int64 `tfsdk:"planned_usage_threshold"`
	EmergencyUsageThreshold types.Int64 `tfsdk:"emergency_usage_threshold"`
}

var DiskSizeAutoscalingAttrTypes = map[string]attr.Type{
	"disk_size_limit":           types.Int64Type,
	"disk_size_gb_limit":        types.Int64Type,
	"planned_usage_threshold":   types.Int64Type,
	"emergency_usage_threshold": types.Int64Type,
}

func diskSizeAutoscalingToObject(ctx context.Context, r *opensearch.DiskSizeAutoscaling) (types.Object, diag.Diagnostics) {
	if isEmptyDiskSizeAutoscaling(r) {
		return types.ObjectNull(DiskSizeAutoscalingAttrTypes), diag.Diagnostics{}
	}

	bytes := r.GetDiskSizeLimit()
	return types.ObjectValueFrom(ctx, DiskSizeAutoscalingAttrTypes, DiskSizeAutoscaling{
		DiskSizeLimit:           types.Int64Value(bytes),
		DiskSizeGbLimit:         types.Int64Value(datasize.ToGigabytes(bytes)),
		PlannedUsageThreshold:   types.Int64Value(r.GetPlannedUsageThreshold()),
		EmergencyUsageThreshold: types.Int64Value(r.GetEmergencyUsageThreshold()),
	})
}

func isEmptyDiskSizeAutoscaling(r *opensearch.DiskSizeAutoscaling) bool {
	return r == nil ||
		(r.DiskSizeLimit == 0 && r.PlannedUsageThreshold == 0 && r.EmergencyUsageThreshold == 0)
}

type WithDiskSizeAutoscaling interface {
	GetDiskSizeAutoscaling() types.Object
}

func ParseNodeDiskSizeAutoscaling(ctx context.Context, ng WithDiskSizeAutoscaling) (*DiskSizeAutoscaling, diag.Diagnostics) {
	res := &DiskSizeAutoscaling{}
	diags := ng.GetDiskSizeAutoscaling().As(ctx, res, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}

// EffectiveDiskSizeLimitBytes returns the autoscaling disk size limit (in bytes) from either
// disk_size_limit or disk_size_gb_limit. When both are present (e.g. after Read), disk_size_limit
// is authoritative.
func EffectiveDiskSizeLimitBytes(r *DiskSizeAutoscaling) (int64, diag.Diagnostics) {
	var diags diag.Diagnostics
	hasBytes := !r.DiskSizeLimit.IsNull() && !r.DiskSizeLimit.IsUnknown()
	hasGb := !r.DiskSizeGbLimit.IsNull() && !r.DiskSizeGbLimit.IsUnknown()
	if hasBytes {
		return r.DiskSizeLimit.ValueInt64(), diags
	}
	if hasGb {
		return datasize.ToBytes(r.DiskSizeGbLimit.ValueInt64()), diags
	}
	return 0, diags
}
