package postgresql

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"regexp"

	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"

	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test"
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

`

var postgresql_versions = [...]string{"12", "12-1c", "13", "13-1c", "14", "14-1c", "15", "15-1c", "16", "17"}

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
		},
	}
}

// Test that a PostgreSQL Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_basic(t *testing.T) {
	t.Parallel()

	version := postgresql_versions[rand.Intn(len(postgresql_versions))]
	log.Printf("TestAccMDBPostgreSQLCluster_basic: version %s", version)
	var cluster postgresql.Cluster
	clusterName := acctest.RandomWithPrefix("tf-postgresql-cluster-basic")
	clusterResource := "yandex_mdb_postgresql_cluster_beta.cluster_basic_test"
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
				Config: testAccMDBPGClusterBasic(clusterName, description, "PRESTABLE", labels, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact("PRESTABLE")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("version"), knownvalue.StringExact(version)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPostgreSQLClusterExists(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
				),
			},
			mdbPGClusterImportStep(clusterResource),
			// Update PostgreSQL Cluster
			{
				Config: testAccMDBPGClusterBasic(clusterName, descriptionUpdated, "PRESTABLE", labelsUpdated, version),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("name"), knownvalue.StringExact(clusterName)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("description"), knownvalue.StringExact(descriptionUpdated)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("environment"), knownvalue.StringExact("PRESTABLE")),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("network_id"), knownvalue.NotNull()), // TODO write check that network_id is not empty
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("folder_id"), knownvalue.StringExact(folderID)),
					statecheck.ExpectKnownValue(clusterResource, tfjsonpath.New("config").AtMapKey("version"), knownvalue.StringExact(version)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPostgreSQLClusterExists(clusterResource, &cluster, 1),
					testAccCheckClusterLabelsExact(&cluster, map[string]string{"key4": "value4"}),
					testAccCheckClusterHasResources(&cluster, "s2.micro", "network-ssd", 10*1024*1024*1024),
				),
			},
			mdbPGClusterImportStep(clusterResource),
		},
	})
}

// Test that a PostgreSQL HA Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_HostTests(t *testing.T) {
	t.Parallel()

	version := postgresql_versions[rand.Intn(len(postgresql_versions))]
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPostgreSQLClusterExists(clusterResource, &cluster, 3),
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPostgreSQLClusterExists(clusterResource, &cluster, 1),
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPostgreSQLClusterExists(clusterResource, &cluster, 3),
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPostgreSQLClusterExists(clusterResource, &cluster, 3),
				),
			},
		},
	})
}

// Test that a PostgreSQL HA Cluster can be created, updated and destroyed
func TestAccMDBPostgreSQLCluster_HostSpecialCaseTests(t *testing.T) {
	t.Parallel()

	version := postgresql_versions[rand.Intn(len(postgresql_versions))]
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPostgreSQLClusterExists(clusterResource, &cluster, 3),
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPostgreSQLClusterExists(clusterResource, &cluster, 3),
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

func testAccCheckMDBPostgreSQLClusterExists(n string, r *postgresql.Cluster, hosts int) resource.TestCheckFunc {
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

func testAccMDBPGClusterBasic(name, description, environment, labels, version string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster_beta" "cluster_basic_test" {
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
      resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
    }
  }
}
`, name, description, environment, labels, version)
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
