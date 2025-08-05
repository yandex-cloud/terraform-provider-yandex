package spark_cluster_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
	sparkv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

func infraResources(t *testing.T, randSuffix string) string {
	type params struct {
		RandSuffix string
		FolderID   string
	}
	p := params{
		RandSuffix: randSuffix,
		FolderID:   os.Getenv("YC_FOLDER_ID"),
	}
	tpl, err := template.New("spark").Parse(`
resource "yandex_vpc_network" "spark-net" {}

resource "yandex_vpc_subnet" "spark-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.spark-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_security_group" "spark-sg1" {
  description = "Test security group 1"
  network_id  = yandex_vpc_network.spark-net.id
}

resource "yandex_iam_service_account" "spark-sa-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  name      = "spark-{{ .RandSuffix }}"
}

resource "yandex_resourcemanager_folder_iam_member" "spark-sa-bindings-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  role      = "managed-spark.integrationProvider"
  member    = "serviceAccount:${yandex_iam_service_account.spark-sa-{{ .RandSuffix }}.id}"
}
`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, p))
	return b.String()
}

type sparkClusterConfigParams struct {
	RandSuffix                string
	Description               string
	IncludeBlockLabels        bool
	Labels                    map[string]string
	DriverResourcePresetID    string
	DriverSize                int64
	DriverMinSize             int64
	DriverMaxSize             int64
	ExecutorResourcePresetID  string
	ExecutorSize              int64
	ExecutorMinSize           int64
	ExecutorMaxSize           int64
	IncludeBlockDependencies  bool
	PipPackage                string
	DebPackage                string
	IncludeBlockHistoryServer bool
	HistoryServerEnabled      bool
	IncludeBlockMetastore     bool
	MetastoreClusterID        string
	SecurityGroup             bool
	DeletionProtection        bool
	LoggingFolderID           string
	LoggingLogGroupID         string
	IncludeBlockMaintenance   bool
	MaintenanceWindowWeekly   bool
	MaintenanceWindowDay      string
	MaintenanceWindowHour     int64
	IncludeBlockTimeouts      bool
}

func sparkClusterConfig(t *testing.T, params sparkClusterConfigParams) string {
	tpl, err := template.New("spark").Parse(`
resource "yandex_spark_cluster" "spark_cluster" {

  name = "spark-{{ .RandSuffix }}"

  {{ if .Description }}
  description = "{{ .Description }}"
  {{ end }}

  {{ if .IncludeBlockLabels }}
  labels = {
    {{ range $key, $val := .Labels }}
    {{ $key }} = "{{ $val }}"
    {{ end }}
  }
  {{ end }}

  config = {
    resource_pools = {
      driver = {
        resource_preset_id = "{{ .DriverResourcePresetID }}"
        {{ if .DriverSize }}
        size = {{ .DriverSize }}
        {{ else }}
        min_size = {{ .DriverMinSize }}
        max_size = {{ .DriverMaxSize }}
        {{ end }}
      }
      executor = {
        resource_preset_id = "{{ .ExecutorResourcePresetID }}"
        {{ if .ExecutorSize }}
        size = {{ .ExecutorSize }}
        {{ else }}
        min_size = {{ .ExecutorMinSize }}
        max_size = {{ .ExecutorMaxSize }}
        {{ end }}
      }
    }

    {{ if .IncludeBlockDependencies }}
    dependencies = {
      {{ if .PipPackage }}
      pip_packages = ["{{ .PipPackage }}"]
      {{ end }}
      {{ if .DebPackage }}
      deb_packages = ["{{ .DebPackage }}"]
      {{ end }}
    }
    {{ end }}

    {{ if .IncludeBlockHistoryServer }}
    history_server = {
      enabled = {{ .HistoryServerEnabled }}
    }
    {{ end }}

    {{ if .IncludeBlockMetastore }}
    metastore = {
      cluster_id = "{{ .MetastoreClusterID }}"
    }
    {{ end }}
  }

  network = {
    subnet_ids = [yandex_vpc_subnet.spark-a.id]
    {{ if .SecurityGroup }}
    security_group_ids = [yandex_vpc_security_group.spark-sg1.id]
    {{ end }}
  }

  {{ if .DeletionProtection }}
  deletion_protection = {{ .DeletionProtection }}
  {{ end }}

  service_account_id = yandex_iam_service_account.spark-sa-{{ .RandSuffix }}.id

  logging = {
    {{ if .LoggingFolderID }}
    enabled = true
    folder_id = "{{ .LoggingFolderID }}"
    {{ else if .LoggingLogGroupID }}
    enabled = true
    log_group_id = "{{ .LoggingLogGroupID }}"
    {{ else }}
    enabled = false
    {{ end }}
  }

  {{ if .IncludeBlockMaintenance }}
  maintenance_window = {
    {{ if .MaintenanceWindowWeekly }}
    type = "WEEKLY"
    day = "{{ .MaintenanceWindowDay }}"
    hour = {{ .MaintenanceWindowHour }}
    {{ else }}
    type = "ANYTIME"
    {{ end }}
  }
  {{ end }}

  {{ if .IncludeBlockTimeouts }}
  timeouts {
    create = "50m"
    update = "50m"
    delete = "50m"
  }
  {{ end }}

  depends_on = [
    yandex_resourcemanager_folder_iam_member.spark-sa-bindings-{{ .RandSuffix }}
  ]
}`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, params))

	return fmt.Sprintf("%s\n%s", infraResources(t, params.RandSuffix), b.String())
}

func testAccCheckSparkClusterDestroy(s *terraform.State) error {
	sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_spark_cluster" {
			continue
		}

		_, err := sdk.Spark().Cluster().Get(context.Background(), &sparkv1.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Spark Cluster still exists")
		}
	}

	return nil
}

func sparkClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccCheckSparkExists(name string, cluster *sparkv1.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK
		found, err := sdk.Spark().Cluster().Get(context.Background(), &sparkv1.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Spark cluster not found")
		}

		if cluster != nil {
			*cluster = *found
		}

		return nil
	}
}

func TestAccSparkCluster_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	folderID := os.Getenv("YC_FOLDER_ID")
	metastoreClusterID := os.Getenv("YC_METASTORE_ID")
	var cluster sparkv1.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckSparkClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: sparkClusterConfig(t, sparkClusterConfigParams{
					RandSuffix:               randSuffix,
					Description:              "acc-basic-step-01 [created with terraform]",
					DriverResourcePresetID:   "c2-m8",
					DriverSize:               1,
					ExecutorResourcePresetID: "c4-m16",
					ExecutorMinSize:          1,
					ExecutorMaxSize:          2,
					LoggingFolderID:          folderID,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSparkExists("yandex_spark_cluster.spark_cluster", &cluster),
					resource.TestCheckResourceAttrSet("yandex_spark_cluster.spark_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_spark_cluster.spark_cluster", "network.subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "name", fmt.Sprintf("spark-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.driver.resource_preset_id", "c2-m8"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.driver.size", "1"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.resource_preset_id", "c4-m16"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.min_size", "1"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.max_size", "2"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.history_server.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.metastore.cluster_id", ""),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "logging.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "logging.folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "maintenance_window.type", "ANYTIME"),
				),
			},
			sparkClusterImportStep("yandex_spark_cluster.spark_cluster"),
			{
				Config: sparkClusterConfig(t, sparkClusterConfigParams{
					RandSuffix:         randSuffix,
					Description:        "acc-step-02 [created with terraform]",
					IncludeBlockLabels: true,
					Labels: map[string]string{
						"my_label": "my_value",
					},
					DriverResourcePresetID:    "c2-m16",
					DriverSize:                2,
					ExecutorResourcePresetID:  "c8-m32",
					ExecutorMinSize:           2,
					ExecutorMaxSize:           4,
					IncludeBlockDependencies:  true,
					PipPackage:                "numpy==2.2.2",
					IncludeBlockHistoryServer: true,
					HistoryServerEnabled:      true,
					IncludeBlockMetastore:     true,
					MetastoreClusterID:        metastoreClusterID,
					SecurityGroup:             true,
					DeletionProtection:        false,
					LoggingFolderID:           folderID,
					IncludeBlockMaintenance:   true,
					MaintenanceWindowWeekly:   true,
					MaintenanceWindowDay:      "TUE",
					MaintenanceWindowHour:     10,
					IncludeBlockTimeouts:      true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSparkExists("yandex_spark_cluster.spark_cluster", &cluster),
					resource.TestCheckResourceAttrSet("yandex_spark_cluster.spark_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_spark_cluster.spark_cluster", "network.subnet_ids.0"),
					resource.TestCheckResourceAttrSet("yandex_spark_cluster.spark_cluster", "network.security_group_ids.0"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "name", fmt.Sprintf("spark-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "description", "acc-step-02 [created with terraform]"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "labels.my_label", "my_value"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.driver.resource_preset_id", "c2-m16"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.driver.size", "2"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.resource_preset_id", "c8-m32"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.min_size", "2"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.max_size", "4"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.dependencies.pip_packages.0", "numpy==2.2.2"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.history_server.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.metastore.cluster_id", metastoreClusterID),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "logging.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "maintenance_window.day", "TUE"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "maintenance_window.hour", "10"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "timeouts.create", "50m"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "timeouts.update", "50m"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "timeouts.delete", "50m"),
				),
			},
			sparkClusterImportStep("yandex_spark_cluster.spark_cluster"),
			{
				Config: sparkClusterConfig(t, sparkClusterConfigParams{
					RandSuffix:         randSuffix,
					Description:        "acc-basic-step-03 [created with terraform]",
					IncludeBlockLabels: true,
					Labels: map[string]string{
						"my_label_1": "my_value_1",
					},
					DriverResourcePresetID:    "c2-m8",
					DriverSize:                1,
					ExecutorResourcePresetID:  "c4-m16",
					ExecutorMinSize:           1,
					ExecutorMaxSize:           2,
					IncludeBlockDependencies:  false,
					IncludeBlockHistoryServer: true,
					HistoryServerEnabled:      true,
					IncludeBlockMetastore:     false,
					SecurityGroup:             false,
					DeletionProtection:        false,
					LoggingFolderID:           folderID,
					IncludeBlockMaintenance:   true,
					MaintenanceWindowWeekly:   false,
					IncludeBlockTimeouts:      false,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSparkExists("yandex_spark_cluster.spark_cluster", &cluster),
					resource.TestCheckResourceAttrSet("yandex_spark_cluster.spark_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_spark_cluster.spark_cluster", "network.subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "name", fmt.Sprintf("spark-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.driver.resource_preset_id", "c2-m8"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.driver.size", "1"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.resource_preset_id", "c4-m16"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.min_size", "1"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.max_size", "2"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.history_server.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "config.metastore.cluster_id", ""),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "logging.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "logging.folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "maintenance_window.type", "ANYTIME"),
				),
			},
			sparkClusterImportStep("yandex_spark_cluster.spark_cluster"),
		},
	})
}
