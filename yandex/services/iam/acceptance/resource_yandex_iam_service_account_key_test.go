package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	lockboxsdk "github.com/yandex-cloud/go-sdk/services/lockbox/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/vault/helper/pgpkeys"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	svc_iam "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/pluginsdk"
	svc_lockbox "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/lockbox/pluginsdk"
)

// Test that a service account key can be created and destroyed
func TestAccServiceAccountKey_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountKeyConfig(accountName, accountDesc, "description for test"),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckServiceAccountKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description for test"),
					resource.TestCheckResourceAttrSet(resourceName, "key_algorithm"),
					resource.TestCheckResourceAttrSet(resourceName, "public_key"),
					resource.TestCheckResourceAttrSet(resourceName, "private_key"),
					statecheck.CheckCreatedAtAttr(resourceName),
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
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	publicKey := pgpkeys.TestPubKey1
	fingerprints, _ := pgpkeys.GetFingerprints([]string{publicKey}, nil)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountKeyConfigEncrypted(accountName, accountDesc, publicKey),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckServiceAccountKeyExists(resourceName),
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
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	lockboxVersionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				// output_to_lockbox is defined, so sensitive fields are stored in Lockbox
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, svc_lockbox.TestAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "privateKeyIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckServiceAccountKeyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "public_key"),
					resource.TestCheckResourceAttr(resourceName, "private_key", ""), // value is not set in the state
					resource.TestCheckResourceAttrSet(resourceName, svc_iam.LockboxOutputVersionIdAttr),
					func(s *terraform.State) error {
						// get version ID, to check later
						versionID, err := statecheck.ResourceAttrValue(s, resourceName, svc_iam.LockboxOutputVersionIdAttr)
						lockboxVersionID = versionID
						return err
					},
				),
			},
			{
				// output_to_lockbox is removed, so private_key value is restored from the Lockbox secret to the state, and Lockbox version is deleted
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "private_key"), // value recovered from lockbox
					func(s *terraform.State) error {
						return suite.testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID)
					},
				),
			},
		},
	})
}

func TestAccServiceAccountKey_output_to_lockbox_on_destroy(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	lockboxVersionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				// output_to_lockbox is defined, so sensitive fields are stored in Lockbox
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, svc_lockbox.TestAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "privateKeyIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get Lockbox version ID, to check later
						versionID, err := statecheck.ResourceAttrValue(s, resourceName, svc_iam.LockboxOutputVersionIdAttr)
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
						return suite.testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID)
					},
				),
			},
		},
	})
}

func TestAccServiceAccountKey_output_to_lockbox_added_and_removed(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	originalPrivateKey := ""
	lockboxVersionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				// initially, output_to_lockbox is not defined
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckServiceAccountKeyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "key_algorithm"),
					resource.TestCheckResourceAttrSet(resourceName, "public_key"),
					resource.TestCheckResourceAttrSet(resourceName, "private_key"),
					func(s *terraform.State) error {
						// get private_key, to compare later
						privateKey, err := statecheck.ResourceAttrValue(s, resourceName, "private_key")
						originalPrivateKey = privateKey
						return err
					},
				),
			},
			{
				// output_to_lockbox is added, so private_key value is moved from the state (which is cleared) to Lockbox
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, svc_lockbox.TestAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "privateKeyIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "private_key", ""), // value is cleared
					resource.TestCheckResourceAttrSet(resourceName, svc_iam.LockboxOutputVersionIdAttr),
					func(s *terraform.State) error {
						// get Lockbox version ID, to check later
						versionID, err := statecheck.ResourceAttrValue(s, resourceName, svc_iam.LockboxOutputVersionIdAttr)
						lockboxVersionID = versionID
						return err
					},
				),
			},
			{
				// output_to_lockbox is removed, so private_key value is restored from the Lockbox secret to the state, and Lockbox version is deleted
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceAttrWithValueFactory(resourceName, "private_key", func() string {
						return originalPrivateKey
					}),
					func(s *terraform.State) error {
						return suite.testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID)
					},
				),
			},
		},
	})
}

func TestAccServiceAccountKey_output_to_lockbox_updated_secret(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	originalPrivateKey := ""
	lockboxVersionID1 := ""
	lockboxVersionID2 := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get private_key, to compare later
						privateKey, err := statecheck.ResourceAttrValue(s, resourceName, "private_key")
						originalPrivateKey = privateKey
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, svc_lockbox.TestAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "privateKeyIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get Lockbox version ID, to check later
						versionID, err := statecheck.ResourceAttrValue(s, resourceName, svc_iam.LockboxOutputVersionIdAttr)
						lockboxVersionID1 = versionID
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, svc_lockbox.TestAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret_2.id", "private_key", "privateKeyIsHere", // changed secret
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return suite.testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID1)
					},
					func(s *terraform.State) error {
						// get the new Lockbox version ID, to check later
						versionID, err := statecheck.ResourceAttrValue(s, resourceName, svc_iam.LockboxOutputVersionIdAttr)
						lockboxVersionID2 = versionID
						if lockboxVersionID1 == lockboxVersionID2 {
							return fmt.Errorf("a new version should have been created, but got the same version %s", lockboxVersionID1)
						}
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceAttrWithValueFactory(resourceName, "private_key", func() string {
						return originalPrivateKey
					}),
					func(s *terraform.State) error {
						return suite.testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret_2", lockboxVersionID2)
					},
				),
			},
		},
	})
}

// This test is almost equal to the previous one, but here we change entries
func TestAccServiceAccountKey_output_to_lockbox_updated_entries(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	resourceName := "yandex_iam_service_account_key.acceptance"
	accountName := "sa" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	originalPrivateKey := ""
	lockboxVersionID1 := ""
	lockboxVersionID2 := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckServiceAccountKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get private_key, to compare later
						privateKey, err := statecheck.ResourceAttrValue(s, resourceName, "private_key")
						originalPrivateKey = privateKey
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, svc_lockbox.TestAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "privateKeyIsHere",
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// get Lockbox version ID, to check later
						versionID, err := statecheck.ResourceAttrValue(s, resourceName, svc_iam.LockboxOutputVersionIdAttr)
						lockboxVersionID1 = versionID
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, svc_lockbox.TestAccOutputToLockbox(
					"yandex_lockbox_secret.target_secret.id", "private_key", "nowIsHere", // changed entry key
				)),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return suite.testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID1)
					},
					func(s *terraform.State) error {
						// get the new Lockbox version ID, to check later
						versionID, err := statecheck.ResourceAttrValue(s, resourceName, svc_iam.LockboxOutputVersionIdAttr)
						lockboxVersionID2 = versionID
						if lockboxVersionID1 == lockboxVersionID2 {
							return fmt.Errorf("a new version should have been created, but got the same version %s", lockboxVersionID1)
						}
						return err
					},
				),
			},
			{
				Config: testAccServiceAccountKeyConfigOutputToLockbox(accountName, accountDesc, ""),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceAttrWithValueFactory(resourceName, "private_key", func() string {
						return originalPrivateKey
					}),
					func(s *terraform.State) error {
						return suite.testAccCheckLockboxVersionDestroyed(s, "yandex_lockbox_secret.target_secret", lockboxVersionID2)
					},
				),
			},
		},
	})
}

func (suite testSuite) testAccCheckServiceAccountKeyDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_service_account_key" {
			continue
		}

		client := iamsdk.NewKeyClient(config.SDK)

		_, err := client.Get(context.Background(), &iam.GetKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("ServiceAccountKey still exists")
		}
	}

	return nil
}

func (suite testSuite) testAccCheckServiceAccountKeyExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := iamsdk.NewKeyClient(config.SDK)

		_, err := client.Get(context.Background(), &iam.GetKeyRequest{
			KeyId: rs.Primary.ID,
		})

		return err
	}
}

func (suite testSuite) testAccCheckLockboxVersionDestroyed(s *terraform.State, secretResourceName, versionID string) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	secretID, err := sdkcommon.GetResourceID(secretResourceName, s)
	if err != nil {
		return err
	}
	client := lockboxsdk.NewSecretClient(config.SDK)

	versions, _ := client.ListVersions(context.Background(), &lockbox.ListVersionsRequest{SecretId: secretID})
	for _, version := range versions.Versions {
		if version.GetId() == versionID {
			if version.GetStatus() == lockbox.Version_DESTROYED || version.GetStatus() == lockbox.Version_SCHEDULED_FOR_DESTRUCTION {
				return nil
			} else {
				return fmt.Errorf("the Lockbox version %s in secret %s should be DESTROYED or SCHEDULED_FOR_DESTRUCTION, but it's %v", versionID, secretID, version.GetStatus())
			}
		}
	}
	return fmt.Errorf("the lockbox version %s was not found in secret %s", versionID, secretID)
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

resource "yandex_lockbox_secret" "target_secret_2" {
  name = "%s 2"
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
`, name, name, name, desc, outputBlock)
}

func testAccServiceAccountKeyConfigJustSecret(name string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "target_secret" {
  name = "%s"
}

resource "yandex_lockbox_secret" "target_secret_2" {
  name = "%s 2"
}
`, name, name)
}
