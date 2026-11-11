package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_ydb_database_dedicated":  ResourceYandexYDBDatabaseDedicated(),
		"yandex_ydb_database_serverless": ResourceYandexYDBDatabaseServerless(),
		"yandex_ydb_topic":               ResourceYandexYDBTopic(),
		"yandex_ydb_table":               ResourceYandexYDBTable(),
		"yandex_ydb_table_changefeed":    ResourceYandexYDBTableChangefeed(),
		"yandex_ydb_table_index":         ResourceYandexYDBTableIndex(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_ydb_database_dedicated":  DataSourceYandexYDBDatabaseDedicated(),
		"yandex_ydb_database_serverless": DataSourceYandexYDBDatabaseServerless(),
	}
}
