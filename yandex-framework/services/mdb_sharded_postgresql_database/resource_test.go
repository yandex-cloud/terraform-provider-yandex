package mdb_sharded_postgresql_database_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	clusterResourceName     = "yandex_mdb_sharded_postgresql_cluster.foo"
	dbResourceNameTestdb    = "yandex_mdb_sharded_postgresql_database.testdb"
	dbResourceNameAnotherdb = "yandex_mdb_sharded_postgresql_database.anotherdb"

	VPCDependencies = `
	resource "yandex_vpc_network" "foo" {}
	
	resource "yandex_vpc_subnet" "foo" {
	  zone           = "ru-central1-a"
	  network_id     = yandex_vpc_network.foo.id
	  v4_cidr_blocks = ["10.1.0.0/24"]
	}
	`
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// Test that a Sharded PostgreSQL Database can be created, updated and destroyed
func TestAccMDBShardedPostgreSQLDatabase_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-sharded_postgresql-database")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBShardedPostgreSQLDatabaseConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dbResourceNameTestdb, "name", "testdb"),
				),
			},
			mdbShardedPostgreSQLDatabaseImportStep(dbResourceNameTestdb),
			{
				Config: testAccMDBShardedPostgreSQLDatabaseConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dbResourceNameAnotherdb, "name", "anotherdb"),
				),
			},
			mdbShardedPostgreSQLDatabaseImportStep(dbResourceNameTestdb),
			mdbShardedPostgreSQLDatabaseImportStep(dbResourceNameAnotherdb),
			{
				Config: testAccMDBShardedPostgreSQLDatabaseConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dbResourceNameTestdb, "name", "testdb"),
				),
			},
			mdbShardedPostgreSQLDatabaseImportStep(dbResourceNameTestdb),
		},
	})
}

func mdbShardedPostgreSQLDatabaseImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", // password is not returned
		},
	}
}

func testAccMDBShardedPostgreSQLDatabaseConfigStep0(name string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_sharded_postgresql_cluster" "foo" {
	name        = "%s"
	description = "Sharded PostgreSQL User Terraform Test"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id

	config = {
		sharded_postgresql_config = {
			common = {
				"console_password": "P@ssw0rd"
			}
			router = {
				resources = {
					resource_preset_id = "s2.micro"
					disk_size          = 10
					disk_type_id       = "network-ssd"
				}
			}
		}
	}

	hosts = {
		"router1" = {
			zone    = "ru-central1-a"
			subnet_id  = yandex_vpc_subnet.foo.id
			type	   = "ROUTER"
		}
	}
}
`, name)
}

func testAccMDBShardedPostgreSQLDatabaseConfigStep1(name string) string {
	return testAccMDBShardedPostgreSQLDatabaseConfigStep0(name) + `
resource "yandex_mdb_sharded_postgresql_database" "testdb" {
	cluster_id = yandex_mdb_sharded_postgresql_cluster.foo.id
	name       = "testdb"
}`
}

func testAccMDBShardedPostgreSQLDatabaseConfigStep2(name string) string {
	return testAccMDBShardedPostgreSQLDatabaseConfigStep1(name) + `
resource "yandex_mdb_sharded_postgresql_database" "anotherdb" {
	cluster_id = yandex_mdb_sharded_postgresql_cluster.foo.id
	name       = "anotherdb"
}`
}

func testAccMDBShardedPostgreSQLDatabaseConfigStep3(name string) string {
	return testAccMDBShardedPostgreSQLDatabaseConfigStep0(name) + `
resource "yandex_mdb_sharded_postgresql_database" "testdb" {
	cluster_id = yandex_mdb_sharded_postgresql_cluster.foo.id
	name       = "testdb"
}`
}
