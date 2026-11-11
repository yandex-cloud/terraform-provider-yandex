package acceptance

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
	certificatemanagersdk "github.com/yandex-cloud/go-sdk/services/certificatemanager/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_certificate_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/certificatemanager/pluginsdk"
)

func TestAccCMCertificateIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var certificate certificatemanager.Certificate
	certificateName := acctest.RandomWithPrefix("tf-cm-certificate")
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
				Config: testAccCMCertificateIamBindingBasic(certificateName, certificateContent, privateKey, "", ""),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					suite.testAccCheckCMCertificateEmptyIam("yandex_cm_certificate.test"),
				),
			},
			{
				Config: testAccCMCertificateIamBindingBasic(certificateName, certificateContent, privateKey, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					suite.testAccCheckCMCertificateIam("yandex_cm_certificate.test", role, []string{userID}),
				),
			},
			{
				ResourceName: "yandex_cm_certificate_iam_binding.test-binding",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return certificate.Id + " " + role, nil
				},
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "certificate_id",
			},
			{
				Config: testAccCMCertificateIamBindingBasic(certificateName, certificateContent, privateKey, "", ""),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					suite.testAccCheckCMCertificateEmptyIam("yandex_cm_certificate.test"),
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

func (suite testSuite) testAccCheckCMCertificateEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getCMCertificateResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("binding found but expected empty for %s", resourceName)
	}
}

func (suite testSuite) testAccCheckCMCertificateIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getCMCertificateResourceAccessBindings(s, resourceName)
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

func (suite testSuite) getCMCertificateResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config, configErr := suite.Config()
	if configErr != nil {
		return nil, configErr
	}

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return svc_certificate_manager.GetCMCertificateAccessBindings(config.Context(), config, rs.Primary.ID)
}

func (suite testSuite) testAccCheckCMCertificateExists(name string, certificate *certificatemanager.Certificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := certificatemanagersdk.NewCertificateClient(config.SDK)

		found, err := client.Get(context.Background(), &certificatemanager.GetCertificateRequest{
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
