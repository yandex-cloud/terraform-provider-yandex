package sweep

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
)

func registerResourceYandexVpcGateway(add func(string, *resource.Sweeper)) {
	add("yandex_vpc_gateway", &resource.Sweeper{
		Name: "yandex_vpc_gateway",
		F:    sweepAllVPCGateways,
		Dependencies: []string{
			"yandex_vpc_route_table",
		},
	})
}

func sweepAllVPCGateways(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListGatewaysRequest{FolderId: conf.FolderID}
	client := vpcsdk.NewGatewayClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCGateway(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC gateway %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepVPCGateway(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepVPCGatewayOnce, conf, "VPC Gateway", id)
}

func sweepVPCGatewayOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_vpc.YandexVPCGatewayDefaultTimeout)
	defer cancel()

	client := vpcsdk.NewGatewayClient(conf.SDK)

	op, err := client.Delete(ctx, &vpc.DeleteGatewayRequest{
		GatewayId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
