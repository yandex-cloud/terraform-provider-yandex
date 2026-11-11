package mdb_sharded_postgresql_shard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
)

var shardSpecTypeMap = map[string]func(attrs map[string]attr.Value, diags *diag.Diagnostics) spqr.ShardSpec_Spec{
	"mdb_postgresql": func(attrs map[string]attr.Value, diags *diag.Diagnostics) spqr.ShardSpec_Spec {
		mdbPostgresql, ok := attrs["mdb_postgresql"]
		if !ok {
			diags.AddError("Invalid type of Shard Spec", "mdb_postgresql is expected to be in shard spec")
			return nil
		}
		cidStr, ok := mdbPostgresql.(types.String)
		if !ok {
			diags.AddError("Invalid type of Shard Spec", "mdb_postgresql is expected to be of type string")
			return nil
		}
		cid := cidStr.ValueString()
		return &spqr.ShardSpec_MdbPostgresql{MdbPostgresql: &spqr.MDBPostgreSQL{ClusterId: cid}}
	},
}

func shardFromState(ctx context.Context, state *Shard) (*spqr.ShardSpec, diag.Diagnostics) {
	spec, diags := expandShardSpec(ctx, state.ShardSpec)
	shard := &spqr.ShardSpec{
		ShardName: state.Name.ValueString(),
		Spec:      spec,
	}
	return shard, diags
}

func expandShardSpec(ctx context.Context, shardSpec types.Object) (spqr.ShardSpec_Spec, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := shardSpec.Attributes()
	// FIXME: figure out how to understand which oneOf option is used
	expand, ok := shardSpecTypeMap["mdb_postgresql"]
	if !ok {
		diags.AddError("Internal provider error", "failed to find type of shard spec in map")
		return nil, diags
	}

	return expand(attrs, &diags), diags
}
