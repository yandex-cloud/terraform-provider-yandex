package yandex

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/broker/v1"
)

const iotBrokerResource = "yandex_iot_core_broker.test-broker"

func init() {
	resource.AddTestSweepers("yandex_iot_core_broker", &resource.Sweeper{
		Name:         "yandex_iot_core_broker",
		F:            testSweepIoTCoreBroker,
		Dependencies: []string{},
	})
}

func testSweepIoTCoreBroker(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &iot.ListBrokersRequest{FolderId: conf.FolderID}
	it := conf.sdk.IoT().Broker().Broker().BrokerIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepIoTCoreBroker(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep IoT Core Broker %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepIoTCoreBroker(conf *Config, id string) bool {
	return sweepWithRetry(sweepIoTCoreBrokerOnce, conf, "IoT Core Broker", id)
}

func sweepIoTCoreBrokerOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexIoTDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.IoT().Broker().Broker().Delete(ctx, &iot.DeleteBrokerRequest{
		BrokerId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexIoTCoreBroker_basic(t *testing.T) {
	t.Parallel()

	var broker iot.Broker
	brokerName := acctest.RandomWithPrefix("tf-iot-core-broker")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreBrokerBasic(brokerName),
				Check: resource.ComposeTestCheckFunc(
					testYandexIoTCoreBrokerExists(iotBrokerResource, &broker),
					resource.TestCheckResourceAttr(iotBrokerResource, "name", brokerName),
					resource.TestCheckResourceAttrSet(iotBrokerResource, "folder_id"),
					resource.TestCheckResourceAttr(iotBrokerResource, "folder_id", folderID),
					testYandexIoTCoreBrokerContainsLabel(&broker, "tf-label", "tf-label-value"),
					testYandexIoTCoreBrokerContainsLabel(&broker, "empty-label", ""),
					testAccCheckCreatedAtAttr(iotBrokerResource),
				),
			},
		},
	})
}

func TestAccYandexIoTCoreBroker_update(t *testing.T) {
	t.Parallel()

	var broker iot.Broker
	brokerName := acctest.RandomWithPrefix("tf-iot-core-broker")

	authInfo := &yandexIotCoreAuth{}

	certDefault, _ := ioutil.ReadFile("test-fixtures/iot/brk_default.pub")
	certUpdated, _ := ioutil.ReadFile("test-fixtures/iot/brk_updated.pub")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreBrokerFull(
					brokerName,
					"description",
					"label_key",
					"label",
					string(certDefault)),
				Check: resource.ComposeTestCheckFunc(
					testYandexIoTCoreBrokerExists(iotBrokerResource, &broker),
					resource.TestCheckResourceAttr(iotBrokerResource, "name", brokerName),
					resource.TestCheckResourceAttr(iotBrokerResource, "description", "description"),
					testYandexIoTCoreBrokerContainsLabel(&broker, "label_key", "label"),
					testYandexIoTCoreStoreBrokerCertificates(authInfo, &broker),
				),
			},
			{
				Config: testYandexIoTCoreBrokerFull(
					brokerName+"_updated",
					"description_updated",
					"label_key_updated",
					"label_updated",
					string(certDefault)),
				Check: resource.ComposeTestCheckFunc(
					testYandexIoTCoreBrokerExists(iotBrokerResource, &broker),
					resource.TestCheckResourceAttr(iotBrokerResource, "name", brokerName+"_updated"),
					resource.TestCheckResourceAttr(iotBrokerResource, "description", "description_updated"),
					testYandexIoTCoreBrokerContainsLabel(&broker, "label_key_updated", "label_updated"),
					testYandexIoTCoreNoChangeBrokerCertificates(authInfo, &broker),
				),
			},
			{
				Config: testYandexIoTCoreBrokerFull(
					brokerName+"_updated",
					"description_updated",
					"label_key_updated",
					"label_updated",
					string(certUpdated)),
				Check: resource.ComposeTestCheckFunc(
					testYandexIoTCoreBrokerExists(iotBrokerResource, &broker),
					resource.TestCheckResourceAttr(iotBrokerResource, "name", brokerName+"_updated"),
					resource.TestCheckResourceAttr(iotBrokerResource, "description", "description_updated"),
					testYandexIoTCoreBrokerContainsLabel(&broker, "label_key_updated", "label_updated"),
					testYandexIoTCoreChangeBrokerCertificates(authInfo, &broker),
					testYandexIoTCoreStoreBrokerCertificates(authInfo, &broker),
				),
			},
		},
	})
}

func testYandexIoTCoreBrokerExists(name string, broker *iot.Broker) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetBrokerByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("IoT Broker not found")
		}

		*broker = *found
		return nil
	}
}

func testYandexIoTCoreBrokerDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iot_core_broker" {
			continue
		}

		_, err := testGetBrokerByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Broker still exists")
		}
	}

	return nil
}

func testYandexIoTCoreBrokerContainsLabel(broker *iot.Broker, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := broker.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexIoTCoreStoreBrokerCertificates(authInfo *yandexIotCoreAuth, broker *iot.Broker) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		authInfo.certificates, err = testGetBrokerCertificatesByID(testAccProvider.Meta().(*Config), broker.Id)
		return err
	}
}

func testYandexIoTCoreNoChangeBrokerCertificates(authInfo *yandexIotCoreAuth, broker *iot.Broker) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		certificatesNew, err := testGetBrokerCertificatesByID(testAccProvider.Meta().(*Config), broker.Id)
		if err != nil {
			return err
		}
		if !testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must not be changed, but it is")
		}
		return nil
	}
}

func testYandexIoTCoreChangeBrokerCertificates(authInfo *yandexIotCoreAuth, broker *iot.Broker) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		certificatesNew, err := testGetBrokerCertificatesByID(testAccProvider.Meta().(*Config), broker.Id)
		if err != nil {
			return err
		}
		if testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must be changed, but it isn't")
		}
		return nil
	}
}

func testYandexIoTCoreBrokerBasic(name string) string {
	return fmt.Sprintf(`
resource "yandex_iot_core_broker" "test-broker" {
  name = "%s"
  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
	`, name)
}

func testYandexIoTCoreBrokerFull(name string, descr string, labelKey string, label string, cert string) string {
	return templateConfig(`
resource "yandex_iot_core_broker" "test-broker" {
  name        = "{{.Name}}"
  description = "{{.Descr}}"
  labels = {
    {{.LabelKey}} = "{{.Label}}"
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

func testGetBrokerByID(config *Config, ID string) (*iot.Broker, error) {
	req := iot.GetBrokerRequest{
		BrokerId: ID,
	}

	return config.sdk.IoT().Broker().Broker().Get(context.Background(), &req)
}

func testGetBrokerCertificatesByID(config *Config, ID string) (map[string]interface{}, error) {
	certs, err := config.sdk.IoT().Broker().Broker().ListCertificates(context.Background(), &iot.ListBrokerCertificatesRequest{BrokerId: ID})
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, cert := range certs.Certificates {
		res[cert.CertificateData] = nil
	}
	return res, nil
}
