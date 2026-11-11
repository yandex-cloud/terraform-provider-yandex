package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func DataSourceYandexMDBKafkaTopic() *schema.Resource {
	dataSource := sdkcommon.ConvertResourceToDataSource(ResourceYandexMDBKafkaTopic())

	dataSource.Description = "Get information about a topic of the Yandex Managed Kafka cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts)."

	dataSource.Schema["cluster_id"].Computed = false
	dataSource.Schema["cluster_id"].Required = true
	dataSource.Schema["name"].Computed = false
	dataSource.Schema["name"].Required = true
	// TODO: SA1019: dataSource.Read is deprecated: Use ReadContext or ReadWithoutTimeout instead. This implementation does not support request cancellation initiated by Terraform, such as a system or practitioner sending SIGINT (Ctrl-c). This implementation also does not support warning diagnostics. (staticcheck)
	dataSource.Read = dataSourceYandexMDBKafkaTopicRead
	return dataSource
}

func dataSourceYandexMDBKafkaTopicRead(d *schema.ResourceData, meta interface{}) error {
	clusterID := d.Get("cluster_id").(string)
	topicName := d.Get("name").(string)
	topicID := sdkcommon.ConstructResourceId(clusterID, topicName)
	d.SetId(topicID)
	return resourceYandexMDBKafkaTopicRead(d, meta)
}
