package acceptance

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	kms "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccKMSAsymmetricEncryptionKeyIamMember_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var kmsKey kms.AsymmetricEncryptionKey
	kmsKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-encryption-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricEncryptionKeyIamMemberBasic(kmsKeyName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricEncryptionKeyExists("yandex_kms_asymmetric_encryption_key.test", &kmsKey),
					suite.testAccCheckKMSAsymmetricEncryptionKeyEmptyIam("yandex_kms_asymmetric_encryption_key.test"),
				),
			},
			{
				Config: testAccKMSAsymmetricEncryptionKeyIamMemberBasic(kmsKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricEncryptionKeyExists("yandex_kms_asymmetric_encryption_key.test", &kmsKey),
					suite.testAccCheckKMSAsymmetricEncryptionKeyIam("yandex_kms_asymmetric_encryption_key.test", role, []string{userID}),
				),
			},
			{
				ResourceName: "yandex_kms_asymmetric_encryption_key_iam_member.test-member",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return kmsKey.Id + "," + role + "," + userID, nil
				},
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "asymmetric_encryption_key_id",
			},
			{
				Config: testAccKMSAsymmetricEncryptionKeyIamMemberBasic(kmsKeyName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricEncryptionKeyExists("yandex_kms_asymmetric_encryption_key.test", &kmsKey),
					suite.testAccCheckKMSAsymmetricEncryptionKeyEmptyIam("yandex_kms_asymmetric_encryption_key.test"),
				),
			},
		},
	})
}

func testAccKMSAsymmetricEncryptionKeyIamMemberBasic(keyName, role, member string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`
resource "yandex_kms_asymmetric_encryption_key" "test" {
  name = "%s"
}
		`, keyName))

	if role != "" && member != "" {
		builder.WriteString(fmt.Sprintf(`
resource "yandex_kms_asymmetric_encryption_key_iam_member" "test-member" {
  asymmetric_encryption_key_id = yandex_kms_asymmetric_encryption_key.test.id
  role   = "%s"
  member = "%s"
}
		`, role, member))
	}
	return builder.String()
}
