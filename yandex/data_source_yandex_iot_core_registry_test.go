package yandex

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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

func testYandexIoTCoreDataSourceRegistryFull(name string, descr string, labelKey string, label string, password string, cert string) string {
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
  passwords = [
    "{{.Password}}"
  ]
  certificates = [<<EOF
{{.Cert}}
EOF
  ]
}
	`, map[string]interface{}{
		"Name":     name,
		"Descr":    descr,
		"LabelKey": labelKey,
		"Label":    label,
		"Password": password,
		"Cert":     cert,
	})
}
