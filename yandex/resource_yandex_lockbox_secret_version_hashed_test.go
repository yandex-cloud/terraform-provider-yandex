package yandex

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccLockboxVersionHashed_basic(t *testing.T) {
	commonTestAccLockboxVersion_basic(t, lockboxVersionHashedOptions)
}

func TestAccLockboxVersionHashed_update_description(t *testing.T) {
	commonTestAccLockboxVersion_update_description(t, lockboxVersionHashedOptions)
}

func TestAccLockboxVersionHashed_update_entries(t *testing.T) {
	commonTestAccLockboxVersion_update_entries(t, lockboxVersionHashedOptions)
}

func TestAccLockboxVersionHashed_add_and_delete(t *testing.T) {
	commonTestAccLockboxVersion_add_and_delete(t, lockboxVersionHashedOptions)
}

func TestAccLockboxVersionHashed_delete_current_version(t *testing.T) {
	commonTestAccLockboxVersion_delete_current_version(t, lockboxVersionHashedOptions)
}

func TestAccLockboxVersionHashed_values_hashed_in_state(t *testing.T) {
	versionOptions := &lockboxVersionOptions{
		resourceType: "yandex_lockbox_secret_version_hashed",
		entriesToHcl: linesForSafeEntries,
	}
	secretName := "a" + acctest.RandString(10)
	versionResource := versionOptions.resourceType + ".basic_version"
	entries := []*lockboxEntryCheck{
		{Key: "key1", Val: "some password"},
		{Key: "key2", Val: "another secret"},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "basic_version", Description: "basic", Entries: entries},
					},
				}),
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources[versionResource]
					if !ok {
						return fmt.Errorf("not found resource: %s", versionResource)
					}
					hashedValue1 := rs.Primary.Attributes["text_value_1"]
					hashedValue2 := rs.Primary.Attributes["text_value_2"]
					if len(hashedValue1) < 50 {
						return fmt.Errorf("text_value_1 hash of version %s is suspiciously short: %s", versionResource, hashedValue1)
					}
					if len(hashedValue2) < 50 {
						return fmt.Errorf("text_value_2 hash of version %s is suspiciously short: %s", versionResource, hashedValue2)
					}
					// Check that value in the state is not the original value
					if strings.Contains(hashedValue1, "some password") {
						return fmt.Errorf("text_value_1 of version %s contains secret value", versionResource)
					}
					if strings.Contains(hashedValue2, "another secret") {
						return fmt.Errorf("text_value_2 of version %s contains secret value", versionResource)
					}
					return nil
				},
			},
		},
	})
}

var lockboxVersionHashedOptions = &lockboxVersionOptions{
	resourceType: "yandex_lockbox_secret_version_hashed",
	entriesToHcl: linesForSafeEntries,
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
