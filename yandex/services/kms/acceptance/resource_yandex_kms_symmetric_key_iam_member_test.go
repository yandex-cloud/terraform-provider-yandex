package acceptance

import (
	"fmt"
	"strings"
	"testing"

	kmsfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccKMSSymmetricKeyIamMember_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var kmsKey kms.SymmetricKey
	kmsKeyName := acctest.RandomWithPrefix("tf-kms-symmetric-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccKMSSymmetricKeyIamMemberBasic(kmsKeyName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					kmsfixtures.CheckSymmetricKeyExists(suite.Config, "yandex_kms_symmetric_key.test", &kmsKey),
					suite.testAccCheckKMSSymmetricKeyEmptyIam("yandex_kms_symmetric_key.test"),
				),
			},
			{
				Config: testAccKMSSymmetricKeyIamMemberBasic(kmsKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					kmsfixtures.CheckSymmetricKeyExists(suite.Config, "yandex_kms_symmetric_key.test", &kmsKey),
					suite.testAccCheckKMSSymmetricKeyIam("yandex_kms_symmetric_key.test", role, []string{userID}),
				),
			},
			{
				ResourceName: "yandex_kms_symmetric_key_iam_member.test-member",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return kmsKey.Id + "," + role + "," + userID, nil
				},
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "symmetric_key_id",
			},
			{
				Config: testAccKMSSymmetricKeyIamMemberBasic(kmsKeyName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					kmsfixtures.CheckSymmetricKeyExists(suite.Config, "yandex_kms_symmetric_key.test", &kmsKey),
					suite.testAccCheckKMSSymmetricKeyEmptyIam("yandex_kms_symmetric_key.test"),
				),
			},
		},
	})
}

func testAccKMSSymmetricKeyIamMemberBasic(keyName, role, member string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`
resource "yandex_kms_symmetric_key" "test" {
  name = "%s"
}
		`, keyName))

	if role != "" && member != "" {
		builder.WriteString(fmt.Sprintf(`
resource "yandex_kms_symmetric_key_iam_member" "test-member" {
  symmetric_key_id = yandex_kms_symmetric_key.test.id
  role   = "%s"
  member = "%s"
}
		`, role, member))
	}
	return builder.String()
}
