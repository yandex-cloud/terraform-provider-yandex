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

type Zookeeper struct {
	Resources           types.Object `tfsdk:"resources"`
	DiskSizeAutoscaling types.Object `tfsdk:"disk_size_autoscaling"`
}

var ZookeeperAttrTypes = map[string]attr.Type{
	"resources":             types.ObjectType{AttrTypes: ResourcesAttrTypes},
	"disk_size_autoscaling": types.ObjectType{AttrTypes: DiskSizeAutoscalingAttrTypes},
}

func FlattenZooKeeper(ctx context.Context, zookeeper *clickhouse.ClusterConfig_Zookeeper, diags *diag.Diagnostics) types.Object {
	if zookeeper == nil {
		return types.ObjectNull(ZookeeperAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, ZookeeperAttrTypes, Zookeeper{
			Resources:           mdbcommon.FlattenResources(ctx, zookeeper.Resources, diags),
			DiskSizeAutoscaling: FlattenDiskSizeAutoscaling(ctx, zookeeper.DiskSizeAutoscaling, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func ExpandZooKeeper(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouse.ConfigSpec_Zookeeper {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var zookeeperData Zookeeper
	diags.Append(c.As(ctx, &zookeeperData, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouse.ConfigSpec_Zookeeper{
		Resources:           mdbcommon.ExpandResources[clickhouse.Resources](ctx, zookeeperData.Resources, diags),
		DiskSizeAutoscaling: ExpandDiskSizeAutoscaling(ctx, zookeeperData.DiskSizeAutoscaling, diags),
	}
}

// ZooKeeper is configured only when its resources are configured
func (z *Zookeeper) IsConfigured(ctx context.Context, diags *diag.Diagnostics) bool {
	if z == nil {
		return false
	}

	if z.Resources.IsNull() {
		return false
	}

	var resources Resources
	diags.Append(z.Resources.As(ctx, &resources, datasize.DefaultOpts)...)
	if diags.HasError() {
		return false
	}

	return resources.IsConfigured()
}
