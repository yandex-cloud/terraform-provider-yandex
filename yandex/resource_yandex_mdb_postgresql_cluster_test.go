package yandex

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

const pgResource = "yandex_mdb_postgresql_cluster.foo"

func init() {
	resource.AddTestSweepers("yandex_mdb_postgresql_cluster", &resource.Sweeper{
		Name: "yandex_mdb_postgresql_cluster",
		F:    testSweepMDBPostgreSQLCluster,
	})
}

func testSweepMDBPostgreSQLCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().PostgreSQL().Cluster().List(conf.Context(), &postgresql.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting PostgreSQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBPostgreSQLCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep PostgreSQL cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBPostgreSQLCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBPostgreSQLClusterOnce, conf, "PostgreSQL cluster", id)
}

func sweepMDBPostgreSQLClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexMDBPostgreSQLClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.sdk.MDB().PostgreSQL().Cluster().Update(ctx, &postgresql.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.MDB().PostgreSQL().Cluster().Delete(ctx, &postgresql.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func mdbPGClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"user",                           // passwords are not returned
			"health",                         // volatile value
			"host.0.name",                    // not returned
			"host.1.name",                    // not returned
			"host.2.name",                    // not returned
			"host.1.replication_source_name", // not returned
			"host_master_name",               // not returned
		},
	}
}

// Test that a PostgreSQL Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_full(t *testing.T) {
	t.Parallel()

	var cluster postgresql.Cluster
	pgName := acctest.RandomWithPrefix("tf-postgresql")
	pgDesc := "PostgreSQL Cluster Terraform Test"
	pgDesc2 := "PostgreSQL Cluster Terraform Test Updated"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			//Create PostgreSQL Cluster
			{
				Config: testAccMDBPGClusterConfigMain(pgName, pgDesc, "PRESTABLE", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResource, &cluster, 1),
					resource.TestCheckResourceAttr(pgResource, "name", pgName),
					resource.TestCheckResourceAttr(pgResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(pgResource, "description", pgDesc),
					resource.TestCheckResourceAttr(pgResource, "database.0.lc_collate", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(pgResource, "database.0.lc_type", "en_US.UTF-8"),
					resource.TestCheckResourceAttrSet(pgResource, "host.0.fqdn"),
					testAccCheckMDBPGClusterContainsLabel(&cluster, "test_key", "test_value"),
					testAccCheckMDBPGClusterHasResources(&cluster, "s2.micro", "network-ssd", 10737418240),
					testAccCheckMDBPGClusterHasUsers(pgResource, map[string][]string{"alice": {"testdb"}}),
					testAccCheckMDBPGClusterHasDatabases(pgResource, []string{"testdb"}),
					testAccCheckCreatedAtAttr(pgResource),
					resource.TestCheckResourceAttr(pgResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(pgResource, "maintenance_window.0.day", "SAT"),
					resource.TestCheckResourceAttr(pgResource, "maintenance_window.0.hour", "12"),
					resource.TestCheckResourceAttr(pgResource, "deletion_protection", "true"),
				),
			},
			mdbPGClusterImportStep(pgResource),
			// uncheck 'deletion_protection'
			{
				Config: testAccMDBPGClusterConfigMain(pgName, pgDesc, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResource, &cluster, 1),
					resource.TestCheckResourceAttr(pgResource, "deletion_protection", "false"),
				),
			},
			mdbPGClusterImportStep(pgResource),
			// check 'deletion_protection'
			{
				Config: testAccMDBPGClusterConfigMain(pgName, pgDesc, "PRESTABLE", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResource, &cluster, 1),
					resource.TestCheckResourceAttr(pgResource, "deletion_protection", "true"),
				),
			},
			mdbPGClusterImportStep(pgResource),
			// trigger deletion by changing environment
			{
				Config:      testAccMDBPGClusterConfigMain(pgName, pgDesc, "PRODUCTION", true),
				ExpectError: regexp.MustCompile(".*The operation was rejected because cluster has 'deletion_protection' = ON.*"),
			},
			// uncheck 'deletion_protection'
			{
				Config: testAccMDBPGClusterConfigMain(pgName, pgDesc, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResource, &cluster, 1),
					resource.TestCheckResourceAttr(pgResource, "deletion_protection", "false"),
				),
			},
			{
				Config:      testAccMDBPGClusterConfigDisallowedUpdateLocale(pgName, pgDesc),
				ExpectError: regexp.MustCompile("impossible to change lc_collate or lc_type for PostgreSQL Cluster database .*"),
			},
			{
				Config:      testAccMDBPGClusterConfigDisallowedUpdateOwner(pgName, pgDesc),
				ExpectError: regexp.MustCompile("impossible to change owner for PostgreSQL Cluster database .*"),
			},
			// Change some options
			{
				Config: testAccMDBPGClusterConfigUpdated(pgName, pgDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResource, &cluster, 1),
					resource.TestCheckResourceAttr(pgResource, "name", pgName),
					resource.TestCheckResourceAttr(pgResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(pgResource, "description", pgDesc2),
					resource.TestCheckResourceAttrSet(pgResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(pgResource, "config.0.access.0.web_sql"),
					resource.TestCheckResourceAttrSet(pgResource, "config.0.access.0.serverless"),
					testAccCheckMDBPGClusterContainsLabel(&cluster, "new_key", "new_value"),
					testAccCheckMDBPGClusterHasResources(&cluster, "s2.micro", "network-ssd", 19327352832),
					testAccCheckMDBPGClusterHasPoolerConfig(&cluster, "TRANSACTION", false),
					testAccCheckMDBPGClusterHasUsers(pgResource, map[string][]string{"alice": {"testdb", "newdb"}, "bob": {"newdb", "fornewuserdb"}}),
					testAccCheckClusterSettingsAccessWebSQL(pgResource),
					testAccCheckClusterSettingsPerformanceDiagnostics(pgResource),
					testAccCheckConnLimitUpdateUserSettings(pgResource),
					testAccCheckMDBPGClusterHasDatabases(pgResource, []string{"testdb", "newdb", "fornewuserdb"}),
					testAccCheckSettingsUpdateUserSettings(pgResource),
					testAccCheckPostgresqlConfigUpdate(pgResource),
					testAccCheckCreatedAtAttr(pgResource),
					resource.TestCheckResourceAttr(pgResource, "security_group_ids.#", "2"),

					resource.TestCheckResourceAttr(pgResource, "maintenance_window.0.day", "WED"),
					resource.TestCheckResourceAttr(pgResource, "maintenance_window.0.hour", "22"),
					resource.TestCheckResourceAttr(pgResource, "config.0.backup_retain_period_days", "12"),
				),
			},
			mdbPGClusterImportStep(pgResource),
			//Add host
			{
				Config: testAccMDBPGClusterConfigHA(pgName, pgDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResource, &cluster, 3),
					resource.TestCheckResourceAttr(pgResource, "name", pgName),
					resource.TestCheckResourceAttr(pgResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(pgResource, "description", pgDesc2),
					resource.TestCheckResourceAttrSet(pgResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(pgResource, "host.1.fqdn"),
					testAccCheckMDBPGClusterContainsLabel(&cluster, "new_key", "new_value"),
					testAccCheckMDBPGClusterHasResources(&cluster, "s2.micro", "network-ssd", 19327352832),
					testAccCheckMDBPGClusterHasPoolerConfig(&cluster, "TRANSACTION", false),
					testAccCheckMDBPGClusterHasUsers(pgResource, map[string][]string{"alice": {"testdb", "newdb"}, "bob": {"newdb", "fornewuserdb"}}),
					testAccCheckMDBPGClusterHasDatabases(pgResource, []string{"testdb", "newdb", "fornewuserdb"}),
					resource.TestCheckResourceAttr(pgResource, "host.2.zone", "ru-central1-c"),
					resource.TestCheckResourceAttr(pgResource, "host.1.zone", "ru-central1-b"),
					testAccCheckCreatedAtAttr(pgResource),
					resource.TestCheckResourceAttr(pgResource, "security_group_ids.#", "1"),

					resource.TestCheckResourceAttr(pgResource, "maintenance_window.0.type", "ANYTIME"),
				),
			},
			mdbPGClusterImportStep(pgResource),
			//Add named host
			{
				Config: testAccMDBPGClusterConfigHANamed(pgName, pgDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResource, &cluster, 3),
					resource.TestCheckResourceAttr(pgResource, "name", pgName),
					resource.TestCheckResourceAttr(pgResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(pgResource, "description", pgDesc2),
					resource.TestCheckResourceAttrSet(pgResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(pgResource, "host.1.fqdn"),
					testAccCheckMDBPGClusterContainsLabel(&cluster, "new_key", "new_value"),
					testAccCheckMDBPGClusterHasResources(&cluster, "s2.micro", "network-ssd", 19327352832),
					testAccCheckMDBPGClusterHasPoolerConfig(&cluster, "TRANSACTION", false),
					testAccCheckMDBPGClusterHasUsers(pgResource, map[string][]string{"alice": {"testdb", "newdb"}, "bob": {"newdb", "fornewuserdb"}}),
					testAccCheckMDBPGClusterHasDatabases(pgResource, []string{"testdb", "newdb", "fornewuserdb"}),
					resource.TestCheckResourceAttr(pgResource, "host.0.assign_public_ip", "true"),
					resource.TestCheckResourceAttr(pgResource, "host.1.zone", "ru-central1-b"),
					resource.TestCheckResourceAttrSet(pgResource, "host.1.replication_source"),
					resource.TestCheckResourceAttr(pgResource, "host.2.priority", "2"),
					testAccCheckCreatedAtAttr(pgResource),
					resource.TestCheckResourceAttr(pgResource, "security_group_ids.#", "1"),
				),
			},
			mdbPGClusterImportStep(pgResource),
			// change some options
			{
				Config: testAccMDBPGClusterConfigHANamedUpdated(pgName, pgDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResource, &cluster, 3),
					resource.TestCheckResourceAttr(pgResource, "host.0.zone", "ru-central1-a"),
					resource.TestCheckResourceAttr(pgResource, "host.1.zone", "ru-central1-b"),
					resource.TestCheckResourceAttr(pgResource, "host.2.zone", "ru-central1-c"),
					resource.TestCheckResourceAttr(pgResource, "host.0.priority", "1"),
					resource.TestCheckResourceAttrSet(pgResource, "host.1.replication_source"),
					resource.TestCheckResourceAttr(pgResource, "host.2.assign_public_ip", "true"),
				),
			},
		},
	})
}

func testAccCheckMDBPGClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_postgresql_cluster" {
			continue
		}

		_, err := config.sdk.MDB().PostgreSQL().Cluster().Get(context.Background(), &postgresql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("PostgreSQL Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBPGClusterExists(n string, r *postgresql.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().PostgreSQL().Cluster().Get(context.Background(), &postgresql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("PostgreSQL Cluster not found")
		}

		*r = *found

		resp, err := config.sdk.MDB().PostgreSQL().Cluster().ListHosts(context.Background(), &postgresql.ListClusterHostsRequest{
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

func testAccCheckMDBPGClusterHasResources(r *postgresql.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Resources
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

func testAccCheckMDBPGClusterHasPoolerConfig(r *postgresql.Cluster, poolingMode string, poolDiscard bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.PoolerConfig
		if rs == nil {
			return fmt.Errorf("Expected pooling mode %v, pool discard %v, got empty pooler config", poolingMode, poolDiscard)
		}

		if rs.PoolDiscard.GetValue() != poolDiscard {
			return fmt.Errorf("Expected pool discard %v, got %v", poolDiscard, rs.PoolDiscard.GetValue())
		}

		if rs.PoolingMode.String() != poolingMode {
			return fmt.Errorf("Expected pooling mode %v, got %v", poolingMode, rs.PoolingMode.String())
		}

		return nil
	}
}

func testAccCheckMDBPGClusterHasUsers(r string, perms map[string][]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().PostgreSQL().User().List(context.Background(), &postgresql.ListUsersRequest{
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

func testAccCheckClusterSettingsAccessWebSQL(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().PostgreSQL().Cluster().Get(context.Background(), &postgresql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if !found.Config.Access.WebSql {
			return fmt.Errorf("Cluster Config.Access.WebSql must be enabled, current %v",
				found.Config.Access.WebSql)
		}

		return nil
	}
}

func testAccCheckClusterSettingsPerformanceDiagnostics(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().PostgreSQL().Cluster().Get(context.Background(), &postgresql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Config.PerformanceDiagnostics.SessionsSamplingInterval != 9 {
			return fmt.Errorf("Cluster Config.PerformanceDiagnostics.SessionsSamplingInterval must be 9, current %v",
				found.Config.PerformanceDiagnostics.SessionsSamplingInterval)
		}

		if found.Config.PerformanceDiagnostics.StatementsSamplingInterval != 8 {
			return fmt.Errorf("Cluster Config.PerformanceDiagnostics.SessionsSamplingInterval must be 8, current %v",
				found.Config.PerformanceDiagnostics.StatementsSamplingInterval)
		}

		return nil
	}
}

var defaultUserSettings = map[string]interface{}{
	"conn_limit": int64(50),
}
var testAccMDBPGClusterConfigUpdatedCheckConnLimitMap = map[string]int64{
	"alice": 42,
}

func testAccCheckConnLimitUpdateUserSettings(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().PostgreSQL().User().List(context.Background(), &postgresql.ListUsersRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		defaultConnLimit := defaultUserSettings["conn_limit"].(int64)
		for _, user := range resp.Users {
			v, ok := testAccMDBPGClusterConfigUpdatedCheckConnLimitMap[user.Name]
			if ok {
				if user.ConnLimit != v {
					return fmt.Errorf("Field 'conn_limit' wasn`t changed for user %s with value %d ",
						user.Name, user.ConnLimit)
				}
			} else if user.ConnLimit != defaultConnLimit {
				return fmt.Errorf("Unmodified field 'conn_limit' was changed for user %s with value %d ",
					user.Name, user.ConnLimit)
			}
		}
		return nil
	}
}

var defaultTransactionIsolationPerUser = map[string]postgresql.UserSettings_TransactionIsolation{
	"alice": postgresql.UserSettings_TRANSACTION_ISOLATION_READ_UNCOMMITTED,
	"bob":   postgresql.UserSettings_TRANSACTION_ISOLATION_READ_COMMITTED,
}

func testAccCheckSettingsUpdateUserSettings(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().PostgreSQL().User().List(context.Background(), &postgresql.ListUsersRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		for _, user := range resp.Users {
			v, ok := defaultTransactionIsolationPerUser[user.Name]
			if ok {
				if user.Settings.DefaultTransactionIsolation != v {
					return fmt.Errorf("Field 'settings.default_transaction_isolation' wasn`t changed for user %s with value %d ",
						user.Name, user.Settings.DefaultTransactionIsolation)
				}
				if user.Settings.LogMinDurationStatement.GetValue() != 5000 {
					return fmt.Errorf("Field 'settings.log_min_duration_statement' wasn`t changed for user %s with value %d ",
						user.Name, user.Settings.LogMinDurationStatement.GetValue())
				}
			}
		}
		return nil
	}
}

func testAccCheckPostgresqlConfigUpdate(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		cluster, err := config.sdk.MDB().PostgreSQL().Cluster().Get(context.Background(), &postgresql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if cluster.Config.GetPostgresqlConfig_12().UserConfig.MaxConnections.GetValue() != 395 {
			return fmt.Errorf("Field 'config.postgresql_config.max_connections' wasn`t changed for with value 395. Current value is %v",
				cluster.Config.GetPostgresqlConfig_12().UserConfig.MaxConnections.GetValue())
		}

		if !cluster.Config.GetPostgresqlConfig_12().UserConfig.EnableParallelHash.GetValue() {
			return fmt.Errorf("Field 'config.postgresql_config.enable_parallel_hash' wasn`t changed for with value true. Current value is %v",
				cluster.Config.GetPostgresqlConfig_12().UserConfig.EnableParallelHash.GetValue())
		}

		if cluster.Config.GetPostgresqlConfig_12().UserConfig.VacuumCleanupIndexScaleFactor.GetValue() != 0.2 {
			return fmt.Errorf("Field 'config.postgresql_config.vacuum_cleanup_index_scale_factor' wasn`t changed for with value 0.2. Current value is %v",
				cluster.Config.GetPostgresqlConfig_12().UserConfig.VacuumCleanupIndexScaleFactor.GetValue())
		}

		if cluster.Config.GetPostgresqlConfig_12().UserConfig.DefaultTransactionIsolation != 1 {
			return fmt.Errorf("Field 'config.postgresql_config.default_transaction_isolation' wasn`t changed for with value 1. Current value is %v",
				cluster.Config.GetPostgresqlConfig_12().UserConfig.DefaultTransactionIsolation)
		}

		return nil
	}
}

func testAccCheckMDBPGClusterHasDatabases(r string, databases []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().PostgreSQL().Database().List(context.Background(), &postgresql.ListDatabasesRequest{
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

func testAccCheckMDBPGClusterContainsLabel(r *postgresql.Cluster, key string, value string) resource.TestCheckFunc {
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

const pgVPCDependencies = `
resource "yandex_vpc_network" "mdb-pg-test-net" {}

resource "yandex_vpc_subnet" "mdb-pg-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.mdb-pg-test-net.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-pg-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.mdb-pg-test-net.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-pg-test-subnet-c" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.mdb-pg-test-net.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_security_group" "mdb-pg-test-sg-x" {
  network_id     = "${yandex_vpc_network.mdb-pg-test-net.id}"
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

resource "yandex_vpc_security_group" "mdb-pg-test-sg-y" {
  network_id     = "${yandex_vpc_network.mdb-pg-test-net.id}"
  
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

func testAccMDBPGClusterConfigMain(name, desc, environment string, deletionProtection bool) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "%s"
  network_id  = "${yandex_vpc_network.mdb-pg-test-net.id}"

  labels = {
    test_key = "test_value"
  }

  maintenance_window {
	type = "WEEKLY"
	day  = "SAT"
	hour = 12
  }

  config {
    version = 12

    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
    }
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
    }
  }

  host {
	zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.mdb-pg-test-subnet-a.id}"
  }

  database {
    owner      = "alice"
    name       = "testdb"
    lc_collate = "en_US.UTF-8"
    lc_type    = "en_US.UTF-8"
  }

  security_group_ids = ["${yandex_vpc_security_group.mdb-pg-test-sg-x.id}"]
  deletion_protection = %t
}
`, name, desc, environment, deletionProtection)
}

func testAccMDBPGClusterConfigDisallowedUpdateLocale(name, desc string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-pg-test-net.id}"

  labels = {
    test_key = "test_value"
  }

  config {
    version = 12

    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
    }
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
    }
  }

  host {
	zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.mdb-pg-test-subnet-a.id}"
  }

  database {
    owner      = "alice"
    name       = "testdb"
    lc_collate = "C"
    lc_type    = "en_US.UTF-8"
  }

  security_group_ids = ["${yandex_vpc_security_group.mdb-pg-test-sg-x.id}"]
}
`, name, desc)
}

func testAccMDBPGClusterConfigDisallowedUpdateOwner(name, desc string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-pg-test-net.id}"

  labels = {
    test_key = "test_value"
  }

  config {
    version = 12

    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
    }
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
    }
  }

  user {
    name     = "bob"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
    }
  }

  host {
	zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.mdb-pg-test-subnet-a.id}"
  }

  database {
    owner      = "bob"
    name       = "testdb"
    lc_collate = "en_US.UTF-8"
    lc_type    = "en_US.UTF-8"
  }

  security_group_ids = ["${yandex_vpc_security_group.mdb-pg-test-sg-x.id}"]
}
`, name, desc)
}

func testAccMDBPGClusterConfigUpdated(name, desc string) string {

	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-pg-test-net.id}"

  labels = {
    new_key = "new_value"
  }

  maintenance_window {
	type = "WEEKLY"
    day  = "WED"
	hour = 22
  }

  config {
    version = 12

    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 18
      disk_type_id       = "network-ssd"
    }
    access {
      web_sql    = true
      serverless = true
    }
    performance_diagnostics {
      sessions_sampling_interval   = 9
      statements_sampling_interval = 8
    }
    
    backup_retain_period_days = 12
    
    pooler_config {
      pooling_mode = "TRANSACTION"
      pool_discard = false
    }

    postgresql_config = {
      max_connections                   = 395
      enable_parallel_hash              = true
      vacuum_cleanup_index_scale_factor = 0.2
      autovacuum_vacuum_scale_factor    = 0.34
      default_transaction_isolation     = "TRANSACTION_ISOLATION_READ_UNCOMMITTED"
    }
  }

  user {
    name       = "alice"
    password   = "mysecurepassword"
    conn_limit = 42

    permission {
      database_name = "testdb"
    }

    permission {
      database_name = "newdb"
    }

    settings = {
      default_transaction_isolation = "read uncommitted"
      log_min_duration_statement    = 5000
    }
  }

  user {
    name     = "bob"
    password = "anothersecurepassword"

    permission {
      database_name = "newdb"
    }

    permission {
      database_name = "fornewuserdb"
    }

    settings = {
	  default_transaction_isolation = "read committed"
	  log_min_duration_statement    = 5000
    }
  }

  host {
	zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.mdb-pg-test-subnet-a.id}"
  }

  database {
    owner      = "alice"
    name       = "testdb"
    lc_collate = "en_US.UTF-8"
    lc_type    = "en_US.UTF-8"
  }

  database {
    owner = "alice"
    name  = "newdb"
  }

  database {
    owner = "bob"
    name  = "fornewuserdb"
  }

  security_group_ids = ["${yandex_vpc_security_group.mdb-pg-test-sg-x.id}", "${yandex_vpc_security_group.mdb-pg-test-sg-y.id}"]
}
`, name, desc)
}

func testAccMDBPGClusterConfigHA(name, desc string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-pg-test-net.id}"

  labels = {
    new_key = "new_value"
  }

  maintenance_window {
    type = "ANYTIME"
  }

  config {
    version = 12

    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 18
      disk_type_id       = "network-ssd"
    }

    pooler_config {
      pooling_mode = "TRANSACTION"
      pool_discard = false
    }
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
    }

    permission {
      database_name = "newdb"
    }
  }

  user {
    name     = "bob"
    password = "anothersecurepassword"

    permission {
      database_name = "newdb"
    }

    permission {
      database_name = "fornewuserdb"
    }

    settings = {
      default_transaction_isolation = "read uncommitted"
      log_min_duration_statement    = -1
    }
  }

  host {
	zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.mdb-pg-test-subnet-a.id}"
  }
  host {
	zone                    = "ru-central1-b"
    subnet_id               = "${yandex_vpc_subnet.mdb-pg-test-subnet-b.id}"
  }
  host {
	zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.mdb-pg-test-subnet-c.id}"
  }

  database {
    owner      = "alice"
    name       = "testdb"
    lc_collate = "en_US.UTF-8"
    lc_type    = "en_US.UTF-8"
  }

  database {
    owner = "alice"
    name  = "newdb"
  }

   database {
    owner = "bob"
    name  = "fornewuserdb"
  }

  security_group_ids = ["${yandex_vpc_security_group.mdb-pg-test-sg-y.id}"]
}
`, name, desc)
}

func testAccMDBPGClusterConfigHANamed(name, desc string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-pg-test-net.id}"

  labels = {
    new_key = "new_value"
  }

  config {
    version = 12

    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 18
      disk_type_id       = "network-ssd"
    }

    pooler_config {
      pooling_mode = "TRANSACTION"
      pool_discard = false
    }
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
    }

    permission {
      database_name = "newdb"
    }
  }

  user {
    name     = "bob"
    password = "anothersecurepassword"

    permission {
      database_name = "newdb"
    }

    permission {
      database_name = "fornewuserdb"
    }
  }

  host {
    zone      		 = "ru-central1-a"
    name      		 = "na"
    subnet_id 		 = "${yandex_vpc_subnet.mdb-pg-test-subnet-a.id}"
    assign_public_ip = true
  }
  host {
    zone                    = "ru-central1-b"
    name                    = "nb"
    replication_source_name = "nc"
    subnet_id               = "${yandex_vpc_subnet.mdb-pg-test-subnet-b.id}"
  }
  host {
    zone      = "ru-central1-c"
    name      = "nc"
    priority  = 2
    subnet_id = "${yandex_vpc_subnet.mdb-pg-test-subnet-c.id}"
  }

  database {
    owner      = "alice"
    name       = "testdb"
    lc_collate = "en_US.UTF-8"
    lc_type    = "en_US.UTF-8"
  }

  database {
    owner = "alice"
    name  = "newdb"
  }

  database {
    owner = "bob"
    name  = "fornewuserdb"
  }

  security_group_ids = ["${yandex_vpc_security_group.mdb-pg-test-sg-y.id}"]
}
`, name, desc)
}

func testAccMDBPGClusterConfigHANamedUpdated(name, desc string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-pg-test-net.id}"

  labels = {
    new_key = "new_value"
  }

  config {
    version = 12

    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 18
      disk_type_id       = "network-ssd"
    }

    pooler_config {
      pooling_mode = "TRANSACTION"
      pool_discard = false
    }
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
    }

    permission {
      database_name = "newdb"
    }
  }

  user {
    name     = "bob"
    password = "anothersecurepassword"

    permission {
      database_name = "newdb"
    }

    permission {
      database_name = "fornewuserdb"
    }
  }

  host {
    zone      = "ru-central1-a"
    name      = "na"
    subnet_id = "${yandex_vpc_subnet.mdb-pg-test-subnet-a.id}"
    priority  = 1
  }
  host {
    zone                    = "ru-central1-b"
    name                    = "nb"
    replication_source_name = "na"
    subnet_id               = "${yandex_vpc_subnet.mdb-pg-test-subnet-b.id}"
  }
  host {
    zone      		 = "ru-central1-c"
    name      		 = "nc"
    assign_public_ip = true
    subnet_id 		 = "${yandex_vpc_subnet.mdb-pg-test-subnet-c.id}"
  }

  database {
    owner      = "alice"
    name       = "testdb"
    lc_collate = "en_US.UTF-8"
    lc_type    = "en_US.UTF-8"
  }

  database {
    owner = "alice"
    name  = "newdb"
  }

  database {
    owner = "bob"
    name  = "fornewuserdb"
  }

  security_group_ids = ["${yandex_vpc_security_group.mdb-pg-test-sg-y.id}"]
}
`, name, desc)
}
