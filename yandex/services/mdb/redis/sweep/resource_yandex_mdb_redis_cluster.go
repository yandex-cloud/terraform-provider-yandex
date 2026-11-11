package sweep

import (
	"fmt"
	"strings"

	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	redissdk "github.com/yandex-cloud/go-sdk/services/mdb/redis/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_redis "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/redis/pluginsdk"
)

func registerResourceYandexMdbRedisCluster(add func(string, *resource.Sweeper)) {
	add(redisResourceType, &resource.Sweeper{
		Name: redisResourceType,
		F:    sweepAllMDBRedisCluster,
	})
}

const (
	redisResourceType    = "yandex_mdb_redis_cluster"
	redisResourceFoo     = redisResourceType + ".foo"
	redisResourceSharded = redisResourceType + ".bar"
)

func sweepAllMDBRedisCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := redissdk.NewClusterClient(conf.SDK).List(conf.Context(), &redis.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
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

func sweepMDBRedisCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMDBRedisClusterOnce, conf, "Redis cluster", id)
}

func sweepMDBRedisClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_redis.YandexMDBRedisClusterCreateTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := redissdk.NewClusterClient(conf.SDK).Update(ctx, &redis.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = sweep.HandleSweepOperationV2(ctx, op, err)
	if err != nil && !strings.EqualFold(sdkcommon.ErrorMessage(err), "no changes detected") {
		return err
	}

	deleteOp, err := redissdk.NewClusterClient(conf.SDK).Delete(ctx, &redis.DeleteClusterRequest{
		ClusterId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, deleteOp, err)
}
