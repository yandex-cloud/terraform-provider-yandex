package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceComputeInstance(t *testing.T) {
	t.Parallel()

	instanceName := fmt.Sprintf("data-instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceComputeInstanceConfig(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceComputeInstanceCheck("data.yandex_compute_instance.bar", "yandex_compute_instance.foo"),
					resource.TestMatchResourceAttr("data.yandex_compute_instance.bar", "fqdn", regexp.MustCompile(instanceName)),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "boot_disk.0.auto_delete", "true"),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "boot_disk.0.initialize_params.0.size", "4"),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "boot_disk.0.initialize_params.0.type", "network-hdd"),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "network_interface.#", "1"),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "network_interface.0.nat", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeInstance_ipv6(t *testing.T) {
	t.Skip("waiting ipv6 support in subnets")
	t.Parallel()

	instanceName := fmt.Sprintf("data-instance-test-ipv6-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceComputeInstanceConfigIpv6(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceComputeInstanceCheck("data.yandex_compute_instance.bar", "yandex_compute_instance.foo"),
					resource.TestMatchResourceAttr("data.yandex_compute_instance.bar", "fqdn", regexp.MustCompile(instanceName)),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "boot_disk.0.auto_delete", "true"),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "boot_disk.0.initialize_params.0.size", "4"),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "boot_disk.0.initialize_params.0.type", "network-hdd"),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "network_interface.#", "1"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_instance.bar", "network_interface.0.ip_address"),
					resource.TestCheckResourceAttr("data.yandex_compute_instance.bar", "network_interface.0.ipv6", "true"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_instance.bar", "network_interface.0.ipv6_address"),
				),
			},
		},
	})
}

func testAccDataSourceComputeInstanceCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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

func testAccDataSourceComputeInstanceConfig(instanceName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foo" {
  name        = "%s"
  hostname    = "%s"
  description = "description"
  zone        = "ru-central1-a"

  resources {
    cores  = 1
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata {
    foo = "bar"
    baz = "qux"
  }

  metadata {
    startup-script = "echo Hello"
  }

  labels {
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

data "yandex_compute_instance" "bar" {
  instance_id = "${yandex_compute_instance.foo.id}"
}
`, instanceName, instanceName)
}

func testAccDataSourceComputeInstanceConfigIpv6(instanceName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foo" {
  name        = "%s"
  hostname    = "%s"
  description = "description"
  zone        = "ru-central1-a"

  resources {
    cores  = 1
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
    ipv6      = true
  }

  metadata {
    foo = "bar"
    baz = "qux"
  }

  metadata {
    startup-script = "echo Hello"
  }

  labels {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
  v6_cidr_blocks = ["fd00:aabb:ccdd:eeff::/64"]
}

data "yandex_compute_instance" "bar" {
  instance_id = "${yandex_compute_instance.foo.id}"
}
`, instanceName, instanceName)
}
