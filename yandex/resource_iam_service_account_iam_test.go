package yandex

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
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

func TestAccServiceAccountIamPolicy(t *testing.T) {
	var serviceAccount iam.ServiceAccount
	cloudID := getExampleCloudID()
	serviceAccountName := acctest.RandomWithPrefix("tf-test")
	userID := getExampleUserID2()
	role := "resource-manager.clouds.member"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountIamPolicy_basic(cloudID, serviceAccountName, userID),
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

func testAccCheckServiceAccountIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		resp, err := config.sdk.IAM().ServiceAccount().ListAccessBindings(context.Background(), &access.ListAccessBindingsRequest{
			ResourceId: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range resp.AccessBindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
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

func testAccServiceAccountIamMember_basic(cloudID, accountName, role, userID string) string {
	prerequisiteMembership, deps := testAccCloudAssignCloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_iam_service_account" "test_account" {
  name        = "%s"
  description = "Iam Testing Account"
}

resource "yandex_iam_service_account_iam_member" "foo" {
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  role   = "%s"
  member = "userAccount:%s"

  depends_on = [%s]
}
`, accountName, role, userID, deps)
}

func testAccServiceAccountIamPolicy_basic(cloudID, accountName, userID string) string {
	prerequisiteMembership, deps := testAccCloudAssignCloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_iam_service_account" "test_account" {
  name        = "%s"
  description = "Iam Testing Account"
}

data "yandex_iam_policy" "foo" {
	binding {
		role = "resource-manager.clouds.member"
		members = ["userAccount:%s"]
	}
}

resource "yandex_iam_service_account_iam_policy" "foo" {
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  policy_data        = "${data.yandex_iam_policy.foo.policy_data}"

  depends_on = [%s]
}
`, accountName, userID, deps)
}
