package acceptance

import (
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	rmfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccServiceAccountIamPolicy(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var serviceAccount iam.ServiceAccount
	cloudID := acceptanceenv.CloudID()
	serviceAccountName := acctest.RandomWithPrefix("tf-test")
	userID := acceptanceenv.UserID2()
	role := "editor"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountIamPolicy_basic(cloudID, serviceAccountName, userID, role),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexIAMServiceAccountExistsWithID("yandex_iam_service_account.test_account", &serviceAccount),
					suite.testAccCheckServiceAccountIam("yandex_iam_service_account.test_account", role, []string{"userAccount:" + userID}),
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
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID)
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
