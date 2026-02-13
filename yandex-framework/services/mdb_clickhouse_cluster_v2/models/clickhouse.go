package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type Clickhouse struct {
	Config              types.Object `tfsdk:"config"`
	Resources           types.Object `tfsdk:"resources"`
	DiskSizeAutoscaling types.Object `tfsdk:"disk_size_autoscaling"`
}

var ClickhouseAttrTypes = map[string]attr.Type{
	"config":                types.ObjectType{AttrTypes: ClickhouseConfigAttrTypes},
	"resources":             types.ObjectType{AttrTypes: ResourcesAttrTypes},
	"disk_size_autoscaling": types.ObjectType{AttrTypes: DiskSizeAutoscalingAttrTypes},
}

func FlattenClickHouse(ctx context.Context, state *Cluster, clickhouse *clickhouse.ClusterConfig_Clickhouse, diags *diag.Diagnostics) types.Object {
	if clickhouse == nil {
		return types.ObjectNull(ClickhouseAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, ClickhouseAttrTypes, Clickhouse{
			Config:              FlattenClickHouseConfig(ctx, state, clickhouse.Config.EffectiveConfig, diags),
			Resources:           mdbcommon.FlattenResources(ctx, clickhouse.Resources, diags),
			DiskSizeAutoscaling: FlattenDiskSizeAutoscaling(ctx, clickhouse.DiskSizeAutoscaling, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func ExpandClickHouse(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouse.ConfigSpec_Clickhouse {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var clickhouseData Clickhouse
	diags.Append(c.As(ctx, &clickhouseData, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouse.ConfigSpec_Clickhouse{
		Config:              ExpandClickHouseConfig(ctx, clickhouseData.Config, diags),
		Resources:           mdbcommon.ExpandResources[clickhouse.Resources](ctx, clickhouseData.Resources, diags),
		DiskSizeAutoscaling: ExpandDiskSizeAutoscaling(ctx, clickhouseData.DiskSizeAutoscaling, diags),
	}
}
