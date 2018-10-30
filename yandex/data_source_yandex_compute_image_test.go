package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceComputeImage_byID(t *testing.T) {
	t.Parallel()

	family := "ubuntu-1804-lts"
	name := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckComputeImageDestroy,
			testAccCheckComputeDiskDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomImageConfig(family, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.yandex_compute_image.from_id",
						"name", name),
					resource.TestCheckResourceAttr("data.yandex_compute_image.from_id",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.from_id",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.from_id",
						"created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeImage_StandardByFamily(t *testing.T) {
	t.Parallel()

	family := "ubuntu-1804-lts"
	re := regexp.MustCompile("ubuntu")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceStandardImageByFamily(family),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.yandex_compute_image.by_family",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.by_family",
						"id"),
					resource.TestMatchResourceAttr("data.yandex_compute_image.by_family", "name", re),
				),
			},
		},
	})
}

func testAccDataSourceCustomImageConfig(family, name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_image" "image" {
    family        = "%s"
    name          = "%s"
    source_disk   = "${yandex_compute_disk.disk.id}"
    min_disk_size = 10
    os_type       = "linux"
}

resource "yandex_compute_disk" "disk" {
    name     = "%s-disk"
    zone     = "ru-central1-a"
    size     = 4
}

data "yandex_compute_image" "from_id" {
    image_id = "${yandex_compute_image.image.id}"
}
`, family, name, name)
}

func testAccDataSourceStandardImageByFamily(family string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "by_family" {
    family = "%s"
}
`, family)
}
