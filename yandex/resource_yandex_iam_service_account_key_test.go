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

// Test that a service account key can be created and destroyed
func TestAccServiceAccountKey_basic(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountKeyConfig(accountName, accountDesc, "description for test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttrSet(resourceName, "key_algorithm"),
					resource.TestCheckResourceAttrSet(resourceName, "public_key"),
					resource.TestCheckResourceAttrSet(resourceName, "private_key"),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
			{
				Config: testAccServiceAccountKeyConfig(accountName, accountDesc, "updated description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
				),
			},
		},
	})
}

func TestAccServiceAccountKey_encrypted(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	publicKey := pgpkeys.TestPubKey1
	fingerprints, _ := pgpkeys.GetFingerprints([]string{publicKey}, nil)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountKeyConfigEncrypted(accountName, accountDesc, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttr(resourceName, "key_fingerprint", fingerprints[0]),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_private_key"),
					resource.TestCheckNoResourceAttr(resourceName, "private_key"),
					testDecryptKeyAndTest(resourceName, "encrypted_private_key", pgpkeys.TestPrivKey1),
				),
			},
		},
	})
}

func TestAccServiceAccountKey_output_to_lockbox_on_create(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				// output_to_lockbox is defined, so sensitive fields are stored in Lockbox
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "privateKeyIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountKeyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "public_key"),
					resource.TestCheckResourceAttr(resourceName, "private_key", ""), // value is not set in the state
					resource.TestCheckResourceAttrSet(resourceName, "output_to_lockbox_version_id"),
				),
			},
			{
				// output_to_lockbox is removed, so private_key value is taken from the Lockbox secret to the state
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "private_key"), // value recovered from lockbox
				),
			},
		},
	})
}

func TestAccServiceAccountKey_output_to_lockbox_on_update(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	originalPrivateKey := ""
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				// initially, output_to_lockbox is not defined
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountKeyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "key_algorithm"),
					resource.TestCheckResourceAttrSet(resourceName, "public_key"),
					resource.TestCheckResourceAttrSet(resourceName, "private_key"),
					func(s *terraform.State) error {
						// get private_key, to compare later
						privateKey, err := getAttributeFromPrimaryInstanceState(s, resourceName, "private_key")
						originalPrivateKey = privateKey
						return err
					},
				),
			},
			{
				// output_to_lockbox is added, so private_key value is moved from the state (which is cleared) to Lockbox
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "privateKeyIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "private_key", ""), // value is cleared
					resource.TestCheckResourceAttrSet(resourceName, "output_to_lockbox_version_id"),
				),
			},
			{
				// output_to_lockbox is removed, so private_key value is restored from the Lockbox secret to the state
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "private_key"),
					func(s *terraform.State) error {
						// check that the value is restored correctly
						restoredPrivateKey, err := getAttributeFromPrimaryInstanceState(s, resourceName, "private_key")
						if err != nil {
							return err
						}
						if restoredPrivateKey != originalPrivateKey {
							return fmt.Errorf("restored private_key is different from the original private_key")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccServiceAccountKey_output_to_lockbox_secret_cannot_be_updated(t *testing.T) {
	t.Parallel()

	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "privateKeyIsHere",
				)),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"\"dummySecretId\"", "private_key", "privateKeyIsHere",
				)),
				ExpectError: regexp.MustCompile("changing secret_id is not allowed"),
			},
		},
	})
}

func TestAccServiceAccountKey_output_to_lockbox_entries_cannot_be_updated(t *testing.T) {
	t.Parallel()

	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "privateKeyIsHere",
				)),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "nowHere",
				)),
				ExpectError: regexp.MustCompile("changing entry keys is not allowed"),
			},
		},
	})
}

func testAccCheckServiceAccountKeyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_service_account_key" {
			continue
		}

		_, err := config.sdk.IAM().Key().Get(context.Background(), &iam.GetKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("ServiceAccountKey still exists")
		}
	}

	return nil
}

func testAccCheckServiceAccountKeyExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		config := testAccProvider.Meta().(*Config)

		_, err := config.sdk.IAM().Key().Get(context.Background(), &iam.GetKeyRequest{
			KeyId: rs.Primary.ID,
		})

		return err
	}
}

func testAccServiceAccountKeyConfig(name, desc, keyDesc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "%s"
}
`, name, desc, keyDesc)
}

func testAccServiceAccountKeyConfigEncrypted(name, desc, key string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"
  pgp_key            = <<EOF
%s
EOF
}
`, name, desc, key)
}

func testAccServiceAccountKeyConfigOutputToLockbox(name, desc, outputBlock string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "target_secret" {
  name = "%s"
}

resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"

  %s
}
`, name, name, desc, outputBlock)
}
