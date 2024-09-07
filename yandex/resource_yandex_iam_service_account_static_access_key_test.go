package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func TestAccServiceAccountStaticAccessKey_output_to_lockbox_on_create(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_static_access_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	lockboxVersionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountStaticAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				// output_to_lockbox is defined, so sensitive fields are stored in Lockbox
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockboxAsMap(
					"yandex_lockbox_secret.target_secret.id", map[string]string{"access_key": "accessKeyIsHere", "secret_key": "secretKeyIsHere"},
				)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountStaticAccessKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_key", ""),
					resource.TestCheckResourceAttr(resourceName, "secret_key", ""),
					resource.TestCheckResourceAttrSet(resourceName, lockboxOutputVersionIdAttr),
					func(s *terraform.State) error {
						// get Lockbox version ID, to check late
						versionID, err := getResourceAttrValue(s, resourceName, lockboxOutputVersionIdAttr)
						lockboxVersionID = versionID
						return err
					},
				),
			},
			{
				// output_to_lockbox is removed, so values are restored from the Lockbox secret to the state, and Lockbox version is deleted
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "access_key"), // value recovered from lockbox
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"), // value recovered from lockbox
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID)
					},
				),
			},
		},
	})
}

func TestAccServiceAccountStaticAccessKey_output_to_lockbox_on_destroy(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_static_access_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	lockboxVersionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountStaticAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				// output_to_lockbox is defined, so sensitive fields are stored in Lockbox
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockboxAsMap(
					"yandex_lockbox_secret.target_secret.id", map[string]string{"access_key": "accessKeyIsHere", "secret_key": "secretKeyIsHere"},
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get Lockbox version ID, to check late
						versionID, err := getResourceAttrValue(s, resourceName, lockboxOutputVersionIdAttr)
						lockboxVersionID = versionID
						return err
					},
				),
			},
			{
				// IAM key is removed, so Lockbox version is destroyed
				Config: testAccServiceAccountKeyConfigJustSecret(accountName),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID)
					},
				),
			},
		},
	})
}

func TestAccServiceAccountStaticAccessKey_output_to_lockbox_on_update(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_static_access_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	originalAccessKey := ""
	originalSecretKey := ""
	lockboxVersionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountStaticAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				// initially, output_to_lockbox is not defined
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountStaticAccessKeyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_key"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					func(s *terraform.State) error {
						// get access_key and secret_key, to compare later
						accessKey, err := getResourceAttrValue(s, resourceName, "access_key")
						if err != nil {
							return err
						}
						originalAccessKey = accessKey
						secretKey, err := getResourceAttrValue(s, resourceName, "secret_key")
						originalSecretKey = secretKey
						return err
					},
				),
			},
			{
				// output_to_lockbox is added, so access_key and secret_key values are moved from the state to Lockbox
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockboxAsMap(
					"yandex_lockbox_secret.target_secret.id", map[string]string{"access_key": "accessKeyIsHere", "secret_key": "secretKeyIsHere"},
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "access_key", ""), // value is cleared
					resource.TestCheckResourceAttr(resourceName, "secret_key", ""), // value is cleared
					resource.TestCheckResourceAttrSet(resourceName, lockboxOutputVersionIdAttr),
					func(s *terraform.State) error {
						// get Lockbox version ID, to check later
						versionID, err := getResourceAttrValue(s, resourceName, lockboxOutputVersionIdAttr)
						lockboxVersionID = versionID
						return err
					},
				),
			},
			{
				// output_to_lockbox is removed, so values are restored from the Lockbox secret to the state, and Lockbox version is deleted
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrWithValueFactory(resourceName, "access_key", func() string {
						return originalAccessKey
					}),
					testAccCheckResourceAttrWithValueFactory(resourceName, "secret_key", func() string {
						return originalSecretKey
					}),
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID)
					},
				),
			},
		},
	})
}

func TestAccServiceAccountStaticAccessKey_output_to_lockbox_updated_secret(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_static_access_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	originalAccessKey := ""
	originalSecretKey := ""
	lockboxVersionID1 := ""
	lockboxVersionID2 := ""
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get access_key and secret_key, to compare later
						accessKey, err := getResourceAttrValue(s, resourceName, "access_key")
						if err != nil {
							return err
						}
						originalAccessKey = accessKey
						secretKey, err := getResourceAttrValue(s, resourceName, "secret_key")
						originalSecretKey = secretKey
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockboxAsMap(
					"yandex_lockbox_secret.target_secret.id", map[string]string{"access_key": "accessKeyIsHere", "secret_key": "secretKeyIsHere"},
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get Lockbox version ID, to check later
						versionID, err := getResourceAttrValue(s, resourceName, lockboxOutputVersionIdAttr)
						lockboxVersionID1 = versionID
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockboxAsMap(
					"yandex_lockbox_secret.target_secret_2.id", map[string]string{"access_key": "accessKeyIsHere", "secret_key": "secretKeyIsHere"}, // changed secret
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID1)
					},
					func(s *terraform.State) error {
						// get the new Lockbox version ID, to check later
						versionID, err := getResourceAttrValue(s, resourceName, lockboxOutputVersionIdAttr)
						lockboxVersionID2 = versionID
						if lockboxVersionID1 == lockboxVersionID2 {
							return fmt.Errorf("a new version should have been created, but got the same version %s", lockboxVersionID1)
						}
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrWithValueFactory(resourceName, "access_key", func() string {
						return originalAccessKey
					}),
					testAccCheckResourceAttrWithValueFactory(resourceName, "secret_key", func() string {
						return originalSecretKey
					}),
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret_2", lockboxVersionID2)
					},
				),
			},
		},
	})
}

// This test is almost equal to the previous one, but here we change entries
func TestAccServiceAccountStaticAccessKey_output_to_lockbox_updated_entries(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_service_account_static_access_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	originalAccessKey := ""
	originalSecretKey := ""
	lockboxVersionID1 := ""
	lockboxVersionID2 := ""
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get access_key and secret_key, to compare later
						accessKey, err := getResourceAttrValue(s, resourceName, "access_key")
						if err != nil {
							return err
						}
						originalAccessKey = accessKey
						secretKey, err := getResourceAttrValue(s, resourceName, "secret_key")
						originalSecretKey = secretKey
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockboxAsMap(
					"yandex_lockbox_secret.target_secret.id", map[string]string{"access_key": "accessKeyIsHere", "secret_key": "secretKeyIsHere"},
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get Lockbox version ID, to check later
						versionID, err := getResourceAttrValue(s, resourceName, lockboxOutputVersionIdAttr)
						lockboxVersionID1 = versionID
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, testAccOutputToLockboxAsMap(
					"yandex_lockbox_secret.target_secret.id", map[string]string{"access_key": "nowIsHere", "secret_key": "secretKeyIsHere"}, // changed entry key
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID1)
					},
					func(s *terraform.State) error {
						// get the new Lockbox version ID, to check later
						versionID, err := getResourceAttrValue(s, resourceName, lockboxOutputVersionIdAttr)
						lockboxVersionID2 = versionID
						if lockboxVersionID1 == lockboxVersionID2 {
							return fmt.Errorf("a new version should have been created, but got the same version %s", lockboxVersionID1)
						}
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrWithValueFactory(resourceName, "access_key", func() string {
						return originalAccessKey
					}),
					testAccCheckResourceAttrWithValueFactory(resourceName, "secret_key", func() string {
						return originalSecretKey
					}),
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID2)
					},
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

func testAccServiceAccountStaticAccessKeyConfigOutputToLockbox(name, desc, outputBlock string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "target_secret" {
  name = "%s"
}

resource "yandex_lockbox_secret" "target_secret_2" {
  name = "%s"
}

resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "%s"
}

resource "yandex_iam_service_account_static_access_key" "acceptance" {
  service_account_id = "${yandex_iam_service_account.acceptance.id}"
  description        = "description for test"

  %s
}
`, name, name, name, desc, outputBlock)
}
