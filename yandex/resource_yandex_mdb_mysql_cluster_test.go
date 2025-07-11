package yandex

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"google.golang.org/genproto/protobuf/field_mask"

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

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.sdk.MDB().MySQL().Cluster().Update(ctx, &mysql.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.MDB().MySQL().Cluster().Delete(ctx, &mysql.DeleteClusterRequest{
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
			"database",                       // not returned
			"user",                           // not returned
			"health",                         // volatile value
			"host",                           // the order of hosts differs
			"allow_regeneration_host",        // Only state flag
			"host.0.name",                    // not returned
			"host.1.name",                    // not returned
			"host.2.name",                    // not returned
			"host.0.replication_source_name", // not returned
			"host.1.replication_source_name", // not returned
			"host.2.replication_source_name", // not returned
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
	var hostNames *[]string = new([]string)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBMysqlClusterDestroy,
		Steps: []resource.TestStep{
			// Create MySQL Cluster
			{
				Config: testAccMDBMySQLClusterConfigMain(mysqlName, mysqlDesc, "PRESTABLE", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),
					resource.TestCheckResourceAttr(mysqlResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(mysqlResource, "description", mysqlDesc),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.0.fqdn"),
					resource.TestCheckResourceAttr(mysqlResource, "host.0.assign_public_ip", "false"),
					testAccCheckMDBMySQLClusterHasDatabases(mysqlResource, []string{"testdb"}),
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]MockPermission{
						"john": {MockPermission{"testdb", []string{"ALL", "INSERT"}}}}),
					testAccCheckMDBMysqlClusterHasResources(&cluster, "s2.micro", "network-ssd", 17179869184),
					testAccCheckMDBMysqlClusterHasBackupWindow(&cluster, 3, 22),
					testAccCheckMDBMysqlClusterContainsLabel(&cluster, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(mysqlResource),
					testAccCheckMDBMysqlClusterHasHosts(mysqlResource, 1),
					resource.TestCheckResourceAttr(mysqlResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(mysqlResource, "deletion_protection", "true"),

					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.day", "SAT"),
					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.hour", "12"),

					resource.TestCheckResourceAttr(mysqlResource, "mysql_config.sql_mode", "ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"),
					resource.TestCheckResourceAttr(mysqlResource, "mysql_config.innodb_print_all_deadlocks", "true"),

					resource.TestCheckResourceAttr(mysqlResource, "backup_retain_period_days", "12"),

					testAccCheckMDBMysqlClusterSettingsPerformanceDiagnostics(mysqlResource, true, 300, 400),
					testAccMDBMysqlGetHostNames(mysqlResource, hostNames),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			// uncheck 'deletion_protection'
			{
				Config: testAccMDBMySQLClusterConfigMain(mysqlName, mysqlDesc, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "deletion_protection", "false"),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			// check 'deletion_protection'
			{
				Config: testAccMDBMySQLClusterConfigMain(mysqlName, mysqlDesc, "PRESTABLE", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "deletion_protection", "true"),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			// trigger deletion by changing environment
			{
				Config:      testAccMDBMySQLClusterConfigMain(mysqlName, mysqlDesc, "PRODUCTION", true),
				ExpectError: regexp.MustCompile(".*The operation was rejected because cluster has 'deletion_protection' = ON.*"),
			},
			// uncheck 'deletion_protection'
			{
				Config: testAccMDBMySQLClusterConfigMain(mysqlName, mysqlDesc, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "deletion_protection", "false"),
				),
			},
			// Change some options
			{
				Config: testAccMDBMySQLClusterVersionUpdate(mysqlName, mysqlDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),
					resource.TestCheckResourceAttr(mysqlResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(mysqlResource, "version", "8.0"),
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
					resource.TestCheckResourceAttr(mysqlResource, "host.0.assign_public_ip", "true"),
					testAccCheckMDBMySQLClusterHasDatabases(mysqlResource, []string{"testdb", "new_testdb"}),
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]MockPermission{
						"john": {MockPermission{"testdb", []string{"ALL", "DROP", "DELETE"}}},
						"mary": {MockPermission{"testdb", []string{"ALL", "INSERT"}}, MockPermission{"new_testdb", []string{"ALL", "INSERT"}}}}),
					testAccCheckMDBMysqlClusterHasResources(&cluster, "s2.micro", "network-ssd", 25769803776),
					testAccCheckMDBMysqlClusterHasBackupWindow(&cluster, 5, 44),
					testAccCheckMDBMysqlClusterContainsLabel(&cluster, "new_key", "new_value"),
					testAccCheckCreatedAtAttr(mysqlResource),
					testAccCheckMDBMysqlClusterHasHosts(mysqlResource, 1),
					resource.TestCheckResourceAttr(mysqlResource, "security_group_ids.#", "2"),

					resource.TestCheckResourceAttr(mysqlResource, "user.0.connection_limits.0.max_questions_per_hour", "10"),
					resource.TestCheckResourceAttr(mysqlResource, "user.0.global_permissions.#", "2"),
					resource.TestCheckResourceAttr(mysqlResource, "user.0.authentication_plugin", "SHA256_PASSWORD"),

					resource.TestCheckResourceAttr(mysqlResource, "access.0.web_sql", "true"),
					resource.TestCheckResourceAttr(mysqlResource, "access.0.data_lens", "true"),
					resource.TestCheckResourceAttr(mysqlResource, "access.0.data_transfer", "true"),
					resource.TestCheckResourceAttr(mysqlResource, "mysql_config.sql_mode", "IGNORE_SPACE,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,HIGH_NOT_PRECEDENCE"),
					resource.TestCheckResourceAttr(mysqlResource, "mysql_config.max_connections", "10"),
					resource.TestCheckResourceAttr(mysqlResource, "mysql_config.default_authentication_plugin", "MYSQL_NATIVE_PASSWORD"),
					resource.TestCheckResourceAttr(mysqlResource, "mysql_config.innodb_print_all_deadlocks", "true"),

					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.day", "WED"),
					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.hour", "22"),

					resource.TestCheckResourceAttr(mysqlResource, "backup_retain_period_days", "13"),
					testAccMDBMysqlCompareHostNames(mysqlResource, hostNames),
				),
			},
		},
	},
	)
}

func TestAccMDBMySQLClusterHA_update(t *testing.T) {
	t.Parallel()

	var cluster mysql.Cluster
	mysqlName := acctest.RandomWithPrefix("tf-mysql")
	var hostNames *[]string = new([]string)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBMysqlClusterDestroy,
		Steps: []resource.TestStep{
			//Add new host
			{
				Config: testAccMDBMysqlClusterHA(mysqlName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),

					testAccCheckMDBMysqlClusterHasHosts(mysqlResource, 3),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.1.fqdn"),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.2.fqdn"),
					resource.TestCheckResourceAttr(mysqlResource, "host.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(mysqlResource, "host.1.assign_public_ip", "true"),
					resource.TestCheckResourceAttr(mysqlResource, "host.2.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.type", "ANYTIME"),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			//Add new host 2 cc
			{
				Config: testAccMDBMysqlClusterHA2(mysqlName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),

					testAccCheckMDBMysqlClusterHasHosts(mysqlResource, 4),
					resource.TestCheckResourceAttr(mysqlResource, "host.0.zone", "ru-central1-a"),
					resource.TestCheckResourceAttr(mysqlResource, "host.1.zone", "ru-central1-b"),
					resource.TestCheckResourceAttr(mysqlResource, "host.2.zone", "ru-central1-d"),
					resource.TestCheckResourceAttr(mysqlResource, "host.3.zone", "ru-central1-d"),
					resource.TestCheckResourceAttr(mysqlResource, "host.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(mysqlResource, "host.1.assign_public_ip", "true"),
					resource.TestCheckResourceAttr(mysqlResource, "host.2.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(mysqlResource, "host.3.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.type", "ANYTIME"),
					testAccMDBMysqlGetHostNames(mysqlResource, hostNames),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			// Configure cascade replica
			{
				Config: testAccMDBMysqlClusterHANamedWithCascade(mysqlName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.0.replication_source"),
					resource.TestCheckResourceAttr(mysqlResource, "host.0.replication_source_name", "nb"),
					resource.TestCheckResourceAttrSet(mysqlResource, "host.1.replication_source"),
					resource.TestCheckResourceAttr(mysqlResource, "host.1.replication_source_name", "na"),
					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.type", "ANYTIME"),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			// Change public IP for 2 hosts
			{
				Config: testAccMDBMysqlClusterHANamedChangePublicIP(mysqlName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),
					resource.TestCheckResourceAttr(mysqlResource, "host.0.assign_public_ip", "true"),
					resource.TestCheckResourceAttr(mysqlResource, "host.1.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(mysqlResource, "host.2.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(mysqlResource, "host.3.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.type", "ANYTIME"),
					testAccMDBMysqlCompareHostNames(mysqlResource, hostNames),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			// Change backup priority for 2 hosts
			{
				Config: testAccMDBMysqlClusterWithBackupPriorities(mysqlName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),
					resource.TestCheckResourceAttr(mysqlResource, "host.0.backup_priority", "10"),
					resource.TestCheckResourceAttr(mysqlResource, "host.1.backup_priority", "5"),
					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.type", "ANYTIME"),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
			// Change host priority for 2 hosts
			{
				Config: testAccMDBMysqlClusterWithPriorities(mysqlName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResource, &cluster),
					resource.TestCheckResourceAttr(mysqlResource, "name", mysqlName),
					resource.TestCheckResourceAttr(mysqlResource, "host.0.priority", "10"),
					resource.TestCheckResourceAttr(mysqlResource, "host.1.priority", "5"),
					resource.TestCheckResourceAttr(mysqlResource, "maintenance_window.0.type", "ANYTIME"),
				),
			},
			mdbMysqlClusterImportStep(mysqlResource),
		},
	},
	)
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

func testAccCheckMDBMySQLClusterHasDatabases(resource string, databases []string) resource.TestCheckFunc {
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

func testAccCheckMDBMysqlClusterSettingsPerformanceDiagnostics(r string, enabled bool, sessionSamplingInterval int, statementSamplingInterval int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
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

		if found.Config.PerformanceDiagnostics.Enabled != enabled {
			return fmt.Errorf("Cluster.Config.PerformanceDiagnostics.Enabled must be %t, current %v",
				enabled, found.Config.PerformanceDiagnostics.Enabled)
		}

		if found.Config.PerformanceDiagnostics.SessionsSamplingInterval != int64(sessionSamplingInterval) {
			return fmt.Errorf("Cluster.Config.PerformanceDiagnostics.SessionsSamplingInterval must be %d, current %v",
				sessionSamplingInterval, found.Config.PerformanceDiagnostics.SessionsSamplingInterval)
		}

		if found.Config.PerformanceDiagnostics.StatementsSamplingInterval != int64(statementSamplingInterval) {
			return fmt.Errorf("Cluster.Config.PerformanceDiagnostics.SessionsSamplingInterval must be %d, current %v",
				statementSamplingInterval, found.Config.PerformanceDiagnostics.StatementsSamplingInterval)
		}

		return nil
	}
}

func testAccMDBMysqlGetHostNames(resource string, hostNames *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		hosts, err := listMysqlHosts(context.Background(), config, rs.Primary.ID)
		if err != nil {
			return err
		}

		for _, host := range hosts {
			*hostNames = append(*hostNames, host.Name)
		}

		return nil
	}
}

func testAccMDBMysqlCompareHostNames(resource string, oldHosts *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		currentHosts, err := listMysqlHosts(context.Background(), config, rs.Primary.ID)
		if err != nil {
			return err
		}

		oldMap := make(map[string]struct{})

		for _, host := range *oldHosts {
			oldMap[host] = struct{}{}
		}

		miss := 0
		for _, host := range currentHosts {
			if _, ok := oldMap[host.Name]; !ok {
				miss++
			}
		}

		if miss > len(currentHosts)-len(*oldHosts) {
			return fmt.Errorf("some MySQL host names changed")
		}

		return nil
	}
}

// TODO: add more zones when v2 platform becomes available.
const mysqlVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo_c" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_subnet" "foo_b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.4.0.0/24"]
}

resource "yandex_vpc_subnet" "foo_a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}

resource "yandex_vpc_security_group" "sg-x" {
  network_id     = yandex_vpc_network.foo.id
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_security_group" "sg-y" {
  network_id     = yandex_vpc_network.foo.id
  
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}
`

func testAccMDBMySQLClusterConfigMain(name, desc, environment string, deletionProtection bool) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "%s"
  network_id  = yandex_vpc_network.foo.id
  version     = "5.7"
  labels = {
    test_key = "test_value"
  }

  mysql_config = {
    sql_mode                      = "ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"
    innodb_print_all_deadlocks    = true
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

  maintenance_window {
	type = "WEEKLY"
	day  = "SAT"
	hour = 12
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
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo_c.id
  }

  security_group_ids = [yandex_vpc_security_group.sg-x.id]
  deletion_protection = %t

  performance_diagnostics {
    enabled                      = true
    sessions_sampling_interval   = 300
    statements_sampling_interval = 400
  }

  backup_retain_period_days = 12
}
`, name, desc, environment, deletionProtection)
}

func testAccMDBMySQLClusterVersionUpdate(name, desc string) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"
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
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo_c.id
  }

  security_group_ids = [yandex_vpc_security_group.sg-x.id]
}
`, name, desc)
}

func testAccMDBMySQLClusterConfigUpdated(name, desc string) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  labels = {
    new_key = "new_value"
  }
  
  maintenance_window {
	type = "WEEKLY"
    day  = "WED"
	hour = 22
  }
  
  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 24
  }

  mysql_config = {
    sql_mode                      = "IGNORE_SPACE,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,HIGH_NOT_PRECEDENCE"
    max_connections               = 10
    default_authentication_plugin = "MYSQL_NATIVE_PASSWORD"
    innodb_print_all_deadlocks    = true
  }

  access {
    web_sql = true
    data_lens = true
    data_transfer = true
  }

  backup_window_start {
    hours   = 5
    minutes = 44
  }
    
  backup_retain_period_days = 13

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

    connection_limits {
      max_questions_per_hour = 10
    }

    global_permissions = ["REPLICATION_SLAVE", "PROCESS"]

    authentication_plugin = "SHA256_PASSWORD"
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
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo_c.id
	assign_public_ip = true
  }

  security_group_ids = [yandex_vpc_security_group.sg-x.id, yandex_vpc_security_group.sg-y.id]
  deletion_protection = false
}
`, name, desc)
}

func testAccMDBMysqlClusterHABasic(name, hosts string) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "%s"
  description = "MySQL High Availability Cluster Terraform Test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 24
  }
 
  %s
}
`, name, hosts)
}

func testAccMDBMysqlClusterHA(name string) string {
	return testAccMDBMysqlClusterHABasic(name, `
  host {
	zone      = "ru-central1-a"
	subnet_id = yandex_vpc_subnet.foo_a.id
  }

  host {
    zone      = "ru-central1-b"
    subnet_id = yandex_vpc_subnet.foo_b.id
	assign_public_ip = true
  }

  host {
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo_c.id
  }

  maintenance_window {
	type = "ANYTIME"
  }
`)
}

func testAccMDBMysqlClusterHA2(name string) string {
	return testAccMDBMysqlClusterHABasic(name, `
  host {
	zone      = "ru-central1-a"
	subnet_id = yandex_vpc_subnet.foo_a.id
  }
	
  host {
    zone      = "ru-central1-b"
    subnet_id = yandex_vpc_subnet.foo_b.id
	assign_public_ip = true
  }

  host {
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo_c.id
  }

  host {
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo_c.id
  }

  maintenance_window {
	type = "ANYTIME"
  }
`)
}

func testAccMDBMysqlClusterHANamedWithCascade(name string) string {
	return testAccMDBMysqlClusterHABasic(name, `

  host {
	zone                    = "ru-central1-a"
	subnet_id               = yandex_vpc_subnet.foo_a.id
	name                    = "nd"
	replication_source_name = "nb"
  }

  host {
    zone      				= "ru-central1-b"
    subnet_id 				= yandex_vpc_subnet.foo_b.id
    name      				= "nc"
	replication_source_name = "na"
	assign_public_ip 		= true
  }

  host {
    zone             = "ru-central1-d"
    subnet_id        = yandex_vpc_subnet.foo_c.id
    name             = "na"
  }


  host {
    zone 		= "ru-central1-d"
    subnet_id   = yandex_vpc_subnet.foo_c.id
    name        = "nb"
  }

  maintenance_window {
	type = "ANYTIME"
  }
`)
}

func testAccMDBMysqlClusterHANamedChangePublicIP(name string) string {
	return testAccMDBMysqlClusterHABasic(name, `
  host {
	zone                    = "ru-central1-a"
	subnet_id               = yandex_vpc_subnet.foo_a.id
	name                    = "nd"
	replication_source_name = "nb"
	assign_public_ip 		= true
  }

  host {
	zone      				= "ru-central1-b"
	subnet_id 				= yandex_vpc_subnet.foo_b.id
	name      				= "nc"
	replication_source_name = "na"
  }

  host {
	zone             = "ru-central1-d"
	subnet_id        = yandex_vpc_subnet.foo_c.id
	name             = "na"
  }

  host {
	zone 		= "ru-central1-d"
	subnet_id   = yandex_vpc_subnet.foo_c.id
	name        = "nb"
  }

  maintenance_window {
	type = "ANYTIME"
  }
`)
}

func testAccMDBMysqlClusterWithBackupPriorities(name string) string {
	return testAccMDBMysqlClusterHABasic(name, `
  host {
	zone                    = "ru-central1-a"
	subnet_id               = yandex_vpc_subnet.foo_a.id
	name                    = "nd"
	replication_source_name = "nb"
	backup_priority 		= 10
  }

  host {
	zone      				= "ru-central1-b"
	subnet_id 				= yandex_vpc_subnet.foo_b.id
	name      				= "nc"
	replication_source_name = "na"
	assign_public_ip 		= true
	backup_priority 		= 5
  }

  host {
	zone             = "ru-central1-d"
	subnet_id        = yandex_vpc_subnet.foo_c.id
	name             = "na"
  }

  host {
	zone 		= "ru-central1-d"
	subnet_id   = yandex_vpc_subnet.foo_c.id
	name        = "nb"
  }

  maintenance_window {
	type = "ANYTIME"
  }
`)
}

func testAccMDBMysqlClusterWithPriorities(name string) string {
	return testAccMDBMysqlClusterHABasic(name, `
  host {
	zone                    = "ru-central1-a"
	subnet_id               = yandex_vpc_subnet.foo_a.id
	name                    = "nd"
	replication_source_name = "nb"
	priority 				= 10
  }

  host {
	zone      				= "ru-central1-b"
	subnet_id 				= yandex_vpc_subnet.foo_b.id
	name      				= "nc"
	replication_source_name = "na"
	assign_public_ip 		= true
	priority 				= 5
  }

  host {
	zone             = "ru-central1-d"
	subnet_id        = yandex_vpc_subnet.foo_c.id
	name             = "na"
  }

  host {
	zone 		= "ru-central1-d"
	subnet_id   = yandex_vpc_subnet.foo_c.id
	name        = "nb"
  }

  maintenance_window {
	type = "ANYTIME"
  }
`)
}
