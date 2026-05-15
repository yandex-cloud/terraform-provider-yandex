package datalens_workbook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDatalensWorkbook_dataSource_basic(t *testing.T) {
	t.Parallel()

	title := test.ResourceName(63)
	desc := acctest.RandStringFromCharSet(64, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensWorkbookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensWorkbookConfig_basic(title, desc) + fmt.Sprintf(`
data "yandex_datalens_workbook" "test" {
  id              = yandex_datalens_workbook.test.id
  organization_id = "%s"
}`, test.GetExampleOrganizationID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dlWorkbookDataSource, "id", dlWorkbookResource, "id"),
					resource.TestCheckResourceAttrPair(dlWorkbookDataSource, "title", dlWorkbookResource, "title"),
					resource.TestCheckResourceAttrPair(dlWorkbookDataSource, "description", dlWorkbookResource, "description"),
				),
			},
		},
	})
}
