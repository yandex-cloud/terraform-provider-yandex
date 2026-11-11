package acceptance

import (
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	rmfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

// Test that an IAM member can be applied to a cloud
func TestAccCloudIamMember_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudID := acceptanceenv.CloudID()
	role := "viewer"
	userID := acceptanceenv.UserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply an IAM member
			{
				Config: testAccCloudAssociateMemberBasic(cloudID, role, userID),
			},
			{
				ResourceName:                         "yandex_resourcemanager_cloud_iam_member.acceptance",
				ImportStateId:                        fmt.Sprintf("%s,%s,%s", cloudID, role, "userAccount:"+userID),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "cloud_id",
			},
		},
	})
}

func testAccCloudAssociateMemberBasic(cloudID, role, userID string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID)

	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_resourcemanager_cloud_iam_member" "acceptance" {
  cloud_id = "%s"
  role     = "%s"
  member   = "userAccount:%s"

  depends_on = [%s]
}
`, cloudID, role, userID, deps)
}
