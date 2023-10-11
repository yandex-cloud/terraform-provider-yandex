package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func init() {
	resource.AddTestSweepers("yandex_compute_disk_placement_group", &resource.Sweeper{
		Name: "yandex_compute_disk_placement_group",
		F:    testSweepComputeDiskPlacementGroups,
		Dependencies: []string{
			"yandex_compute_disk",
		},
	})
}

func sweepComputeDiskPlacementGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexComputeDiskPlacementGroupDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Compute().DiskPlacementGroup().Delete(ctx, &compute.DeleteDiskPlacementGroupRequest{
		DiskPlacementGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func testSweepComputeDiskPlacementGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListDiskPlacementGroupsRequest{FolderId: conf.FolderID}
	it := conf.sdk.Compute().DiskPlacementGroup().DiskPlacementGroupIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepWithRetry(sweepComputeDiskPlacementGroupOnce, conf, "Placement group", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep compute Placement Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func TestAccComputeDisk_createDiskPlacementGroup(t *testing.T) {
	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var disk compute.Disk

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskPlacementGroup(diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.foobar", &disk),
					testAccCheckNonEmptyDiskPlacementGroup(&disk),
				),
			},
		},
	})
}

func TestAccComputeDisk_createAndEraseDiskPlacementGroup(t *testing.T) {
	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var disk compute.Disk

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskPlacementGroup(diskName),
			},
			{
				Config: testAccDiskNoPlacementGroup(diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.foobar", &disk),
					testAccCheckEmptyDiskPlacementGroup(&disk),
				),
			},
		},
	})
}

func TestAccComputeDisk_createEmptyDiskPlacementGroupAndAssignLater(t *testing.T) {
	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var disk compute.Disk

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskNoPlacementGroup(diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.foobar", &disk),
					testAccCheckEmptyDiskPlacementGroup(&disk),
				),
			},
			{
				Config: testAccDiskPlacementGroup(diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.foobar", &disk),
					testAccCheckNonEmptyDiskPlacementGroup(&disk),
				),
			},
		},
	})
}

func testAccCheckNonEmptyDiskPlacementGroup(disk *compute.Disk) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if disk.DiskPlacementPolicy != nil && disk.DiskPlacementPolicy.PlacementGroupId != "" {
			return nil
		}
		return fmt.Errorf("disk placement_group_id is invalid")
	}
}

func testAccCheckEmptyDiskPlacementGroup(disk *compute.Disk) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if disk.DiskPlacementPolicy != nil && disk.DiskPlacementPolicy.PlacementGroupId == "" {
			return nil
		}
		return fmt.Errorf("disk placement_group_id is invalid")
	}
}

func testAccDiskPlacementGroup(disk string) string {
	// language=tf
	return fmt.Sprintf(`
resource yandex_compute_disk foobar {
  name = "%s"
  size = 93
  type = "network-ssd-nonreplicated"
  zone = "ru-central1-b"

  disk_placement_policy {
    disk_placement_group_id = yandex_compute_disk_placement_group.pg.id
  }
}

resource yandex_compute_disk_placement_group pg {
  zone = "ru-central1-b"
}
`, disk)
}

func testAccDiskNoPlacementGroup(instance string) string {
	// language=tf
	return fmt.Sprintf(`
resource yandex_compute_disk foobar {
  name = "%s"
  size = 93
  type = "network-ssd-nonreplicated"
  zone = "ru-central1-b"

  disk_placement_policy {
    disk_placement_group_id = ""
  }
}

resource yandex_compute_disk_placement_group pg {
  zone = "ru-central1-b"
}
`, instance)
}
