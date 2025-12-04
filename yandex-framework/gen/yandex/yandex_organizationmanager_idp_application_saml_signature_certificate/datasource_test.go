package yandex_organizationmanager_idp_application_saml_signature_certificate_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceOrganizationManagerIdpApplicationSamlSignatureCertificate_byID(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc-test-saml-app")
	certName := acctest.RandomWithPrefix(signatureCertificateNamePrefix)
	organizationID := test.GetExampleOrganizationID()
	dataSourceName := "data.yandex_organizationmanager_idp_application_saml_signature_certificate.source"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpSamlSignatureCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIdpSamlSignatureCertificateConfig(appName, certName, "data source certificate", organizationID),
				Check: resource.ComposeTestCheckFunc(
					test.AccCheckResourceIDField(dataSourceName, "signature_certificate_id"),
					resource.TestCheckResourceAttr(dataSourceName, "name", certName),
					resource.TestCheckResourceAttr(dataSourceName, "description", "data source certificate"),
					resource.TestCheckResourceAttr(dataSourceName, "status", "INACTIVE"),
					resource.TestCheckResourceAttrSet(dataSourceName, "signature_certificate_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "fingerprint"),
					resource.TestCheckResourceAttrSet(dataSourceName, "application_id"),
					test.AccCheckCreatedAtAttr(dataSourceName),
				),
			},
		},
	})
}

func testAccDataSourceIdpSamlSignatureCertificateConfig(appName, certificateName, description, organizationID string) string {
	return fmt.Sprintf(`
%s

data "yandex_organizationmanager_idp_application_saml_signature_certificate" "source" {
  signature_certificate_id = yandex_organizationmanager_idp_application_saml_signature_certificate.foobar.signature_certificate_id
}
`, testAccIdpSamlSignatureCertificateConfig(appName, certificateName, description, organizationID))
}
