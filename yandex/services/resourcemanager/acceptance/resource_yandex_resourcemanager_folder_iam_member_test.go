package acceptance

import (
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	rmfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

// Test that an IAM binding can be applied to a folder
func TestAccFolderIamMember_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var folder resourcemanager.Folder
	cloudID := acceptanceenv.CloudID()
	folderID := acceptanceenv.FolderID()
	userID1 := acceptanceenv.UserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Use an example folder
			{
				Config: testAccFolderIamBasic(folderID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderExists("data.yandex_resourcemanager_folder.acceptance", &folder),
					suite.testAccFolderExistingPolicy(&folder),
				),
			},
			// Apply an IAM binding
			{
				Config: testAccFolderAssociateMemberBasic(cloudID, folderID, userID1),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
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
	suite := testSuite{acceptance.New(t)}
	var folder resourcemanager.Folder
	cloudID := acceptanceenv.CloudID()
	folderID := acceptanceenv.FolderID()
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Use an example folder
			{
				Config: testAccFolderIamBasic(folderID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderExists("data.yandex_resourcemanager_folder.acceptance", &folder),
					suite.testAccFolderExistingPolicy(&folder),
				),
			},
			// Apply an IAM binding
			{
				Config: testAccFolderAssociateMemberBasic(cloudID, folderID, userID1),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
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
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID2,
						}}),
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
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
	suite := testSuite{acceptance.New(t)}
	var folder resourcemanager.Folder
	cloudID := acceptanceenv.CloudID()
	folderID := acceptanceenv.FolderID()
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Use an example folder
			{
				Config: testAccFolderIamBasic(folderID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderExists("data.yandex_resourcemanager_folder.acceptance", &folder),
					suite.testAccFolderExistingPolicy(&folder),
				),
			},
			// Apply multiple IAM bindings
			{
				Config: testAccFolderAssociateMemberMultiple(cloudID, folderID, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID2,
						}}),
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
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
					suite.testAccCheckYandexResourceManagerFolderExists("data.yandex_resourcemanager_folder.acceptance", &folder),
					suite.testAccFolderExistingPolicy(&folder),
				),
			},
		},
	})
}

func testAccFolderAssociateMemberBasic(cloudID, folderID, userID string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
data "yandex_resourcemanager_folder" "acceptance" {
  folder_id = "%s"
}

resource "yandex_resourcemanager_folder_iam_member" "acceptance" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  member    = "userAccount:%s"
  role      = "resource-manager.editor"

  depends_on = [%s]
}
`, folderID, userID, deps)
}

func testAccFolderAssociateMemberMultiple(cloudID, folderID, userID1, userID2 string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID1, userID2)

	return prerequisiteMembership + fmt.Sprintf(`
data "yandex_resourcemanager_folder" "acceptance" {
  folder_id = "%s"
}

resource "yandex_resourcemanager_folder_iam_member" "acceptance" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  member    = "userAccount:%s"
  role      = "resource-manager.editor"

  depends_on = [%s]
}

resource "yandex_resourcemanager_folder_iam_member" "multiple" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  member    = "userAccount:%s"
  role      = "resource-manager.editor"

  depends_on = [%s]
}
`, folderID, userID1, deps, userID2, deps)
}
