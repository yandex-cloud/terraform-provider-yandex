package datalens_dashboard_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDatalensDashboard_dataSource_basic(t *testing.T) {
	t.Parallel()

	dashName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensDashboardConfig_basic(dashName) + fmt.Sprintf(`
data "yandex_datalens_dashboard" "test" {
  id              = yandex_datalens_dashboard.test.id
  organization_id = "%s"
}`, test.GetExampleOrganizationID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dlDashboardDataSource, "id", dlDashboardResource, "id"),
					resource.TestCheckResourceAttrPair(dlDashboardDataSource, "entry.name", dlDashboardResource, "entry.name"),
				),
			},
		},
	})
}
