package yandex

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
)

func TestAccKMSAsymmetricSignatureKeyIamMember_basic(t *testing.T) {
	var kmsKey kms.AsymmetricSignatureKey
	kmsKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-signature-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricSignatureKeyIamMemberBasic(kmsKeyName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.test", &kmsKey),
					testAccCheckKMSAsymmetricSignatureKeyEmptyIam("yandex_kms_asymmetric_signature_key.test"),
				),
			},
			{
				Config: testAccKMSAsymmetricSignatureKeyIamMemberBasic(kmsKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.test", &kmsKey),
					testAccCheckKMSAsymmetricSignatureKeyIam("yandex_kms_asymmetric_signature_key.test", role, []string{userID}),
				),
			},
			{
				ResourceName: "yandex_kms_asymmetric_signature_key_iam_member.test-member",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return kmsKey.Id + " " + role + " " + userID, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKMSAsymmetricSignatureKeyIamMemberBasic(kmsKeyName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.test", &kmsKey),
					testAccCheckKMSAsymmetricSignatureKeyEmptyIam("yandex_kms_asymmetric_signature_key.test"),
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
