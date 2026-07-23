package datalens_workbook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDatalensWorkbook_basic(t *testing.T) {
	t.Parallel()

	title := test.ResourceName(63)
	desc := acctest.RandStringFromCharSet(64, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensWorkbookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensWorkbookConfig_basic(title, desc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatalensWorkbookExists(dlWorkbookResource),
					resource.TestCheckResourceAttr(dlWorkbookResource, "title", title),
					resource.TestCheckResourceAttr(dlWorkbookResource, "description", desc),
					resource.TestCheckResourceAttrSet(dlWorkbookResource, "id"),
					resource.TestCheckResourceAttrSet(dlWorkbookResource, "tenant_id"),
					resource.TestCheckResourceAttrSet(dlWorkbookResource, "created_at"),
					resource.TestCheckResourceAttrSet(dlWorkbookResource, "status"),
				),
			},
			datalensWorkbookImportStep(),
		},
	})
}

func TestAccDatalensWorkbook_update(t *testing.T) {
	t.Parallel()

	title := test.ResourceName(63)
	desc := acctest.RandStringFromCharSet(64, acctest.CharSetAlpha)
	descUpdated := acctest.RandStringFromCharSet(64, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensWorkbookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensWorkbookConfig_basic(title, desc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatalensWorkbookExists(dlWorkbookResource),
					resource.TestCheckResourceAttr(dlWorkbookResource, "description", desc),
				),
			},
			{
				Config: testAccDatalensWorkbookConfig_basic(title, descUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatalensWorkbookExists(dlWorkbookResource),
					resource.TestCheckResourceAttr(dlWorkbookResource, "description", descUpdated),
				),
			},
			datalensWorkbookImportStep(),
		},
	})
}

func testAccDatalensWorkbookConfig_basic(title, description string) string {
	return fmt.Sprintf(`
resource "yandex_datalens_workbook" "test" {
  title           = "%s"
  description     = "%s"
  organization_id = "%s"
}
`, title, description, test.GetExampleOrganizationID())
}
