package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceComputeDisk_byID(t *testing.T) {
	t.Parallel()

	family := "ubuntu-1804-lts"
	diskName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomDiskConfig(family, diskName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_disk.source", "disk_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"name", diskName),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"instance_ids.#", "0"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"type", "network-hdd"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"zone"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"block_size", "4096"),
					testAccCheckCreatedAtAttr("data.yandex_compute_disk.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeDisk_byName(t *testing.T) {
	t.Parallel()

	family := "ubuntu-1804-lts"
	diskName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomDiskConfig(family, diskName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_disk.source", "disk_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"name", diskName),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"instance_ids.#", "0"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"type", "network-hdd"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"zone"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"block_size", "4096"),
					testAccCheckCreatedAtAttr("data.yandex_compute_disk.source"),
				),
			},
		},
	})
}

func testAccDataSourceCustomDiskResourceConfig(family, name string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "%s"
}

resource "yandex_compute_disk" "foo" {
  name       = "%s"
  zone       = "ru-central1-a"
  image_id   = "${data.yandex_compute_image.ubuntu.id}"
  size       = 8
  block_size = 4096

  labels = {
    my-label = "my-label-value"
  }
}
`, family, name)
}

func testAccDataSourceCustomDiskConfig(family, name string, useID bool) string {
	if useID {
		return testAccDataSourceCustomDiskResourceConfig(family, name) + computeDiskDataByIDConfig
	}

	return testAccDataSourceCustomDiskResourceConfig(family, name) + computeDiskDataByNameConfig
}

const computeDiskDataByIDConfig = `
data "yandex_compute_disk" "source" {
  disk_id = "${yandex_compute_disk.foo.id}"
}
`

const computeDiskDataByNameConfig = `
data "yandex_compute_disk" "source" {
  name = "${yandex_compute_disk.foo.name}"
}
`
