package trino_cluster_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
	trinov1 "github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/trino_cluster"
)

var (
	//go:embed test-data/CA1.pem
	caCert1 string
	//go:embed test-data/CA2.pem
	caCert2 string
	//go:embed test-data/resource-groups-1.json
	resourceGroups1 string
	//go:embed test-data/resource-groups-2.json
	resourceGroups2 string

	//go:embed test-data/expected-resource-groups-1.json
	expectedResourceGroups1 string
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
	tpl, err := template.New("trino").Parse(`
resource "yandex_vpc_network" "trino-net" {}

resource "yandex_vpc_subnet" "trino-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.trino-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_security_group" "trino-sg1" {
  description = "Test security group 1"
  network_id  = yandex_vpc_network.trino-net.id
}

resource "yandex_iam_service_account" "trino-sa-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  name      = "trino-{{ .RandSuffix }}"
}

resource "yandex_resourcemanager_folder_iam_member" "trino-sa-bindings-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  role      = "managed-trino.integrationProvider"
  member    = "serviceAccount:${yandex_iam_service_account.trino-sa-{{ .RandSuffix }}.id}"
}
`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, p))
	return b.String()
}

type trinoClusterConfigParams struct {
	RandSuffix         string
	FolderID           string
	Coordinator        trinoComponentParams
	Worker             trinoWorkerParams
	Labels             map[string]string
	MaintenanceWindow  *MaintenanceWindow
	DeletionProtection bool
	AdditionalParams   bool
	RetryPolicy        *RetryPolicyParams
	Version            string
	TrustedCerts       []string
	ResourceGroups     string
	QueryProperties    map[string]string
}

type MaintenanceWindow struct {
	Type string
	Hour int
	Day  string
}

type trinoComponentParams struct {
	ResourcePresetID string
}

type trinoWorkerParams struct {
	ResourcePresetID string
	FixedScale       *FixedScaleParams
	AutoScale        *AutoScaleParams
}

type FixedScaleParams struct {
	Count int
}

type AutoScaleParams struct {
	MinCount int
	MaxCount int
}

type RetryPolicyParams struct {
	Policy               string
	AdditionalProperties map[string]string
	ExchangeManager      ExchangeManagerParams
}

type ExchangeManagerParams struct {
	AdditionalProperties map[string]string
}

func trinoClusterConfig(t *testing.T, params trinoClusterConfigParams) string {
	tpl, err := template.New("trino").Parse(`
resource "yandex_trino_cluster" "trino_cluster" {
  name               = "trino-{{ .RandSuffix }}"
  service_account_id = yandex_iam_service_account.trino-sa-{{ .RandSuffix }}.id
  subnet_ids = [yandex_vpc_subnet.trino-a.id]
  coordinator = {
    resource_preset_id = "{{ .Coordinator.ResourcePresetID }}"
  }
  worker = {
    resource_preset_id = "{{ .Worker.ResourcePresetID }}"
    {{ if .Worker.FixedScale }}
    fixed_scale = {
      count = {{ .Worker.FixedScale.Count }}
    }
    {{ end }}
    {{ if .Worker.AutoScale }}
    auto_scale = {
      min_count = {{ .Worker.AutoScale.MinCount }}
      max_count = {{ .Worker.AutoScale.MaxCount }}
    }
    {{ end }}
  }
  deletion_protection = {{ .DeletionProtection }}

  {{ if .Version }}
  version = "{{ .Version }}"
  {{ end }}

  {{ if .TrustedCerts }}
  tls = {
    trusted_certificates = [
	{{ range .TrustedCerts}}
<<EOT
{{ . -}}
EOT
,
	{{ end }}
	]
  }
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

  {{ if .RetryPolicy }}
  retry_policy = {
    policy = "{{ .RetryPolicy.Policy }}"
    {{ if .RetryPolicy.AdditionalProperties }}
    additional_properties = {
      {{ range $key, $val := .RetryPolicy.AdditionalProperties}}
      {{ $key }} = "{{ $val }}"
      {{ end }}
    }
    {{ end }}
    exchange_manager = {
      service_s3 = {}
      {{ if .RetryPolicy.ExchangeManager.AdditionalProperties }}
      additional_properties = {
        {{ range $key, $val := .RetryPolicy.ExchangeManager.AdditionalProperties}}
        {{ $key }} = "{{ $val }}"
        {{ end }}
      }
      {{ end }}
    }
  }
  {{ end }}

  {{ if .AdditionalParams }}
  security_group_ids = [yandex_vpc_security_group.trino-sg1.id]
  description = "trino-cluster"
  logging = {
    enabled   = true
    folder_id = "{{ .FolderID }}"
    min_level = "INFO"
  }


  {{ end }}

  {{ if .ResourceGroups }}
  resource_groups_json = <<-EOT
	{{ .ResourceGroups }}
  EOT
  {{ end }}

  {{ if .QueryProperties }}
  query_properties = {
    {{ range $key, $val := .QueryProperties}}
    "{{ $key }}" = "{{ $val }}"
    {{ end }}
  }
  {{ end }}

  timeouts {
		create = "50m"
		update = "50m"
		delete = "50m"
  }

  depends_on = [
    yandex_resourcemanager_folder_iam_member.trino-sa-bindings-{{ .RandSuffix }}
  ]
}`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, params))

	return fmt.Sprintf("%s\n%s", infraResources(t, params.RandSuffix), b.String())
}

func testAccCheckTrinoClusterDestroy(s *terraform.State) error {
	sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_trino_cluster" {
			continue
		}

		_, err := sdk.Trino().Cluster().Get(context.Background(), &trinov1.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Trino Cluster still exists")
		}
	}

	return nil
}

func trinoClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"health",               // volatile value
			"resource_groups_json", // json value is can cause problems with import
		},
	}
}

func testAccCheckTrinoExists(name string, cluster *trinov1.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK
		found, err := sdk.Trino().Cluster().Get(context.Background(), &trinov1.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Trino cluster not found")
		}

		if cluster != nil {
			*cluster = *found
		}

		return nil
	}
}

func TestAccMDBTrinoCluster_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	folderID := os.Getenv("YC_FOLDER_ID")
	var cluster trinov1.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckTrinoClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: trinoClusterConfig(t, trinoClusterConfigParams{
					RandSuffix: randSuffix,
					FolderID:   folderID,
					Coordinator: trinoComponentParams{
						ResourcePresetID: "c4-m16",
					},
					Worker: trinoWorkerParams{
						ResourcePresetID: "c4-m16",
						FixedScale: &FixedScaleParams{
							Count: 1,
						},
					},
					Version:        "468",
					TrustedCerts:   []string{caCert1},
					ResourceGroups: resourceGroups1,
					QueryProperties: map[string]string{
						"query.max-memory-per-node": "7GB",
						"query.max-cpu-time":        "11h",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoExists("yandex_trino_cluster.trino_cluster", &cluster),
					resource.TestCheckResourceAttrSet("yandex_trino_cluster.trino_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_trino_cluster.trino_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "name", fmt.Sprintf("trino-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "coordinator.resource_preset_id", "c4-m16"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "worker.resource_preset_id", "c4-m16"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "worker.fixed_scale.count", "1"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "maintenance_window.type", "ANYTIME"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "version", "468"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "tls.trusted_certificates.0", caCert1),
					testCheckResourceGroupsEqual("yandex_trino_cluster.trino_cluster", "resource_groups_json", expectedResourceGroups1),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "query_properties.query.max-memory-per-node", "7GB"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "query_properties.query.max-cpu-time", "11h"),
				),
			},
			trinoClusterImportStep("yandex_trino_cluster.trino_cluster"),
			{
				Config: trinoClusterConfig(t, trinoClusterConfigParams{
					RandSuffix: randSuffix,
					FolderID:   folderID,
					Labels: map[string]string{
						"label": "value",
					},
					Coordinator: trinoComponentParams{
						ResourcePresetID: "c8-m32",
					},
					Worker: trinoWorkerParams{
						ResourcePresetID: "c8-m32",
						FixedScale: &FixedScaleParams{
							Count: 2,
						},
					},
					MaintenanceWindow: &MaintenanceWindow{
						Type: "WEEKLY",
						Day:  "MON",
						Hour: 2,
					},
					AdditionalParams: true,
					RetryPolicy: &RetryPolicyParams{
						Policy: "TASK",
						AdditionalProperties: map[string]string{
							"fault-tolerant-execution-max-task-split-count": "1024",
						},
						ExchangeManager: ExchangeManagerParams{
							AdditionalProperties: map[string]string{},
						},
					},
					Version:        "476",
					TrustedCerts:   []string{caCert2},
					ResourceGroups: resourceGroups2,
					QueryProperties: map[string]string{
						"query.max-memory-per-node": "7MB",
						"query.max-run-time":        "23h",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoExists("yandex_trino_cluster.trino_cluster", &cluster),
					resource.TestCheckResourceAttrSet("yandex_trino_cluster.trino_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_trino_cluster.trino_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "name", fmt.Sprintf("trino-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "folder_id", folderID),
					// New specified
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "labels.label", "value"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "maintenance_window.day", "MON"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "maintenance_window.hour", "2"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "coordinator.resource_preset_id", "c8-m32"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "worker.resource_preset_id", "c8-m32"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "worker.fixed_scale.count", "2"),
					// Additional
					resource.TestCheckResourceAttrSet("yandex_trino_cluster.trino_cluster", "security_group_ids.0"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "description", "trino-cluster"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "logging.enabled", "true"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "logging.folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "logging.min_level", "INFO"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "retry_policy.policy", "TASK"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "retry_policy.additional_properties.fault-tolerant-execution-max-task-split-count", "1024"),

					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "timeouts.create", "50m"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "timeouts.update", "50m"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "timeouts.delete", "50m"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "version", "476"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "tls.trusted_certificates.0", caCert2),
					testCheckResourceGroupsEqual("yandex_trino_cluster.trino_cluster", "resource_groups_json", resourceGroups2),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "query_properties.query.max-memory-per-node", "7MB"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "query_properties.query.max-run-time", "23h"),
				),
			},
			trinoClusterImportStep("yandex_trino_cluster.trino_cluster"),
			{
				Config: trinoClusterConfig(t, trinoClusterConfigParams{
					RandSuffix: randSuffix,
					FolderID:   folderID,
					Coordinator: trinoComponentParams{
						ResourcePresetID: "c4-m16",
					},
					Worker: trinoWorkerParams{
						ResourcePresetID: "c4-m16",
						FixedScale: &FixedScaleParams{
							Count: 1,
						},
					},
					MaintenanceWindow: &MaintenanceWindow{
						Type: "ANYTIME",
					},
					Version: "468",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrinoExists("yandex_trino_cluster.trino_cluster", &cluster),
					resource.TestCheckResourceAttrSet("yandex_trino_cluster.trino_cluster", "service_account_id"),
					resource.TestCheckResourceAttrSet("yandex_trino_cluster.trino_cluster", "subnet_ids.0"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "name", fmt.Sprintf("trino-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "coordinator.resource_preset_id", "c4-m16"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "worker.resource_preset_id", "c4-m16"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "worker.fixed_scale.count", "1"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "maintenance_window.type", "ANYTIME"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "version", "468"),
					resource.TestCheckResourceAttr("yandex_trino_cluster.trino_cluster", "tls.trusted_certificates.#", "0"),
					resource.TestCheckNoResourceAttr("yandex_trino_cluster.trino_cluster", "resource_groups_json"),
					resource.TestCheckNoResourceAttr("yandex_trino_cluster.trino_cluster", "query_properties"),
				),
			},
			trinoClusterImportStep("yandex_trino_cluster.trino_cluster"),
		},
	})
}

func testCheckResourceGroupsEqual(resourceName, attrName, expectedJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		actualJSON := rs.Primary.Attributes[attrName]
		if actualJSON == "" {
			return fmt.Errorf("attribute %s not found in resource %s", attrName, resourceName)
		}

		expectedConfig := &trino_cluster.ResourceGroups{}
		if err := json.Unmarshal([]byte(expectedJSON), expectedConfig); err != nil {
			return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
		}

		actualConfig := &trino_cluster.ResourceGroups{}
		if err := json.Unmarshal([]byte(actualJSON), actualConfig); err != nil {
			return fmt.Errorf("failed to unmarshal actual JSON: %w", err)
		}

		if !expectedConfig.Equal(actualConfig) {
			return fmt.Errorf("resource_groups_json mismatch:\nexpected: %+v\nactual: %+v", expectedConfig, actualConfig)
		}

		return nil
	}
}
