package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	rmfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	svc_resource_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/pluginsdk"
)

// Test that an IAM binding can be applied to a folder
func TestAccFolderIamBinding_basic(t *testing.T) {
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
				Config: testAccFolderAssociateBindingBasic(cloudID, folderID, userID1),
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

// Test that multiple IAM bindings can be applied to a folder, one at a time
func TestAccFolderIamBinding_multiple(t *testing.T) {
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
				Config: testAccFolderAssociateBindingBasic(cloudID, folderID, userID1),
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
				Config: testAccFolderAssociateBindingMultiple(cloudID, folderID, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.viewer",
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

// Test that multiple IAM bindings can be applied to a folder all at once
func TestAccFolderIamBinding_multipleAtOnce(t *testing.T) {
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
				Config: testAccFolderAssociateBindingMultiple(cloudID, folderID, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID1,
						}}),
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.viewer",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID2,
						}}),
				),
			},
		},
	})
}

// Test that an IAM binding can be updated once applied to a folder
func TestAccFolderIamBinding_update(t *testing.T) {
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
				Config: testAccFolderAssociateBindingBasic(cloudID, folderID, userID1),
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
			// Apply an updated IAM binding
			{
				Config: testAccFolderAssociateBindingUpdated(cloudID, folderID, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID1,
						},
					}),
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID2,
						},
					}),
				),
			},
			// Drop the original member
			{
				Config: testAccFolderAssociateBindingDropMemberFromBasic(cloudID, folderID, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID2,
						},
					}),
				),
			},
		},
	})
}

// Test that an IAM binding can be removed from a folder
func TestAccFolderIamBinding_remove(t *testing.T) {
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
				Config: testAccFolderAssociateBindingMultiple(cloudID, folderID, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.viewer",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID2,
						},
					}),
					suite.testAccCheckYandexResourceManagerFolderIamBindingExists(&folder, &access.AccessBinding{
						RoleId: "resource-manager.editor",
						Subject: &access.Subject{
							Type: "userAccount",
							Id:   userID1,
						},
					}),
				),
			},
			// Remove the bindings
			{
				Config: testAccFolderIamBasic(folderID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccFolderExistingPolicy(&folder),
				),
			},
		},
	})
}

func (suite testSuite) testAccCheckYandexResourceManagerFolderIamBindingExists(folder *resourcemanager.Folder, expected *access.AccessBinding) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		folderPolicy, err := getFolderIamPolicyByFolderID(folder.Id, config)
		if err != nil {
			return fmt.Errorf("Failed to retrieve IAM policy for folder %q: %s", folder.Id, err)
		}

		// TODO: Check count of members

		if checkBindingInPolicy(folderPolicy, expected) {
			return nil
		}

		return fmt.Errorf("Expected access binding for role %q of folder %q to be %v, got %v", expected.RoleId, folder.Id, expected, folderPolicy)
	}
}

// Confirm that a folder has an IAM policy with at least 1 binding
func (suite testSuite) testAccFolderExistingPolicy(folder *resourcemanager.Folder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		if _, err := svc_resource_manager.GetFolderAccessBindings(context.Background(), c, folder.Id); err != nil {
			return err
		}

		return nil
	}
}

func checkBindingInPolicy(policy *sdkcommon.Policy, tBinding *access.AccessBinding) bool {
	for _, cBinding := range policy.Bindings {
		if cBinding.RoleId != tBinding.RoleId {
			continue
		} else {
			if cBinding.Subject.Id == tBinding.Subject.Id && cBinding.Subject.Type == tBinding.Subject.Type {
				return true
			}
		}
	}
	return false
}

func (suite testSuite) testAccCheckYandexResourceManagerFolderExists(n string, folder *resourcemanager.Folder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := resourcemanagersdk.NewFolderClient(config.SDK)

		found, err := client.Get(context.Background(), &resourcemanager.GetFolderRequest{
			FolderId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Folder not found")
		}

		*folder = *found

		return nil
	}
}

func testAccFolderIamBasic(folderID string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_folder" "acceptance" {
  folder_id = "%s"
}
`, folderID)
}

func testAccFolderAssociateBindingBasic(cloudID, folderID, userID string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
data "yandex_resourcemanager_folder" "acceptance" {
  folder_id = "%s"
}

resource "yandex_resourcemanager_folder_iam_binding" "acceptance" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  members   = ["userAccount:%s"]
  role      = "resource-manager.editor"

  depends_on = [%s]
}
`, folderID, userID, deps)
}

func testAccFolderAssociateBindingMultiple(cloudID, folderID, userID1, userID2 string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID1, userID2)
	return prerequisiteMembership + fmt.Sprintf(`
data "yandex_resourcemanager_folder" "acceptance" {
  folder_id = "%s"
}

resource "yandex_resourcemanager_folder_iam_binding" "acceptance" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  members   = ["userAccount:%s"]
  role      = "resource-manager.editor"

  depends_on = [%s]
}

resource "yandex_resourcemanager_folder_iam_binding" "multiple" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  members   = ["userAccount:%s"]
  role      = "resource-manager.viewer"

  depends_on = [%s]
}
`, folderID, userID1, deps, userID2, deps)
}

func testAccFolderAssociateBindingUpdated(cloudID, folderID, userID1, userID2 string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID1, userID2)
	return prerequisiteMembership + fmt.Sprintf(`
data "yandex_resourcemanager_folder" "acceptance" {
  folder_id = "%s"
}

resource "yandex_resourcemanager_folder_iam_binding" "acceptance" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  members   = ["userAccount:%s", "userAccount:%s"]
  role      = "resource-manager.editor"

  depends_on = [%s]
}
`, folderID, userID1, userID2, deps)
}

func testAccFolderAssociateBindingDropMemberFromBasic(cloudID, folderID, userID string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
data "yandex_resourcemanager_folder" "acceptance" {
  folder_id = "%s"
}

resource "yandex_resourcemanager_folder_iam_binding" "dropped" {
  folder_id = "${data.yandex_resourcemanager_folder.acceptance.id}"
  members   = ["userAccount:%s"]
  role      = "resource-manager.editor"

  depends_on = [%s]
}
`, folderID, userID, deps)
}
