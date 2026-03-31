package mdb_mysql_database_v2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func waitForOperations(seconds time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		time.Sleep(seconds * time.Second)
		return nil
	}
}

func TestAccMDBMySQLDatabase_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-mysql-database")
	dbName := acctest.RandomWithPrefix("testdb")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLDatabaseBasic(clusterName, dbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					testAccCheckMDBMySQLDatabaseResourceIDField(mysqlDatabaseResourceName),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName, "name", dbName,
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					resource.TestCheckResourceAttrSet(
						mysqlDatabaseResourceName, "cluster_id",
					),
				),
			},
			mdbMySQLDatabaseImportStep(mysqlDatabaseResourceName),
		},
	})
}

func TestAccMDBMySQLDatabase_update(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-mysql-database")
	dbName := acctest.RandomWithPrefix("testdb")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLDatabaseBasic(clusterName, dbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
				),
			},
			{
				Config: testAccMDBMySQLDatabaseWithDeletionProtection(clusterName, dbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_ENABLED",
					),
				),
			},
			mdbMySQLDatabaseImportStep(mysqlDatabaseResourceName),
			{
				Config: testAccMDBMySQLDatabaseBasic(clusterName, dbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
		},
	})
}

func TestAccMDBMySQLDatabase_full(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix(testMySQLDatabasePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLDatabaseConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName, "name", "testdb",
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
				),
			},
			mdbMySQLDatabaseImportStep(mysqlDatabaseResourceName),
			{
				Config: testAccMDBMySQLDatabaseConfigStep2(clusterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBMySQLClusterHasDatabase(
						t, "testdb", "DELETION_PROTECTION_MODE_ENABLED",
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_ENABLED",
					),
				),
			},
			mdbMySQLDatabaseImportStep(mysqlDatabaseResourceName),
			{
				Config: testAccMDBMySQLDatabaseConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName1),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName1, "name", "testdb1",
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName1,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_INHERITED",
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName, "name", "testdb",
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
				),
			},
			mdbMySQLDatabaseImportStep(mysqlDatabaseResourceName1),
			{
				Config: testAccMDBMySQLDatabaseConfigStep4(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName),
					testAccCheckMDBMySQLDatabaseExists(mysqlDatabaseResourceName1),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName1,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					resource.TestCheckResourceAttr(
						mysqlDatabaseResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
		},
	})
}

func testAccMDBMySQLDatabaseBasic(clusterName, dbName string) string {
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
`, clusterName, dbName)
}

func testAccMDBMySQLDatabaseWithDeletionProtection(clusterName, dbName string) string {
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
`, clusterName, dbName)
}

func testAccMDBMySQLDatabaseConfigStep0(name string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_mysql_cluster_v2" "foo" {
  name        = "%s"
  description = "MySQL Database Terraform Test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  hosts = {
    "host1" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
    "host2" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.bar.id
    }
  }
}
`, name)
}

func testAccMDBMySQLDatabaseConfigStep1(name string) string {
	return testAccMDBMySQLDatabaseConfigStep0(name) + `
resource "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "testdb"
}
`
}

func testAccMDBMySQLDatabaseConfigStep2(name string) string {
	return testAccMDBMySQLDatabaseConfigStep0(name) + `
resource "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id               = yandex_mdb_mysql_cluster_v2.foo.id
  name                     = "testdb"
  deletion_protection_mode = "DELETION_PROTECTION_MODE_ENABLED"
}
`
}

func testAccMDBMySQLDatabaseConfigStep3(name string) string {
	return testAccMDBMySQLDatabaseConfigStep0(name) + `
resource "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "testdb"
}

resource "yandex_mdb_mysql_database_v2" "testdb1" {
  cluster_id               = yandex_mdb_mysql_cluster_v2.foo.id
  name                     = "testdb1"
  deletion_protection_mode = "DELETION_PROTECTION_MODE_INHERITED"

  depends_on = [yandex_mdb_mysql_database_v2.testdb]
}
`
}

func testAccMDBMySQLDatabaseConfigStep4(name string) string {
	return testAccMDBMySQLDatabaseConfigStep0(name) + `
resource "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "testdb"
}

resource "yandex_mdb_mysql_database_v2" "testdb1" {
  cluster_id               = yandex_mdb_mysql_cluster_v2.foo.id
  name                     = "testdb1"
  deletion_protection_mode = "DELETION_PROTECTION_MODE_DISABLED"

  depends_on = [yandex_mdb_mysql_database_v2.testdb]
}
`
}
