package yandex

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

//revive:disable:var-naming
func TestAccComputeSnapshot_basic(t *testing.T) {
	t.Parallel()

	snapshotName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var snapshot compute.Snapshot
	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeSnapshot_basic(snapshotName, diskName, "my-value-for-tag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeSnapshotExists(
						"yandex_compute_snapshot.foobar", &snapshot),
					testAccCheckCreatedAtAttr("yandex_compute_snapshot.foobar"),
					resource.TestCheckResourceAttr("yandex_compute_snapshot.foobar", "hardware_generation.#", "1"),
					resource.TestCheckResourceAttr("yandex_compute_snapshot.foobar", "hardware_generation.0.generation2_features.#", "1"),
				),
			},
		},
	})
}

func TestAccComputeSnapshot_update(t *testing.T) {
	t.Parallel()

	var snapshot compute.Snapshot
	snapshotName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))
	diskName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeSnapshot_basic(snapshotName, diskName, "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeSnapshotExists(
						"yandex_compute_snapshot.foobar", &snapshot),
				),
			},
			{
				Config: testAccComputeSnapshot_basic(snapshotName, diskName, "my-updated-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeSnapshotExists(
						"yandex_compute_snapshot.foobar", &snapshot),
				),
			},
			{
				ResourceName:      "yandex_compute_snapshot.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckComputeSnapshotDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_snapshot" {
			continue
		}

		_, err := config.sdk.Compute().Snapshot().Get(context.Background(), &compute.GetSnapshotRequest{
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

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Compute().Snapshot().Get(context.Background(), &compute.GetSnapshotRequest{
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

		foundDisk, errDisk := config.sdk.Compute().Disk().Get(context.Background(), &compute.GetDiskRequest{
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

func testAccComputeSnapshot_basic(snapshotName, diskName, labelValue string) string {
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

  hardware_generation {
    legacy_features {
	  pci_topology = "PCI_TOPOLOGY_V2"
	}
  }
}

resource "yandex_compute_snapshot" "foobar" {
  name           = "%s"
  source_disk_id = "${yandex_compute_disk.foobar.id}"

  labels = {
    test_label = "%s"
  }

  hardware_generation {
    generation2_features {}
  }
}
`, diskName, snapshotName, labelValue)
}
