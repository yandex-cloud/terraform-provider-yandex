package datasphere_project_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	fixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datasphere/testfixtures"
)

const testProjectDataSourceName = "data.yandex_datasphere_project.test-project-data"

func TestAccDatasphereProjectDataSource(t *testing.T) {

	var (
		projectName   = test.ResourceName(63)
		communityName = test.ResourceName(63)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             fixtures.AccCheckProjectDestroy(projectClient),
		Steps: []resource.TestStep{
			{
				Config: testDatasphereProjectDataSourceConfig(communityName, projectName),
				Check: resource.ComposeTestCheckFunc(
					fixtures.ProjectExists(projectClient, testProjectDataSourceName),
					resource.TestCheckResourceAttr(testProjectDataSourceName, "name", projectName),
					resource.TestCheckResourceAttrSet(testProjectDataSourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testProjectDataSourceName, "created_by"),
					test.AccCheckCreatedAtAttr(testProjectDataSourceName),
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
`, communityName, test.GetExampleOrganizationID(), test.GetBillingAccountId(), projectName)
}
