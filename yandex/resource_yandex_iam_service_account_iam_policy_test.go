package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func TestAccServiceAccountIamPolicy(t *testing.T) {
	var serviceAccount iam.ServiceAccount
	cloudID := getExampleCloudID()
	serviceAccountName := acctest.RandomWithPrefix("tf-test")
	userID := getExampleUserID2()
	role := "editor"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountIamPolicy_basic(cloudID, serviceAccountName, userID, role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountExistsWithID("yandex_iam_service_account.test_account", &serviceAccount),
					testAccCheckServiceAccountIam("yandex_iam_service_account.test_account", role, []string{"userAccount:" + userID}),
				),
			},
			{
				ResourceName: "yandex_iam_service_account_iam_policy.foo",
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return serviceAccount.Id, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//revive:disable:var-naming
func testAccServiceAccountIamPolicy_basic(cloudID, accountName, userID, role string) string {
	prerequisiteMembership, deps := testAccCloudAssignCloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_iam_service_account" "test_account" {
  name        = "%s"
  description = "Iam Testing Account"
}

data "yandex_iam_policy" "foo" {
  binding {
    role    = "%s"
    members = ["userAccount:%s"]
  }
}

resource "yandex_iam_service_account_iam_policy" "foo" {
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  policy_data        = "${data.yandex_iam_policy.foo.policy_data}"

  depends_on = [%s]
}
`, accountName, role, userID, deps)
}
