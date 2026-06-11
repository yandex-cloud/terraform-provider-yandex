package datalens_chart_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDatalensChart_dataSource_basic(t *testing.T) {
	t.Parallel()

	chartName := test.ResourceName(50)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensChartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensChartConfig_wizard(chartName) + fmt.Sprintf(`
data "yandex_datalens_chart" "test" {
  id              = yandex_datalens_chart.test.id
  type            = "wizard"
  organization_id = "%s"
}`, test.GetExampleOrganizationID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dlChartDataSource, "id", dlChartResource, "id"),
					resource.TestCheckResourceAttrPair(dlChartDataSource, "name", dlChartResource, "name"),
				),
			},
		},
	})
}
