package yandex

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
)

func TestAccCMCertificateIamBinding_basic(t *testing.T) {
	var certificate certificatemanager.Certificate
	certificateName := acctest.RandomWithPrefix("tf-cm-certificate")
	certificateContent, privateKey, err := getSelfSignedCertificate()
	if err != nil {
		t.Fatal(err)
	}

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCMCertificateIamBindingBasic(certificateName, certificateContent, privateKey, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					testAccCheckCMCertificateEmptyIam("yandex_cm_certificate.test"),
				),
			},
			{
				Config: testAccCMCertificateIamBindingBasic(certificateName, certificateContent, privateKey, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					testAccCheckCMCertificateIam("yandex_cm_certificate.test", role, []string{userID}),
				),
			},
			{
				ResourceName: "yandex_cm_certificate_iam_binding.test-binding",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return certificate.Id + " " + role, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCMCertificateIamBindingBasic(certificateName, certificateContent, privateKey, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					testAccCheckCMCertificateEmptyIam("yandex_cm_certificate.test"),
				),
			},
		},
	})
}

func testAccCMCertificateIamBindingBasic(name, certificateContent, privateKey, role, member string) string {
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
resource "yandex_cm_certificate_iam_binding" "test-binding" {
  certificate_id = yandex_cm_certificate.test.id
  role   = "%s"
  members = ["%s"]
}
		`, role, member))
	}

	return builder.String()
}

func testAccCheckCMCertificateEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getCMCertificateResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckCMCertificateIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getCMCertificateResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
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

		return fmt.Errorf("binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func getCMCertificateResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := testAccProvider.Meta().(*Config)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getCMCertificateAccessBindings(config.Context(), config, rs.Primary.ID)
}

func testAccCheckCMCertificateExists(name string, certificate *certificatemanager.Certificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Certificates().Certificate().Get(context.Background(), &certificatemanager.GetCertificateRequest{
			CertificateId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("certificate not found")
		}

		*certificate = *found

		return nil
	}
}

func getSelfSignedCertificate() (string, string, error) {
	return CMCertificateTestSelfSignedCertificate, CMCertificateTestPrivateKey, nil
}
