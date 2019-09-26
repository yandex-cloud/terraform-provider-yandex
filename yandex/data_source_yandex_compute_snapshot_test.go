package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceComputeSnapshot_byID(t *testing.T) {
	t.Parallel()

	diskName := acctest.RandomWithPrefix("tf-disk")
	snapshotName := acctest.RandomWithPrefix("tf-snap")
	label := acctest.RandomWithPrefix("label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckComputeDiskDestroy,
			testAccCheckComputeSnapshotDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSnapshotConfig(diskName, snapshotName, label, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_snapshot.source", "snapshot_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_snapshot.source",
						"name", snapshotName),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"source_disk_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_snapshot.source",
						"labels.test_label", label),
					testAccCheckCreatedAtAttr("data.yandex_compute_snapshot.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeSnapshot_byName(t *testing.T) {
	t.Parallel()

	diskName := acctest.RandomWithPrefix("tf-disk")
	snapshotName := acctest.RandomWithPrefix("tf-snap")
	label := acctest.RandomWithPrefix("label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckComputeDiskDestroy,
			testAccCheckComputeSnapshotDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSnapshotConfig(diskName, snapshotName, label, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_snapshot.source", "snapshot_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_snapshot.source",
						"name", snapshotName),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"source_disk_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_snapshot.source",
						"labels.test_label", label),
					testAccCheckCreatedAtAttr("data.yandex_compute_snapshot.source"),
				),
			},
		},
	})
}

func testAccDataSourceSnapshotResourceConfig(diskName, snapshotName, labelValue string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
  size     = 4
}

resource "yandex_compute_snapshot" "foobar" {
  name           = "%s"
  source_disk_id = "${yandex_compute_disk.foobar.id}"

  labels = {
    test_label = "%s"
  }
}
`, diskName, snapshotName, labelValue)
}

func testAccDataSourceSnapshotConfig(diskName, snapshotName, labelValue string, useID bool) string {
	if useID {
		return testAccDataSourceSnapshotResourceConfig(diskName, snapshotName, labelValue) + computeSnapshotDataByIDConfig
	}

	return testAccDataSourceSnapshotResourceConfig(diskName, snapshotName, labelValue) + computeSnapshotDataByNameConfig
}

const computeSnapshotDataByIDConfig = `
data "yandex_compute_snapshot" "source" {
  snapshot_id = "${yandex_compute_snapshot.foobar.id}"
}
`

const computeSnapshotDataByNameConfig = `
data "yandex_compute_snapshot" "source" {
  name = "${yandex_compute_snapshot.foobar.name}"
}
`
