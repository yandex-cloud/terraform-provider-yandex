package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
)

const redisResource = "yandex_mdb_redis_cluster.foo"

func mdbRedisClusterImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      redisResource,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"config.0.password", // not returned
			"host.0.subnet_id",  // computed on server side
			"host.1.subnet_id",  // computed on server side
		},
	}
}

// Test that a Redis Cluster can be created, updated and destroyed
func TestAccMDBRedisCluster_full(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-redis")
	redisDesc := "Redis Cluster Terraform Test"
	redisDesc2 := "Redis Cluster Terraform Test Updated"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			// Create Redis Cluster
			{
				Config: testAccMDBRedisClusterConfigMain(redisName, redisDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					testAccCheckMDBRedisClusterHasConfig(&r, "ALLKEYS_LRU", 100),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.nano", 17179869184),
					testAccCheckMDBRedisClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(redisResource),
				),
			},
			mdbRedisClusterImportStep(),
			// Change some options
			{
				Config: testAccMDBRedisClusterConfigUpdated(redisName, redisDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc2),
					testAccCheckMDBRedisClusterHasConfig(&r, "VOLATILE_LFU", 200),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.micro", 25769803776),
					testAccCheckMDBRedisClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckCreatedAtAttr(redisResource),
				),
			},
			mdbRedisClusterImportStep(),
			// Add new host
			{
				Config: testAccMDBRedisClusterConfigAddedHost(redisName, redisDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 2),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc2),
					testAccCheckMDBRedisClusterHasConfig(&r, "VOLATILE_LFU", 200),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.micro", 25769803776),
					testAccCheckMDBRedisClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckCreatedAtAttr(redisResource),
				),
			},
			mdbRedisClusterImportStep(),
		},
	})
}

func testAccCheckMDBRedisClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_redis_cluster" {
			continue
		}

		_, err := config.sdk.MDB().Redis().Cluster().Get(context.Background(), &redis.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Redis Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBRedisClusterExists(n string, r *redis.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().Redis().Cluster().Get(context.Background(), &redis.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Redis Cluster not found")
		}

		*r = *found

		resp, err := config.sdk.MDB().Redis().Cluster().ListHosts(context.Background(), &redis.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != hosts {
			return fmt.Errorf("Expected %d hosts, got %d", hosts, len(resp.Hosts))
		}

		return nil
	}
}

func testAccCheckMDBRedisClusterHasConfig(r *redis.Cluster, maxmemoryPolicy string, timeout int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := extractRedisConfig(r.Config)
		if c.maxmemoryPolicy != maxmemoryPolicy {
			return fmt.Errorf("Expected config.maxmemory_policy '%s', got '%s'", maxmemoryPolicy, c.maxmemoryPolicy)
		}
		if c.timeout != timeout {
			return fmt.Errorf("Expected config.timeout '%d', got '%d'", timeout, c.timeout)
		}
		return nil
	}
}
func testAccCheckMDBRedisClusterHasResources(r *redis.Cluster, resourcePresetID string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("Expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("Expected label with key '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBRedisClusterContainsLabel(r *redis.Cluster, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := r.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

const redisVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}
`

func testAccMDBRedisClusterConfigMain(name, desc string) string {
	return fmt.Sprintf(redisVPCDependencies+`
resource "yandex_mdb_redis_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  labels = {
    test_key = "test_value"
  }

  config {
    password         = "passw0rd"
    timeout          = 100
    maxmemory_policy = "ALLKEYS_LRU"
  }

  resources {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }
}
`, name, desc)
}

func testAccMDBRedisClusterConfigUpdated(name, desc string) string {
	return fmt.Sprintf(redisVPCDependencies+`
resource "yandex_mdb_redis_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  labels = {
    new_key = "new_value"
  }

  config {
    password         = "passw0rd"
    timeout          = 200
    maxmemory_policy = "VOLATILE_LFU"
  }

  resources {
    resource_preset_id = "hm1.micro"
    disk_size          = 24
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }
}
`, name, desc)
}

func testAccMDBRedisClusterConfigAddedHost(name, desc string) string {
	return fmt.Sprintf(redisVPCDependencies+`
resource "yandex_mdb_redis_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  labels = {
    new_key = "new_value"
  }

  config {
    password         = "passw0rd"
    timeout          = 200
    maxmemory_policy = "VOLATILE_LFU"
  }

  resources {
    resource_preset_id = "hm1.micro"
    disk_size          = 24
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }

  host {
    zone      = "ru-central1-b"
    subnet_id = "${yandex_vpc_subnet.bar.id}"
  }
}
`, name, desc)
}
