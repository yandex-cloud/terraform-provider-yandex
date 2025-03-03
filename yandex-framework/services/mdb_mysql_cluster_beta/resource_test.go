package mdb_mysql_cluster_beta_test

import (
	"context"
	"fmt"
	"log"
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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"

	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	defaultMDBPageSize                 = 1000
	msResource                         = "yandex_mdb_mysql_cluster_beta.foo"
	msRestoreBackupId                  = "c9qrbucrcvm6a50tblv2:c9q698sst87e4vhkvrsm"
	yandexMDBMySQLClusterCreateTimeout = 30 * time.Minute // TODO refactor
	yandexMDBMySQLClusterDeleteTimeout = 15 * time.Minute
	yandexMDBMySQLClusterUpdateTimeout = 60 * time.Minute
)

const msVPCDependencies = `
resource "yandex_vpc_network" "mdb-ms-test-net" {}

resource "yandex_vpc_subnet" "mdb-ms-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.mdb-ms-test-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-ms-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.mdb-ms-test-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-ms-test-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.mdb-ms-test-net.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_security_group" "sgroup1" {
  description = "Test security group 1"
  network_id  = yandex_vpc_network.mdb-ms-test-net.id
}

resource "yandex_vpc_security_group" "sgroup2" {
  description = "Test security group 2"
  network_id  = yandex_vpc_network.mdb-ms-test-net.id
}

`

var (
	msVersions = [...]string{"5.7", "8.0"}
)

func init() {
	resource.AddTestSweepers("yandex_mdb_mysql_cluster_beta", &resource.Sweeper{
		Name: "yandex_mdb_mysql_cluster_beta",
		F:    testSweepMDBMySQLCluster,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepMDBMySQLCluster(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.MDB().MySQL().Cluster().List(context.Background(), &mysql.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting MySQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBMySQLCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep MySQL cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBMySQLCluster(conf *config.Config, id string) bool {
	return test.SweepWithRetry(sweepMDBMySQLClusterOnce, conf, "MySQL cluster", id)
}

func sweepMDBMySQLClusterOnce(conf *config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexMDBMySQLClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	op, err := conf.SDK.MDB().MySQL().Cluster().Update(ctx, &mysql.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = test.HandleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(test.ErrorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.SDK.MDB().MySQL().Cluster().Delete(ctx, &mysql.DeleteClusterRequest{
		ClusterId: id,
	})
	return test.HandleSweepOperation(ctx, conf, op, err)
}

func mdbMySQLClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"health", // volatile value
			"hosts",  // volatile value
		},
	}
}

// Test that a MySQL Cluster can be created, updated and destroyed
func TestAccMDBMySQLCluster_basic(t *testing.T) {
	t.Parallel()

	version := msVersions[0]
	versionUpdate := msVersions[1]

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

	log.Printf("TestAccMDBMySQLCluster_basic: version %s", version)
	var cluster mysql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-mysql-cluster-basic")
	resourceId := "cluster_basic_test"
	clusterResource := "yandex_mdb_mysql_cluster_beta." + resourceId
	description := "MySQL Cluster Terraform Test Basic"
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

	firstBasicStateChecks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(description)),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("version"), knownvalue.StringExact(version)),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("performance_diagnostics"), knownvalue.ObjectExact(map[string]knownvalue.Check{
			"enabled":                      knownvalue.Bool(false),
			"sessions_sampling_interval":   knownvalue.Int64Exact(60),
			"statements_sampling_interval": knownvalue.Int64Exact(600),
		})),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("access"), knownvalue.ObjectExact(
			map[string]knownvalue.Check{
				"data_lens":     knownvalue.Bool(false),
				"data_transfer": knownvalue.Bool(false),
				"web_sql":       knownvalue.Bool(false),
			},
		)),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_retain_period_days"), knownvalue.Int64Exact(7)),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_window_start"), knownvalue.ObjectExact(map[string]knownvalue.Check{
			"hours":   knownvalue.Int64Exact(0),
			"minutes": knownvalue.Int64Exact(0),
		})),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(map[string]knownvalue.Check{
			"type": knownvalue.StringExact("ANYTIME"),
			"day":  knownvalue.Null(),
			"hour": knownvalue.Null(),
		})),
		statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("security_group_ids"), knownvalue.SetSizeExact(0)),
	}

	firstBasicApiChecks := []resource.TestCheckFunc{
		testAccCheckClusterLabelsExact(&cluster, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}),
		testAccCheckClusterHasResources(&cluster, "s2.micro", "network-hdd", 10*1024*1024*1024),
		testAccCheckClusterDeletionProtectionExact(&cluster, false),
		testAccCheckClusterAccessExact(&cluster, &mysql.Access{
			DataLens:     false,
			DataTransfer: false,
			WebSql:       false,
		}),
		testAccCheckClusterPerformanceDiagnosticsExact(&cluster, &mysql.PerformanceDiagnostics{
			Enabled:                    false,
			SessionsSamplingInterval:   60,
			StatementsSamplingInterval: 600,
		}),
		testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
		testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
			Hours:   0,
			Minutes: 0,
		}),
		testAccCheckClusterMaintenanceWindow(&cluster, &mysql.MaintenanceWindow{
			Policy: &mysql.MaintenanceWindow_Anytime{
				Anytime: &mysql.AnytimeMaintenanceWindow{},
			},
		}),
		testAccCheckClusterSecurityGroupIdsExact(&cluster, nil),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLClusterDestroy,
		Steps: []resource.TestStep{
			// Create MySQL Cluster
			{
				Config: testAccMDBMySQLClusterBasic(resourceId, clusterName, description, "PRODUCTION", labels, version, resources),
				ConfigStateChecks: append([]statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact("PRODUCTION")),
				}, firstBasicStateChecks...),
				Check: resource.ComposeAggregateTestCheckFunc(
					append(
						[]resource.TestCheckFunc{
							testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 1),
							testAccCheckClusterStringExact(&cluster, mysql.Cluster_PRODUCTION),
						},
						firstBasicApiChecks...,
					)...,
				),
			},
			mdbMySQLClusterImportStep(clusterResource),
			// Update MySQL Cluster Environment
			{
				Config: testAccMDBMySQLClusterBasic(resourceId, clusterName, description, "PRESTABLE", labels, version, resources),
				ConfigStateChecks: append([]statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact("PRESTABLE")),
				}, firstBasicStateChecks...),
				Check: resource.ComposeAggregateTestCheckFunc(
					append(
						[]resource.TestCheckFunc{
							testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 1),
							testAccCheckClusterStringExact(&cluster, mysql.Cluster_PRESTABLE),
						},
						firstBasicApiChecks...,
					)...,
				),
			},
			// Update MySQL Cluster Environment
			mdbMySQLClusterImportStep(clusterResource),
			{
				Config: testAccMDBMySQLClusterBasic(resourceId, clusterName, descriptionUpdated, "PRESTABLE", labelsUpdated, versionUpdate, resourcesUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionUpdated)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact("PRESTABLE")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("version"), knownvalue.StringExact(versionUpdate)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(false),
							"data_transfer": knownvalue.Bool(false),
							"web_sql":       knownvalue.Bool(false),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("performance_diagnostics"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"enabled":                      knownvalue.Bool(false),
						"sessions_sampling_interval":   knownvalue.Int64Exact(60),
						"statements_sampling_interval": knownvalue.Int64Exact(600),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_window_start"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"hours":   knownvalue.Int64Exact(0),
						"minutes": knownvalue.Int64Exact(0),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_retain_period_days"), knownvalue.Int64Exact(7)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("maintenance_window"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"type": knownvalue.StringExact("ANYTIME"),
						"day":  knownvalue.Null(),
						"hour": knownvalue.Null(),
					})),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("security_group_ids"), knownvalue.SetSizeExact(0)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key4": "value4"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-hdd", 12*1024*1024*1024),
					testAccCheckClusterDeletionProtectionExact(&cluster, false),
					testAccCheckClusterAccessExact(&cluster, &mysql.Access{
						DataLens:     false,
						DataTransfer: false,
						WebSql:       false,
					}),
					testAccCheckClusterPerformanceDiagnosticsExact(&cluster, &mysql.PerformanceDiagnostics{
						Enabled:                    false,
						SessionsSamplingInterval:   60,
						StatementsSamplingInterval: 600,
					}),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   0,
						Minutes: 0,
					}),
					testAccCheckClusterMaintenanceWindow(&cluster, &mysql.MaintenanceWindow{
						Policy: &mysql.MaintenanceWindow_Anytime{
							Anytime: &mysql.AnytimeMaintenanceWindow{},
						},
					}),
					testAccCheckClusterSecurityGroupIdsExact(&cluster, nil),
				),
			},
			mdbMySQLClusterImportStep(clusterResource),
		},
	})
}

// Test that a MySQL Cluster can be created, updated and destroyed
func TestAccMDBMySQLCluster_full(t *testing.T) {
	t.Parallel()

	version := msVersions[0]
	versionUpdate := msVersions[1]

	resources := `
	  resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
	`

	log.Printf("TestAccMDBMySQLCluster_full: version %s", version)
	var cluster mysql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-mysql-cluster-full")

	resourceId := "cluster_full_test"
	clusterResource := "yandex_mdb_mysql_cluster_beta." + resourceId

	description := "MySQL Cluster Terraform Test Full"
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
		data_lens = false
	`

	accessUpdated := `
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
		CheckDestroy:             testAccCheckMDBMySQLClusterDestroy,
		Steps: []resource.TestStep{
			// Create MySQL Cluster
			{
				Config: testAccMDBMySQLClusterFull(
					resourceId, clusterName, description,
					environment, labels, version,
					resources, access,
					performanceDiagnostics,
					backupWindowStart,
					maintenanceWindow,
					backupRetainPeriodDays, true,
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
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("version"), knownvalue.StringExact(version)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(false),
							"data_transfer": knownvalue.Bool(true),
							"web_sql":       knownvalue.Bool(true),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("performance_diagnostics"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"enabled":                      knownvalue.Bool(true),
							"sessions_sampling_interval":   knownvalue.Int64Exact(60),
							"statements_sampling_interval": knownvalue.Int64Exact(600),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_retain_period_days"), knownvalue.Int64Exact(7)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_window_start"), knownvalue.ObjectExact(
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
					testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckClusterDeletionProtectionExact(&cluster, true),
					testAccCheckClusterAccessExact(&cluster, &mysql.Access{
						DataLens:     false,
						DataTransfer: true,
						WebSql:       true,
					}),
					testAccCheckClusterPerformanceDiagnosticsExact(
						&cluster,
						&mysql.PerformanceDiagnostics{
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
					testAccCheckClusterMaintenanceWindow(&cluster, &mysql.MaintenanceWindow{
						Policy: &mysql.MaintenanceWindow_Anytime{
							Anytime: &mysql.AnytimeMaintenanceWindow{},
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
			mdbMySQLClusterImportStep(clusterResource),
			{
				Config: testAccMDBMySQLClusterFull(
					resourceId, clusterName, descriptionUpdated,
					environment, labelsUpdated, versionUpdate, resources, accessUpdated,
					performanceDiagnosticsUpdated,
					backupWindowStartUpdated,
					maintenanceWindowUpdated,
					backupRetainPeriodDaysUpdated, false,
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
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("version"), knownvalue.StringExact(versionUpdate)),

					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("access"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"data_lens":     knownvalue.Bool(true),
							"data_transfer": knownvalue.Bool(false),
							"web_sql":       knownvalue.Bool(false),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("performance_diagnostics"), knownvalue.ObjectExact(
						map[string]knownvalue.Check{
							"enabled":                      knownvalue.Bool(false),
							"sessions_sampling_interval":   knownvalue.Int64Exact(500),
							"statements_sampling_interval": knownvalue.Int64Exact(1000),
						},
					)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_retain_period_days"), knownvalue.Int64Exact(14)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_window_start"), knownvalue.ObjectExact(
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
					testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key4": "value4"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckClusterDeletionProtectionExact(&cluster, false),
					testAccCheckClusterAccessExact(&cluster, &mysql.Access{
						DataLens:     true,
						DataTransfer: false,
						WebSql:       false,
					}),
					testAccCheckClusterPerformanceDiagnosticsExact(
						&cluster,
						&mysql.PerformanceDiagnostics{
							SessionsSamplingInterval:   500,
							StatementsSamplingInterval: 1000,
						},
					),
					testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(14)),
					testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
						Hours:   10,
						Minutes: 3,
					}),
					testAccCheckClusterMaintenanceWindow(&cluster, &mysql.MaintenanceWindow{
						Policy: &mysql.MaintenanceWindow_WeeklyMaintenanceWindow{
							WeeklyMaintenanceWindow: &mysql.WeeklyMaintenanceWindow{
								Day:  mysql.WeeklyMaintenanceWindow_MON,
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
			mdbMySQLClusterImportStep(clusterResource),
		},
	})
}

func TestAccMDBMySQLCluster_mixed(t *testing.T) {
	t.Parallel()

	version := msVersions[1]

	log.Printf("TestAccMDBMySQLCluster_mixed: version %s", version)
	var cluster mysql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-mysql-cluster-mixed")

	resourceId := "cluster_mixed_test"
	clusterResource := "yandex_mdb_mysql_cluster_beta." + resourceId

	folderID := test.GetExampleFolderID()

	descriptionFull := "Cluster test mixed: full"
	descriptionBasic := "Cluster test mixed: basic"

	environment := "PRODUCTION"
	labels := `
		key = "value"
	`

	access := `
	data_lens = false
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

	resources := `
		resource_preset_id = "s2.micro"
		disk_size = 10
		disk_type_id = "network-ssd"
	`

	stepsFullBasic := [2]resource.TestStep{
		{
			Config: testAccMDBMySQLClusterFull(
				resourceId, clusterName, descriptionFull, environment, labels, version, resources, access,
				performanceDiagnostics,
				backupWindowStart,
				maintenanceWindow,
				backupRetainPeriodDays,
				false, []string{},
			),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionFull)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact(environment)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("version"), knownvalue.StringExact(version)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("access"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"data_lens":     knownvalue.Bool(false),
						"data_transfer": knownvalue.Bool(false),
						"web_sql":       knownvalue.Bool(false),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("performance_diagnostics"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"enabled":                      knownvalue.Bool(false),
						"sessions_sampling_interval":   knownvalue.Int64Exact(60),
						"statements_sampling_interval": knownvalue.Int64Exact(600),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_window_start"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"hours":   knownvalue.Int64Exact(0),
						"minutes": knownvalue.Int64Exact(0),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_retain_period_days"), knownvalue.Int64Exact(7)),
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
				testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 1),
				testAccCheckClusterLabelsExact(&cluster, map[string]string{"key": "value"}),
				testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),

				testAccCheckClusterDeletionProtectionExact(&cluster, false),
				testAccCheckClusterAccessExact(&cluster, &mysql.Access{
					DataLens:     false,
					DataTransfer: false,
					WebSql:       false,
				}),
				testAccCheckClusterPerformanceDiagnosticsExact(&cluster, &mysql.PerformanceDiagnostics{
					Enabled:                    false,
					SessionsSamplingInterval:   60,
					StatementsSamplingInterval: 600,
				}),
				testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
					Hours:   0,
					Minutes: 0,
				}),
				testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
				testAccCheckClusterMaintenanceWindow(&cluster, &mysql.MaintenanceWindow{
					Policy: &mysql.MaintenanceWindow_Anytime{
						Anytime: &mysql.AnytimeMaintenanceWindow{},
					},
				}),
				testAccCheckClusterSecurityGroupIdsExact(&cluster, nil),
			),
		},
		{
			Config: testAccMDBMySQLClusterBasic(resourceId, clusterName, descriptionBasic, environment, labels, version, resources),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionBasic)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact(environment)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("version"), knownvalue.StringExact(version)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("deletion_protection"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("access"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"data_lens":     knownvalue.Bool(false),
						"data_transfer": knownvalue.Bool(false),
						"web_sql":       knownvalue.Bool(false),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("performance_diagnostics"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"enabled":                      knownvalue.Bool(false),
						"sessions_sampling_interval":   knownvalue.Int64Exact(60),
						"statements_sampling_interval": knownvalue.Int64Exact(600),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_window_start"), knownvalue.ObjectExact(
					map[string]knownvalue.Check{
						"hours":   knownvalue.Int64Exact(0),
						"minutes": knownvalue.Int64Exact(0),
					},
				)),
				statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("backup_retain_period_days"), knownvalue.Int64Exact(7)),
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
				testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 1),
				testAccCheckClusterLabelsExact(&cluster, map[string]string{"key": "value"}),
				testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
				testAccCheckClusterDeletionProtectionExact(&cluster, false),
				testAccCheckClusterAccessExact(&cluster, &mysql.Access{
					DataLens:     false,
					DataTransfer: false,
					WebSql:       false,
				}),
				testAccCheckClusterPerformanceDiagnosticsExact(&cluster, &mysql.PerformanceDiagnostics{
					Enabled:                    false,
					SessionsSamplingInterval:   60,
					StatementsSamplingInterval: 600,
				}),
				testAccCheckClusterBackupRetainPeriodDaysExact(&cluster, wrapperspb.Int64(7)),
				testAccCheckClusterBackupWindowStartExact(&cluster, &timeofday.TimeOfDay{
					Hours:   0,
					Minutes: 0,
				}),
				testAccCheckClusterMaintenanceWindow(&cluster, &mysql.MaintenanceWindow{
					Policy: &mysql.MaintenanceWindow_Anytime{
						Anytime: &mysql.AnytimeMaintenanceWindow{},
					},
				}),
				testAccCheckClusterSecurityGroupIdsExact(&cluster, nil),
			),
		},
	}

	for i := 0; i < 2; i++ {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProviderFactories,
			CheckDestroy:             testAccCheckMDBMySQLClusterDestroy,
			Steps: []resource.TestStep{
				stepsFullBasic[i],
				stepsFullBasic[i^1],
			},
		})
	}
}

// Test that a MySQL HA Cluster can be created, updated and destroyed
func TestAccMDBMySQLCluster_HostTests(t *testing.T) {
	t.Parallel()

	version := msVersions[1]
	log.Printf("TestAccMDBMySQLCluster_HostTests: version %s", version)
	var cluster mysql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-mysql-cluster-hosts-test")
	clusterResource := "yandex_mdb_mysql_cluster_beta.cluster_host_tests"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccMDBMySQLClusterHostsStep0(clusterName, version, "# no hosts section specified"),
				ExpectError: regexp.MustCompile(`Error: Missing required argument`),
			},
			{
				Config: testAccMDBMySQLClusterHostsStep1(clusterName, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("na").AtMapKey("zone"), knownvalue.StringExact("ru-central1-a")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("zone"), knownvalue.StringExact("ru-central1-b")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nd").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 3),
					resource.TestCheckResourceAttrSet(clusterResource, `hosts.na.fqdn`),
					resource.TestCheckResourceAttrSet(clusterResource, `hosts.nb.fqdn`),
					resource.TestCheckResourceAttrSet(clusterResource, `hosts.nd.fqdn`),
				),
			},
			{
				Config: testAccMDBMySQLClusterHostsStep2(clusterName, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("nb").AtMapKey("zone"), knownvalue.StringExact("ru-central1-b")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 1),
				),
			},
			{
				Config: testAccMDBMySQLClusterHostsStep3(clusterName, version),
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
					testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 3),
				),
			},
			{
				Config: testAccMDBMySQLClusterHostsStep4(clusterName, version),
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
					testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 3),
				),
			},
		},
	})
}

// Test that a MySQL HA Cluster can be created, updated and destroyed
func TestAccMDBMySQLCluster_HostSpecialCaseTests(t *testing.T) {
	t.Parallel()

	version := msVersions[1]
	log.Printf("TestAccMDBMySQLCluster_HostTests: version %s", version)
	var cluster mysql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-mysql-cluster-hosts-special-test")
	clusterResource := "yandex_mdb_mysql_cluster_beta.cluster_hosts_special_case_tests"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLClusterHostsSpecialCaseStep1(clusterName, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("lol").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("kek").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("hosts").AtMapKey("cheburek").AtMapKey("zone"), knownvalue.StringExact("ru-central1-d")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 3),
				),
			},
			{
				Config: testAccMDBMySQLClusterHostsSpecialCaseStep2(clusterName, version),
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
					testAccCheckExistsAndParseMDBMySQLCluster(clusterResource, &cluster, 3),
				),
			},
		},
	})
}

func testAccCheckMDBMySQLClusterDestroy(s *terraform.State) error {
	config := test.AccProvider.(*provider.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_mysql_cluster_beta" {
			continue
		}

		_, err := config.SDK.MDB().MySQL().Cluster().Get(context.Background(), &mysql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("MySQL Cluster still exists")
		}
	}

	return nil
}

func testAccCheckExistsAndParseMDBMySQLCluster(n string, r *mysql.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*provider.Provider).GetConfig()

		found, err := config.SDK.MDB().MySQL().Cluster().Get(context.Background(), &mysql.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("MySQL Cluster not found")
		}

		*r = *found

		resp, err := config.SDK.MDB().MySQL().Cluster().ListHosts(context.Background(), &mysql.ListClusterHostsRequest{
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

func testAccCheckClusterStringExact(r *mysql.Cluster, expected mysql.Cluster_Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if r.Environment == expected {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched labels.\nActual:   %+v\nExpected: %+v", r.Name, r.Environment, expected.String())
	}
}

func testAccCheckClusterLabelsExact(r *mysql.Cluster, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.Labels, expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched labels.\nActual:   %+v\nExpected: %+v", r.Name, r.Labels, expected)
	}
}

func testAccCheckClusterDeletionProtectionExact(r *mysql.Cluster, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if r.GetDeletionProtection() == expected {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config deletion_protection.\nActual:   %+v\nExpected: %+v", r.Name, r.GetDeletionProtection(), expected)
	}
}

func testAccCheckClusterSecurityGroupIdsExact(r *mysql.Cluster, expectedResourceNames []string) resource.TestCheckFunc {
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

func testAccCheckClusterAccessExact(r *mysql.Cluster, expected *mysql.Access) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetAccess(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config access.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetAccess(), expected)
	}
}

func testAccCheckClusterPerformanceDiagnosticsExact(r *mysql.Cluster, expected *mysql.PerformanceDiagnostics) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetPerformanceDiagnostics(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config performance_diagnostics.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetPerformanceDiagnostics(), expected)
	}
}

func testAccCheckClusterBackupRetainPeriodDaysExact(r *mysql.Cluster, expected *wrapperspb.Int64Value) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetBackupRetainPeriodDays(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config backup_retain_period_days.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetBackupWindowStart(), expected)
	}
}

func testAccCheckClusterBackupWindowStartExact(r *mysql.Cluster, expected *timeofday.TimeOfDay) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetConfig().GetBackupWindowStart(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched config backup_window_start.\nActual:   %+v\nExpected: %+v", r.Name, r.GetConfig().GetBackupWindowStart(), expected)
	}
}

func testAccCheckClusterMaintenanceWindow(r *mysql.Cluster, expected *mysql.MaintenanceWindow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if reflect.DeepEqual(r.GetMaintenanceWindow(), expected) {
			return nil
		}
		return fmt.Errorf("Cluster %s has mismatched maintenance_window.\nActual:   %+v\nExpected: %+v", r.Name, r.GetMaintenanceWindow(), expected)
	}
}

func testAccCheckClusterHasResources(r *mysql.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
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

func testAccMDBMySQLClusterBasic(resourceId, name, description, environment, labels, version, resources string) string {
	return fmt.Sprintf(msVPCDependencies+`
resource "yandex_mdb_mysql_cluster_beta" "%s" {
  name        = "%s"
  description = "%s"
  environment = "%s"
  network_id  = yandex_vpc_network.mdb-ms-test-net.id

  labels = {
%s
  }

  hosts = {
    "na" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-a.id
    }
  }

  version = "%s"
  resources {
	%s
  }
}
`, resourceId, name, description, environment, labels, version, resources)
}

func testAccMDBMySQLClusterFull(
	resourceId, clusterName, description, environment, labels,
	version, resources,
	access,
	performanceDiagnostics,
	backupWindowStart,
	maintenanceWindow string, backupRetainPeriodDays int, deletionProtection bool, confSecurityGroupIds []string,
) string {
	return fmt.Sprintf(msVPCDependencies+`
resource "yandex_mdb_mysql_cluster_beta" "%s" {
  name        = "%s"
  description = "%s"
  environment = "%s" 
  network_id  = yandex_vpc_network.mdb-ms-test-net.id

  labels = {
%s
  }

  hosts = {
    "host" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-a.id
    }
  }


  version = "%s"
  resources {
	%s
  }
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
  
  
  maintenance_window = {
	%s
  }

  deletion_protection = %t
  security_group_ids = [%s]

}
`, resourceId, clusterName, description, environment,
		labels, version, resources, access,
		performanceDiagnostics, backupRetainPeriodDays, backupWindowStart,
		maintenanceWindow, deletionProtection, strings.Join(confSecurityGroupIds, ", "),
	)
}

func testAccMDBMySQLClusterHostsStep0(name, version, hosts string) string {
	return fmt.Sprintf(msVPCDependencies+`
resource "yandex_mdb_mysql_cluster_beta" "cluster_host_tests" {
	name        = "%s"
	description = "MySQL Cluster Hosts Terraform Test"
	network_id  = yandex_vpc_network.mdb-ms-test-net.id
	environment = "PRESTABLE"

	version = "%s"
	resources {
		resource_preset_id = "s2.micro"
		disk_size          = 10
		disk_type_id       = "network-ssd"
	}
  
%s
}
`, name, version, hosts)
}

// Init hosts configuration
func testAccMDBMySQLClusterHostsStep1(name, version string) string {
	return testAccMDBMySQLClusterHostsStep0(name, version, `
  hosts = {
    "na" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-a.id
    }
    "nb" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-b.id
    }
    "nd" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-d.id
    }
  }
`)
}

// Drop some hosts
func testAccMDBMySQLClusterHostsStep2(name, version string) string {
	return testAccMDBMySQLClusterHostsStep0(name, version, `
  hosts = {
    "nb" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-b.id
    }
  }
`)
}

// Add some hosts back with all possible options
func testAccMDBMySQLClusterHostsStep3(name, version string) string {
	return testAccMDBMySQLClusterHostsStep0(name, version, `
  hosts = {
    "na" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-a.id
      assign_public_ip = true
    }
    "nb" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-b.id
    }
    "nd" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-d.id
      assign_public_ip = true
    }
  }
`)
}

// Update Hosts
func testAccMDBMySQLClusterHostsStep4(name, version string) string {
	return testAccMDBMySQLClusterHostsStep0(name, version, `
  hosts = {
    "na" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-a.id
      assign_public_ip = false
    }
    "nb" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-b.id
      assign_public_ip = true
    }
    "nd" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-d.id
      assign_public_ip = false
    }
  }
`)
}

func testAccMDBMySQLClusterHostsSpecialCaseStep0(name, version, hosts string) string {
	return fmt.Sprintf(msVPCDependencies+`
resource "yandex_mdb_mysql_cluster_beta" "cluster_hosts_special_case_tests" {
  name        = "%s"
  description = "MySQL Cluster Hosts Terraform Test"
  network_id  = yandex_vpc_network.mdb-ms-test-net.id
  environment = "PRESTABLE"


  version = "%s"
  resources {
	resource_preset_id = "s2.micro"
	disk_size          = 10
	disk_type_id       = "network-ssd"
  }
  
%s
}
`, name, version, hosts)
}

// Init hosts special case configuration
func testAccMDBMySQLClusterHostsSpecialCaseStep1(name, version string) string {
	return testAccMDBMySQLClusterHostsSpecialCaseStep0(name, version, `
  hosts = {
    "lol" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-d.id
    }
    "kek" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-d.id
    }
    "cheburek" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-d.id
    }
  }
`)
}

// Change some options
func testAccMDBMySQLClusterHostsSpecialCaseStep2(name, version string) string {
	return testAccMDBMySQLClusterHostsSpecialCaseStep0(name, version, `
  hosts = {
    "lol" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-d.id
    }
    "kek" = {
      zone      = "ru-central1-d"
	  assign_public_ip = true
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-d.id
    }
    "cheburek" = {
      zone      = "ru-central1-d"
      subnet_id = yandex_vpc_subnet.mdb-ms-test-subnet-d.id
    }
  }
`)
}

// func testAccMDBMySQLClusterConfigHANamedSwitchMaster(name, version string) string
// func testAccMDBMySQLClusterConfigHANamedChangePublicIP(name, version string) string
// func testAccMDBMySQLClusterConfigHANamedWithCascade(name, version string) string
