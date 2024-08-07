package yandex

import (
	"fmt"
	"testing"
)

func TestAccLockboxVersion_basic(t *testing.T) {
	commonTestAccLockboxVersion_basic(t, lockboxVersionOriginalOptions)
}

func TestAccLockboxVersion_update_description(t *testing.T) {
	commonTestAccLockboxVersion_update_description(t, lockboxVersionOriginalOptions)
}

func TestAccLockboxVersion_update_entries(t *testing.T) {
	commonTestAccLockboxVersion_update_entries(t, lockboxVersionOriginalOptions)
}

func TestAccLockboxVersion_add_and_delete(t *testing.T) {
	commonTestAccLockboxVersion_add_and_delete(t, lockboxVersionOriginalOptions)
}

func TestAccLockboxVersion_delete_current_version(t *testing.T) {
	commonTestAccLockboxVersion_delete_current_version(t, lockboxVersionOriginalOptions)
}

var lockboxVersionOriginalOptions = &lockboxVersionOptions{
	resourceType: "yandex_lockbox_secret_version",
	entriesToHcl: linesForEntries,
}

func linesForEntries(entries []*lockboxEntryCheck) string {
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

/*
func TestAccLockboxVersion_command(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	versionResource := "yandex_lockbox_secret_version.exec_version"
	versionID := ""

	// TODO - although checkTestFilesFolder output looks fine, we can't read the script file
	//  we get: Error: fork/exec /go/src/github.com/terraform-providers/terraform-provider-yandex/yandex/test-fixtures/fake_secret_generator.sh: no such file or directory
	checkTestFilesFolder(t)
	// TODO - this also doen't work, the file is created but it's not found later (same "no such file or directory" error)
	scriptFile := createTempFile(t, "fake_secret_generator.sh", `#!/bin/bash
set -e
# As a proof of concept, we return an argument, an env var and a random number.
echo -n "arg: $1, var: $VALUE, rnd: $RANDOM"
`)
	defer os.Remove(scriptFile.Name())
	script := scriptFile.Name()
	t.Logf("Created a temp file for the script: %v", script)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret and version
				Config: testAccLockboxSecretVersionWithCommand(secretName, "echo", "dummy value", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID),
					testAccCheckYandexLockboxVersionEntries(versionResource, []*lockboxEntryCheck{
						{Key: "k1", Val: "dummy value\n"},
						{Key: "k2", Val: "plain value"},
					}),
				),
			},
			{
				// Change exec values (change arg)
				Config: testAccLockboxSecretVersionWithCommand(secretName, script, "other value", "just a var"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID),
					testAccCheckYandexLockboxVersionEntries(versionResource, []*lockboxEntryCheck{
						{Key: "k1", Regexp: regexp.MustCompile(`arg: other value, var: just a var, rnd: \d+`)},
						{Key: "k2", Val: "plain value"},
					}),
				),
			},
			{
				// Change exec values (remove args and env)
				Config: testAccLockboxSecretVersionWithCommand(secretName, script, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(versionResource, &versionID),
					testAccCheckYandexLockboxVersionEntries(versionResource, []*lockboxEntryCheck{
						{Key: "k1", Regexp: regexp.MustCompile(`arg: , var: , rnd: \d+`)},
						{Key: "k2", Val: "plain value"},
					}),
				),
			},
		},
	})
}

func testAccLockboxSecretVersionWithCommand(name, cmd, arg, env string) string {
	if arg != "" {
		arg = fmt.Sprintf(`args = ["%s"]`, arg)
	}
	if env != "" {
		env = fmt.Sprintf(`env = { VALUE = "%s" }`, env)
	}
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "exec_secret" {
  name        = "%v"
}

resource "yandex_lockbox_secret_version" "exec_version" {
  secret_id = yandex_lockbox_secret.exec_secret.id
  entries {
    key = "k1"
    command {
      path = "%v"
      %v
      %v
    }
  }
  entries {
    key = "k2"
    text_value = "plain value"
  }
}
`, name, cmd, arg, env)
}

func getYandexDir() string {
	pwd, _ := os.Getwd()
	if strings.HasSuffix(pwd, "yandex") {
		// when running `go test ./yandex ...`, like explained in:
		// https://wiki.yandex-team.ru/cloud/devel/terraform/acceptance-tests/
		return pwd
	}
	return pwd + "/yandex" // when tests are executed from the repo root folder (in builds)
}

func checkTestFilesFolder(t *testing.T) {
	yandexDir := getYandexDir()
	testFixturesDir := yandexDir + "/test-fixtures"
	t.Logf("Checking files in test-fixtures dir: %v", testFixturesDir)
	files, err := os.ReadDir(testFixturesDir)
	if err == nil {
		for _, file := range files {
			t.Logf("- %v", file.Name())
		}
	} else {
		t.Errorf("%v", err)
	}

	script := yandexDir + "/test-fixtures/fake_secret_generator.sh"
	t.Logf("Script file is expected at: %v", script)
}

func createTempFile(t *testing.T, filename, content string) *os.File {
	scriptFile, err := os.CreateTemp("", filename)
	if err != nil {
		t.Error(err)
	}

	_, err = scriptFile.WriteString(content)
	if err != nil {
		t.Error(err)
	}

	err = scriptFile.Chmod(0755)
	if err != nil {
		t.Error(err)
	}

	return scriptFile
}
*/
