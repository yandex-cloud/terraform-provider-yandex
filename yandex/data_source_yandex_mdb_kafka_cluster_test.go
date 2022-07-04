package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const mdbKafkaClusterByIDConfig = `
data "yandex_mdb_kafka_cluster" "bar" {
	cluster_id = yandex_mdb_kafka_cluster.foo.id
}
`

const mdbKafkaClusterByNameConfig = `
data "yandex_mdb_kafka_cluster" "bar" {
	name = yandex_mdb_kafka_cluster.foo.name
}
`

const mdbKafkaTopicDataSourceConfig = `
data "yandex_mdb_kafka_topic" "baz" {
	cluster_id = yandex_mdb_kafka_cluster.foo.id
	name = "raw_events"
}
`

func TestAccDataSourceMDBKafkaClusterAndTopic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("ds-kf-by-id")
	description := "KafkaCluster Terraform Datasource Test"
	resourceName := "yandex_mdb_kafka_cluster.foo"
	clusterDatasource := "data.yandex_mdb_kafka_cluster.bar"
	topicDatasource := "data.yandex_mdb_kafka_topic.baz"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBKafkaClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBKafkaClusterConfig(clusterName, description, true),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceMDBKafkaClusterAttributesCheck(clusterDatasource, resourceName),
					testAccCheckResourceIDField(clusterDatasource, "cluster_id"),
					resource.TestCheckResourceAttr(clusterDatasource, "name", clusterName),
					resource.TestCheckResourceAttr(clusterDatasource, "folder_id", getExampleFolderID()),
					resource.TestCheckResourceAttr(clusterDatasource, "description", description),
					resource.TestCheckResourceAttr(clusterDatasource, "environment", "PRESTABLE"),
					resource.TestCheckResourceAttr(clusterDatasource, "labels.test_key", "test_value"),
					resource.TestCheckResourceAttr(clusterDatasource, "config.0.brokers_count", "1"),
					resource.TestCheckResourceAttr(clusterDatasource, "config.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(clusterDatasource, "config.0.version", "2.8"),
					resource.TestCheckResourceAttr(clusterDatasource, "zookeeper.#", "0"),
					resource.TestCheckResourceAttr(clusterDatasource, "topic.#", "2"),
					resource.TestCheckResourceAttr(clusterDatasource, "user.#", "2"),
					resource.TestCheckResourceAttr(clusterDatasource, "deletion_protection", "false"),
					testAccCheckCreatedAtAttr(clusterDatasource),

					resource.TestCheckResourceAttr(topicDatasource, "partitions", "1"),
					resource.TestCheckResourceAttr(topicDatasource, "replication_factor", "1"),
					resource.TestCheckResourceAttr(topicDatasource, "topic_config.0.cleanup_policy", "CLEANUP_POLICY_COMPACT_AND_DELETE"),
					resource.TestCheckResourceAttr(topicDatasource, "topic_config.0.max_message_bytes", "777216"),
					resource.TestCheckResourceAttr(topicDatasource, "topic_config.0.segment_bytes", "134217728"),
					resource.TestCheckResourceAttr(topicDatasource, "topic_config.0.flush_ms", "9223372036854775807"),
				),
			},
		},
	})
}

func testAccDataSourceMDBKafkaClusterConfig(kfName, kfDesc string, useDataID bool) string {
	wholeConfig := testAccMDBKafkaClusterConfigMain(kfName, kfDesc, "PRESTABLE")

	if useDataID {
		wholeConfig += mdbKafkaClusterByIDConfig
	} else {
		wholeConfig += mdbKafkaClusterByNameConfig
	}

	return wholeConfig + mdbKafkaTopicDataSourceConfig
}

func testAccDataSourceMDBKafkaClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		if ds.Primary.ID != rs.Primary.ID {
			return fmt.Errorf("cluster `data source` ID does not match `resource` ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
		}

		datasourceAttributes := ds.Primary.Attributes
		resourceAttributes := rs.Primary.Attributes

		instanceAttrsToTest := []string{
			"name",
			"folder_id",
			"network_id",
			"created_at",
			"description",
			"labels",
			"environment",
			"config",
			"config.0.kafka",
			"config.0.zookeeper",
			"config.0.assign_public_ip",
			"topics",
			"users",
			"security_group_ids",
			"deletion_protection",
		}

		for _, attrToCheck := range instanceAttrsToTest {
			if datasourceAttributes[attrToCheck] != resourceAttributes[attrToCheck] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrToCheck,
					datasourceAttributes[attrToCheck],
					resourceAttributes[attrToCheck],
				)
			}
		}

		return nil
	}
}
