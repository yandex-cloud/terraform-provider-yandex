package yandex

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/vault/helper/pgpkeys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

func CreateResourceData(t *testing.T, schemaObject map[string]*schema.Schema, rawInitialState map[string]interface{},
	diffAttributes map[string]*terraform.ResourceAttrDiff) *schema.ResourceData {

	t.Helper()
	ctx := context.Background()
	internalMap := schema.InternalMap(schemaObject)

	emptyState := terraform.NewInstanceStateShimmedFromValue(cty.ObjectVal(map[string]cty.Value{}), 1)
	initialDiff, err := internalMap.Diff(ctx, emptyState.DeepCopy(), terraform.NewResourceConfigRaw(rawInitialState), nil, nil, true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	targetDiff := &terraform.InstanceDiff{Attributes: diffAttributes}

	resourceData, err := internalMap.Data(emptyState.MergeDiff(initialDiff), targetDiff)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return resourceData
}

func TestJoinedStrings(t *testing.T) {
	testKeys := []string{"key1", "key2", "key3"}
	joinedKeys := getJoinedKeys(testKeys)
	assert.Equal(t, "`key1`, `key2`, `key3`", joinedKeys)

	testKey := []string{"key1"}
	joinedKey := getJoinedKeys(testKey)
	assert.Equal(t, "`key1`", joinedKey)
}

func memberType(ab *access.AccessBinding) string {
	return ab.Subject.Type
}

func memberID(ab *access.AccessBinding) string {
	return ab.Subject.Id
}

func userAccountIDs(p *Policy) []string {
	usersMap := map[string]bool{}
	for _, b := range p.Bindings {
		if memberType(b) == "userAccount" {
			usersMap[memberID(b)] = true
		}
	}
	userIDs := []string{}
	for userID := range usersMap {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

func testAccCloudAssignCloudMemberRole(cloudID string, usersID ...string) (string, string) {
	var config bytes.Buffer
	var resourceRefs []string

	for _, userID := range usersID {
		resType := "yandex_resourcemanager_cloud_iam_member"
		resName := fmt.Sprintf("membership-%s-%s", cloudID, userID)

		config.WriteString(fmt.Sprintf(`
// Make user member of cloud to allow assign another roles
resource "%s" "%s" {
  cloud_id = "%s"
  role     = "resource-manager.clouds.member"
  member   = "userAccount:%s"
}
`, resType, resName, cloudID, userID))

		resourceRefs = append(resourceRefs, fmt.Sprintf("\"%s.%s\"", resType, resName))
	}

	return config.String(), strings.Join(resourceRefs, ",")
}

func getFolderIamPolicyByFolderID(folderID string, config *Config) (*Policy, error) {
	f := FolderIamUpdater{
		folderID: folderID,
		Config:   config,
	}
	return f.GetResourceIamPolicy()
}

func checkWithState(fn func() resource.TestCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return fn()(s)
	}
}

func testAccCheckFunctionIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		bindings, err := getFunctionAccessBindings(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func testAccCheckServiceAccountIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		bindings, err := getServiceAccountAccessBindings(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func testAccCheckCreatedAtAttr(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const createdAtAttrName = "created_at"
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		createdAt, ok := rs.Primary.Attributes[createdAtAttrName]
		if !ok {
			return fmt.Errorf("can't find '%s' attr for %s resource", createdAtAttrName, resourceName)
		}

		if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
			return fmt.Errorf("can't parse timestamp in attr '%s': %s", createdAtAttrName, createdAt)
		}
		return nil
	}
}

func testAccCheckResourceIDField(resourceName string, idFieldName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if rs.Primary.Attributes[idFieldName] != rs.Primary.ID {
			return fmt.Errorf("Resource: %s id field: %s, doesn't match resource ID", resourceName, idFieldName)
		}

		return nil
	}
}
func testExistsElementWithAttrTrimmedValue(resourceName, path, field, value string, fullPath *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", resourceName, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", resourceName, ms.Path)
		}

		reStr := fmt.Sprintf(`(%s\.\d+)\.%s`, path, field)
		re, err := regexp.Compile(reStr)
		if err != nil {
			return err
		}
		for k, v := range is.Attributes {
			trimmedValue := strings.TrimSpace(v)
			if re.MatchString(k) && trimmedValue == value {
				sm := re.FindStringSubmatch(k)
				*fullPath = sm[1]
				return nil
			}
		}

		return fmt.Errorf(
			"Can't find key %s.*.%s in resource: %s with value %s", path, field, resourceName, value,
		)
	}
}

func testExistsElementWithAttrValue(resourceName, path, field, value string, fullPath *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", resourceName, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", resourceName, ms.Path)
		}
		reStr := fmt.Sprintf(`(%s\.\d+)\.%s`, path, field)
		re := regexp.MustCompile(reStr)

		for k, v := range is.Attributes {
			if re.MatchString(k) && v == value {
				sm := re.FindStringSubmatch(k)
				*fullPath = sm[1]
				return nil
			}
		}

		return fmt.Errorf(
			"Can't find key %s.*.%s in resource: %s with value %s", path, field, resourceName, value,
		)
	}
}

func testExistsFirstElementWithAttr(resourceName, path, field string, fullPath *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", resourceName, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", resourceName, ms.Path)
		}

		for k := range is.Attributes {
			reStr := fmt.Sprintf(`(%s\.\d+)\.%s`, path, field)
			re := regexp.MustCompile(reStr)
			if re.MatchString(k) {
				sm := re.FindStringSubmatch(k)
				*fullPath = sm[1]
				return nil
			}
		}

		return fmt.Errorf(
			"Can't find key %s.*.%s in resource: %s", path, field, resourceName,
		)
	}
}

func testCheckResourceSubAttr(resourceName string, path *string, field string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", resourceName, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", resourceName, ms.Path)
		}

		fullPath := fmt.Sprintf("%s.%s", *path, field)
		actualValue, ok := is.Attributes[fullPath]
		if !ok {
			return fmt.Errorf("Can't find path %s in resource: %s", fullPath, resourceName)
		}

		if actualValue != value {
			return fmt.Errorf(
				"Can't match values for path %s in resource: %s. %s != %s", fullPath, resourceName, value, actualValue,
			)
		}

		return nil
	}
}

func testCheckResourceSubAttrFn(resourceName string, path *string, field string, checkfn func(string) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", resourceName, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", resourceName, ms.Path)
		}

		fullPath := fmt.Sprintf("%s.%s", *path, field)
		value, ok := is.Attributes[fullPath]
		if !ok {
			return fmt.Errorf("Can't find path %s in resource: %s", fullPath, resourceName)
		}

		err := checkfn(value)
		if err != nil {
			return err
		}

		return nil
	}
}

func testDecryptKeyAndTest(name, key, pgpKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		ciphertext, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("can't find '%s' attr for %s resource", key, name)
		}

		// We can't verify that the decrypted ciphertext is correct, because we don't
		// have it. We can verify that decrypting it does not error
		_, err := pgpkeys.DecryptBytes(ciphertext, pgpKey)
		if err != nil {
			return fmt.Errorf("error decrypting ciphertext: %s", err)
		}

		return nil
	}
}

func DefaultAndEmptyFolderProviders() []map[string]*schema.Provider {
	return []map[string]*schema.Provider{
		testAccProviders,
		testAccProviderEmptyFolder,
	}
}

func CustomProvidersTest(t *testing.T, providers []map[string]*schema.Provider, testCase resource.TestCase) {
	for _, provider := range providers {
		customTest := testCase
		customTest.Providers = provider
		resource.Test(t, customTest)
	}
}

// Helpers for tests.
// Config for generated access key and secret key.
func testAccCommonIamDependenciesEditorConfig(randInt int) string {
	return testAccCommonIamDependenciesConfigImpl(randInt, "editor")
}

func testAccCommonIamDependenciesAdminConfig(randInt int) string {
	return testAccCommonIamDependenciesConfigImpl(randInt, "admin")
}

func testAccCommonIamDependenciesConfigImpl(randInt int, role string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "sa" {
	name = "test-sa-for-tf-test-%[1]d"
}

resource "yandex_resourcemanager_folder_iam_member" "binding" {
	folder_id   = "%[3]s"
	member      = "serviceAccount:${yandex_iam_service_account.sa.id}"
	role        = "%[2]s"
	sleep_after = 30
}

resource "yandex_iam_service_account_static_access_key" "sa-key" {
	service_account_id = "${yandex_iam_service_account.sa.id}"

	depends_on = [
		yandex_resourcemanager_folder_iam_member.binding
	]
}
`, randInt, role, getExampleFolderID())
}

func TestSortInterfaceListByTemplate(t *testing.T) {

	name := "some_key"

	listToSort := []interface{}{
		map[string]interface{}{name: "a"},
		map[string]interface{}{name: "d"},
		map[string]interface{}{name: "b"},
		map[string]interface{}{name: "c"},
		map[string]interface{}{name: "h"},
	}
	templateList := []interface{}{
		map[string]interface{}{name: "b"},
		map[string]interface{}{name: "c"},
		map[string]interface{}{name: "d"},
		map[string]interface{}{name: "e"},
	}

	checkList := []string{"b", "c", "d", "a", "h"}

	sortInterfaceListByTemplate(listToSort, templateList, name)

	for i, v := range checkList {
		if getField(listToSort[i], name) != v {
			t.Errorf("sortInterfaceListByTemplate: after sort %v value should be \"%v\" but value is \"%v\"", i, v, getField(listToSort[i], name))
		}
	}
}

func TestSortInterfaceListByTemplateNoIntersection(t *testing.T) {

	name := "some_key"

	listToSort := []interface{}{
		map[string]interface{}{name: "a"},
		map[string]interface{}{name: "d"},
		map[string]interface{}{name: "b"},
		map[string]interface{}{name: "c"},
		map[string]interface{}{name: "h"},
	}
	templateList := []interface{}{
		map[string]interface{}{name: "m"},
		map[string]interface{}{name: "n"},
		map[string]interface{}{name: "o"},
		map[string]interface{}{name: "p"},
	}

	checkList := []string{"a", "b", "c", "d", "h"}

	sortInterfaceListByTemplate(listToSort, templateList, name)

	for i, v := range checkList {
		if getField(listToSort[i], name) != v {
			t.Errorf("sortInterfaceListByTemplate: after sort %v value should be \"%v\" but value is \"%v\"", i, v, getField(listToSort[i], name))
		}
	}
}

func TestSortInterfaceListByTemplateEmptyTemplate(t *testing.T) {

	name := "some_key"

	listToSort := []interface{}{
		map[string]interface{}{name: "a"},
		map[string]interface{}{name: "d"},
		map[string]interface{}{name: "b"},
		map[string]interface{}{name: "c"},
		map[string]interface{}{name: "h"},
	}
	templateList := []interface{}{}

	// NO sorting
	checkList := []string{"a", "d", "b", "c", "h"}

	sortInterfaceListByTemplate(listToSort, templateList, name)

	for i, v := range checkList {
		if getField(listToSort[i], name) != v {
			t.Errorf("sortInterfaceListByTemplate: after sort %v value should be \"%v\" but value is \"%v\"", i, v, getField(listToSort[i], name))
		}
	}
}

func getRandAccTestResourceName() string {
	return fmt.Sprintf("tf-test-%s", acctest.RandString(10))
}

func testAccCheckResourceAttrWithValueFactory(name, key string, valueFactory func() string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't get resource '%s'", name)
		}

		instanceState := resourceState.Primary
		if instanceState == nil {
			return fmt.Errorf("there is no primary instance state for '%s'", name)
		}

		expected := valueFactory()
		if expected == "0" && (strings.HasSuffix(key, ".#") || strings.HasSuffix(key, ".%")) {
			return fmt.Errorf("testAccCheckResourceAttrWithValueFactory does not know how to perform empty check =(")
		}

		actual, ok := instanceState.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: there is no '%s' attribute", name, key)
		}

		if expected != actual {
			return fmt.Errorf("%s: expected attribute '%s' to have value '%s', got '%s'", name, key, expected, actual)
		}

		return nil
	}
}

func TestParseDuration(t *testing.T) {
	d := (&schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}).TestResourceData()
	i := d.Get("name")
	r, err := parseDuration(i.(string))
	require.NoError(t, err)
	require.Nil(t, r)
}
