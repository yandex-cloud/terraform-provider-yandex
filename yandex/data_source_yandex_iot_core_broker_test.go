package yandex

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const iotDataSourceBrokerResource = "data.yandex_iot_core_broker.test-brk-ds"

func TestAccYandexDataSourceIoTBroker_byID(t *testing.T) {
	t.Parallel()

	brokerName := acctest.RandomWithPrefix("tf-iot-core-broker")
	folderID := getExampleFolderID()
	brokerDesc := acctest.RandomWithPrefix("descriprion-for-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIoTBrokerConfigByID(brokerName, brokerDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "name", brokerName),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "description", brokerDesc),
					resource.TestCheckResourceAttrSet(iotDataSourceBrokerResource, "folder_id"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "folder_id", folderID),
					testAccCheckCreatedAtAttr(iotDataSourceBrokerResource),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "certificates.#", "0"),
				),
			},
		},
	})
}

func TestAccYandexDataSourceIoTBroker_byName(t *testing.T) {
	t.Parallel()

	brokerName := acctest.RandomWithPrefix("tf-iot-core-broker")
	folderID := getExampleFolderID()
	brokerDesc := acctest.RandomWithPrefix("descriprion-for-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIoTBrokerConfigByName(brokerName, brokerDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "name", brokerName),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "description", brokerDesc),
					resource.TestCheckResourceAttrSet(iotDataSourceBrokerResource, "folder_id"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "folder_id", folderID),
					testAccCheckCreatedAtAttr(iotDataSourceBrokerResource),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "certificates.#", "0"),
				),
			},
		},
	})
}

func TestAccYandexDataSourceIoTCoreBroker_full(t *testing.T) {
	t.Parallel()

	brokerName := acctest.RandomWithPrefix("tf-iot-core-broker")
	folderID := getExampleFolderID()

	cert, _ := ioutil.ReadFile("test-fixtures/iot/brk.pub")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreDataSourceBrokerFull(
					brokerName,
					"description",
					"label_key",
					"label",
					string(cert)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "name", brokerName),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "description", "description"),
					resource.TestCheckResourceAttrSet(iotDataSourceBrokerResource, "folder_id"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "labels.label_key", "label"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "certificates.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceIoTBrokerConfigByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_iot_core_broker" "test-brk-ds" {
  broker_id = "${yandex_iot_core_broker.test-brk.id}"
}

resource "yandex_iot_core_broker" "test-brk" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}

func testAccDataSourceIoTBrokerConfigByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_iot_core_broker" "test-brk-ds" {
  name = "${yandex_iot_core_broker.test-brk.name}"
}

resource "yandex_iot_core_broker" "test-brk" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}

func testYandexIoTCoreDataSourceBrokerFull(name string, descr string, labelKey string, label string, cert string) string {
	return templateConfig(`
data "yandex_iot_core_broker" "test-brk-ds" {
  broker_id = "${yandex_iot_core_broker.test-brk.id}"
}

resource "yandex_iot_core_broker" "test-brk" {
  name        = "{{.Name}}"
  description = "{{.Descr}}"
  labels = {
    {{.LabelKey}} = "{{.Label}}",
    empty-label   = ""
  }
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
		"Cert":     cert,
	})
}
