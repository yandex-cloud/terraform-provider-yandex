package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hashicorp/vault/helper/pgpkeys"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

// Test that a service account API key can be created and destroyed
func TestAccServiceAccountAPIKey_basic(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_api_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountAPIKeyConfig(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountAPIKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
		},
	})
}

func TestAccServiceAccountAPIKey_encrypted(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_api_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	publicKey := pgpkeys.TestPubKey1
	fingerprints, _ := pgpkeys.GetFingerprints([]string{publicKey}, nil)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountAPIKeyConfigEncrypted(accountName, accountDesc, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountAPIKeyExists(resourceName),
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

func testAccCheckServiceAccountAPIKeyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_service_account_api_key" {
			continue
		}

		_, err := config.sdk.IAM().ApiKey().Get(context.Background(), &iam.GetApiKeyRequest{
			ApiKeyId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("ServiceAccountAPIKey still exists")
		}
	}

	return nil
}

func testAccCheckServiceAccountAPIKeyExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		config := testAccProvider.Meta().(*Config)

		_, err := config.sdk.IAM().ApiKey().Get(context.Background(), &iam.GetApiKeyRequest{
			ApiKeyId: rs.Primary.ID,
		})

		return err
	}
}

func testAccServiceAccountAPIKeyConfig(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_api_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"
}
`, name, desc)
}

func testAccServiceAccountAPIKeyConfigEncrypted(name, desc, key string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_api_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"
  pgp_key            = <<EOF
%s
EOF
}
`, name, desc, key)
}
