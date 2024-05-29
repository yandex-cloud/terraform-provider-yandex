package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLockboxVersionHashed_basic(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	secretDesc := "Terraform test secret"
	versionDesc := "Terraform test version"
	secretResource := "yandex_lockbox_secret.basic_secret"
	versionResource := "yandex_lockbox_secret_version_hashed.basic_version"
	versionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret and version
				Config: testAccLockboxSecretVersionHashedBasic(secretName, secretDesc, versionDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID), // stores current versionID
					resource.TestCheckResourceAttr(versionResource, "description", versionDesc),
					testAccCheckYandexLockboxVersionEntries(versionResource, []*lockboxEntryCheck{
						{Key: "key1", Val: "val1"},
						{Key: "key2", Val: "val2"},
					}),
				),
			},
			{
				// update version description (will add a new version to the secret)
				Config: testAccLockboxSecretVersionHashedBasic(secretName, secretDesc, versionDesc+" updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID), // checks that now versionID is different
					resource.TestCheckResourceAttr(versionResource, "description", versionDesc+" updated"),
					testAccCheckYandexLockboxVersionEntries(versionResource, []*lockboxEntryCheck{
						{Key: "key1", Val: "val1"},
						{Key: "key2", Val: "val2"},
					}),
				),
			},
		},
	})
}

func TestAccLockboxVersionHashed_update_entries(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	secretDesc := "Terraform test secret"
	versionDesc := "Terraform test version"
	secretResource := "yandex_lockbox_secret.basic_secret"
	versionResource := "yandex_lockbox_secret_version_hashed.basic_version"
	versionID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret and version
				Config: testAccLockboxSecretVersionHashed(secretName, secretDesc, versionDesc, []*lockboxEntryCheck{
					{Key: "key1", Val: "val1"},
					{Key: "key2", Val: "val2"},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID),
					resource.TestCheckResourceAttr(versionResource, "description", versionDesc),
					testAccCheckYandexLockboxVersionEntries(versionResource, []*lockboxEntryCheck{
						{Key: "key1", Val: "val1"},
						{Key: "key2", Val: "val2"},
					}),
				),
			},
			{
				// modify entries
				Config: testAccLockboxSecretVersionHashed(secretName, secretDesc, versionDesc, []*lockboxEntryCheck{
					// {Key: "key1", Val: "val1"}, // remove
					{Key: "key2", Val: "val22"}, // modify
					{Key: "key3", Val: "val3"},  // add
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, &versionID),
					resource.TestCheckResourceAttr(versionResource, "description", versionDesc),
					testAccCheckYandexLockboxVersionEntries(versionResource, []*lockboxEntryCheck{
						{Key: "key2", Val: "val22"},
						{Key: "key3", Val: "val3"},
					}),
				),
			},
		},
	})
}

func testAccLockboxSecretVersionHashedBasic(name, secretDesc, versionDesc string) string {
	entries := []*lockboxEntryCheck{
		{Key: "key1", Val: "val1"},
		{Key: "key2", Val: "val2"},
	}
	return testAccLockboxSecretVersionHashed(name, secretDesc, versionDesc, entries)
}

func testAccLockboxSecretVersionHashed(name, secretDesc, versionDesc string, entries []*lockboxEntryCheck) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "basic_secret" {
  name        = "%v"
  description = "%v"
}

resource "yandex_lockbox_secret_version_hashed" "basic_version" {
  secret_id = yandex_lockbox_secret.basic_secret.id
  description = "%v"
  %v
}
`, name, secretDesc, versionDesc, linesForSafeEntries(entries))
}

func linesForSafeEntries(entries []*lockboxEntryCheck) string {
	result := ""
	for i, e := range entries {
		result += lineForSafeEntry(i+1, e.Key, e.Val) // safe entries start with index 1 (key_1, key_2, etc)
	}
	return result
}

func lineForSafeEntry(i int, k string, v string) string {
	return fmt.Sprintf(`
  key_%v        = "%v"
  text_value_%v = "%v"
`, i, k, i, v)
}
