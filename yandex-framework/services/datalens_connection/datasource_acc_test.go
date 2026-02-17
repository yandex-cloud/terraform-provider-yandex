package datalens_connection_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceDatalensConnection_basic(t *testing.T) {
	t.Parallel()

	saName := test.ResourceName(63)
	dbName := test.ResourceName(63)
	connName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensConnectionDataSourceConfig(saName, dbName, connName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatalensConnectionExists(dlConnectionDataSource),
					// Verify data source attributes match the resource
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "id",
						dlConnectionResource, "id",
					),
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "name",
						dlConnectionResource, "name",
					),
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "type",
						dlConnectionResource, "type",
					),
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "created_at",
						dlConnectionResource, "created_at",
					),
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "updated_at",
						dlConnectionResource, "updated_at",
					),
					// Verify YDB nested attributes
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "ydb.host",
						dlConnectionResource, "ydb.host",
					),
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "ydb.port",
						dlConnectionResource, "ydb.port",
					),
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "ydb.db_name",
						dlConnectionResource, "ydb.db_name",
					),
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "ydb.cloud_id",
						dlConnectionResource, "ydb.cloud_id",
					),
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "ydb.folder_id",
						dlConnectionResource, "ydb.folder_id",
					),
					resource.TestCheckResourceAttrPair(
						dlConnectionDataSource, "ydb.service_account_id",
						dlConnectionResource, "ydb.service_account_id",
					),
				),
			},
		},
	})
}

// testAccDatalensConnectionDataSourceConfig creates infrastructure (SA + YDB),
// a connection resource, and then reads it back via a data source.
func testAccDatalensConnectionDataSourceConfig(saName, dbName, name string) string {
	return testAccDatalensConnectionInfraConfig(saName, dbName) + fmt.Sprintf(`
resource "yandex_datalens_connection" "test" {
  name            = "%s"
  type            = "ydb"
  organization_id = "%s"

  depends_on = [yandex_resourcemanager_folder_iam_member.datalens_test_sa_ydb_editor]

  ydb = {
    host                  = local.ydb_host
    port                  = 2135
    db_name               = yandex_ydb_database_serverless.test.database_path
    cloud_id              = "%s"
    folder_id             = "%s"
    service_account_id    = yandex_iam_service_account.datalens_test_sa.id
    workbook_id           = "c7e0nsirdp1cw"
    data_export_forbidden = "off"
  }
}

data "yandex_datalens_connection" "test" {
  id              = yandex_datalens_connection.test.id
  organization_id = "%s"
}
`, name, test.GetExampleOrganizationID(), test.GetExampleCloudID(), test.GetExampleFolderID(), test.GetExampleOrganizationID())
}
