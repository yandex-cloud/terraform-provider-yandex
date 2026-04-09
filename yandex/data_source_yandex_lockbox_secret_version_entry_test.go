package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
)

// TestAccDataSourceLockboxSecretVersionEntry_basic verifies that both entries from the same
// version are correctly retrieved by key in a single Terraform config.
func TestAccDataSourceLockboxSecretVersionEntry_basic(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	dataSource1 := "data.yandex_lockbox_secret_version_entry.entry_key1"
	dataSource2 := "data.yandex_lockbox_secret_version_entry.entry_key2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccLockboxSecretVersionEntryConfig(secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLockboxSecretVersionEntryExists(dataSource1),
					resource.TestCheckResourceAttr(dataSource1, "key", "key1"),
					resource.TestCheckResourceAttr(dataSource1, "text_value", "val1"),
					resource.TestCheckResourceAttrSet(dataSource1, "version_id"),
					resource.TestCheckResourceAttrSet(dataSource1, "secret_id"),

					testAccDataSourceLockboxSecretVersionEntryExists(dataSource2),
					resource.TestCheckResourceAttr(dataSource2, "key", "key2"),
					resource.TestCheckResourceAttr(dataSource2, "text_value", "val2"),
					resource.TestCheckResourceAttrSet(dataSource2, "version_id"),
				),
			},
		},
	})
}

func TestAccDataSourceLockboxSecretVersionEntry_noVersionID(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	dataSource := "data.yandex_lockbox_secret_version_entry.entry_no_version"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccLockboxSecretVersionEntryNoVersionIDConfig(secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLockboxSecretVersionEntryExists(dataSource),
					resource.TestCheckResourceAttr(dataSource, "key", "api_key"),
					resource.TestCheckResourceAttr(dataSource, "text_value", "my-api-key"),
					// version_id must be populated even though it was not specified
					resource.TestCheckResourceAttrSet(dataSource, "version_id"),
				),
			},
		},
	})
}

func testAccDataSourceLockboxSecretVersionEntryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		secretID := ds.Primary.Attributes["secret_id"]
		versionID := ds.Primary.Attributes["version_id"]

		found, err := config.sdk.LockboxPayload().Payload().Get(context.Background(), &lockbox.GetPayloadRequest{
			SecretId:  secretID,
			VersionId: versionID,
		})
		if err != nil {
			return err
		}

		if found.VersionId != versionID {
			return fmt.Errorf("secret version not found: %v (in secret %v)", versionID, secretID)
		}

		return nil
	}
}

func testAccLockboxSecretVersionEntryConfig(secretName string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "test_secret" {
  name = "%s"
}

resource "yandex_lockbox_secret_version" "test_version" {
  secret_id = yandex_lockbox_secret.test_secret.id
  entries {
    key        = "key1"
    text_value = "val1"
  }
  entries {
    key        = "key2"
    text_value = "val2"
  }
}

data "yandex_lockbox_secret_version_entry" "entry_key1" {
  secret_id  = yandex_lockbox_secret.test_secret.id
  version_id = yandex_lockbox_secret_version.test_version.id
  key        = "key1"
}

data "yandex_lockbox_secret_version_entry" "entry_key2" {
  secret_id  = yandex_lockbox_secret.test_secret.id
  version_id = yandex_lockbox_secret_version.test_version.id
  key        = "key2"
}
`, secretName)
}

func testAccLockboxSecretVersionEntryNoVersionIDConfig(secretName string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "test_secret" {
  name = "%s"
}

resource "yandex_lockbox_secret_version" "test_version" {
  secret_id = yandex_lockbox_secret.test_secret.id
  entries {
    key        = "api_key"
    text_value = "my-api-key"
  }
  entries {
    key        = "username"
    text_value = "admin"
  }
}

data "yandex_lockbox_secret_version_entry" "entry_no_version" {
  secret_id = yandex_lockbox_secret.test_secret.id
  # version_id is intentionally omitted — should resolve to the latest version
  key = "api_key"

  depends_on = [yandex_lockbox_secret_version.test_version]
}
`, secretName)
}
