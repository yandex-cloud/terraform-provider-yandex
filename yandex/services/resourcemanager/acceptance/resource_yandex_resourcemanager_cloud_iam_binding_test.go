package acceptance

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	rmfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_resource_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/pluginsdk"
)

// Test that an IAM binding can be applied to a cloud
func TestAccCloudIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudID := acceptanceenv.CloudID()
	role := "viewer"
	userID := acceptanceenv.UserID2()

	t.Log(testAccCloudAssociateBindingBasic(cloudID, role, userID))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccCloudAssociateBindingBasic(cloudID, role, userID),
				Check: suite.testAccCheckCloudIam(
					cloudID, role, []string{"userAccount:" + userID}),
			},
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.acceptance", cloudID, role),
		},
	})
}

// Test that multiple IAM bindings can be applied to a cloud, one at a time
func TestAccCloudIamBinding_multiple(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudID := acceptanceenv.CloudID()
	role1 := "editor"
	role2 := "viewer"
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccCloudAssociateBindingBasic(cloudID, role1, userID1),
				Check:  suite.testAccCheckCloudIam(cloudID, role1, []string{"userAccount:" + userID1}),
			},
			// Apply another IAM binding
			{
				Config: testAccCloudAssociateBindingMultiple(cloudID, role1, role2, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckCloudIam(cloudID, role1,
						[]string{"userAccount:" + userID1, "userAccount:" + userID2}),
					suite.testAccCheckCloudIam(cloudID, role2,
						[]string{"userAccount:" + userID1, "userAccount:" + userID2}),
				),
			},
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.acceptance", cloudID, role1),
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.multiple", cloudID, role2),
		},
	})
}

// Test that multiple IAM bindings can be applied to a cloud all at once
func TestAccCloudIamBinding_multipleAtOnce(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudID := acceptanceenv.CloudID()
	role := "editor"
	role2 := "viewer"
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare data source about cloud ID
			{
				Config: rmfixtures.CloudDataSourceByID(cloudID),
			},
			// Apply an IAM binding
			{
				Config: testAccCloudAssociateBindingMultiple(cloudID, role, role2, userID1, userID2),
			},
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.acceptance", cloudID, role),
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.multiple", cloudID, role2),
		},
	})
}

// Test that an IAM binding can be updated once applied to a cloud
func TestAccCloudIamBinding_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudID := acceptanceenv.CloudID()
	role := "editor"
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccCloudAssociateBindingBasic(cloudID, role, userID1),
			},
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.acceptance", cloudID, role),
			// Apply an updated IAM binding
			{
				Config: testAccCloudAssociateBindingUpdated(cloudID, role, userID1, userID2),
			},
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.acceptance", cloudID, role),
			// Drop the original member
			{
				Config: testAccCloudAssociateBindingDropMemberFromBasic(cloudID, role, userID1),
			},
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.acceptance", cloudID, role),
		},
	})
}

// Test that an IAM binding can be removed from a cloud
func TestAccCloudIamBinding_remove(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudID := acceptanceenv.CloudID()
	role1 := "editor"
	role2 := "viewer"
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare data source about cloud ID
			{
				Config: rmfixtures.CloudDataSourceByID(cloudID),
			},
			// Apply multiple IAM bindings
			{
				Config: testAccCloudAssociateBindingMultiple(cloudID, role1, role2, userID1, userID2),
			},
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.acceptance", cloudID, role1),
			cloudIamBindingImportStep("yandex_resourcemanager_cloud_iam_binding.multiple", cloudID, role2),
			// Remove the bindings
			{
				Config: rmfixtures.CloudDataSourceByID(cloudID),
			},
		},
	})
}

func (suite testSuite) testAccCheckCloudIam(cloudID, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		bindings, err := svc_resource_manager.GetCloudAccessBindings(config.Context(), config, cloudID)
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

func testAccCloudAssociateBindingUpdated(cloudID, role, userID1, userID2 string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID1, userID2)
	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_resourcemanager_cloud_iam_binding" "acceptance" {
  cloud_id = "%s"
  role     = "%s"
  members  = ["userAccount:%s", "userAccount:%s"]

  depends_on = [%s]
}
`, cloudID, role, userID1, userID2, deps)
}

func testAccCloudAssociateBindingMultiple(cloudID, role1, role2, userID1, userID2 string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID1, userID2)

	multiple1 := fmt.Sprintf(`
resource "yandex_resourcemanager_cloud_iam_binding" "acceptance" {
  cloud_id = "%s"
  role     = "%s"
  members  = ["userAccount:%s", "userAccount:%s"]

  depends_on = [%s]
}
`, cloudID, role1, userID1, userID2, deps)

	multiple2 := fmt.Sprintf(`
resource "yandex_resourcemanager_cloud_iam_binding" "multiple" {
  cloud_id = "%s"
  role     = "%s"
  members  = ["userAccount:%s", "userAccount:%s"]

  depends_on = [%s]
}
`, cloudID, role2, userID1, userID2, deps)

	return prerequisiteMembership + multiple1 + multiple2
}

func cloudIamBindingImportStep(resourceName, cloudID, role string) resource.TestStep {
	return resource.TestStep{
		ResourceName:                         resourceName,
		ImportStateId:                        fmt.Sprintf("%s,%s", cloudID, role),
		ImportState:                          true,
		ImportStateVerifyIdentifierAttribute: "cloud_id",
	}
}

func testAccCloudAssociateBindingBasic(cloudID, role, userID string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_resourcemanager_cloud_iam_binding" "acceptance" {
  cloud_id = "%s"
  role     = "%s"
  members  = ["userAccount:%s"]

  depends_on = [%s]
}
`, cloudID, role, userID, deps)
}

func testAccCloudAssociateBindingDropMemberFromBasic(cloudID, role, userID string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_resourcemanager_cloud_iam_binding" "acceptance" {
  cloud_id = "%s"
  role     = "%s"
  members  = ["userAccount:%s"]

  depends_on = [%s]
}
`, cloudID, role, userID, deps)
}
