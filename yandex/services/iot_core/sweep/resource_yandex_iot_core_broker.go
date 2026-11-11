package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/broker/v1"
	brokersdk "github.com/yandex-cloud/go-sdk/services/iot/broker/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_iot_core "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iot_core/pluginsdk"
)

func registerResourceYandexIotCoreBroker(add func(string, *resource.Sweeper)) {
	add("yandex_iot_core_broker", &resource.Sweeper{
		Name:         "yandex_iot_core_broker",
		F:            sweepAllIoTCoreBroker,
		Dependencies: []string{},
	})
}

func sweepAllIoTCoreBroker(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &iot.ListBrokersRequest{FolderId: conf.FolderID}
	client := brokersdk.NewBrokerClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepIoTCoreBroker(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep IoT Core Broker %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepIoTCoreBroker(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepIoTCoreBrokerOnce, conf, "IoT Core Broker", id)
}

func sweepIoTCoreBrokerOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_iot_core.YandexIoTDefaultTimeout)
	defer cancel()

	client := brokersdk.NewBrokerClient(conf.SDK)

	op, err := client.Delete(ctx, &iot.DeleteBrokerRequest{
		BrokerId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
