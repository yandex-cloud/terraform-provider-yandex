package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexMDBKafkaConnector() *schema.Resource {
	dataSource := convertResourceToDataSource(resourceYandexMDBKafkaConnector())
	dataSource.Schema["cluster_id"].Computed = false
	dataSource.Schema["cluster_id"].Required = true
	dataSource.Schema["name"].Computed = false
	dataSource.Schema["name"].Required = true
	dataSource.Read = resourceYandexMDBKafkaConnectorRead
	return dataSource
}
