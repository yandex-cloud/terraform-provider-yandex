package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceLockboxVersion_basic(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	basicData1 := "data.yandex_lockbox_secret_version.basic_version1"
	basicData2 := "data.yandex_lockbox_secret_version.basic_version2"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccLockboxSecretVersionKeyResource(secretName, BASIC_VERSION1_RESOURCE_AND_DATA),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLockboxSecretVersionExists(basicData1),
					testAccCheckResourceIDField(basicData1, "version_id"),
					testAccCheckYandexLockboxVersionStateEntries(basicData1, []*lockboxEntryCheck{
						{Key: "key1", Val: "val1"},
						{Key: "key2", Val: "val2"},
					}),
				),
			},
			{
				Config: testAccLockboxSecretVersionKeyResource(secretName, BASIC_VERSION2_RESOURCE_AND_DATA),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLockboxSecretVersionExists(basicData2),
					testAccCheckResourceIDField(basicData2, "version_id"),
					testAccCheckYandexLockboxVersionStateEntries(basicData2, []*lockboxEntryCheck{
						{Key: "key2", Val: "val2"},
						{Key: "key3", Val: "val3"},
					}),
				),
			},
		},
	})
}

func testAccLockboxSecretVersionKeyResource(name string, secret_version_config string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "basic_secret" {
  name        = "%v"
}
`, name) + secret_version_config
}

const BASIC_VERSION1_RESOURCE_AND_DATA = `
resource "yandex_lockbox_secret_version" "basic_version1" {
  secret_id   = yandex_lockbox_secret.basic_secret.id
  entries {
      key        = "key1"
      text_value = "val1"
  }
  entries {
      key        = "key2"
      text_value = "val2"
  }
}

data "yandex_lockbox_secret_version" "basic_version1" {
  secret_id = yandex_lockbox_secret.basic_secret.id
  version_id = yandex_lockbox_secret_version.basic_version1.id
}
`
const BASIC_VERSION2_RESOURCE_AND_DATA = `
resource "yandex_lockbox_secret_version" "basic_version2" {
  secret_id   = yandex_lockbox_secret.basic_secret.id
  entries {
      key        = "key2"
      text_value = "val2"
  }
  entries {
      key        = "key3"
      text_value = "val3"
  }
}

data "yandex_lockbox_secret_version" "basic_version2" {
  secret_id = yandex_lockbox_secret.basic_secret.id
  version_id = yandex_lockbox_secret_version.basic_version2.id
}
`

func testAccDataSourceLockboxSecretVersionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.LockboxPayload().Payload().Get(context.Background(), &lockbox.GetPayloadRequest{
			SecretId:  ds.Primary.Attributes["secret_id"],
			VersionId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.VersionId != ds.Primary.ID {
			return fmt.Errorf("secret version not found: %v (in secret %v)", ds.Primary.ID, ds.Primary.Attributes["secret_id"])
		}

		return nil
	}
}

// Checks expectedEntries in the state of versionResource
func testAccCheckYandexLockboxVersionStateEntries(versionResource string, expectedEntries []*lockboxEntryCheck) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[versionResource]
		attributes := r.Primary.Attributes
		count, _ := strconv.Atoi(attributes["entries.#"])
		if len(expectedEntries) > count {
			return fmt.Errorf("entries has %d values but we expected %d entries: %v", count, len(expectedEntries), expectedEntries)
		}
		for i := 0; i < count; i++ {
			key := attributes[fmt.Sprintf("entries.%d.key", i)]
			value := attributes[fmt.Sprintf("entries.%d.text_value", i)]
			expectedEntry := expectedEntries[i]
			if key != expectedEntry.Key {
				return fmt.Errorf("entry at index %d should have key '%s' but has key '%s'", i, expectedEntry.Key, key)
			}
			if value != expectedEntry.Val {
				return fmt.Errorf("entry at index %d should have value '%s' but has value '%s'", i, expectedEntry.Val, value)
			}
		}
		return nil
	}
}
