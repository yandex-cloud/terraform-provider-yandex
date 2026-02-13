package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type Resources struct {
	ResourcePresetID types.String `tfsdk:"resource_preset_id"`
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	DiskTypeID       types.String `tfsdk:"disk_type_id"`
}

var ResourcesAttrTypes = map[string]attr.Type{
	"resource_preset_id": types.StringType,
	"disk_size":          types.Int64Type,
	"disk_type_id":       types.StringType,
}

func GetClickHouseResources(ctx context.Context, cluster Cluster, diags *diag.Diagnostics) (types.Object, bool) {
	if cluster.ClickHouse.IsNull() || cluster.ClickHouse.IsUnknown() {
		return types.ObjectNull(ResourcesAttrTypes), false
	}

	var ch Clickhouse
	diags.Append(cluster.ClickHouse.As(ctx, &ch, datasize.DefaultOpts)...)
	if diags.HasError() {
		return types.ObjectNull(ResourcesAttrTypes), false
	}

	if ch.Resources.IsNull() || ch.Resources.IsUnknown() {
		return ch.Resources, false
	}

	return ch.Resources, true
}
