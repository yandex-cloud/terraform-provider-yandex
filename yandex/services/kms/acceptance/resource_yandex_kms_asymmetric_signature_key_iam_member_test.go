package acceptance

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	kms "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccKMSAsymmetricSignatureKeyIamMember_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var kmsKey kms.AsymmetricSignatureKey
	kmsKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-signature-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricSignatureKeyIamMemberBasic(kmsKeyName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.test", &kmsKey),
					suite.testAccCheckKMSAsymmetricSignatureKeyEmptyIam("yandex_kms_asymmetric_signature_key.test"),
				),
			},
			{
				Config: testAccKMSAsymmetricSignatureKeyIamMemberBasic(kmsKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.test", &kmsKey),
					suite.testAccCheckKMSAsymmetricSignatureKeyIam("yandex_kms_asymmetric_signature_key.test", role, []string{userID}),
				),
			},
			{
				ResourceName: "yandex_kms_asymmetric_signature_key_iam_member.test-member",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return kmsKey.Id + "," + role + "," + userID, nil
				},
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "asymmetric_signature_key_id",
			},
			{
				Config: testAccKMSAsymmetricSignatureKeyIamMemberBasic(kmsKeyName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.test", &kmsKey),
					suite.testAccCheckKMSAsymmetricSignatureKeyEmptyIam("yandex_kms_asymmetric_signature_key.test"),
				),
			},
		},
	})
}

func testAccKMSAsymmetricSignatureKeyIamMemberBasic(keyName, role, member string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`
resource "yandex_kms_asymmetric_signature_key" "test" {
  name = "%s"
}
		`, keyName))

	if role != "" && member != "" {
		builder.WriteString(fmt.Sprintf(`
resource "yandex_kms_asymmetric_signature_key_iam_member" "test-member" {
  asymmetric_signature_key_id = yandex_kms_asymmetric_signature_key.test.id
  role   = "%s"
  member = "%s"
}
		`, role, member))
	}
	return builder.String()
}
