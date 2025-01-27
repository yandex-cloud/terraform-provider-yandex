package yandex

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

const (
	snapshotScheduleResource = "yandex_compute_snapshot_schedule.foobar"
)

//revive:disable:var-naming
func TestAccComputeSnapshotSchedule_basic(t *testing.T) {
	t.Parallel()

	scheduleName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	diskName := "disk1"
	var schedule compute.SnapshotSchedule
	snapshotDescription := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckComputeSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeSnapshotSchedule_basic(diskName, scheduleName, snapshotDescription, "my-value-for-tag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeSnapshotScheduleExists(snapshotScheduleResource, &schedule),
					testAccCheckCreatedAtAttr(snapshotScheduleResource),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "id"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "labels.%"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "snapshot_spec.#"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "snapshot_spec.0.description"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "snapshot_spec.0.labels.%"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "schedule_policy.#"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "schedule_policy.0.start_at"),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "schedule_policy.0.expression", "0 0 1 1 *"),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "folder_id", getExampleFolderID()),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "name", scheduleName),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "disk_ids.#", "1"),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "snapshot_count", "1"),
				),
			},
			{
				// make sure state is unchanged
				Config: testAccComputeSnapshotSchedule_basic(diskName, scheduleName, snapshotDescription, "my-value-for-tag"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "id"),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "disk_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccComputeSnapshotSchedule_update(t *testing.T) {
	t.Parallel()

	var schedule compute.SnapshotSchedule
	scheduleName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))
	snapshotDescription := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeSnapshotSchedule_basic("disk1", scheduleName, snapshotDescription, "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeSnapshotScheduleExists(
						"yandex_compute_snapshot_schedule.foobar", &schedule),
				),
			},
			{
				Config: testAccComputeSnapshotSchedule_basic("disk2", scheduleName, snapshotDescription, "my-updated-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeSnapshotScheduleExists(
						"yandex_compute_snapshot_schedule.foobar", &schedule),
				),
			},
			{
				ResourceName:      "yandex_compute_snapshot_schedule.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckComputeSnapshotScheduleDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_snapshot_schedule" {
			continue
		}

		_, err := config.sdk.Compute().SnapshotSchedule().Get(context.Background(), &compute.GetSnapshotScheduleRequest{
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

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Compute().SnapshotSchedule().Get(context.Background(), &compute.GetSnapshotScheduleRequest{
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

func testAccComputeSnapshotSchedule_basic(diskName, scheduleName, snapshotDescription, labelValue string) string {
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
`, scheduleName, snapshotDescription, labelValue, diskName)
}

func Test_makeUpdateSnapshotScheduleDisksRequest(t *testing.T) {
	tests := []struct {
		name     string
		oldDisks map[string]bool
		newDisks map[string]bool
		want     *compute.UpdateSnapshotScheduleDisksRequest
	}{
		{
			name: "empty",
			want: &compute.UpdateSnapshotScheduleDisksRequest{},
		},
		{
			name:     "add",
			newDisks: map[string]bool{"disk1": true},
			want: &compute.UpdateSnapshotScheduleDisksRequest{
				Add: []string{"disk1"},
			},
		},
		{
			name:     "remove",
			oldDisks: map[string]bool{"disk1": true},
			want: &compute.UpdateSnapshotScheduleDisksRequest{
				Remove: []string{"disk1"},
			},
		},
		{
			name:     "add and remove",
			oldDisks: map[string]bool{"disk1": true, "disk": true},
			newDisks: map[string]bool{"disk2": true, "disk": true},
			want: &compute.UpdateSnapshotScheduleDisksRequest{
				Remove: []string{"disk1"},
				Add:    []string{"disk2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeUpdateSnapshotScheduleDisksRequest(tt.oldDisks, tt.newDisks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeUpdateSnapshotScheduleDisksRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
