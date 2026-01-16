package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
)

type Host struct {
	FQDN           types.String `tfsdk:"fqdn"`
	Zone           types.String `tfsdk:"zone"`
	Type           types.String `tfsdk:"type"`
	ShardName      types.String `tfsdk:"shard_name"`
	SubnetId       types.String `tfsdk:"subnet_id"`
	AssignPublicIp types.Bool   `tfsdk:"assign_public_ip"`
}

var HostType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"fqdn":             types.StringType,
		"zone":             types.StringType,
		"type":             types.StringType,
		"shard_name":       types.StringType,
		"subnet_id":        types.StringType,
		"assign_public_ip": types.BoolType,
	},
}

func (h Host) GetFQDN() types.String {
	return h.FQDN
}

func (h Host) GetShard() string {
	if h.Type.ValueString() == clickhouse.Host_ZOOKEEPER.String() {
		return "zk"
	}
	return h.ShardName.ValueString()
}
