package mdb_postgresql_cluster_beta_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"regexp"
	"sort"

	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	pconfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	defaultMDBPageSize                      = 1000
	pgResource                              = "yandex_mdb_postgresql_cluster_beta.foo"
	pgRestoreBackupId                       = "c9qrbucrcvm6a50tblv2:c9q698sst87e4vhkvrsm"
	yandexMDBPostgreSQLClusterCreateTimeout = 30 * time.Minute // TODO refactor
	yandexMDBPostgreSQLClusterDeleteTimeout = 15 * time.Minute
	yandexMDBPostgreSQLClusterUpdateTimeout = 60 * time.Minute
)

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

resource "yandex_vpc_subnet" "mdb-pg-test-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.mdb-pg-test-net.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_security_group" "sgroup1" {
  description = "Test security group 1"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id
}

resource "yandex_vpc_security_group" "sgroup2" {
  description = "Test security group 2"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id
}

`

var (
	pgVersions   = [...]string{"13", "14", "15", "16", "17"}
	pg1CVersions = [...]string{"13-1c", "14-1c", "15-1c"} //nolint:unused
)

func init() {
	resource.AddTestSweepers("yandex_mdb_postgresql_cluster_beta", &resource.Sweeper{
		Name: "yandex_mdb_postgresql_cluster_beta",
		F:    testSweepMDBPostgreSQLCluster,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepMDBPostgreSQLCluster(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.MDB().PostgreSQL().Cluster().List(context.Background(), &postgresql.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
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

func sweepMDBPostgreSQLCluster(conf *config.Config, id string) bool {
	return test.SweepWithRetry(sweepMDBPostgreSQLClusterOnce, conf, "PostgreSQL cluster", id)
}

func sweepMDBPostgreSQLClusterOnce(conf *config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexMDBPostgreSQLClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	op, err := conf.SDK.MDB().PostgreSQL().Cluster().Update(ctx, &postgresql.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = test.HandleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(test.ErrorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.SDK.MDB().PostgreSQL().Cluster().Delete(ctx, &postgresql.DeleteClusterRequest{
		ClusterId: id,
	})
	return test.HandleSweepOperation(ctx, conf, op, err)
}

func mdbPGClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"health", // volatile value
			"hosts",  // volatile value
			// API returns inconsistent result with importing postgresql_config
			"config.postgresql_config", // volatile value
		},
	}
}

// Test that a PostgreSQL Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_basic(t *testing.T) {
	t.Parallel()

	version := "13"
	versionUpdate := "14"

	resources := `
	  resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-hdd"
	`

	resourcesUpdated := `
	  resource_preset_id = "s2.micro"
	  disk_size		  = 12
	  disk_type_id	   = "network-hdd"
	`

	log.Printf("TestAccMDBPostgreSQLCluster_basic: version %s", version)
	var cluster postgresql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-postgresql-cluster-basic")
	resourceId := "cluster_basic_test"
	clusterResource := "yandex_mdb_postgresql_cluster_beta." + resourceId
	description := "PostgreSQL Cluster Terraform Test Basic"
	descriptionUpdated := fmt.Sprintf("%s Updated", description)
	folderID := test.GetExampleFolderID()

	labels := `
    key1 = "value1"
    key2 = "value2"
    key3 = "value3"
    `
	labelsUpdated := `
    key4 = "value4"
    `

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			// Create PostgreSQL Cluster
			{
				Config: testAccMDBPGClusterBasic(resourceId, clusterName, description, "PRESTABLE", labels, version, resources),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact("PRESTABLE")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("version"), knownvalue.StringExact(version)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("autofailover"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("performance_diagnostics"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"enabled":                      knownvalue.Bool(false),
						"sessions_sampling_interval":   knownvalue.Int64Exact(60),
						"statements_sampling_interval": knownvalue.Int64Exact(600),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(false),
							"data_transfer": knownvalue.Bool(false),
							"web_sql":       knownvalue.Bool(false),
							"serverless":    knownvalue.Bool(false),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(7)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"hours":   knownvalue.Int64Exact(0),
						"minutes": knownvalue.Int64Exact(0),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("pooler_config"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"pool_discard": knownvalue.Null(),
							"pooling_mode": knownvalue.StringExact(postgresql.ConnectionPoolerConfig_POOLING_MODE_UNSPECIFIED.String()),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("disk_size_autoscaling"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"disk_size_limit":           knownvalue.Int64Exact(0),
						"emergency_usage_threshold": knownvalue.Int64Exact(0),
						"planned_usage_threshold":   knownvalue.Int64Exact(0),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"type": knownvalue.StringExact("ANYTIME"),
						"day":  knownvalue.Null(),
						"hour": knownvalue.Null(),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("security_group_ids"), knownvalue.SetSizeExact(0)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-hdd", 10*1024*1024*1024),
					testAccCheckClusterAutofailoverExact(&cluster, true),
					testAccCheckClusterDeletionProtectionExact(&cluster, false),
					testAccCheckClusterAccessExact(&cluster, &postgresql.Access{
						DataLens:     false,
						DataTransfer: false,
						WebSql:       false,
						Serverless:   false,
					}),
					testAccCheckClusterPerformanceDiagnosticsExact(&cluster, &postgresql.PerformanceDiagnostics{
						Enabled:                    false,
						SessionsSamplingInterval:   60,
						StatementsSamplingInterval: 600,
					}),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   0,
						Minutes: 0,
					}),
					testAccCheckClusterPoolerConfigExact(&cluster, nil),
					testAccCheckClusterDiskSizeAutoscalingExact(&cluster, &postgresql.DiskSizeAutoscaling{}),
					testAccCheckClusterMaintenanceWindow(&cluster, &postgresql.MaintenanceWindow{
						Policy: &postgresql.MaintenanceWindow_Anytime{
							Anytime: &postgresql.AnytimeMaintenanceWindow{},
						},
					}),
					testAccCheckClusterSecurityGroupIdsExact(&cluster, nil),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			// Update PostgreSQL Cluster
			{
				Config: testAccMDBPGClusterBasic(resourceId, clusterName, descriptionUpdated, "PRESTABLE", labelsUpdated, versionUpdate, resourcesUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionUpdated)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact("PRESTABLE")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("version"), knownvalue.StringExact(versionUpdate)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("autofailover"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(false),
							"data_transfer": knownvalue.Bool(false),
							"web_sql":       knownvalue.Bool(false),
							"serverless":    knownvalue.Bool(false),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("performance_diagnostics"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"enabled":                      knownvalue.Bool(false),
						"sessions_sampling_interval":   knownvalue.Int64Exact(60),
						"statements_sampling_interval": knownvalue.Int64Exact(600),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"hours":   knownvalue.Int64Exact(0),
						"minutes": knownvalue.Int64Exact(0),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(7)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("pooler_config"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"pool_discard": knownvalue.Null(),
							"pooling_mode": knownvalue.StringExact(postgresql.ConnectionPoolerConfig_POOLING_MODE_UNSPECIFIED.String()),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("disk_size_autoscaling"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"disk_size_limit":           knownvalue.Int64Exact(0),
						"emergency_usage_threshold": knownvalue.Int64Exact(0),
						"planned_usage_threshold":   knownvalue.Int64Exact(0),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"type": knownvalue.StringExact("ANYTIME"),
						"day":  knownvalue.Null(),
						"hour": knownvalue.Null(),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("security_group_ids"), knownvalue.SetSizeExact(0)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key4": "value4"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-hdd", 12*1024*1024*1024),
					testAccCheckClusterAutofailoverExact(&cluster, true),
					testAccCheckClusterDeletionProtectionExact(&cluster, false),
					testAccCheckClusterAccessExact(&cluster, &postgresql.Access{
						DataLens:     false,
						DataTransfer: false,
						WebSql:       false,
						Serverless:   false,
					}),
					testAccCheckClusterPerformanceDiagnosticsExact(&cluster, &postgresql.PerformanceDiagnostics{
						Enabled:                    false,
						SessionsSamplingInterval:   60,
						StatementsSamplingInterval: 600,
					}),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   0,
						Minutes: 0,
					}),
					testAccCheckClusterDiskSizeAutoscalingExact(&cluster, &postgresql.DiskSizeAutoscaling{}),
					testAccCheckClusterPoolerConfigExact(&cluster, nil),
					testAccCheckClusterMaintenanceWindow(&cluster, &postgresql.MaintenanceWindow{
						Policy: &postgresql.MaintenanceWindow_Anytime{
							Anytime: &postgresql.AnytimeMaintenanceWindow{},
						},
					}),
					testAccCheckClusterSecurityGroupIdsExact(&cluster, nil),
				),
			},
			mdbPGClusterImportStep(clusterResource),
		},
	})
}

// Test that a PostgreSQL Cluster can be created, updated and destroyed
// TODO: Check enable and disable pooler_config
func TestAccMDBPostgreSQLCluster_full(t *testing.T) {
	t.Parallel()

	version := "14-1c"
	versionUpdate := "15-1c"

	resources := `
	  resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
	`

	log.Printf("TestAccMDBPostgreSQLCluster_full: version %s", version)
	var cluster postgresql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-postgresql-cluster-full")

	resourceId := "cluster_full_test"
	clusterResource := "yandex_mdb_postgresql_cluster_beta." + resourceId

	description := "PostgreSQL Cluster Terraform Test Full"
	descriptionUpdated := fmt.Sprintf("%s Updated", description)
	folderID := test.GetExampleFolderID()

	environment := "PRODUCTION"

	labels := `
    key1 = "value1"
    key2 = "value2"
    key3 = "value3"
    `
	labelsUpdated := `
    key4 = "value4"
    `

	access := `
		data_transfer = true
		web_sql = true
		serverless = false
		data_lens = false
	`

	accessUpdated := `
		serverless = true
		data_lens = true
		data_transfer = false
		web_sql = false
	`

	performanceDiagnostics := `
		enabled = true
		sessions_sampling_interval = 60
		statements_sampling_interval = 600
	`

	performanceDiagnosticsUpdated := `
		sessions_sampling_interval = 500
		statements_sampling_interval = 1000
	`

	backupRetainPeriodDays := 7
	backupRetainPeriodDaysUpdated := 14

	backupWindowStart := `
		hours = 5
		minutes = 4
	`

	backupWindowStartUpdated := `
		hours = 10
		minutes = 3
	`

	postgresqlConfig := `
		max_connections                = 100
		enable_parallel_hash           = true
		autovacuum_vacuum_scale_factor = 0.34
		default_transaction_isolation  = "TRANSACTION_ISOLATION_READ_COMMITTED"
		shared_preload_libraries       = "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
	`

	postgresqlConfigUpdated := `
		max_connections				= 200
		enable_parallel_hash		   = false
		autovacuum_vacuum_scale_factor = 0.35
		default_transaction_isolation  = "TRANSACTION_ISOLATION_REPEATABLE_READ"
		shared_preload_libraries       = "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"
	`

	poolerCfg := `
		pooling_mode = "TRANSACTION"
		pool_discard = true
	`

	poolerCfgUpdated := `
		pooling_mode = "SESSION"
		pool_discard = false
	`

	dsa := `
		disk_size_limit = 15
		emergency_usage_threshold = 20
	`

	dsaUpdate := `
		disk_size_limit = 20
		emergency_usage_threshold = 15
		planned_usage_threshold = 10
	`

	maintenanceWindow := `
		type = "ANYTIME"
	`

	maintenanceWindowUpdated := `
		type = "WEEKLY"
		day  = "MON"
		hour = 5
	`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			// Create PostgreSQL Cluster
			{
				Config: testAccMDBPGClusterFull(
					resourceId, clusterName, description,
					environment, labels, version,
					resources, access,
					performanceDiagnostics,
					backupWindowStart,
					poolerCfg,
					dsa,
					postgresqlConfig,
					maintenanceWindow,
					backupRetainPeriodDays, true, true,
					[]string{
						"yandex_vpc_security_group.sgroup1.id",
					},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact(environment)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("version"), knownvalue.StringExact(version)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("autofailover"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(false),
							"data_transfer": knownvalue.Bool(true),
							"web_sql":       knownvalue.Bool(true),
							"serverless":    knownvalue.Bool(false),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("performance_diagnostics"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"enabled":                      knownvalue.Bool(true),
							"sessions_sampling_interval":   knownvalue.Int64Exact(60),
							"statements_sampling_interval": knownvalue.Int64Exact(600),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(7)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"hours":   knownvalue.Int64Exact(5),
							"minutes": knownvalue.Int64Exact(4),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("pooler_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"pooling_mode": knownvalue.StringExact(postgresql.ConnectionPoolerConfig_PoolingMode_name[2]),
						"pool_discard": knownvalue.Bool(true),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("disk_size_autoscaling"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"disk_size_limit":           knownvalue.Int64Exact(15),
						"emergency_usage_threshold": knownvalue.Int64Exact(20),
						"planned_usage_threshold":   knownvalue.Int64Exact(0),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("postgresql_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"max_connections":                knownvalue.StringExact("100"),
						"enable_parallel_hash":           knownvalue.StringExact("true"),
						"autovacuum_vacuum_scale_factor": knownvalue.StringExact("0.34"),
						"default_transaction_isolation":  knownvalue.StringExact("TRANSACTION_ISOLATION_READ_COMMITTED"),
						"shared_preload_libraries":       knownvalue.StringExact("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"type": knownvalue.StringExact("ANYTIME"),
							"day":  knownvalue.Null(),
							"hour": knownvalue.Null(),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("security_group_ids"), knownvalue.SetSizeExact(1)),
					statecheck.CompareValueCollection(
						clusterResource,
						[]tfjsonpath.Path{
							tfjsonpath.New("security_group_ids"),
						},
						"yandex_vpc_security_group.sgroup1",
						tfjsonpath.New("id"), compare.ValuesSame(),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckClusterAutofailoverExact(&cluster, true),
					testAccCheckClusterDeletionProtectionExact(&cluster, true),
					testAccCheckClusterAccessExact(&cluster, &postgresql.Access{
						DataLens:     false,
						DataTransfer: true,
						WebSql:       true,
						Serverless:   false,
					}),
					testAccCheckClusterPerformanceDiagnosticsExact(
						&cluster,
						&postgresql.PerformanceDiagnostics{
							Enabled:                    true,
							SessionsSamplingInterval:   60,
							StatementsSamplingInterval: 600,
						},
					),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   5,
						Minutes: 4,
					}),
					testAccCheckClusterPoolerConfigExact(&cluster, &postgresql.ConnectionPoolerConfig{
						PoolingMode: postgresql.ConnectionPoolerConfig_TRANSACTION,
						PoolDiscard: wrapperspb.Bool(true),
					}),
					testAccCheckClusterDiskSizeAutoscalingExact(&cluster, &postgresql.DiskSizeAutoscaling{
						DiskSizeLimit:           datasize.ToBytes(15),
						EmergencyUsageThreshold: 20,
					}),
					testAccCheckClusterPostgresqlConfigExact(&cluster, &pconfig.PostgresqlConfig14_1C{
						MaxConnections:              wrapperspb.Int64(100),
						EnableParallelHash:          wrapperspb.Bool(true),
						AutovacuumVacuumScaleFactor: wrapperspb.Double(0.34),
						DefaultTransactionIsolation: pconfig.PostgresqlConfig14_1C_TRANSACTION_ISOLATION_READ_COMMITTED,
						SharedPreloadLibraries: []pconfig.PostgresqlConfig14_1C_SharedPreloadLibraries{
							pconfig.PostgresqlConfig14_1C_SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,
							pconfig.PostgresqlConfig14_1C_SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN,
						},
					}, []string{
						"MaxConnections",
						"EnableParallelHash",
						"AutovacuumVacuumScaleFactor",
						"DefaultTransactionIsolation",
						"SharedPreloadLibraries",
					}),
					testAccCheckClusterMaintenanceWindow(&cluster, &postgresql.MaintenanceWindow{
						Policy: &postgresql.MaintenanceWindow_Anytime{
							Anytime: &postgresql.AnytimeMaintenanceWindow{},
						},
					}),
					testAccCheckClusterSecurityGroupIdsExact(
						&cluster,
						[]string{
							"yandex_vpc_security_group.sgroup1",
						},
					),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			{
				Config: testAccMDBPGClusterFull(
					resourceId, clusterName, descriptionUpdated,
					environment, labelsUpdated, versionUpdate, resources, accessUpdated,
					performanceDiagnosticsUpdated,
					backupWindowStartUpdated,
					poolerCfgUpdated,
					dsaUpdate,
					postgresqlConfigUpdated,
					maintenanceWindowUpdated,
					backupRetainPeriodDaysUpdated, false, false,
					[]string{
						"yandex_vpc_security_group.sgroup2.id",
					},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionUpdated)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact(environment)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("version"), knownvalue.StringExact(versionUpdate)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("autofailover"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(true),
							"data_transfer": knownvalue.Bool(false),
							"web_sql":       knownvalue.Bool(false),
							"serverless":    knownvalue.Bool(true),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("performance_diagnostics"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"enabled":                      knownvalue.Bool(false),
							"sessions_sampling_interval":   knownvalue.Int64Exact(500),
							"statements_sampling_interval": knownvalue.Int64Exact(1000),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(14)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"hours":   knownvalue.Int64Exact(10),
							"minutes": knownvalue.Int64Exact(3),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("pooler_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"pooling_mode": knownvalue.StringExact(postgresql.ConnectionPoolerConfig_PoolingMode_name[1]),
						"pool_discard": knownvalue.Bool(false),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("disk_size_autoscaling"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"disk_size_limit":           knownvalue.Int64Exact(20),
						"emergency_usage_threshold": knownvalue.Int64Exact(15),
						"planned_usage_threshold":   knownvalue.Int64Exact(10),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("postgresql_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"max_connections":                knownvalue.StringExact("200"),
						"enable_parallel_hash":           knownvalue.StringExact("false"),
						"autovacuum_vacuum_scale_factor": knownvalue.StringExact("0.35"),
						"default_transaction_isolation":  knownvalue.StringExact("TRANSACTION_ISOLATION_REPEATABLE_READ"),
						"shared_preload_libraries":       knownvalue.StringExact("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"type": knownvalue.StringExact("WEEKLY"),
							"day":  knownvalue.StringExact("MON"),
							"hour": knownvalue.Int64Exact(5),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("security_group_ids"), knownvalue.SetSizeExact(1)),
					statecheck.CompareValueCollection(
						clusterResource,
						[]tfjsonpath.Path{
							tfjsonpath.New("security_group_ids"),
						},
						"yandex_vpc_security_group.sgroup2",
						tfjsonpath.New("id"), compare.ValuesSame(),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key4": "value4"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckClusterAutofailoverExact(&cluster, false),
					testAccCheckClusterDeletionProtectionExact(&cluster, false),
					testAccCheckClusterAccessExact(&cluster, &postgresql.Access{
						DataLens:     true,
						DataTransfer: false,
						WebSql:       false,
						Serverless:   true,
					}),
					testAccCheckClusterPerformanceDiagnosticsExact(
						&cluster,
						&postgresql.PerformanceDiagnostics{
							SessionsSamplingInterval:   500,
							StatementsSamplingInterval: 1000,
						},
					),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(14)),
					testAccCheckClusterPostgresqlConfigExact(&cluster, &pconfig.PostgresqlConfig15_1C{
						MaxConnections:              wrapperspb.Int64(200),
						EnableParallelHash:          wrapperspb.Bool(false),
						AutovacuumVacuumScaleFactor: wrapperspb.Double(0.35),
						DefaultTransactionIsolation: pconfig.PostgresqlConfig15_1C_TRANSACTION_ISOLATION_REPEATABLE_READ,
						SharedPreloadLibraries: []pconfig.PostgresqlConfig15_1C_SharedPreloadLibraries{
							pconfig.PostgresqlConfig15_1C_SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,
						},
					}, []string{
						"MaxConnections",
						"EnableParallelHash",
						"AutovacuumVacuumScaleFactor",
						"DefaultTransactionIsolation",
						"SharedPreloadLibraries",
					}),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   10,
						Minutes: 3,
					}),
					testAccCheckClusterPoolerConfigExact(&cluster, &postgresql.ConnectionPoolerConfig{
						PoolingMode: postgresql.ConnectionPoolerConfig_SESSION,
						PoolDiscard: wrapperspb.Bool(false),
					}),
					testAccCheckClusterDiskSizeAutoscalingExact(&cluster, &postgresql.DiskSizeAutoscaling{
						DiskSizeLimit:           datasize.ToBytes(20),
						EmergencyUsageThreshold: 15,
						PlannedUsageThreshold:   10,
					}),
					testAccCheckClusterMaintenanceWindow(&cluster, &postgresql.MaintenanceWindow{
						Policy: &postgresql.MaintenanceWindow_WeeklyMaintenanceWindow{
							WeeklyMaintenanceWindow: &postgresql.WeeklyMaintenanceWindow{
								Day:  postgresql.WeeklyMaintenanceWindow_MON,
								Hour: 5,
							},
						},
					}),
					testAccCheckClusterSecurityGroupIdsExact(
						&cluster,
						[]string{
							"yandex_vpc_security_group.sgroup2",
						},
					),
				),
			},
			mdbPGClusterImportStep(clusterResource),
		},
	})
}

// TODO: Check enable and disable disk_size_autoscaling when fix api
func TestAccMDBPostgreSQLCluster_mixed(t *testing.T) {
	t.Parallel()

	version := "17"

	log.Printf("TestAccMDBPostgreSQLCluster_mixed: version %s", version)
	var cluster postgresql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-postgresql-cluster-mixed")

	resourceId := "cluster_mixed_test"
	clusterResource := "yandex_mdb_postgresql_cluster_beta." + resourceId

	folderID := test.GetExampleFolderID()

	descriptionFull := "Cluster test mixed: full"
	descriptionBasic := "Cluster test mixed: basic"

	environment := "PRODUCTION"
	labels := `
		key = "value"
	`

	access := `
	data_lens = false
	serverless = false
	`

	performanceDiagnostics := `
		enabled = false
		sessions_sampling_interval = 60
		statements_sampling_interval = 600
	`

	maintenanceWindow := `
		type = "ANYTIME"
	`

	backupRetainPeriodDays := 7
	backupWindowStart := `
		hours = 0
		minutes = 0
	`

	// TODO: Replace to empty when fix api
	postgresqlConfig := `
		password_encryption = "PASSWORD_ENCRYPTION_SCRAM_SHA_256"
	`

	resources := `
		resource_preset_id = "s2.micro"
		disk_size = 10
		disk_type_id = "network-ssd"
	`

	poolerConfig := `
			pool_discard = null
			pooling_mode = "POOLING_MODE_UNSPECIFIED"
		`

	dsa := `
		disk_size_limit = 20
	`

	stepsFullBasic := []resource.TestStep{
		{
			Config: testAccMDBPGClusterBasic(resourceId, clusterName, descriptionBasic, environment, labels, version, resources),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionBasic)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact(environment)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("version"), knownvalue.StringExact(version)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("autofailover"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"data_lens":     knownvalue.Bool(false),
						"data_transfer": knownvalue.Bool(false),
						"web_sql":       knownvalue.Bool(false),
						"serverless":    knownvalue.Bool(false),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("performance_diagnostics"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"enabled":                      knownvalue.Bool(false),
						"sessions_sampling_interval":   knownvalue.Int64Exact(60),
						"statements_sampling_interval": knownvalue.Int64Exact(600),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"hours":   knownvalue.Int64Exact(0),
						"minutes": knownvalue.Int64Exact(0),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(7)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("pooler_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
					"pool_discard": knownvalue.Null(),
					"pooling_mode": knownvalue.StringExact(postgresql.ConnectionPoolerConfig_POOLING_MODE_UNSPECIFIED.String()),
				})),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("disk_size_autoscaling"), knownvalue.ObjectExact(map[string]knownvalue.Check{
					"disk_size_limit":           knownvalue.Int64Exact(0),
					"emergency_usage_threshold": knownvalue.Int64Exact(0),
					"planned_usage_threshold":   knownvalue.Int64Exact(0),
				})),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("postgresql_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
					"password_encryption": knownvalue.StringExact("PASSWORD_ENCRYPTION_SCRAM_SHA_256"),
				})),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"type": knownvalue.StringExact("ANYTIME"),
						"day":  knownvalue.Null(),
						"hour": knownvalue.Null(),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("security_group_ids"), knownvalue.SetSizeExact(0)),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 1),
				testAccCheckClusterLabelsExact(&cluster, map[string]string{"key": "value"}),
				testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
				testAccCheckClusterAutofailoverExact(&cluster, true),
				testAccCheckClusterDeletionProtectionExact(&cluster, false),
				testAccCheckClusterAccessExact(&cluster, &postgresql.Access{
					DataLens:     false,
					DataTransfer: false,
					WebSql:       false,
					Serverless:   false,
				}),
				testAccCheckClusterPerformanceDiagnosticsExact(&cluster, &postgresql.PerformanceDiagnostics{
					Enabled:                    false,
					SessionsSamplingInterval:   60,
					StatementsSamplingInterval: 600,
				}),
				testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
				testAccCheckClusterPoolerConfigExact(&cluster, nil),
				testAccCheckClusterDiskSizeAutoscalingExact(&cluster, &postgresql.DiskSizeAutoscaling{}),
				testAccCheckClusterPostgresqlConfigExact(&cluster, &pconfig.PostgresqlConfig17{
					PasswordEncryption: pconfig.PostgresqlConfig17_PASSWORD_ENCRYPTION_SCRAM_SHA_256,
				}, nil),
				testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
					Hours:   0,
					Minutes: 0,
				}),
				testAccCheckClusterMaintenanceWindow(&cluster, &postgresql.MaintenanceWindow{
					Policy: &postgresql.MaintenanceWindow_Anytime{
						Anytime: &postgresql.AnytimeMaintenanceWindow{},
					},
				}),
				testAccCheckClusterSecurityGroupIdsExact(&cluster, nil),
			),
		},
		{
			Config: testAccMDBPGClusterFull(
				resourceId, clusterName, descriptionFull, environment, labels, version, resources, access,
				performanceDiagnostics,
				backupWindowStart,
				poolerConfig,
				dsa,
				postgresqlConfig,
				maintenanceWindow,
				backupRetainPeriodDays,
				true, false, []string{},
			),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionFull)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact(environment)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("version"), knownvalue.StringExact(version)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("autofailover"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"data_lens":     knownvalue.Bool(false),
						"data_transfer": knownvalue.Bool(false),
						"web_sql":       knownvalue.Bool(false),
						"serverless":    knownvalue.Bool(false),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("performance_diagnostics"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"enabled":                      knownvalue.Bool(false),
						"sessions_sampling_interval":   knownvalue.Int64Exact(60),
						"statements_sampling_interval": knownvalue.Int64Exact(600),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"hours":   knownvalue.Int64Exact(0),
						"minutes": knownvalue.Int64Exact(0),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(7)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("pooler_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
					"pool_discard": knownvalue.Null(),
					"pooling_mode": knownvalue.StringExact(postgresql.ConnectionPoolerConfig_POOLING_MODE_UNSPECIFIED.String()),
				})),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("disk_size_autoscaling"), knownvalue.ObjectExact(map[string]knownvalue.Check{
					"disk_size_limit":           knownvalue.Int64Exact(20),
					"emergency_usage_threshold": knownvalue.Int64Exact(0),
					"planned_usage_threshold":   knownvalue.Int64Exact(0),
				})),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("postgresql_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
					"password_encryption": knownvalue.StringExact("PASSWORD_ENCRYPTION_SCRAM_SHA_256"),
				})),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"type": knownvalue.StringExact("ANYTIME"),
						"day":  knownvalue.Null(),
						"hour": knownvalue.Null(),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("security_group_ids"), knownvalue.SetSizeExact(0)),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 1),
				testAccCheckClusterLabelsExact(&cluster, map[string]string{"key": "value"}),
				testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
				testAccCheckClusterAutofailoverExact(&cluster, true),
				testAccCheckClusterDeletionProtectionExact(&cluster, false),
				testAccCheckClusterAccessExact(&cluster, &postgresql.Access{
					DataLens:     false,
					DataTransfer: false,
					WebSql:       false,
					Serverless:   false,
				}),
				testAccCheckClusterPerformanceDiagnosticsExact(&cluster, &postgresql.PerformanceDiagnostics{
					Enabled:                    false,
					SessionsSamplingInterval:   60,
					StatementsSamplingInterval: 600,
				}),
				testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
					Hours:   0,
					Minutes: 0,
				}),
				testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
				testAccCheckClusterPoolerConfigExact(&cluster, nil),
				testAccCheckClusterDiskSizeAutoscalingExact(&cluster, &postgresql.DiskSizeAutoscaling{
					DiskSizeLimit:           datasize.ToBytes(20),
					PlannedUsageThreshold:   0,
					EmergencyUsageThreshold: 0,
				}),
				testAccCheckClusterPostgresqlConfigExact(&cluster, &pconfig.PostgresqlConfig17{
					PasswordEncryption: pconfig.PostgresqlConfig17_PASSWORD_ENCRYPTION_SCRAM_SHA_256,
				}, nil),
				testAccCheckClusterMaintenanceWindow(&cluster, &postgresql.MaintenanceWindow{
					Policy: &postgresql.MaintenanceWindow_Anytime{
						Anytime: &postgresql.AnytimeMaintenanceWindow{},
					},
				}),
				testAccCheckClusterSecurityGroupIdsExact(&cluster, nil),
			),
		},
		{
			Config: testAccMDBPGClusterBasic(resourceId, clusterName, descriptionBasic, environment, labels, version, resources),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionBasic)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact(environment)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("version"), knownvalue.StringExact(version)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("autofailover"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"data_lens":     knownvalue.Bool(false),
						"data_transfer": knownvalue.Bool(false),
						"web_sql":       knownvalue.Bool(false),
						"serverless":    knownvalue.Bool(false),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("performance_diagnostics"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"enabled":                      knownvalue.Bool(false),
						"sessions_sampling_interval":   knownvalue.Int64Exact(60),
						"statements_sampling_interval": knownvalue.Int64Exact(600),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"hours":   knownvalue.Int64Exact(0),
						"minutes": knownvalue.Int64Exact(0),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(7)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("pooler_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
					"pool_discard": knownvalue.Null(),
					"pooling_mode": knownvalue.StringExact(postgresql.ConnectionPoolerConfig_POOLING_MODE_UNSPECIFIED.String()),
				})),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("disk_size_autoscaling"), knownvalue.ObjectExact(map[string]knownvalue.Check{
					"disk_size_limit":           knownvalue.Int64Exact(20),
					"emergency_usage_threshold": knownvalue.Int64Exact(0),
					"planned_usage_threshold":   knownvalue.Int64Exact(0),
				})),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("postgresql_config"), knownvalue.ObjectExact(map[string]knownvalue.Check{
					"password_encryption": knownvalue.StringExact("PASSWORD_ENCRYPTION_SCRAM_SHA_256"),
				})),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"type": knownvalue.StringExact("ANYTIME"),
						"day":  knownvalue.Null(),
						"hour": knownvalue.Null(),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("security_group_ids"), knownvalue.SetSizeExact(0)),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 1),
				testAccCheckClusterLabelsExact(&cluster, map[string]string{"key": "value"}),
				testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
				testAccCheckClusterAutofailoverExact(&cluster, true),
				testAccCheckClusterDeletionProtectionExact(&cluster, false),
				testAccCheckClusterAccessExact(&cluster, &postgresql.Access{
					DataLens:     false,
					DataTransfer: false,
					WebSql:       false,
					Serverless:   false,
				}),
				testAccCheckClusterPerformanceDiagnosticsExact(&cluster, &postgresql.PerformanceDiagnostics{
					Enabled:                    false,
					SessionsSamplingInterval:   60,
					StatementsSamplingInterval: 600,
				}),
				testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
				testAccCheckClusterPoolerConfigExact(&cluster, nil),
				testAccCheckClusterDiskSizeAutoscalingExact(&cluster, &postgresql.DiskSizeAutoscaling{
					DiskSizeLimit:           datasize.ToBytes(20),
					PlannedUsageThreshold:   0,
					EmergencyUsageThreshold: 0,
				}),
				testAccCheckClusterPostgresqlConfigExact(&cluster, &pconfig.PostgresqlConfig17{
					PasswordEncryption: pconfig.PostgresqlConfig17_PASSWORD_ENCRYPTION_SCRAM_SHA_256,
				}, nil),
				testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
					Hours:   0,
					Minutes: 0,
				}),
				testAccCheckClusterMaintenanceWindow(&cluster, &postgresql.MaintenanceWindow{
					Policy: &postgresql.MaintenanceWindow_Anytime{
						Anytime: &postgresql.AnytimeMaintenanceWindow{},
					},
				}),
				testAccCheckClusterSecurityGroupIdsExact(&cluster, nil),
			),
		},
	}

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProviderFactories,
			CheckDestroy:             testAccCheckMDBPGClusterDestroy,
			Steps:                    stepsFullBasic[:],
		},
	)
}

// Test that a PostgreSQL HA Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_HostTests(t *testing.T) {
	t.Parallel()

	version := pgVersions[rand.Intn(len(pgVersions))]
	log.Printf("TestAccMDBPostgreSQLCluster_HostTests: version %s", version)
	var cluster postgresql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-postgresql-cluster-hosts-test")
	clusterResource := "yandex_mdb_postgresql_cluster_beta.cluster_host_tests"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccMDBPGClusterHostsStep0(clusterName, version, "# no hosts section specified"),
				ExpectError: regexp.MustCompile(`Error: Missing required argument`),
			},
			{
				Config: testAccMDBPGClusterHostsStep1(clusterName, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("na").AtMapKey("zone"), knownvalue.StringExact("ru-central1-a")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("zone"), knownvalue.StringExact("ru-central1-b")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nd").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 3),
					resource.TestCheckResourceAttrSet(clusterResource, `hosts.na.fqdn`),
					resource.TestCheckResourceAttrSet(clusterResource, `hosts.nb.fqdn`),
					resource.TestCheckResourceAttrSet(clusterResource, `hosts.nd.fqdn`),
				),
			},
			{
				Config: testAccMDBPGClusterHostsStep2(clusterName, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("zone"), knownvalue.StringExact("ru-central1-b")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 1),
				),
			},
			{
				Config: testAccMDBPGClusterHostsStep3(clusterName, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("na").AtMapKey("zone"), knownvalue.StringExact("ru-central1-a")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("na").AtMapKey("assign_public_ip"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("zone"), knownvalue.StringExact("ru-central1-b")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("assign_public_ip"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nd").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nd").AtMapKey("assign_public_ip"), knownvalue.Bool(true)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 3),
				),
			},
			{
				Config: testAccMDBPGClusterHostsStep4(clusterName, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("na").AtMapKey("zone"), knownvalue.StringExact("ru-central1-a")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("na").AtMapKey("assign_public_ip"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("zone"), knownvalue.StringExact("ru-central1-b")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("assign_public_ip"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nd").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nd").AtMapKey("assign_public_ip"), knownvalue.Bool(false)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 3),
				),
			},
		},
	})
}

// Test that a PostgreSQL HA Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_HostSpecialCaseTests(t *testing.T) {
	t.Parallel()

	version := pgVersions[rand.Intn(len(pgVersions))]
	log.Printf("TestAccMDBPostgreSQLCluster_HostTests: version %s", version)
	var cluster postgresql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-postgresql-cluster-hosts-special-test")
	clusterResource := "yandex_mdb_postgresql_cluster_beta.cluster_hosts_special_case_tests"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBPGClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPGClusterHostsSpecialCaseStep1(clusterName, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("lol").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("kek").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("cheburek").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 3),
				),
			},
			{
				Config: testAccMDBPGClusterHostsSpecialCaseStep2(clusterName, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("lol").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("lol").AtMapKey("assign_public_ip"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("kek").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("kek").AtMapKey("assign_public_ip"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("cheburek").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("cheburek").AtMapKey("assign_public_ip"), knownvalue.Bool(false)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBPostgreSQLCluster(clusterResource, &cluster, 3),
				),
			},
		},
	})
}

func testAccCheckMDBPGClusterDestroy(s *terraform.State) error {
	config := test.AccProvider.(*provider.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_postgresql_cluster_beta" {
			continue
		}

		_, err := config.SDK.MDB().PostgreSQL().Cluster().Get(context.Background(), &postgresql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("PostgreSQL Cluster still exists")
		}
	}

	return nil
}

func testAccCheckExistsAndParseMDBPostgreSQLCluster(n string, r *postgresql.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*provider.Provider).GetConfig()

		found, err := config.SDK.MDB().PostgreSQL().Cluster().Get(context.Background(), &postgresql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("PostgreSQL Cluster not found")
		}

		*r = *found

		resp, err := config.SDK.MDB().PostgreSQL().Cluster().ListHosts(context.Background(), &postgresql.ListClusterHostsRequest{
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

func testAccCheckClusterLabelsExact(r *postgresql.Cluster, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.Labels, expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched labels.\nActual:   %+v\nExpected: %+v", r.Name, r.Labels, expected)
	}
}

func testAccCheckClusterAutofailoverExact(r *postgresql.Cluster, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if r.Config.GetAutofailover().GetValue() == expected {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config autofailover.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetAutofailover().GetValue(), expected)
	}
}

func testAccCheckClusterDeletionProtectionExact(r *postgresql.Cluster, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if r.GetDeletionProtection() == expected {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config deletion_protection.\nActual:   %+v\nExpected: %+v", r.Name, r.GetDeletionProtection(), expected)
	}
}

func testAccCheckClusterSecurityGroupIdsExact(r *postgresql.Cluster, expectedResourceNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rootModule := s.RootModule()

		expectedResourceIds := make([]string, len(expectedResourceNames))
		for idx, resName := range expectedResourceNames {
			expectedResourceIds[idx] = rootModule.Resources[resName].Primary.ID
		}

		if len(r.GetSecurityGroupIds()) == 0 && len(expectedResourceIds) == 0 {
			return nil
		}

		sort.Strings(r.SecurityGroupIds)
		sort.Strings(expectedResourceIds)

		if reflect.DeepEqual(expectedResourceIds, r.SecurityGroupIds) {
			return nil
		}
		return fmt.Errorf(
			"Cluster %s has mismatched config security_group_ids.\nActual:   %+v\nExpected: %+v", r.Name, r.GetSecurityGroupIds(), expectedResourceIds,
		)
	}
}

func testAccCheckClusterAccessExact(r *postgresql.Cluster, expected *postgresql.Access) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetAccess(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config access.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetAccess(), expected)
	}
}

func testAccCheckClusterPerformanceDiagnosticsExact(r *postgresql.Cluster, expected *postgresql.PerformanceDiagnostics) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetPerformanceDiagnostics(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config performance_diagnostics.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetPerformanceDiagnostics(), expected)
	}
}

func testAccCheckClusterBackupRetainPeriodDaysExact(r *postgresql.Cluster, expected *wrapperspb.Int64Value) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetBackupRetainPeriodDays(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config backup_retain_period_days.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetBackupWindowStart(), expected)
	}
}

func testAccCheckClusterPoolerConfigExact(r *postgresql.Cluster, expected *postgresql.ConnectionPoolerConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.Config.GetPoolerConfig(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched pooler_config.\nActual:   %+v\nExpected: %+v", r.Name, r.Config.GetPoolerConfig(), expected)
	}
}

func testAccCheckClusterBackupWindowStartExact(r *postgresql.Cluster, expected *timeofday.TimeOfDay) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetBackupWindowStart(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config backup_window_start.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetBackupWindowStart(), expected)
	}
}

func testAccCheckClusterPostgresqlConfigExact(r *postgresql.Cluster, expectedUserConfig interface{}, checkFields []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var cmpObj interface{}
		switch expectedUserConfig.(type) {
		case *pconfig.PostgresqlConfig10:
			cmpObj = r.GetConfig().GetPostgresqlConfig_10().GetUserConfig()
		case *pconfig.PostgresqlConfig10_1C:
			cmpObj = r.GetConfig().GetPostgresqlConfig_10_1C().GetUserConfig()
		case *pconfig.PostgresqlConfig11:
			cmpObj = r.GetConfig().GetPostgresqlConfig_11().GetUserConfig()
		case *pconfig.PostgresqlConfig11_1C:
			cmpObj = r.GetConfig().GetPostgresqlConfig_11_1C().GetUserConfig()
		case *pconfig.PostgresqlConfig12:
			cmpObj = r.GetConfig().GetPostgresqlConfig_12().GetUserConfig()
		case *pconfig.PostgresqlConfig12_1C:
			cmpObj = r.GetConfig().GetPostgresqlConfig_12_1C().GetUserConfig()
		case *pconfig.PostgresqlConfig13:
			cmpObj = r.GetConfig().GetPostgresqlConfig_13().GetUserConfig()
		case *pconfig.PostgresqlConfig13_1C:
			cmpObj = r.GetConfig().GetPostgresqlConfig_13_1C().GetUserConfig()
		case *pconfig.PostgresqlConfig14:
			cmpObj = r.GetConfig().GetPostgresqlConfig_14().GetUserConfig()
		case *pconfig.PostgresqlConfig14_1C:
			cmpObj = r.GetConfig().GetPostgresqlConfig_14_1C().GetUserConfig()
		case *pconfig.PostgresqlConfig15:
			cmpObj = r.GetConfig().GetPostgresqlConfig_15().GetUserConfig()
		case *pconfig.PostgresqlConfig15_1C:
			cmpObj = r.GetConfig().GetPostgresqlConfig_15_1C().GetUserConfig()
		case *pconfig.PostgresqlConfig16:
			cmpObj = r.GetConfig().GetPostgresqlConfig_16().GetUserConfig()
		case *pconfig.PostgresqlConfig17:
			cmpObj = r.GetConfig().GetPostgresqlConfig_17().GetUserConfig()
		default:
			return fmt.Errorf("unsupported expectedUserConfig type %T", expectedUserConfig)
		}

		actual := reflect.ValueOf(cmpObj)
		expected := reflect.ValueOf(expectedUserConfig)

		if actual.IsNil() != expected.IsNil() {
			return fmt.Errorf("Cluster %s has mismatched postgresql config existence.\nActual:   %+v\nExpected: %+v", r.Name, actual.IsNil(), expected.IsNil())
		}

		actual = actual.Elem()
		expected = expected.Elem()

		for _, field := range checkFields {
			actualF := actual.FieldByName(field).Interface()
			expectedF := expected.FieldByName(field).Interface()
			if !reflect.DeepEqual(actualF, expectedF) {
				return fmt.Errorf("Cluster %s has mismatched postgresql config field %s.\nActual:   %+v, %T\nExpected: %+v, %T", r.Name, field, actualF, actualF, expectedF, expectedF)
			}
		}

		return nil
	}
}

func testAccCheckClusterDiskSizeAutoscalingExact(r *postgresql.Cluster, expected *postgresql.DiskSizeAutoscaling) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetDiskSizeAutoscaling(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config disk_size_autoscaling.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetDiskSizeAutoscaling(), expected)
	}
}

func testAccCheckClusterMaintenanceWindow(r *postgresql.Cluster, expected *postgresql.MaintenanceWindow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetMaintenanceWindow(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched maintenance_window.\nActual:   %+v\nExpected: %+v", r.Name, r.GetMaintenanceWindow(), expected)
	}
}

func testAccCheckClusterHasResources(r *postgresql.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccMDBPGClusterBasic(resourceId, name, description, environment, labels, version, resources string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster_beta" "%s" {
  name        = "%s"
  description = "%s"
  environment = "%s"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id

  labels = {
%s
  }

  hosts = {
    "na" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
    }
  }

  config {
    version = "%s"
    resources {
      %s
    }
  }
}
`, resourceId, name, description, environment, labels, version, resources)
}

func testAccMDBPGClusterFull(
	resourceId, clusterName, description, environment, labels,
	version, resources,
	access,
	performanceDiagnostics,
	backupWindowStart,
	poolerConfig,
	dsa,
	pgConfig,
	maintenanceWindow string, backupRetainPeriodDays int, autofailover, deletionProtection bool, confSecurityGroupIds []string,
) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster_beta" "%s" {
  name        = "%s"
  description = "%s"
  environment = "%s" 
  network_id  = yandex_vpc_network.mdb-pg-test-net.id

  labels = {
%s
  }

  hosts = {
    "host" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
    }
  }

  config {
    version = "%s"
    resources {
      %s
    }
    autofailover = %t
    access = {
    %s
    }
	performance_diagnostics = {
		%s
	}
	backup_retain_period_days = %d
	backup_window_start = {
	%s
	}

	pooler_config = {
		%s
	}
	
	disk_size_autoscaling = {
		%s
	}

	postgresql_config = {
		%s
	}
  }
  
  maintenance_window = {
	%s
  }

  deletion_protection = %t
  security_group_ids = [%s]

}
`, resourceId, clusterName, description, environment,
		labels, version, resources, autofailover, access,
		performanceDiagnostics, backupRetainPeriodDays, backupWindowStart, poolerConfig, dsa, pgConfig,
		maintenanceWindow, deletionProtection, strings.Join(confSecurityGroupIds, ", "),
	)
}

func testAccMDBPGClusterHostsStep0(name, version, hosts string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster_beta" "cluster_host_tests" {
  name        = "%s"
  description = "PostgreSQL Cluster Hosts Terraform Test"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id
  environment = "PRESTABLE"

  config {
    version = "%s"
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
    }
  }
%s
}
`, name, version, hosts)
}

// Init hosts configuration
func testAccMDBPGClusterHostsStep1(name, version string) string {
	return testAccMDBPGClusterHostsStep0(name, version, `
  hosts = {
    "na" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
    }
    "nb" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
    }
    "nd" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-d.id
    }
  }
`)
}

// Drop some hosts
func testAccMDBPGClusterHostsStep2(name, version string) string {
	return testAccMDBPGClusterHostsStep0(name, version, `
  hosts = {
    "nb" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
    }
  }
`)
}

// Add some hosts back with all possible options
func testAccMDBPGClusterHostsStep3(name, version string) string {
	return testAccMDBPGClusterHostsStep0(name, version, `
  hosts = {
    "na" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
      assign_public_ip = true
    }
    "nb" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
    }
    "nd" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-d.id
      assign_public_ip = true
    }
  }
`)
}

// Update Hosts
func testAccMDBPGClusterHostsStep4(name, version string) string {
	return testAccMDBPGClusterHostsStep0(name, version, `
  hosts = {
    "na" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
      assign_public_ip = false
    }
    "nb" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-b.id
      assign_public_ip = true
    }
    "nd" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-d.id
      assign_public_ip = false
    }
  }
`)
}

func testAccMDBPGClusterHostsSpecialCaseStep0(name, version, hosts string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster_beta" "cluster_hosts_special_case_tests" {
  name        = "%s"
  description = "PostgreSQL Cluster Hosts Terraform Test"
  network_id  = yandex_vpc_network.mdb-pg-test-net.id
  environment = "PRESTABLE"

  config {
    version = "%s"
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
    }
  }
%s
}
`, name, version, hosts)
}

// Init hosts special case configuration
func testAccMDBPGClusterHostsSpecialCaseStep1(name, version string) string {
	return testAccMDBPGClusterHostsSpecialCaseStep0(name, version, `
  hosts = {
    "lol" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-d.id
    }
    "kek" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-d.id
    }
    "cheburek" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-d.id
    }
  }
`)
}

// Change some options
func testAccMDBPGClusterHostsSpecialCaseStep2(name, version string) string {
	return testAccMDBPGClusterHostsSpecialCaseStep0(name, version, `
  hosts = {
    "lol" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-d.id
    }
    "kek" = {
      zone      = "ru-central1-d"
	  assign_public_ip = true
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-d.id
    }
    "cheburek" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-pg-test-subnet-d.id
    }
  }
`)
}

// func testAccMDBPGClusterConfigHANamedSwitchMaster(name, version string) string
// func testAccMDBPGClusterConfigHANamedChangePublicIP(name, version string) string
// func testAccMDBPGClusterConfigHANamedWithCascade(name, version string) string
