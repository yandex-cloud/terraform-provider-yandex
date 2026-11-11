package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_mysql_cluster":  ResourceYandexMDBMySQLCluster(),
		"yandex_mdb_mysql_database": ResourceYandexMDBMySQLDatabase(),
		"yandex_mdb_mysql_user":     ResourceYandexMDBMySQLUser(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_mysql_cluster":  DataSourceYandexMDBMySQLCluster(),
		"yandex_mdb_mysql_database": DataSourceYandexMDBMySQLDatabase(),
		"yandex_mdb_mysql_user":     DataSourceYandexMDBMySQLUser(),
	}
}
