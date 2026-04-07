package mdb_mysql_database_v2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	mysqlDatabaseDataSourceName = "data.yandex_mdb_mysql_database_v2.testdb"
)

func TestAccDataSourceMDBMySQLDatabase_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-mysql-database-datasource")
	dbName := acctest.RandomWithPrefix("testdb")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMySQLDatabaseConfig(clusterName, dbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					testAccCheckMDBMySQLDatabaseResourceIDField(mysqlDatabaseResourceName),
					testAccDataSourceMDBMySQLDatabaseAttributesCheck(
						mysqlDatabaseDataSourceName,
						mysqlDatabaseResourceName,
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseDataSourceName, "name", dbName,
					),
					resource.TestCheckResourceAttrSet(
						mysqlDatabaseDataSourceName, "cluster_id",
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseDataSourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
				),
			},
		},
	})
}

func TestAccDataSourceMDBMySQLDatabase_withDeletionProtection(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-mysql-database-datasource")
	dbName := acctest.RandomWithPrefix("testdb")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMySQLDatabaseWithDeletionProtectionConfig(
					clusterName, dbName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					testAccDataSourceMDBMySQLDatabaseAttributesCheck(
						mysqlDatabaseDataSourceName,
						mysqlDatabaseResourceName,
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseDataSourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_ENABLED",
					),
				),
			},
			{
				Config: testAccDataSourceMDBMySQLDatabaseConfig(clusterName, dbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					resource.TestCheckResourceAttr(
						mysqlDatabaseDataSourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
		},
	})
}

func testAccDataSourceMDBMySQLDatabaseAttributesCheck(
	datasourceName, resourceName string,
) resource.TestCheckFunc {
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
			return fmt.Errorf(
				"datasource ID %s does not match resource ID %s",
				ds.Primary.ID, rs.Primary.ID,
			)
		}

		attrsToCheck := []struct {
			dsPath string
			rsPath string
		}{
			{"cluster_id", "cluster_id"},
			{"name", "name"},
			{"deletion_protection_mode", "deletion_protection_mode"},
		}

		for _, attr := range attrsToCheck {
			dsVal, dsOk := ds.Primary.Attributes[attr.dsPath]
			if !dsOk {
				return fmt.Errorf(
					"%s is not present in datasource attributes", attr.dsPath,
				)
			}

			rsVal, rsOk := rs.Primary.Attributes[attr.rsPath]
			if !rsOk {
				return fmt.Errorf(
					"%s is not present in resource attributes", attr.rsPath,
				)
			}

			if dsVal != rsVal {
				return fmt.Errorf(
					"attribute %s: datasource has %q, resource has %q",
					attr.dsPath, dsVal, rsVal,
				)
			}
		}

		return nil
	}
}

func testAccDataSourceMDBMySQLDatabaseConfig(clusterName, dbName string) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster_v2" "foo" {
  name        = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  hosts = {
    "host1" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
  }

  resources {
    resource_preset_id = "s2.micro"
    disk_size          = 10
    disk_type_id       = "network-ssd"
  }
}

resource "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "%s"
}

data "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = yandex_mdb_mysql_database_v2.testdb.name
}
`, clusterName, dbName)
}

func testAccDataSourceMDBMySQLDatabaseWithDeletionProtectionConfig(
	clusterName, dbName string,
) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster_v2" "foo" {
  name        = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  hosts = {
    "host1" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
  }

  resources {
    resource_preset_id = "s2.micro"
    disk_size          = 10
    disk_type_id       = "network-ssd"
  }
}

resource "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id               = yandex_mdb_mysql_cluster_v2.foo.id
  name                     = "%s"
  deletion_protection_mode = "DELETION_PROTECTION_MODE_ENABLED"
}

data "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = yandex_mdb_mysql_database_v2.testdb.name
}
`, clusterName, dbName)
}
