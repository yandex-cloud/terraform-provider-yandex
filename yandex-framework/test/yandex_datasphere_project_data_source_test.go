package test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const testProjectDataSourceName = "data.yandex_datasphere_project.test-project-data"

func TestAccDatasphereProjectDataSource(t *testing.T) {

	projectName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	communityName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasphereProjectDataSourceConfig(communityName, projectName),
				Check: resource.ComposeTestCheckFunc(
					testDatasphereProjectExists(testProjectDataSourceName),
					resource.TestCheckResourceAttr(testProjectDataSourceName, "name", projectName),
					resource.TestCheckResourceAttrSet(testProjectDataSourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testProjectDataSourceName, "created_by"),
					testAccCheckCreatedAtAttr(testProjectDataSourceName),
				),
			},
		},
	})
}

func testDatasphereProjectDataSourceConfig(communityName, projectName string) string {
	return fmt.Sprintf(`
data "yandex_datasphere_project" "test-project-data" {
  id = yandex_datasphere_project.test-project.id
}

resource "yandex_datasphere_community" "test-community" {
  name = "%s"
  organization_id = "%s"
  billing_account_id = "%s"
}

resource "yandex_datasphere_project" "test-project" {
  name = "%s"
  community_id = yandex_datasphere_community.test-community.id
}
`, communityName, getExampleOrganizationID(), getBillingAccountId(), projectName)
}
