package yandex

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
)

func init() {
	resource.AddTestSweepers("yandex_kms_symmetric_key", &resource.Sweeper{
		Name: "yandex_kms_symmetric_key",
		F:    testSweepKMSSymmetricKey,
	})
}

func TestAccKMSSymmetricKey_basic(t *testing.T) {
	t.Parallel()

	var symmetricKey1 kms.SymmetricKey
	var symmetricKey2 kms.SymmetricKey
	var symmetricKey3 kms.SymmetricKey

	key1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key3Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKMSSymmetricKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSSymmetricKey_basic(key1Name, key2Name, key3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSymmetricKeyExists(
						"yandex_kms_symmetric_key.key-a", &symmetricKey1),
					testAccCheckKMSSymmetricKeyExists(
						"yandex_kms_symmetric_key.key-b", &symmetricKey2),
					testAccCheckKMSSymmetricKeyExists(
						"yandex_kms_symmetric_key.key-c", &symmetricKey3),
					testAccCheckDuration("yandex_kms_symmetric_key.key-a", "rotation_period", "24h"),
					testAccCheckDuration("yandex_kms_symmetric_key.key-b", "rotation_period", "8760h"),
					testAccCheckDuration("yandex_kms_symmetric_key.key-c", "rotation_period", ""),
					testAccCheckCreatedAtAttr("yandex_kms_symmetric_key.key-a"),
					testAccCheckCreatedAtAttr("yandex_kms_symmetric_key.key-b"),
					testAccCheckCreatedAtAttr("yandex_kms_symmetric_key.key-c"),
				),
			},
			{
				ResourceName:      "yandex_kms_symmetric_key.key-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_symmetric_key.key-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_symmetric_key.key-c",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKMSSymmetricKey_deletion_protection(t *testing.T) {
	t.Parallel()

	var symmetricKey1 kms.SymmetricKey
	var symmetricKey2 kms.SymmetricKey
	var symmetricKey3 kms.SymmetricKey

	key1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key3Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKMSSymmetricKeyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccKMSSymmetricKey_deletion_protection(key1Name, key2Name, key3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSymmetricKeyExists(
						"yandex_kms_symmetric_key.key-a", &symmetricKey1),
					testAccCheckKMSSymmetricKeyExists(
						"yandex_kms_symmetric_key.key-b", &symmetricKey2),
					testAccCheckKMSSymmetricKeyExists(
						"yandex_kms_symmetric_key.key-c", &symmetricKey3),
					testAccCheckBoolValue("yandex_kms_symmetric_key.key-a", "deletion_protection", true),
					testAccCheckBoolValue("yandex_kms_symmetric_key.key-b", "deletion_protection", false),
					testAccCheckBoolValue("yandex_kms_symmetric_key.key-c", "deletion_protection", false),
					testAccCheckCreatedAtAttr("yandex_kms_symmetric_key.key-a"),
					testAccCheckCreatedAtAttr("yandex_kms_symmetric_key.key-b"),
					testAccCheckCreatedAtAttr("yandex_kms_symmetric_key.key-c"),
				),
			},
			{
				ResourceName:      "yandex_kms_symmetric_key.key-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_symmetric_key.key-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_symmetric_key.key-c",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKmsSymmetricKeyDeletionProtection_update(key1Name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoolValue("yandex_kms_symmetric_key.key-a", "deletion_protection", false),
				),
			},
		},
	})
}

func TestAccKMSSymmetricKey_update(t *testing.T) {
	t.Parallel()

	var symmetricKey1 kms.SymmetricKey
	var symmetricKey2 kms.SymmetricKey
	var symmetricKey3 kms.SymmetricKey

	key1Name := acctest.RandomWithPrefix("tf-key-a")
	key2Name := acctest.RandomWithPrefix("tf-key-b")
	key3Name := acctest.RandomWithPrefix("tf-key-c")
	updatedKey1Name := key1Name + "-update"
	updatedKey2Name := key2Name + "-update"
	updatedKey3Name := key3Name + "-update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKMSSymmetricKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSSymmetricKey_basic(key1Name, key2Name, key3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSymmetricKeyExists("yandex_kms_symmetric_key.key-a", &symmetricKey1),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-a", "name", key1Name),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-a", "description", "description for key-a"),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-a", "default_algorithm", "AES_128"),
					testAccCheckDuration("yandex_kms_symmetric_key.key-a", "rotation_period", "24h"),

					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey1, "tf-label", "tf-label-value-a"),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey1, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_kms_symmetric_key.key-a"),

					testAccCheckKMSSymmetricKeyExists("yandex_kms_symmetric_key.key-b", &symmetricKey2),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-b", "name", key2Name),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-b", "description", "description for key-b"),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-b", "default_algorithm", "AES_256"),
					testAccCheckDuration("yandex_kms_symmetric_key.key-b", "rotation_period", "8760h"),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey2, "tf-label", "tf-label-value-b"),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey2, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_kms_symmetric_key.key-b"),

					testAccCheckKMSSymmetricKeyExists("yandex_kms_symmetric_key.key-c", &symmetricKey3),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-c", "name", key3Name),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-c", "description", "description for key-c"),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-c", "default_algorithm", "AES_256"),
					testAccCheckDuration("yandex_kms_symmetric_key.key-c", "rotation_period", ""),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey3, "tf-label", "tf-label-value-c"),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey3, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_kms_symmetric_key.key-c"),
				),
			},
			{
				Config: testAccKMSSymmetricKey_update(updatedKey1Name, updatedKey2Name, updatedKey3Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSymmetricKeyExists("yandex_kms_symmetric_key.key-a", &symmetricKey1),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-a", "name", updatedKey1Name),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-a", "default_algorithm", "AES_192"),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-a", "rotation_period", ""),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey1, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey1, "new-field", "only-shows-up-when-updated"),

					testAccCheckKMSSymmetricKeyExists("yandex_kms_symmetric_key.key-b", &symmetricKey2),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-b", "name", updatedKey2Name),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-b", "default_algorithm", "AES_192"),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-b", "rotation_period", ""),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey2, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey2, "new-field", "only-shows-up-when-updated"),

					testAccCheckKMSSymmetricKeyExists("yandex_kms_symmetric_key.key-c", &symmetricKey3),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-c", "name", updatedKey3Name),
					resource.TestCheckResourceAttr("yandex_kms_symmetric_key.key-c", "default_algorithm", "AES_192"),
					testAccCheckDuration("yandex_kms_symmetric_key.key-c", "rotation_period", "8760h"),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey3, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckKMSSymmetricKeyContainsLabel(&symmetricKey3, "new-field", "only-shows-up-when-updated"),
				),
			},
			{
				ResourceName:      "yandex_kms_symmetric_key.key-a",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  checkImportFolderID(getExampleFolderID()),
			},
			{
				ResourceName:      "yandex_kms_symmetric_key.key-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_kms_symmetric_key.key-c",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkImportFolderID(folderID string) resource.ImportStateCheckFunc {
	return func(s []*terraform.InstanceState) error {
		if len(s) == 0 {
			return errors.New("No InstanceState found")
		}

		if len(s) != 1 {
			return fmt.Errorf("Expected one InstanceState, found: %d", len(s))
		}

		fID := s[0].Attributes["folder_id"]
		if fID != folderID {
			return fmt.Errorf("Expected folder_id %q, got %q", folderID, fID)
		}

		return nil
	}
}

func testAccCheckKMSSymmetricKeyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kms_symmetric_key" {
			continue
		}

		_, err := config.sdk.KMS().SymmetricKey().Get(context.Background(), &kms.GetSymmetricKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("KMS Symmetric Key still exists")
		}
	}

	return nil
}

func testAccCheckKMSSymmetricKeyExists(name string, symmetricKey *kms.SymmetricKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.KMS().SymmetricKey().Get(context.Background(), &kms.GetSymmetricKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("KMS Symmetric Key not found")
		}

		*symmetricKey = *found

		return nil
	}
}

func testAccCheckKMSSymmetricKeyContainsLabel(symmetricKey *kms.SymmetricKey, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := symmetricKey.Labels[key]
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
func testAccKMSSymmetricKey_basic(key1Name, key2Name, key3Name string) string {
	return fmt.Sprintf(`
resource "yandex_kms_symmetric_key" "key-a" {
  name              = "%s"
  description       = "description for key-a"
  default_algorithm = "AES_128"
  rotation_period   = "24h"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }
}

resource "yandex_kms_symmetric_key" "key-b" {
  name              = "%s"
  description       = "description for key-b"
  default_algorithm = "AES_256"
  rotation_period   = "8760h"   // equal 1 year

  labels = {
    tf-label    = "tf-label-value-b"
    empty-label = ""
  }
}

resource "yandex_kms_symmetric_key" "key-c" {
  name              = "%s"
  description       = "description for key-c"
  default_algorithm = "AES_256"

  labels = {
    tf-label    = "tf-label-value-c"
    empty-label = ""
  }
}

`, key1Name, key2Name, key3Name)
}

//revive:disable:var-naming
func testAccKMSSymmetricKey_deletion_protection(key1Name, key2Name, key3Name string) string {
	return fmt.Sprintf(`
resource "yandex_kms_symmetric_key" "key-a" {
  name                = "%s"
  description         = "description for key-a"
  deletion_protection = true
}

resource "yandex_kms_symmetric_key" "key-b" {
  name                = "%s"
  description         = "description for key-b"
  deletion_protection = false

}

resource "yandex_kms_symmetric_key" "key-c" {
  name        = "%s"
  description = "description for key-c"
}

`, key1Name, key2Name, key3Name)
}

func testAccKmsSymmetricKeyDeletionProtection_update(keyName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "yandex_kms_symmetric_key" "key-a" {
  name                = "%s"
  description         = "update deletion protection for key-a"
  deletion_protection = "%t"
}
`, keyName, deletionProtection)
}

func testAccKMSSymmetricKey_update(key1Name, key2Name, key3Name string) string {
	return fmt.Sprintf(`
resource "yandex_kms_symmetric_key" "key-a" {
  name              = "%s"
  description       = "description with update for key-a"
  default_algorithm = "AES_192"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_kms_symmetric_key" "key-b" {
  name              = "%s"
  description       = "description with update for key-b"
  default_algorithm = "AES_192"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_kms_symmetric_key" "key-c" {
  name              = "%s"
  description       = "description with update for key-c"
  default_algorithm = "AES_192"
  rotation_period   = "8760h"   // equal 1 year

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, key1Name, key2Name, key3Name)
}

func testSweepKMSSymmetricKey(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &kms.ListSymmetricKeysRequest{FolderId: conf.FolderID}
	it := conf.sdk.KMS().SymmetricKey().SymmetricKeyIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepKMSSymmetricKey(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep KSM symmetric key %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepKMSSymmetricKey(conf *Config, id string) bool {
	return sweepWithRetry(sweepKMSSymmetricKeyOnce, conf, "KMS Symmetric Key", id)
}

func sweepKMSSymmetricKeyOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexKMSSymmetricKeyDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.KMS().SymmetricKey().Delete(ctx, &kms.DeleteSymmetricKeyRequest{
		KeyId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
