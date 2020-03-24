package yandex

import (
	"context"
	"fmt"
	"sort"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
)

const chResource = "yandex_mdb_clickhouse_cluster.foo"
const chResourceSharded = "yandex_mdb_clickhouse_cluster.bar"

func init() {
	resource.AddTestSweepers("yandex_mdb_clickhouse_cluster", &resource.Sweeper{
		Name: "yandex_mdb_clickhouse_cluster",
		F:    testSweepMDBClickHouseCluster,
	})
}

func testSweepMDBClickHouseCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().Clickhouse().Cluster().List(conf.Context(), &clickhouse.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting ClickHouse clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBClickHouseCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ClickHouse cluster %q", c.Id))
		} else {
			if !sweepVPCNetwork(conf, c.NetworkId) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep VPC network %q", c.NetworkId))
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBClickHouseCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBClickHouseClusterOnce, conf, "ClickHouse cluster", id)
}

func sweepMDBClickHouseClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexMDBClickHouseClusterDeleteTimeout)
	defer cancel()

	op, err := conf.sdk.MDB().Clickhouse().Cluster().Delete(ctx, &clickhouse.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func mdbClickHouseClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"user",      // passwords are not returned
			"host",      // zookeeper hosts are not imported by default
			"zookeeper", // zookeeper spec is not imported by default
			"health",    // volatile value
		},
	}
}

// Test that a ClickHouse Cluster can be created, updated and destroyed
func TestAccMDBClickHouseCluster_full(t *testing.T) {
	t.Parallel()

	var r clickhouse.Cluster
	chName := acctest.RandomWithPrefix("tf-clickhouse")
	chDesc := "ClickHouse Cluster Terraform Test"
	chDesc2 := "ClickHouse Cluster Terraform Test Updated"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			// Create ClickHouse Cluster
			{
				Config: testAccMDBClickHouseClusterConfigMain(chName, chDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &r, 1),
					resource.TestCheckResourceAttr(chResource, "name", chName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "description", chDesc),
					resource.TestCheckResourceAttrSet(chResource, "host.0.fqdn"),
					testAccCheckMDBClickHouseClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBClickHouseClusterHasResources(&r, "s2.micro", "network-ssd", 17179869184),
					testAccCheckMDBClickHouseClusterHasUsers(chResource, map[string][]string{"john": {"testdb"}}),
					testAccCheckMDBClickHouseClusterHasDatabases(chResource, []string{"testdb"}),
					testAccCheckCreatedAtAttr(chResource),
				),
			},
			mdbClickHouseClusterImportStep(chResource),
			// Change some options
			{
				Config: testAccMDBClickHouseClusterConfigUpdated(chName, chDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &r, 1),
					resource.TestCheckResourceAttr(chResource, "name", chName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "description", chDesc2),
					resource.TestCheckResourceAttrSet(chResource, "host.0.fqdn"),
					testAccCheckMDBClickHouseClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckMDBClickHouseClusterHasResources(&r, "s2.micro", "network-ssd", 19327352832),
					testAccCheckMDBClickHouseClusterHasUsers(chResource, map[string][]string{"john": {"testdb"}, "mary": {"newdb", "testdb"}}),
					testAccCheckMDBClickHouseClusterHasDatabases(chResource, []string{"testdb", "newdb"}),
					testAccCheckCreatedAtAttr(chResource),
				),
			},
			mdbClickHouseClusterImportStep(chResource),
			// Add host, creates implicit ZooKeeper subcluster
			{
				Config: testAccMDBClickHouseClusterConfigHA(chName, chDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &r, 5),
					resource.TestCheckResourceAttr(chResource, "name", chName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "description", chDesc2),
					resource.TestCheckResourceAttrSet(chResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(chResource, "host.1.fqdn"),
					testAccCheckMDBClickHouseClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckMDBClickHouseClusterHasResources(&r, "s2.micro", "network-ssd", 19327352832),
					testAccCheckMDBClickHouseZooKeeperSubclusterHasResources(&r, "s2.micro", "network-ssd", 10737418240),
					testAccCheckMDBClickHouseClusterHasUsers(chResource, map[string][]string{"john": {"testdb"}, "mary": {"newdb", "testdb"}}),
					testAccCheckMDBClickHouseClusterHasDatabases(chResource, []string{"testdb", "newdb"}),
					testAccCheckCreatedAtAttr(chResource),
				),
			},
			mdbClickHouseClusterImportStep(chResource),
		},
	})
}

// Test that a sharded ClickHouse Cluster can be created, updated and destroyed
func TestAccMDBClickHouseCluster_sharded(t *testing.T) {
	t.Skip("CLOUD-38473")

	t.Parallel()

	var r clickhouse.Cluster
	chName := acctest.RandomWithPrefix("tf-clickhouse-sharded")
	chDesc := "ClickHouse Sharded Cluster Terraform Test"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			// Create sharded ClickHouse Cluster
			{
				Config: testAccMDBClickHouseClusterConfigSharded(chName, chDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceSharded, &r, 2),
					resource.TestCheckResourceAttr(chResourceSharded, "name", chName),
					resource.TestCheckResourceAttr(chResourceSharded, "folder_id", folderID),
					resource.TestCheckResourceAttrSet(chResourceSharded, "host.0.fqdn"),
					testAccCheckMDBClickHouseClusterHasShards(&r, []string{"shard1", "shard2"}),
					testAccCheckMDBClickHouseClusterHasResources(&r, "s2.micro", "network-ssd", 10737418240),
					testAccCheckMDBClickHouseClusterHasUsers(chResourceSharded, map[string][]string{"john": {"testdb"}}),
					testAccCheckMDBClickHouseClusterHasDatabases(chResourceSharded, []string{"testdb"}),
					testAccCheckCreatedAtAttr(chResourceSharded),
				),
			},
			mdbClickHouseClusterImportStep(chResourceSharded),
			// Add new shard, delete old shard
			{
				Config: testAccMDBClickHouseClusterConfigShardedUpdated(chName, chDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceSharded, &r, 2),
					resource.TestCheckResourceAttr(chResourceSharded, "name", chName),
					resource.TestCheckResourceAttr(chResourceSharded, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResourceSharded, "description", chDesc),
					resource.TestCheckResourceAttrSet(chResourceSharded, "host.0.fqdn"),
					testAccCheckMDBClickHouseClusterHasShards(&r, []string{"shard1", "shard3"}),
					testAccCheckMDBClickHouseClusterHasResources(&r, "s2.micro", "network-ssd", 10737418240),
					testAccCheckMDBClickHouseClusterHasUsers(chResourceSharded, map[string][]string{"john": {"testdb"}}),
					testAccCheckMDBClickHouseClusterHasDatabases(chResourceSharded, []string{"testdb"}),
					testAccCheckCreatedAtAttr(chResourceSharded),
				),
			},
			mdbClickHouseClusterImportStep(chResourceSharded),
		},
	})
}

func testAccCheckMDBClickHouseClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_clickhouse_cluster" {
			continue
		}

		_, err := config.sdk.MDB().Clickhouse().Cluster().Get(context.Background(), &clickhouse.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("ClickHouse Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBClickHouseClusterExists(n string, r *clickhouse.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().Clickhouse().Cluster().Get(context.Background(), &clickhouse.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("ClickHouse Cluster not found")
		}

		*r = *found

		resp, err := config.sdk.MDB().Clickhouse().Cluster().ListHosts(context.Background(), &clickhouse.ListClusterHostsRequest{
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

func testAccCheckMDBClickHouseClusterHasResources(r *clickhouse.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Clickhouse.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("Expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("Expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("Expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

//nolint:unused
func testAccCheckMDBClickHouseClusterHasShards(r *clickhouse.Cluster, shards []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().Clickhouse().Cluster().ListShards(context.Background(), &clickhouse.ListClusterShardsRequest{
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

func testAccCheckMDBClickHouseZooKeeperSubclusterHasResources(r *clickhouse.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Zookeeper.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("Expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("Expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("Expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBClickHouseClusterHasUsers(r string, perms map[string][]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().Clickhouse().User().List(context.Background(), &clickhouse.ListUsersRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}
		users := resp.Users

		if len(users) != len(perms) {
			return fmt.Errorf("Expected %d users, found %d", len(perms), len(users))
		}

		for _, u := range users {
			ps, ok := perms[u.Name]
			if !ok {
				return fmt.Errorf("Unexpected user: %s", u.Name)
			}

			ups := []string{}
			for _, p := range u.Permissions {
				ups = append(ups, p.DatabaseName)
			}

			sort.Strings(ps)
			sort.Strings(ups)
			if fmt.Sprintf("%v", ps) != fmt.Sprintf("%v", ups) {
				return fmt.Errorf("User %s has wrong permissions, %v. Expected %v", u.Name, ups, ps)
			}
		}

		return nil
	}
}

func testAccCheckMDBClickHouseClusterHasDatabases(r string, databases []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().Clickhouse().Database().List(context.Background(), &clickhouse.ListDatabasesRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}
		dbs := []string{}
		for _, d := range resp.Databases {
			dbs = append(dbs, d.Name)
		}

		if len(dbs) != len(databases) {
			return fmt.Errorf("Expected %d dbs, found %d", len(databases), len(dbs))
		}

		sort.Strings(dbs)
		sort.Strings(databases)
		if fmt.Sprintf("%v", dbs) != fmt.Sprintf("%v", databases) {
			return fmt.Errorf("Cluster has wrong databases, %v. Expected %v", dbs, databases)
		}

		return nil
	}
}

func testAccCheckMDBClickHouseClusterContainsLabel(r *clickhouse.Cluster, key string, value string) resource.TestCheckFunc {
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

const clickHouseVPCDependencies = `
resource "yandex_vpc_network" "mdb-ch-test-net" {}

resource "yandex_vpc_subnet" "mdb-ch-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-ch-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-ch-test-subnet-c" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}
`

func testAccMDBClickHouseClusterConfigMain(name, desc string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+`
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-ch-test-net.id}"

  labels = {
    test_key = "test_value"
  }

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  database {
    name = "testdb"
  }

  user {
    name     = "john"
    password = "password"
    permission {
      database_name = "testdb"
    }
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
  }
}
`, name, desc)
}

func testAccMDBClickHouseClusterConfigUpdated(name, desc string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+`
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-ch-test-net.id}"

  labels = {
    new_key = "new_value"
  }

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 18
    }
  }

  database {
    name = "testdb"
  }

  database {
    name = "newdb"
  }

  user {
    name     = "john"
    password = "password"
    permission {
      database_name = "testdb"
    }
  }

  user {
    name     = "mary"
    password = "qwerty123"
    permission {
      database_name = "newdb"
    }
    permission {
      database_name = "testdb"
    }
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
  }
}
`, name, desc)
}

func testAccMDBClickHouseClusterConfigHA(name, desc string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+`
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-ch-test-net.id}"

  labels = {
    new_key = "new_value"
  }

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 18
    }
  }

  zookeeper {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  database {
    name = "testdb"
  }

  database {
    name = "newdb"
  }

  user {
    name     = "john"
    password = "password"
    permission {
      database_name = "testdb"
    }
  }

  user {
    name     = "mary"
    password = "qwerty123"
    permission {
      database_name = "newdb"
    }
    permission {
      database_name = "testdb"
    }
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-b"
    subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-b.id}"
  }

  host {
    type      = "ZOOKEEPER"
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
  }

  host {
    type      = "ZOOKEEPER"
    zone      = "ru-central1-b"
    subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-b.id}"
  }

  host {
    type      = "ZOOKEEPER"
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-c.id}"
  }
}
`, name, desc)
}

//nolint:unused
func testAccMDBClickHouseClusterConfigSharded(name, desc string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+`
resource "yandex_mdb_clickhouse_cluster" "bar" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-ch-test-net.id}"

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  database {
    name = "testdb"
  }

  user {
    name     = "john"
    password = "password"
    permission {
      database_name = "testdb"
    }
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-a"
    subnet_id  = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
	shard_name = "shard1"
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-b"
    subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-b.id}"
	shard_name = "shard2"
  }
}
`, name, desc)
}

//nolint:unused
func testAccMDBClickHouseClusterConfigShardedUpdated(name, desc string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+`
resource "yandex_mdb_clickhouse_cluster" "bar" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-ch-test-net.id}"

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  database {
    name = "testdb"
  }

  user {
    name     = "john"
    password = "password"
    permission {
      database_name = "testdb"
    }
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-a"
    subnet_id  = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
	shard_name = "shard1"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.mdb-ch-test-subnet-c.id}"
	shard_name = "shard3"
  }
}
`, name, desc)
}
