package mdb_redis_cluster_v2_test

import (
	"context"
	"fmt"

	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"

	"github.com/hashicorp/go-multierror"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/protobuf/field_mask"

	"strings"
	"time"
)

func testSweepMDBRedisCluster(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	resp, err := conf.SDK.MDB().Redis().Cluster().List(ctx, &redis.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
		PageSize: 1000,
	})
	if err != nil {
		return fmt.Errorf("error getting Redis clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBRedisCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Redis cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBRedisCluster(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepMDBRedisClusterOnce, conf, "Redis cluster", id)
}

func sweepMDBRedisClusterOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.SDK.MDB().Redis().Cluster().Update(ctx, &redis.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = test.HandleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(err.Error(), "no changes detected") {
		return err
	}

	op, err = conf.SDK.MDB().Redis().Cluster().Delete(ctx, &redis.DeleteClusterRequest{
		ClusterId: id,
	})
	return test.HandleSweepOperation(ctx, conf, op, err)
}
