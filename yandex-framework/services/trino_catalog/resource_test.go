package trino_catalog_test

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
	trinov1 "github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

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

resource "yandex_iam_service_account" "trino-sa-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  name      = "trino-{{ .RandSuffix }}"
}

resource "yandex_resourcemanager_folder_iam_member" "trino-sa-bindings-{{ .RandSuffix }}" {
  folder_id = "{{ .FolderID }}"
  role      = "managed-trino.integrationProvider"
  member    = "serviceAccount:${yandex_iam_service_account.trino-sa-{{ .RandSuffix }}.id}"
}

resource "yandex_trino_cluster" "trino_cluster" {
  name               = "trino-{{ .RandSuffix }}"
  service_account_id = yandex_iam_service_account.trino-sa-{{ .RandSuffix }}.id
  subnet_ids = [
    yandex_vpc_subnet.trino-a.id
  ]
  coordinator = {
    resource_preset_id = "c4-m16"
  }
  worker = {
    resource_preset_id = "c4-m16"
    fixed_scale = {
      count = 1
    }
  }
  depends_on = [
    yandex_resourcemanager_folder_iam_member.trino-sa-bindings-{{ .RandSuffix }}
  ]
}
`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, p))
	return b.String()
}

type trinoCatalogConfigParams struct {
	RandSuffix    string
	FolderID      string
	CatalogName   string
	Description   string
	Labels        map[string]string
	ConnectorType string
	Postgresql    *PostgresqlConnectorConfig
	Hive          *HiveConnectorConfig
}

type PostgresqlConnectorConfig struct {
	OnPremise            *OnPremise
	ConnectionManager    *ConnectionManager
	AdditionalProperties map[string]string
}

type OnPremise struct {
	ConnectionURL string
	UserName      string
	Password      string
}

type ConnectionManager struct {
	ConectionID          string
	Database             string
	ConnectionProperties map[string]string
}

type HiveConnectorConfig struct {
	MetaStoreURI         string
	AdditionalProperties map[string]string
}

func trinoCatalogConfig(t *testing.T, params trinoCatalogConfigParams) string {
	tpl, err := template.New("trino_catalog").Parse(`
resource "yandex_trino_catalog" "trino_catalog" {
  name       = "{{ .CatalogName }}"
  cluster_id = yandex_trino_cluster.trino_cluster.id
  {{ if .Description }}
  description = "{{ .Description }}"
  {{ end }}

  {{ if .Labels }}
  labels = {
    {{ range $key, $val := .Labels}}
    {{ $key }} = "{{ $val }}"
    {{ end }}
  }
  {{ end }}

  {{ if eq .ConnectorType "postgresql" }}
  postgresql = {
		{{ if .Postgresql.OnPremise }}
    on_premise = {
      connection_url = "{{ .Postgresql.OnPremise.ConnectionURL }}"
      user_name      = "{{ .Postgresql.OnPremise.UserName }}"
      password       = "{{ .Postgresql.OnPremise.Password }}"
    }
		{{ end }}
		
		{{ if .Postgresql.ConnectionManager }}
    connection_manager = {
      connection_id = "{{ .Postgresql.ConnectionManager.ConnectionURL }}"
      database      = "{{ .Postgresql.ConnectionManager.UserName }}"
				connection_properties
    	
			{{ if .Postgresql.ConnectionManager.ConnectionProperties }}
			additional_properties = {
				{{ range $key, $val := .Postgresql.ConnectionManager.ConnectionProperties }}
				"{{ $key }}" = "{{ $val }}"
		    {{ end }}
			}
			{{ end }}
    }
    {{ end }}

		{{ if .Postgresql.AdditionalProperties }}
		additional_properties = {
			{{ range $key, $val := .Postgresql.AdditionalProperties }}
			"{{ $key }}" = "{{ $val }}"
			{{ end }}
		}
		{{ end }}
  }
  {{ end }}


	{{ if eq .ConnectorType "hive" }}
  hive = {
		metastore = {
			uri = "{{ .Hive.MetaStoreURI }}"
		}
		file_system = {
			s3 = {}	
		}

		{{ if .Hive.AdditionalProperties }}
		additional_properties = {
			{{ range $key, $val := .Hive.AdditionalProperties }}
			"{{ $key }}" = "{{ $val }}"
			{{ end }}
		{{ end }}
  }
  {{ end }}


  {{ if eq .ConnectorType "tpch" }}
  tpch = {}
  {{ end }}


  timeouts {
    create = "50m"
    update = "50m"
    delete = "50m"
  }
}`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, params))

	return fmt.Sprintf("%s\n%s", infraResources(t, params.RandSuffix), b.String())
}

func testAccCheckTrinoCatalogDestroy(s *terraform.State) error {
	sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_trino_catalog" {
			continue
		}

		clusterID := rs.Primary.Attributes["cluster_id"]
		_, err := sdk.Trino().Catalog().Get(context.Background(), &trinov1.GetCatalogRequest{
			ClusterId: clusterID,
			CatalogId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Trino Catalog still exists")
		}
	}

	return nil
}

func trinoCatalogImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:            name,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"postgresql.on_premise.password"},
		ImportStateIdFunc: func(s *terraform.State) (string, error) {
			rs, ok := s.RootModule().Resources[name]
			if !ok {
				return "", fmt.Errorf("resource not found: %s", name)
			}
			clusterID := rs.Primary.Attributes["cluster_id"]
			catalogID := rs.Primary.ID
			return resourceid.Construct(clusterID, catalogID), nil
		},
	}
}

func testAccCheckTrinoCatalogExists(name string, catalog *trinov1.Catalog) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK
		clusterID := rs.Primary.Attributes["cluster_id"]
		found, err := sdk.Trino().Catalog().Get(context.Background(), &trinov1.GetCatalogRequest{
			ClusterId: clusterID,
			CatalogId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Trino catalog not found")
		}

		if catalog != nil {
			*catalog = *found
		}

		return nil
	}
}

func TestAccMDBTrinoCatalog_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	folderID := os.Getenv("YC_FOLDER_ID")
	var catalog trinov1.Catalog

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckTrinoCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("catalog-%s", randSuffix),
					ConnectorType: "tpch",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("catalog-%s", randSuffix)),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "cluster_id"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "id"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "tpch.%"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "timeouts.create", "50m"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "timeouts.update", "50m"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "timeouts.delete", "50m"),
				),
			},
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("pg-catalog-%s", randSuffix),
					Description:   "PostgreSQL catalog",
					ConnectorType: "postgresql",
					Postgresql: &PostgresqlConnectorConfig{
						OnPremise: &OnPremise{
							ConnectionURL: "jdbc:postgresql://localhost:5432/testdb",
							UserName:      "testuser",
							Password:      "testpassword",
						},
						AdditionalProperties: map[string]string{
							"postgresql.fetch-size": "1024",
						},
					},
					Labels: map[string]string{
						"env":     "test",
						"version": "v1",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("pg-catalog-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "PostgreSQL catalog"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "cluster_id"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "id"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "labels.env", "test"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "labels.version", "v1"),
					// Connector
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "postgresql.on_premise.connection_url", "jdbc:postgresql://localhost:5432/testdb"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "postgresql.on_premise.user_name", "testuser"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "postgresql.on_premise.password"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "postgresql.additional_properties.postgresql.fetch-size", "1024"),
				),
			},
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("pg-catalog-%s", randSuffix),
					ConnectorType: "postgresql",
					Postgresql: &PostgresqlConnectorConfig{
						OnPremise: &OnPremise{
							ConnectionURL: "jdbc:postgresql://localhost:5432/newdb",
							UserName:      "user2",
							Password:      "password",
						},
						AdditionalProperties: map[string]string{
							"postgresql.fetch-size": "1024",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("pg-catalog-%s", randSuffix)),
					resource.TestCheckNoResourceAttr("yandex_trino_catalog.trino_catalog", "description"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "cluster_id"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "id"),
					// Connector
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "postgresql.on_premise.connection_url", "jdbc:postgresql://localhost:5432/newdb"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "postgresql.on_premise.user_name", "user2"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "postgresql.on_premise.password"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "postgresql.additional_properties.postgresql.fetch-size", "1024"),
				),
			},
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					Description:   "new-description",
					CatalogName:   fmt.Sprintf("hive-catalog-%s", randSuffix),
					ConnectorType: "hive",
					Hive: &HiveConnectorConfig{
						MetaStoreURI: "thrift://10.10.0.15:9083",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("hive-catalog-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "new-description"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "cluster_id"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "id"),
					// Connector
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "hive.metastore.uri", "thrift://10.10.0.15:9083"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "hive.file_system.s3.%"),
				),
			},
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
		},
	},
	)
}
