package yandex

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/vault/helper/pgpkeys"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

// Test that an OAuth client secret can be created and destroyed
func TestAccIAMOAuthClientSecret_basic(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_oauth_client_secret.acceptance"
	clientName := "oauth-client-" + acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckOAuthClientSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOAuthClientSecretConfig(clientName, "description for test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOAuthClientSecretExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_value"),
					resource.TestCheckResourceAttrSet(resourceName, "masked_secret"),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
		},
	})
}

func TestAccIAMOAuthClientSecret_encrypted(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_oauth_client_secret.acceptance"
	clientName := "oauth-client-" + acctest.RandString(10)
	publicKey := pgpkeys.TestPubKey1
	fingerprints, _ := pgpkeys.GetFingerprints([]string{publicKey}, nil)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckOAuthClientSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOAuthClientSecretConfigEncrypted(clientName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOAuthClientSecretExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttr(resourceName, "key_fingerprint", fingerprints[0]),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_secret_value"),
					resource.TestCheckNoResourceAttr(resourceName, "secret_value"),
					testDecryptKeyAndTest(resourceName, "encrypted_secret_value", pgpkeys.TestPrivKey1),
				),
			},
		},
	})
}

func TestAccIAMOAuthClientSecret_output_to_lockbox_on_create(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_oauth_client_secret.acceptance"
	clientName := "oauth-client-" + acctest.RandString(10)
	lockboxVersionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckOAuthClientSecretDestroy,
		Steps: []resource.TestStep{
			{
				// output_to_lockbox is defined, so sensitive fields are stored in Lockbox
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_value", "secretValueIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOAuthClientSecretExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "masked_secret"),
					resource.TestCheckResourceAttr(resourceName, "secret_value", ""), // value is not set in the state
					resource.TestCheckResourceAttrSet(resourceName, lockboxOutputVersionIdAttr),
					func(s *terraform.State) error {
						// get version ID, to check later
						versionID, err := getResourceAttrValue(s, resourceName, lockboxOutputVersionIdAttr)
						lockboxVersionID = versionID
						return err
					},
				),
			},
			{
				// output_to_lockbox is removed, so secret_value value is restored from the Lockbox secret to the state, and Lockbox version is deleted
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "secret_value"), // value recovered from lockbox
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID)
					},
				),
			},
		},
	})
}

func TestAccIAMOAuthClientSecret_output_to_lockbox_on_destroy(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_oauth_client_secret.acceptance"
	clientName := "oauth-client-" + acctest.RandString(10)
	lockboxVersionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckOAuthClientSecretDestroy,
		Steps: []resource.TestStep{
			{
				// output_to_lockbox is defined, so sensitive fields are stored in Lockbox
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_value", "secretValueIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get Lockbox version ID, to check later
						versionID, err := getResourceAttrValue(s, resourceName, lockboxOutputVersionIdAttr)
						lockboxVersionID = versionID
						return err
					},
				),
			},
			{
				// OAuth client secret is removed, so Lockbox version is destroyed
				Config: testAccOAuthClientSecretConfigJustSecret(clientName),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID)
					},
				),
			},
		},
	})
}

func TestAccIAMOAuthClientSecret_output_to_lockbox_added_and_removed(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_oauth_client_secret.acceptance"
	clientName := "oauth-client-" + acctest.RandString(10)
	originalSecretValue := ""
	lockboxVersionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckOAuthClientSecretDestroy,
		Steps: []resource.TestStep{
			{
				// initially, output_to_lockbox is not defined
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOAuthClientSecretExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "masked_secret"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_value"),
					func(s *terraform.State) error {
						// get secret_value, to compare later
						secretValue, err := getResourceAttrValue(s, resourceName, "secret_value")
						originalSecretValue = secretValue
						return err
					},
				),
			},
			{
				// output_to_lockbox is added, so secret_value value is moved from the state (which is cleared) to Lockbox
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_value", "secretValueIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "secret_value", ""), // value is cleared
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
				// output_to_lockbox is removed, so secret_value value is restored from the Lockbox secret to the state, and Lockbox version is deleted
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrWithValueFactory(resourceName, "secret_value", func() string {
						return originalSecretValue
					}),
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID)
					},
				),
			},
		},
	})
}

func TestAccIAMOAuthClientSecret_output_to_lockbox_updated_secret(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_oauth_client_secret.acceptance"
	clientName := "oauth-client-" + acctest.RandString(10)
	originalSecretValue := ""
	lockboxVersionID1 := ""
	lockboxVersionID2 := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckOAuthClientSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, ""),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get secret_value, to compare later
						secretValue, err := getResourceAttrValue(s, resourceName, "secret_value")
						originalSecretValue = secretValue
						return err
					},
				),
			},
			{
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_value", "secretValueIsHere",
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
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret_2.id", "secret_value", "secretValueIsHere", // changed secret
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
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrWithValueFactory(resourceName, "secret_value", func() string {
						return originalSecretValue
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
func TestAccIAMOAuthClientSecret_output_to_lockbox_updated_entries(t *testing.T) {
	t.Parallel()

	resourceName := "yandex_iam_oauth_client_secret.acceptance"
	clientName := "oauth-client-" + acctest.RandString(10)
	originalSecretValue := ""
	lockboxVersionID1 := ""
	lockboxVersionID2 := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckOAuthClientSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, ""),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get secret_value, to compare later
						secretValue, err := getResourceAttrValue(s, resourceName, "secret_value")
						originalSecretValue = secretValue
						return err
					},
				),
			},
			{
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_value", "secretValueIsHere",
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
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, testAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "secret_value", "nowIsHere", // changed entry key
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
				Config: testAccOAuthClientSecretConfigOutputToLockbox(clientName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrWithValueFactory(resourceName, "secret_value", func() string {
						return originalSecretValue
					}),
					func(s *terraform.State) error {
						return testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID2)
					},
				),
			},
		},
	})
}

func testAccCheckOAuthClientSecretDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_oauth_client_secret" {
			continue
		}

		_, err := config.sdk.IAM().OAuthClientSecret().Get(context.Background(), &iam.GetOAuthClientSecretRequest{
			OauthClientSecretId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("OAuthClientSecret still exists")
		}
	}

	return nil
}

func testAccCheckOAuthClientSecretExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		config := testAccProvider.Meta().(*Config)

		_, err := config.sdk.IAM().OAuthClientSecret().Get(context.Background(), &iam.GetOAuthClientSecretRequest{
			OauthClientSecretId: rs.Primary.ID,
		})

		return err
	}
}

func testAccOAuthClientSecretConfig(clientName, secretDesc string) string {
	testFolderID := os.Getenv("YC_FOLDER_ID")
	return fmt.Sprintf(`
resource "yandex_iam_oauth_client" "acceptance" {
  name       = "%s"
  folder_id  = "%s"
  scopes     = ["iam"]
}

resource "yandex_iam_oauth_client_secret" "acceptance" {
  oauth_client_id = "${yandex_iam_oauth_client.acceptance.id}"
  description     = "%s"
}
`, clientName, testFolderID, secretDesc)
}

func testAccOAuthClientSecretConfigEncrypted(clientName, key string) string {
	testFolderID := os.Getenv("YC_FOLDER_ID")
	return fmt.Sprintf(`
resource "yandex_iam_oauth_client" "acceptance" {
  name       = "%s"
  folder_id  = "%s"
  scopes     = ["iam"]
}

resource "yandex_iam_oauth_client_secret" "acceptance" {
  oauth_client_id = "${yandex_iam_oauth_client.acceptance.id}"
  description     = "description for test"
  pgp_key         = <<EOF
%s
EOF
}
`, clientName, testFolderID, key)
}

func testAccOAuthClientSecretConfigOutputToLockbox(clientName, outputBlock string) string {
	testFolderID := os.Getenv("YC_FOLDER_ID")
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "target_secret" {
  name = "%s"
}

resource "yandex_lockbox_secret" "target_secret_2" {
  name = "%s 2"
}

resource "yandex_iam_oauth_client" "acceptance" {
  name       = "%s"
  folder_id  = "%s"
  scopes     = ["iam"]
}

resource "yandex_iam_oauth_client_secret" "acceptance" {
  oauth_client_id = "${yandex_iam_oauth_client.acceptance.id}"
  description     = "description for test"

  %s
}
`, clientName, clientName, clientName, testFolderID, outputBlock)
}

func testAccOAuthClientSecretConfigJustSecret(clientName string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "target_secret" {
  name = "%s"
}

resource "yandex_lockbox_secret" "target_secret_2" {
  name = "%s 2"
}
`, clientName, clientName)
}
