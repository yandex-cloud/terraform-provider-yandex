package diskplacement

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test/compute/diskcheck"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test/compute/iam"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test/compute/name"
	"golang.org/x/net/context"
)

const timeout = 15 * time.Minute

var (
	diskPlacementName = name.GenerateNameForResource(10)
)

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
		CheckDestroy:             testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskPlacementGroupWithIAMMember(diskPlacementName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					diskcheck.TestAccCheckComputeDiskPlacementGroupExists(
						"yandex_compute_disk_placement_group.pg",
						&placement,
						timeout,
					),
					iam.TestAccCheckIamBindingExists(ctx, func() iam.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return cfg.SDK.Compute().DiskPlacementGroup()
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

func testAccCheckComputeDiskDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_disk" {
			continue
		}

		_, err := config.SDK.Compute().Disk().Get(context.Background(), &compute.GetDiskRequest{
			DiskId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Disk still exists")
		}
	}

	return nil
}
