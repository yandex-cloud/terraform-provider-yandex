package acceptance

import (
	"fmt"
	"testing"

	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_lockbox "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/lockbox/pluginsdk"
)

func TestAccLockboxVersion_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	suite.commonTestAccLockboxVersion_basic(t, lockboxVersionOriginalOptions)
}

func TestAccLockboxVersion_update_description(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	suite.commonTestAccLockboxVersion_update_description(t, lockboxVersionOriginalOptions)
}

func TestAccLockboxVersion_update_entries(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	suite.commonTestAccLockboxVersion_update_entries(t, lockboxVersionOriginalOptions)
}

func TestAccLockboxVersion_add_and_delete(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	suite.commonTestAccLockboxVersion_add_and_delete(t, lockboxVersionOriginalOptions)
}

func TestAccLockboxVersion_delete_current_version(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	suite.commonTestAccLockboxVersion_delete_current_version(t, lockboxVersionOriginalOptions)
}

var lockboxVersionOriginalOptions = &lockboxVersionOptions{
	resourceType: "yandex_lockbox_secret_version",
	entriesToHcl: linesForEntries,
}

func linesForEntries(entries []*svc_lockbox.LockboxEntryCheck) string {
	result := ""
	for _, e := range entries {
		result += lineForEntry(e.Key, e.Val)
	}
	return result
}

func lineForEntry(k string, v string) string {
	return fmt.Sprintf(`
entries {
    key        = "%v"
    text_value = "%v"
}
`, k, v)
}
