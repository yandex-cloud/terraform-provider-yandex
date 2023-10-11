package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceComputePlacementGroup_byID(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputePlacementConfig(true),
				Check: testAccDataSourceComputePlacementGroupCheck(
					"data.yandex_compute_placement_group.pgd",
					"yandex_compute_placement_group.pgr"),
			},
		},
	})
}

func TestAccDataSourceComputePlacementGroup_byName(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputePlacementConfig(false),
				Check: testAccDataSourceComputePlacementGroupCheck(
					"data.yandex_compute_placement_group.pgd",
					"yandex_compute_placement_group.pgr"),
			},
		},
	})
}

func testAccDataSourceComputePlacementGroupAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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

func testAccDataSourceComputePlacementGroupCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceComputePlacementGroupAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "group_id"),
	)
}

func testAccDataSourceComputePlacementConfig(useDataID bool) string {
	if useDataID {
		// language=tf
		return `
data yandex_compute_placement_group pgd {
  group_id = yandex_compute_placement_group.pgr.id
}

resource yandex_compute_placement_group pgr {
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
data yandex_compute_placement_group pgd {
  name = yandex_compute_placement_group.pgr.name
}

resource yandex_compute_placement_group pgr {
  name        = "my-group"
  description = "my description"
  labels = {
    first  = "xxx"
    second = "yyy"
  }
}
`
}
