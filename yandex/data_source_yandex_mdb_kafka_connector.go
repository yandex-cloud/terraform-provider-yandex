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
	// TODO: SA1019: dataSource.Read is deprecated: Use ReadContext or ReadWithoutTimeout instead. This implementation does not support request cancellation initiated by Terraform, such as a system or practitioner sending SIGINT (Ctrl-c). This implementation also does not support warning diagnostics. (staticcheck)
	dataSource.Read = resourceYandexMDBKafkaConnectorRead
	return dataSource
}
