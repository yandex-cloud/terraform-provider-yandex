package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceComputeInstanceGroup_byID(t *testing.T) {
	t.Parallel()

	igName := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeInstanceGroupConfig(igName, saName),
				Check:  testAccDataSourceComputeInstanceGroupCheck("data.yandex_compute_instance_group.bar", "yandex_compute_instance_group.group1"),
			},
		},
	})
}

func TestAccDataSourceComputeInstanceGroupGpus_byID(t *testing.T) {
	igName := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeInstanceGroupGpusConfig(igName, saName),
				Check:  testAccDataSourceComputeInstanceGroupCheck("data.yandex_compute_instance_group.bar", "yandex_compute_instance_group.group1"),
			},
		},
	})
}

const computeInstanceGroupDataByIDConfig = `
data "yandex_compute_instance_group" "bar" {
  instance_group_id = "${yandex_compute_instance_group.group1.id}"
}
`

func testAccDataSourceComputeInstanceGroupConfig(igName string, saName string) string {
	return testAccComputeInstanceGroupConfigMain(igName, saName) + computeInstanceGroupDataByIDConfig
}

func testAccDataSourceComputeInstanceGroupGpusConfig(igName string, saName string) string {
	return testAccComputeInstanceGroupConfigGpus(igName, saName) + computeInstanceGroupDataByIDConfig
}

func testAccDataSourceComputeInstanceGroupCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
			return fmt.Errorf("instance group `data source` ID does not match `resource` ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
		}

		datasourceAttributes := ds.Primary.Attributes
		resourceAttributes := rs.Primary.Attributes

		instanceAttrsToTest := []string{
			"name",
			"folder_id",
			"description",
			"instance_template.0.labels.%",
			"instance_template.0.metadata.%",
			"instance_template.0.boot_disk.0.initialize_params.0.type",
			"instance_template.0.resources.0.core_fraction",
			"scale_policy.0.fixed_scale.0.size",
			"allocation_policy.#",
			"deploy_policy.0.startup_duration",
			"scale_policy.#",
			"service_account_id",
			"allocation_policy.0.zones.#",
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
