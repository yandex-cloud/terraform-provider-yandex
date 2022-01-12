package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMDBElasticsearchCluster_byID(t *testing.T) {
	t.Parallel()

	esName := acctest.RandomWithPrefix("ds-es-by-id")
	esDesc := "ElasticsearchCluster Terraform Datasource Test"
	randInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBElasticsearchClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBElasticsearchClusterConfig(esName, esDesc, randInt, true),
				Check: testAccDataSourceMDBElasticsearchClusterCheck(
					"data.yandex_mdb_elasticsearch_cluster.bar",
					"yandex_mdb_elasticsearch_cluster.foo", esName, esDesc),
			},
		},
	})
}

func TestAccDataSourceMDBElasticsearchCluster_byName(t *testing.T) {
	t.Parallel()

	esName := acctest.RandomWithPrefix("ds-es-by-name")
	esDesc := "ElasticseachCluster Terraform Datasource Test"
	randInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBElasticsearchClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBElasticsearchClusterConfig(esName, esDesc, randInt, false),
				Check: testAccDataSourceMDBElasticsearchClusterCheck(
					"data.yandex_mdb_elasticsearch_cluster.bar",
					"yandex_mdb_elasticsearch_cluster.foo", esName, esDesc),
			},
		},
	})
}

func testAccDataSourceMDBElasticseachClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
			"host",
			"config",
			"security_group_ids",
			"service_account_id",
			"deletion_protection",
			"maintenance_window.0.type",
			"maintenance_window.0.day",
			"maintenance_window.0.hour",
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

func testAccDataSourceMDBElasticsearchClusterCheck(datasourceName string, resourceName string, name string, desc string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBElasticseachClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", name),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "description", desc),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
		resource.TestCheckResourceAttr(datasourceName, "config.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "host.#", "5"),
		resource.TestCheckResourceAttrSet(datasourceName, "service_account_id"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
		// our host stored in set and indexed by hashcode, not order.
		// resource.TestCheckResourceAttrSet(datasourceName, "host.0.fqdn"),
		testAccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.type", "WEEKLY"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.day", "FRI"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.hour", "20"),
	)
}

const mdbElasticsearchClusterByIDConfig = `
data "yandex_mdb_elasticsearch_cluster" "bar" {
  cluster_id = "${yandex_mdb_elasticsearch_cluster.foo.id}"
}
`

const mdbElasticsearchClusterByNameConfig = `
data "yandex_mdb_elasticsearch_cluster" "bar" {
  name = "${yandex_mdb_elasticsearch_cluster.foo.name}"
}
`

func testAccDataSourceMDBElasticsearchClusterConfig(name, desc string, randInt int, useDataID bool) string {
	if useDataID {
		return testAccMDBElasticsearchClusterConfig(name, desc, "PRESTABLE", false, randInt) + mdbElasticsearchClusterByIDConfig
	}

	return testAccMDBElasticsearchClusterConfig(name, desc, "PRESTABLE", false, randInt) + mdbElasticsearchClusterByNameConfig
}
