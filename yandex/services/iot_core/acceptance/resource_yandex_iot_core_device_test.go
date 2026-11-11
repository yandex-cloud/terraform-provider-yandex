package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	iotfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iot_core/testfixtures"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	devicessdk "github.com/yandex-cloud/go-sdk/services/iot/devices/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const iotRegistryResourceForDevices = "yandex_iot_core_registry.test-registry"

const iotDeviceResource = "yandex_iot_core_device.test-device"

func TestAccYandexIoTCoreDevice_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var registry iot.Registry
	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")

	var device iot.Device
	deviceName := acctest.RandomWithPrefix("tf-iot-core-device")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexIoTCoreDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreDeviceBasic(registryName, deviceName),
				Check: resource.ComposeTestCheckFunc(
					iotfixtures.CheckDeviceExists(suite.Config, iotRegistryResourceForDevices, iotDeviceResource, &registry, &device),
					resource.TestCheckResourceAttr(iotDeviceResource, "name", deviceName),
					resource.TestCheckResourceAttrSet(iotDeviceResource, "registry_id"),
					testYandexIoTCoreDeviceRegistryID(iotDeviceResource, &registry),
					testYandexIoTCoreDeviceContainsLabel(&device, "key", "value"),
					testYandexIoTCoreDeviceContainsLabel(&device, "empty", ""),
					statecheck.CheckCreatedAtAttr(iotDeviceResource),
				),
			},
		},
	})
}

func TestAccYandexIoTCoreDevice_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var registry iot.Registry
	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")

	var device iot.Device
	deviceName := acctest.RandomWithPrefix("tf-iot-core-device")

	authInfo := &yandexIotCoreAuth{}

	certDefault, _ := ioutil.ReadFile("testdata/dev_default.pub")
	certUpdated, _ := ioutil.ReadFile("testdata/dev_updated.pub")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexIoTCoreDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreDeviceFull(
					registryName,
					deviceName,
					"description",
					"$devices/{id}/events",
					"aaa/bbb",
					"0123456789_abcd",
					string(certDefault)),
				Check: resource.ComposeTestCheckFunc(
					iotfixtures.CheckDeviceExists(suite.Config, iotRegistryResourceForDevices, iotDeviceResource, &registry, &device),
					resource.TestCheckResourceAttr(iotDeviceResource, "name", deviceName),
					resource.TestCheckResourceAttr(iotDeviceResource, "description", "description"),
					suite.testYandexIoTCoreStoreDeviceCertificates(authInfo, &device),
					suite.testYandexIoTCoreStoreDevicePasswords(authInfo, &device),
					testYandexIoTCoreDeviceContainsAlias(&device, "$devices/{id}/events", "aaa/bbb"),
				),
			},
			{
				Config: testYandexIoTCoreDeviceFull(
					registryName,
					deviceName+"_updated",
					"description_updated",
					"$devices/{id}/events/updated",
					"aaa/bbb",
					"0123456789_updated",
					string(certDefault)),
				Check: resource.ComposeTestCheckFunc(
					iotfixtures.CheckDeviceExists(suite.Config, iotRegistryResourceForDevices, iotDeviceResource, &registry, &device),
					resource.TestCheckResourceAttr(iotDeviceResource, "name", deviceName+"_updated"),
					resource.TestCheckResourceAttr(iotDeviceResource, "description", "description_updated"),
					suite.testYandexIoTCoreNoChangeDeviceCertificates(authInfo, &device),
					suite.testYandexIoTCoreChangeDevicePasswords(authInfo, &device),
					suite.testYandexIoTCoreStoreDevicePasswords(authInfo, &device),
					testYandexIoTCoreDeviceContainsAlias(&device, "$devices/{id}/events/updated", "aaa/bbb"),
				),
			},
			{
				Config: testYandexIoTCoreDeviceFull(
					registryName,
					deviceName+"_updated",
					"description_updated",
					"$devices/{id}/events/updated",
					"aaa/bbb_updated",
					"0123456789_updated",
					string(certUpdated)),
				Check: resource.ComposeTestCheckFunc(
					iotfixtures.CheckDeviceExists(suite.Config, iotRegistryResourceForDevices, iotDeviceResource, &registry, &device),
					resource.TestCheckResourceAttr(iotDeviceResource, "name", deviceName+"_updated"),
					resource.TestCheckResourceAttr(iotDeviceResource, "description", "description_updated"),
					suite.testYandexIoTCoreChangeDeviceCertificates(authInfo, &device),
					suite.testYandexIoTCoreStoreDeviceCertificates(authInfo, &device),
					suite.testYandexIoTCoreNoChangeDevicePasswords(authInfo, &device),
					testYandexIoTCoreDeviceContainsAlias(&device, "$devices/{id}/events/updated", "aaa/bbb_updated"),
				),
			},
		},
	})
}

func testYandexIoTCoreDeviceRegistryID(name string, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		runFunc := resource.TestCheckResourceAttr(name, "registry_id", registry.Id)
		return runFunc(s)
	}
}

func testYandexIoTCoreDeviceContainsLabel(device *iot.Device, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := device.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func (suite testSuite) testYandexIoTCoreDeviceDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iot_core_device" {
			continue
		}

		_, err := iotfixtures.GetDeviceByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Device still exists")
		}
	}

	return nil
}

func (suite testSuite) testYandexIoTCoreStoreDeviceCertificates(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		authInfo.certificates, err = testGetDeviceCertificatesByID(providerConfig, device.Id)
		return err
	}
}

func (suite testSuite) testYandexIoTCoreStoreDevicePasswords(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		authInfo.passwords, err = testGetDevicePasswordsByID(providerConfig, device.Id)
		return err
	}
}

func (suite testSuite) testYandexIoTCoreNoChangeDeviceCertificates(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		certificatesNew, err := testGetDeviceCertificatesByID(providerConfig, device.Id)
		if err != nil {
			return err
		}
		if !testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must not be changed, but it is")
		}
		return nil
	}
}

func (suite testSuite) testYandexIoTCoreNoChangeDevicePasswords(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		passwordsNew, err := testGetDevicePasswordsByID(providerConfig, device.Id)
		if err != nil {
			return err
		}
		if !testIsEqualSets(authInfo.passwords, passwordsNew) {
			return fmt.Errorf("Passwords must not be changed, but it is")
		}
		return nil
	}
}

func testYandexIoTCoreDeviceContainsAlias(device *iot.Device, topic string, alias string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := device.TopicAliases[alias]
		if !ok {
			return fmt.Errorf("Expected alias with key '%s' not found", alias)
		}
		if v != topic {
			topic = strings.ReplaceAll(topic, "{id}", device.Id)
		}
		if v != topic {
			return fmt.Errorf("Incorrect alias value for key '%s': expected '%s' but found '%s'", alias, topic, v)
		}
		return nil
	}
}

func (suite testSuite) testYandexIoTCoreChangeDeviceCertificates(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		certificatesNew, err := testGetDeviceCertificatesByID(providerConfig, device.Id)
		if err != nil {
			return err
		}
		if testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must be changed, but it isn't")
		}
		return nil
	}
}

func (suite testSuite) testYandexIoTCoreChangeDevicePasswords(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		passwordsNew, err := testGetDevicePasswordsByID(providerConfig, device.Id)
		if err != nil {
			return err
		}
		if testIsEqualSets(authInfo.passwords, passwordsNew) {
			return fmt.Errorf("Passwords must be changed, but it isn't")
		}
		return nil
	}
}

func testYandexIoTCoreDeviceBasic(registryName string, deviceName string) string {
	return fmt.Sprintf(`
resource "yandex_iot_core_registry" "test-registry" {
  name = "%s"
}

resource "yandex_iot_core_device" "test-device" {
  registry_id = "${yandex_iot_core_registry.test-registry.id}"
  name        = "%s"
  labels = {
    key    = "value"
    empty  = ""
  }
}
	`, registryName, deviceName)
}

func testYandexIoTCoreDeviceFull(registrtyName string, deviceName string, descr string, aliasKey string, aliasValue string, password string, cert string) string {
	return sdkcommon.TemplateConfig(`
resource "yandex_iot_core_registry" "test-registry" {
  name = "{{.RegName}}"
}

resource "yandex_iot_core_device" "test-device" {
  registry_id = "${yandex_iot_core_registry.test-registry.id}"
  name        = "{{.Name}}"
  description = "{{.Descr}}"
  aliases = {
    "{{.AliasValue}}" = "{{.AliasKey}}"
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
		"RegName":    registrtyName,
		"Name":       deviceName,
		"Descr":      descr,
		"AliasKey":   aliasKey,
		"AliasValue": aliasValue,
		"Password":   password,
		"Cert":       cert,
	})
}

func testGetDeviceCertificatesByID(config *providerconfig.Config, ID string) (map[string]interface{}, error) {
	client := devicessdk.NewDeviceClient(config.SDK)

	certs, err := client.ListCertificates(context.Background(), &iot.ListDeviceCertificatesRequest{DeviceId: ID})
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, cert := range certs.Certificates {
		res[cert.CertificateData] = nil
	}
	return res, nil
}

func testGetDevicePasswordsByID(config *providerconfig.Config, ID string) (map[string]interface{}, error) {
	client := devicessdk.NewDeviceClient(config.SDK)

	passwords, err := client.ListPasswords(context.Background(), &iot.ListDevicePasswordsRequest{DeviceId: ID})
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, pass := range passwords.Passwords {
		res[pass.Id] = nil
	}
	return res, nil
}
