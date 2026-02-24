package mdb_sharded_postgresql_shard_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	clusterResourceName     = "yandex_mdb_sharded_postgresql_cluster.foo"
	shardResourceNameShard1 = "yandex_mdb_sharded_postgresql_shard.shard1"
	shardResourceNameShard2 = "yandex_mdb_sharded_postgresql_shard.shard2"

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

// Test that a Sharded PostgreSQL shard can be created, updated and destroyed
func TestAccMDBShardedPostgreSQLShard_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-sharded_postgresql-shard")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBShardedPostgreSQLShardConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(shardResourceNameShard1, "name", "shard1"),
				),
			},
			mdbShardedPostgreSQLShardImportStep(shardResourceNameShard1),
			{
				Config: testAccMDBShardedPostgreSQLShardConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(shardResourceNameShard1, "name", "shard1"),
					resource.TestCheckResourceAttr(shardResourceNameShard2, "name", "shard2"),
				),
			},
			mdbShardedPostgreSQLShardImportStep(shardResourceNameShard1),
			mdbShardedPostgreSQLShardImportStep(shardResourceNameShard2),
			{
				Config: testAccMDBShardedPostgreSQLShardConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(shardResourceNameShard1, "name", "shard1"),
				),
			},
			mdbShardedPostgreSQLShardImportStep(shardResourceNameShard1),
		},
	})
}

func mdbShardedPostgreSQLShardImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", // password is not returned
		},
	}
}

func testAccMDBPostgreSQLCluster(name string) string {
	return fmt.Sprintf(`
resource "yandex_mdb_postgresql_cluster_v2" "%s" {
  name        = "%s"
  environment = "PRODUCTION" 
  network_id  = yandex_vpc_network.foo.id

  hosts = {
    "host" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
  }

  config {
    version = "17"
	resources {
		resource_preset_id = "s2.micro"
		disk_size          = 10
		disk_type_id       = "network-ssd"
	}
  }
}
`, name, name)
}

func testAccMDBShardedPostgreSQLShardConfigStep0(name string) string {
	return fmt.Sprintf(VPCDependencies+`
%s

%s

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
`, testAccMDBPostgreSQLCluster(fmt.Sprintf("%s-shard1", name)), testAccMDBPostgreSQLCluster(fmt.Sprintf("%s-shard2", name)), name)
}

func testAccMDBShardedPostgreSQLShardConfigStep1(name string) string {
	return testAccMDBShardedPostgreSQLShardConfigStep0(name) + fmt.Sprintf(`
resource "yandex_mdb_sharded_postgresql_shard" "shard1" {
	cluster_id = yandex_mdb_sharded_postgresql_cluster.foo.id
	name       = "shard1"
	shard_spec = {
		mdb_postgresql = yandex_mdb_postgresql_cluster_v2.%s-shard1.id
	}
}`, name)
}

func testAccMDBShardedPostgreSQLShardConfigStep2(name string) string {
	return testAccMDBShardedPostgreSQLShardConfigStep1(name) + fmt.Sprintf(`
resource "yandex_mdb_sharded_postgresql_shard" "shard2" {
	cluster_id = yandex_mdb_sharded_postgresql_cluster.foo.id
	name       = "shard2"
	shard_spec = {
		mdb_postgresql = yandex_mdb_postgresql_cluster_v2.%s-shard2.id
	}
}`, name)
}

func testAccMDBShardedPostgreSQLShardConfigStep3(name string) string {
	return testAccMDBShardedPostgreSQLShardConfigStep0(name) + fmt.Sprintf(`
resource "yandex_mdb_sharded_postgresql_shard" "shard1" {
	cluster_id = yandex_mdb_sharded_postgresql_cluster.foo.id
	name       = "shard1"
	shard_spec = {
		mdb_postgresql = yandex_mdb_postgresql_cluster_v2.%s-shard1.id
	}
}`, name)
}
