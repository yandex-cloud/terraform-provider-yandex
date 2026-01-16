package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
)

type ShardGroup struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ShardNames  types.List   `tfsdk:"shard_names"`
}

var ShardGroupAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"description": types.StringType,
	"shard_names": types.ListType{ElemType: types.StringType},
}

func flattenShardGroup(ctx context.Context, group *clickhouse.ShardGroup, diags *diag.Diagnostics) types.Object {
	if group == nil {
		return types.ObjectNull(ShardGroupAttrTypes)
	}

	shardNames, d := types.ListValueFrom(ctx, types.StringType, group.ShardNames)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(
		ctx, ShardGroupAttrTypes, ShardGroup{
			Name:        types.StringValue(group.Name),
			Description: types.StringValue(group.Description),
			ShardNames:  shardNames,
		},
	)
	diags.Append(d...)

	return obj
}

func FlattenListShardGroup(ctx context.Context, groups []*clickhouse.ShardGroup, diags *diag.Diagnostics) types.List {
	if groups == nil {
		return types.ListNull(types.ObjectType{AttrTypes: ShardGroupAttrTypes})
	}

	tfGroups := make([]types.Object, len(groups))
	for i, r := range groups {
		tfGroups[i] = flattenShardGroup(ctx, r, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: ShardGroupAttrTypes}, tfGroups)
	diags.Append(d...)

	return list
}

func ExpandListShardGroup(ctx context.Context, g types.List, cid string, diags *diag.Diagnostics) []*clickhouse.ShardGroup {
	emptyList := []*clickhouse.ShardGroup{}

	if g.IsNull() || g.IsUnknown() {
		return emptyList
	}

	result := make([]*clickhouse.ShardGroup, 0, len(g.Elements()))
	groups := make([]ShardGroup, 0, len(g.Elements()))
	diags.Append(g.ElementsAs(ctx, &groups, false)...)
	if diags.HasError() {
		return emptyList
	}

	for _, group := range groups {
		var shardNames []string
		if !group.ShardNames.IsNull() && !group.ShardNames.IsUnknown() {
			diags.Append(group.ShardNames.ElementsAs(ctx, &shardNames, false)...)
			if diags.HasError() {
				return emptyList
			}
		}

		result = append(result, &clickhouse.ShardGroup{
			Name:        group.Name.ValueString(),
			ClusterId:   cid,
			Description: group.Description.ValueString(),
			ShardNames:  shardNames,
		})
	}

	return result
}
