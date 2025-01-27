package compute_placement_group_iam_binding_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const timeout = 15 * time.Minute

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeInstance_createPlacementGroupIamMember(t *testing.T) {
	var (
		placementGroup compute.PlacementGroup
		pgName         = fmt.Sprintf("%s-%s", test.TestPrefix(), fmt.Sprintf("instance-test-%s", acctest.RandString(10)))
		userID         = "allUsers"
		role           = "editor"
		ctx, cancel    = context.WithTimeout(context.Background(), timeout)
	)
	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstancePlacementGroupWithPartitionStrategy(role, userID, pgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_placement_group.pg", &placementGroup),
					test.TestAccCheckIamBindingExists(ctx, func() test.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return cfg.SDK.Compute().PlacementGroup()
					}, &placementGroup, role, []string{"system:" + userID}),
				),
			},
		},
	})
}

func testAccCheckComputeInstanceDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_instance" {
			continue
		}

		_, err := config.SDK.Compute().Instance().Get(context.Background(), &compute.GetInstanceRequest{
			InstanceId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Instance still exists")
		}
	}

	return nil
}

func testAccCheckComputeInstanceExists(n string, instance *compute.PlacementGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.Compute().PlacementGroup().Get(context.Background(), &compute.GetPlacementGroupRequest{
			PlacementGroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Placement group not found")
		}

		*instance = *found

		return nil
	}
}

func testAccComputeInstancePlacementGroupWithPartitionStrategy(role, userID, name string) string {
	// language=tf
	return fmt.Sprintf(`
resource yandex_compute_placement_group pg {
  name = "%s"
  placement_strategy_partitions = 3
}

resource "yandex_compute_placement_group_iam_binding" "test-placement-iam" {
  role = "%s"
  members = ["system:%s"]
  placement_group_id = yandex_compute_placement_group.pg.id
}
`, name, role, userID)
}
