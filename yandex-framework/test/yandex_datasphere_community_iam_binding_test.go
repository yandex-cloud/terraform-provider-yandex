package test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework"
	yandex_datasphere_community "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/yandex-datasphere/community"
)

func TestAccDatasphereCommunityResourceIamBinding(t *testing.T) {
	var (
		communityName = testResourseName(63)

		userID = "allUsers"
		role   = "datasphere.community-projects.viewer"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckCommunityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCommunityIamBindingConfig(communityName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testCommunityExists(testCommunityResourceName),
					testAccCheckCommunityIam(testCommunityResourceName, role, []string{"system:" + userID}),
				),
			},
			{
				ResourceName:                         "yandex_datasphere_community_iam_binding.test-community-binding",
				ImportStateIdFunc:                    importIamBindingIdFunc(testCommunityResourceName, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "community_id",
			},
		},
	})
}

func testAccCommunityIamBindingConfig(name, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_datasphere_community" "test-community" {
  name = "%s"
  billing_account_id = "%s"
  organization_id = "%s"
}

resource "yandex_datasphere_community_iam_binding" "test-community-binding" {
  role = "%s"
  members = ["system:%s"]
  community_id = yandex_datasphere_community.test-community.id
}
`, name, getBillingAccountId(), getExampleOrganizationID(), role, userID)
}

func testAccCheckCommunityIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}
		communityUpdater := yandex_datasphere_community.CommunityIAMUpdater{
			CommunityId:    rs.Primary.ID,
			ProviderConfig: &config,
		}

		bindings, err := communityUpdater.GeAccessBindings(context.Background(), rs.Primary.ID)
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
