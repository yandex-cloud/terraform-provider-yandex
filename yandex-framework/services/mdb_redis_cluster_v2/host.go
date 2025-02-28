package mdb_redis_cluster_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	redisproto "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type Host struct {
	Zone            types.String `tfsdk:"zone"`
	ShardName       types.String `tfsdk:"shard_name"`
	SubnetId        types.String `tfsdk:"subnet_id"`
	FQDN            types.String `tfsdk:"fqdn"`
	ReplicaPriority types.Int64  `tfsdk:"replica_priority"`
	AssignPublicIp  types.Bool   `tfsdk:"assign_public_ip"`
}

var HostType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"zone":             types.StringType,
		"shard_name":       types.StringType,
		"subnet_id":        types.StringType,
		"fqdn":             types.StringType,
		"replica_priority": types.Int64Type,
		"assign_public_ip": types.BoolType,
	},
}

var redisHostService = &RedisHostService{}

type RedisHostService struct {
}

func (r RedisHostService) FullyMatch(planHost Host, stateHost Host) bool {
	return planHost.Zone.ValueString() == stateHost.Zone.ValueString() &&
		(planHost.SubnetId.IsUnknown() || planHost.SubnetId.ValueString() == stateHost.SubnetId.ValueString()) &&
		planHost.ReplicaPriority.ValueInt64() == stateHost.ReplicaPriority.ValueInt64() &&
		planHost.AssignPublicIp.ValueBool() == stateHost.AssignPublicIp.ValueBool() &&
		(planHost.ShardName.IsUnknown() || planHost.ShardName.ValueString() == stateHost.ShardName.ValueString())
}

func (r RedisHostService) PartialMatch(planHost Host, stateHost Host) bool {
	return planHost.Zone.Equal(stateHost.Zone) &&
		(planHost.FQDN.IsUnknown() || planHost.FQDN.Equal(stateHost.FQDN)) &&
		(planHost.SubnetId.IsUnknown() || planHost.SubnetId.Equal(stateHost.SubnetId)) &&
		(planHost.ShardName.IsUnknown() || planHost.ShardName.Equal(stateHost.ShardName))
}

func (r RedisHostService) GetChanges(plan Host, state Host) (*redisproto.UpdateHostSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !r.PartialMatch(plan, state) {
		diags.AddError(
			"Wrong changes for host",
			"Attributes shard_name, zone, subnet_id can't be changed. Try to replace this host to new one",
		)
		return nil, diags
	}
	if plan.AssignPublicIp.Equal(state.AssignPublicIp) && plan.ReplicaPriority.Equal(state.ReplicaPriority) {
		return nil, nil
	}
	return &redisproto.UpdateHostSpec{
		HostName: state.FQDN.ValueString(),
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"assign_public_ip", "replica_priority"},
		},
		AssignPublicIp:  plan.AssignPublicIp.ValueBool(),
		ReplicaPriority: wrappers.Int64FromTF(plan.ReplicaPriority),
	}, diags
}

func (r RedisHostService) ConvertToProto(h Host) *redisproto.HostSpec {
	return &redisproto.HostSpec{
		ZoneId:          h.Zone.ValueString(),
		ShardName:       h.ShardName.ValueString(),
		SubnetId:        h.SubnetId.ValueString(),
		AssignPublicIp:  h.AssignPublicIp.ValueBool(),
		ReplicaPriority: wrappers.Int64FromTF(h.ReplicaPriority),
	}
}

func (r RedisHostService) ConvertFromProto(apiHost *redisproto.Host) Host {
	return Host{
		Zone:            types.StringValue(apiHost.ZoneId),
		ShardName:       types.StringValue(apiHost.ShardName),
		SubnetId:        types.StringValue(apiHost.SubnetId),
		AssignPublicIp:  types.BoolValue(apiHost.AssignPublicIp),
		ReplicaPriority: types.Int64Value(apiHost.ReplicaPriority.GetValue()),
		FQDN:            types.StringValue(apiHost.Name),
	}
}

func (h Host) GetFQDN() types.String {
	return h.FQDN
}
func (h Host) GetShard() string {
	return h.ShardName.ValueString()
}
