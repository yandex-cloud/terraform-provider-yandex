package yandex

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

// Acceptance test functions for one type of resource share common prefix with underscore as separator.
// This allows you to run all tests for one specific resource type like:
// $ make testacc TESTARGS='-run=TestAccFolderIamPolicy'
//
// Disable revive due to "don't use underscores in Go names" rule: eg func testAccFolderIamPolicy_basic should be testAccFolderIamPolicyBasic

// revive:disable:var-naming
func TestAccFolderIamPolicy_basic(t *testing.T) {
	cloudID := getExampleCloudID()
	folderID := getExampleFolderID()
	userID1 := getExampleUserID2()

	policy := &Policy{
		Bindings: []*access.AccessBinding{
			{
				RoleId: "viewer",
				Subject: &access.Subject{
					Type: "userAccount",
					Id:   userID1,
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckYandexResourceManagerFolderIamPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFolderIamPolicy_basic(cloudID, folderID, policy),
				Check:  testAccCheckYandexResourceManagerFolderIamPolicy("yandex_resourcemanager_folder_iam_policy.test", policy),
			},
		},
	})
}

func TestAccFolderIamPolicy_update(t *testing.T) {
	cloudID := getExampleCloudID()
	folderID := getExampleFolderID()
	userID1 := getExampleUserID1()

	policy1 := &Policy{
		Bindings: []*access.AccessBinding{
			{
				RoleId: "viewer",
				Subject: &access.Subject{
					Type: "userAccount",
					Id:   userID1,
				},
			},
		},
	}
	policy2 := &Policy{
		Bindings: []*access.AccessBinding{
			{
				RoleId: "editor",
				Subject: &access.Subject{
					Type: "userAccount",
					Id:   userID1,
				},
			},
			{
				RoleId: "viewer",
				Subject: &access.Subject{
					Type: "userAccount",
					Id:   userID1,
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckYandexResourceManagerFolderIamPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFolderIamPolicy_basic(cloudID, folderID, policy1),
				Check:  testAccCheckYandexResourceManagerFolderIamPolicy("yandex_resourcemanager_folder_iam_policy.test", policy1),
			},
			{
				Config: testAccFolderIamPolicy_basic(cloudID, folderID, policy2),
				Check:  testAccCheckYandexResourceManagerFolderIamPolicy("yandex_resourcemanager_folder_iam_policy.test", policy2),
			},
		},
	})
}

func testAccCheckYandexResourceManagerFolderIamPolicyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_resourcemanager_folder_iam_policy" {
			continue
		}

		folderID := rs.Primary.Attributes["folder_id"]

		bindings, err := getFolderAccessBindings(config.Context(), config, folderID)
		if err != nil && bindings != nil && len(bindings) > 0 {
			return fmt.Errorf("Folder '%s' policy hasn't been deleted", folderID)
		}
	}
	return nil
}

func testAccCheckYandexResourceManagerFolderIamPolicy(n string, policy *Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		bindings, err := getFolderAccessBindings(config.Context(), config, rs.Primary.ID)
		if err != nil {
			return err
		}

		sort.Sort(sortableBindings(bindings))
		sort.Sort(sortableBindings(policy.Bindings))
		if !reflect.DeepEqual(bindings, policy.Bindings) {
			return fmt.Errorf("Incorrect IAM policy bindings. Expected '%s', got '%s'", policy.Bindings, bindings)
		}

		return nil
	}
}

func testAccFolderIamPolicy_basic(cloudID, folderID string, policy *Policy) string {
	prerequisiteMembership, deps := testAccCloudAssignCloudMemberRole(cloudID, userAccountIDs(policy)...)

	var bindingBuffer bytes.Buffer
	rolesMap := rolesToMembersMap(policy.Bindings)

	for role, members := range rolesMap {
		bindingBuffer.WriteString("binding {\n")
		bindingBuffer.WriteString(fmt.Sprintf("role = \"%s\"\n", role))
		bindingBuffer.WriteString("members = [\n")
		for m := range members {
			bindingBuffer.WriteString(fmt.Sprintf("\"%s\",\n", m))
		}
		bindingBuffer.WriteString("]\n}\n")
	}

	return prerequisiteMembership + fmt.Sprintf(`
data "yandex_resourcemanager_folder" "permissiontest" {
  folder_id = "%s"
}

data "yandex_iam_policy" "test" {
  %s
}

resource "yandex_resourcemanager_folder_iam_policy" "test" {
  folder_id   = "${data.yandex_resourcemanager_folder.permissiontest.id}"
  policy_data = "${data.yandex_iam_policy.test.policy_data}"

  depends_on = [%s]
}
`, folderID, bindingBuffer.String(), deps)
}
