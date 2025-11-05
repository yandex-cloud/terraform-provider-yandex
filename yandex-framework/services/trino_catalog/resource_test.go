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

	Postgresql *PostgresqlConnectorConfig
	Clickhouse *ClickhouseConnectorConfig

	Hive      *HiveConnectorConfig
	DeltaLake *DeltaLakeConnectorConfig
	Iceberg   *IcebergConnectorConfig
	Hudi      *HudiConnectorConfig

	Oracle    *OracleConnectorConfig
	Sqlserver *SqlserverConnectorConfig
}

type PostgresqlConnectorConfig struct {
	OnPremise            *OnPremise
	AdditionalProperties map[string]string
}

type OnPremise struct {
	ConnectionURL string
	UserName      string
	Password      string
}

type HiveConnectorConfig struct {
	MetaStoreURI         string
	UseExternalS3        bool
	ExternalS3Config     *ExternalS3Config
	AdditionalProperties map[string]string
}

type ClickhouseConnectorConfig struct {
	OnPremise            *OnPremise
	AdditionalProperties map[string]string
}

type DeltaLakeConnectorConfig struct {
	MetaStoreURI         string
	UseExternalS3        bool
	ExternalS3Config     *ExternalS3Config
	AdditionalProperties map[string]string
}

type IcebergConnectorConfig struct {
	MetaStoreURI         string
	UseExternalS3        bool
	ExternalS3Config     *ExternalS3Config
	AdditionalProperties map[string]string
}

type HudiConnectorConfig struct {
	MetaStoreURI         string
	UseExternalS3        bool
	ExternalS3Config     *ExternalS3Config
	AdditionalProperties map[string]string
}

type OracleConnectorConfig struct {
	OnPremise            *OnPremise
	AdditionalProperties map[string]string
}

type SqlserverConnectorConfig struct {
	OnPremise            *OnPremise
	AdditionalProperties map[string]string
}

type TpcdsConnectorConfig struct {
	AdditionalProperties map[string]string
}

type ExternalS3Config struct {
	AwsAccessKey string
	AwsSecretKey string
	AwsEndpoint  string
	AwsRegion    string
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
			{{ if .Hive.UseExternalS3 }}
			external_s3 = {
				aws_access_key = "{{ .Hive.ExternalS3Config.AwsAccessKey }}"
				aws_secret_key = "{{ .Hive.ExternalS3Config.AwsSecretKey }}"
				aws_endpoint   = "{{ .Hive.ExternalS3Config.AwsEndpoint }}"
				aws_region     = "{{ .Hive.ExternalS3Config.AwsRegion }}"
			}
			{{ else }}
			s3 = {}
			{{ end }}
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


	{{ if eq .ConnectorType "clickhouse" }}
	clickhouse = {
		{{ if .Clickhouse.OnPremise }}
    on_premise = {
      connection_url = "{{ .Clickhouse.OnPremise.ConnectionURL }}"
      user_name      = "{{ .Clickhouse.OnPremise.UserName }}"
      password       = "{{ .Clickhouse.OnPremise.Password }}"
    }
		{{ end }}

		{{ if .Clickhouse.AdditionalProperties }}
		additional_properties = {
			{{ range $key, $val := .Clickhouse.AdditionalProperties }}
			"{{ $key }}" = "{{ $val }}"
			{{ end }}
		}
		{{ end }}
  }
  {{ end }}


	{{ if eq .ConnectorType "delta_lake" }}
  delta_lake = {
		metastore = {
			uri = "{{ .DeltaLake.MetaStoreURI }}"
		}
		file_system = {
			{{ if .DeltaLake.UseExternalS3 }}
			external_s3 = {
				aws_access_key = "{{ .DeltaLake.ExternalS3Config.AwsAccessKey }}"
				aws_secret_key = "{{ .DeltaLake.ExternalS3Config.AwsSecretKey }}"
				aws_endpoint   = "{{ .DeltaLake.ExternalS3Config.AwsEndpoint }}"
				aws_region     = "{{ .DeltaLake.ExternalS3Config.AwsRegion }}"
			}
			{{ else }}
			s3 = {}
			{{ end }}
		}

		{{ if .DeltaLake.AdditionalProperties }}
		additional_properties = {
			{{ range $key, $val := .DeltaLake.AdditionalProperties }}
			"{{ $key }}" = "{{ $val }}"
			{{ end }}
		}
		{{ end }}
  }
  {{ end }}


	{{ if eq .ConnectorType "iceberg" }}
  iceberg = {
		metastore = {
			uri = "{{ .Iceberg.MetaStoreURI }}"
		}
		file_system = {
			{{ if .Iceberg.UseExternalS3 }}
			external_s3 = {
				aws_access_key = "{{ .Iceberg.ExternalS3Config.AwsAccessKey }}"
				aws_secret_key = "{{ .Iceberg.ExternalS3Config.AwsSecretKey }}"
				aws_endpoint   = "{{ .Iceberg.ExternalS3Config.AwsEndpoint }}"
				aws_region     = "{{ .Iceberg.ExternalS3Config.AwsRegion }}"
			}
			{{ else }}
			s3 = {}
			{{ end }}
		}

		{{ if .Iceberg.AdditionalProperties }}
		additional_properties = {
			{{ range $key, $val := .Iceberg.AdditionalProperties }}
			"{{ $key }}" = "{{ $val }}"
			{{ end }}
		}
		{{ end }}
  }
  {{ end }}


	{{ if eq .ConnectorType "hudi" }}
  hudi = {
		  metastore = {
			  uri = "{{ .Hudi.MetaStoreURI }}"
		  }
		  file_system = {
			  {{ if .Hudi.UseExternalS3 }}
			  external_s3 = {
				  aws_access_key = "{{ .Hudi.ExternalS3Config.AwsAccessKey }}"
				  aws_secret_key = "{{ .Hudi.ExternalS3Config.AwsSecretKey }}"
				  aws_endpoint   = "{{ .Hudi.ExternalS3Config.AwsEndpoint }}"
				  aws_region     = "{{ .Hudi.ExternalS3Config.AwsRegion }}"
			  }
			  {{ else }}
			  s3 = {}
			  {{ end }}
		  }
  
		  {{ if .Hudi.AdditionalProperties }}
		  additional_properties = {
			  {{ range $key, $val := .Hudi.AdditionalProperties }}
			  "{{ $key }}" = "{{ $val }}"
			  {{ end }}
		  }
		  {{ end }}
	}
	{{ end }}


	{{ if eq .ConnectorType "oracle" }}
  oracle = {
		{{ if .Oracle.OnPremise }}
    on_premise = {
      connection_url = "{{ .Oracle.OnPremise.ConnectionURL }}"
      user_name      = "{{ .Oracle.OnPremise.UserName }}"
      password       = "{{ .Oracle.OnPremise.Password }}"
    }
		{{ end }}

		{{ if .Oracle.AdditionalProperties }}
		additional_properties = {
			{{ range $key, $val := .Oracle.AdditionalProperties }}
			"{{ $key }}" = "{{ $val }}"
			{{ end }}
		}
		{{ end }}
  }
  {{ end }}


	{{ if eq .ConnectorType "sqlserver" }}
  sqlserver = {
		{{ if .Sqlserver.OnPremise }}
    on_premise = {
      connection_url = "{{ .Sqlserver.OnPremise.ConnectionURL }}"
      user_name      = "{{ .Sqlserver.OnPremise.UserName }}"
      password       = "{{ .Sqlserver.OnPremise.Password }}"
    }
		{{ end }}

		{{ if .Sqlserver.AdditionalProperties }}
		additional_properties = {
			{{ range $key, $val := .Sqlserver.AdditionalProperties }}
			"{{ $key }}" = "{{ $val }}"
			{{ end }}
		}
		{{ end }}
  }
  {{ end }}


	{{ if eq .ConnectorType "tpcds" }}
  tpcds = {}
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
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"postgresql.on_premise.password",
			"clickhouse.on_premise.password",
			"oracle.on_premise.password",
			"sqlserver.on_premise.password",
			"oracle.on_premise.password",
			"hive.file_system.external_s3.aws_access_key",
			"hive.file_system.external_s3.aws_secret_key",
			"delta_lake.file_system.external_s3.aws_access_key",
			"delta_lake.file_system.external_s3.aws_secret_key",
			"iceberg.file_system.external_s3.aws_access_key",
			"iceberg.file_system.external_s3.aws_secret_key",
			"hudi.file_system.external_s3.aws_secret_key",
			"hudi.file_system.external_s3.aws_access_key",
		},
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
			// PostgreSQL catalog
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
						UseExternalS3: false,
						MetaStoreURI:  "thrift://10.10.0.15:9083",
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
			// Clickhouse catalog
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("clickhouse-catalog-%s", randSuffix),
					Description:   "Clickhouse catalog",
					ConnectorType: "clickhouse",
					Clickhouse: &ClickhouseConnectorConfig{
						OnPremise: &OnPremise{
							ConnectionURL: "jdbc:clickhouse://localhost:8123/testdb",
							UserName:      "clickhouse_user",
							Password:      "clickhouse_password",
						},
						AdditionalProperties: map[string]string{
							"clickhouse.map-string-as-varchar": "true",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("clickhouse-catalog-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "Clickhouse catalog"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "clickhouse.on_premise.connection_url", "jdbc:clickhouse://localhost:8123/testdb"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "clickhouse.on_premise.user_name", "clickhouse_user"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "clickhouse.on_premise.password"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "clickhouse.additional_properties.clickhouse.map-string-as-varchar", "true"),
				),
			},
			// Oracle catalog
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("oracle-catalog-%s", randSuffix),
					Description:   "Oracle catalog",
					ConnectorType: "oracle",
					Oracle: &OracleConnectorConfig{
						OnPremise: &OnPremise{
							ConnectionURL: "jdbc:oracle:thin:@example.net:1521:XE",
							UserName:      "oracle_user",
							Password:      "oracle_password",
						},
						AdditionalProperties: map[string]string{
							"oracle.connection-pool.max-size": "10",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("oracle-catalog-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "Oracle catalog"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "oracle.on_premise.connection_url", "jdbc:oracle:thin:@example.net:1521:XE"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "oracle.on_premise.user_name", "oracle_user"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "oracle.on_premise.password"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "oracle.additional_properties.oracle.connection-pool.max-size", "10"),
				),
			},
			// SQL Server catalog
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("sqlserver-catalog-%s", randSuffix),
					Description:   "SQL Server catalog",
					ConnectorType: "sqlserver",
					Sqlserver: &SqlserverConnectorConfig{
						OnPremise: &OnPremise{
							ConnectionURL: "jdbc:sqlserver://sqlserver.trino.internal:1433;databaseName=trino;encrypt=false;trustServerCertificate=true",
							UserName:      "sqlserver_user",
							Password:      "sqlserver_password",
						},
						AdditionalProperties: map[string]string{
							"sqlserver.snapshot-isolation.disabled": "false",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("sqlserver-catalog-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "SQL Server catalog"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "sqlserver.on_premise.connection_url", "jdbc:sqlserver://sqlserver.trino.internal:1433;databaseName=trino;encrypt=false;trustServerCertificate=true"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "sqlserver.on_premise.user_name", "sqlserver_user"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "sqlserver.on_premise.password"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "sqlserver.additional_properties.sqlserver.snapshot-isolation.disabled", "false"),
				),
			},
			// TPCDS catalog
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("tpcds-catalog-%s", randSuffix),
					Description:   "TPCDS catalog",
					ConnectorType: "tpcds",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("tpcds-catalog-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "TPCDS catalog"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "tpcds.%"),
				),
			},
			// Delta Lake catalog
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("deltalake-catalog-%s", randSuffix),
					Description:   "Delta Lake catalog",
					ConnectorType: "delta_lake",
					DeltaLake: &DeltaLakeConnectorConfig{
						MetaStoreURI:  "thrift://10.10.0.15:9083",
						UseExternalS3: false,
						AdditionalProperties: map[string]string{
							"delta.enable-non-concurrent-writes": "true",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("deltalake-catalog-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "Delta Lake catalog"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "delta_lake.metastore.uri", "thrift://10.10.0.15:9083"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "delta_lake.file_system.s3.%"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "delta_lake.additional_properties.delta.enable-non-concurrent-writes", "true"),
				),
			},
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("deltalake-external-s3-%s", randSuffix),
					Description:   "Delta Lake with External S3",
					ConnectorType: "delta_lake",
					DeltaLake: &DeltaLakeConnectorConfig{
						MetaStoreURI:  "thrift://10.10.0.20:9083",
						UseExternalS3: true,
						ExternalS3Config: &ExternalS3Config{
							AwsAccessKey: "test-access-key",
							AwsSecretKey: "test-secret-key",
							AwsEndpoint:  "https://storage.yandexcloud.net",
							AwsRegion:    "ru-central1",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("deltalake-external-s3-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "Delta Lake with External S3"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "delta_lake.metastore.uri", "thrift://10.10.0.20:9083"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "delta_lake.file_system.external_s3.aws_access_key"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "delta_lake.file_system.external_s3.aws_secret_key"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "delta_lake.file_system.external_s3.aws_endpoint", "https://storage.yandexcloud.net"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "delta_lake.file_system.external_s3.aws_region", "ru-central1"),
				),
			},
			// Iceberg catalog
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("iceberg-catalog-%s", randSuffix),
					Description:   "Iceberg catalog",
					ConnectorType: "iceberg",
					Iceberg: &IcebergConnectorConfig{
						MetaStoreURI:  "thrift://10.10.0.15:9083",
						UseExternalS3: false,
						AdditionalProperties: map[string]string{
							"iceberg.add-files-procedure.enabled": "true",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("iceberg-catalog-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "Iceberg catalog"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "iceberg.metastore.uri", "thrift://10.10.0.15:9083"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "iceberg.file_system.s3.%"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "iceberg.additional_properties.iceberg.add-files-procedure.enabled", "true"),
				),
			},
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("iceberg-updated-%s", randSuffix),
					Description:   "Iceberg catalog updated",
					ConnectorType: "iceberg",
					Iceberg: &IcebergConnectorConfig{
						MetaStoreURI:  "thrift://10.10.0.30:9083",
						UseExternalS3: true,
						ExternalS3Config: &ExternalS3Config{
							AwsAccessKey: "test-access-key",
							AwsSecretKey: "test-secret-key",
							AwsEndpoint:  "https://storage.yandexcloud.net",
							AwsRegion:    "ru-central1",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("iceberg-updated-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "Iceberg catalog updated"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "iceberg.metastore.uri", "thrift://10.10.0.30:9083"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "iceberg.file_system.external_s3.aws_access_key"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "iceberg.file_system.external_s3.aws_secret_key"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "iceberg.file_system.external_s3.aws_endpoint", "https://storage.yandexcloud.net"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "iceberg.file_system.external_s3.aws_region", "ru-central1"),
				),
			},
			// Hudi catalog
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("hudi-catalog-%s", randSuffix),
					Description:   "Hudi catalog",
					ConnectorType: "hudi",
					Hudi: &HudiConnectorConfig{
						MetaStoreURI:  "thrift://10.10.0.15:9083",
						UseExternalS3: false,
						AdditionalProperties: map[string]string{
							"hudi.parquet.use-column-names": "true",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("hudi-catalog-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "Hudi catalog"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "hudi.metastore.uri", "thrift://10.10.0.15:9083"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "hudi.file_system.s3.%"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "hudi.additional_properties.hudi.parquet.use-column-names", "true"),
				),
			},
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
			{
				Config: trinoCatalogConfig(t, trinoCatalogConfigParams{
					RandSuffix:    randSuffix,
					FolderID:      folderID,
					CatalogName:   fmt.Sprintf("hudi-updated-%s", randSuffix),
					Description:   "Hudi catalog updated",
					ConnectorType: "hudi",
					Hudi: &HudiConnectorConfig{
						MetaStoreURI:  "thrift://10.10.0.30:9083",
						UseExternalS3: true,
						ExternalS3Config: &ExternalS3Config{
							AwsAccessKey: "test-access-key",
							AwsSecretKey: "test-secret-key",
							AwsEndpoint:  "https://storage.yandexcloud.net",
							AwsRegion:    "ru-central1",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.trino_catalog", &catalog),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "name", fmt.Sprintf("hudi-updated-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "description", "Hudi catalog updated"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "hudi.metastore.uri", "thrift://10.10.0.30:9083"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "hudi.file_system.external_s3.aws_access_key"),
					resource.TestCheckResourceAttrSet("yandex_trino_catalog.trino_catalog", "hudi.file_system.external_s3.aws_secret_key"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "hudi.file_system.external_s3.aws_endpoint", "https://storage.yandexcloud.net"),
					resource.TestCheckResourceAttr("yandex_trino_catalog.trino_catalog", "hudi.file_system.external_s3.aws_region", "ru-central1"),
				),
			},
			trinoCatalogImportStep("yandex_trino_catalog.trino_catalog"),
		},
	},
	)
}
