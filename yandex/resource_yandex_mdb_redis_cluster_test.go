package yandex

import (
	"context"
	"fmt"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
)

const redisResource = "yandex_mdb_redis_cluster.foo"
const redisResourceSharded = "yandex_mdb_redis_cluster.bar"

func init() {
	resource.AddTestSweepers("yandex_mdb_redis_cluster", &resource.Sweeper{
		Name: "yandex_mdb_redis_cluster",
		F:    testSweepMDBRedisCluster,
	})
}

func testSweepMDBRedisCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().Redis().Cluster().List(conf.Context(), &redis.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Redis clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBRedisCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Redis cluster %q", c.Id))
		} else {
			if !sweepVPCNetwork(conf, c.NetworkId) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep VPC network %q", c.NetworkId))
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBRedisCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBRedisClusterOnce, conf, "Redis cluster", id)
}

func sweepMDBRedisClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexMDBRedisClusterDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.MDB().Redis().Cluster().Delete(ctx, &redis.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func mdbRedisClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"config.0.password", // not returned
			"health",            // volatile value
			"host",              // the order of hosts differs
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
				Config: testAccMDBRedisClusterConfigMain(redisName, redisDesc, "5.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckResourceAttrSet(redisResource, "host.0.fqdn"),
					testAccCheckMDBRedisClusterHasConfig(&r, "ALLKEYS_LRU", 100, "5.0"),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.nano", 17179869184),
					testAccCheckMDBRedisClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(redisResource),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Change some options
			{
				Config: testAccMDBRedisClusterConfigUpdated(redisName, redisDesc2, "5.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc2),
					resource.TestCheckResourceAttrSet(redisResource, "host.0.fqdn"),
					testAccCheckMDBRedisClusterHasConfig(&r, "VOLATILE_LFU", 200, "5.0"),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.micro", 25769803776),
					testAccCheckMDBRedisClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckCreatedAtAttr(redisResource),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Add new host
			{
				Config: testAccMDBRedisClusterConfigAddedHost(redisName, redisDesc2, "5.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 2),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc2),
					resource.TestCheckResourceAttrSet(redisResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(redisResource, "host.1.fqdn"),
					testAccCheckMDBRedisClusterHasConfig(&r, "VOLATILE_LFU", 200, "5.0"),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.micro", 25769803776),
					testAccCheckMDBRedisClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckCreatedAtAttr(redisResource),
				),
			},
			mdbRedisClusterImportStep(redisResource),
		},
	})
}

// Test that a sharded Redis Cluster can be created, updated and destroyed
func TestAccMDBRedisCluster_sharded(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-sharded-redis")
	redisDesc := "Sharded Redis Cluster Terraform Test"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			// Create Redis Cluster
			{
				Config: testAccMDBRedisShardedClusterConfig(redisName, redisDesc, "5.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResourceSharded, &r, 3),
					resource.TestCheckResourceAttr(redisResourceSharded, "name", redisName),
					resource.TestCheckResourceAttr(redisResourceSharded, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResourceSharded, "description", redisDesc),
					testAccCheckMDBRedisClusterHasShards(&r, []string{"first", "second", "third"}),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.nano", 17179869184),
					testAccCheckCreatedAtAttr(redisResourceSharded),
				),
			},
			mdbRedisClusterImportStep(redisResourceSharded),
			// Add new shard, delete old shard
			{
				Config: testAccMDBRedisShardedClusterConfigUpdated(redisName, redisDesc, "5.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResourceSharded, &r, 3),
					resource.TestCheckResourceAttr(redisResourceSharded, "name", redisName),
					resource.TestCheckResourceAttr(redisResourceSharded, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResourceSharded, "description", redisDesc),
					testAccCheckMDBRedisClusterHasShards(&r, []string{"first", "second", "new"}),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.nano", 17179869184),
					testAccCheckCreatedAtAttr(redisResourceSharded),
				),
			},
			mdbRedisClusterImportStep(redisResourceSharded),
		},
	})
}

func TestAccMDBRedis6Cluster_full(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-redis")
	redisDesc := "Redis 6 Cluster Terraform Test"
	redisDesc2 := "Redis 6 Cluster Terraform Test Updated"
	folderID := getExampleFolderID()
	version := "6.0"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			// Create Redis Cluster
			{
				Config: testAccMDBRedisClusterConfigMain(redisName, redisDesc, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckResourceAttrSet(redisResource, "host.0.fqdn"),
					testAccCheckMDBRedisClusterHasConfig(&r, "ALLKEYS_LRU", 100, version),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.nano", 17179869184),
					testAccCheckMDBRedisClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(redisResource),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Change some options
			{
				Config: testAccMDBRedisClusterConfigUpdated(redisName, redisDesc2, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc2),
					resource.TestCheckResourceAttrSet(redisResource, "host.0.fqdn"),
					testAccCheckMDBRedisClusterHasConfig(&r, "VOLATILE_LFU", 200, version),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.micro", 25769803776),
					testAccCheckMDBRedisClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckCreatedAtAttr(redisResource),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Add new host
			{
				Config: testAccMDBRedisClusterConfigAddedHost(redisName, redisDesc2, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 2),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc2),
					resource.TestCheckResourceAttrSet(redisResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(redisResource, "host.1.fqdn"),
					testAccCheckMDBRedisClusterHasConfig(&r, "VOLATILE_LFU", 200, version),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.micro", 25769803776),
					testAccCheckMDBRedisClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckCreatedAtAttr(redisResource),
				),
			},
			mdbRedisClusterImportStep(redisResource),
		},
	})
}

// Test that a sharded Redis Cluster can be created, updated and destroyed
func TestAccMDBRedis6Cluster_sharded(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-sharded-redis")
	redisDesc := "Sharded Redis Cluster Terraform Test"
	folderID := getExampleFolderID()
	version := "6.0"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			// Create Redis Cluster
			{
				Config: testAccMDBRedisShardedClusterConfig(redisName, redisDesc, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResourceSharded, &r, 3),
					resource.TestCheckResourceAttr(redisResourceSharded, "name", redisName),
					resource.TestCheckResourceAttr(redisResourceSharded, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResourceSharded, "description", redisDesc),
					testAccCheckMDBRedisClusterHasShards(&r, []string{"first", "second", "third"}),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.nano", 17179869184),
					testAccCheckCreatedAtAttr(redisResourceSharded),
				),
			},
			mdbRedisClusterImportStep(redisResourceSharded),
			// Add new shard, delete old shard
			{
				Config: testAccMDBRedisShardedClusterConfigUpdated(redisName, redisDesc, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResourceSharded, &r, 3),
					resource.TestCheckResourceAttr(redisResourceSharded, "name", redisName),
					resource.TestCheckResourceAttr(redisResourceSharded, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResourceSharded, "description", redisDesc),
					testAccCheckMDBRedisClusterHasShards(&r, []string{"first", "second", "new"}),
					testAccCheckMDBRedisClusterHasResources(&r, "hm1.nano", 17179869184),
					testAccCheckCreatedAtAttr(redisResourceSharded),
				),
			},
			mdbRedisClusterImportStep(redisResourceSharded),
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

func testAccCheckMDBRedisClusterHasShards(r *redis.Cluster, shards []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().Redis().Cluster().ListShards(context.Background(), &redis.ListClusterShardsRequest{
			ClusterId: r.Id,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Shards) != len(shards) {
			return fmt.Errorf("Expected %d shards, got %d", len(shards), len(resp.Shards))
		}
		for _, s := range shards {
			found := false
			for _, rs := range resp.Shards {
				if s == rs.Name {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("Shard '%s' not found", s)
			}
		}
		return nil
	}
}

func testAccCheckMDBRedisClusterHasConfig(r *redis.Cluster, maxmemoryPolicy string, timeout int64,
	version string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := extractRedisConfig(r.Config)
		if c.maxmemoryPolicy != maxmemoryPolicy {
			return fmt.Errorf("Expected config.maxmemory_policy '%s', got '%s'", maxmemoryPolicy, c.maxmemoryPolicy)
		}
		if c.timeout != timeout {
			return fmt.Errorf("Expected config.timeout '%d', got '%d'", timeout, c.timeout)
		}
		if c.version != version {
			return fmt.Errorf("Expected config.version '%s', got '%s'", version, c.version)
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

// TODO: add more zones when v2 platform becomes available.
const redisVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}
`

func testAccMDBRedisClusterConfigMain(name, desc string, version string) string {
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
	version			 = "%s"
  }

  resources {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  host {
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }
}
`, name, desc, version)
}

func testAccMDBRedisClusterConfigUpdated(name, desc string, version string) string {
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
	version			 = "%s"
  }

  resources {
    resource_preset_id = "hm1.micro"
    disk_size          = 24
  }

  host {
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }
}
`, name, desc, version)
}

func testAccMDBRedisClusterConfigAddedHost(name, desc string, version string) string {
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
	version			 = "%s"
  }

  resources {
    resource_preset_id = "hm1.micro"
    disk_size          = 24
  }

  host {
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }

  host {
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }
}
`, name, desc, version)
}

func testAccMDBRedisShardedClusterConfig(name, desc string, version string) string {
	return fmt.Sprintf(redisVPCDependencies+`
resource "yandex_mdb_redis_cluster" "bar" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"
  sharded     = true

  config {
    password = "passw0rd"
	version  = "%s"
  }

  resources {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  host {
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.foo.id}"
	shard_name = "first"
  }

  host {
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.foo.id}"
	shard_name = "second"
  }

  host {
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.foo.id}"
	shard_name = "third"
  }
}
`, name, desc, version)
}

func testAccMDBRedisShardedClusterConfigUpdated(name, desc string, version string) string {
	return fmt.Sprintf(redisVPCDependencies+`
resource "yandex_mdb_redis_cluster" "bar" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"
  sharded     = true

  config {
    password = "passw0rd"
	version	 = "%s"
  }

  resources {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  host {
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.foo.id}"
	shard_name = "first"
  }

  host {
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.foo.id}"
	shard_name = "second"
  }

  host {
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.foo.id}"
	shard_name = "new"
  }
}
`, name, desc, version)
}
