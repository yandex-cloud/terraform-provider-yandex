package mdb_sharded_postgresql_cluster_test

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	defaultMDBPageSize = 1000

	yandexMDBShardedPostgreSQLClusterResourceType = "yandex_mdb_sharded_postgresql_cluster"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccMDBShardedPostgreSQLCluster_basic(t *testing.T) {
	t.Parallel()

	var cluster spqr.Cluster
	clusterName := acctest.RandomWithPrefix("tf-sharded-postgresql-cluster-basic")
	resourceId := "cluster_basic_test"
	clusterResource := "yandex_mdb_sharded_postgresql_cluster." + resourceId
	description := "Sharded Postgresql Cluster Terraform Test Basic"
	descriptionUpdated := fmt.Sprintf("%s Updated", description)
	folderID := testhelpers.GetExampleFolderID()

	resources := `
		resource_preset_id = "s2.micro"
		disk_size          = 10
		disk_type_id       = "network-hdd"
	`

	resourcesUpdated := `
		resource_preset_id = "s2.micro"
		disk_size		   = 12
		disk_type_id	   = "network-hdd"
	`

	labels := `
		key1 = "value1"
		key2 = "value2"
		key3 = "value3"
    `
	labelsUpdated := `
		key4 = "value4"
    `

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBShardedPostgreSQLClusterDestroy,
		Steps: []resource.TestStep{
			// Create Sharded Postgresql Cluster
			{
				Config: testAccMDBShardedPostgreSQLClusterBasic(resourceId, clusterName, description, "PRESTABLE", labels, resources),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact("PRESTABLE")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
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
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"type": knownvalue.StringExact("ANYTIME"),
						"day":  knownvalue.Null(),
						"hour": knownvalue.Null(),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-hdd", 10*1024*1024*1024),
					testAccCheckClusterDeletionProtectionExact(&cluster, false),
					testAccCheckClusterAccessExact(&cluster, &spqr.Access{
						DataLens:     false,
						DataTransfer: false,
						WebSql:       false,
						Serverless:   false,
					}),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   0,
						Minutes: 0,
					}),
					testAccCheckClusterMaintenanceWindow(&cluster, &spqr.MaintenanceWindow{
						Policy: &spqr.MaintenanceWindow_Anytime{
							Anytime: &spqr.AnytimeMaintenanceWindow{},
						},
					}),
				),
			},
			mdbShardedPostgreSQLClusterImportStep(clusterResource),
			// Update Sharded Postgresql Cluster
			{
				Config: testAccMDBShardedPostgreSQLClusterBasic(resourceId, clusterName, descriptionUpdated, "PRESTABLE", labelsUpdated, resourcesUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionUpdated)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact("PRESTABLE")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(false),
							"data_transfer": knownvalue.Bool(false),
							"web_sql":       knownvalue.Bool(false),
							"serverless":    knownvalue.Bool(false),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"hours":   knownvalue.Int64Exact(0),
						"minutes": knownvalue.Int64Exact(0),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(7)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"type": knownvalue.StringExact("ANYTIME"),
						"day":  knownvalue.Null(),
						"hour": knownvalue.Null(),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key4": "value4"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-hdd", 12*1024*1024*1024),
					testAccCheckClusterDeletionProtectionExact(&cluster, false),
					testAccCheckClusterAccessExact(&cluster, &spqr.Access{
						DataLens:     false,
						DataTransfer: false,
						WebSql:       false,
						Serverless:   false,
					}),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   0,
						Minutes: 0,
					}),
					testAccCheckClusterMaintenanceWindow(&cluster, &spqr.MaintenanceWindow{
						Policy: &spqr.MaintenanceWindow_Anytime{
							Anytime: &spqr.AnytimeMaintenanceWindow{},
						},
					}),
				),
			},
			mdbShardedPostgreSQLClusterImportStep(clusterResource),
		},
	})
}

func TestAccMDBShardedPostgreSQLCluster_full(t *testing.T) {
	t.Parallel()

	log.Printf("TestAccMDBShardedPostgreSQLCluster_full")
	var cluster spqr.Cluster
	clusterName := acctest.RandomWithPrefix("tf-sharded-postgresql-cluster-full")

	resourceId := "cluster_full_test"
	clusterResource := "yandex_mdb_sharded_postgresql_cluster." + resourceId

	description := "Sharded Postgresql Cluster Terraform Test Full"
	descriptionUpdated := fmt.Sprintf("%s Updated", description)
	folderID := testhelpers.GetExampleFolderID()

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
		data_transfer = false
		web_sql = false
		serverless = false
		data_lens = false
	`

	accessUpdated := `
		serverless = false
		data_lens = false
		data_transfer = false
		web_sql = false
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

	shardedPostgresqlConfig := `
		router = {
			config = {
				show_notice_messages = false
			}
			resources = {
				resource_preset_id = "s2.micro"
				disk_size          = 10
				disk_type_id       = "network-ssd"
			}
		}
	`

	shardedPostgresqlConfigUpdated := `
		router = {
			config = {
				show_notice_messages = true
			}
			resources = {
				resource_preset_id = "s2.micro"
				disk_size          = 16
				disk_type_id       = "network-ssd"
			}
		}
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
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBShardedPostgreSQLClusterDestroy,
		Steps: []resource.TestStep{
			// Create Sharded Postgresql Cluster
			{
				Config: testAccMDBShardedPostgreSQLClusterFull(
					resourceId, clusterName, description,
					environment, labels,
					access,
					backupWindowStart,
					shardedPostgresqlConfig,
					maintenanceWindow,
					backupRetainPeriodDays, false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact(environment)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(false),
							"data_transfer": knownvalue.Bool(false),
							"web_sql":       knownvalue.Bool(false),
							"serverless":    knownvalue.Bool(false),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(int64(backupRetainPeriodDays))),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"hours":   knownvalue.Int64Exact(5),
							"minutes": knownvalue.Int64Exact(4),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"type": knownvalue.StringExact("ANYTIME"),
							"day":  knownvalue.Null(),
							"hour": knownvalue.Null(),
						},
					)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckClusterDeletionProtectionExact(&cluster, false),
					testAccCheckClusterAccessExact(&cluster, &spqr.Access{
						DataLens:     false,
						DataTransfer: false,
						WebSql:       false,
						Serverless:   false,
					}),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(int64(backupRetainPeriodDays))),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   5,
						Minutes: 4,
					}),
					testAccCheckClusterShardedPostgreSQLConfigExact(&cluster, &spqr.SPQRConfig{
						Router: &spqr.RouterConfig{
							Config: &spqr.RouterSettings{
								ShowNoticeMessages: wrapperspb.Bool(false),
							},
							Resources: &spqr.Resources{
								ResourcePresetId: "s2.micro",
								DiskSize:         10,
								DiskTypeId:       "network-ssd",
							},
						},
					}, []string{}),
					testAccCheckClusterMaintenanceWindow(&cluster, &spqr.MaintenanceWindow{
						Policy: &spqr.MaintenanceWindow_Anytime{
							Anytime: &spqr.AnytimeMaintenanceWindow{},
						},
					}),
				),
			},
			mdbShardedPostgreSQLClusterImportStep(clusterResource),
			{
				Config: testAccMDBShardedPostgreSQLClusterFull(
					resourceId, clusterName, descriptionUpdated,
					environment, labelsUpdated, accessUpdated,
					backupWindowStartUpdated,
					shardedPostgresqlConfigUpdated,
					maintenanceWindowUpdated,
					backupRetainPeriodDaysUpdated, false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionUpdated)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact(environment)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(false),
							"data_transfer": knownvalue.Bool(false),
							"web_sql":       knownvalue.Bool(false),
							"serverless":    knownvalue.Bool(false),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_retain_period_days"), knownvalue.Int64Exact(int64(backupRetainPeriodDaysUpdated))),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("backup_window_start"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"hours":   knownvalue.Int64Exact(10),
							"minutes": knownvalue.Int64Exact(3),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"type": knownvalue.StringExact("WEEKLY"),
							"day":  knownvalue.StringExact("MON"),
							"hour": knownvalue.Int64Exact(5),
						},
					)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key4": "value4"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 16*1024*1024*1024),
					testAccCheckClusterDeletionProtectionExact(&cluster, false),
					testAccCheckClusterAccessExact(&cluster, &spqr.Access{
						DataLens:     false,
						DataTransfer: false,
						WebSql:       false,
						Serverless:   false,
					}),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(int64(backupRetainPeriodDaysUpdated))),
					testAccCheckClusterShardedPostgreSQLConfigExact(&cluster, &spqr.SPQRConfig{
						Router: &spqr.RouterConfig{
							Config: &spqr.RouterSettings{
								ShowNoticeMessages: wrapperspb.Bool(true),
							},
							Resources: &spqr.Resources{
								ResourcePresetId: "s2.micro",
								DiskSize:         16,
								DiskTypeId:       "network-ssd",
							},
						},
					}, []string{}),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   10,
						Minutes: 3,
					}),
					testAccCheckClusterMaintenanceWindow(&cluster, &spqr.MaintenanceWindow{
						Policy: &spqr.MaintenanceWindow_WeeklyMaintenanceWindow{
							WeeklyMaintenanceWindow: &spqr.WeeklyMaintenanceWindow{
								Day:  spqr.WeeklyMaintenanceWindow_MON,
								Hour: 5,
							},
						},
					}),
				),
			},
			mdbShardedPostgreSQLClusterImportStep(clusterResource),
		},
	})
}

func TestAccMDBShardedPostgreSQLCluster_HostTests(t *testing.T) {
	t.Parallel()

	log.Printf("TestAccMDBShardedPostgreSQLCluster_HostTests")
	var cluster spqr.Cluster
	clusterName := acctest.RandomWithPrefix("tf-sharded_postgresql-cluster-hosts-test")
	clusterResource := "yandex_mdb_sharded_postgresql_cluster.cluster_host_tests"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBShardedPostgreSQLClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccMDBShardedPostgreSQLClusterHostsStep0(clusterName, "# no hosts section specified"),
				ExpectError: regexp.MustCompile(`Error: Missing required argument`),
			},
			{
				Config: testAccMDBShardedPostgreSQLClusterHostsStep1(clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("na").AtMapKey("zone"), knownvalue.StringExact("ru-central1-a")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("zone"), knownvalue.StringExact("ru-central1-b")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nd").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(clusterResource, &cluster, 3),
					resource.TestCheckResourceAttrSet(clusterResource, `hosts.na.fqdn`),
					resource.TestCheckResourceAttrSet(clusterResource, `hosts.nb.fqdn`),
					resource.TestCheckResourceAttrSet(clusterResource, `hosts.nd.fqdn`),
				),
			},
			{
				Config: testAccMDBShardedPostgreSQLClusterHostsStep2(clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("zone"), knownvalue.StringExact("ru-central1-b")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(clusterResource, &cluster, 2),
				),
			},
			{
				Config: testAccMDBShardedPostgreSQLClusterHostsStep3(clusterName),
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
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(clusterResource, &cluster, 3),
				),
			},
			{
				Config: testAccMDBShardedPostgreSQLClusterHostsStep4(clusterName),
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
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(clusterResource, &cluster, 3),
				),
			},
		},
	})
}

func mdbShardedPostgreSQLClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"health", // volatile value
			"hosts",  // volatile value
			"config.sharded_postgresql_config.common.%",
		},
	}
}

func testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(n string, r *spqr.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testhelpers.AccProvider.(*provider.Provider).GetConfig()

		found, err := config.SDK.MDB().SPQR().Cluster().Get(context.Background(), &spqr.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Sharded Postgresql Cluster not found")
		}

		*r = *found

		resp, err := config.SDK.MDB().SPQR().Cluster().ListHosts(context.Background(), &spqr.ListClusterHostsRequest{
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

func testAccCheckClusterShardedPostgreSQLConfigExact(r *spqr.Cluster, expectedUserConfig interface{}, checkFields []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cmpObj := r.Config.SpqrConfig

		actual := reflect.ValueOf(cmpObj)
		expected := reflect.ValueOf(expectedUserConfig)

		if actual.IsNil() != expected.IsNil() {
			return fmt.Errorf("Cluster %s has mismatched sharded_postgresql config existence.\nActual:   %+v\nExpected: %+v", r.Name, actual.IsNil(), expected.IsNil())
		}

		actual = actual.Elem()
		expected = expected.Elem()

		for _, field := range checkFields {
			actualF := actual.FieldByName(field).Interface()
			expectedF := expected.FieldByName(field).Interface()
			if !reflect.DeepEqual(actualF, expectedF) {
				return fmt.Errorf("Cluster %s has mismatched sharded_postgresql config field %s.\nActual:   %+v, %T\nExpected: %+v, %T", r.Name, field, actualF, actualF, expectedF, expectedF)
			}
		}

		return nil
	}
}

func testAccCheckClusterLabelsExact(r *spqr.Cluster, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.Labels, expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched labels.\nActual:   %+v\nExpected: %+v", r.Name, r.Labels, expected)
	}
}

func testAccCheckClusterHasResources(r *spqr.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		check := func(rs *spqr.Resources) error {
			if rs == nil {
				return nil
			}
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
		if r.Config.SpqrConfig.Router != nil {
			if err := check(r.Config.SpqrConfig.Router.Resources); err != nil {
				return err
			}
		}

		if r.Config.SpqrConfig.Coordinator != nil {
			if err := check(r.Config.SpqrConfig.Coordinator.Resources); err != nil {
				return err
			}
		}

		if r.Config.SpqrConfig.Infra != nil {
			if err := check(r.Config.SpqrConfig.Infra.Resources); err != nil {
				return err
			}
		}

		if r.Config.SpqrConfig.Postgresql != nil {
			if err := check(r.Config.SpqrConfig.Postgresql.Resources); err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccCheckClusterDeletionProtectionExact(r *spqr.Cluster, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if r.GetDeletionProtection() == expected {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config deletion_protection.\nActual:   %+v\nExpected: %+v", r.Name, r.GetDeletionProtection(), expected)
	}
}

func testAccCheckClusterAccessExact(r *spqr.Cluster, expected *spqr.Access) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetAccess(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config access.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetAccess(), expected)
	}
}

func testAccCheckClusterBackupRetainPeriodDaysExact(r *spqr.Cluster, expected *wrapperspb.Int64Value) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetBackupRetainPeriodDays(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config backup_retain_period_days.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetBackupWindowStart(), expected)
	}
}

func testAccCheckClusterBackupWindowStartExact(r *spqr.Cluster, expected *timeofday.TimeOfDay) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetBackupWindowStart(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config backup_window_start.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetBackupWindowStart(), expected)
	}
}

func testAccCheckClusterMaintenanceWindow(r *spqr.Cluster, expected *spqr.MaintenanceWindow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetMaintenanceWindow(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched maintenance_window.\nActual:   %+v\nExpected: %+v", r.Name, r.GetMaintenanceWindow(), expected)
	}
}

func testAccCheckMDBShardedPostgreSQLClusterDestroy(s *terraform.State) error {
	config := testhelpers.AccProvider.(*provider.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != yandexMDBShardedPostgreSQLClusterResourceType {
			continue
		}

		_, err := config.SDK.MDB().SPQR().Cluster().Get(context.Background(), &spqr.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Sharded Postgresql Cluster still exists")
		}
	}

	return nil
}

const shardedPostgreSQLVPCDependencies = `
resource "yandex_vpc_network" "mdb-sharded_postgresql-test-net" {}

resource "yandex_vpc_subnet" "mdb-sharded_postgresql-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.mdb-sharded_postgresql-test-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-sharded_postgresql-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.mdb-sharded_postgresql-test-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-sharded_postgresql-test-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.mdb-sharded_postgresql-test-net.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}

`

func testAccMDBShardedPostgreSQLClusterBasic(resourceId, name, description, environment, labels, resources string) string {
	return fmt.Sprintf(shardedPostgreSQLVPCDependencies+`
resource "yandex_mdb_sharded_postgresql_cluster" "%s" {
	name        = "%s"
	description = "%s"
	environment = "%s"
	network_id  = yandex_vpc_network.mdb-sharded_postgresql-test-net.id

	labels = {
%s
	}

	hosts = {
		"na" = {
			zone      = "ru-central1-a"
			subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-a.id
			type	  = "ROUTER"
		}
	}

	config = {
		sharded_postgresql_config = {
			router = {
				resources = {
%s
				}
			}
		}
	}
}
`, resourceId, name, description, environment, labels, resources)
}

func testAccMDBShardedPostgreSQLClusterFull(
	resourceId, clusterName, description, environment, labels,
	access,
	backupWindowStart,
	cfg,
	maintenanceWindow string, backupRetainPeriodDays int, deletionProtection bool,
) string {
	return fmt.Sprintf(shardedPostgreSQLVPCDependencies+`
resource "yandex_mdb_sharded_postgresql_cluster" "%s" {
	name        = "%s"
	description = "%s"
	environment = "%s"
	network_id  = yandex_vpc_network.mdb-sharded_postgresql-test-net.id

	labels = {
%s
	}

	hosts = {
		"host" = {
			zone      = "ru-central1-a"
			subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-a.id
			type	  = "ROUTER"
		}
	}

	config = {
		access = {
%s
    	}
		backup_retain_period_days = %d
		backup_window_start = {
%s
		}

		sharded_postgresql_config = {
%s
		}
  }

	maintenance_window = {
%s
	}

  deletion_protection = %t
}
`, resourceId, clusterName, description, environment,
		labels, access,
		backupRetainPeriodDays, backupWindowStart, cfg,
		maintenanceWindow, deletionProtection,
	)
}

func testAccMDBShardedPostgreSQLClusterHostsStep0(name, hosts string) string {
	return fmt.Sprintf(shardedPostgreSQLVPCDependencies+`
resource "yandex_mdb_sharded_postgresql_cluster" "cluster_host_tests" {
	name        = "%s"
	description = "Sharded Postgresql Cluster Hosts Terraform Test"
	network_id  = yandex_vpc_network.mdb-sharded_postgresql-test-net.id
	environment = "PRESTABLE"

	config = {
		sharded_postgresql_config = {
			router = {
				resources = {
					resource_preset_id = "s2.micro"
					disk_size          = 10
					disk_type_id       = "network-ssd"
				}
			}
			coordinator = {
				resources = {
					resource_preset_id = "s2.micro"
					disk_size          = 10
					disk_type_id       = "network-ssd"
				}
			}
		}
	}
%s
}
`, name, hosts)
}

// Init hosts configuration
func testAccMDBShardedPostgreSQLClusterHostsStep1(name string) string {
	return testAccMDBShardedPostgreSQLClusterHostsStep0(name, `
	hosts = {
		"na" = {
    		zone      = "ru-central1-a"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-a.id
			type	  = "ROUTER"
    	}
		"nb" = {
    		zone      = "ru-central1-b"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-b.id
			type	  = "ROUTER"
		}
		"nd" = {
    		zone      = "ru-central1-d"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-d.id
			type	  = "COORDINATOR"
		}
	}
`)
}

// Drop some hosts
func testAccMDBShardedPostgreSQLClusterHostsStep2(name string) string {
	return testAccMDBShardedPostgreSQLClusterHostsStep0(name, `
	hosts = {
		"nb" = {
    		zone      = "ru-central1-b"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-b.id
			type	  = "ROUTER"
		}
		"nd" = {
    		zone      = "ru-central1-d"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-d.id
			type	  = "COORDINATOR"
		}
  	}
`)
}

// Add some hosts back with all possible options
func testAccMDBShardedPostgreSQLClusterHostsStep3(name string) string {
	return testAccMDBShardedPostgreSQLClusterHostsStep0(name, `
	hosts = {
		"na" = {
    		zone      = "ru-central1-a"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-a.id
    		assign_public_ip = true
			type	  = "ROUTER"
		}
		"nb" = {
    		zone      = "ru-central1-b"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-b.id
			type	  = "ROUTER"
    	}
		"nd" = {
    		zone      = "ru-central1-d"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-d.id
    		assign_public_ip = true
			type	  = "COORDINATOR"
		}
  	}
`)
}

// Update Hosts
func testAccMDBShardedPostgreSQLClusterHostsStep4(name string) string {
	return testAccMDBShardedPostgreSQLClusterHostsStep0(name, `
	hosts = {
		"na" = {
    		zone      = "ru-central1-a"
			subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-a.id
    		assign_public_ip = false
			type	  = "ROUTER"
    	}
    	"nb" = {
			zone      = "ru-central1-b"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-b.id
    		assign_public_ip = true
			type	  = "ROUTER"
		}
		"nd" = {
    		zone      = "ru-central1-d"
    		subnet_id = yandex_vpc_subnet.mdb-sharded_postgresql-test-subnet-d.id
    		assign_public_ip = false
			type	  = "COORDINATOR"
		}
  	}
`)
}
