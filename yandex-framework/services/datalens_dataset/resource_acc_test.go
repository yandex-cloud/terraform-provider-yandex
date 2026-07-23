package datalens_dataset_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDatalensDataset_basic(t *testing.T) {
	t.Parallel()

	saName := test.ResourceName(63)
	dbName := test.ResourceName(63)
	connName := test.ResourceName(63)
	datasetName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensDatasetConfig_basic(saName, dbName, connName, datasetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatalensDatasetExists(dlDatasetResource),
					resource.TestCheckResourceAttr(dlDatasetResource, "name", datasetName),
					resource.TestCheckResourceAttrSet(dlDatasetResource, "id"),
					resource.TestCheckResourceAttr(dlDatasetResource, "dataset.sources.#", "1"),
				),
			},
			datalensDatasetImportStep(),
		},
	})
}

// testAccDatalensDatasetConfig_basic builds a minimal but valid dataset on top
// of a fresh YDB connection. The dataset has one source (a YDB table), one
// avatar, and one DIMENSION result-schema field.
func testAccDatalensDatasetConfig_basic(saName, dbName, connName, datasetName string) string {
	ydbResource, ydbHost, ydbPath := datalensTestYDB(dbName)
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "datalens_test_sa" {
  name      = "%[1]s"
  folder_id = "%[6]s"
}

resource "yandex_resourcemanager_folder_iam_member" "datalens_test_sa_ydb_editor" {
  folder_id = "%[6]s"
  role      = "ydb.editor"
  member    = "serviceAccount:${yandex_iam_service_account.datalens_test_sa.id}"
}
%[8]s
locals {
  ydb_host    = %[9]s
  ydb_db_path = %[10]s
}

resource "yandex_datalens_workbook" "test" {
  title           = "tf-acc-dataset-%[7]s"
  organization_id = "%[4]s"
}

resource "yandex_datalens_connection" "test" {
  name            = "%[3]s"
  type            = "ydb"
  organization_id = "%[4]s"
  depends_on      = [yandex_resourcemanager_folder_iam_member.datalens_test_sa_ydb_editor]

  ydb = {
    host                  = local.ydb_host
    port                  = 2135
    db_name               = local.ydb_db_path
    cloud_id              = "%[5]s"
    folder_id             = "%[6]s"
    service_account_id    = yandex_iam_service_account.datalens_test_sa.id
    workbook_id           = yandex_datalens_workbook.test.id
    data_export_forbidden = "off"
  }
}

resource "yandex_datalens_dataset" "test" {
  name            = "%[7]s"
  organization_id = "%[4]s"
  workbook_id     = yandex_datalens_workbook.test.id

  dataset = {
    sources = [{
      id            = "src-1"
      title         = "subselect"
      source_type   = "YDB_SUBSELECT"
      connection_id = yandex_datalens_connection.test.id
      parameters = {
        subsql = "SELECT 1 AS id, 'hello' AS name"
      }
      raw_schema = [
        {
          name      = "id"
          title     = "id"
          user_type = "integer"
          native_type = {
            name     = "integer"
            nullable = false
          }
        },
        {
          name      = "name"
          title     = "name"
          user_type = "string"
          native_type = {
            name     = "text"
            nullable = false
          }
        },
      ]
    }]
    source_avatars = [{
      id        = "ava-1"
      source_id = "src-1"
      title     = "avatar"
      is_root   = true
    }]
  }
}
`, saName, dbName, connName,
		test.GetExampleOrganizationID(),
		test.GetExampleCloudID(),
		test.GetExampleFolderID(),
		datasetName,
		ydbResource,
		ydbHost,
		ydbPath,
	)
}
