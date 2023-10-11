package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
)

func TestAccDataSourceCMCertificate_selfManaged(t *testing.T) {
	certName := "crt" + acctest.RandString(10) + "-self-managed"
	certDesc := "Terraform Test Self Managed Certificate"
	folderID := getExampleFolderID()
	dataName := "data.yandex_cm_certificate.self_managed_certificate"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexCMCertificateAllDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccCMCertificateSelfManagedResourceAndData(certName, certDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceCheckYandexCMCertificateResourceExists(dataName),
					resource.TestCheckResourceAttr(dataName, "folder_id", folderID),
					resource.TestCheckResourceAttr(dataName, "name", certName),
					resource.TestCheckResourceAttr(dataName, "description", certDesc),
					resource.TestCheckResourceAttr(dataName, "labels.%", "2"),
					resource.TestCheckResourceAttr(dataName, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(dataName, "labels.key2", "value2"),
					resource.TestCheckResourceAttr(dataName, "domains.#", "1"),
					resource.TestCheckResourceAttr(dataName, "domains.0", "example.com"),
					resource.TestCheckResourceAttr(dataName, "serial", "9d1134c1a824ad86"),
					resource.TestCheckResourceAttr(dataName, "challenges.#", "0"),
					resource.TestCheckResourceAttr(dataName, "not_after", "7499-02-13T09:48:13Z"),
					resource.TestCheckResourceAttr(dataName, "not_before", "2023-04-23T09:48:13Z"),
					resource.TestCheckResourceAttr(dataName, "status",
						certificatemanager.Certificate_Status_name[int32(certificatemanager.Certificate_ISSUED)]),
					resource.TestCheckResourceAttr(dataName, "type",
						certificatemanager.CertificateType_name[int32(certificatemanager.CertificateType_IMPORTED)]),
					testAccCheckCreatedAtAttr(dataName),
				),
			},
		},
	})
}

func testAccCMCertificateSelfManagedResourceAndData(name, desc string) string {
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

data "yandex_cm_certificate" "self_managed_certificate" {
	certificate_id = yandex_cm_certificate.self_managed_certificate.id
}
`,
		name,
		desc,
		CMCertificateTestSelfSignedCertificate,
		CMCertificateTestPrivateKey,
	)
}

func testAccDataSourceCheckYandexCMCertificateResourceExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for: %s", r)
		}
		return nil
	}
}
