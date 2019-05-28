package yandex

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func TestAccComputeDisk_basic(t *testing.T) {
	t.Parallel()

	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var disk compute.Disk

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeDisk_basic(diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.foobar", &disk),
					testAccCheckComputeDiskHasLabel(&disk, "my-label", "my-label-value"),
					testAccCheckCreatedAtAttr("yandex_compute_disk.foobar"),
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

func TestAccComputeDisk_timeout(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccComputeDisk_timeout(),
				ExpectError: regexp.MustCompile("DeadlineExceeded|deadline exceeded"),
			},
		},
	})
}

func TestAccComputeDisk_update(t *testing.T) {
	t.Parallel()

	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var disk compute.Disk

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeDisk_basic(diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.foobar", &disk),
					resource.TestCheckResourceAttr("yandex_compute_disk.foobar", "size", "4"),
					testAccCheckComputeDiskHasLabel(&disk, "my-label", "my-label-value"),
					testAccCheckCreatedAtAttr("yandex_compute_disk.foobar"),
				),
			},
			{
				Config: testAccComputeDisk_updated(diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.foobar", &disk),
					resource.TestCheckResourceAttr("yandex_compute_disk.foobar", "size", "8"),
					testAccCheckComputeDiskHasLabel(&disk, "my-label", "my-updated-label-value"),
					testAccCheckComputeDiskHasLabel(&disk, "a-new-label", "a-new-label-value"),
					testAccCheckCreatedAtAttr("yandex_compute_disk.foobar"),
				),
			},
		},
	})
}

func TestAccComputeDisk_fromSnapshot(t *testing.T) {
	t.Parallel()

	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	firstDiskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	snapshotName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	var disk compute.Disk

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeDisk_fromSnapshot(firstDiskName, snapshotName, diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.seconddisk", &disk),
				),
			},
		},
	})
}

func TestAccComputeDisk_deleteDetach(t *testing.T) {
	t.Skip("enable when instance disk attach/detach operation will be supported")
	t.Parallel()

	diskName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	instanceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var disk compute.Disk

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeDisk_deleteDetach(instanceName, diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.foo", &disk),
				),
			},
			// this needs to be a second step so we refresh and see the instance
			// listed as attached to the disk; the instance is created after the
			// disk. and the disk's properties aren't refreshed unless there's
			// another step
			{
				Config: testAccComputeDisk_deleteDetach(instanceName, diskName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeDiskExists(
						"yandex_compute_disk.foo", &disk),
				),
			},
		},
	})
}

func testAccCheckComputeDiskDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_disk" {
			continue
		}

		_, err := config.sdk.Compute().Disk().Get(context.Background(), &compute.GetDiskRequest{
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

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Compute().Disk().Get(context.Background(), &compute.GetDiskRequest{
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

func testAccCheckComputeDiskHasLabel(disk *compute.Disk, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		val, ok := disk.Labels[key]
		if !ok {
			return fmt.Errorf("Label with key %s not found", key)
		}

		if val != value {
			return fmt.Errorf("Label value did not match for key %s: expected %s but found %s", key, value, val)
		}
		return nil
	}
}

//revive:disable:var-naming
func testAccComputeDisk_basic(diskName string) string {
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
`, diskName)
}

func testAccComputeDisk_timeout() string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
  size     = 4
  type     = "network-hdd"

  timeouts {
    create = "1s"
  }
}
`, acctest.RandomWithPrefix("tf-disk"))
}

func testAccComputeDisk_updated(diskName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
  size     = 8
  type     = "network-hdd"

  labels = {
    my-label    = "my-updated-label-value"
    a-new-label = "a-new-label-value"
  }
}
`, diskName)
}

func testAccComputeDisk_fromSnapshot(firstDiskName, snapshotName, secondDiskName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "d1-%s"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
  size     = 4
  type     = "network-hdd"
}

resource "yandex_compute_snapshot" "snapdisk" {
  name           = "%s"
  source_disk_id = "${yandex_compute_disk.foobar.id}"
}

resource "yandex_compute_disk" "seconddisk" {
  name        = "d2-%s"
  size        = 6
  snapshot_id = "${yandex_compute_snapshot.snapdisk.id}"
  type        = "network-hdd"
}
`, firstDiskName, snapshotName, secondDiskName)
}

func testAccComputeDisk_deleteDetach(instanceName, diskName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foo" {
  name     = "%s"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
  size     = 4
  type     = "network-hdd"
}

resource "yandex_compute_instance" "bar" {
  name = "%s"

  resources {
    cores  = 1
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  secondary_disk {
    disk_id = "${yandex_compute_disk.foo.id}"
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.bar-subnet.id}"
  }
}

resource "yandex_vpc_network" "foo-network" {}

resource "yandex_vpc_subnet" "bar-subnet" {
  zone           = "ru-central1-a"
  name           = "testacccomputedisk-deletedetach"
  network_id     = "${yandex_vpc_network.foo-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, diskName, instanceName)
}
