package yandex

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
)

// Here we write tests that can be common both in the original resource, and in the hashed variant

// lockboxVersionOptions struct is different for each version variant (original and hashed)
type lockboxVersionOptions struct {
	resourceType string
	entriesToHcl EntriesToHcl
}

type lockboxVersionsData struct {
	options  *lockboxVersionOptions
	versions []*lockboxVersionData
}

type lockboxVersionData struct {
	ResourceName string
	Description  string
	Entries      []*lockboxEntryCheck
}

type EntriesToHcl func(entries []*lockboxEntryCheck) string

func commonTestAccLockboxVersion_basic(t *testing.T, versionOptions *lockboxVersionOptions) {
	secretName := "a" + acctest.RandString(10)
	secretResource := "yandex_lockbox_secret.basic_secret"
	versionResource := versionOptions.resourceType + ".basic_version"
	entries := []*lockboxEntryCheck{
		{Key: "key1", Val: "val1"},
		{Key: "key2", Val: "val2"},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret and version
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "basic_version", Description: "basic", Entries: entries},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource, nil),
					resource.TestCheckResourceAttr(versionResource, "description", "basic"),
					testAccCheckYandexLockboxVersionEntries(versionResource, entries),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"ACTIVE": 1,
					}),
				),
			},
		},
	})
}

func commonTestAccLockboxVersion_add_and_delete(t *testing.T, versionOptions *lockboxVersionOptions) {
	secretName := "a" + acctest.RandString(10)
	secretResource := "yandex_lockbox_secret.basic_secret"
	versionResource1 := versionOptions.resourceType + ".version1"
	versionResource2 := versionOptions.resourceType + ".version2"
	entries1 := []*lockboxEntryCheck{
		{Key: "key1", Val: "val1"},
		{Key: "key2", Val: "val2"},
	}
	entries2 := []*lockboxEntryCheck{
		{Key: "key1", Val: "val11"},
		{Key: "key3", Val: "val3"},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret and version
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "version1", Description: "first", Entries: entries1},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource1, nil),
					resource.TestCheckResourceAttr(versionResource1, "description", "first"),
					testAccCheckYandexLockboxVersionEntries(versionResource1, entries1),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"ACTIVE": 1,
					}),
				),
			},
			{
				// define additional version
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "version1", Description: "first", Entries: entries1},
						{ResourceName: "version2", Description: "second", Entries: entries2},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource1, nil),
					testAccCheckYandexLockboxResourceExists(versionResource2, nil),
					resource.TestCheckResourceAttr(versionResource2, "description", "second"),
					testAccCheckYandexLockboxVersionEntries(versionResource1, entries1),
					testAccCheckYandexLockboxVersionEntries(versionResource2, entries2),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"ACTIVE": 2,
					}),
				),
			},
			{
				// delete old version
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "version2", Description: "second", Entries: entries2},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceDoesNotExist(versionResource1),
					testAccCheckYandexLockboxResourceExists(versionResource2, nil),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"ACTIVE":                    1,
						"SCHEDULED_FOR_DESTRUCTION": 1,
					}),
				),
			},
		},
	})
}

func commonTestAccLockboxVersion_update_description(t *testing.T, versionOptions *lockboxVersionOptions) {
	secretName := "a" + acctest.RandString(10)
	secretResource := "yandex_lockbox_secret.basic_secret"
	versionResource := versionOptions.resourceType + ".same_version"
	versionID := ""
	entries := []*lockboxEntryCheck{
		{Key: "key1", Val: "val1"},
		{Key: "key2", Val: "val2"},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret and version
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "same_version", Description: "first", Entries: entries},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID),
					resource.TestCheckResourceAttr(versionResource, "description", "first"),
					testAccCheckYandexLockboxVersionEntries(versionResource, entries),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"ACTIVE": 1,
					}),
				),
			},
			{
				// update description will create a new version
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "same_version", Description: "second", Entries: entries},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID),
					resource.TestCheckResourceAttr(versionResource, "description", "second"),
					testAccCheckYandexLockboxVersionEntries(versionResource, entries),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"ACTIVE":                    1,
						"SCHEDULED_FOR_DESTRUCTION": 1,
					}),
				),
			},
		},
	})
}

func commonTestAccLockboxVersion_update_entries(t *testing.T, versionOptions *lockboxVersionOptions) {
	secretName := "a" + acctest.RandString(10)
	secretResource := "yandex_lockbox_secret.basic_secret"
	versionResource := versionOptions.resourceType + ".same_version"
	versionID := ""
	entries1 := []*lockboxEntryCheck{
		{Key: "password", Val: "initial"},
	}
	entries2 := []*lockboxEntryCheck{
		{Key: "password", Val: "changed"},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret and version
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "same_version", Entries: entries1},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID),
					testAccCheckYandexLockboxVersionEntries(versionResource, entries1),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"ACTIVE": 1,
					}),
				),
			},
			{
				// update entries will create a new version
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "same_version", Entries: entries2},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID),
					testAccCheckYandexLockboxVersionEntries(versionResource, entries2),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"ACTIVE":                    1,
						"SCHEDULED_FOR_DESTRUCTION": 1,
					}),
				),
			},
		},
	})
}

func commonTestAccLockboxVersion_delete_current_version(t *testing.T, versionOptions *lockboxVersionOptions) {
	secretName := "a" + acctest.RandString(10)
	secretResource := "yandex_lockbox_secret.basic_secret"
	versionResource := versionOptions.resourceType + ".basic_version"
	entries := []*lockboxEntryCheck{
		{Key: "key1", Val: "val1"},
		{Key: "key2", Val: "val2"},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret and version
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options: versionOptions,
					versions: []*lockboxVersionData{
						{ResourceName: "basic_version", Description: "basic", Entries: entries},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceExists(versionResource, nil),
					resource.TestCheckResourceAttr(versionResource, "description", "basic"),
					testAccCheckYandexLockboxVersionEntries(versionResource, entries),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"ACTIVE": 1,
					}),
				),
			},
			{
				// delete version (it's the current one)
				Config: testAccLockboxSecretAndVersions(secretName, &lockboxVersionsData{
					options:  versionOptions,
					versions: []*lockboxVersionData{},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, nil),
					testAccCheckYandexLockboxResourceDoesNotExist(versionResource),
					testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource, map[string]int{
						"SCHEDULED_FOR_DESTRUCTION": 1,
					}),
				),
			},
		},
	})
}

func testAccLockboxSecretAndVersions(secretName string, versionsData *lockboxVersionsData) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "basic_secret" {
  name        = "%s"
}

%s
`, secretName, testAccLockboxSecretVersions(versionsData))
}

func testAccLockboxSecretVersions(versionsData *lockboxVersionsData) string {
	strArr := make([]string, len(versionsData.versions))
	for i, v := range versionsData.versions {
		strArr[i] = testAccLockboxSecretVersion(v, versionsData.options.resourceType, versionsData.options.entriesToHcl)
	}
	return strings.Join(strArr, "")
}

func testAccLockboxSecretVersion(version *lockboxVersionData, versionResourceType string, entriesToHcl EntriesToHcl) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
  secret_id = yandex_lockbox_secret.basic_secret.id
  description = "%v"
  %v
}
`, versionResourceType, version.ResourceName, version.Description, entriesToHcl(version.Entries))
}

func testAccCheckYandexLockboxSecretVersionStatusCounts(secretResource string, expectedStatusCounts map[string]int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		rs, ok := s.RootModule().Resources[secretResource]
		if !ok {
			return fmt.Errorf("not found resource: %s", secretResource)
		}
		response, err := config.sdk.LockboxSecret().Secret().ListVersions(context.Background(), &lockbox.ListVersionsRequest{
			SecretId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		var statusCounts = make(map[string]int)
		for _, version := range response.Versions {
			if _, ok := statusCounts[version.Status.String()]; !ok {
				statusCounts[version.Status.String()] = 0
			}
			statusCounts[version.Status.String()]++
		}

		if !reflect.DeepEqual(expectedStatusCounts, statusCounts) {
			return fmt.Errorf("expected %v but found %v", expectedStatusCounts, statusCounts)
		}
		return nil
	}
}

// Checks expectedEntries in the real secret version of the versionResource
// We can't check entries in state because the resource doesn't read the entries.
func testAccCheckYandexLockboxVersionEntries(versionResource string, expectedEntries []*lockboxEntryCheck) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		rs, ok := s.RootModule().Resources[versionResource]
		if !ok {
			return fmt.Errorf("not found resource: %s", versionResource)
		}
		payload, err := config.sdk.LockboxPayload().Payload().Get(context.Background(), &lockbox.GetPayloadRequest{
			SecretId:  rs.Primary.Attributes["secret_id"],
			VersionId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}
		if len(expectedEntries) != len(payload.GetEntries()) {
			return fmt.Errorf("expected %d entries but found %d", len(expectedEntries), len(payload.GetEntries()))
		}
		for i, entry := range payload.GetEntries() {
			expectedEntry := expectedEntries[i]
			if entry.Key != expectedEntry.Key {
				return fmt.Errorf("entry at index %d should have key '%s' but has key '%s'", i, expectedEntry.Key, entry.Key)
			}
			if expectedEntry.Regexp != nil {
				if !expectedEntry.Regexp.MatchString(entry.GetTextValue()) {
					return fmt.Errorf("entry at index %d should have value that matches '%v' but has value '%s'", i, expectedEntry.Regexp, entry.GetTextValue())
				}
			} else {
				if entry.GetTextValue() != expectedEntry.Val {
					return fmt.Errorf("entry at index %d should have value '%s' but has value '%s'", i, expectedEntry.Val, entry.GetTextValue())
				}
			}
		}
		return nil
	}
}
