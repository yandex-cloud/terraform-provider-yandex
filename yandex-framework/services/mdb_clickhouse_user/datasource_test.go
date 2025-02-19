package mdb_clickhouse_user_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceMDBClickHouseUser_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-clickhouse-user-datasource")
	description := "Clickhouse User Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBClickHouseUserConfig(clusterName, description),
				Check: testAccDataSourceMDBClickHouseUserCheck(
					"data.yandex_mdb_clickhouse_user.dsleonardo", "yandex_mdb_clickhouse_user.leonardo",
				),
			},
		},
	})
}

func testAccDataSourceMDBClickHouseUserAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
			return fmt.Errorf("instance `data source` ID does not match `resource` ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
		}

		datasourceAttributes := ds.Primary.Attributes
		resourceAttributes := rs.Primary.Attributes

		instanceAttrsToTest := []struct {
			dataSourcePath string
			resourcePath   string
		}{
			{
				"cluster_id",
				"cluster_id",
			},
			{
				"name",
				"name",
			},
			{
				"permission.#",
				"permission.#",
			},
			{
				"quota.#",
				"quota.#",
			},
			{
				"settings.%",
				"settings.%",
			},
		}

		log.Printf("[DEBUG] data attributes: %v\n", datasourceAttributes)
		log.Printf("[DEBUG] resource attributes: %v\n", resourceAttributes)
		for _, attrToCheck := range instanceAttrsToTest {
			if _, ok := datasourceAttributes[attrToCheck.dataSourcePath]; !ok {
				return fmt.Errorf("%s is not present in data source attributes", attrToCheck.dataSourcePath)
			}
			if _, ok := resourceAttributes[attrToCheck.resourcePath]; !ok {
				return fmt.Errorf("%s is not present in resource attributes", attrToCheck.resourcePath)
			}
			if datasourceAttributes[attrToCheck.dataSourcePath] != resourceAttributes[attrToCheck.resourcePath] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrToCheck.dataSourcePath,
					datasourceAttributes[attrToCheck.dataSourcePath],
					resourceAttributes[attrToCheck.resourcePath],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceMDBClickHouseUserCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBClickHouseUserAttributesCheck(datasourceName, resourceName),
		testAccCheckMDBClickHouseUserResourceIDField(resourceName),
		resource.TestCheckResourceAttr(datasourceName, "name", "leonardo"),
	)
}

func testAccDataSourceMDBClickHouseUserConfig(name string, description string) string {
	return testAccMDBClickHouseClusterConfigMain(name, description) + `

	resource "yandex_mdb_clickhouse_user" "leonardo" {
		cluster_id = yandex_mdb_clickhouse_cluster.sewage.id
		name       = "leonardo"
		password   = "mysecureP@ssw0rd"
		permission {
	      database_name = yandex_mdb_clickhouse_database.pepperoni.name
	  	}
		quota {
		  interval_duration = 79800000
		  queries           = 5000
		}
		settings {
          readonly = 0
          allow_ddl = true
          connect_timeout = 30000
          distributed_product_mode = "local"
          join_algorithm = [ "partial_merge", "full_sorting_merge" ]
          max_block_size = 5008
		}

	}

	data "yandex_mdb_clickhouse_user" "dsleonardo" {
		cluster_id = yandex_mdb_clickhouse_cluster.sewage.id
		name       = yandex_mdb_clickhouse_user.leonardo.name
	}
	
	`
}
