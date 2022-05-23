package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMDBPostgreSQLCluster_byID(t *testing.T) {
	t.Parallel()

	pgName := acctest.RandomWithPrefix("ds-pg-by-id")
	pgDesc := "PostgreSQL Cluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBPGClusterConfig(pgName, pgDesc, true),
				Check: testAccDataSourceMDBPGClusterCheck(
					"data.yandex_mdb_postgresql_cluster.bar",
					"yandex_mdb_postgresql_cluster.foo", pgName, pgDesc),
			},
		},
	})
}

func TestAccDataSourceMDBPostgreSQLCluster_byName(t *testing.T) {
	t.Parallel()

	pgName := acctest.RandomWithPrefix("ds-pg-by-name")
	pgDesc := "PostgreSQL Cluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBPGClusterConfig(pgName, pgDesc, false),
				Check: testAccDataSourceMDBPGClusterCheck(
					"data.yandex_mdb_postgresql_cluster.bar",
					"yandex_mdb_postgresql_cluster.foo", pgName, pgDesc),
			},
		},
	})
}

func testAccDataSourceMDBPGClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
				"config.#",
				"config.#",
			},
			{
				"config.0.access.#",
				"config.0.access.#",
			},
			{
				"config.0.access.0.data_lens",
				"config.0.access.0.data_lens",
			},
			{
				"config.0.autofailover",
				"config.0.autofailover",
			},
			{
				"config.0.backup_window_start.#",
				"config.0.backup_window_start.#",
			},
			{
				"config.0.backup_window_start.0.hours",
				"config.0.backup_window_start.0.hours",
			},
			{
				"config.0.backup_window_start.0.minutes",
				"config.0.backup_window_start.0.minutes",
			},
			{
				"config.0.pooler_config.#",
				"config.0.pooler_config.#",
			},
			{
				"config.0.resources.#",
				"config.0.resources.#",
			},
			{
				"config.0.resources.0.disk_size",
				"config.0.resources.0.disk_size",
			},
			{
				"config.0.resources.0.disk_type_id",
				"config.0.resources.0.disk_type_id",
			},
			{
				"config.0.resources.0.resource_preset_id",
				"config.0.resources.0.resource_preset_id",
			},
			{
				"config.0.version",
				"config.0.version",
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
				"environment",
				"environment",
			},
			{
				"folder_id",
				"folder_id",
			},
			{
				"host.#",
				"host.#",
			},
			{
				"host.0.assign_public_ip",
				"host.0.assign_public_ip",
			},
			{
				"host.0.fqdn",
				"host.0.fqdn",
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
				"host.0.role",
				"host.0.role",
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
				"labels.%",
				"labels.%",
			},
			{
				"labels.test_key",
				"labels.test_key",
			},
			{
				"name",
				"name",
			},
			{
				"network_id",
				"network_id",
			},
			{
				"status",
				"status",
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

func testAccDataSourceMDBPGClusterCheck(datasourceName string, resourceName string, pgName string, desc string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBPGClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", pgName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
		resource.TestCheckResourceAttr(datasourceName, "config.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "host.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "config.0.access.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "config.0.backup_window_start.#", "1"),
		resource.TestCheckResourceAttrSet(datasourceName, "host.0.fqdn"),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
		testAccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
	)
}

const mdbPGClusterByIDConfig = `
data "yandex_mdb_postgresql_cluster" "bar" {
  cluster_id = yandex_mdb_postgresql_cluster.foo.id
}
`

const mdbPGClusterByNameConfig = `
data "yandex_mdb_postgresql_cluster" "bar" {
  name = yandex_mdb_postgresql_cluster.foo.name
}
`

func testAccDataSourceMDBPGClusterConfig(pgName, pgDesc string, useDataID bool) string {
	if useDataID {
		return testAccMDBPGClusterConfigMain(pgName, pgDesc, "PRESTABLE", false) + mdbPGClusterByIDConfig
	}

	return testAccMDBPGClusterConfigMain(pgName, pgDesc, "PRESTABLE", false) + mdbPGClusterByNameConfig
}
