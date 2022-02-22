package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMDBClickHouseCluster_byID(t *testing.T) {
	t.Parallel()

	chName := acctest.RandomWithPrefix("ds-ch-by-id")
	chDesc := "ClickHouseCluster Terraform Datasource Test"
	bucketName := acctest.RandomWithPrefix("tf-test-clickhouse-bucket")
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBClickHouseClusterConfig(chName, chDesc, bucketName, true, rInt),
				Check: testAccDataSourceMDBClickHouseClusterCheck(
					"data.yandex_mdb_clickhouse_cluster.bar",
					"yandex_mdb_clickhouse_cluster.foo", chName, chDesc),
			},
		},
	})
}

func TestAccDataSourceMDBClickHouseCluster_byName(t *testing.T) {
	t.Parallel()

	chName := acctest.RandomWithPrefix("ds-ch-by-name")
	chDesc := "ClickHouseCluster Terraform Datasource Test"
	bucketName := acctest.RandomWithPrefix("tf-test-clickhouse-bucket")
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBClickHouseClusterConfig(chName, chDesc, bucketName, false, rInt),
				Check: testAccDataSourceMDBClickHouseClusterCheck(
					"data.yandex_mdb_clickhouse_cluster.bar",
					"yandex_mdb_clickhouse_cluster.foo", chName, chDesc),
			},
		},
	})
}

func testAccDataSourceMDBClickHouseClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
			return fmt.Errorf("instance `data source` ID does not match `resource` ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
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
			"resources",
			"host",
			"clickhouse",
			"database",
			"user.0.name",
			"user.0.permission",
			"database",
			"access",
			"backup_window_start",
			"security_group_ids",
			"sql_user_management",
			"sql_database_management",
			"embedded_keeper",
			"service_account_id",
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

func testAccDataSourceMDBClickHouseClusterCheck(datasourceName string, resourceName string, chName string, desc string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBClickHouseClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", chName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
		resource.TestCheckResourceAttr(datasourceName, "user.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "database.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "clickhouse.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "host.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "access.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "backup_window_start.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "sql_user_management", "false"),
		resource.TestCheckResourceAttr(datasourceName, "sql_database_management", "false"),
		resource.TestCheckResourceAttr(datasourceName, "embedded_keeper", "false"),
		resource.TestCheckResourceAttrSet(datasourceName, "service_account_id"),
		resource.TestCheckResourceAttrSet(datasourceName, "host.0.fqdn"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
		testAccCheckCreatedAtAttr(datasourceName),
	)
}

const mdbClickHouseClusterByIDConfig = `
data "yandex_mdb_clickhouse_cluster" "bar" {
  cluster_id = "${yandex_mdb_clickhouse_cluster.foo.id}"
}
`

const mdbClickHouseClusterByNameConfig = `
data "yandex_mdb_clickhouse_cluster" "bar" {
  name = "${yandex_mdb_clickhouse_cluster.foo.name}"
}
`

func testAccDataSourceMDBClickHouseClusterConfig(chName, chDesc, bucket string, useDataID bool, randInt int) string {
	if useDataID {
		return testAccMDBClickHouseClusterConfigMain(chName, chDesc, "PRESTABLE", false, bucket, randInt) + mdbClickHouseClusterByIDConfig
	}

	return testAccMDBClickHouseClusterConfigMain(chName, chDesc, "PRESTABLE", false, bucket, randInt) + mdbClickHouseClusterByNameConfig
}
