package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
)

func init() {
	resource.AddTestSweepers("yandex_kms_asymmetric_signature_key", &resource.Sweeper{
		Name: "yandex_kms_asymmetric_signature_key",
		F:    testSweepKMSAsymmetricSignatureKey,
	})
}

func TestAccKMSAsymmetricSignatureKey_basic(t *testing.T) {
	t.Parallel()

	var asymmetricSignatureKey1 kms.AsymmetricSignatureKey
	var asymmetricSignatureKey2 kms.AsymmetricSignatureKey
	var asymmetricSignatureKey3 kms.AsymmetricSignatureKey

	key1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key3Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKMSAsymmetricSignatureKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricSignatureKey_basic(key1Name, key2Name, key3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyExists(
						"yandex_kms_asymmetric_signature_key.key-a", &asymmetricSignatureKey1),
					testAccCheckKMSAsymmetricSignatureKeyExists(
						"yandex_kms_asymmetric_signature_key.key-b", &asymmetricSignatureKey2),
					testAccCheckKMSAsymmetricSignatureKeyExists(
						"yandex_kms_asymmetric_signature_key.key-c", &asymmetricSignatureKey3),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_signature_key.key-a"),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_signature_key.key-b"),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_signature_key.key-c"),
				),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key.key-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key.key-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key.key-c",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKMSAsymmetricSignatureKey_deletion_protection(t *testing.T) {
	t.Parallel()

	var asymmetricSignatureKey1 kms.AsymmetricSignatureKey
	var asymmetricSignatureKey2 kms.AsymmetricSignatureKey
	var asymmetricSignatureKey3 kms.AsymmetricSignatureKey

	key1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key3Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKMSAsymmetricSignatureKeyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricSignatureKey_deletion_protection(key1Name, key2Name, key3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyExists(
						"yandex_kms_asymmetric_signature_key.key-a", &asymmetricSignatureKey1),
					testAccCheckKMSAsymmetricSignatureKeyExists(
						"yandex_kms_asymmetric_signature_key.key-b", &asymmetricSignatureKey2),
					testAccCheckKMSAsymmetricSignatureKeyExists(
						"yandex_kms_asymmetric_signature_key.key-c", &asymmetricSignatureKey3),
					testAccCheckBoolValue("yandex_kms_asymmetric_signature_key.key-a", "deletion_protection", true),
					testAccCheckBoolValue("yandex_kms_asymmetric_signature_key.key-b", "deletion_protection", false),
					testAccCheckBoolValue("yandex_kms_asymmetric_signature_key.key-c", "deletion_protection", false),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_signature_key.key-a"),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_signature_key.key-b"),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_signature_key.key-c"),
				),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key.key-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key.key-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key.key-c",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKmsAsymmetricSignatureKeyDeletionProtection_update(key1Name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoolValue("yandex_kms_asymmetric_signature_key.key-a", "deletion_protection", false),
				),
			},
		},
	})
}

func TestAccKMSAsymmetricSignatureKey_update(t *testing.T) {
	t.Parallel()

	var asymmetricSignatureKey1 kms.AsymmetricSignatureKey
	var asymmetricSignatureKey2 kms.AsymmetricSignatureKey
	var asymmetricSignatureKey3 kms.AsymmetricSignatureKey

	key1Name := acctest.RandomWithPrefix("tf-key-a")
	key2Name := acctest.RandomWithPrefix("tf-key-b")
	key3Name := acctest.RandomWithPrefix("tf-key-c")
	updatedKey1Name := key1Name + "-update"
	updatedKey2Name := key2Name + "-update"
	updatedKey3Name := key3Name + "-update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKMSAsymmetricSignatureKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricSignatureKey_basic(key1Name, key2Name, key3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.key-a", &asymmetricSignatureKey1),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-a", "name", key1Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-a", "description", "description for key-a"),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-a", "signature_algorithm", "RSA_2048_SIGN_PSS_SHA_256"),

					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey1, "tf-label", "tf-label-value-a"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey1, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_signature_key.key-a"),

					testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.key-b", &asymmetricSignatureKey2),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-b", "name", key2Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-b", "description", "description for key-b"),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-b", "signature_algorithm", "RSA_4096_SIGN_PSS_SHA_256"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey2, "tf-label", "tf-label-value-b"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey2, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_signature_key.key-b"),

					testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.key-c", &asymmetricSignatureKey3),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-c", "name", key3Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-c", "description", "description for key-c"),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-c", "signature_algorithm", "RSA_3072_SIGN_PSS_SHA_256"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey3, "tf-label", "tf-label-value-c"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey3, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_signature_key.key-c"),
				),
			},
			{
				Config: testAccKMSAsymmetricSignatureKey_update(updatedKey1Name, updatedKey2Name, updatedKey3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.key-a", &asymmetricSignatureKey1),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-a", "name", updatedKey1Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-a", "signature_algorithm", "RSA_2048_SIGN_PSS_SHA_256"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey1, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey1, "new-field", "only-shows-up-when-updated"),

					testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.key-b", &asymmetricSignatureKey2),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-b", "name", updatedKey2Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-b", "signature_algorithm", "RSA_4096_SIGN_PSS_SHA_256"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey2, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey2, "new-field", "only-shows-up-when-updated"),

					testAccCheckKMSAsymmetricSignatureKeyExists("yandex_kms_asymmetric_signature_key.key-c", &asymmetricSignatureKey3),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-c", "name", updatedKey3Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_signature_key.key-c", "signature_algorithm", "RSA_3072_SIGN_PSS_SHA_256"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey3, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckKMSAsymmetricSignatureKeyContainsLabel(&asymmetricSignatureKey3, "new-field", "only-shows-up-when-updated"),
				),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key.key-a",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  checkImportFolderID(getExampleFolderID()),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key.key-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key.key-c",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckKMSAsymmetricSignatureKeyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kms_asymmetric_signature_key" {
			continue
		}

		_, err := config.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().Get(context.Background(), &kms.GetAsymmetricSignatureKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("KMS AsymmetricSignatureKey still exists")
		}
	}

	return nil
}

func testAccCheckKMSAsymmetricSignatureKeyExists(name string, asymmetricSignatureKey *kms.AsymmetricSignatureKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().Get(context.Background(), &kms.GetAsymmetricSignatureKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("KMS AsymmetricSignatureKey not found")
		}

		*asymmetricSignatureKey = *found

		return nil
	}
}

func testAccCheckKMSAsymmetricSignatureKeyContainsLabel(asymmetricSignatureKey *kms.AsymmetricSignatureKey, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := asymmetricSignatureKey.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

//revive:disable:var-naming
func testAccKMSAsymmetricSignatureKey_basic(key1Name, key2Name, key3Name string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_signature_key" "key-a" {
  name              = "%s"
  description       = "description for key-a"
  signature_algorithm = "RSA_2048_SIGN_PSS_SHA_256"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }
}

resource "yandex_kms_asymmetric_signature_key" "key-b" {
  name              = "%s"
  description       = "description for key-b"
  signature_algorithm = "RSA_4096_SIGN_PSS_SHA_256"

  labels = {
    tf-label    = "tf-label-value-b"
    empty-label = ""
  }
}

resource "yandex_kms_asymmetric_signature_key" "key-c" {
  name              = "%s"
  description       = "description for key-c"
  signature_algorithm = "RSA_3072_SIGN_PSS_SHA_256"

  labels = {
    tf-label    = "tf-label-value-c"
    empty-label = ""
  }
}

`, key1Name, key2Name, key3Name)
}

//revive:disable:var-naming
func testAccKMSAsymmetricSignatureKey_deletion_protection(key1Name, key2Name, key3Name string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_signature_key" "key-a" {
  name                = "%s"
  description         = "description for key-a"
  deletion_protection = true
}

resource "yandex_kms_asymmetric_signature_key" "key-b" {
  name                = "%s"
  description         = "description for key-b"
  deletion_protection = false

}

resource "yandex_kms_asymmetric_signature_key" "key-c" {
  name        = "%s"
  description = "description for key-c"
}

`, key1Name, key2Name, key3Name)
}

func testAccKmsAsymmetricSignatureKeyDeletionProtection_update(keyName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_signature_key" "key-a" {
  name                = "%s"
  description         = "update deletion protection for key-a"
  deletion_protection = "%t"
}
`, keyName, deletionProtection)
}

func testAccKMSAsymmetricSignatureKey_update(key1Name, key2Name, key3Name string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_signature_key" "key-a" {
  name              = "%s"
  description       = "description with update for key-a"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_kms_asymmetric_signature_key" "key-b" {
  name              = "%s"
  description       = "description with update for key-b"
  signature_algorithm = "RSA_4096_SIGN_PSS_SHA_256"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_kms_asymmetric_signature_key" "key-c" {
  name              = "%s"
  description       = "description with update for key-c"
  signature_algorithm = "RSA_3072_SIGN_PSS_SHA_256"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, key1Name, key2Name, key3Name)
}

func testSweepKMSAsymmetricSignatureKey(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &kms.ListAsymmetricSignatureKeysRequest{FolderId: conf.FolderID}
	it := conf.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().AsymmetricSignatureKeyIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepKMSAsymmetricSignatureKey(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep KSMS Asymmetric Signature Key %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepKMSAsymmetricSignatureKey(conf *Config, id string) bool {
	return sweepWithRetry(sweepKMSAsymmetricSignatureKeyOnce, conf, "KMS Asymmetric Signature Key", id)
}

func sweepKMSAsymmetricSignatureKeyOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexKMSAsymmetricSignatureKeyDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().Delete(ctx, &kms.DeleteAsymmetricSignatureKeyRequest{
		KeyId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
