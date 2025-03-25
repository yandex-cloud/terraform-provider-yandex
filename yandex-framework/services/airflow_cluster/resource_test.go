package airflow_cluster_test

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
	afv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/airflow/v1"
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
	tpl, err := template.New("airflow").Parse(`
resource "yandex_vpc_network" "airflow-net" {}

resource "yandex_vpc_subnet" "airflow-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.airflow-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "airflow-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.airflow-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "airflow-d" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.airflow-net.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_security_group" "airflow-sg1" {
  description = "Test security group 1"
  network_id  = yandex_vpc_network.airflow-net.id
}

resource "yandex_iam_service_account" "airflow-sa-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  name      = "airflow-{{ .RandSuffix }}"
}

resource "yandex_resourcemanager_folder_iam_member" "airflow-sa-bindings-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  role      = "managed-airflow.integrationProvider"
  member    = "serviceAccount:${yandex_iam_service_account.airflow-sa-{{ .RandSuffix }}.id}"
}

resource "yandex_storage_bucket" "airflow-bucket-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  bucket    = "airflow-tf-{{ .RandSuffix }}"
}
`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, p))
	return b.String()
}

type airflowClusterConfigParams struct {
	RandSuffix         string
	FolderID           string
	Webserver          airflowComponentParams
	Scheduler          airflowComponentParams
	Triggerer          *airflowComponentParams
	Worker             airflowWorkerParams
	Labels             map[string]string
	DeletionProtection bool
	AdditionalParams   bool
}

type airflowComponentParams struct {
	Count            int
	ResourcePresetID string
}

type airflowWorkerParams struct {
	MinCount         int
	MaxCount         int
	ResourcePresetID string
}

func airflowClusterConfig(t *testing.T, params airflowClusterConfigParams) string {
	tpl, err := template.New("airflow").Parse(`
resource "yandex_airflow_cluster" "airflow_cluster" {
  name = "airflow-{{ .RandSuffix }}"
  admin_password = "sTr0nGp@sSw0rD"
  code_sync = {
    s3 = {
      bucket = yandex_storage_bucket.airflow-bucket-{{ .RandSuffix }}.bucket
    }
  }
  service_account_id = yandex_iam_service_account.airflow-sa-{{ .RandSuffix }}.id
  subnet_ids = [
    yandex_vpc_subnet.airflow-a.id,
	yandex_vpc_subnet.airflow-b.id,
	yandex_vpc_subnet.airflow-d.id
  ]
  webserver = {
    resource_preset_id = "{{ .Webserver.ResourcePresetID }}"
    count              = {{ .Webserver.Count }}
  }
  scheduler = {
    resource_preset_id = "{{ .Scheduler.ResourcePresetID }}"
    count              = {{ .Scheduler.Count }}
  }
  worker = {
    resource_preset_id = "{{ .Worker.ResourcePresetID }}"
    min_count          = {{ .Worker.MinCount }}
    max_count          = {{ .Worker.MaxCount }}
  }
  deletion_protection = {{ .DeletionProtection }}

  {{ if .Triggerer }}
  triggerer = {
    resource_preset_id = "{{ .Triggerer.ResourcePresetID }}"
    count              = {{ .Triggerer.Count }}
  }
  {{ end }}

  {{ if .Labels }}
  labels = {
    {{ range $key, $val := .Labels}}
	{{ $key }} = "{{ $val }}"
    {{ end }}
  }
  {{ end }}

  {{ if .AdditionalParams }}
  security_group_ids = [yandex_vpc_security_group.airflow-sg1.id]
  pip_packages = ["dbt"]
  deb_packages = ["tree"]
  description = "airflow-cluster"
  airflow_config = {
    "api" = {
      "auth_backends" = "airflow.api.auth.backend.basic_auth,airflow.api.auth.backend.session"
    }
  }
  lockbox_secrets_backend = {
    enabled = true
  }
  logging = {
    enabled   = true
    folder_id = "{{ .FolderID }}"
    min_level = "INFO"
  }
  timeouts {
	create = "50m"
	update = "50m"
	delete = "50m"
  }
  {{ end }}
  depends_on = [
    yandex_resourcemanager_folder_iam_member.airflow-sa-bindings-{{ .RandSuffix }}
  ]
}`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, params))

	return fmt.Sprintf("%s\n%s", infraResources(t, params.RandSuffix), b.String())
}

func testAccCheckAirflowClusterDestroy(s *terraform.State) error {
	sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_airflow_cluster" {
			continue
		}

		_, err := sdk.Airflow().Cluster().Get(context.Background(), &afv1.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Airflow Cluster still exists")
		}
	}

	return nil
}

func airflowClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:            name,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"admin_password"},
	}
}

func testAccCheckAirflowExists(name string, cluster *afv1.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK
		found, err := sdk.Airflow().Cluster().Get(context.Background(), &afv1.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Airflow cluster not found")
		}

		if cluster != nil {
			*cluster = *found
		}

		return nil
	}
}

func TestAccMDBAirflowCluster_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	folderID := os.Getenv("YC_FOLDER_ID")
	var cluster afv1.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckAirflowClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: airflowClusterConfig(t, airflowClusterConfigParams{
					RandSuffix: randSuffix,
					FolderID:   folderID,
					Webserver: airflowComponentParams{
						Count:            1,
						ResourcePresetID: "c1-m4",
					},
					Scheduler: airflowComponentParams{
						Count:            1,
						ResourcePresetID: "c1-m4",
					},
					Worker: airflowWorkerParams{
						MinCount:         1,
						MaxCount:         1,
						ResourcePresetID: "c1-m4",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAirflowExists("yandex_airflow_cluster.airflow_cluster", &cluster),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "admin_password"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "subnet_ids.1"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "subnet_ids.2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "code_sync.s3.bucket", fmt.Sprintf("airflow-tf-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "name", fmt.Sprintf("airflow-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "webserver.count", "1"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "webserver.resource_preset_id", "c1-m4"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "scheduler.count", "1"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "scheduler.resource_preset_id", "c1-m4"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "worker.min_count", "1"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "worker.max_count", "1"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "worker.resource_preset_id", "c1-m4"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "deletion_protection", "false"),
				),
			},
			airflowClusterImportStep("yandex_airflow_cluster.airflow_cluster"),
			{
				Config: airflowClusterConfig(t, airflowClusterConfigParams{
					RandSuffix: randSuffix,
					FolderID:   folderID,
					Webserver: airflowComponentParams{
						Count:            2,
						ResourcePresetID: "c1-m2",
					},
					Scheduler: airflowComponentParams{
						Count:            2,
						ResourcePresetID: "c1-m2",
					},
					Worker: airflowWorkerParams{
						MinCount:         2,
						MaxCount:         2,
						ResourcePresetID: "c1-m2",
					},
					Triggerer: &airflowComponentParams{
						Count:            2,
						ResourcePresetID: "c1-m2",
					},
					Labels: map[string]string{
						"label": "value",
					},
					DeletionProtection: true,
					AdditionalParams:   true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAirflowExists("yandex_airflow_cluster.airflow_cluster", &cluster),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "admin_password"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "subnet_ids.1"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "subnet_ids.2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "code_sync.s3.bucket", fmt.Sprintf("airflow-tf-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "name", fmt.Sprintf("airflow-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "webserver.count", "2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "webserver.resource_preset_id", "c1-m2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "scheduler.count", "2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "scheduler.resource_preset_id", "c1-m2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "worker.min_count", "2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "worker.max_count", "2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "worker.resource_preset_id", "c1-m2"),
					// New specified
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "triggerer.count", "2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "triggerer.resource_preset_id", "c1-m2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "labels.label", "value"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "deletion_protection", "true"),
					// Additional
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "airflow_config.api.auth_backends"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "security_group_ids.0"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "pip_packages.0", "dbt"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "deb_packages.0", "tree"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "description", "airflow-cluster"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "lockbox_secrets_backend.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "logging.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "logging.folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "logging.min_level", "INFO"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "timeouts.create", "50m"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "timeouts.update", "50m"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "timeouts.delete", "50m"),
				),
			},
			airflowClusterImportStep("yandex_airflow_cluster.airflow_cluster"),
			{
				Config: airflowClusterConfig(t, airflowClusterConfigParams{
					RandSuffix: randSuffix,
					FolderID:   folderID,
					Webserver: airflowComponentParams{
						Count:            1,
						ResourcePresetID: "c1-m4",
					},
					Scheduler: airflowComponentParams{
						Count:            1,
						ResourcePresetID: "c1-m4",
					},
					Worker: airflowWorkerParams{
						MinCount:         1,
						MaxCount:         1,
						ResourcePresetID: "c1-m4",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAirflowExists("yandex_airflow_cluster.airflow_cluster", &cluster),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "admin_password"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "subnet_ids.1"),
					resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "subnet_ids.2"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "code_sync.s3.bucket", fmt.Sprintf("airflow-tf-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "name", fmt.Sprintf("airflow-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "webserver.count", "1"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "webserver.resource_preset_id", "c1-m4"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "scheduler.count", "1"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "scheduler.resource_preset_id", "c1-m4"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "worker.min_count", "1"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "worker.max_count", "1"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "worker.resource_preset_id", "c1-m4"),
					resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "deletion_protection", "false"),
				),
			},
			airflowClusterImportStep("yandex_airflow_cluster.airflow_cluster"),
		},
	})
}
