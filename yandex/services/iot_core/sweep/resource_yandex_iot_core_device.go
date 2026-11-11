package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	devicessdk "github.com/yandex-cloud/go-sdk/services/iot/devices/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_iot_core "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iot_core/pluginsdk"
)

func registerResourceYandexIotCoreDevice(add func(string, *resource.Sweeper)) {
	add("yandex_iot_core_device", &resource.Sweeper{
		Name: "yandex_iot_core_device",
		F:    sweepAllIoTCoreDevice,
		Dependencies: []string{
			"yandex_function_trigger",
		},
	})
}

func sweepAllIoTCoreDevice(_ string) error {
	conf, err := sweep.ConfigForSweepers()
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

func listDevicesWithRetry(conf *providerconfig.Config, folderId string) (resp *iot.ListDevicesResponse, err error) {
	client := devicessdk.NewDeviceClient(conf.SDK)

	for i := 1; i <= conf.MaxRetries; i++ {
		resp, err = client.List(conf.Context(), &iot.ListDevicesRequest{
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

func sweepIoTCoreDevice(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepIoTCoreDeviceOnce, conf, "IoT Core Device", id)
}

func sweepIoTCoreDeviceOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_iot_core.YandexIoTDefaultTimeout)
	defer cancel()
	client := devicessdk.NewDeviceClient(conf.SDK)

	op, err := client.Delete(ctx, &iot.DeleteDeviceRequest{
		DeviceId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
