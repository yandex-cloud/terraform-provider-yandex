package trino_access_control_test

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
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func infraResources(t *testing.T, randSuffix, folderID string) string {
	type params struct {
		RandSuffix string
		FolderID   string
	}
	p := params{
		RandSuffix: randSuffix,
		FolderID:   folderID,
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

resource "yandex_trino_catalog" "tpch" {
  cluster_id = yandex_trino_cluster.trino_cluster.id
  name = "tpch-{{ .RandSuffix }}"
  tpch = {}
}
`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, p))
	return b.String()
}

type trinoAccessControlConfigParams struct {
	RandSuffix                  string
	FolderID                    string
	CatalogRules                []CatalogRule
	SchemaRules                 []SchemaRule
	TableRules                  []TableRule
	FunctionRules               []FunctionRule
	ProcedureRules              []ProcedureRule
	QueryRules                  []QueryRule
	SystemSessionPropertyRules  []SystemSessionPropertyRule
	CatalogSessionPropertyRules []CatalogSessionPropertyRule
}

type CatalogRule struct {
	CatalogIDs    []string
	CatalogRegexp string
	Users         []string
	Groups        []string
	Description   string
	Permission    string
}

type SchemaRule struct {
	CatalogIDs    []string
	CatalogRegexp string
	SchemaNames   []string
	SchemaRegexp  string
	Users         []string
	Groups        []string
	Description   string
	Owner         string
}

type TableRule struct {
	CatalogIDs    []string
	CatalogRegexp string
	SchemaNames   []string
	SchemaRegexp  string
	TableNames    []string
	TableRegexp   string
	Columns       []ColumnRule
	Users         []string
	Groups        []string
	Description   string
	Filter        string
	Privileges    []string
}

type ColumnRule struct {
	Name   string
	Mask   string
	Access string
}

type FunctionRule struct {
	CatalogIDs     []string
	CatalogRegexp  string
	SchemaNames    []string
	SchemaRegexp   string
	FunctionNames  []string
	FunctionRegexp string
	Users          []string
	Groups         []string
	Description    string
	Privileges     []string
}

type ProcedureRule struct {
	CatalogIDs      []string
	CatalogRegexp   string
	SchemaNames     []string
	SchemaRegexp    string
	ProcedureNames  []string
	ProcedureRegexp string
	Users           []string
	Groups          []string
	Description     string
	Privileges      []string
}

type QueryRule struct {
	Users       []string
	Groups      []string
	QueryOwners []string
	Description string
	Privileges  []string
}

type SystemSessionPropertyRule struct {
	PropertyNames  []string
	PropertyRegexp string
	Users          []string
	Groups         []string
	Description    string
	Allow          string
}

type CatalogSessionPropertyRule struct {
	CatalogIDs     []string
	CatalogRegexp  string
	PropertyNames  []string
	PropertyRegexp string
	Users          []string
	Groups         []string
	Description    string
	Allow          string
}

func trinoAccessControlConfig(t *testing.T, params trinoAccessControlConfigParams) string {
	tpl, err := template.New("trino_access_control").Parse(`
resource "yandex_trino_access_control" "trino_access_control" {
  cluster_id = yandex_trino_cluster.trino_cluster.id
  
  {{ if .CatalogRules }}
  catalogs = [
    {{ range .CatalogRules }}
    {
      {{ if .CatalogIDs }}
      catalog = {
        ids = [
		  {{ range .CatalogIDs }}
		  {{ . }},
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .CatalogRegexp }}
      catalog = {
        name_regexp = "{{ .CatalogRegexp }}"
      }
      {{ end }}
      {{ if .Users }}
      users = [
	  	{{ range .Users }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Groups }}
      groups = [
	  	{{ range .Groups }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Description }}
      description = "{{ .Description }}"
      {{ end }}
      permission = "{{ .Permission }}"
    },
    {{ end }}
  ]
  {{ end }}

  {{ if .SchemaRules }}
  schemas = [
    {{ range .SchemaRules }}
    {
      {{ if .CatalogIDs }}
      catalog = {
        ids = [
		  {{ range .CatalogIDs }}
		  {{ . }},
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .CatalogRegexp }}
      catalog = {
        name_regexp = "{{ .CatalogRegexp }}"
      }
      {{ end }}
      {{ if .SchemaNames }}
      schema = {
        names = [
		  {{ range .SchemaNames }}
		  "{{ . }}",
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .SchemaRegexp }}
      schema = {
        name_regexp = "{{ .SchemaRegexp }}"
      }
      {{ end }}
      {{ if .Users }}
      users = [
	  	{{ range .Users }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Groups }}
      groups = [
	  	{{ range .Groups }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Description }}
      description = "{{ .Description }}"
      {{ end }}
      owner = "{{ .Owner }}"
    },
    {{ end }}
  ]
  {{ end }}

  {{ if .TableRules }}
  tables = [
    {{ range .TableRules }}
    {
      {{ if .CatalogIDs }}
      catalog = {
        ids = [
		  {{ range .CatalogIDs }}
		  {{ . }},
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .CatalogRegexp }}
      catalog = {
        name_regexp = "{{ .CatalogRegexp }}"
      }
      {{ end }}
      {{ if .SchemaNames }}
      schema = {
        names = [
		  {{ range .SchemaNames }}
		  "{{ . }}",
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .SchemaRegexp }}
      schema = {
        name_regexp = "{{ .SchemaRegexp }}"
      }
      {{ end }}
      {{ if .TableNames }}
      table = {
        names = [
		  {{ range .TableNames }}
		  "{{ . }}",
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .TableRegexp }}
      table = {
        name_regexp = "{{ .TableRegexp }}"
      }
      {{ end }}
      {{ if .Users }}
      users = [
	  	{{ range .Users }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Groups }}
      groups = [
	  	{{ range .Groups }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Description }}
      description = "{{ .Description }}"
      {{ end }}
      {{ if .Filter }}
      filter = "{{ .Filter }}"
      {{ end }}
      {{ if .Privileges }}
      privileges = [
	  	{{ range .Privileges }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Columns }}
      columns = [
        {{ range .Columns }}
        {
          name = "{{ .Name }}"
          access = "{{ .Access }}"
          {{ if .Mask }}
          mask = "{{ .Mask }}"
          {{ end }}
        },
        {{ end }}
      ]
      {{ end }}
    },
    {{ end }}
  ]
  {{ end }}

  {{ if .FunctionRules }}
  functions = [
    {{ range .FunctionRules }}
    {
      {{ if .CatalogIDs }}
      catalog = {
        ids = [
		  {{ range .CatalogIDs }}
		  {{ . }},
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .CatalogRegexp }}
      catalog = {
        name_regexp = "{{ .CatalogRegexp }}"
      }
      {{ end }}
      {{ if .SchemaNames }}
      schema = {
        names = [
		  {{ range .SchemaNames }}
		  "{{ . }}",
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .SchemaRegexp }}
      schema = {
        name_regexp = "{{ .SchemaRegexp }}"
      }
      {{ end }}
      {{ if .FunctionNames }}
      function = {
        names = [
		  {{ range .FunctionNames }}
		  "{{ . }}",
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .FunctionRegexp }}
      function = {
        name_regexp = "{{ .FunctionRegexp }}"
      }
      {{ end }}
      {{ if .Users }}
      users = [
	  	{{ range .Users }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Groups }}
      groups = [
	  	{{ range .Groups }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Description }}
      description = "{{ .Description }}"
      {{ end }}
      {{ if .Privileges }}
      privileges = [
	  	{{ range .Privileges }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
    },
    {{ end }}
  ]
  {{ end }}

  {{ if .ProcedureRules }}
  procedures = [
    {{ range .ProcedureRules }}
    {
      {{ if .CatalogIDs }}
      catalog = {
        ids = [
		  {{ range .CatalogIDs }}
		  {{ . }},
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .CatalogRegexp }}
      catalog = {
        name_regexp = "{{ .CatalogRegexp }}"
      }
      {{ end }}
      {{ if .SchemaNames }}
      schema = {
        names = [
		  {{ range .SchemaNames }}
		  "{{ . }}",
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .SchemaRegexp }}
      schema = {
        name_regexp = "{{ .SchemaRegexp }}"
      }
      {{ end }}
      {{ if .ProcedureNames }}
      procedure = {
        names = [
		  {{ range .ProcedureNames }}
		  "{{ . }}",
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .ProcedureRegexp }}
      procedure = {
        name_regexp = "{{ .ProcedureRegexp }}"
      }
      {{ end }}
      {{ if .Users }}
      users = [
	  	{{ range .Users }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Groups }}
      groups = [
	  	{{ range .Groups }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Description }}
      description = "{{ .Description }}"
      {{ end }}
      {{ if .Privileges }}
      privileges = [
	  	{{ range .Privileges }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
    },
    {{ end }}
  ]
  {{ end }}

  {{ if .QueryRules }}
  queries = [
    {{ range .QueryRules }}
    {
      {{ if .Users }}
      users = [
	  	{{ range .Users }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Groups }}
      groups = [
	  	{{ range .Groups }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
	  {{ if .QueryOwners }}
      query_owners = [
	  	{{ range .QueryOwners }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Description }}
      description = "{{ .Description }}"
      {{ end }}
      {{ if .Privileges }}
      privileges = [
	  	{{ range .Privileges }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
    },
    {{ end }}
  ]
  {{ end }}

  {{ if .SystemSessionPropertyRules }}
  system_session_properties = [
    {{ range .SystemSessionPropertyRules }}
    {
      {{ if .PropertyNames }}
      property = {
        names = [
		  {{ range .PropertyNames }}
		  "{{ . }}",
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .PropertyRegexp }}
      property = {
        name_regexp = "{{ .PropertyRegexp }}"
      }
      {{ end }}
      {{ if .Users }}
      users = [
	  	{{ range .Users }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Groups }}
      groups = [
	  	{{ range .Groups }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Description }}
      description = "{{ .Description }}"
      {{ end }}
      allow = "{{ .Allow }}"
    },
    {{ end }}
  ]
  {{ end }}

  {{ if .CatalogSessionPropertyRules }}
  catalog_session_properties = [
    {{ range .CatalogSessionPropertyRules }}
    {
      {{ if .CatalogIDs }}
      catalog = {
        ids = [
		  {{ range .CatalogIDs }}
		  {{ . }},
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .CatalogRegexp }}
      catalog = {
        name_regexp = "{{ .CatalogRegexp }}"
      }
      {{ end }}
      {{ if .PropertyNames }}
      property = {
        names = [
		  {{ range .PropertyNames }}
		  "{{ . }}",
		  {{ end }}
		]
      }
      {{ end }}
      {{ if .PropertyRegexp }}
      property = {
        name_regexp = "{{ .PropertyRegexp }}"
      }
      {{ end }}
      {{ if .Users }}
      users = [
	  	{{ range .Users }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Groups }}
      groups = [
	  	{{ range .Groups }}
		"{{ . }}",
		{{ end }}
	  ]
      {{ end }}
      {{ if .Description }}
      description = "{{ .Description }}"
      {{ end }}
      allow = "{{ .Allow }}"
    },
    {{ end }}
  ]
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

	return fmt.Sprintf("%s\n%s", infraResources(t, params.RandSuffix, params.FolderID), b.String())
}

func testAccCheckTrinoAccessControlDestroy(s *terraform.State) error {
	sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_trino_access_control" {
			continue
		}

		clusterID := rs.Primary.Attributes["cluster_id"]
		cluster, err := sdk.Trino().Cluster().Get(context.Background(), &trinov1.GetClusterRequest{
			ClusterId: clusterID,
		})

		if err != nil {
			if status.Code(err) == codes.NotFound {
				return nil
			}
			return fmt.Errorf("get Trino cluster: %w", err)
		}

		if cluster.Trino.AccessControl != nil {
			return fmt.Errorf("Trino cluster access control still exists")
		}
	}

	return nil
}

func trinoAccessControlImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:                         name,
		ImportState:                          true,
		ImportStateVerify:                    true,
		ImportStateVerifyIdentifierAttribute: "cluster_id",
		ImportStateIdFunc: func(s *terraform.State) (string, error) {
			rs, ok := s.RootModule().Resources[name]
			if !ok {
				return "", fmt.Errorf("resource not found: %s", name)
			}
			clusterID := rs.Primary.Attributes["cluster_id"]
			return clusterID, nil
		},
	}
}

func testAccCheckTrinoAccessControlExists(name string) resource.TestCheckFunc {
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
		found, err := sdk.Trino().Cluster().Get(context.Background(), &trinov1.GetClusterRequest{
			ClusterId: clusterID,
		})
		if err != nil {
			return err
		}
		if found.Trino.AccessControl == nil {
			return fmt.Errorf("Trino cluster access control does not exist")
		}
		return nil
	}
}

func TestAccMDBTrinoAccessControl_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	folderID := os.Getenv("YC_FOLDER_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckTrinoAccessControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: trinoAccessControlConfig(t, trinoAccessControlConfigParams{
					RandSuffix: randSuffix,
					FolderID:   folderID,
					CatalogRules: []CatalogRule{
						{
							CatalogIDs:  []string{"yandex_trino_catalog.tpch.id"},
							Users:       []string{"u1", "u2"},
							Groups:      []string{"g1", "g2"},
							Description: "catalog rule",
							Permission:  "ALL",
						},
						{
							CatalogRegexp: "tpch.*",
							Permission:    "READ_ONLY",
						},
						{
							Permission: "NONE",
						},
					},
					SchemaRules: []SchemaRule{
						{
							CatalogIDs:  []string{"yandex_trino_catalog.tpch.id"},
							SchemaNames: []string{"information_schema"},
							Users:       []string{"u1", "u2"},
							Groups:      []string{"g1", "g2"},
							Description: "schema rule",
							Owner:       "YES",
						},
						{
							CatalogRegexp: "tpch.*",
							SchemaRegexp:  ".*",
							Owner:         "NO",
						},
						{
							Owner: "NO",
						},
					},
					TableRules: []TableRule{
						{
							CatalogIDs:  []string{"yandex_trino_catalog.tpch.id"},
							SchemaNames: []string{"information_schema"},
							TableNames:  []string{"t1", "t2"},
							Users:       []string{"u1", "u2"},
							Groups:      []string{"g1", "g2"},
							Columns: []ColumnRule{
								{
									Name:   "c1",
									Mask:   "substring(c1, -2)",
									Access: "ALL",
								},
								{
									Name:   "c2",
									Access: "NONE",
								},
							},
							Filter:      "year > 2020",
							Description: "table rule",
							Privileges:  []string{"SELECT", "UPDATE", "DELETE", "INSERT", "OWNERSHIP", "GRANT_SELECT"},
						},
						{
							CatalogRegexp: "tpch.*",
							SchemaRegexp:  ".*",
							TableRegexp:   "t.*",
							Privileges:    []string{"SELECT"},
						},
						{},
					},
					FunctionRules: []FunctionRule{
						{
							CatalogIDs:    []string{"yandex_trino_catalog.tpch.id"},
							SchemaNames:   []string{"information_schema"},
							FunctionNames: []string{"t1", "t2"},
							Users:         []string{"u1", "u2"},
							Groups:        []string{"g1", "g2"},
							Description:   "function rule",
							Privileges:    []string{"EXECUTE", "GRANT_EXECUTE", "OWNERSHIP"},
						},
						{
							CatalogRegexp:  "tpch.*",
							SchemaRegexp:   ".*",
							FunctionRegexp: "t.*",
							Privileges:     []string{"EXECUTE"},
						},
						{},
					},
					ProcedureRules: []ProcedureRule{
						{
							CatalogIDs:     []string{"yandex_trino_catalog.tpch.id"},
							SchemaNames:    []string{"information_schema"},
							ProcedureNames: []string{"t1", "t2"},
							Users:          []string{"u1", "u2"},
							Groups:         []string{"g1", "g2"},
							Description:    "procedure rule",
							Privileges:     []string{"EXECUTE"},
						},
						{
							CatalogRegexp:   "tpch.*",
							SchemaRegexp:    ".*",
							ProcedureRegexp: "t.*",
							Privileges:      []string{"EXECUTE"},
						},
						{},
					},
					QueryRules: []QueryRule{
						{
							Users:       []string{"u1", "u2"},
							Groups:      []string{"g1", "g2"},
							QueryOwners: []string{"u3", "u4"},
							Description: "query rule",
							Privileges:  []string{"VIEW", "KILL"},
						},
						{
							Privileges: []string{"EXECUTE"},
						},
						{},
					},
					SystemSessionPropertyRules: []SystemSessionPropertyRule{
						{
							PropertyNames: []string{"a", "b"},
							Users:         []string{"u1", "u2"},
							Groups:        []string{"g1", "g2"},
							Description:   "system session property rule",
							Allow:         "YES",
						},
						{
							PropertyRegexp: ".*",
							Allow:          "NO",
						},
					},
					CatalogSessionPropertyRules: []CatalogSessionPropertyRule{
						{
							CatalogIDs:    []string{"yandex_trino_catalog.tpch.id"},
							PropertyNames: []string{"a", "b"},
							Users:         []string{"u1", "u2"},
							Groups:        []string{"g1", "g2"},
							Description:   "system session property rule",
							Allow:         "YES",
						},
						{
							CatalogRegexp:  ".*",
							PropertyRegexp: ".*",
							Allow:          "NO",
						},
					},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrinoAccessControlExists("yandex_trino_access_control.trino_access_control"),
					resource.TestCheckResourceAttrSet("yandex_trino_access_control.trino_access_control", "cluster_id"),

					// catalogs rules
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.#", "3"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.users.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.users.0", "u1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.users.1", "u2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.groups.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.groups.0", "g1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.groups.1", "g2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.description", "catalog rule"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.permission", "ALL"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.1.catalog.name_regexp", "tpch.*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.1.permission", "READ_ONLY"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.2.permission", "NONE"),

					// schema rules
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.#", "3"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.catalog.ids.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_trino_access_control.trino_access_control", "schemas.0.catalog.ids.0"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.schema.names.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.schema.names.0", "information_schema"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.users.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.users.0", "u1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.users.1", "u2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.groups.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.groups.0", "g1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.groups.1", "g2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.description", "schema rule"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.0.owner", "YES"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.1.catalog.name_regexp", "tpch.*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.1.schema.name_regexp", ".*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.1.owner", "NO"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "schemas.2.owner", "NO"),

					// table rules
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.#", "3"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.catalog.ids.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_trino_access_control.trino_access_control", "tables.0.catalog.ids.0"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.schema.names.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.schema.names.0", "information_schema"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.table.names.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.table.names.0", "t1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.table.names.1", "t2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.users.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.users.0", "u1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.users.1", "u2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.groups.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.groups.0", "g1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.groups.1", "g2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.columns.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.columns.0.name", "c1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.columns.0.mask", "substring(c1, -2)"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.columns.0.access", "ALL"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.columns.1.name", "c2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.columns.1.access", "NONE"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.filter", "year > 2020"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.description", "table rule"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.privileges.#", "6"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.privileges.0", "SELECT"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.privileges.1", "UPDATE"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.privileges.2", "DELETE"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.privileges.3", "INSERT"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.privileges.4", "OWNERSHIP"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.0.privileges.5", "GRANT_SELECT"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.1.catalog.name_regexp", "tpch.*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.1.schema.name_regexp", ".*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.1.table.name_regexp", "t.*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.1.privileges.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.1.privileges.0", "SELECT"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "tables.2.%", "9"),

					// function rules
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.#", "3"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.catalog.ids.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_trino_access_control.trino_access_control", "functions.0.catalog.ids.0"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.schema.names.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.schema.names.0", "information_schema"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.function.names.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.function.names.0", "t1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.function.names.1", "t2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.users.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.users.0", "u1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.users.1", "u2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.groups.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.groups.0", "g1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.groups.1", "g2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.description", "function rule"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.privileges.#", "3"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.privileges.0", "EXECUTE"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.privileges.1", "GRANT_EXECUTE"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.0.privileges.2", "OWNERSHIP"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.1.catalog.name_regexp", "tpch.*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.1.schema.name_regexp", ".*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.1.function.name_regexp", "t.*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.1.privileges.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.1.privileges.0", "EXECUTE"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "functions.2.%", "7"),

					// procedure rules
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.#", "3"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.catalog.ids.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_trino_access_control.trino_access_control", "procedures.0.catalog.ids.0"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.schema.names.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.schema.names.0", "information_schema"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.procedure.names.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.procedure.names.0", "t1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.procedure.names.1", "t2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.users.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.users.0", "u1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.users.1", "u2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.groups.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.groups.0", "g1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.groups.1", "g2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.description", "procedure rule"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.privileges.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.0.privileges.0", "EXECUTE"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.1.catalog.name_regexp", "tpch.*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.1.schema.name_regexp", ".*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.1.procedure.name_regexp", "t.*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.1.privileges.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.1.privileges.0", "EXECUTE"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "procedures.2.%", "7"),

					// query rules
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.#", "3"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.users.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.users.0", "u1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.users.1", "u2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.groups.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.groups.0", "g1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.groups.1", "g2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.query_owners.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.query_owners.0", "u3"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.query_owners.1", "u4"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.description", "query rule"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.privileges.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.privileges.0", "VIEW"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.0.privileges.1", "KILL"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.1.privileges.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.1.privileges.0", "EXECUTE"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "queries.2.%", "5"),

					// system session property rules
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.users.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.users.0", "u1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.users.1", "u2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.groups.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.groups.0", "g1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.groups.1", "g2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.property.names.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.property.names.0", "a"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.property.names.1", "b"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.description", "system session property rule"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.0.allow", "YES"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.1.property.name_regexp", ".*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties.1.allow", "NO"),

					// catalog session property rules
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.catalog.ids.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.catalog.ids.0"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.users.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.users.0", "u1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.users.1", "u2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.groups.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.groups.0", "g1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.groups.1", "g2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.property.names.#", "2"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.property.names.0", "a"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.property.names.1", "b"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.description", "system session property rule"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.0.allow", "YES"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.1.catalog.name_regexp", ".*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.1.property.name_regexp", ".*"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties.1.allow", "NO"),

					// Check timeouts
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "timeouts.create", "50m"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "timeouts.update", "50m"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "timeouts.delete", "50m"),
				),
			},
			trinoAccessControlImportStep("yandex_trino_access_control.trino_access_control"),
			{
				Config: trinoAccessControlConfig(t, trinoAccessControlConfigParams{
					RandSuffix: randSuffix,
					FolderID:   folderID,
					CatalogRules: []CatalogRule{
						{
							Description: "updated rule",
							Permission:  "READ_ONLY",
						},
					},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrinoAccessControlExists("yandex_trino_access_control.trino_access_control"),
					resource.TestCheckResourceAttrSet("yandex_trino_access_control.trino_access_control", "cluster_id"),

					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.#", "1"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.description", "updated rule"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "catalogs.0.permission", "READ_ONLY"),

					resource.TestCheckNoResourceAttr("yandex_trino_access_control.trino_access_control", "schemas"),
					resource.TestCheckNoResourceAttr("yandex_trino_access_control.trino_access_control", "tables"),
					resource.TestCheckNoResourceAttr("yandex_trino_access_control.trino_access_control", "functions"),
					resource.TestCheckNoResourceAttr("yandex_trino_access_control.trino_access_control", "procedures"),
					resource.TestCheckNoResourceAttr("yandex_trino_access_control.trino_access_control", "queries"),
					resource.TestCheckNoResourceAttr("yandex_trino_access_control.trino_access_control", "system_session_properties"),
					resource.TestCheckNoResourceAttr("yandex_trino_access_control.trino_access_control", "catalog_session_properties"),

					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "timeouts.create", "50m"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "timeouts.update", "50m"),
					resource.TestCheckResourceAttr("yandex_trino_access_control.trino_access_control", "timeouts.delete", "50m"),
				),
			},
			trinoAccessControlImportStep("yandex_trino_access_control.trino_access_control"),
		},
	})
}
