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

func TestAccDataSourceMDBMySQLUser_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("ds-pg-by-id")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBMySQLUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMySQLUserConfig(clusterName, true),
				Check: testAccDataSourceMDBMySQLUserCheck(
					"data.yandex_mdb_mysql_user.john",
					"yandex_mdb_mysql_user.john", clusterName),
			},
		},
	})
}

func testAccDataSourceMDBMySQLUserAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
				"permission.#",
				"permission.#",
			},
			{
				"permission.0.database_name",
				"permission.0.database_name",
			},
			{
				"permission.0.roles.#",
				"permission.0.roles.#",
			},
			{
				"permission.0.roles.0",
				"permission.0.roles.0",
			},
			{
				"permission.0.roles.1",
				"permission.0.roles.1",
			},
			{
				"global_permissions.#",
				"global_permissions.#",
			},
			{
				"connection_limits.0.max_questions_per_hour",
				"connection_limits.0.max_questions_per_hour",
			},
			{
				"connection_limits.0.max_updates_per_hour",
				"connection_limits.0.max_updates_per_hour",
			},
			{
				"connection_limits.0.max_connections_per_hour",
				"connection_limits.0.max_connections_per_hour",
			},
			{
				"connection_limits.0.max_user_connections",
				"connection_limits.0.max_user_connections",
			},
			{
				"authentication_plugin",
				"authentication_plugin",
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

func testAccDataSourceMDBMySQLUserCheck(datasourceName string, resourceName string, clusterName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBMySQLUserAttributesCheck(datasourceName, resourceName),
		testAccDataSourceMDBMySQLUserCheckResourceIDField(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "name", "john"),
		resource.TestCheckResourceAttr(datasourceName, "connection_limits.0.max_questions_per_hour", "10"),
		resource.TestCheckResourceAttr(datasourceName, "connection_limits.0.max_updates_per_hour", "20"),
		resource.TestCheckResourceAttr(datasourceName, "connection_limits.0.max_connections_per_hour", "30"),
		resource.TestCheckResourceAttr(datasourceName, "connection_limits.0.max_user_connections", "40"),
		resource.TestCheckResourceAttr(datasourceName, "global_permissions.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "authentication_plugin", "SHA256_PASSWORD"),
	)
}

func testAccDataSourceMDBMySQLUserCheckResourceIDField(resourceName string) resource.TestCheckFunc {
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

func testAccDataSourceMDBMySQLUserConfig(name string, useDataID bool) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
	name        = "%s"
	description = "MySQL User Terraform Datasource Test"
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

resource "yandex_mdb_mysql_database" "testdb" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
	name       = "testdb"
}

resource "yandex_mdb_mysql_user" "john" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
    name       = "john"
    password   = "password"

    permission {
      database_name = yandex_mdb_mysql_database.testdb.name
      roles         = ["ALL", "DROP", "DELETE"]
    }

	connection_limits {
	  max_questions_per_hour   = 10
	  max_updates_per_hour     = 20
	  max_connections_per_hour = 30
	  max_user_connections     = 40
	}
    
	global_permissions = ["PROCESS"]

	authentication_plugin = "SHA256_PASSWORD"
}

data "yandex_mdb_mysql_user" "john" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
	name       = yandex_mdb_mysql_user.john.name
}
`, name)
}

func testAccCheckMDBMySQLUserDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_mysql_user" {
			continue
		}

		clusterId, username, err := deconstructResourceId(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = config.sdk.MDB().MySQL().User().Get(context.Background(), &mysql.GetUserRequest{
			ClusterId: clusterId,
			UserName:  username,
		})

		if err == nil {
			return fmt.Errorf("MySQL User still exists")
		}
	}

	return nil
}
