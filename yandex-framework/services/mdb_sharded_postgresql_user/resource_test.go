package mdb_sharded_postgresql_user_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	clusterResourceName   = "yandex_mdb_sharded_postgresql_cluster.foo"
	userResourceNameAlice = "yandex_mdb_sharded_postgresql_user.alice"
	userResourceNameBob   = "yandex_mdb_sharded_postgresql_user.bob"

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

// Test that a Sharded PostgreSQL User can be created, updated and destroyed
func TestAccMDBShardedPostgreSQLUser_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-sharded_postgresql-user")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBShardedPostgreSQLUserConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(userResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(userResourceNameAlice, "settings.connection_limit", "5"),
					resource.TestCheckResourceAttr(userResourceNameAlice, "settings.connection_retries", "5"),
				),
			},
			mdbShardedPostgreSQLUserImportStep(userResourceNameAlice),
			{
				Config: testAccMDBShardedPostgreSQLUserConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(userResourceNameBob, "name", "bob"),
				),
			},
			mdbShardedPostgreSQLUserImportStep(userResourceNameAlice),
			mdbShardedPostgreSQLUserImportStep(userResourceNameBob),
			{
				Config: testAccMDBShardedPostgreSQLUserConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(userResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(userResourceNameAlice, "settings.connection_limit", "10"),
					resource.TestCheckResourceAttr(userResourceNameAlice, "settings.connection_retries", "10"),
				),
			},
			mdbShardedPostgreSQLUserImportStep(userResourceNameAlice),
		},
	})
}

func mdbShardedPostgreSQLUserImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", // password is not returned
		},
	}
}

func testAccMDBShardedPostgreSQLUserConfigStep0(name string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_sharded_postgresql_cluster" "foo" {
	name        = "%s"
	description = "Sharded PostgreSQL User Terraform Test"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id

	config = {
		sharded_postgresql_config = {
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

// Create cluster, user and database
func testAccMDBShardedPostgreSQLUserConfigStep1(name string) string {
	return testAccMDBShardedPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_sharded_postgresql_user" "alice" {
	cluster_id = yandex_mdb_sharded_postgresql_cluster.foo.id
	name       = "alice"
	password   = "P@ssw0rd"
	settings = {
		connection_limit = 5
		connection_retries = 5
	}
	grants = ["reader", "writer"]
}`
}

// Create another user with grants and settings
func testAccMDBShardedPostgreSQLUserConfigStep2(name string) string {
	return testAccMDBShardedPostgreSQLUserConfigStep1(name) + `
resource "yandex_mdb_sharded_postgresql_user" "bob" {
	cluster_id = yandex_mdb_sharded_postgresql_cluster.foo.id
	name       = "bob"
	password   = "P@ssw0rd"
	grants = ["reader", "writer"]
	settings = {
	}
}`
}

// Change Alice's settings
func testAccMDBShardedPostgreSQLUserConfigStep3(name string) string {
	return testAccMDBShardedPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_sharded_postgresql_user" "alice" {
	cluster_id = yandex_mdb_sharded_postgresql_cluster.foo.id
	name       = "alice"
	password   = "P@ssw0rd"
	grants = ["reader"]
	settings = {
		connection_limit = 10
		connection_retries = 10
	}
}`
}
