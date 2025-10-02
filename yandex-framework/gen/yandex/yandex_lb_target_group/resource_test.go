package yandex_lb_target_group_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	loadbalancersdk "github.com/yandex-cloud/go-sdk/services/loadbalancer/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const tgResource = "yandex_lb_target_group.test-tg"

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccLBTargetGroup_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()

	tgName := acctest.RandomWithPrefix("tf-target-group")
	folderID := test.GetExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccLBTargetGroupBasic(tgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tgResource, "name", tgName),
					resource.TestCheckResourceAttrSet(tgResource, "folder_id"),
					resource.TestCheckResourceAttr(tgResource, "folder_id", folderID),
					test.AccCheckCreatedAtAttr(tgResource),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccLBTargetGroupBasic(tgName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckLBTargetGroupDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_lb_target_group" {
			continue
		}

		_, err := loadbalancersdk.NewTargetGroupClient(config.SDKv2).Get(context.Background(), &loadbalancer.GetTargetGroupRequest{
			TargetGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("TargetGroup still exists")
		}
	}

	return nil
}

func testAccLBTargetGroupBasic(name string) string {
	return fmt.Sprintf(`
resource "yandex_lb_target_group" "test-tg" {
  name		= "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
`, name)
}
