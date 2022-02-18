package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMDBMySQLCluster_byID(t *testing.T) {
	t.Parallel()

	mysqlName := acctest.RandomWithPrefix("ds-mysql-by-id")
	mysqlDesc := "MySQL Cluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBMysqlClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMysqlClusterConfig(mysqlName, mysqlDesc, true),
				Check: testAccDataSourceMDBMysqlClusterCheck(
					"data.yandex_mdb_mysql_cluster.bar",
					"yandex_mdb_mysql_cluster.foo", mysqlName, mysqlDesc),
			},
		},
	})
}

func TestAccDataSourceMDBMySQLCluster_byName(t *testing.T) {
	t.Parallel()

	mysqlName := acctest.RandomWithPrefix("ds-mysql-by-name")
	mysqlDesc := "MySQL Cluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBMysqlClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMysqlClusterConfig(mysqlName, mysqlDesc, false),
				Check: testAccDataSourceMDBMysqlClusterCheck(
					"data.yandex_mdb_mysql_cluster.bar",
					"yandex_mdb_mysql_cluster.foo", mysqlName, mysqlDesc),
			},
		},
	})
}

func testAccDataSourceMDBMysqlClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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

		instanceAttrsToTest := []struct {
			dataSourcePath string
			resourcePath   string
		}{
			{
				"name",
				"name",
			},
			{
				"folder_id",
				"folder_id",
			},
			{
				"network_id",
				"network_id",
			},
			{
				"created_at",
				"created_at",
			},
			{
				"description",
				"description",
			},
			{
				"labels.%",
				"labels.%",
			},
			{
				"labels.test_key",
				"labels.test_key",
			},
			{
				"environment",
				"environment",
			},
			{
				"resources.0.disk_size",
				"resources.0.disk_size",
			},
			{
				"resources.0.disk_type_id",
				"resources.0.disk_type_id",
			},
			{
				"resources.0.resource_preset_id",
				"resources.0.resource_preset_id",
			},
			{
				"version",
				"version",
			},
			{
				"user.#",
				"user.#",
			},
			{
				"user.0.name",
				"user.0.name",
			},
			{
				"user.0.permission.#",
				"user.0.permission.#",
			},
			{
				"user.0.permission.0.database_name",
				"user.0.permission.0.database_name",
			},
			{
				"user.0.permission.0.roles.#",
				"user.0.permission.0.roles.#",
			},
			{
				"user.0.permission.0.roles.0",
				"user.0.permission.0.roles.0",
			},
			{
				"user.0.permission.0.roles.1",
				"user.0.permission.0.roles.1",
			},
			{
				"database.#",
				"database.#",
			},
			{
				"database.#",
				"database.#",
			},
			{
				"database.0.name",
				"database.0.name",
			},
			{
				"host.#",
				"host.#",
			},
			{
				"host.0.fqdn",
				"host.0.fqdn",
			},
			{
				"host.0.assign_public_ip",
				"host.0.assign_public_ip",
			},
			{
				"host.0.subnet_id",
				"host.0.subnet_id",
			},
			{
				"host.0.zone",
				"host.0.zone",
			},
			{
				"host.0.replication_source",
				"host.0.replication_source",
			},
			{
				"host.0.priority",
				"host.0.priority",
			},
			{
				"host.0.backup_priority",
				"host.0.backup_priority",
			},
			{
				"security_group_ids.#",
				"security_group_ids.#",
			},
			{
				"deletion_protection",
				"deletion_protection",
			},
		}

		for _, attrToCheck := range instanceAttrsToTest {
			if _, ok := datasourceAttributes[attrToCheck.dataSourcePath]; !ok {
				return fmt.Errorf("%s is not present in data source attributes", attrToCheck.dataSourcePath)
			}
			if _, ok := resourceAttributes[attrToCheck.resourcePath]; !ok {
				return fmt.Errorf("%s is not present in resource attributes", attrToCheck.resourcePath)
			}
			if datasourceAttributes[attrToCheck.dataSourcePath] != resourceAttributes[attrToCheck.resourcePath] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrToCheck.dataSourcePath,
					datasourceAttributes[attrToCheck.dataSourcePath],
					resourceAttributes[attrToCheck.resourcePath],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceMDBMysqlClusterCheck(datasourceName string, resourceName string, mysqlName string, desc string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBMysqlClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", mysqlName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
		resource.TestCheckResourceAttr(datasourceName, "user.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "database.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "host.#", "1"),
		resource.TestCheckResourceAttrSet(datasourceName, "host.0.fqdn"),
		testAccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
	)
}

const mdbMysqlClusterByIDConfig = `
data "yandex_mdb_mysql_cluster" "bar" {
  cluster_id = "${yandex_mdb_mysql_cluster.foo.id}"
}
`

const mdbMysqlClusterByNameConfig = `
data "yandex_mdb_mysql_cluster" "bar" {
  name = "${yandex_mdb_mysql_cluster.foo.name}"
}
`

func testAccDataSourceMDBMysqlClusterConfig(mysqlName, mysqlDesc string, useDataID bool) string {
	if useDataID {
		return testAccMDBMySQLClusterConfigMain(mysqlName, mysqlDesc, "PRESTABLE", false) + mdbMysqlClusterByIDConfig
	}

	return testAccMDBMySQLClusterConfigMain(mysqlName, mysqlDesc, "PRESTABLE", false) + mdbMysqlClusterByNameConfig
}
