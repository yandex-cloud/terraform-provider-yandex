package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceMDBMongoDBCluster_byName(t *testing.T) {
	t.Parallel()

	mongodbName := acctest.RandomWithPrefix("ds-mongodb-by-name")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBMongoDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMongoDBClusterConfig(mongodbName),
				Check: testAccDataSourceMDBMongoDBClusterCheck(
					"data.yandex_mdb_mongodb_cluster.bar",
					"yandex_mdb_mongodb_cluster.foo", mongodbName),
			},
		},
	})
}

func testAccDataSourceMDBMongoDBClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
			"database",
			"user.0.name",
			"user.0.permission",
			"user.0.database",
			"host",
			"sharded",
			"cluster_config.0.version",
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

func testAccDataSourceMDBMongoDBClusterCheck(datasourceName string, resourceName string, mongodbName string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBMongoDBClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", mongodbName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
		resource.TestCheckResourceAttr(datasourceName, "sharded", "false"),
		resource.TestCheckResourceAttr(datasourceName, "host.#", "2"),
		testAccCheckCreatedAtAttr(datasourceName),
	)
}

const mdbMongoDBClusterByNameConfig = `
data "yandex_mdb_mongodb_cluster" "bar" {
  name = "${yandex_mdb_mongodb_cluster.foo.name}"
}
`

func testAccDataSourceMDBMongoDBClusterConfig(mongodbName string) string {
	return testAccMDBMongoDBClusterConfigMain(mongodbName) + mdbMongoDBClusterByNameConfig
}
