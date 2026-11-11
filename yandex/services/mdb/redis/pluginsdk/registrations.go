package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_redis_cluster": ResourceYandexMDBRedisCluster(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_redis_cluster": DataSourceYandexMDBRedisCluster(),
	}
}
