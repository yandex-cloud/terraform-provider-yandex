package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
)

func init() {
	resource.AddTestSweepers("yandex_kms_asymmetric_encryption_key", &resource.Sweeper{
		Name: "yandex_kms_asymmetric_encryption_key",
		F:    testSweepKMSAsymmetricEncryptionKey,
	})
}

func TestAccKMSAsymmetricEncryptionKey_basic(t *testing.T) {
	t.Parallel()

	var asymmetricEncryptionKey1 kms.AsymmetricEncryptionKey
	var asymmetricEncryptionKey2 kms.AsymmetricEncryptionKey
	var asymmetricEncryptionKey3 kms.AsymmetricEncryptionKey

	key1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key3Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKMSAsymmetricEncryptionKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricEncryptionKey_basic(key1Name, key2Name, key3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricEncryptionKeyExists(
						"yandex_kms_asymmetric_encryption_key.key-a", &asymmetricEncryptionKey1),
					testAccCheckKMSAsymmetricEncryptionKeyExists(
						"yandex_kms_asymmetric_encryption_key.key-b", &asymmetricEncryptionKey2),
					testAccCheckKMSAsymmetricEncryptionKeyExists(
						"yandex_kms_asymmetric_encryption_key.key-c", &asymmetricEncryptionKey3),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_encryption_key.key-a"),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_encryption_key.key-b"),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_encryption_key.key-c"),
				),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key.key-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key.key-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key.key-c",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKMSAsymmetricEncryptionKey_deletion_protection(t *testing.T) {
	t.Parallel()

	var asymmetricEncryptionKey1 kms.AsymmetricEncryptionKey
	var asymmetricEncryptionKey2 kms.AsymmetricEncryptionKey
	var asymmetricEncryptionKey3 kms.AsymmetricEncryptionKey

	key1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key3Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKMSAsymmetricEncryptionKeyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricEncryptionKey_deletion_protection(key1Name, key2Name, key3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricEncryptionKeyExists(
						"yandex_kms_asymmetric_encryption_key.key-a", &asymmetricEncryptionKey1),
					testAccCheckKMSAsymmetricEncryptionKeyExists(
						"yandex_kms_asymmetric_encryption_key.key-b", &asymmetricEncryptionKey2),
					testAccCheckKMSAsymmetricEncryptionKeyExists(
						"yandex_kms_asymmetric_encryption_key.key-c", &asymmetricEncryptionKey3),
					testAccCheckBoolValue("yandex_kms_asymmetric_encryption_key.key-a", "deletion_protection", true),
					testAccCheckBoolValue("yandex_kms_asymmetric_encryption_key.key-b", "deletion_protection", false),
					testAccCheckBoolValue("yandex_kms_asymmetric_encryption_key.key-c", "deletion_protection", false),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_encryption_key.key-a"),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_encryption_key.key-b"),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_encryption_key.key-c"),
				),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key.key-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key.key-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key.key-c",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKmsAsymmetricEncryptionKeyDeletionProtection_update(key1Name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoolValue("yandex_kms_asymmetric_encryption_key.key-a", "deletion_protection", false),
				),
			},
		},
	})
}

func TestAccKMSAsymmetricEncryptionKey_update(t *testing.T) {
	t.Parallel()

	var asymmetricEncryptionKey1 kms.AsymmetricEncryptionKey
	var asymmetricEncryptionKey2 kms.AsymmetricEncryptionKey
	var asymmetricEncryptionKey3 kms.AsymmetricEncryptionKey

	key1Name := acctest.RandomWithPrefix("tf-key-a")
	key2Name := acctest.RandomWithPrefix("tf-key-b")
	key3Name := acctest.RandomWithPrefix("tf-key-c")
	updatedKey1Name := key1Name + "-update"
	updatedKey2Name := key2Name + "-update"
	updatedKey3Name := key3Name + "-update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKMSAsymmetricEncryptionKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricEncryptionKey_basic(key1Name, key2Name, key3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricEncryptionKeyExists("yandex_kms_asymmetric_encryption_key.key-a", &asymmetricEncryptionKey1),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-a", "name", key1Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-a", "description", "description for key-a"),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-a", "encryption_algorithm", "RSA_2048_ENC_OAEP_SHA_256"),

					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey1, "tf-label", "tf-label-value-a"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey1, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_encryption_key.key-a"),

					testAccCheckKMSAsymmetricEncryptionKeyExists("yandex_kms_asymmetric_encryption_key.key-b", &asymmetricEncryptionKey2),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-b", "name", key2Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-b", "description", "description for key-b"),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-b", "encryption_algorithm", "RSA_4096_ENC_OAEP_SHA_256"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey2, "tf-label", "tf-label-value-b"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey2, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_encryption_key.key-b"),

					testAccCheckKMSAsymmetricEncryptionKeyExists("yandex_kms_asymmetric_encryption_key.key-c", &asymmetricEncryptionKey3),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-c", "name", key3Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-c", "description", "description for key-c"),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-c", "encryption_algorithm", "RSA_3072_ENC_OAEP_SHA_256"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey3, "tf-label", "tf-label-value-c"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey3, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_kms_asymmetric_encryption_key.key-c"),
				),
			},
			{
				Config: testAccKMSAsymmetricEncryptionKey_update(updatedKey1Name, updatedKey2Name, updatedKey3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricEncryptionKeyExists("yandex_kms_asymmetric_encryption_key.key-a", &asymmetricEncryptionKey1),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-a", "name", updatedKey1Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-a", "encryption_algorithm", "RSA_2048_ENC_OAEP_SHA_256"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey1, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey1, "new-field", "only-shows-up-when-updated"),

					testAccCheckKMSAsymmetricEncryptionKeyExists("yandex_kms_asymmetric_encryption_key.key-b", &asymmetricEncryptionKey2),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-b", "name", updatedKey2Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-b", "encryption_algorithm", "RSA_4096_ENC_OAEP_SHA_256"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey2, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey2, "new-field", "only-shows-up-when-updated"),

					testAccCheckKMSAsymmetricEncryptionKeyExists("yandex_kms_asymmetric_encryption_key.key-c", &asymmetricEncryptionKey3),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-c", "name", updatedKey3Name),
					resource.TestCheckResourceAttr("yandex_kms_asymmetric_encryption_key.key-c", "encryption_algorithm", "RSA_3072_ENC_OAEP_SHA_256"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey3, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(&asymmetricEncryptionKey3, "new-field", "only-shows-up-when-updated"),
				),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key.key-a",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  checkImportFolderID(getExampleFolderID()),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key.key-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key.key-c",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckKMSAsymmetricEncryptionKeyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kms_asymmetric_encryption_key" {
			continue
		}

		_, err := config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().Get(context.Background(), &kms.GetAsymmetricEncryptionKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("KMS AsymmetricEncryptionKey still exists")
		}
	}

	return nil
}

func testAccCheckKMSAsymmetricEncryptionKeyExists(name string, asymmetricEncryptionKey *kms.AsymmetricEncryptionKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().Get(context.Background(), &kms.GetAsymmetricEncryptionKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("KMS AsymmetricEncryptionKey not found")
		}

		*asymmetricEncryptionKey = *found

		return nil
	}
}

func testAccCheckKMSAsymmetricEncryptionKeyContainsLabel(asymmetricEncryptionKey *kms.AsymmetricEncryptionKey, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := asymmetricEncryptionKey.Labels[key]
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
func testAccKMSAsymmetricEncryptionKey_basic(key1Name, key2Name, key3Name string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_encryption_key" "key-a" {
  name              = "%s"
  description       = "description for key-a"
  encryption_algorithm = "RSA_2048_ENC_OAEP_SHA_256"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }
}

resource "yandex_kms_asymmetric_encryption_key" "key-b" {
  name              = "%s"
  description       = "description for key-b"
  encryption_algorithm = "RSA_4096_ENC_OAEP_SHA_256"

  labels = {
    tf-label    = "tf-label-value-b"
    empty-label = ""
  }
}

resource "yandex_kms_asymmetric_encryption_key" "key-c" {
  name              = "%s"
  description       = "description for key-c"
  encryption_algorithm = "RSA_3072_ENC_OAEP_SHA_256"

  labels = {
    tf-label    = "tf-label-value-c"
    empty-label = ""
  }
}

`, key1Name, key2Name, key3Name)
}

//revive:disable:var-naming
func testAccKMSAsymmetricEncryptionKey_deletion_protection(key1Name, key2Name, key3Name string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_encryption_key" "key-a" {
  name                = "%s"
  description         = "description for key-a"
  deletion_protection = true
}

resource "yandex_kms_asymmetric_encryption_key" "key-b" {
  name                = "%s"
  description         = "description for key-b"
  deletion_protection = false

}

resource "yandex_kms_asymmetric_encryption_key" "key-c" {
  name        = "%s"
  description = "description for key-c"
}

`, key1Name, key2Name, key3Name)
}

func testAccKmsAsymmetricEncryptionKeyDeletionProtection_update(keyName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_encryption_key" "key-a" {
  name                = "%s"
  description         = "update deletion protection for key-a"
  deletion_protection = "%t"
}
`, keyName, deletionProtection)
}

func testAccKMSAsymmetricEncryptionKey_update(key1Name, key2Name, key3Name string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_encryption_key" "key-a" {
  name              = "%s"
  description       = "description with update for key-a"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_kms_asymmetric_encryption_key" "key-b" {
  name              = "%s"
  description       = "description with update for key-b"
  encryption_algorithm = "RSA_4096_ENC_OAEP_SHA_256"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_kms_asymmetric_encryption_key" "key-c" {
  name              = "%s"
  description       = "description with update for key-c"
  encryption_algorithm = "RSA_3072_ENC_OAEP_SHA_256"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, key1Name, key2Name, key3Name)
}

func testSweepKMSAsymmetricEncryptionKey(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &kms.ListAsymmetricEncryptionKeysRequest{FolderId: conf.FolderID}
	it := conf.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().AsymmetricEncryptionKeyIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepKMSAsymmetricEncryptionKey(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep KSMS Asymmetric Encryption Key %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepKMSAsymmetricEncryptionKey(conf *Config, id string) bool {
	return sweepWithRetry(sweepKMSAsymmetricEncryptionKeyOnce, conf, "KMS Asymmetric Encryption Key", id)
}

func sweepKMSAsymmetricEncryptionKeyOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexKMSAsymmetricEncryptionKeyDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().Delete(ctx, &kms.DeleteAsymmetricEncryptionKeyRequest{
		KeyId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
