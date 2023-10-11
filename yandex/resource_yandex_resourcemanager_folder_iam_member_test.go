package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
)

// Test that an IAM binding can be applied to a folder
func TestAccFolderIamMember_basic(t *testing.T) {
	var folder resourcemanager.Folder
	cloudID := getExampleCloudID()
	folderID := getExampleFolderID()
	userID1 := getExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Use an example folder
			{
				Config: testAccFolderIamBasic(folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexResourceManagerFolderExists("data.yandex_resourcemanager_folder.acceptance", &folder),
					testAccFolderExistingPolicy(&folder),
				),
			},
			// Apply an IAM binding
			{
				Config: testAccFolderAssociateMemberBasic(cloudID, folderID, userID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "admin",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID1,
						},
					}),
				),
			},
		},
	})
}

// Test that multiple IAM bindings can be applied to a folder
func TestAccFolderIamMember_multiple(t *testing.T) {
	var folder resourcemanager.Folder
	cloudID := getExampleCloudID()
	folderID := getExampleFolderID()
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Use an example folder
			{
				Config: testAccFolderIamBasic(folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexResourceManagerFolderExists("data.yandex_resourcemanager_folder.acceptance", &folder),
					testAccFolderExistingPolicy(&folder),
				),
			},
			// Apply an IAM binding
			{
				Config: testAccFolderAssociateMemberBasic(cloudID, folderID, userID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "admin",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID1,
						},
					}),
				),
			},
			// Apply another IAM binding
			{
				Config: testAccFolderAssociateMemberMultiple(cloudID, folderID, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "admin",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID2,
						}}),
					testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "admin",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID1,
						}}),
				),
			},
		},
	})
}

// Test that an IAM binding can be removed from a folder
func TestAccFolderIamMember_remove(t *testing.T) {
	var folder resourcemanager.Folder
	cloudID := getExampleCloudID()
	folderID := getExampleFolderID()
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Use an example folder
			{
				Config: testAccFolderIamBasic(folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexResourceManagerFolderExists("data.yandex_resourcemanager_folder.acceptance", &folder),
					testAccFolderExistingPolicy(&folder),
				),
			},
			// Apply multiple IAM bindings
			{
				Config: testAccFolderAssociateMemberMultiple(cloudID, folderID, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "admin",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID2,
						}}),
					testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "admin",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID1,
						}}),
				),
			},
			// Remove the bindings
			{
				Config: testAccFolderIamBasic(folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexResourceManagerFolderExists("data.yandex_resourcemanager_folder.acceptance", &folder),
					testAccFolderExistingPolicy(&folder),
				),
			},
		},
	})
}

func testAccFolderAssociateMemberBasic(cloudID, folderID, userID string) string {
	prerequisiteMembership, deps := testAccCloudAssignCloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
data "yandex_resourcemanager_folder" "acceptance" {
  folder_id = "%s"
}

resource "yandex_resourcemanager_folder_iam_member" "acceptance" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  member    = "userAccount:%s"
  role      = "admin"

  depends_on = [%s]
}
`, folderID, userID, deps)
}

func testAccFolderAssociateMemberMultiple(cloudID, folderID, userID1, userID2 string) string {
	prerequisiteMembership, deps := testAccCloudAssignCloudMemberRole(cloudID, userID1, userID2)

	return prerequisiteMembership + fmt.Sprintf(`
data "yandex_resourcemanager_folder" "acceptance" {
  folder_id = "%s"
}

resource "yandex_resourcemanager_folder_iam_member" "acceptance" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  member    = "userAccount:%s"
  role      = "admin"

  depends_on = [%s]
}

resource "yandex_resourcemanager_folder_iam_member" "multiple" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  member    = "userAccount:%s"
  role      = "admin"

  depends_on = [%s]
}
`, folderID, userID1, deps, userID2, deps)
}
