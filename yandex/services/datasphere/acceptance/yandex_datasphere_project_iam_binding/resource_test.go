package yandex_datasphere_project_iam_binding_test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	dataspheresdk "github.com/yandex-cloud/go-sdk/services/datasphere/v2"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	fixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datasphere/testfixtures"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.

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
		CheckDestroy:             fixtures.AccCheckProjectDestroy(projectClient),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasphereProjectIamBindingConfig(communityName, projectName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					fixtures.ProjectExists(projectClient, fixtures.ProjectResourceName),
					testAccCheckDatasphereProjectIam(fixtures.ProjectResourceName, role, []string{"system:" + userID}),
				),
			},
			{
				ResourceName:                         "yandex_datasphere_project_iam_binding.test-project-binding",
				ImportStateIdFunc:                    test.ImportIamBindingIdFunc(fixtures.ProjectResourceName, role),
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
		bindings, err := dataspheresdk.NewProjectClient(config.SDKv2).ListAccessBindings(context.Background(), &access.ListAccessBindingsRequest{
			ResourceId: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("get access bindings for resource id(%s):  %w", rs.Primary.ID, err)
		}

		var roleMembers []string
		for _, binding := range bindings.AccessBindings {
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
