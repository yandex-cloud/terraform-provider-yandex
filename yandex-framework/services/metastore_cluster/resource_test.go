package metastore_cluster_test

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
	"github.com/yandex-cloud/go-genproto/yandex/cloud/metastore/v1"
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
	tpl, err := template.New("metastore").Parse(`
resource "yandex_vpc_network" "metastore-net" {}

resource "yandex_vpc_subnet" "metastore-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.metastore-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "metastore-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.metastore-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_security_group" "metastore-sg1" {
  description = "Test security group 1"
  network_id  = yandex_vpc_network.metastore-net.id
}

resource "yandex_iam_service_account" "metastore-sa-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  name      = "metastore-{{ .RandSuffix }}"
}

resource "yandex_resourcemanager_folder_iam_member" "metastore-sa-bindings-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  role      = "managed-metastore.integrationProvider"
  member    = "serviceAccount:${yandex_iam_service_account.metastore-sa-{{ .RandSuffix }}.id}"
}
`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, p))
	return b.String()
}

type metastoreClusterConfigParams struct {
	RandSuffix         string
	FolderID           string
	FolderIDSpecified  bool
	Labels             map[string]string
	MaintenanceWindow  *MaintenanceWindow
	DeletionProtection optional[bool]
	LoggingEnabled     optional[bool]
	Description        optional[string]
	SGIDsSpecified     optional[bool]
	SubnetIDVar        string
	ResourcePreset     string
}

type MaintenanceWindow struct {
	Type string
	Hour int
	Day  string
}

type optional[T any] struct {
	Valid bool
	Value T
}

func newOptional[T any](val T) optional[T] {
	return optional[T]{
		Valid: true,
		Value: val,
	}
}

func metastoreClusterConfig(t *testing.T, params metastoreClusterConfigParams) string {
	tpl, err := template.New("metastore").Parse(`
resource "yandex_metastore_cluster" "metastore_cluster" {
  name = "metastore-{{ .RandSuffix }}"
  service_account_id = yandex_iam_service_account.metastore-sa-{{ .RandSuffix }}.id
  subnet_ids = [{{ .SubnetIDVar }}]

  cluster_config = {
    resource_preset_id = "{{ .ResourcePreset }}"
  }
  
  {{ if .DeletionProtection.Valid }}
  deletion_protection = {{ .DeletionProtection.Value }}
  {{ end }}
 
  {{ if .Labels }}
  labels = {
    {{ range $key, $val := .Labels}}
	{{ $key }} = "{{ $val }}"
    {{ end }}
  }
  {{ end }}

  {{ if .MaintenanceWindow }}
  maintenance_window = {
	type = "{{ .MaintenanceWindow.Type }}"
	{{ if eq .MaintenanceWindow.Type "WEEKLY"}}
	day  = "{{ .MaintenanceWindow.Day }}"
	hour = {{ .MaintenanceWindow.Hour }}
	{{ end }}
  }
  {{ end }}

  {{ if .Description.Valid }}
  description = "{{ .Description.Value }}"
  {{ end }}
  
  {{ if .SGIDsSpecified.Valid }}
  {{ if .SGIDsSpecified.Value }}
  security_group_ids = [yandex_vpc_security_group.metastore-sg1.id]
  {{ else }}
  security_group_ids = []
  {{ end }}
  {{ end }}

  {{ if .FolderIDSpecified }}
  folder_id = "{{ .FolderID }}"
  {{ end }}

  {{ if .LoggingEnabled.Valid }}
  logging = {
    {{ if .LoggingEnabled.Value }}
    enabled   = true
    folder_id = "{{ .FolderID }}"
    min_level = "INFO"
	{{ else }}
	enabled = false
	{{ end }}
  }
  {{ end }}

  timeouts {
	create = "50m"
	update = "50m"
	delete = "50m"
  }
  depends_on = [
    yandex_resourcemanager_folder_iam_member.metastore-sa-bindings-{{ .RandSuffix }}
  ]
}`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, params))

	return fmt.Sprintf("%s\n%s", infraResources(t, params.RandSuffix), b.String())
}

func testAccCheckMetastoreClusterDestroy(s *terraform.State) error {
	sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_metastore_cluster" {
			continue
		}

		_, err := sdk.Metastore().Cluster().Get(context.Background(), &metastore.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Metastore Cluster still exists")
		}
	}

	return nil
}

func metastoreClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:            name,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"health"},
	}
}

func testAccCheckMetastoreExists(name string, cluster *metastore.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK
		found, err := sdk.Metastore().Cluster().Get(context.Background(), &metastore.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Metastore cluster not found")
		}

		if cluster != nil {
			*cluster = *found
		}

		return nil
	}
}

func TestAccMDBMetastoreCluster_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	anotherRandSuffix := fmt.Sprintf("%d", acctest.RandInt())
	folderID := os.Getenv("YC_FOLDER_ID")
	var cluster metastore.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMetastoreClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: metastoreClusterConfig(t, metastoreClusterConfigParams{
					RandSuffix:     randSuffix,
					FolderID:       folderID,
					SubnetIDVar:    "yandex_vpc_subnet.metastore-a.id",
					ResourcePreset: "c2-m4",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetastoreExists("yandex_metastore_cluster.metastore_cluster", &cluster),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "name", fmt.Sprintf("metastore-%s", randSuffix)),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "cluster_config.resource_preset_id", "c2-m4"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "maintenance_window.type", "ANYTIME"),
					// Not set
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "description"),
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "security_group_ids"),
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "logging"),
				),
			},
			metastoreClusterImportStep("yandex_metastore_cluster.metastore_cluster"),
			{
				Config: metastoreClusterConfig(t, metastoreClusterConfigParams{
					RandSuffix:        anotherRandSuffix,
					FolderID:          folderID,
					FolderIDSpecified: true,
					Labels: map[string]string{
						"label": "value",
					},
					MaintenanceWindow: &MaintenanceWindow{
						Type: "WEEKLY",
						Day:  "MON",
						Hour: 2,
					},
					DeletionProtection: newOptional(true),
					LoggingEnabled:     newOptional(true),
					Description:        newOptional("description"),
					SGIDsSpecified:     newOptional(true),
					SubnetIDVar:        "yandex_vpc_subnet.metastore-a.id",
					ResourcePreset:     "c2-m8",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetastoreExists("yandex_metastore_cluster.metastore_cluster", &cluster),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "name", fmt.Sprintf("metastore-%s", anotherRandSuffix)),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "cluster_config.resource_preset_id", "c2-m8"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "deletion_protection", "true"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "maintenance_window.day", "MON"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "maintenance_window.hour", "2"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "description", "description"),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "security_group_ids.0"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "logging.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "logging.min_level", "INFO"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "logging.folder_id", folderID),
				),
			},
			metastoreClusterImportStep("yandex_metastore_cluster.metastore_cluster"),
			{
				Config: metastoreClusterConfig(t, metastoreClusterConfigParams{
					RandSuffix:         randSuffix,
					FolderID:           folderID,
					FolderIDSpecified:  true,
					Labels:             nil,
					MaintenanceWindow:  nil,
					DeletionProtection: newOptional(false),
					LoggingEnabled:     newOptional(false),
					Description:        newOptional(""),
					SGIDsSpecified:     newOptional(false),
					SubnetIDVar:        "yandex_vpc_subnet.metastore-a.id",
					ResourcePreset:     "c2-m4",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetastoreExists("yandex_metastore_cluster.metastore_cluster", &cluster),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "name", fmt.Sprintf("metastore-%s", randSuffix)),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "cluster_config.resource_preset_id", "c2-m4"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "maintenance_window.type", "ANYTIME"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "description", ""),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "logging.enabled", "false"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "logging.folder_id", folderID), // is returned by metastore API
					// Not set
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "security_group_ids.0"),
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "maintenance_window.day"),
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "maintenance_window.hour"),
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "logging.min_level"),
				),
			},
		},
	})
}

func TestAccMDBMetastoreCluster_recreate(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	folderID := os.Getenv("YC_FOLDER_ID")
	var cluster metastore.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMetastoreClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: metastoreClusterConfig(t, metastoreClusterConfigParams{
					RandSuffix:     randSuffix,
					FolderID:       folderID,
					ResourcePreset: "c2-m4",
					SubnetIDVar:    "yandex_vpc_subnet.metastore-a.id",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetastoreExists("yandex_metastore_cluster.metastore_cluster", &cluster),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "name", fmt.Sprintf("metastore-%s", randSuffix)),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "cluster_config.resource_preset_id", "c2-m4"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "maintenance_window.type", "ANYTIME"),
					// Not set
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "description"),
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "security_group_ids"),
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "logging"),
				),
			},
			metastoreClusterImportStep("yandex_metastore_cluster.metastore_cluster"),
			{
				Config: metastoreClusterConfig(t, metastoreClusterConfigParams{
					RandSuffix:     randSuffix,
					FolderID:       folderID,
					SubnetIDVar:    "yandex_vpc_subnet.metastore-b.id",
					ResourcePreset: "c2-m4",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetastoreExists("yandex_metastore_cluster.metastore_cluster", &cluster),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "name", fmt.Sprintf("metastore-%s", randSuffix)),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_metastore_cluster.metastore_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "cluster_config.resource_preset_id", "c2-m4"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_metastore_cluster.metastore_cluster", "maintenance_window.type", "ANYTIME"),
					// Not set
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "description"),
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "security_group_ids"),
					resource.TestCheckNoResourceAttr("yandex_metastore_cluster.metastore_cluster", "logging"),
				),
			},
		},
	})
}
