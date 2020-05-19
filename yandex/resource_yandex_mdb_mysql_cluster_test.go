package yandex

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

const mysqlResource = "yandex_mdb_mysql_cluster.foo"

func init() {
	resource.AddTestSweepers("yandex_mdb_mysql_cluster", &resource.Sweeper{
		Name: "yandex_mdb_mysql_cluster",
		F:    testSweepMDBMySQLCluster,
	})
}

func testSweepMDBMySQLCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().MySQL().Cluster().List(conf.Context(), &mysql.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting MySQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBMysqlCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep MySQL cluster %q", c.Id))
		} else {
			if !sweepVPCNetwork(conf, c.NetworkId) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep VPC network %q", c.NetworkId))
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBMysqlCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBMysqlClusterOnce, conf, "MySQL cluster", id)
}

func sweepMDBMysqlClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexMDBMySQLClusterDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.MDB().MySQL().Cluster().Delete(ctx, &mysql.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

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

type MockPermission struct {
	DatabaseName string
	Roles        []string
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
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]MockPermission{
						"john": {MockPermission{"testdb", []string{"ALL", "INSERT"}}}}),
					testAccCheckMDBMysqlClusterHasResources(&cluster, "s2.micro", "network-ssd", 17179869184),
					testAccCheckMDBMysqlClusterHasBackupWindow(&cluster, 3, 22),
					testAccCheckMDBMysqlClusterContainsLabel(&cluster, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(mysqlResource),
					testAccCheckMDBMysqlClusterHasHosts(mysqlResource, 1),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			{
				Config: testAccMDBMySQLClusterConfigDisallowedUpdatePublicIP(mysqlName, mysqlDesc),
				ExpectError: regexp.MustCompile("forbidden to change assign_public_ip setting for existing host .* " +
					"in resource_yandex_mdb_mysql_cluster, if you really need it you should delete one host and add another"),
			},
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
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]MockPermission{
						"john": {MockPermission{"testdb", []string{"ALL", "DROP", "DELETE"}}},
						"mary": {MockPermission{"testdb", []string{"ALL", "INSERT"}}, MockPermission{"new_testdb", []string{"ALL", "INSERT"}}}}),
					testAccCheckMDBMysqlClusterHasResources(&cluster, "s2.micro", "network-ssd", 25769803776),
					testAccCheckMDBMysqlClusterHasBackupWindow(&cluster, 5, 44),
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
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]MockPermission{
						"john": {MockPermission{"testdb", []string{"ALL", "DROP", "DELETE"}}},
						"mary": {MockPermission{"testdb", []string{"ALL", "INSERT"}}, MockPermission{"new_testdb", []string{"ALL", "INSERT"}}}}),
					testAccCheckMDBMysqlClusterHasResources(&cluster, "s2.micro", "network-ssd", 25769803776),
					testAccCheckMDBMysqlClusterHasBackupWindow(&cluster, 5, 44),
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

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "yandex_vpc_network" {
			continue
		}

		_, err := config.sdk.VPC().Network().Get(context.Background(), &vpc.GetNetworkRequest{
			NetworkId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Network still exists")
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

func testAccCheckMDBMysqlClusterHasUsers(resource string, perms map[string][]MockPermission) resource.TestCheckFunc {
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

			databases := []string{}
			for _, permission := range u.Permissions {
				databases = append(databases, permission.DatabaseName)
				err = checkRoles(u.Name, permission, ps)
				if err != nil {
					return err
				}
			}

			expectedDatabases := []string{}
			for _, permission := range ps {
				expectedDatabases = append(expectedDatabases, permission.DatabaseName)
			}

			sort.Strings(expectedDatabases)
			sort.Strings(databases)
			if fmt.Sprintf("%v", expectedDatabases) != fmt.Sprintf("%v", databases) {
				return fmt.Errorf("User %s has wrong permissions, %v. Expected %v", u.Name, databases, expectedDatabases)
			}
		}

		return nil
	}
}

func checkRoles(name string, permission *mysql.Permission, expectedPermissions []MockPermission) error {
	for _, expectedPermission := range expectedPermissions {
		if permission.DatabaseName != expectedPermission.DatabaseName {
			continue
		}
		roles := permission.Roles
		expectedRoles, err := bindDatabaseRoles(expectedPermission.Roles)
		if err != nil {
			return err
		}
		if fmt.Sprintf("%v", roles) != fmt.Sprintf("%v", expectedRoles) {
			return fmt.Errorf("User %s has wrong permissions, wrong roles, %v. Expected %v", name, roles, expectedRoles)
		}
	}
	return nil
}

func testAccCheckMDBMysqlClusterHasBackupWindow(resource *mysql.Cluster, hours, minutes int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		window := resource.Config.BackupWindowStart
		if window == nil {
			return fmt.Errorf("Missing backup_window_start for '%s'", resource.Id)
		}
		if window.Hours != hours {
			return fmt.Errorf("Expected backup_window_start hours '%d', got '%d'", hours, window.Hours)
		}
		if window.Minutes != minutes {
			return fmt.Errorf("Expected backup_window_start minutes '%d', got '%d'", minutes, window.Minutes)
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

  backup_window_start {
    hours   = 3
    minutes = 22
  }

  database {
    name = "testdb"
  }

  user {
    name     = "john"
    password = "password"
    permission {
      database_name = "testdb"
      roles         = ["ALL", "INSERT"]
    }
  }

  host {
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.foo_c.id}"
  }
}
`, name, desc)
}

func testAccMDBMySQLClusterConfigDisallowedUpdatePublicIP(name, desc string) string {
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

  backup_window_start {
    hours   = 3
    minutes = 22
  }

  database {
    name = "testdb"
  }

  user {
    name     = "john"
    password = "password"
    permission {
      database_name = "testdb"
      roles         = ["ALL", "INSERT"]
    }
  }

  host {
    zone             = "ru-central1-c"
    subnet_id        = "${yandex_vpc_subnet.foo_c.id}"
    assign_public_ip = true
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

  backup_window_start {
    hours   = 5
    minutes = 44
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
      roles         = ["ALL", "DROP", "DELETE"]
    }
  }

  user {
    name     = "mary"
    password = "password"

    permission {
      database_name = "testdb"
      roles         = ["ALL", "INSERT"]
    }

    permission {
      database_name = "new_testdb"
      roles         = ["ALL", "INSERT"]
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

  backup_window_start {
    hours   = 5
    minutes = 44
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
      roles         = ["ALL", "DROP", "DELETE"]
    }
  }

  user {
    name     = "mary"
    password = "password"

    permission {
      database_name = "testdb"
      roles         = ["ALL", "INSERT"]
    }

    permission {
      database_name = "new_testdb"
      roles         = ["ALL", "INSERT"]
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
