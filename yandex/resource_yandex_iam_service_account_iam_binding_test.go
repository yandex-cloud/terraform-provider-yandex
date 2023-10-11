package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func importIDFunc(serviceAccount *iam.ServiceAccount, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return serviceAccount.Id + " " + role, nil
	}
}

func TestAccServiceAccountIamBinding(t *testing.T) {
	var serviceAccount iam.ServiceAccount
	serviceAccountName := acctest.RandomWithPrefix("tf-test")
	cloudID := getExampleCloudID()
	userID := getExampleUserID1()
	role := "viewer"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountIamBinding_basic(cloudID, serviceAccountName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountExistsWithID("yandex_iam_service_account.test_account", &serviceAccount),
					testAccCheckServiceAccountIam("yandex_iam_service_account.test_account", role, []string{"userAccount:" + userID}),
				),
			},
			{
				ResourceName:      "yandex_iam_service_account_iam_binding.foo",
				ImportStateIdFunc: importIDFunc(&serviceAccount, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//revive:disable:var-naming
func testAccServiceAccountIamBinding_basic(cloudID, accountName, role, userID string) string {
	prerequisiteMembership, deps := testAccCloudAssignCloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_iam_service_account" "test_account" {
  name        = "%s"
  description = "Iam Testing Account"
}

resource "yandex_iam_service_account_iam_binding" "foo" {
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  role               = "%s"
  members            = ["userAccount:%s"]

  depends_on = [%s]
}
`, accountName, role, userID, deps)
}
