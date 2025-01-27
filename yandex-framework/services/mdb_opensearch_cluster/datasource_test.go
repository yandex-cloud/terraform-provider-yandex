package mdb_opensearch_cluster_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceMDBOpenSearchCluster_byID(t *testing.T) {
	t.Parallel()

	osName := acctest.RandomWithPrefix("ds-os-by-id")
	osDesc := "OpenSearchCluster Terraform Datasource Test"
	randInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBOpenSearchClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBOpenSearchClusterConfig(osName, osDesc, randInt, true),
				Check:  testAccDataSourceMDBOpenSearchClusterCheck("data.yandex_mdb_opensearch_cluster.bar", osName, osDesc),
			},
		},
	})
}

func TestAccDataSourceMDBOpenSearchCluster_byName(t *testing.T) {
	t.Parallel()

	osName := acctest.RandomWithPrefix("ds-os-by-name")
	osDesc := "OpenSearchCluster Terraform Datasource Test"
	randInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBOpenSearchClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBOpenSearchClusterConfig(osName, osDesc, randInt, false),
				Check: testAccDataSourceMDBOpenSearchClusterCheck(
					"data.yandex_mdb_opensearch_cluster.bar", osName, osDesc),
			},
		},
	})
}

func testAccDataSourceMDBOpenSearchClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
			"hosts",
			"config",
			"security_group_ids",
			"service_account_id",
			"deletion_protection",
			"maintenance_window.type",
			"maintenance_window.day",
			"maintenance_window.hour",
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

func testAccDataSourceMDBOpenSearchClusterCheck(datasourceName string, name string, desc string) resource.TestCheckFunc {
	folderID := test.GetExampleFolderID()
	env := "PRESTABLE"
	resourceName := openSearchResourcePrefix + name

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBOpenSearchClusterAttributesCheck(datasourceName, resourceName),
		test.AccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", name),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
		// resource.TestCheckResourceAttr(datasourceName, "config", "1"),
		resource.TestCheckResourceAttrSet(datasourceName, "service_account_id"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
		resource.TestCheckResourceAttr(datasourceName, "hosts.#", "2"),
		resource.TestCheckResourceAttrSet(datasourceName, "hosts.0.fqdn"),
		resource.TestCheckResourceAttrSet(datasourceName, "hosts.1.fqdn"),
		test.AccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.type", "WEEKLY"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.day", "FRI"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.hour", "20"),
	)
}

const mdbOpenSearchClusterByIDConfig = `
data "yandex_mdb_opensearch_cluster" "bar" {
  cluster_id = "${yandex_mdb_opensearch_cluster.%s.id}"
}
`

const mdbOpenSearchClusterByNameConfig = `
data "yandex_mdb_opensearch_cluster" "bar" {
  name = "${yandex_mdb_opensearch_cluster.%s.name}"
}
`

func testAccDataSourceMDBOpenSearchClusterConfig(name, desc string, randInt int, useDataID bool) string {
	if useDataID {
		return testAccMDBOpenSearchClusterConfig(name, desc, "PRESTABLE", false, randInt) +
			fmt.Sprintf(mdbOpenSearchClusterByIDConfig, name)
	}

	return testAccMDBOpenSearchClusterConfig(name, desc, "PRESTABLE", false, randInt) +
		fmt.Sprintf(mdbOpenSearchClusterByNameConfig, name)
}
