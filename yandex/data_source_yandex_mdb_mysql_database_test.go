package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
)

func TestAccDataSourceMDBMySQLDatabase_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("ds-pg-by-id")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBMySQLDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMySQLDatabaseConfig(clusterName, true),
				Check: testAccDataSourceMDBMySQLDatabaseCheck(
					"data.yandex_mdb_mysql_database.bar",
					"yandex_mdb_mysql_database.foo", clusterName),
			},
		},
	})
}

func testAccDataSourceMDBMySQLDatabaseAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
				"cluster_id",
				"cluster_id",
			},
			{
				"name",
				"name",
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

func testAccDataSourceMDBMySQLDatabaseCheck(datasourceName string, resourceName string, clusterName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBMySQLDatabaseAttributesCheck(datasourceName, resourceName),
		testAccDataSourceMDBMySQLDatabaseCheckResourceIDField(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "name", "foo"),
	)
}

func testAccDataSourceMDBMySQLDatabaseCheckResourceIDField(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		expectedResourceId := constructResourceId(rs.Primary.Attributes["cluster_id"], rs.Primary.Attributes["name"])

		if expectedResourceId != rs.Primary.ID {
			return fmt.Errorf("Wrong resource %s id. Expected %s, got %s", resourceName, expectedResourceId, rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataSourceMDBMySQLDatabaseConfig(name string, useDataID bool) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
	name        = "%s"
	description = "MySQL Database Terraform Datasource Test"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id
	version     = "8.0"
	
	resources {
		resource_preset_id = "s2.micro"
		disk_type_id       = "network-ssd"
		disk_size          = 24
	}

	host {
		zone      = "ru-central1-c"
		subnet_id = yandex_vpc_subnet.foo_c.id
	}
}

resource "yandex_mdb_mysql_database" "foo" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
	name       = "foo"
}

data "yandex_mdb_mysql_database" "bar" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
	name       = yandex_mdb_mysql_database.foo.name
}
`, name)
}

func testAccCheckMDBMySQLDatabaseDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_mysql_database" {
			continue
		}

		clusterId, dbname, err := deconstructResourceId(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = config.sdk.MDB().MySQL().Database().Get(context.Background(), &mysql.GetDatabaseRequest{
			ClusterId:    clusterId,
			DatabaseName: dbname,
		})

		if err == nil {
			return fmt.Errorf("MySQL Database still exists")
		}
	}

	return nil
}
