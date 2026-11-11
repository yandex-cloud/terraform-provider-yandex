package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_postgresql_cluster":  ResourceYandexMDBPostgreSQLCluster(),
		"yandex_mdb_postgresql_database": ResourceYandexMDBPostgreSQLDatabase(),
		"yandex_mdb_postgresql_user":     ResourceYandexMDBPostgreSQLUser(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_postgresql_cluster":  DataSourceYandexMDBPostgreSQLCluster(),
		"yandex_mdb_postgresql_database": DataSourceYandexMDBPostgreSQLDatabase(),
		"yandex_mdb_postgresql_user":     DataSourceYandexMDBPostgreSQLUser(),
	}
}
