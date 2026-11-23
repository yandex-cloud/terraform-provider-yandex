package models

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ExternalShardReplica struct {
	Host              types.String `tfsdk:"host"`
	Port              types.Int64  `tfsdk:"port"`
	Secure            types.Bool   `tfsdk:"secure"`
	User              types.String `tfsdk:"user"`
	Password          types.String `tfsdk:"password"`
	PasswordWo        types.String `tfsdk:"password_wo"`
	PasswordWoVersion types.Int64  `tfsdk:"password_wo_version"`
	Priority          types.Int64  `tfsdk:"priority"`
}

var ExternalShardReplicaAttrTypes = map[string]attr.Type{
	"host":                types.StringType,
	"port":                types.Int64Type,
	"secure":              types.BoolType,
	"user":                types.StringType,
	"password":            types.StringType,
	"password_wo":         types.StringType,
	"password_wo_version": types.Int64Type,
	"priority":            types.Int64Type,
}

type ExternalShardModel struct {
	Name    types.String `tfsdk:"name"`
	Weight  types.Int64  `tfsdk:"weight"`
	Replica types.List   `tfsdk:"replica"`
}

var ExternalShardAttrTypes = map[string]attr.Type{
	"name":    types.StringType,
	"weight":  types.Int64Type,
	"replica": types.ListType{ElemType: types.ObjectType{AttrTypes: ExternalShardReplicaAttrTypes}},
}

type ShardGroup struct {
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ShardNames     types.List   `tfsdk:"shard_names"`
	ExternalShards types.List   `tfsdk:"external_shard"`
}

var ShardGroupAttrTypes = map[string]attr.Type{
	"name":           types.StringType,
	"description":    types.StringType,
	"shard_names":    types.ListType{ElemType: types.StringType},
	"external_shard": types.ListType{ElemType: types.ObjectType{AttrTypes: ExternalShardAttrTypes}},
}

var DataSourceShardGroupAttrTypes = func() map[string]attr.Type {
	replica := maps.Clone(ExternalShardReplicaAttrTypes)
	delete(replica, "password_wo")
	delete(replica, "password_wo_version")
	shard := maps.Clone(ExternalShardAttrTypes)
	shard["replica"] = types.ListType{ElemType: types.ObjectType{AttrTypes: replica}}
	group := maps.Clone(ShardGroupAttrTypes)
	group["external_shard"] = types.ListType{ElemType: types.ObjectType{AttrTypes: shard}}
	return group
}()

func flattenExternalShardReplica(ctx context.Context, r *clickhouse.ExternalShard_Replica, prev types.Object, diags *diag.Diagnostics) types.Object {
	attrTypes := prev.AttributeTypes(ctx)
	if r == nil {
		return types.ObjectNull(attrTypes)
	}

	password := types.StringValue(r.Password)
	version := types.Int64Null()
	_, isResource := attrTypes["password_wo_version"]
	if isResource {
		password = types.StringNull()
		if !prev.IsNull() && !prev.IsUnknown() {
			password = prev.Attributes()["password"].(types.String)
			version = prev.Attributes()["password_wo_version"].(types.Int64)
		}
	}

	values := map[string]attr.Value{
		"host":     types.StringValue(r.Host),
		"port":     mdbcommon.FlattenInt64Wrapper(ctx, r.Port, diags),
		"secure":   mdbcommon.FlattenBoolWrapper(ctx, r.Secure, diags),
		"user":     types.StringValue(r.User),
		"password": password,
		"priority": mdbcommon.FlattenInt64Wrapper(ctx, r.Priority, diags),
	}
	if isResource {
		values["password_wo"] = types.StringNull()
		values["password_wo_version"] = version
	}
	obj, d := types.ObjectValue(attrTypes, values)
	diags.Append(d...)
	return obj
}

func flattenExternalShard(ctx context.Context, s *clickhouse.ExternalShard, prevShardObj types.Object, diags *diag.Diagnostics) types.Object {
	attrTypes := prevShardObj.AttributeTypes(ctx)
	replicaType := attrTypes["replica"].(types.ListType).ElemType.(types.ObjectType)
	if s == nil {
		return types.ObjectNull(attrTypes)
	}

	var prevShard ExternalShardModel
	if !prevShardObj.IsNull() && !prevShardObj.IsUnknown() {
		diags.Append(prevShardObj.As(ctx, &prevShard, datasize.DefaultOpts)...)
	}

	prevReplicaByHost := map[string]types.Object{}
	for _, elem := range prevShard.Replica.Elements() {
		replica := elem.(types.Object)
		prevReplicaByHost[replica.Attributes()["host"].(types.String).ValueString()] = replica
	}

	replicaObjs := make([]types.Object, len(s.Replicas))
	for i, r := range s.Replicas {
		prevReplica := types.ObjectNull(replicaType.AttrTypes)
		if prev, ok := prevReplicaByHost[r.Host]; ok {
			prevReplica = prev
		}
		replicaObjs[i] = flattenExternalShardReplica(ctx, r, prevReplica, diags)
	}
	replicas, d := types.ListValueFrom(ctx, replicaType, replicaObjs)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, attrTypes, ExternalShardModel{
		Name:    types.StringValue(s.Name),
		Weight:  mdbcommon.FlattenInt64Wrapper(ctx, s.Weight, diags),
		Replica: replicas,
	})
	diags.Append(d...)
	return obj
}

func flattenShardGroup(ctx context.Context, group *clickhouse.ShardGroup, prevGroupObj types.Object, diags *diag.Diagnostics) types.Object {
	attrTypes := prevGroupObj.AttributeTypes(ctx)
	shardType := attrTypes["external_shard"].(types.ListType).ElemType.(types.ObjectType)
	if group == nil {
		return types.ObjectNull(attrTypes)
	}

	var prevGroup ShardGroup
	if !prevGroupObj.IsNull() && !prevGroupObj.IsUnknown() {
		diags.Append(prevGroupObj.As(ctx, &prevGroup, datasize.DefaultOpts)...)
	}

	prevShardByName := map[string]types.Object{}
	if !prevGroup.ExternalShards.IsNull() && !prevGroup.ExternalShards.IsUnknown() {
		for _, elem := range prevGroup.ExternalShards.Elements() {
			prevShardObj := elem.(types.Object)
			name := prevShardObj.Attributes()["name"].(types.String).ValueString()
			prevShardByName[name] = prevShardObj
		}
	}

	shardNames, d := types.ListValueFrom(ctx, types.StringType, group.ShardNames)
	diags.Append(d...)
	if shardNames.IsNull() {
		shardNames = types.ListValueMust(types.StringType, []attr.Value{})
	}

	externalShardObjs := make([]types.Object, len(group.ExternalShards))
	for i, s := range group.ExternalShards {
		prevShardObj := types.ObjectNull(shardType.AttrTypes)
		if prev, ok := prevShardByName[s.Name]; ok {
			prevShardObj = prev
		}
		externalShardObjs[i] = flattenExternalShard(ctx, s, prevShardObj, diags)
	}
	externalShards, d := types.ListValueFrom(ctx, shardType, externalShardObjs)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, attrTypes, ShardGroup{
		Name:           types.StringValue(group.Name),
		Description:    types.StringValue(group.Description),
		ShardNames:     shardNames,
		ExternalShards: externalShards,
	})
	diags.Append(d...)
	return obj
}

func FlattenListShardGroup(ctx context.Context, groups []*clickhouse.ShardGroup, prevShardGroups types.List, diags *diag.Diagnostics) types.List {
	groupType := prevShardGroups.ElementType(ctx).(types.ObjectType)
	if groups == nil {
		return types.ListNull(groupType)
	}

	prevGroupByName := map[string]types.Object{}
	if !prevShardGroups.IsNull() && !prevShardGroups.IsUnknown() {
		for _, elem := range prevShardGroups.Elements() {
			prevGroupObj := elem.(types.Object)
			name := prevGroupObj.Attributes()["name"].(types.String).ValueString()
			prevGroupByName[name] = prevGroupObj
		}
	}

	tfGroups := make([]types.Object, len(groups))
	for i, r := range groups {
		prevGroupObj := types.ObjectNull(groupType.AttrTypes)
		if prev, ok := prevGroupByName[r.Name]; ok {
			prevGroupObj = prev
		}
		tfGroups[i] = flattenShardGroup(ctx, r, prevGroupObj, diags)
	}

	list, d := types.ListValueFrom(ctx, groupType, tfGroups)
	diags.Append(d...)
	return list
}

func expandExternalShardReplica(r ExternalShardReplica) *clickhouse.ExternalShard_Replica {
	replica := &clickhouse.ExternalShard_Replica{
		Host:     r.Host.ValueString(),
		User:     r.User.ValueString(),
		Password: r.Password.ValueString(),
	}
	if !r.Port.IsNull() && !r.Port.IsUnknown() {
		replica.Port = wrapperspb.Int64(r.Port.ValueInt64())
	}
	if !r.Secure.IsNull() && !r.Secure.IsUnknown() {
		replica.Secure = wrapperspb.Bool(r.Secure.ValueBool())
	}
	if !r.Priority.IsNull() && !r.Priority.IsUnknown() {
		replica.Priority = wrapperspb.Int64(r.Priority.ValueInt64())
	}
	return replica
}

func expandExternalShard(ctx context.Context, s ExternalShardModel, diags *diag.Diagnostics) *clickhouse.ExternalShard {
	shard := &clickhouse.ExternalShard{
		Name: s.Name.ValueString(),
	}
	if !s.Weight.IsNull() && !s.Weight.IsUnknown() {
		shard.Weight = wrapperspb.Int64(s.Weight.ValueInt64())
	}
	if !s.Replica.IsNull() && !s.Replica.IsUnknown() {
		replicas := make([]ExternalShardReplica, 0, len(s.Replica.Elements()))
		diags.Append(s.Replica.ElementsAs(ctx, &replicas, false)...)
		if diags.HasError() {
			return shard
		}
		shard.Replicas = make([]*clickhouse.ExternalShard_Replica, len(replicas))
		for i, r := range replicas {
			shard.Replicas[i] = expandExternalShardReplica(r)
		}
	}
	return shard
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

		var externalShards []*clickhouse.ExternalShard
		if !group.ExternalShards.IsNull() && !group.ExternalShards.IsUnknown() {
			shardModels := make([]ExternalShardModel, 0, len(group.ExternalShards.Elements()))
			diags.Append(group.ExternalShards.ElementsAs(ctx, &shardModels, false)...)
			if diags.HasError() {
				return emptyList
			}
			for _, sm := range shardModels {
				externalShards = append(externalShards, expandExternalShard(ctx, sm, diags))
				if diags.HasError() {
					return emptyList
				}
			}
		}

		result = append(result, &clickhouse.ShardGroup{
			Name:           group.Name.ValueString(),
			ClusterId:      cid,
			Description:    group.Description.ValueString(),
			ShardNames:     shardNames,
			ExternalShards: externalShards,
		})
	}

	return result
}
