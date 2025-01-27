package compute_snapshot_schedule_iam_binding_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	snapshotScheduleResource = "yandex_compute_snapshot_schedule.foobar"
	timeout                  = time.Minute * 15
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

//revive:disable:var-naming
func TestAccComputeSnapshotSchedule_basicIamMember(t *testing.T) {

	var (
		schedule            compute.SnapshotSchedule
		snapshotDescription = test.GenerateNameForResource(10)
		scheduleName        = test.GenerateNameForResource(10)
		diskName            = "disk1"
		userID              = "allUsers"
		role                = "editor"
		ctx, cancel         = context.WithTimeout(context.Background(), timeout)
	)
	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeSnapshotSchedule_basic(diskName, scheduleName, snapshotDescription, "my-value-for-tag", role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeSnapshotScheduleExists(snapshotScheduleResource, &schedule),
					test.TestAccCheckIamBindingExists(ctx, func() test.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return cfg.SDK.Compute().SnapshotSchedule()
					}, &schedule, role, []string{"system:" + userID}),
				),
			},
		},
	})
}

func testAccCheckComputeSnapshotScheduleDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_snapshot_schedule" {
			continue
		}

		_, err := config.SDK.Compute().SnapshotSchedule().Get(context.Background(), &compute.GetSnapshotScheduleRequest{
			SnapshotScheduleId: rs.Primary.ID,
		})
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
		}
		return fmt.Errorf("SnapshotSchedule still exists")
	}

	return nil
}

func testAccCheckComputeSnapshotScheduleExists(n string, schedule *compute.SnapshotSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.Compute().SnapshotSchedule().Get(context.Background(), &compute.GetSnapshotScheduleRequest{
			SnapshotScheduleId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Snapshot schedule %s not found", n)
		}

		//goland:noinspection GoVetCopyLock (this comment suppress warning in Idea IDE about coping sync.Mutex)
		*schedule = *found

		return nil
	}
}

func testAccComputeSnapshotSchedule_basic(diskName, scheduleName, snapshotDescription, labelValue, role, userID string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "disk1" {
  image_id = "${data.yandex_compute_image.ubuntu.id}"
  size     = 4
}

resource "yandex_compute_disk" "disk2" {
	image_id = "${data.yandex_compute_image.ubuntu.id}"
	size     = 4
  }

resource "yandex_compute_snapshot_schedule" "foobar" {
  name           = "%s"

  schedule_policy {
	expression = "0 0 1 1 *"
  }

  snapshot_count = 1

  snapshot_spec {
	description = "%s"
	labels = {
	  test_label = "test_value"
	}
  }

  labels = {
    test_label = "%s"
  }

  disk_ids = ["${yandex_compute_disk.%s.id}"]
}

resource "yandex_compute_snapshot_schedule_iam_binding" "test-binding" {
  role = "%s"
  members = ["system:%s"]
  snapshot_schedule_id = yandex_compute_snapshot_schedule.foobar.id
}

`, scheduleName, snapshotDescription, labelValue, diskName, role, userID)
}
