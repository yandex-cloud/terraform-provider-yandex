package mdb_sharded_postgresql_shard

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
)

func shardToState(shard *spqr.Shard, state *Shard, cid string) diag.Diagnostics {
	state.ClusterID = types.StringValue(cid)
	state.Name = types.StringValue(shard.Name)
	state.ShardSpec = types.ObjectValueMust(shardSpecType.AttrTypes, map[string]attr.Value{
		"mdb_postgresql": types.StringValue(shard.ClusterId),
	})
	return diag.Diagnostics{}
}
