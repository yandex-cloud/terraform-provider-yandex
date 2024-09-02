package yandex

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func TestAccServiceAccountAPIKey_scoped(t *testing.T) {
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
				Config: testAccServiceAccountAPIKeyConfigScoped(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountAPIKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttr(resourceName, "scope", "yc.ydb.topics.manage"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
		},
	})
}

func TestAccServiceAccountAPIKey_expired(t *testing.T) {
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
				Config: testAccServiceAccountAPIKeyConfigExpired(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountAPIKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttr(resourceName, "expires_at", "2099-11-11T22:33:44Z"),
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

func TestAccServiceAccountAPIKey_output_to_lockbox_on_create(t *testing.T) {
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
				// output_to_lockbox is defined, so sensitive fields are stored in Lockbox
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_key", "secretKeyIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountAPIKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "secret_key", ""),
					resource.TestCheckResourceAttrSet(resourceName, "output_to_lockbox_version_id"),
				),
			},
			{
				// output_to_lockbox is removed, so private_key value is taken from the Lockbox secret to the state
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"), // value recovered from lockbox
				),
			},
		},
	})
}

func TestAccServiceAccountAPIKey_output_to_lockbox_on_update(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_api_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	originalSecretKey := ""
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				// initially, output_to_lockbox is not defined
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountAPIKeyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					func(s *terraform.State) error {
						// get secret_key, to compare later
						secretKey, err := getAttributeFromPrimaryInstanceState(s, resourceName, "secret_key")
						originalSecretKey = secretKey
						return err
					},
				),
			},
			{
				// output_to_lockbox is added, so secret_key value is moved from the state (which is cleared) to Lockbox
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_key", "secretKeyIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "secret_key", ""), // value is cleared
					resource.TestCheckResourceAttrSet(resourceName, "output_to_lockbox_version_id"),
				),
			},
			{
				// output_to_lockbox is removed, so secret_key value is restored from the Lockbox secret to the state
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					func(s *terraform.State) error {
						// check that the value is restored correctly
						restoredSecretKey, err := getAttributeFromPrimaryInstanceState(s, resourceName, "secret_key")
						if err != nil {
							return err
						}
						if restoredSecretKey != originalSecretKey {
							return fmt.Errorf("restored secret_key is different from the original secret_key")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccServiceAccountAPIKey_output_to_lockbox_secret_cannot_be_updated(t *testing.T) {
	t.Parallel()

	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, ""),
			},
			{
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_key", "secretKeyIsHere",
				)),
			},
			{
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"\"dummySecretId\"", "secret_key", "secretKeyIsHere",
				)),
				ExpectError: regexp.MustCompile("changing secret_id is not allowed"),
			},
		},
	})
}

func TestAccServiceAccountAPIKey_output_to_lockbox_entries_cannot_be_updated(t *testing.T) {
	t.Parallel()

	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, ""),
			},
			{
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_key", "secretKeyIsHere",
				)),
			},
			{
				Config: testAccServiceAccountAPIKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_key", "nowHere",
				)),
				ExpectError: regexp.MustCompile("changing entry keys is not allowed"),
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

func testAccServiceAccountAPIKeyConfigScoped(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_api_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"
  scope        		 = "yc.ydb.topics.manage"
}
`, name, desc)
}
func testAccServiceAccountAPIKeyConfigExpired(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_api_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"
  expires_at   		 = "2099-11-11T22:33:44Z"
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

func testAccServiceAccountAPIKeyConfigOutputToLockbox(name, desc, outputBlock string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "target_secret" {
  name = "%s"
}

resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_api_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"

  %s
}
`, name, name, desc, outputBlock)
}
