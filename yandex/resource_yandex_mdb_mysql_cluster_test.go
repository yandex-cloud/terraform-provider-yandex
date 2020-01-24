package yandex

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
)

const mysqlResource = "yandex_mdb_mysql_cluster.foo"

func mdbMysqlClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"user",   // not returned
			"health", // volatile value
			"host",   // the order of hosts differs
		},
	}
}

// Test that a MySQL Cluster can be created, updated and destroyed
func TestAccMDBMySQLCluster_full(t *testing.T) {
	t.Parallel()

	var cluster mysql.Cluster
	mysqlName := acctest.RandomWithPrefix("tf-mysql")
	mysqlDesc := "MySQL Cluster Terraform Test"
	mysqlDesc2 := "MySQL Cluster Terraform Test Updated"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBMysqlClusterDestroy,
		Steps: []resource.TestStep{
			// Create MySQL Cluster
			{
				Config: testAccMDBMySQLClusterConfigMain(mysqlName, mysqlDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),
					resource.TestCheckResourceAttr(mysqlResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(mysqlResource, "description", mysqlDesc),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.0.fqdn"),
					testAccCheckMDBMysqlClusterHasDatabases(mysqlResource, []string{"testdb"}),
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]string{"john": {"testdb"}}),
					testAccCheckMDBMysqlClusterHasResources(&cluster, "s2.micro", "network-ssd", 17179869184),
					testAccCheckMDBMysqlClusterContainsLabel(&cluster, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(mysqlResource),
					testAccCheckMDBMysqlClusterHasHosts(mysqlResource, 1),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			// Change some options
			{
				Config: testAccMDBMySQLClusterConfigUpdated(mysqlName, mysqlDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),
					resource.TestCheckResourceAttr(mysqlResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(mysqlResource, "description", mysqlDesc2),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.0.fqdn"),
					testAccCheckMDBMysqlClusterHasDatabases(mysqlResource, []string{"testdb", "new_testdb"}),
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]string{"john": {"testdb"}, "mary": {"testdb", "new_testdb"}}),
					testAccCheckMDBMysqlClusterHasResources(&cluster, "s2.micro", "network-ssd", 25769803776),
					testAccCheckMDBMysqlClusterContainsLabel(&cluster, "new_key", "new_value"),
					testAccCheckCreatedAtAttr(mysqlResource),
					testAccCheckMDBMysqlClusterHasHosts(mysqlResource, 1),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			//Add new host
			{
				Config: testAccMDBMysqlClusterHA(mysqlName, mysqlDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),
					resource.TestCheckResourceAttr(mysqlResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(mysqlResource, "description", mysqlDesc2),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.1.fqdn"),
					testAccCheckMDBMysqlClusterHasDatabases(mysqlResource, []string{"testdb", "new_testdb"}),
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]string{"john": {"testdb"}, "mary": {"testdb", "new_testdb"}}),
					testAccCheckMDBMysqlClusterHasResources(&cluster, "s2.micro", "network-ssd", 25769803776),
					testAccCheckMDBMysqlClusterContainsLabel(&cluster, "new_key", "new_value"),
					testAccCheckCreatedAtAttr(mysqlResource),
					testAccCheckMDBMysqlClusterHasHosts(mysqlResource, 3),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
		},
	})
}

func testAccCheckMDBMysqlClusterDestroy(state *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "yandex_mdb_mysql_cluster" {
			continue
		}

		_, err := config.sdk.MDB().MySQL().Cluster().Get(context.Background(), &mysql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("MySQL Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBMySQLClusterExists(resource string, cluster *mysql.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().MySQL().Cluster().Get(context.Background(), &mysql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("MySQL Cluster not found")
		}

		*cluster = *found

		return nil
	}
}

func testAccCheckMDBMysqlClusterHasDatabases(resource string, databases []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().MySQL().Database().List(context.Background(), &mysql.ListDatabasesRequest{
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

func testAccCheckMDBMysqlClusterHasHosts(resource string, expectedHostCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().MySQL().Cluster().ListHosts(context.Background(), &mysql.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != expectedHostCount {
			return fmt.Errorf("Expected %d hosts, found %d", expectedHostCount, len(resp.Hosts))
		}

		return nil
	}
}

func testAccCheckMDBMysqlClusterHasUsers(resource string, perms map[string][]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().MySQL().User().List(context.Background(), &mysql.ListUsersRequest{
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

func testAccCheckMDBMysqlClusterHasResources(resource *mysql.Cluster, resourcePresetID, diskTypeID string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := resource.Config.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("Expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskTypeID {
			return fmt.Errorf("Expected disk type id '%d', got '%d'", diskSize, rs.DiskSize)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("Expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBMysqlClusterContainsLabel(resource *mysql.Cluster, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := resource.Labels[key]
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
const mysqlVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo_c" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_subnet" "foo_b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.4.0.0/24"]
}

resource "yandex_vpc_subnet" "foo_a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.5.0.0/24"]
}
`

func testAccMDBMySQLClusterConfigMain(name, desc string) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"
  version     = "5.7"
  labels = {
    test_key = "test_value"
  }

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
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
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.foo_c.id}"
  }
}
`, name, desc)
}

func testAccMDBMySQLClusterConfigUpdated(name, desc string) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"
  version     = "5.7"

  labels = {
    new_key = "new_value"
  }

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 24
  }

  database {
    name = "testdb"
  }

  database {
    name = "new_testdb"
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
    password = "password"

    permission {
      database_name = "testdb"
    }

    permission {
      database_name = "new_testdb"
    }
  }

  host {
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.foo_c.id}"
  }
}
`, name, desc)
}

func testAccMDBMysqlClusterHA(name, desc string) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"
  version     = "5.7"

  labels = {
    new_key = "new_value"
  }

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 24
  }

  database {
    name = "testdb"
  }

  database {
    name = "new_testdb"
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
    password = "password"

    permission {
      database_name = "testdb"
    }

    permission {
      database_name = "new_testdb"
    }
  }

  host {
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.foo_c.id}"
  }

  host {
    zone      = "ru-central1-b"
    subnet_id = "${yandex_vpc_subnet.foo_b.id}"
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo_a.id}"
  }
}
`, name, desc)
}
