package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Test that an IAM member can be applied to a cloud
func TestAccCloudIamMember_basic(t *testing.T) {
	cloudID := getExampleCloudID()
	role := "viewer"
	userID := getExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Apply an IAM member
			{
				Config: testAccCloudAssociateMemberBasic(cloudID, role, userID),
			},
			{
				ResourceName:      "yandex_resourcemanager_cloud_iam_member.acceptance",
				ImportStateId:     fmt.Sprintf("%s %s %s", cloudID, role, "userAccount:"+userID),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCloudAssociateMemberBasic(cloudID, role, userID string) string {
	prerequisiteMembership, deps := testAccCloudAssignCloudMemberRole(cloudID, userID)

	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_resourcemanager_cloud_iam_member" "acceptance" {
  cloud_id = "%s"
  role     = "%s"
  member   = "userAccount:%s"

  depends_on = [%s]
}
`, cloudID, role, userID, deps)
}
