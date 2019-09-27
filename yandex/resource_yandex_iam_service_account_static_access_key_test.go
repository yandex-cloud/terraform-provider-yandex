package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hashicorp/vault/helper/pgpkeys"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
)

// Test that a service account key can be created and destroyed
func TestAccServiceAccountStaticAccessKey_basic(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_static_access_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountStaticAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountStaticAccessKeyConfig(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountStaticAccessKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttrSet(resourceName, "access_key"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
		},
	})
}

func TestAccServiceAccountStaticAccessKey_encrypted(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_static_access_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	publicKey := pgpkeys.TestPubKey1
	fingerprints, _ := pgpkeys.GetFingerprints([]string{publicKey}, nil)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountStaticAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountStaticAccessKeyConfigEncrypted(accountName, accountDesc, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountStaticAccessKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttr(resourceName, "key_fingerprint", fingerprints[0]),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_secret_key"),
					resource.TestCheckNoResourceAttr(resourceName, "secret_key"),
					testDecryptKeyAndTest(resourceName, "encrypted_secret_key", pgpkeys.TestPrivKey1),
				),
			},
		},
	})
}

func testAccCheckServiceAccountStaticAccessKeyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_service_account_static_access_key" {
			continue
		}

		_, err := config.sdk.IAM().AWSCompatibility().AccessKey().Get(context.Background(), &awscompatibility.GetAccessKeyRequest{
			AccessKeyId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("ServiceAccountStaticAccessKey still exists")
		}
	}

	return nil
}

func testAccCheckServiceAccountStaticAccessKeyExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		config := testAccProvider.Meta().(*Config)

		_, err := config.sdk.IAM().AWSCompatibility().AccessKey().Get(context.Background(), &awscompatibility.GetAccessKeyRequest{
			AccessKeyId: rs.Primary.ID,
		})

		return err
	}
}

func testAccServiceAccountStaticAccessKeyConfig(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_static_access_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"
}
`, name, desc)
}

func testAccServiceAccountStaticAccessKeyConfigEncrypted(name, desc, key string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_static_access_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"
  pgp_key            = <<EOF
%s
EOF
}
`, name, desc, key)
}
