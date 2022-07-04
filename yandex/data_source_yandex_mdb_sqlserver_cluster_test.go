package yandex

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMDBSQLServerCluster_byID(t *testing.T) {
	if os.Getenv("TF_SQL_LICENSE_ACCEPTED") != "1" {
		t.Skip()
	}
	t.Parallel()

	sqlserverName := acctest.RandomWithPrefix("ds-sqlserver-by-id")
	sqlserverDesc := "SQLServer Cluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBSQLServerClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBSQLServerClusterConfig(sqlserverName, sqlserverDesc, true),
				Check: testAccDataSourceMDBSQLServerClusterCheck(
					"data.yandex_mdb_sqlserver_cluster.bar",
					"yandex_mdb_sqlserver_cluster.foo", sqlserverName, sqlserverDesc),
			},
		},
	})
}

func TestAccDataSourceMDBSQLServerCluster_byName(t *testing.T) {
	if os.Getenv("TF_SQL_LICENSE_ACCEPTED") != "1" {
		t.Skip()
	}
	t.Parallel()

	sqlserverName := acctest.RandomWithPrefix("ds-sqlserver-by-name")
	sqlserverDesc := "SQLServer Cluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBSQLServerClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBSQLServerClusterConfig(sqlserverName, sqlserverDesc, false),
				Check: testAccDataSourceMDBSQLServerClusterCheck(
					"data.yandex_mdb_sqlserver_cluster.bar",
					"yandex_mdb_sqlserver_cluster.foo", sqlserverName, sqlserverDesc),
			},
		},
	})
}

func testAccDataSourceMDBSQLServerClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
				"labels.test_key_create",
				"labels.test_key_create",
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
				"security_group_ids.#",
				"security_group_ids.#",
			},
			{
				"deletion_protection",
				"deletion_protection",
			},
			{
				"sqlcollation",
				"sqlcollation",
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

func testAccDataSourceMDBSQLServerClusterCheck(datasourceName string, resourceName string, sqlserverName string, desc string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBSQLServerClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", sqlserverName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key_create", "test_value_create"),
		resource.TestCheckResourceAttr(datasourceName, "user.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "database.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "host.#", "1"),
		resource.TestCheckResourceAttrSet(datasourceName, "host.0.fqdn"),
		testAccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
		resource.TestCheckResourceAttr(datasourceName, "sqlcollation", "Cyrillic_General_CI_AI"),
	)
}

const mdbSQLServerClusterByIDConfig = `
data "yandex_mdb_sqlserver_cluster" "bar" {
  cluster_id = "${yandex_mdb_sqlserver_cluster.foo.id}"
}
`

const mdbSQLServerClusterByNameConfig = `
data "yandex_mdb_sqlserver_cluster" "bar" {
  name = "${yandex_mdb_sqlserver_cluster.foo.name}"
}
`

func testAccDataSourceMDBSQLServerClusterConfig(sqlserverName, sqlserverDesc string, useDataID bool) string {
	if useDataID {
		return testAccMDBSQLServerClusterConfigMain(sqlserverName, sqlserverDesc, "PRESTABLE", false) + mdbSQLServerClusterByIDConfig
	}

	return testAccMDBSQLServerClusterConfigMain(sqlserverName, sqlserverDesc, "PRESTABLE", false) + mdbSQLServerClusterByNameConfig
}
