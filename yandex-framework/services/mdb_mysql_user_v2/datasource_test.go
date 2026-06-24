package mdb_mysql_user_v2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	mysqlUserV2DataSourceName = "data.yandex_mdb_mysql_user_v2.testuser"
)

func TestAccDataSourceMDBMySQLUserV2_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-mysql-user-v2-ds")
	userName := acctest.RandomWithPrefix("testuser")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLUserV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMySQLUserV2Config(clusterName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					testAccCheckMDBMySQLUserV2ResourceIDField(mysqlUserV2ResourceName),
					testAccDataSourceMDBMySQLUserV2AttributesCheck(
						mysqlUserV2DataSourceName,
						mysqlUserV2ResourceName,
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2DataSourceName, "name", userName,
					),
					resource.TestCheckResourceAttrSet(
						mysqlUserV2DataSourceName, "cluster_id",
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2DataSourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
		},
	})
}

func TestAccDataSourceMDBMySQLUserV2_withDeletionProtection(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-mysql-user-v2-ds")
	userName := acctest.RandomWithPrefix("testuser")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLUserV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMySQLUserV2WithDeletionProtectionConfig(clusterName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					testAccDataSourceMDBMySQLUserV2AttributesCheck(
						mysqlUserV2DataSourceName,
						mysqlUserV2ResourceName,
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2DataSourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_ENABLED",
					),
				),
			},
			{
				Config: testAccDataSourceMDBMySQLUserV2Config(clusterName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					resource.TestCheckResourceAttr(
						mysqlUserV2DataSourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
		},
	})
}

func testAccDataSourceMDBMySQLUserV2AttributesCheck(
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

		attrsToCheck := []struct{ dsPath, rsPath string }{
			{"cluster_id", "cluster_id"},
			{"name", "name"},
			{"deletion_protection_mode", "deletion_protection_mode"},
		}
		for _, attr := range attrsToCheck {
			dsVal, dsOk := ds.Primary.Attributes[attr.dsPath]
			if !dsOk {
				return fmt.Errorf("%s is not present in datasource attributes", attr.dsPath)
			}
			rsVal, rsOk := rs.Primary.Attributes[attr.rsPath]
			if !rsOk {
				return fmt.Errorf("%s is not present in resource attributes", attr.rsPath)
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

func testAccDataSourceMDBMySQLUserV2Config(clusterName, userName string) string {
	return fmt.Sprintf(mysqlUserV2VPCDependencies+`
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
  name       = "testdb"
}

resource "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "%s"
  password   = "Password123!"

  permission {
    database_name = yandex_mdb_mysql_database_v2.testdb.name
    roles         = ["ALL"]
  }
}

data "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = yandex_mdb_mysql_user_v2.testuser.name
}
`, clusterName, userName)
}

func testAccDataSourceMDBMySQLUserV2WithDeletionProtectionConfig(clusterName, userName string) string {
	return fmt.Sprintf(mysqlUserV2VPCDependencies+`
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
  name       = "testdb"
}

resource "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id               = yandex_mdb_mysql_cluster_v2.foo.id
  name                     = "%s"
  password                 = "Password123!"
  deletion_protection_mode = "DELETION_PROTECTION_MODE_ENABLED"

  permission {
    database_name = yandex_mdb_mysql_database_v2.testdb.name
    roles         = ["ALL"]
  }
}

data "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = yandex_mdb_mysql_user_v2.testuser.name
}
`, clusterName, userName)
}
