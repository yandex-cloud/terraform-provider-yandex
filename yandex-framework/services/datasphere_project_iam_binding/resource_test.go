package datasphere_project_iam_binding_test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	iam_binding "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/datasphere_project_iam_binding"
	//dataspheretest "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/tests/datasphere"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccDatasphereProjectResourceIamBinding(t *testing.T) {
	var (
		communityName = test.ResourceName(63)
		projectName   = test.ResourceName(63)

		userID = "allUsers"
		role   = "datasphere.community-projects.viewer"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasphereProjectIamBindingConfig(communityName, projectName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					test.ProjectExists(test.ProjectResourceName),
					testAccCheckDatasphereProjectIam(test.ProjectResourceName, role, []string{"system:" + userID}),
				),
			},
			{
				ResourceName:                         "yandex_datasphere_project_iam_binding.test-project-binding",
				ImportStateIdFunc:                    test.ImportIamBindingIdFunc(test.ProjectResourceName, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "project_id",
			},
		},
	})
}

func testAccDatasphereProjectIamBindingConfig(communityName, projectName, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_datasphere_community" "test-community" {
  name = "%s"
  billing_account_id = "%s"
  organization_id = "%s"
}

resource "yandex_datasphere_project_iam_binding" "test-project-binding" {
  role = "%s"
  members = ["system:%s"]
  project_id = yandex_datasphere_project.test-project.id
}

resource "yandex_datasphere_project" "test-project" {
  name = "%s"
  community_id = yandex_datasphere_community.test-community.id
}
`, communityName, test.GetBillingAccountId(), test.GetExampleOrganizationID(), role, userID, projectName)
}

func testAccCheckDatasphereProjectIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}
		projectUpdater := iam_binding.ProjectIAMUpdater{
			ProjectId:      rs.Primary.ID,
			ProviderConfig: &config,
		}

		bindings, err := projectUpdater.GeAccessBindings(context.Background(), rs.Primary.ID)
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

		return fmt.Errorf("binding found but expected members is %v, got %v", members, roleMembers)
	}
}
