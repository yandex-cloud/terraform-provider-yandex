package yandex

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
)

func TestAccKMSSecretCiphertext_basic(t *testing.T) {
	t.Parallel()

	keyName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	plaintext := acctest.RandString(18)
	aadContext := acctest.RandString(36)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSSecretCiphertext_basic(keyName, plaintext, aadContext),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSecretCiphertextDecryptable(
						"yandex_kms_secret_ciphertext.ciphertext-a"),
					testAccCheckKMSSecretCiphertextDecryptable(
						"yandex_kms_secret_ciphertext.ciphertext-b"),
				),
			},
		},
	})
}

func testAccCheckKMSSecretCiphertextDecryptable(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		plaintext := rs.Primary.Attributes["plaintext"]
		ciphertext, err := base64.StdEncoding.DecodeString(rs.Primary.Attributes["ciphertext"])
		if err != nil {
			return fmt.Errorf("Cannot decode ciphertext from base64: %s", err)
		}

		req := &kms.SymmetricDecryptRequest{
			KeyId:      rs.Primary.Attributes["key_id"],
			AadContext: []byte(rs.Primary.Attributes["aad_context"]),
			Ciphertext: ciphertext,
		}

		resp, err := config.sdk.KMSCrypto().SymmetricCrypto().Decrypt(context.Background(), req)
		if err != nil {
			return fmt.Errorf("Error while requesting API to decrypt data with KMS symmetric key: %s", err)
		}

		if plaintext != string(resp.Plaintext) {
			return fmt.Errorf("Wrong decoded plaintext: expected '%s', got '%s'", plaintext, resp.Plaintext)
		}

		return nil
	}
}

//revive:disable:var-naming
func testAccKMSSecretCiphertext_basic(keyName, aadContext, plaintext string) string {
	return fmt.Sprintf(`
resource "yandex_kms_symmetric_key" "key" {
  name              = "%s"
  description       = "description with update for key-a"
  default_algorithm = "AES_192"
}

resource "yandex_kms_secret_ciphertext" "ciphertext-a" {
  key_id      = "${yandex_kms_symmetric_key.key.id}"
  aad_context = "%s"
  plaintext   = "%s"
}

resource "yandex_kms_secret_ciphertext" "ciphertext-b" {
  key_id    = "${yandex_kms_symmetric_key.key.id}"
  plaintext = "%s"
}
`, keyName, aadContext, plaintext, plaintext)
}
