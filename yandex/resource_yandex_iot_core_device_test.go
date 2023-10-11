package yandex

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
)

const iotRegistryResourceForDevices = "yandex_iot_core_registry.test-registry"
const iotDeviceResource = "yandex_iot_core_device.test-device"

func init() {
	resource.AddTestSweepers("yandex_iot_core_device", &resource.Sweeper{
		Name: "yandex_iot_core_device",
		F:    testSweepIoTCoreDevice,
		Dependencies: []string{
			"yandex_function_trigger",
		},
	})
}

func testSweepIoTCoreDevice(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}
	for {
		resp, err := listDevicesWithRetry(conf, conf.FolderID)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("failed to list IoT Core Devices"))
			break
		}
		if len(resp.GetDevices()) == 0 {
			break
		}

		for _, device := range resp.GetDevices() {
			if !sweepIoTCoreDevice(conf, device.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep IoT Core Device %q", device.Id))
			}
		}
	}

	return result.ErrorOrNil()
}

func listDevicesWithRetry(conf *Config, folderId string) (resp *iot.ListDevicesResponse, err error) {
	for i := 1; i <= conf.MaxRetries; i++ {
		resp, err = conf.sdk.IoT().Devices().Device().List(conf.Context(), &iot.ListDevicesRequest{
			Id:        &iot.ListDevicesRequest_FolderId{FolderId: conf.FolderID},
			PageSize:  100,
			PageToken: "",
		})
		if err == nil {
			break
		}
	}
	return resp, err
}

func sweepIoTCoreDevice(conf *Config, id string) bool {
	return sweepWithRetry(sweepIoTCoreDeviceOnce, conf, "IoT Core Device", id)
}

func sweepIoTCoreDeviceOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexIoTDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.IoT().Devices().Device().Delete(ctx, &iot.DeleteDeviceRequest{
		DeviceId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexIoTCoreDevice_basic(t *testing.T) {
	t.Parallel()

	var registry iot.Registry
	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")

	var device iot.Device
	deviceName := acctest.RandomWithPrefix("tf-iot-core-device")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreDeviceBasic(registryName, deviceName),
				Check: resource.ComposeTestCheckFunc(
					testYandexIoTCoreDeviceExists(iotRegistryResourceForDevices, iotDeviceResource, &registry, &device),
					resource.TestCheckResourceAttr(iotDeviceResource, "name", deviceName),
					resource.TestCheckResourceAttrSet(iotDeviceResource, "registry_id"),
					testYandexIoTCoreDeviceRegistryID(iotDeviceResource, &registry),
					testAccCheckCreatedAtAttr(iotDeviceResource),
				),
			},
		},
	})
}

func TestAccYandexIoTCoreDevice_update(t *testing.T) {
	t.Parallel()

	var registry iot.Registry
	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")

	var device iot.Device
	deviceName := acctest.RandomWithPrefix("tf-iot-core-device")

	authInfo := &yandexIotCoreAuth{}

	certDefault, _ := ioutil.ReadFile("test-fixtures/iot/dev_default.pub")
	certUpdated, _ := ioutil.ReadFile("test-fixtures/iot/dev_updated.pub")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreDeviceDestroy,
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
					testYandexIoTCoreDeviceExists(iotRegistryResourceForDevices, iotDeviceResource, &registry, &device),
					resource.TestCheckResourceAttr(iotDeviceResource, "name", deviceName),
					resource.TestCheckResourceAttr(iotDeviceResource, "description", "description"),
					testYandexIoTCoreStoreDeviceCertificates(authInfo, &device),
					testYandexIoTCoreStoreDevicePasswords(authInfo, &device),
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
					testYandexIoTCoreDeviceExists(iotRegistryResourceForDevices, iotDeviceResource, &registry, &device),
					resource.TestCheckResourceAttr(iotDeviceResource, "name", deviceName+"_updated"),
					resource.TestCheckResourceAttr(iotDeviceResource, "description", "description_updated"),
					testYandexIoTCoreNoChangeDeviceCertificates(authInfo, &device),
					testYandexIoTCoreChangeDevicePasswords(authInfo, &device),
					testYandexIoTCoreStoreDevicePasswords(authInfo, &device),
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
					testYandexIoTCoreDeviceExists(iotRegistryResourceForDevices, iotDeviceResource, &registry, &device),
					resource.TestCheckResourceAttr(iotDeviceResource, "name", deviceName+"_updated"),
					resource.TestCheckResourceAttr(iotDeviceResource, "description", "description_updated"),
					testYandexIoTCoreChangeDeviceCertificates(authInfo, &device),
					testYandexIoTCoreStoreDeviceCertificates(authInfo, &device),
					testYandexIoTCoreNoChangeDevicePasswords(authInfo, &device),
					testYandexIoTCoreDeviceContainsAlias(&device, "$devices/{id}/events/updated", "aaa/bbb_updated"),
				),
			},
		},
	})
}

func testYandexIoTCoreDeviceExists(registryName string, deviceName string, registry *iot.Registry, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		registryFunc := testYandexIoTCoreRegistryExists(registryName, registry)
		err := registryFunc(s)
		if err != nil {
			return err
		}

		rs, ok := s.RootModule().Resources[deviceName]
		if !ok {
			return fmt.Errorf("Not found: %s", deviceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetDeviceByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("IoT Device not found")
		}

		*device = *found
		return nil
	}
}

func testYandexIoTCoreDeviceRegistryID(name string, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		runFunc := resource.TestCheckResourceAttr(name, "registry_id", registry.Id)
		return runFunc(s)
	}
}

func testYandexIoTCoreDeviceDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iot_core_device" {
			continue
		}

		_, err := testGetDeviceByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Device still exists")
		}
	}

	return nil
}

func testYandexIoTCoreStoreDeviceCertificates(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		authInfo.certificates, err = testGetDeviceCertificatesByID(testAccProvider.Meta().(*Config), device.Id)
		return err
	}
}

func testYandexIoTCoreStoreDevicePasswords(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		authInfo.passwords, err = testGetDevicePasswordsByID(testAccProvider.Meta().(*Config), device.Id)
		return err
	}
}

func testYandexIoTCoreNoChangeDeviceCertificates(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		certificatesNew, err := testGetDeviceCertificatesByID(testAccProvider.Meta().(*Config), device.Id)
		if err != nil {
			return err
		}
		if !testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must not be changed, but it is")
		}
		return nil
	}
}

func testYandexIoTCoreNoChangeDevicePasswords(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		passwordsNew, err := testGetDevicePasswordsByID(testAccProvider.Meta().(*Config), device.Id)
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

func testYandexIoTCoreChangeDeviceCertificates(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		certificatesNew, err := testGetDeviceCertificatesByID(testAccProvider.Meta().(*Config), device.Id)
		if err != nil {
			return err
		}
		if testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must be changed, but it isn't")
		}
		return nil
	}
}

func testYandexIoTCoreChangeDevicePasswords(authInfo *yandexIotCoreAuth, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		passwordsNew, err := testGetDevicePasswordsByID(testAccProvider.Meta().(*Config), device.Id)
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
}
	`, registryName, deviceName)
}

func testYandexIoTCoreDeviceFull(registrtyName string, deviceName string, descr string, aliasKey string, aliasValue string, password string, cert string) string {
	return templateConfig(`
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

func testGetDeviceByID(config *Config, ID string) (*iot.Device, error) {
	req := iot.GetDeviceRequest{
		DeviceId: ID,
	}

	return config.sdk.IoT().Devices().Device().Get(context.Background(), &req)
}

func testGetDeviceCertificatesByID(config *Config, ID string) (map[string]interface{}, error) {
	certs, err := config.sdk.IoT().Devices().Device().ListCertificates(context.Background(), &iot.ListDeviceCertificatesRequest{DeviceId: ID})
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, cert := range certs.Certificates {
		res[cert.CertificateData] = nil
	}
	return res, nil
}

func testGetDevicePasswordsByID(config *Config, ID string) (map[string]interface{}, error) {
	passwords, err := config.sdk.IoT().Devices().Device().ListPasswords(context.Background(), &iot.ListDevicePasswordsRequest{DeviceId: ID})
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, pass := range passwords.Passwords {
		res[pass.Id] = nil
	}
	return res, nil
}
