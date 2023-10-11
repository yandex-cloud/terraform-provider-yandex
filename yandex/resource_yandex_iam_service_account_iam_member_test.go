package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func TestAccServiceAccountIamMember(t *testing.T) {
	var serviceAccount iam.ServiceAccount
	serviceAccountName := acctest.RandomWithPrefix("tf-test")
	cloudID := getExampleCloudID()
	userID := getExampleUserID1()
	role := "editor"
	identity := "userAccount:" + userID

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountIamMember_basic(cloudID, serviceAccountName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountExistsWithID("yandex_iam_service_account.test_account", &serviceAccount),
					testAccCheckServiceAccountIam("yandex_iam_service_account.test_account", role, []string{identity}),
				),
			},
			{
				ResourceName: "yandex_iam_service_account_iam_member.foo",
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return fmt.Sprintf("%s %s %s", serviceAccount.Id, role, identity), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//revive:disable:var-naming
func testAccServiceAccountIamMember_basic(cloudID, accountName, role, userID string) string {
	prerequisiteMembership, deps := testAccCloudAssignCloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_iam_service_account" "test_account" {
  name        = "%s"
  description = "Iam Testing Account"
}

resource "yandex_iam_service_account_iam_member" "foo" {
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  role               = "%s"
  member             = "userAccount:%s"

  depends_on = [%s]
}
`, accountName, role, userID, deps)
}
