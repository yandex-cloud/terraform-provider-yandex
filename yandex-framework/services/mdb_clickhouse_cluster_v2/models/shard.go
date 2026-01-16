package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type Shard struct {
	Weight    types.Int64  `tfsdk:"weight"`
	Resources types.Object `tfsdk:"resources"`
}

var ShardAttrTypes = map[string]attr.Type{
	"weight":    types.Int64Type,
	"resources": types.ObjectType{AttrTypes: ResourcesAttrTypes},
}

func flattenShard(ctx context.Context, shard *clickhouse.Shard, diags *diag.Diagnostics) types.Object {
	if shard == nil {
		return types.ObjectNull(ShardAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, ShardAttrTypes, Shard{
			Weight:    types.Int64Value(shard.Config.Clickhouse.Weight.Value),
			Resources: mdbcommon.FlattenResources(ctx, shard.Config.Clickhouse.Resources, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func FlattenListShard(ctx context.Context, shards []*clickhouse.Shard, diags *diag.Diagnostics) types.Map {
	if shards == nil {
		return types.MapNull(types.ObjectType{AttrTypes: ShardAttrTypes})
	}

	tfShards := make(map[string]attr.Value, len(shards))
	for _, s := range shards {
		tfShards[s.Name] = flattenShard(ctx, s, diags)
	}

	m, d := types.MapValue(types.ObjectType{AttrTypes: ShardAttrTypes}, tfShards)
	diags.Append(d...)

	return m
}

func ExpandListShard(ctx context.Context, m types.Map, cid string, diags *diag.Diagnostics) []*clickhouse.ShardSpec {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}

	result := make([]*clickhouse.ShardSpec, 0, len(m.Elements()))
	var shards map[string]Shard
	diags.Append(m.ElementsAs(ctx, &shards, false)...)
	if diags.HasError() {
		return nil
	}

	for shardName, shard := range shards {
		result = append(result, &clickhouse.ShardSpec{
			Name: shardName,
			ConfigSpec: &clickhouse.ShardConfigSpec{
				Clickhouse: &clickhouse.ShardConfigSpec_Clickhouse{
					Weight:    mdbcommon.ExpandInt64Wrapper(ctx, shard.Weight, diags),
					Resources: mdbcommon.ExpandResources[clickhouse.Resources](ctx, shard.Resources, diags),
				},
			},
		})
	}

	return result
}
