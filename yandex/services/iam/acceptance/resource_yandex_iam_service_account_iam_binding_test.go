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

func importIDFunc(serviceAccount *iam.ServiceAccount, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return serviceAccount.Id + "," + role, nil
	}
}

func TestAccServiceAccountIamBinding(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var serviceAccount iam.ServiceAccount
	serviceAccountName := acctest.RandomWithPrefix("tf-test")
	cloudID := acceptanceenv.CloudID()
	userID := acceptanceenv.UserID1()
	role := "viewer"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountIamBinding_basic(cloudID, serviceAccountName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexIAMServiceAccountExistsWithID("yandex_iam_service_account.test_account", &serviceAccount),
					suite.testAccCheckServiceAccountIam("yandex_iam_service_account.test_account", role, []string{"userAccount:" + userID}),
				),
			},
			{
				ResourceName:                         "yandex_iam_service_account_iam_binding.foo",
				ImportStateIdFunc:                    importIDFunc(&serviceAccount, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "service_account_id",
			},
		},
	})
}

//revive:disable:var-naming
func testAccServiceAccountIamBinding_basic(cloudID, accountName, role, userID string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID)
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
