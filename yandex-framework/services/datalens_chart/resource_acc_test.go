package datalens_chart_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDatalensChart_wizard_basic(t *testing.T) {
	t.Parallel()

	chartName := test.ResourceName(50)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensChartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensChartConfig_wizard(chartName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dlChartResource, "name", chartName),
					resource.TestCheckResourceAttr(dlChartResource, "type", "wizard"),
					resource.TestCheckResourceAttrSet(dlChartResource, "id"),
					resource.TestCheckResourceAttr(dlChartResource, "data.visualization.id", "flatTable"),
					resource.TestCheckResourceAttr(dlChartResource, "data.wizard.datasets_ids.#", "1"),
				),
			},
		},
	})
}

func testAccDatalensChartConfig_wizard(name string) string {
	ydbResource, ydbHost, ydbPath := datalensTestYDB(name + "-db")
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "datalens_test_sa" {
  name      = "%[1]s-sa"
  folder_id = "%[3]s"
}

resource "yandex_resourcemanager_folder_iam_member" "datalens_test_sa_ydb_editor" {
  folder_id = "%[3]s"
  role      = "ydb.editor"
  member    = "serviceAccount:${yandex_iam_service_account.datalens_test_sa.id}"
}
%[5]s
locals {
  ydb_host    = %[6]s
  ydb_db_path = %[7]s
}

resource "yandex_datalens_workbook" "test" {
  title           = "tf-acc-chart-%[1]s"
  organization_id = "%[2]s"
}

resource "yandex_datalens_connection" "test" {
  name            = "%[1]s-conn"
  type            = "ydb"
  organization_id = "%[2]s"
  depends_on      = [yandex_resourcemanager_folder_iam_member.datalens_test_sa_ydb_editor]

  ydb = {
    host                  = local.ydb_host
    port                  = 2135
    db_name               = local.ydb_db_path
    cloud_id              = "%[4]s"
    folder_id             = "%[3]s"
    service_account_id    = yandex_iam_service_account.datalens_test_sa.id
    workbook_id           = yandex_datalens_workbook.test.id
    data_export_forbidden = "off"
  }
}

resource "yandex_datalens_dataset" "test" {
  name            = "%[1]s-ds"
  organization_id = "%[2]s"
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
        { name = "id", title = "id", user_type = "integer", native_type = { name = "integer", nullable = false } },
        { name = "name", title = "name", user_type = "string", native_type = { name = "text", nullable = false } },
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

resource "yandex_datalens_chart" "test" {
  name            = "%[1]s"
  organization_id = "%[2]s"
  workbook_id     = yandex_datalens_workbook.test.id

  annotation = {
    description = "tf acc test chart"
  }

  data = {
    visualization = {
      id = "flatTable"
    }

    extra_settings = {
      limit      = 100
      pagination = "on"
    }

    wizard = {
      datasets_ids = [yandex_datalens_dataset.test.id]
    }
  }
}
`, name, test.GetExampleOrganizationID(), test.GetExampleFolderID(), test.GetExampleCloudID(),
		ydbResource, ydbHost, ydbPath)
}
