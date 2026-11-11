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

func registerResourceYandexIotCoreRegistry(add func(string, *resource.Sweeper)) {
	add("yandex_iot_core_registry", &resource.Sweeper{
		Name: "yandex_iot_core_registry",
		F:    sweepAllIoTCoreRegistry,
		Dependencies: []string{
			"yandex_iot_core_device",
		},
	})
}

func sweepAllIoTCoreRegistry(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &iot.ListRegistriesRequest{FolderId: conf.FolderID}
	client := devicessdk.NewRegistryClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepIoTCoreRegistry(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep IoT Core Registry %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepIoTCoreRegistry(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepIoTCoreRegistryOnce, conf, "IoT Core Registry", id)
}

func sweepIoTCoreRegistryOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_iot_core.YandexIoTDefaultTimeout)
	defer cancel()

	client := devicessdk.NewRegistryClient(conf.SDK)

	op, err := client.Delete(ctx, &iot.DeleteRegistryRequest{
		RegistryId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
