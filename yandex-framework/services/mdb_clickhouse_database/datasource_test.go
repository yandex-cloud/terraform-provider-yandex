package mdb_clickhouse_database_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceMDBClickHouseDatabase_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-clickhouse-database-datasource")
	description := "ClickHouse database terraform datasource test"

	resourceName := fmt.Sprintf("yandex_mdb_clickhouse_database.%s", chDBResourceName1)
	dataSourceName := fmt.Sprintf("data.yandex_mdb_clickhouse_database.%s", chDBResourceName1)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBClickHouseDatabaseDatasourceConfig(clusterName, description, chDBResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseDatabaseResourceIDField(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", chDBResourceName1),
					resource.TestCheckResourceAttr(dataSourceName, "name", chDBResourceName1),
				),
			},
		},
	})
}

func testAccDataSourceMDBClickHouseDatabaseDatasourceConfig(name, description, dbName string) string {
	return testAccMDBClickHouseClusterConfigMain(name, description) + fmt.Sprintf(`

	resource "yandex_mdb_clickhouse_database" "%[2]s" {
		cluster_id = %[1]s
		name       = "%[2]s"
	}

	data "yandex_mdb_clickhouse_database" "%[2]s" {
		cluster_id = %[1]s
		name       = yandex_mdb_clickhouse_database.%[2]s.name
	}
	
	`, chClusterResourceIDLink, dbName)
}
