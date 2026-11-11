package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_greenplum_cluster": ResourceYandexMDBGreenplumCluster(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_greenplum_cluster": DataSourceYandexMDBGreenplumCluster(),
	}
}
