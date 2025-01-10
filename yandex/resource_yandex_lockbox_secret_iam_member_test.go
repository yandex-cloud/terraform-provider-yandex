package yandex

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
)

func TestAccLockboxSecretIamMember_basic(t *testing.T) {
	var secret lockbox.Secret
	secretName := acctest.RandomWithPrefix("tf-lockbox-secret")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccLockboxSecretIamMemberBasic(secretName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLockboxSecretExists("yandex_lockbox_secret.test", &secret),
					testAccCheckLockboxSecretEmptyIam("yandex_lockbox_secret.test"),
				),
			},
			{
				Config: testAccLockboxSecretIamMemberBasic(secretName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLockboxSecretExists("yandex_lockbox_secret.test", &secret),
					testAccCheckLockboxSecretIam("yandex_lockbox_secret.test", role, []string{userID}),
				),
			},
			{
				ResourceName: "yandex_lockbox_secret_iam_member.test-member",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return secret.Id + " " + role + " " + userID, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLockboxSecretIamMemberBasic(secretName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLockboxSecretExists("yandex_lockbox_secret.test", &secret),
					testAccCheckLockboxSecretEmptyIam("yandex_lockbox_secret.test"),
				),
			},
		},
	})
}

func testAccLockboxSecretIamMemberBasic(name, role, member string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`
resource "yandex_lockbox_secret" "test" {
  name = "%s"
}
		`, name))

	if role != "" && member != "" {
		builder.WriteString(fmt.Sprintf(`
resource "yandex_lockbox_secret_iam_member" "test-member" {
  secret_id = yandex_lockbox_secret.test.id
  role   = "%s"
  member = "%s"
}
		`, role, member))
	}
	return builder.String()
}
