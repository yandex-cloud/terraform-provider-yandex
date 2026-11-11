package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func registerMdbRedisClusterV2Resource(add func(string, *resource.Sweeper)) {
	add("yandex_mdb_redis_cluster_v2", &resource.Sweeper{
		Name: "yandex_mdb_redis_cluster_v2",
		F:    sweepAllMDBRedisCluster,
	})
}
