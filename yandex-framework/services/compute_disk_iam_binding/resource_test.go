package compute_disk_iam_binding_test

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

const (
	timeout = 15 * time.Minute
)

var (
	diskName = acctest.RandomWithPrefix(test.TestPrefix())
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeDisk_basicIamMember(t *testing.T) {
	var (
		disk        compute.Disk
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
				Config: testAccComputeDiskWithIAMMember_basic(diskName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					test.TestAccCheckComputeDiskExists(
						"yandex_compute_disk.foobar",
						&disk,
						timeout,
					),
					test.TestAccCheckIamBindingExists(ctx, func() test.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return cfg.SDK.Compute().Disk()
					}, &disk, role, []string{"system:" + userID}),
				),
			},
			{
				ResourceName: "yandex_compute_disk.foobar",
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return disk.Id, nil
				},
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"image"},
				ImportStateVerify:       true,
			},
		},
	})
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

//revive:disable:var-naming
func testAccComputeDiskWithIAMMember_basic(diskName, role, userID string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
  size     = 4
  type     = "network-hdd"

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_compute_disk_iam_binding" "test-disk-binding" {
  role = "%s"
  members = ["system:%s"]
  disk_id = yandex_compute_disk.foobar.id
}

`, diskName, role, userID)
}
