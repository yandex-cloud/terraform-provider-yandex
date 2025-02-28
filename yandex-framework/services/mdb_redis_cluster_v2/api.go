package mdb_redis_cluster_v2

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

const (
	defaultMDBPageSize = 1000
)

var redisAPI = RedisAPI{}

type RedisAPI struct {
}

func (r *RedisAPI) GetCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) *redis.Cluster {
	db, err := sdk.MDB().Redis().Cluster().Get(ctx, &redis.GetClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		diag.AddError(
			"API Error Reading",
			fmt.Sprintf("Error while requesting API to read Redis cluster %q: %s", cid, err.Error()),
		)
		return nil
	}
	return db
}

func (r *RedisAPI) DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) {
	op, err := sdk.WrapOperation(sdk.MDB().Redis().Cluster().Delete(ctx, &redis.DeleteClusterRequest{
		ClusterId: cid,
	}))

	if err != nil {
		diag.AddError(
			"API Error Deleting",
			fmt.Sprintf("Error while requesting API to delete Redis cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"API Error Deleting",
			fmt.Sprintf("Error while waiting for operation %q to delete Redis cluster %q: %s", op.Id(), cid, err.Error()),
		)
	}
}

func (r *RedisAPI) CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *redis.CreateClusterRequest) string {
	op, err := sdk.WrapOperation(sdk.MDB().Redis().Cluster().Create(ctx, req))
	if err != nil {
		diag.AddError(
			"API Error Creating",
			fmt.Sprintf("Error while requesting API to create Redis cluster: %s", err.Error()),
		)
		return ""
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diag.AddError(
			"API Error Creating",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata: %s", op.Id(), err.Error()),
		)
		return ""
	}

	md, ok := protoMetadata.(*redis.CreateClusterMetadata)
	if !ok {
		diag.AddError(
			"API Error Creating",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata", op.Id()),
		)
		return ""
	}

	log.Printf("[DEBUG] Creating Redis Cluster %q", md.ClusterId)

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"API Error Creating",
			fmt.Sprintf("Error while waiting for operation %q to create Redis cluster: %s", op.Id(), err.Error()),
		)
		return ""
	}

	return md.ClusterId
}

func (r *RedisAPI) UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *redis.UpdateClusterRequest) {
	op, err := sdk.WrapOperation(sdk.MDB().Redis().Cluster().Update(ctx, req))
	if err != nil {
		diag.AddError(
			"API Error Updating",
			fmt.Sprintf("Error while requesting API to update Redis cluster: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"API Error Updating",
			fmt.Sprintf("Error while waiting for operation %q to update Redis cluster: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (r *RedisAPI) ListHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) []*redis.Host {
	var hosts []*redis.Host
	pageToken := ""

	for {
		resp, err := sdk.MDB().Redis().Cluster().ListHosts(ctx, &redis.ListClusterHostsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})

		if err != nil {
			diag.AddError(
				"API Error Reading",
				fmt.Sprintf("Error while requesting API to list Redis hosts %q: %s", cid, err.Error()),
			)
			return nil
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return hosts
}

func (r *RedisAPI) MoveCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, folderID string) {

	request := &redis.MoveClusterRequest{
		ClusterId:           cid,
		DestinationFolderId: folderID,
	}
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending Redis cluster move request: %+v", request)
		return sdk.MDB().Redis().Cluster().Move(ctx, request)
	})
	if err != nil {
		diag.AddError(
			"API Error Moving",
			fmt.Sprintf("Error while requesting API to move Redis cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"API Error Moving",
			fmt.Sprintf("Error while waiting for operation %q to move Redis cluster %q: %s", op.Id(), cid, err.Error()),
		)
	}
}

func (r *RedisAPI) EnableShardingRedis(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) {
	op, err := sdk.WrapOperation(sdk.MDB().Redis().Cluster().EnableSharding(ctx, &redis.EnableShardingClusterRequest{ClusterId: cid}))
	if err != nil {
		diag.AddError(
			"API Error EnableSharding",
			fmt.Sprintf("Error while requesting API to enable sharding Redis cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"API Error EnableSharding",
			fmt.Sprintf("Error while waiting for operation %q to enable sharding Redis cluster %q: %s", op.Id(), cid, err.Error()),
		)
	}
}

func (r *RedisAPI) CreateShard(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, shardName string, hostSpecs []*redis.HostSpec) {
	op, err := sdk.WrapOperation(
		sdk.MDB().Redis().Cluster().AddShard(ctx, &redis.AddClusterShardRequest{
			ClusterId: cid,
			ShardName: shardName,
			HostSpecs: hostSpecs,
		}),
	)
	if err != nil {
		diag.AddError(
			"API Error Creating",
			fmt.Sprintf("Error while requesting API to create shard Redis cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"API Error Creating",
			fmt.Sprintf("Error while waiting for operation %q to create shard Redis cluster %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}

	r.RebalanceCluster(ctx, sdk, diag, cid)
}

func (r *RedisAPI) RebalanceCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) {
	op, err := sdk.WrapOperation(
		sdk.MDB().Redis().Cluster().Rebalance(ctx, &redis.RebalanceClusterRequest{
			ClusterId: cid,
		}),
	)
	if err != nil {
		diag.AddError(
			"API Error Rebalance",
			fmt.Sprintf("Error while requesting API to create shard Redis cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"API Error Rebalance",
			fmt.Sprintf("Error while waiting for operation %q to create shard Redis cluster %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}
}

func (r *RedisAPI) CreateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*redis.HostSpec) {
	for _, spec := range specs {
		op, err := sdk.WrapOperation(
			sdk.MDB().Redis().Cluster().AddHosts(ctx, &redis.AddClusterHostsRequest{
				ClusterId: cid,
				HostSpecs: []*redis.HostSpec{spec},
			}),
		)
		if err != nil {
			diag.AddError(
				"API Error Creating",
				fmt.Sprintf("Error while requesting API to create host Redis cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diag.AddError(
				"API Error Creating",
				fmt.Sprintf("Error while waiting for operation %q to create host Redis cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}

func (r *RedisAPI) DeleteShard(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, shardName string) {
	op, err := sdk.WrapOperation(
		sdk.MDB().Redis().Cluster().DeleteShard(ctx, &redis.DeleteClusterShardRequest{
			ClusterId: cid,
			ShardName: shardName,
		}),
	)
	if err != nil {
		diag.AddError(
			"API Error Deleting",
			fmt.Sprintf("Error while requesting API to delete shard Redis cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"API Error Deleting",
			fmt.Sprintf("Error while waiting for operation %q to delete shard Redis cluster %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}
}

func (r *RedisAPI) DeleteHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, fqdns []string) {
	for _, fqdn := range fqdns {
		op, err := sdk.WrapOperation(
			sdk.MDB().Redis().Cluster().DeleteHosts(ctx, &redis.DeleteClusterHostsRequest{
				ClusterId: cid,
				HostNames: []string{fqdn},
			}),
		)
		if err != nil {
			diag.AddError(
				"API Error Creating",
				fmt.Sprintf("Error while requesting API to delete host Redis cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diag.AddError(
				"API Error Creating",
				fmt.Sprintf("Error while waiting for operation %q to delete host Redis cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}

func (r *RedisAPI) UpdateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*redis.UpdateHostSpec) {
	for _, spec := range specs {
		request := &redis.UpdateClusterHostsRequest{
			ClusterId: cid,
			UpdateHostSpecs: []*redis.UpdateHostSpec{
				spec,
			},
		}
		op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
			log.Printf("[DEBUG] Sending Redis cluster update hosts request: %+v", request)
			return sdk.MDB().Redis().Cluster().UpdateHosts(ctx, request)
		})
		if err != nil {
			diag.AddError(
				"API Error Updating",
				fmt.Sprintf("Error while requesting API to update host Redis cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diag.AddError(
				"API Error Updating",
				fmt.Sprintf("Error while waiting for operation %q to update host Redis cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}
