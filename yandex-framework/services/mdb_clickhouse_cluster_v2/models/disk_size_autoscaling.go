package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type DiskSizeAutoscaling struct {
	DiskSizeLimit           types.Int64 `tfsdk:"disk_size_limit"`
	PlannedUsageThreshold   types.Int64 `tfsdk:"planned_usage_threshold"`
	EmergencyUsageThreshold types.Int64 `tfsdk:"emergency_usage_threshold"`
}

var DiskSizeAutoscalingAttrTypes = map[string]attr.Type{
	"disk_size_limit":           types.Int64Type,
	"planned_usage_threshold":   types.Int64Type,
	"emergency_usage_threshold": types.Int64Type,
}

func FlattenDiskSizeAutoscaling(ctx context.Context, diskSizeAutoscaling *clickhouse.DiskSizeAutoscaling, diags *diag.Diagnostics) types.Object {
	if diskSizeAutoscaling == nil {
		return types.ObjectNull(DiskSizeAutoscalingAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DiskSizeAutoscalingAttrTypes, DiskSizeAutoscaling{
			DiskSizeLimit:           types.Int64Value(datasize.ToGigabytes(diskSizeAutoscaling.DiskSizeLimit.GetValue())),
			PlannedUsageThreshold:   types.Int64Value(diskSizeAutoscaling.PlannedUsageThreshold.GetValue()),
			EmergencyUsageThreshold: types.Int64Value(diskSizeAutoscaling.EmergencyUsageThreshold.GetValue()),
		},
	)
	diags.Append(d...)

	return obj
}

func ExpandDiskSizeAutoscaling(ctx context.Context, diskSizeAutoscaling types.Object, diags *diag.Diagnostics) *clickhouse.DiskSizeAutoscaling {
	if diskSizeAutoscaling.IsNull() || diskSizeAutoscaling.IsUnknown() {
		return nil
	}

	var dsa DiskSizeAutoscaling
	if diags.Append(diskSizeAutoscaling.As(ctx, &dsa, datasize.DefaultOpts)...); diags.HasError() {
		return nil
	}

	return &clickhouse.DiskSizeAutoscaling{
		DiskSizeLimit:           wrapperspb.Int64(datasize.ToBytes(dsa.DiskSizeLimit.ValueInt64())),
		EmergencyUsageThreshold: wrapperspb.Int64(dsa.EmergencyUsageThreshold.ValueInt64()),
		PlannedUsageThreshold:   wrapperspb.Int64(dsa.PlannedUsageThreshold.ValueInt64()),
	}
}

func GetClickHouseDiskSizeAutoscaling(ctx context.Context, cluster Cluster, diags *diag.Diagnostics) (types.Object, bool) {
	if cluster.ClickHouse.IsNull() || cluster.ClickHouse.IsUnknown() {
		return types.ObjectNull(DiskSizeAutoscalingAttrTypes), false
	}

	var ch Clickhouse
	diags.Append(cluster.ClickHouse.As(ctx, &ch, datasize.DefaultOpts)...)
	if diags.HasError() {
		return types.ObjectNull(DiskSizeAutoscalingAttrTypes), false
	}

	if ch.DiskSizeAutoscaling.IsNull() || ch.DiskSizeAutoscaling.IsUnknown() {
		return ch.DiskSizeAutoscaling, false
	}

	return ch.DiskSizeAutoscaling, true
}
