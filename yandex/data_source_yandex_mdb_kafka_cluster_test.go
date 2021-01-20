package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

func TestAccDataSourceMDBKafkaCluster_byID(t *testing.T) {
	t.Parallel()

	kfName := acctest.RandomWithPrefix("ds-kf-by-id")
	kfDesc := "KafkaCluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBKafkaClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBKafkaClusterConfig(kfName, kfDesc, true),
				Check: testAccDataSourceMDBKafkaClusterCheck(
					"data.yandex_mdb_kafka_cluster.bar",
					"yandex_mdb_kafka_cluster.foo", kfName, kfDesc),
			},
		},
	})
}

func testAccDataSourceMDBKafkaClusterConfig(kfName, kfDesc string, useDataID bool) string {
	if useDataID {
		return testAccMDBKafkaClusterConfigMain(kfName, kfDesc) + mdbKafkaClusterByIDConfig
	}

	return testAccMDBKafkaClusterConfigMain(kfName, kfDesc) + mdbKafkaClusterByNameConfig
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

func testAccDataSourceMDBKafkaClusterCheck(datasourceName string, resourceName string, chName string, desc string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBKafkaClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", chName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
		resource.TestCheckResourceAttr(datasourceName, "config.0.brokers_count", "1"),
		resource.TestCheckResourceAttr(datasourceName, "config.0.assign_public_ip", "false"),
		resource.TestCheckResourceAttr(datasourceName, "config.0.version", "2.6"),
		resource.TestCheckResourceAttr(datasourceName, "zookeeper.#", "0"),
		resource.TestCheckResourceAttr(datasourceName, "topic.#", "2"),
		resource.TestCheckResourceAttr(datasourceName, "user.#", "2"),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		testAccCheckCreatedAtAttr(datasourceName),
	)
}
