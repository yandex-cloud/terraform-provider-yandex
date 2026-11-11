package acceptance

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccCMCertificateIamMember_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var certificate certificatemanager.Certificate
	name := acctest.RandomWithPrefix("tf-cm-certificate")
	certificateContent, privateKey, err := getSelfSignedCertificate()
	if err != nil {
		t.Fatal(err)
	}

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccCMCertificateIamMemberBasic(name, certificateContent, privateKey, "", ""),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					suite.testAccCheckCMCertificateEmptyIam("yandex_cm_certificate.test"),
				),
			},
			{
				Config: testAccCMCertificateIamMemberBasic(name, certificateContent, privateKey, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					suite.testAccCheckCMCertificateIam("yandex_cm_certificate.test", role, []string{userID}),
				),
			},
			{
				ResourceName: "yandex_cm_certificate_iam_member.test-member",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return certificate.Id + "," + role + "," + userID, nil
				},
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "certificate_id",
			},
			{
				Config: testAccCMCertificateIamMemberBasic(name, certificateContent, privateKey, "", ""),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					suite.testAccCheckCMCertificateEmptyIam("yandex_cm_certificate.test"),
				),
			},
		},
	})
}

func testAccCMCertificateIamMemberBasic(name, certificateContent, privateKey, role, member string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`
resource "yandex_cm_certificate" "test" {
  name = "%s"
  self_managed {
	certificate = <<EOF
%s
EOF
	private_key = <<EOF
%s
EOF
  }
}
		`, name, certificateContent, privateKey))

	if role != "" && member != "" {
		builder.WriteString(fmt.Sprintf(`
resource "yandex_cm_certificate_iam_member" "test-member" {
  certificate_id = yandex_cm_certificate.test.id
  role   = "%s"
  member = "%s"
}
		`, role, member))
	}
	return builder.String()
}
