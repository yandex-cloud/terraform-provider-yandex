package acceptance

import (
	"fmt"
	"io/ioutil"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

const iotDataSourceBrokerResource = "data.yandex_iot_core_broker.test-brk-ds"

func TestAccYandexDataSourceIoTBroker_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	brokerName := acctest.RandomWithPrefix("tf-iot-core-broker")
	folderID := acceptance.FolderID()
	brokerDesc := acctest.RandomWithPrefix("descriprion-for-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexIoTCoreBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIoTBrokerConfigByID(brokerName, brokerDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "name", brokerName),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "description", brokerDesc),
					resource.TestCheckResourceAttrSet(iotDataSourceBrokerResource, "folder_id"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "folder_id", folderID),
					statecheck.CheckCreatedAtAttr(iotDataSourceBrokerResource),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "certificates.#", "0"),
				),
			},
		},
	})
}

func TestAccYandexDataSourceIoTBroker_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	brokerName := acctest.RandomWithPrefix("tf-iot-core-broker")
	folderID := acceptance.FolderID()
	brokerDesc := acctest.RandomWithPrefix("descriprion-for-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexIoTCoreBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIoTBrokerConfigByName(brokerName, brokerDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "name", brokerName),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "description", brokerDesc),
					resource.TestCheckResourceAttrSet(iotDataSourceBrokerResource, "folder_id"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "folder_id", folderID),
					statecheck.CheckCreatedAtAttr(iotDataSourceBrokerResource),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "certificates.#", "0"),
				),
			},
		},
	})
}

func TestAccYandexDataSourceIoTCoreBroker_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	brokerName := acctest.RandomWithPrefix("tf-iot-core-broker")
	folderID := acceptance.FolderID()

	cert, _ := ioutil.ReadFile("testdata/brk.pub")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexIoTCoreBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreDataSourceBrokerFull(
					brokerName,
					"description",
					"label_key",
					"label",
					"ERROR",
					string(cert)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "name", brokerName),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "description", "description"),
					resource.TestCheckResourceAttrSet(iotDataSourceBrokerResource, "folder_id"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "labels.label_key", "label"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "certificates.#", "1"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "log_options.0.disabled", "false"),
					resource.TestCheckResourceAttr(iotDataSourceBrokerResource, "log_options.0.min_level", "ERROR"),
					resource.TestCheckResourceAttrSet(iotDataSourceBrokerResource, "log_options.0.log_group_id"),
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

func testYandexIoTCoreDataSourceBrokerFull(name string, descr string, labelKey string, label string, minLogLevel string, cert string) string {
	return sdkcommon.TemplateConfig(`
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
  log_options {
	log_group_id = yandex_logging_group.logging-group.id
	min_level = "{{.MinLogLevel}}"
  }
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
		"Cert":        cert,
	})
}
