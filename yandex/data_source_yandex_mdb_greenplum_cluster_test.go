package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceMDBGreenplumCluster_byID(t *testing.T) {
	t.Parallel()

	greenplumName := acctest.RandomWithPrefix("ds-greenplum-by-id")
	greenplumDesc := "Greenplum Cluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBGreenplumClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBGreenplumClusterConfig(greenplumName, greenplumDesc, true),
				Check: testAccDataSourceMDBGreenplumClusterCheck(
					"data.yandex_mdb_greenplum_cluster.bar",
					"yandex_mdb_greenplum_cluster.foo", greenplumName, greenplumDesc),
			},
		},
	})
}

func TestAccDataSourceMDBGreenplumCluster_byName(t *testing.T) {
	t.Parallel()

	greenplumName := acctest.RandomWithPrefix("ds-greenplum-by-name")
	greenplumDesc := "Greenplum Cluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBGreenplumClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBGreenplumClusterConfig(greenplumName, greenplumDesc, false),
				Check: testAccDataSourceMDBGreenplumClusterCheck(
					"data.yandex_mdb_greenplum_cluster.bar",
					"yandex_mdb_greenplum_cluster.foo", greenplumName, greenplumDesc),
			},
		},
	})
}

func testAccDataSourceMDBGreenplumClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
				"master_subcluster.0.resources.0.disk_size",
				"master_subcluster.0.resources.0.disk_size",
			},
			{
				"master_subcluster.0.resources.0.disk_type_id",
				"master_subcluster.0.resources.0.disk_type_id",
			},
			{
				"master_subcluster.0.resources.0.resource_preset_id",
				"master_subcluster.0.resources.0.resource_preset_id",
			},
			{
				"version",
				"version",
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

func testAccDataSourceMDBGreenplumClusterCheck(datasourceName string, resourceName string, greenplumName string, desc string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBGreenplumClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", greenplumName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key_create", "test_value_create"),
		testAccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
	)
}

const mdbGreenplumClusterByIDConfig = `
data "yandex_mdb_greenplum_cluster" "bar" {
  cluster_id = "${yandex_mdb_greenplum_cluster.foo.id}"
}
`

const mdbGreenplumClusterByNameConfig = `
data "yandex_mdb_greenplum_cluster" "bar" {
  name = "${yandex_mdb_greenplum_cluster.foo.name}"
}
`

func testAccDataSourceMDBGreenplumClusterConfig(greenplumName, greenplumDesc string, useDataID bool) string {
	if useDataID {
		return testAccMDBGreenplumClusterConfigMain(greenplumName, greenplumDesc, "PRESTABLE", false) + mdbGreenplumClusterByIDConfig
	}

	return testAccMDBGreenplumClusterConfigMain(greenplumName, greenplumDesc, "PRESTABLE", false) + mdbGreenplumClusterByNameConfig
}
