package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceCMCertificateContent_selfManaged(t *testing.T) {
	certName := "crt" + acctest.RandString(10) + "-self-managed"
	certDesc := "Terraform Test Self Managed Certificate"
	dataName := "data.yandex_cm_certificate_content.self_managed_certificate"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexCMCertificateAllDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccCMCertificateContentSelfManagedResourceAndData(certName, certDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceCheckYandexCMCertificateResourceExists(dataName),
					resource.TestCheckResourceAttr(dataName, "certificates.#", "1"),
					resource.TestCheckResourceAttr(dataName, "certificates.0", CMCertificateTestSelfSignedCertificate),
					resource.TestCheckResourceAttr(dataName, "private_key", CMCertificateTestPrivateKey),
				),
			},
		},
	})
}

func testAccCMCertificateContentSelfManagedResourceAndData(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_cm_certificate" "self_managed_certificate" {
 name        = "%v"
 description = "%v"
 labels      = {
   key1 = "value1"
   key2 = "value2"
 }
 deletion_protection = false
 self_managed {
   certificate = <<EOF
%vEOF
   private_key = <<EOF
%vEOF
 }
}

data "yandex_cm_certificate_content" "self_managed_certificate" {
	certificate_id = yandex_cm_certificate.self_managed_certificate.id
}
`,
		name,
		desc,
		CMCertificateTestSelfSignedCertificate,
		CMCertificateTestPrivateKey,
	)
}
