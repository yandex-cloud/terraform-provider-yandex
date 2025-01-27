package datasphere_community_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	//dataspheretest "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/tests/datasphere"
)

const testCommunityDataSourceName = "data.yandex_datasphere_community.test-community-data"

func TestAccDatasphereCommunityDataSource(t *testing.T) {
	var (
		communityName = test.ResourceName(63)
		communityDesc = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey      = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue    = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckCommunityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCommunityDataConfig(communityName, communityDesc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					test.CommunityExists(testCommunityDataSourceName),
					resource.TestCheckResourceAttr(testCommunityDataSourceName, "name", communityName),
					resource.TestCheckResourceAttr(testCommunityDataSourceName, "description", communityDesc),
					resource.TestCheckResourceAttr(testCommunityDataSourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
					resource.TestCheckResourceAttrSet(testCommunityDataSourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testCommunityDataSourceName, "created_by"),
					resource.TestCheckResourceAttr(testCommunityDataSourceName, "organization_id", test.GetExampleOrganizationID()),
					test.AccCheckCreatedAtAttr(testCommunityDataSourceName),
				),
			},
		},
	})
}

func testCommunityDataConfig(name string, desc string, labelKey, labelValue string) string {
	return fmt.Sprintf(`
data "yandex_datasphere_community" "test-community-data" {
  id = yandex_datasphere_community.test-community.id
}

resource "yandex_datasphere_community" "test-community" {
  name = "%s"
  description = "%s"
  billing_account_id = "%s"
  labels = {
    "%s": "%s"
  }
  organization_id = "%s"
}`, name, desc, test.GetBillingAccountId(), labelKey, labelValue, test.GetExampleOrganizationID())
}
