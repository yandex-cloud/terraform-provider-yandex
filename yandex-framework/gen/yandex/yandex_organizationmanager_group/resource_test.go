package yandex_organizationmanager_group_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagerv1sdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"testing"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// The config here should match as closely as possible to the one presented to the user in the docs.
// Serves as a proof that the example config is viable.
func TestAccOrganizationManagerGroup_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()

	config := fmt.Sprintf(`
resource "yandex_organizationmanager_group" group {
  name            = "my-group"
  description     = "My new Group"
  organization_id = "%s"
}
`, test.GetExampleOrganizationID())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: config},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   config, ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_group" {
			continue
		}

		_, err := organizationmanagerv1sdk.NewGroupClient(config.SDKv2).Get(context.Background(), &organizationmanager.GetGroupRequest{
			GroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Group still exists")
		}
	}

	return nil
}
