package datalens_dashboard_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDatalensDashboard_basic(t *testing.T) {
	t.Parallel()

	dashName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensDashboardConfig_basic(dashName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dlDashboardResource, "entry.name", dashName),
					resource.TestCheckResourceAttrSet(dlDashboardResource, "id"),
					resource.TestCheckResourceAttr(dlDashboardResource, "entry.data.salt", "abc"),
					resource.TestCheckResourceAttr(dlDashboardResource, "entry.data.tabs.#", "1"),
				),
			},
			datalensDashboardImportStep(),
		},
	})
}

func testAccDatalensDashboardConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "yandex_datalens_workbook" "test" {
  title           = "tf-acc-dashboard-%[1]s"
  organization_id = "%[2]s"
}

resource "yandex_datalens_dashboard" "test" {
  organization_id = "%[2]s"

  entry = {
    name        = "%[1]s"
    workbook_id = yandex_datalens_workbook.test.id

    data = {
      salt = "abc"

      settings = {
        autoupdate_interval     = 30
        max_concurrent_requests = 1
        silent_loading          = false
        dependent_selectors     = false
        expand_toc              = false
      }

      tabs = [
        {
          id    = "tab-1"
          title = "Overview"
          items = []
        },
      ]
    }
  }
}
`, name, test.GetExampleOrganizationID())
}
