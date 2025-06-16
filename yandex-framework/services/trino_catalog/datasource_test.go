package trino_catalog_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	trinov1 "github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceTrinoCatalog_postgresql(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	var catalog trinov1.Catalog

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckTrinoCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTrinoCatalogConfig(t, randSuffix),
				Check: resource.ComposeTestCheckFunc(
					// Postgresql.
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.postgresql", &catalog),
					testAccDataSourceTrinoCatalogAttributesCheck(
						"data.yandex_trino_catalog.postgresql",
						"yandex_trino_catalog.postgresql"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.postgresql", "name", "postgresql"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.postgresql", "description", "PostgreSQL test catalog"),
					resource.TestCheckResourceAttrSet("data.yandex_trino_catalog.postgresql", "cluster_id"),
					resource.TestCheckResourceAttrSet("data.yandex_trino_catalog.postgresql", "id"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.postgresql", "labels.env", "test"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.postgresql", "labels.type", "postgresql"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.postgresql", "postgresql.on_premise.connection_url", "jdbc:postgresql://localhost:5432/testdb"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.postgresql", "postgresql.on_premise.user_name", "testuser"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.postgresql", "postgresql.additional_properties.postgresql.fetch-size", "1024"),

					// Hive
					testAccCheckTrinoCatalogExists("yandex_trino_catalog.hive", &catalog),
					testAccDataSourceTrinoCatalogAttributesCheck(
						"data.yandex_trino_catalog.hive",
						"yandex_trino_catalog.hive"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.hive", "name", "hive"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.hive", "description", "Hive test catalog"),
					resource.TestCheckResourceAttrSet("data.yandex_trino_catalog.hive", "cluster_id"),
					resource.TestCheckResourceAttrSet("data.yandex_trino_catalog.hive", "id"),
					resource.TestCheckResourceAttr("data.yandex_trino_catalog.hive", "hive.metastore.uri", "thrift://10.10.0.15:9083"),
					resource.TestCheckResourceAttrSet("data.yandex_trino_catalog.hive", "hive.file_system.s3.%"),
				),
			},
		},
	})
}

func testAccDataSourceTrinoCatalogAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		if ds.Primary.ID != rs.Primary.ID {
			return fmt.Errorf("trino catalog `data source` ID does not match `resource` ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
		}

		datasourceAttributes := ds.Primary.Attributes
		resourceAttributes := rs.Primary.Attributes

		catalogAttrsToTest := []string{
			"name",
			"cluster_id",
			"description",
			"labels.%",
		}

		for _, attrToCheck := range catalogAttrsToTest {
			if datasourceAttributes[attrToCheck] != resourceAttributes[attrToCheck] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrToCheck,
					datasourceAttributes[attrToCheck],
					resourceAttributes[attrToCheck],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceTrinoCatalogConfig(t *testing.T, randSuffix string) string {
	infraConfig := infraResources(t, randSuffix)

	catalogConfig := `
resource "yandex_trino_catalog" "postgresql" {
  name        = "postgresql"
  cluster_id  = yandex_trino_cluster.trino_cluster.id
  description = "PostgreSQL test catalog"

  labels = {
    env  = "test"
    type = "postgresql"
  }

  postgresql = {
    on_premise = {
      connection_url = "jdbc:postgresql://localhost:5432/testdb"
      user_name      = "testuser"
      password       = "testpassword"
    }

    additional_properties = {
      "postgresql.fetch-size" = "1024"
    }
  }
}

resource "yandex_trino_catalog" "hive" {
  name        = "hive"
  cluster_id  = yandex_trino_cluster.trino_cluster.id
  description = "Hive test catalog"

  hive = {
    metastore = {
      uri = "thrift://10.10.0.15:9083"
    }
    file_system = {
      s3 = {}
    }
  }

}`

	dataSourceConfig := `
data "yandex_trino_catalog" "postgresql" {
  cluster_id = yandex_trino_cluster.trino_cluster.id
  id         = yandex_trino_catalog.postgresql.id
}

data "yandex_trino_catalog" "hive" {
  cluster_id = yandex_trino_cluster.trino_cluster.id
  name       = yandex_trino_catalog.hive.name
}`

	return fmt.Sprintf("%s\n%s\n%s", infraConfig, catalogConfig, dataSourceConfig)
}
