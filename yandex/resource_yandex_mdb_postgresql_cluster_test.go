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

const (
	pgResource = "yandex_mdb_postgresql_cluster.foo"
)

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
			"user", // passwords are not returned
			"database",
			"health",                         // volatile value
			"host.0.name",                    // not returned
			"host.1.name",                    // not returned
			"host.2.name",                    // not returned
			"host.3.name",                    // not returned
			"host.0.replication_source_name", // not returned
			"host.1.replication_source_name", // not returned
			"host.2.replication_source_name", // not returned
			"host.3.replication_source_name", // not returned
			"host_master_name",               // not returned
		},
	}
}

// Test that a PostgreSQL Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_full(t *testing.T) {
	t.Parallel()

	var cluster postgresql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-postgresql")
	clusterResource := "yandex_mdb_postgresql_cluster.foo"
	pgDesc := "PostgreSQL Cluster Terraform Test"
	pgDesc2 := "PostgreSQL Cluster Terraform Test Updated"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			// 1. Create PostgreSQL Cluster
			{
				Config: testAccMDBPGClusterConfigMain(clusterName, pgDesc, "PRESTABLE", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 1),
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttr(clusterResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(clusterResource, "description", pgDesc),
					resource.TestCheckResourceAttr(clusterResource, "database.0.lc_collate", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(clusterResource, "database.0.lc_type", "en_US.UTF-8"),
					resource.TestCheckResourceAttrSet(clusterResource, "host.0.fqdn"),
					testAccCheckMDBPGClusterContainsLabel(&cluster, "test_key", "test_value"),
					testAccCheckMDBPGClusterHasResources(&cluster, "s2.micro", "network-ssd", 10737418240),
					testAccCheckMDBPGClusterHasUsers(clusterResource, map[string][]string{"alice": {"testdb"}}),
					testAccCheckMDBPGClusterHasDatabases(clusterResource, []string{"testdb"}),
					testAccCheckCreatedAtAttr(clusterResource),
					resource.TestCheckResourceAttr(clusterResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(clusterResource, "maintenance_window.0.day", "SAT"),
					resource.TestCheckResourceAttr(clusterResource, "maintenance_window.0.hour", "12"),
					resource.TestCheckResourceAttr(clusterResource, "deletion_protection", "true"),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			// 3. uncheck 'deletion_protection'
			{
				Config: testAccMDBPGClusterConfigMain(clusterName, pgDesc, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 1),
					resource.TestCheckResourceAttr(clusterResource, "deletion_protection", "false"),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			// 5. check 'deletion_protection'
			{
				Config: testAccMDBPGClusterConfigMain(clusterName, pgDesc, "PRESTABLE", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 1),
					resource.TestCheckResourceAttr(clusterResource, "deletion_protection", "true"),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			// 7. trigger deletion by changing environment
			{
				Config:      testAccMDBPGClusterConfigMain(clusterName, pgDesc, "PRODUCTION", true),
				ExpectError: regexp.MustCompile(".*The operation was rejected because cluster has 'deletion_protection' = ON.*"),
			},
			// 8. uncheck 'deletion_protection'
			{
				Config: testAccMDBPGClusterConfigMain(clusterName, pgDesc, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 1),
					resource.TestCheckResourceAttr(clusterResource, "deletion_protection", "false"),
				),
			},
			{
				Config:      testAccMDBPGClusterConfigDisallowedUpdateLocale(clusterName, pgDesc),
				ExpectError: regexp.MustCompile("impossible to change lc_collate or lc_type for PostgreSQL Cluster database .*"),
			},
			{
				Config:      testAccMDBPGClusterConfigDisallowedUpdateOwner(clusterName, pgDesc),
				ExpectError: regexp.MustCompile("impossible to change owner for PostgreSQL Cluster database .*"),
			},
			mdbPGClusterImportStep(clusterResource),
			// 12. Change some options
			{
				Config: testAccMDBPGClusterConfigUpdated(clusterName, pgDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 1),
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttr(clusterResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(clusterResource, "description", pgDesc2),
					resource.TestCheckResourceAttrSet(clusterResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(clusterResource, "config.0.access.0.web_sql"),
					resource.TestCheckResourceAttrSet(clusterResource, "config.0.access.0.serverless"),
					testAccCheckMDBPGClusterContainsLabel(&cluster, "new_key", "new_value"),
					testAccCheckMDBPGClusterHasResources(&cluster, "s2.micro", "network-ssd", 19327352832),
					testAccCheckMDBPGClusterHasPoolerConfig(&cluster, "TRANSACTION", false),
					testAccCheckMDBPGClusterHasUsers(clusterResource, map[string][]string{"alice": {"testdb", "newdb"}, "bob": {"newdb", "fornewuserdb"}}),
					testAccCheckClusterSettingsAccessWebSQL(clusterResource),
					testAccCheckClusterSettingsPerformanceDiagnostics(clusterResource),
					testAccCheckConnLimitUpdateUserSettings(clusterResource),
					testAccCheckMDBPGClusterHasDatabases(clusterResource, []string{"testdb", "newdb", "fornewuserdb"}),
					testAccCheckSettingsUpdateUserSettings(clusterResource),
					testAccCheckPostgresqlConfigUpdate(clusterResource),
					testAccCheckCreatedAtAttr(clusterResource),
					resource.TestCheckResourceAttr(clusterResource, "security_group_ids.#", "2"),

					resource.TestCheckResourceAttr(clusterResource, "maintenance_window.0.day", "WED"),
					resource.TestCheckResourceAttr(clusterResource, "maintenance_window.0.hour", "22"),
					resource.TestCheckResourceAttr(clusterResource, "config.0.backup_retain_period_days", "12"),
				),
			},
		},
	})
}

// Test that a PostgreSQL HA Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_HAWithoutNames_update(t *testing.T) {
	t.Parallel()

	var cluster postgresql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-postgresql")
	clusterResource := "yandex_mdb_postgresql_cluster.ha_cluster"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPGClusterConfigHA(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 3),
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttrSet(clusterResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(clusterResource, "host.1.fqdn"),
					resource.TestCheckResourceAttrSet(clusterResource, "host.2.fqdn"),
					resource.TestCheckResourceAttr(clusterResource, "host.0.zone", "ru-central1-a"),
					resource.TestCheckResourceAttr(clusterResource, "host.1.zone", "ru-central1-b"),
					resource.TestCheckResourceAttr(clusterResource, "host.2.zone", "ru-central1-c"),
					resource.TestCheckResourceAttr(clusterResource, "host.2.assign_public_ip", "true"),
					testAccCheckCreatedAtAttr(clusterResource),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			{
				Config: testAccMDBPGClusterConfigHAChangePublicIP(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 3),
					resource.TestCheckResourceAttr(clusterResource, "host.0.assign_public_ip", "true"),
				),
			},
		},
	})
}

// Test that a PostgreSQL HA named Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_HAWithNames_update(t *testing.T) {
	t.Parallel()

	var cluster postgresql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-postgresql")
	clusterResource := "yandex_mdb_postgresql_cluster.ha_cluster_with_names"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPGClusterConfigHANamed(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 3),
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttr(clusterResource, "host.0.name", "na"),
					resource.TestCheckResourceAttr(clusterResource, "host.1.name", "nb"),
					resource.TestCheckResourceAttr(clusterResource, "host.2.name", "nc"),
					resource.TestCheckResourceAttr(clusterResource, "host.0.assign_public_ip", "true"),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			{
				Config: testAccMDBPGClusterConfigHANamedChangePublicIP(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 3),
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttr(clusterResource, "host.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(clusterResource, "host.2.assign_public_ip", "true"),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			{
				Config: testAccMDBPGClusterConfigHANamedWithCascade(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 3),
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttrSet(clusterResource, "host.0.replication_source"),
					resource.TestCheckResourceAttr(clusterResource, "host.0.replication_source_name", "nb"),
					resource.TestCheckResourceAttrSet(clusterResource, "host.1.replication_source"),
					resource.TestCheckResourceAttr(clusterResource, "host.1.replication_source_name", "nc"),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			{
				Config: testAccMDBPGClusterConfigHANamedWithPriorities(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(clusterResource, &cluster, 3),
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttr(clusterResource, "host.1.priority", "5"),
					resource.TestCheckResourceAttr(clusterResource, "host.2.priority", "10"),
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

		userConfig := cluster.Config.GetPostgresqlConfig_12().UserConfig

		if userConfig.MaxConnections.GetValue() != 395 {
			return fmt.Errorf("Field 'config.postgresql_config.max_connections' wasn`t changed for with value 395. Current value is %v",
				userConfig.MaxConnections.GetValue())
		}

		if !userConfig.EnableParallelHash.GetValue() {
			return fmt.Errorf("Field 'config.postgresql_config.enable_parallel_hash' wasn`t changed for with value true. Current value is %v",
				userConfig.EnableParallelHash.GetValue())
		}

		if userConfig.VacuumCleanupIndexScaleFactor.GetValue() != 0.2 {
			return fmt.Errorf("Field 'config.postgresql_config.vacuum_cleanup_index_scale_factor' wasn`t changed for with value 0.2. Current value is %v",
				userConfig.VacuumCleanupIndexScaleFactor.GetValue())
		}

		if userConfig.DefaultTransactionIsolation != 1 {
			return fmt.Errorf("Field 'config.postgresql_config.default_transaction_isolation' wasn`t changed for with value 1. Current value is %v",
				userConfig.DefaultTransactionIsolation)
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
  network_id     = yandex_vpc_network.mdb-pg-test-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-pg-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.mdb-pg-test-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-pg-test-subnet-c" {
  zone           = "ru-central1-c"
  network_id     = yandex_vpc_network.mdb-pg-test-net.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_security_group" "mdb-pg-test-sg-x" {
  network_id     = yandex_vpc_network.mdb-pg-test-net.id
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
  network_id     = yandex_vpc_network.mdb-pg-test-net.id
  
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
  network_id  = yandex_vpc_network.mdb-pg-test-net.id

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
    subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
  }

  database {
    owner      = "alice"
    name       = "testdb"
    lc_collate = "en_US.UTF-8"
    lc_type    = "en_US.UTF-8"
  }

  security_group_ids = [yandex_vpc_security_group.mdb-pg-test-sg-x.id]
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
  network_id  = yandex_vpc_network.mdb-pg-test-net.id

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
    subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
  }

  database {
    owner      = "alice"
    name       = "testdb"
    lc_collate = "C"
    lc_type    = "en_US.UTF-8"
  }

  security_group_ids = [yandex_vpc_security_group.mdb-pg-test-sg-x.id]
}
`, name, desc)
}

func testAccMDBPGClusterConfigDisallowedUpdateOwner(name, desc string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id

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
    subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
  }

  database {
    owner      = "bob"
    name       = "testdb"
    lc_collate = "en_US.UTF-8"
    lc_type    = "en_US.UTF-8"
  }

  security_group_ids = [yandex_vpc_security_group.mdb-pg-test-sg-x.id]
}
`, name, desc)
}

func testAccMDBPGClusterConfigUpdated(name, desc string) string {

	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id

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
    zone             = "ru-central1-a"
    subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
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

  security_group_ids = [yandex_vpc_security_group.mdb-pg-test-sg-x.id, yandex_vpc_security_group.mdb-pg-test-sg-y.id]
}
`, name, desc)
}

func testAccMDBPGClusterConfigHABasicConfig(name, hosts string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "ha_cluster" {
  name        = "%s"
  description = "PostgreSQL HA Cluster without names Terraform Test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id

  config {
    version = 13

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

  %s
}
`, name, hosts)
}

func testAccMDBPGClusterConfigHA(name string) string {
	return testAccMDBPGClusterConfigHABasicConfig(name, `
	host {
		zone             = "ru-central1-a"
		subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
	}
	host {
		zone             = "ru-central1-b"
		subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
	}
	host {
		zone             = "ru-central1-c"
		subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-c.id
		assign_public_ip = true
	}
`)
}

func testAccMDBPGClusterConfigHAChangePublicIP(name string) string {
	return testAccMDBPGClusterConfigHABasicConfig(name, `
	host {
		zone             = "ru-central1-a"
		subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
		assign_public_ip = true
	}
	host {
		zone             = "ru-central1-b"
		subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
	}
	host {
		zone             = "ru-central1-c"
		subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-c.id
	}
`)
}

func testAccMDBPGClusterConfigHANamedBasicConfig(name, hosts string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "ha_cluster_with_names" {
  name        = "%s"
  description = "PostgreSQL HA Cluster Terraform Test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id

  labels = {
    new_key = "new_value"
  }

  config {
    version = 14

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

%s
}
`, name, hosts)
}

func testAccMDBPGClusterConfigHANamed(name string) string {
	return testAccMDBPGClusterConfigHANamedBasicConfig(name, `
  host {
    name                    = "na"
    zone                    = "ru-central1-a"
    subnet_id               = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
    
    assign_public_ip = true
  }

  host {
    name                    = "nb"
    zone                    = "ru-central1-b"
    subnet_id               = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
  }

  host {
    name             = "nc"
    zone             = "ru-central1-c"
    subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-c.id
  }
`)
}

func testAccMDBPGClusterConfigHANamedChangePublicIP(name string) string {
	return testAccMDBPGClusterConfigHANamedBasicConfig(name, `
  host {
    name                    = "na"
    zone                    = "ru-central1-a"
    subnet_id               = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
  }

  host {
    name                    = "nb"
    zone                    = "ru-central1-b"
    subnet_id               = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
  }

  host {
    name             = "nc"
    zone             = "ru-central1-c"
    subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-c.id
    
    assign_public_ip = true
  }
`)
}

func testAccMDBPGClusterConfigHANamedWithCascade(name string) string {
	return testAccMDBPGClusterConfigHANamedBasicConfig(name, `
  host {
    name                    = "na"
    zone                    = "ru-central1-a"
    subnet_id               = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
    
    replication_source_name = "nb"
  }

  host {
    name                    = "nb"
    zone                    = "ru-central1-b"
    subnet_id               = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
    
    replication_source_name = "nc"
  }

  host {
    name             = "nc"
    zone             = "ru-central1-c"
    subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-c.id
    assign_public_ip = true
  }
`)
}

func testAccMDBPGClusterConfigHANamedWithPriorities(name string) string {
	return testAccMDBPGClusterConfigHANamedBasicConfig(name, `
  host {
    name                    = "na"
    zone                    = "ru-central1-a"
    subnet_id               = yandex_vpc_subnet.mdb-pg-test-subnet-a.id

    replication_source_name = "nb"
  }

  host {
    name                    = "nb"
    zone                    = "ru-central1-b"
    subnet_id               = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
    replication_source_name = "nc"
    
    priority                = 5
  }

  host {
    name             = "nc"
    zone             = "ru-central1-c"
    subnet_id        = yandex_vpc_subnet.mdb-pg-test-subnet-c.id
    assign_public_ip = true
    
    priority         = 10
  }
`)
}
