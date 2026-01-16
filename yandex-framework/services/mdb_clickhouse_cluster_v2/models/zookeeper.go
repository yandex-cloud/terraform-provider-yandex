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
	Resources types.Object `tfsdk:"resources"`
}

var ZookeeperAttrTypes = map[string]attr.Type{
	"resources": types.ObjectType{AttrTypes: ResourcesAttrTypes},
}

func FlattenZooKeeper(ctx context.Context, zookeeper *clickhouse.ClusterConfig_Zookeeper, diags *diag.Diagnostics) types.Object {
	if zookeeper == nil {
		return types.ObjectNull(ZookeeperAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, ZookeeperAttrTypes, Zookeeper{
			Resources: mdbcommon.FlattenResources(ctx, zookeeper.Resources, diags),
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
		Resources: mdbcommon.ExpandResources[clickhouse.Resources](ctx, zookeeperData.Resources, diags),
	}
}
