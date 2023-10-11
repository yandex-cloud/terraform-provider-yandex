package test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework"
	yandex_datasphere_project "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/yandex-datasphere/project"
	"reflect"
	"sort"
	"testing"
)

func TestAccDatasphereProjectResourceIamBinding(t *testing.T) {
	communityName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	projectName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)

	userID := "allUsers"
	role := "datasphere.community-projects.viewer"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasphereProjectIamBindingConfig(communityName, projectName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testDatasphereProjectExists(testProjectResourceName),
					testAccCheckDatasphereProjectIam(testProjectResourceName, role, []string{"system:" + userID}),
				),
			},
			{
				ResourceName:                         "yandex_datasphere_project_iam_binding.test-project-binding",
				ImportStateIdFunc:                    importIamBindingIdFunc(testProjectResourceName, role),
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
`, communityName, getBillingAccountId(), getExampleOrganizationID(), role, userID, projectName)
}

func testAccCheckDatasphereProjectIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}
		projectUpdater := yandex_datasphere_project.ProjectIAMUpdater{
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
