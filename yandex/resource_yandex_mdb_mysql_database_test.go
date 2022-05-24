package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	mysqlDatabaseResourceName1 = "yandex_mdb_mysql_database.testdb1"
	mysqlDatabaseResourceName2 = "yandex_mdb_mysql_database.testdb2"
)

// Test that a MySQL Database can be created, updated and destroyed
func TestAccMDBMySQLDatabase_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-mysql")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLDatabaseConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mysqlDatabaseResourceName1, "name", "testdb1"),
					testAccCheckMDBMySQLClusterHasDatabases(mysqlResource, []string{"testdb1"}),
				),
			},
			mdbMySQLDatabaseImportStep(mysqlDatabaseResourceName1),
			{
				Config: testAccMDBMySQLDatabaseConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mysqlDatabaseResourceName2, "name", "testdb2"),
					testAccCheckMDBMySQLClusterHasDatabases(mysqlResource, []string{"testdb1", "testdb2"}),
				),
			},
			mdbMySQLDatabaseImportStep(mysqlDatabaseResourceName2),
		},
	})
}

func mdbMySQLDatabaseImportStep(clusterName string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      clusterName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccMDBMySQLDatabaseConfigStep0(clusterName string) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
	name        = "%s"
	description = "MySQL Database Terraform Test"
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
`, clusterName)
}

// Create database
func testAccMDBMySQLDatabaseConfigStep1(clusterName string) string {
	return testAccMDBMySQLDatabaseConfigStep0(clusterName) + `
resource "yandex_mdb_mysql_database" "testdb1" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
	name       = "testdb1"
}
`
}

// Create another database
func testAccMDBMySQLDatabaseConfigStep2(clusterName string) string {
	return testAccMDBMySQLDatabaseConfigStep1(clusterName) + `
resource "yandex_mdb_mysql_database" "testdb2" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
	name       = "testdb2"
}
`
}
