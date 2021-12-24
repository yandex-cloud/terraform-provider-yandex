package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceComputeInstance_byID(t *testing.T) {
	t.Parallel()

	instanceName := fmt.Sprintf("data-instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeInstanceConfig(instanceName, true),
				Check: testAccDataSourceComputeInstanceCheck(
					"data.yandex_compute_instance.bar",
					"yandex_compute_instance.foo", instanceName),
			},
		},
	})
}

func TestAccDataSourceComputeInstance_byName(t *testing.T) {
	t.Parallel()

	instanceName := fmt.Sprintf("data-instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeInstanceConfig(instanceName, false),
				Check: testAccDataSourceComputeInstanceCheck(
					"data.yandex_compute_instance.bar",
					"yandex_compute_instance.foo", instanceName),
			},
		},
	})
}

func testAccDataSourceComputeInstanceAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
			"zone",
			"platform_id",
			"resources",
			"description",
			"labels",
			"metadata",
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

func TestAccDataSourceComputeInstance_GpusById(t *testing.T) {
	instanceName := fmt.Sprintf("data-instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComputeInstanceGpusConfig(instanceName, true),
				Check: testAccDataSourceComputeInstanceCheck(
					"data.yandex_compute_instance.bar",
					"yandex_compute_instance.foo", instanceName),
			},
		},
	})
}

func testAccDataSourceComputeInstanceCheck(datasourceName string, resourceName string, instanceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceComputeInstanceAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "instance_id"),
		resource.TestMatchResourceAttr(datasourceName, "fqdn", regexp.MustCompile(instanceName)),
		resource.TestCheckResourceAttr(datasourceName, "boot_disk.0.auto_delete", "true"),
		resource.TestCheckResourceAttr(datasourceName, "boot_disk.0.initialize_params.0.size", "4"),
		resource.TestCheckResourceAttr(datasourceName, "boot_disk.0.initialize_params.0.block_size", "8192"),
		resource.TestCheckResourceAttr(datasourceName, "boot_disk.0.initialize_params.0.type", "network-hdd"),
		resource.TestCheckResourceAttr(datasourceName, "network_interface.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "network_interface.0.nat", "false"),
		resource.TestCheckResourceAttr(datasourceName, "scheduling_policy.0.preemptible", "false"),
	)
}

func testAccDataSourceComputeInstanceResourceConfig(instanceName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foo" {
  name        = "%s"
  hostname    = "%s"
  platform_id = "standard-v2"
  description = "description"
  zone        = "ru-central1-a"

  resources {
    cores         = 2
    core_fraction = 20
    memory        = 2
  }

  boot_disk {
    initialize_params {
      size       = 4
      block_size = 8192
      image_id   = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instanceName, instanceName)
}

func testAccDataSourceComputeInstanceResourceGpusConfig(instanceName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foo" {
  name        = "%s"
  hostname    = "%s"
  description = "description"
  zone        = "ru-central1-b"
  platform_id = "gpu-standard-v1"

  resources {
    cores  = 8
    memory = 96
    gpus   = 1
  }

  boot_disk {
    initialize_params {
      size       = 4
      block_size = 8192
      image_id   = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}`, instanceName, instanceName)
}

const computeInstanceDataByIDConfig = `
data "yandex_compute_instance" "bar" {
  instance_id = "${yandex_compute_instance.foo.id}"
}
`

const computeInstanceDataByNameConfig = `
data "yandex_compute_instance" "bar" {
  name = "${yandex_compute_instance.foo.name}"
}
`

func testAccDataSourceComputeInstanceConfig(instanceName string, useDataID bool) string {
	if useDataID {
		return testAccDataSourceComputeInstanceResourceConfig(instanceName) + computeInstanceDataByIDConfig
	}

	return testAccDataSourceComputeInstanceResourceConfig(instanceName) + computeInstanceDataByNameConfig
}

func testAccDataSourceComputeInstanceGpusConfig(instanceName string, useDataID bool) string {
	if useDataID {
		return testAccDataSourceComputeInstanceResourceGpusConfig(instanceName) + computeInstanceDataByIDConfig
	}

	return testAccDataSourceComputeInstanceResourceGpusConfig(instanceName) + computeInstanceDataByNameConfig
}
