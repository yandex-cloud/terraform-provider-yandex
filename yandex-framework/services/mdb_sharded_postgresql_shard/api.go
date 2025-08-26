package mdb_sharded_postgresql_shard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

var shardedPostgreSQLAPI = ShardedPostgreSQLAPI{}

type ShardedPostgreSQLAPI struct{}

func (r *ShardedPostgreSQLAPI) ReadShard(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, shardname string) *spqr.Shard {
	shards, err := sdk.MDB().SPQR().Cluster().ListShards(ctx, &spqr.ListClusterShardsRequest{
		ClusterId: cid,
	})
	if err != nil {
		diags.AddError(
			"Failed to Read resources",
			fmt.Sprintf("Error while requesting API to get Sharded PostgreSQL shard: %s", err.Error()),
		)
		return nil
	}

	for _, u := range shards.GetShards() {
		if u.GetName() == shardname {
			return u
		}
	}

	diags.AddError(
		"Failed to Read resource",
		fmt.Sprintf("Sharded PostgreSQL shard %q not found", shardname),
	)
	return nil
}

func (r *ShardedPostgreSQLAPI) CreateShard(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, shardSpec *spqr.ShardSpec) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().SPQR().Cluster().AddShard(ctx, &spqr.AddClusterShardRequest{
			ClusterId: cid,
			ShardSpec: shardSpec,
		})
	})
	if err != nil {
		diags.AddError(
			"Failed to Create resource",
			fmt.Sprintf("Error while requesting API to create Sharded PostgreSQL shard: %s", err.Error()),
		)
		return
	}
	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to Create resource",
			fmt.Sprintf("Error while waiting for operation to create Sharded PostgreSQL shard: %s", err.Error()),
		)
	}
}

func (r *ShardedPostgreSQLAPI) UpdateShard(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, shardSpec *spqr.ShardSpec, updatePaths []string) {
	diags.AddError("update shard is not implemented yet", "")
}

func (r *ShardedPostgreSQLAPI) DeleteShard(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, shardname string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().SPQR().Cluster().DeleteShard(ctx, &spqr.DeleteClusterShardRequest{
			ClusterId: cid,
			ShardName: shardname,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			fmt.Sprintf("Error while requesting API to delete Sharded PostgreSQL shard: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			fmt.Sprintf("Error while waiting for operation to delete Sharded PostgreSQL shard: %s", err.Error()),
		)
	}
}
