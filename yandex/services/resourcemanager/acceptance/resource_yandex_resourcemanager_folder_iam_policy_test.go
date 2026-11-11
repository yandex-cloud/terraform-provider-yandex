package acceptance

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	rmfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	svc_resource_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/pluginsdk"
)

// revive:disable:var-naming
func TestAccFolderIamPolicy_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudID := acceptanceenv.CloudID()
	folderID := acceptanceenv.FolderID()
	userID1 := acceptanceenv.UserID2()

	policy := &sdkcommon.Policy{
		Bindings: []*access.AccessBinding{
			{
				RoleId: "resource-manager.viewer",
				Subject: &access.Subject{
					Type: "userAccount",
					Id:   userID1,
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckYandexResourceManagerFolderIamPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFolderIamPolicy_basic(cloudID, folderID, policy),
				Check:  suite.testAccCheckYandexResourceManagerFolderIamPolicy("yandex_resourcemanager_folder_iam_policy.test", policy),
			},
		},
	})
}

func TestAccFolderIamPolicy_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudID := acceptanceenv.CloudID()
	folderID := acceptanceenv.FolderID()
	userID1 := acceptanceenv.UserID1()

	policy1 := &sdkcommon.Policy{
		Bindings: []*access.AccessBinding{
			{
				RoleId: "resource-manager.viewer",
				Subject: &access.Subject{
					Type: "userAccount",
					Id:   userID1,
				},
			},
		},
	}
	policy2 := &sdkcommon.Policy{
		Bindings: []*access.AccessBinding{
			{
				RoleId: "resource-manager.editor",
				Subject: &access.Subject{
					Type: "userAccount",
					Id:   userID1,
				},
			},
			{
				RoleId: "resource-manager.viewer",
				Subject: &access.Subject{
					Type: "userAccount",
					Id:   userID1,
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckYandexResourceManagerFolderIamPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFolderIamPolicy_basic(cloudID, folderID, policy1),
				Check:  suite.testAccCheckYandexResourceManagerFolderIamPolicy("yandex_resourcemanager_folder_iam_policy.test", policy1),
			},
			{
				Config: testAccFolderIamPolicy_basic(cloudID, folderID, policy2),
				Check:  suite.testAccCheckYandexResourceManagerFolderIamPolicy("yandex_resourcemanager_folder_iam_policy.test", policy2),
			},
		},
	})
}

func (suite testSuite) testAccCheckYandexResourceManagerFolderIamPolicyDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_resourcemanager_folder_iam_policy" {
			continue
		}

		folderID := rs.Primary.Attributes["folder_id"]

		bindings, err := svc_resource_manager.GetFolderAccessBindings(config.Context(), config, folderID)
		if err != nil && bindings != nil && len(bindings) > 0 {
			return fmt.Errorf("Folder '%s' policy hasn't been deleted", folderID)
		}
	}
	return nil
}

func (suite testSuite) testAccCheckYandexResourceManagerFolderIamPolicy(n string, policy *sdkcommon.Policy) resource.TestCheckFunc {
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

		bindings, err := svc_resource_manager.GetFolderAccessBindings(config.Context(), config, rs.Primary.ID)
		if err != nil {
			return err
		}

		sort.Sort(sdkcommon.SortableBindings(bindings))
		sort.Sort(sdkcommon.SortableBindings(policy.Bindings))
		if !reflect.DeepEqual(bindings, policy.Bindings) {
			return fmt.Errorf("Incorrect IAM policy bindings. Expected '%s', got '%s'", policy.Bindings, bindings)
		}

		return nil
	}
}

func testAccFolderIamPolicy_basic(cloudID, folderID string, policy *sdkcommon.Policy) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userAccountIDs(policy)...)

	var bindingBuffer bytes.Buffer
	rolesMap := sdkcommon.RolesToMembersMap(policy.Bindings)

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
