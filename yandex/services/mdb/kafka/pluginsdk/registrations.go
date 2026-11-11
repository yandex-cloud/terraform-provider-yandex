package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_kafka_cluster":   ResourceYandexMDBKafkaCluster(),
		"yandex_mdb_kafka_topic":     ResourceYandexMDBKafkaTopic(),
		"yandex_mdb_kafka_connector": ResourceYandexMDBKafkaConnector(),
		"yandex_mdb_kafka_user":      ResourceYandexMDBKafkaUser(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_mdb_kafka_cluster":   DataSourceYandexMDBKafkaCluster(),
		"yandex_mdb_kafka_topic":     DataSourceYandexMDBKafkaTopic(),
		"yandex_mdb_kafka_connector": DataSourceYandexMDBKafkaConnector(),
		"yandex_mdb_kafka_user":      DataSourceYandexMDBKafkaUser(),
	}
}
