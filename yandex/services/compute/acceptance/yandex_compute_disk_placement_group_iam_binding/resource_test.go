package yandex_compute_disk_placement_group_iam_binding_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const timeout = 15 * time.Minute

var (
	diskPlacementName = test.GenerateNameForResource(10)
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.

func TestAccComputeDisk_createDiskPlacementGroupIamMember(t *testing.T) {
	var (
		placement   compute.DiskPlacementGroup
		userID      = "allUsers"
		role        = "editor"
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	)

	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeDiskPlacementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskPlacementGroupWithIAMMember(diskPlacementName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskPlacementGroupExists(
						"yandex_compute_disk_placement_group.pg",
						&placement,
						timeout,
					),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() (iam.BindingsGetter, error) {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return computesdk.NewDiskPlacementGroupClient(cfg.SDKv2), nil
					}, &placement, role, []string{"system:" + userID}),
				),
			},
		},
	})
}

func testAccDiskPlacementGroupWithIAMMember(diskPlacementName, role, userID string) string {
	// language=tf
	return fmt.Sprintf(`
resource yandex_compute_disk_placement_group pg {
  zone = "ru-central1-b"
  name = "%s"
}

resource "yandex_compute_disk_placement_group_iam_binding" "test-diskplacement-binding" {
  role = "%s"
  members = ["system:%s"]
  disk_placement_group_id = yandex_compute_disk_placement_group.pg.id
}
`, diskPlacementName, role, userID)
}

func testAccCheckComputeDiskPlacementGroupDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	return checkDiskPlacementGroupDestroy(s, computesdk.NewDiskPlacementGroupClient(config.SDKv2))
}

type diskPlacementGroupGetter interface {
	Get(context.Context, *compute.GetDiskPlacementGroupRequest, ...grpc.CallOption) (*compute.DiskPlacementGroup, error)
}

func checkDiskPlacementGroupDestroy(s *terraform.State, client diskPlacementGroupGetter) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_disk_placement_group" {
			continue
		}
		_, err := client.Get(context.Background(), &compute.GetDiskPlacementGroupRequest{DiskPlacementGroupId: rs.Primary.ID})
		if err == nil {
			return fmt.Errorf("disk placement group still exists")
		}
	}
	return nil
}
