package yandex_compute_disk_placement_group_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeDisk_UpgradeFromSDKv2(t *testing.T) {
	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccDiskPlacementGroup(diskName),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccDiskPlacementGroup(diskName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccComputeDisk_createDiskPlacementGroup(t *testing.T) {
	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var disk compute.Disk

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeDiskDestroy,
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
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeDiskDestroy,
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
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeDiskDestroy,
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

func testAccCheckComputeDiskDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_disk" {
			continue
		}

		_, err := computesdk.NewDiskClient(config.SDKv2).Get(context.Background(), &compute.GetDiskRequest{
			DiskId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Disk still exists")
		}
	}

	return nil
}

func testAccCheckComputeDiskExists(n string, disk *compute.Disk) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		//p := getTestProjectFromEnv()
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := computesdk.NewDiskClient(config.SDKv2).Get(context.Background(), &compute.GetDiskRequest{
			DiskId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Disk not found")
		}

		*disk = *found

		return nil
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
