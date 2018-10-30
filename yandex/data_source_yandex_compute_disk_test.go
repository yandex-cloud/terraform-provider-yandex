package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
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
				Config: testAccDataSourceCustomDiskConfig(family, diskName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"name", diskName),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"instance_ids.#", "0"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"source_image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"type", "network-hdd"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"zone"),
				),
			},
		},
	})
}

func testAccDataSourceCustomDiskConfig(family, name string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
	family = "%s"
}

resource "yandex_compute_disk" "source_disk" {
    name     = "%s"
    zone     = "ru-central1-a"
    image_id = "${data.yandex_compute_image.ubuntu.id}"
    size     = 8

	labels {
		my-label = "my-label-value"
	}
}

data "yandex_compute_disk" "source" {
    disk_id = "${yandex_compute_disk.source_disk.id}"
}
`, family, name)
}
