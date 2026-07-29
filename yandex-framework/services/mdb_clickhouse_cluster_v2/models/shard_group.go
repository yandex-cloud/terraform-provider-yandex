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

type ExternalShardReplica struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Secure   types.Bool   `tfsdk:"secure"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	Priority types.Int64  `tfsdk:"priority"`
}

var ExternalShardReplicaAttrTypes = map[string]attr.Type{
	"host":     types.StringType,
	"port":     types.Int64Type,
	"secure":   types.BoolType,
	"user":     types.StringType,
	"password": types.StringType,
	"priority": types.Int64Type,
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

func buildExternalReplicaPasswordByHost(ctx context.Context, prevReplicas types.List, diags *diag.Diagnostics) map[string]types.String {
	result := map[string]types.String{}
	if prevReplicas.IsNull() || prevReplicas.IsUnknown() {
		return result
	}
	var replicas []ExternalShardReplica
	diags.Append(prevReplicas.ElementsAs(ctx, &replicas, false)...)
	for _, r := range replicas {
		result[r.Host.ValueString()] = r.Password
	}
	return result
}

func flattenExternalShardReplica(ctx context.Context, r *clickhouse.ExternalShard_Replica, prevPwd types.String, diags *diag.Diagnostics) types.Object {
	if r == nil {
		return types.ObjectNull(ExternalShardReplicaAttrTypes)
	}

	password := types.StringValue(r.Password)
	if shouldRestorePassword(password, prevPwd) {
		password = prevPwd
	}

	var port types.Int64
	if r.Port != nil {
		port = types.Int64Value(r.Port.Value)
	} else {
		port = types.Int64Null()
	}

	var secure types.Bool
	if r.Secure != nil {
		secure = types.BoolValue(r.Secure.Value)
	} else {
		secure = types.BoolNull()
	}

	var priority types.Int64
	if r.Priority != nil {
		priority = types.Int64Value(r.Priority.Value)
	} else {
		priority = types.Int64Null()
	}

	obj, d := types.ObjectValueFrom(ctx, ExternalShardReplicaAttrTypes, ExternalShardReplica{
		Host:     types.StringValue(r.Host),
		Port:     port,
		Secure:   secure,
		User:     types.StringValue(r.User),
		Password: password,
		Priority: priority,
	})
	diags.Append(d...)
	return obj
}

func flattenExternalShard(ctx context.Context, s *clickhouse.ExternalShard, prevShardObj types.Object, diags *diag.Diagnostics) types.Object {
	if s == nil {
		return types.ObjectNull(ExternalShardAttrTypes)
	}

	var prevShard ExternalShardModel
	if !prevShardObj.IsNull() && !prevShardObj.IsUnknown() {
		diags.Append(prevShardObj.As(ctx, &prevShard, datasize.DefaultOpts)...)
	}

	prevReplicaPasswordByHost := buildExternalReplicaPasswordByHost(ctx, prevShard.Replica, diags)

	var weight types.Int64
	if s.Weight != nil {
		weight = types.Int64Value(s.Weight.Value)
	} else {
		weight = types.Int64Null()
	}

	replicaObjs := make([]types.Object, len(s.Replicas))
	for i, r := range s.Replicas {
		replicaObjs[i] = flattenExternalShardReplica(ctx, r, prevReplicaPasswordByHost[r.Host], diags)
	}
	replicas, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: ExternalShardReplicaAttrTypes}, replicaObjs)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, ExternalShardAttrTypes, ExternalShardModel{
		Name:    types.StringValue(s.Name),
		Weight:  weight,
		Replica: replicas,
	})
	diags.Append(d...)
	return obj
}

func flattenShardGroup(ctx context.Context, group *clickhouse.ShardGroup, prevGroupObj types.Object, diags *diag.Diagnostics) types.Object {
	if group == nil {
		return types.ObjectNull(ShardGroupAttrTypes)
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

	externalShardObjs := make([]types.Object, len(group.ExternalShards))
	for i, s := range group.ExternalShards {
		prevShardObj := types.ObjectNull(ExternalShardAttrTypes)
		if prev, ok := prevShardByName[s.Name]; ok {
			prevShardObj = prev
		}
		externalShardObjs[i] = flattenExternalShard(ctx, s, prevShardObj, diags)
	}
	externalShards, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: ExternalShardAttrTypes}, externalShardObjs)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, ShardGroupAttrTypes, ShardGroup{
		Name:           types.StringValue(group.Name),
		Description:    types.StringValue(group.Description),
		ShardNames:     shardNames,
		ExternalShards: externalShards,
	})
	diags.Append(d...)
	return obj
}

func FlattenListShardGroup(ctx context.Context, groups []*clickhouse.ShardGroup, prevShardGroups types.List, diags *diag.Diagnostics) types.List {
	if groups == nil {
		return types.ListNull(types.ObjectType{AttrTypes: ShardGroupAttrTypes})
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
		prevGroupObj := types.ObjectNull(ShardGroupAttrTypes)
		if prev, ok := prevGroupByName[r.Name]; ok {
			prevGroupObj = prev
		}
		tfGroups[i] = flattenShardGroup(ctx, r, prevGroupObj, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: ShardGroupAttrTypes}, tfGroups)
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
