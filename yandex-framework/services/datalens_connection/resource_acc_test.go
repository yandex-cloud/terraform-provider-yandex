package datalens_connection_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDatalensConnection_basic(t *testing.T) {
	t.Parallel()

	saName := test.ResourceName(63)
	dbName := test.ResourceName(63)
	connName := test.ResourceName(63)
	connDesc := acctest.RandStringFromCharSet(64, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensConnectionConfig_basic(saName, dbName, connName, connDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatalensConnectionExists(dlConnectionResource),
					resource.TestCheckResourceAttr(dlConnectionResource, "name", connName),
					resource.TestCheckResourceAttr(dlConnectionResource, "description", connDesc),
					resource.TestCheckResourceAttr(dlConnectionResource, "type", "ydb"),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "id"),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "created_at"),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "updated_at"),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "organization_id"),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "ydb.host"),
					resource.TestCheckResourceAttr(dlConnectionResource, "ydb.port", "2135"),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "ydb.db_name"),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "ydb.service_account_id"),
					resource.TestCheckResourceAttr(dlConnectionResource, "ydb.data_export_forbidden", "off"),
					testAccCheckDatalensConnectionResourceID(dlConnectionResource),
				),
			},
			datalensConnectionImportStep(),
		},
	})
}

func TestAccDatalensConnection_update(t *testing.T) {
	t.Parallel()

	saName := test.ResourceName(63)
	dbName := test.ResourceName(63)
	connName := test.ResourceName(63)
	connDesc := acctest.RandStringFromCharSet(64, acctest.CharSetAlpha)
	connDescUpdated := acctest.RandStringFromCharSet(64, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensConnectionDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccDatalensConnectionConfig_basic(saName, dbName, connName, connDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatalensConnectionExists(dlConnectionResource),
					resource.TestCheckResourceAttr(dlConnectionResource, "name", connName),
					resource.TestCheckResourceAttr(dlConnectionResource, "description", connDesc),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "ydb.host"),
				),
			},
			datalensConnectionImportStep(),
			// Update description
			{
				Config: testAccDatalensConnectionConfig_basic(saName, dbName, connName, connDescUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatalensConnectionExists(dlConnectionResource),
					resource.TestCheckResourceAttr(dlConnectionResource, "name", connName),
					resource.TestCheckResourceAttr(dlConnectionResource, "description", connDescUpdated),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "ydb.host"),
				),
			},
			datalensConnectionImportStep(),
		},
	})
}

func TestAccDatalensConnection_noDescription(t *testing.T) {
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
				Config: testAccDatalensConnectionConfig_noDescription(saName, dbName, connName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatalensConnectionExists(dlConnectionResource),
					resource.TestCheckResourceAttr(dlConnectionResource, "name", connName),
					resource.TestCheckResourceAttr(dlConnectionResource, "type", "ydb"),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "id"),
					resource.TestCheckResourceAttrSet(dlConnectionResource, "created_at"),
				),
			},
			datalensConnectionImportStep(),
		},
	})
}

func testAccDatalensConnectionConfig_basic(saName, dbName, name, description string) string {
	return testAccDatalensConnectionInfraConfig(saName, dbName) + fmt.Sprintf(`
resource "yandex_datalens_connection" "test" {
  name            = "%s"
  description     = "%s"
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
`, name, description, test.GetExampleOrganizationID(), test.GetExampleCloudID(), test.GetExampleFolderID())
}

func testAccDatalensConnectionConfig_noDescription(saName, dbName, name string) string {
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
`, name, test.GetExampleOrganizationID(), test.GetExampleCloudID(), test.GetExampleFolderID())
}
