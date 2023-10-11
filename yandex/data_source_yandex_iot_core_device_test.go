package yandex

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const iotDataSourceDeviceResource = "data.yandex_iot_core_device.test-dev-ds"

func TestAccYandexDataSourceIoTDevice_byID(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")
	deviceName := acctest.RandomWithPrefix("tf-iot-core-device")
	deviceDesc := acctest.RandomWithPrefix("description-for-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIoTDeviceConfigByID(registryName, deviceName, deviceDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "name", deviceName),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "description", deviceDesc),
					testAccCheckCreatedAtAttr(iotDataSourceDeviceResource),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "certificates.#", "0"),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "passwords.#", "0"),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "aliases.%", "0"),
				),
			},
		},
	})
}

func TestAccYandexDataSourceIoTDevice_byName(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")
	deviceName := acctest.RandomWithPrefix("tf-iot-core-device")
	deviceDesc := acctest.RandomWithPrefix("description-for-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIoTDeviceConfigByName(registryName, deviceName, deviceDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "name", deviceName),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "description", deviceDesc),
					testAccCheckCreatedAtAttr(iotDataSourceDeviceResource),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "certificates.#", "0"),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "passwords.#", "0"),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "aliases.%", "0"),
				),
			},
		},
	})
}

func TestAccYandexDataSourceIoTCoreDevice_full(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")
	deviceName := acctest.RandomWithPrefix("tf-iot-core-device")

	cert, _ := ioutil.ReadFile("test-fixtures/iot/dev.pub")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreDataSourceDeviceFull(
					registryName,
					deviceName,
					"description",
					"$devices/{id}/events",
					"aaa/bbb",
					"0123456789_default",
					string(cert)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "name", deviceName),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "description", "description"),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "certificates.#", "1"),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "passwords.#", "1"),
					resource.TestCheckResourceAttr(iotDataSourceDeviceResource, "aliases.%", "1"),
				),
			},
		},
	})
}

func testAccDataSourceIoTDeviceConfigByID(registryName, name, desc string) string {
	return fmt.Sprintf(`
data "yandex_iot_core_device" "test-dev-ds" {
  device_id = "${yandex_iot_core_device.test-dev.id}"
}

resource "yandex_iot_core_registry" "test-reg" {
  name = "%s"
}

resource "yandex_iot_core_device" "test-dev" {
  registry_id = "${yandex_iot_core_registry.test-reg.id}"
  name        = "%s"
  description = "%s"
}
`, registryName, name, desc)
}

func testAccDataSourceIoTDeviceConfigByName(registryName, name, desc string) string {
	return fmt.Sprintf(`
data "yandex_iot_core_device" "test-dev-ds" {
  name = "${yandex_iot_core_device.test-dev.name}"
}

resource "yandex_iot_core_registry" "test-reg" {
  name = "%s"
}

resource "yandex_iot_core_device" "test-dev" {
  registry_id = "${yandex_iot_core_registry.test-reg.id}"
  name        = "%s"
  description = "%s"
}
`, registryName, name, desc)
}

func testYandexIoTCoreDataSourceDeviceFull(registrtyName string, deviceName string, descr string, aliasKey string, aliasValue string, password string, cert string) string {
	return templateConfig(`
data "yandex_iot_core_device" "test-dev-ds" {
  name = "${yandex_iot_core_device.test-dev.name}"
}

resource "yandex_iot_core_registry" "test-reg" {
  name = "{{.RegName}}"
}

resource "yandex_iot_core_device" "test-dev" {
  registry_id = "${yandex_iot_core_registry.test-reg.id}"
  name        = "{{.Name}}"
  description = "{{.Descr}}"
  passwords = [
    "{{.Password}}"
  ]
  certificates = [<<EOF
{{.Cert}}
EOF
  ]
  aliases = {
    "{{.AliasValue}}" = "{{.AliasKey}}"
  }
}
	`, map[string]interface{}{
		"RegName":    registrtyName,
		"Name":       deviceName,
		"Descr":      descr,
		"AliasKey":   aliasKey,
		"AliasValue": aliasValue,
		"Password":   password,
		"Cert":       cert,
	})
}
