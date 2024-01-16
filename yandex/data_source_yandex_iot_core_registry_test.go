package yandex

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const iotDataSourceResource = "data.yandex_iot_core_registry.test-reg-ds"

func TestAccYandexDataSourceIoTRegistry_byID(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")
	folderID := getExampleFolderID()
	registryDesc := acctest.RandomWithPrefix("descriprion-for-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIoTRegistryConfigByID(registryName, registryDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceResource, "name", registryName),
					resource.TestCheckResourceAttr(iotDataSourceResource, "description", registryDesc),
					resource.TestCheckResourceAttrSet(iotDataSourceResource, "folder_id"),
					resource.TestCheckResourceAttr(iotDataSourceResource, "folder_id", folderID),
					testAccCheckCreatedAtAttr(iotDataSourceResource),
					resource.TestCheckResourceAttr(iotDataSourceResource, "certificates.#", "0"),
					resource.TestCheckResourceAttr(iotDataSourceResource, "passwords.#", "0"),
				),
			},
		},
	})
}

func TestAccYandexDataSourceIoTRegistry_byName(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")
	folderID := getExampleFolderID()
	registryDesc := acctest.RandomWithPrefix("descriprion-for-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIoTRegistryConfigByName(registryName, registryDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceResource, "name", registryName),
					resource.TestCheckResourceAttr(iotDataSourceResource, "description", registryDesc),
					resource.TestCheckResourceAttrSet(iotDataSourceResource, "folder_id"),
					resource.TestCheckResourceAttr(iotDataSourceResource, "folder_id", folderID),
					testAccCheckCreatedAtAttr(iotDataSourceResource),
					resource.TestCheckResourceAttr(iotDataSourceResource, "certificates.#", "0"),
					resource.TestCheckResourceAttr(iotDataSourceResource, "passwords.#", "0"),
				),
			},
		},
	})
}

func TestAccYandexDataSourceIoTCoreRegistry_full(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")
	folderID := getExampleFolderID()

	cert, _ := ioutil.ReadFile("test-fixtures/iot/reg.pub")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreDataSourceRegistryFull(
					registryName,
					"description",
					"label_key",
					"label",
					"ERROR",
					"0123456789_abcd",
					string(cert)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceResource, "name", registryName),
					resource.TestCheckResourceAttr(iotDataSourceResource, "description", "description"),
					resource.TestCheckResourceAttrSet(iotDataSourceResource, "folder_id"),
					resource.TestCheckResourceAttr(iotDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(iotDataSourceResource, "labels.label_key", "label"),
					resource.TestCheckResourceAttr(iotDataSourceResource, "certificates.#", "1"),
					resource.TestCheckResourceAttr(iotDataSourceResource, "passwords.#", "1"),
					resource.TestCheckResourceAttr(iotDataSourceResource, "log_options.0.disabled", "false"),
					resource.TestCheckResourceAttr(iotDataSourceResource, "log_options.0.min_level", "ERROR"),
					resource.TestCheckResourceAttrSet(iotDataSourceResource, "log_options.0.log_group_id"),
				),
			},
		},
	})
}

func testAccDataSourceIoTRegistryConfigByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_iot_core_registry" "test-reg-ds" {
  registry_id = "${yandex_iot_core_registry.test-reg.id}"
}

resource "yandex_iot_core_registry" "test-reg" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}

func testAccDataSourceIoTRegistryConfigByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_iot_core_registry" "test-reg-ds" {
  name = "${yandex_iot_core_registry.test-reg.name}"
}

resource "yandex_iot_core_registry" "test-reg" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}

func testYandexIoTCoreDataSourceRegistryFull(name string, descr string, labelKey string, label string, minLogLevel string, password string, cert string) string {
	return templateConfig(`
data "yandex_iot_core_registry" "test-reg-ds" {
  registry_id = "${yandex_iot_core_registry.test-reg.id}"
}

resource "yandex_iot_core_registry" "test-reg" {
  name        = "{{.Name}}"
  description = "{{.Descr}}"
  labels = {
    {{.LabelKey}} = "{{.Label}}",
    empty-label   = ""
  }
  log_options {
	log_group_id = yandex_logging_group.logging-group.id
	min_level = "{{.MinLogLevel}}"
  }
  passwords = [
    "{{.Password}}"
  ]
  certificates = [<<EOF
{{.Cert}}
EOF
  ]
}

resource "yandex_logging_group" "logging-group" {
}
	`, map[string]interface{}{
		"Name":        name,
		"Descr":       descr,
		"LabelKey":    labelKey,
		"Label":       label,
		"MinLogLevel": minLogLevel,
		"Password":    password,
		"Cert":        cert,
	})
}
