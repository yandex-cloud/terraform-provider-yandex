package testfixtures

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	devicessdk "github.com/yandex-cloud/go-sdk/services/iot/devices/v1"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func CheckDeviceExists(configForTest func() (*providerconfig.Config, error), registryName string, deviceName string, registry *iot.Registry, device *iot.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		registryFunc := CheckRegistryExists(configForTest, registryName, registry)
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

		config, configErr := configForTest()
		if configErr != nil {
			return configErr
		}

		found, err := GetDeviceByID(config, rs.Primary.ID)
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

func GetDeviceByID(config *providerconfig.Config, ID string) (*iot.Device, error) {
	client := devicessdk.NewDeviceClient(config.SDK)

	req := iot.GetDeviceRequest{
		DeviceId: ID,
	}

	return client.Get(context.Background(), &req)
}
func CheckRegistryExists(configForTest func() (*providerconfig.Config, error), name string, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := configForTest()
		if configErr != nil {
			return configErr
		}

		found, err := GetRegistryByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("IoT Registry not found")
		}

		*registry = *found
		return nil
	}
}

func GetRegistryByID(config *providerconfig.Config, ID string) (*iot.Registry, error) {
	client := devicessdk.NewRegistryClient(config.SDK)

	req := iot.GetRegistryRequest{
		RegistryId: ID,
	}

	return client.Get(context.Background(), &req)
}
