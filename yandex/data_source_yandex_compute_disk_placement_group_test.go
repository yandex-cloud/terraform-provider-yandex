package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceComputeDisk_diskPlacementGroupByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeDiskPlacementConfig(true),
				Check: testAccDataSourceComputeDiskPlacementGroupCheck(
					"data.yandex_compute_disk_placement_group.pgd",
					"yandex_compute_disk_placement_group.pgr"),
			},
		},
	})
}

func TestAccDataSourceComputeDisk_diskPlacementGroupByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeDiskPlacementConfig(false),
				Check: testAccDataSourceComputeDiskPlacementGroupCheck(
					"data.yandex_compute_disk_placement_group.pgd",
					"yandex_compute_disk_placement_group.pgr"),
			},
		},
	})
}

func testAccDataSourceComputeDiskPlacementGroupAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		if ds.Primary.ID != rs.Primary.ID {
			return fmt.Errorf("instance `data source` ID does not match `resource` ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
		}

		datasourceAttributes := ds.Primary.Attributes
		resourceAttributes := rs.Primary.Attributes

		instanceAttrsToTest := []string{
			"name",
			"folder_id",
			"description",
			"labels",
			"status",
			"zone",
		}

		for _, attrToCheck := range instanceAttrsToTest {
			if datasourceAttributes[attrToCheck] != resourceAttributes[attrToCheck] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrToCheck,
					datasourceAttributes[attrToCheck],
					resourceAttributes[attrToCheck],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceComputeDiskPlacementGroupCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceComputeDiskPlacementGroupAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "group_id"),
	)
}

func testAccDataSourceComputeDiskPlacementConfig(useDataID bool) string {
	if useDataID {
		// language=tf
		return `
data yandex_compute_disk_placement_group pgd {
  group_id = yandex_compute_disk_placement_group.pgr.id
}

resource yandex_compute_disk_placement_group pgr {
  name        = "my-group"
  description = "my description"
  labels = {
    first  = "xxx"
    second = "yyy"
  }
}
`
	}
	// language=tf
	return `
data yandex_compute_disk_placement_group pgd {
  name = yandex_compute_disk_placement_group.pgr.name
}

resource yandex_compute_disk_placement_group pgr {
  name        = "my-group"
  description = "my description"
  labels = {
    first  = "xxx"
    second = "yyy"
  }
}
`
}
