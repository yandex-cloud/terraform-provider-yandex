package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
				Check:  testAccDataSourceComputeInstanceGroupFixedScaleCheck("data.yandex_compute_instance_group.bar", "yandex_compute_instance_group.group1"),
			},
		},
	})
}

func TestAccDataSourceComputeInstanceGroup_GpusByID(t *testing.T) {
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

func TestAccDataSourceComputeInstanceGroup_AutoscaleByID(t *testing.T) {
	igName := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeInstanceGroupAutoscaleConfig(igName, saName),
				Check:  testAccDataSourceComputeInstanceGroupAutoScaleCheck("data.yandex_compute_instance_group.bar", "yandex_compute_instance_group.group1"),
			},
		},
	})
}

func TestAccDataSourceComputeInstanceGroup_InstanceTagsPool(t *testing.T) {
	igName := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeInstanceGroupInstanceTagsPoolConfig(igName, saName),
				Check:  testAccDataSourceComputeInstanceGroupInstanceTagsPoolCheck("data.yandex_compute_instance_group.bar", "yandex_compute_instance_group.group1"),
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

func testAccDataSourceComputeInstanceGroupAutoscaleConfig(igName string, saName string) string {
	return testAccComputeInstanceGroupConfigAutoScale(igName, saName) + computeInstanceGroupDataByIDConfig
}

func testAccDataSourceComputeInstanceGroupGpusConfig(igName string, saName string) string {
	return testAccComputeInstanceGroupConfigGpus(igName, saName) + computeInstanceGroupDataByIDConfig
}

func testAccDataSourceComputeInstanceGroupInstanceTagsPoolConfig(igName string, saName string) string {
	return testAccComputeInstanceGroupConfigInstanceTagsPool(igName, saName) + computeInstanceGroupDataByIDConfig
}

func testAttrsCheck(datasourceName string, resourceName string, testAttrs []string) resource.TestCheckFunc {
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

		for _, attrToCheck := range testAttrs {
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

var baseAttrsToTest = []string{
	"name",
	"folder_id",
	"description",
	"instance_template.0.labels.%",
	"instance_template.0.metadata.%",
	"instance_template.0.metadata_options.0.gce_http_endpoint",
	"instance_template.0.metadata_options.0.aws_v1_http_endpoint",
	"instance_template.0.metadata_options.0.gce_http_token",
	"instance_template.0.metadata_options.0.aws_v1_http_token",
	"instance_template.0.boot_disk.0.initialize_params.0.type",
	"instance_template.0.resources.0.core_fraction",
	"service_account_id",
	"deploy_policy.#",
	"deploy_policy.0.startup_duration",
	"deploy_policy.0.strategy",
	"scale_policy.#",
	"allocation_policy.#",
	"allocation_policy.0.zones.#",
	"instances.0.fqdn",
	"instances.0.name",
}

func testAccDataSourceComputeInstanceGroupCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return testAttrsCheck(datasourceName, resourceName, baseAttrsToTest)
}

func testAccDataSourceComputeInstanceGroupAutoScaleCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	instanceAttrsToTest := []string{
		"scale_policy.0.auto_scale.0.auto_scale_type",
		"scale_policy.0.auto_scale.0.initial_size",
		"scale_policy.0.auto_scale.0.max_size",
		"scale_policy.0.auto_scale.0.min_zone_size",
		"scale_policy.0.auto_scale.0.measurement_duration",
		"scale_policy.0.auto_scale.0.cpu_utilization_target",
	}

	instanceAttrsToTest = append(instanceAttrsToTest, baseAttrsToTest...)
	return testAttrsCheck(datasourceName, resourceName, instanceAttrsToTest)
}

func testAccDataSourceComputeInstanceGroupFixedScaleCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	instanceAttrsToTest := []string{
		"scale_policy.0.fixed_scale.0.size",
	}

	instanceAttrsToTest = append(instanceAttrsToTest, baseAttrsToTest...)
	return testAttrsCheck(datasourceName, resourceName, instanceAttrsToTest)
}

func testAccDataSourceComputeInstanceGroupInstanceTagsPoolCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	instanceAttrsToTest := []string{
		"allocation_policy.0.instance_tags_pool.#",
		"allocation_policy.0.instance_tags_pool.0.zone",
		"allocation_policy.0.instance_tags_pool.0.tags",
		"allocation_policy.0.instance_tags_pool.1.zone",
		"allocation_policy.0.instance_tags_pool.1.tags",
		"allocation_policy.0.instance_tags_pool.2.zone",
		"allocation_policy.0.instance_tags_pool.2.tags",
	}

	instanceAttrsToTest = append(instanceAttrsToTest, baseAttrsToTest...)
	return testAttrsCheck(datasourceName, resourceName, instanceAttrsToTest)
}
