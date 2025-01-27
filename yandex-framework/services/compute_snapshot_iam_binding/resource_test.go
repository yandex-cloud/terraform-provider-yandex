package compute_snapshot_iam_binding_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
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
	timeout = time.Minute * 15
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeSnapshot_basicIamMember(t *testing.T) {
	var (
		snapshotName = test.GenerateNameForResource(10)
		snapshot     compute.Snapshot
		diskName     = test.GenerateNameForResource(10)
		userID       = "allUsers"
		role         = "editor"
		ctx, cancel  = context.WithTimeout(context.Background(), timeout)
	)
	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeSnapshot_basic(snapshotName, diskName, "my-value-for-tag", role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeSnapshotExists("yandex_compute_snapshot.foobar", &snapshot),
					test.TestAccCheckIamBindingExists(ctx, func() test.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return cfg.SDK.Compute().Snapshot()
					}, &snapshot, role, []string{"system:" + userID}),
				),
			},
		},
	})
}

func testAccCheckComputeSnapshotDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_snapshot" {
			continue
		}

		_, err := config.SDK.Compute().Snapshot().Get(context.Background(), &compute.GetSnapshotRequest{
			SnapshotId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Snapshot still exists")
		}
	}

	return nil
}

func testAccCheckComputeSnapshotExists(n string, snapshot *compute.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.Compute().Snapshot().Get(context.Background(), &compute.GetSnapshotRequest{
			SnapshotId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Snapshot %s not found", n)
		}

		attr := rs.Primary.Attributes["source_disk_id"]
		if found.SourceDiskId != attr {
			return fmt.Errorf("Snapshot %s has mismatched source disk id.\nTF State: %+v.\nYC State: %+v",
				n, attr, found.SourceDiskId)
		}

		foundDisk, errDisk := config.SDK.Compute().Disk().Get(context.Background(), &compute.GetDiskRequest{
			DiskId: rs.Primary.Attributes["source_disk_id"],
		})
		if errDisk != nil {
			return errDisk
		}
		if foundDisk.Id != attr {
			return fmt.Errorf("Snapshot %s has mismatched source disk\nTF State: %+v.\nYC State: %+v",
				n, attr, foundDisk.Id)
		}

		_, ok = rs.Primary.Attributes["labels.%"]
		if !ok {
			return fmt.Errorf("Snapshot %s has no labels map in attributes", n)
		}

		attrMap := make(map[string]string)
		for k, v := range rs.Primary.Attributes {
			if !strings.HasPrefix(k, "labels.") || k == "labels.%" {
				continue
			}
			key := k[len("labels."):]
			attrMap[key] = v
		}
		if (len(attrMap) != 0 || len(found.Labels) != 0) && !reflect.DeepEqual(attrMap, found.Labels) {
			return fmt.Errorf("Snapshot %s has mismatched labels.\nTF State: %+v\nYC State: %+v",
				n, attrMap, found.Labels)
		}

		*snapshot = *found

		return nil
	}
}

func testAccComputeSnapshot_basic(snapshotName, diskName, labelValue, role, userID string) string {
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
    disk_label = "value-of-disk-label"
  }
}

resource "yandex_compute_snapshot" "foobar" {
  name           = "%s"
  source_disk_id = "${yandex_compute_disk.foobar.id}"

  labels = {
    test_label = "%s"
  }
}

resource "yandex_compute_snapshot_iam_binding" "test-snapshot-iam-binding" {
  role = "%s"
  members = ["system:%s"]
  snapshot_id = yandex_compute_snapshot.foobar.id
}
`, diskName, snapshotName, labelValue, role, userID)
}
