package acceptance

import (
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const (
	snapshotScheduleData = "data.yandex_compute_snapshot_schedule.source"
)

func TestAccDataSourceComputeSnapshotSchedule_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	scheduleName := acctest.RandomWithPrefix("tf-disk")
	snapshotDescription := acctest.RandomWithPrefix("tf-snap")
	label := acctest.RandomWithPrefix("label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acceptance.PreCheck(t) },
		ProviderFactories: suite.ProviderFactories(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			suite.testAccCheckComputeSnapshotScheduleDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSnapshotScheduleConfig(scheduleName, snapshotDescription, label, true),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField(snapshotScheduleData, "snapshot_schedule_id"),
					resource.TestCheckResourceAttr(snapshotScheduleData, "name", scheduleName),
					resource.TestCheckResourceAttrSet(snapshotScheduleData, "id"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "labels.%"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "snapshot_spec.#"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "snapshot_spec.0.description"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "schedule_policy.#"),
					resource.TestCheckResourceAttrSet(snapshotScheduleResource, "schedule_policy.0.start_at"),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "schedule_policy.0.expression", "0 0 1 1 *"),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "folder_id", acceptance.FolderID()),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "name", scheduleName),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "disk_ids.#", "1"),
					resource.TestCheckResourceAttr(snapshotScheduleResource, "snapshot_count", "1"),
					statecheck.CheckCreatedAtAttr(snapshotScheduleData),
				),
			},
		},
	})
}

func TestAccDataSourceComputeSnapshotSchedule_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	scheduleName := acctest.RandomWithPrefix("tf-disk")
	snapshotDescription := acctest.RandomWithPrefix("tf-snap")
	label := acctest.RandomWithPrefix("label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			suite.testAccCheckComputeSnapshotScheduleDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSnapshotScheduleConfig(scheduleName, snapshotDescription, label, false),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField(snapshotScheduleData, "snapshot_schedule_id"),
					resource.TestCheckResourceAttr(snapshotScheduleData,
						"name", scheduleName),
					resource.TestCheckResourceAttrSet(snapshotScheduleData,
						"id"),
				),
			},
		},
	})
}

func testAccDataSourceSnapshotScheduleResourceConfig(scheduleName, snapshotDescription, labelValue string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
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
  }

  labels = {
    test_label = "%s"
  }

  disk_ids = ["${yandex_compute_disk.foobar.id}"]
}
`, scheduleName, snapshotDescription, labelValue)
}

func testAccDataSourceSnapshotScheduleConfig(scheduleName, snapshotDescription, labelValue string, useID bool) string {
	if useID {
		return testAccDataSourceSnapshotScheduleResourceConfig(scheduleName, snapshotDescription, labelValue) + computeSnapshotScheduleDataByIDConfig
	}

	return testAccDataSourceSnapshotScheduleResourceConfig(scheduleName, snapshotDescription, labelValue) + computeSnapshotScheduleDataByNameConfig
}

const computeSnapshotScheduleDataByIDConfig = `
data "yandex_compute_snapshot_schedule" "source" {
  snapshot_schedule_id = "${yandex_compute_snapshot_schedule.foobar.id}"
}
`

const computeSnapshotScheduleDataByNameConfig = `
data "yandex_compute_snapshot_schedule" "source" {
  name = "${yandex_compute_snapshot_schedule.foobar.name}"
}
`
