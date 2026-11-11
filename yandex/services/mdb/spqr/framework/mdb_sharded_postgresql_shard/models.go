package mdb_sharded_postgresql_shard

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Shard struct {
	Id        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	Name      types.String `tfsdk:"name"`
	ShardSpec types.Object `tfsdk:"shard_spec"`
}

type ShardSpec struct {
	MDBPostgreSQLId types.String `tfsdk:"mdb_postgresql"`
}

var shardSpecType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"mdb_postgresql": types.StringType,
	},
}
