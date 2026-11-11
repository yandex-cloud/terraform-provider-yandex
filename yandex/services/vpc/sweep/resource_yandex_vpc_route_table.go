package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
)

func registerResourceYandexVpcRouteTable(add func(string, *resource.Sweeper)) {
	add("yandex_vpc_route_table", &resource.Sweeper{
		Name: "yandex_vpc_route_table",
		F:    sweepAllVPCRouteTables,
	})
}

func sweepAllVPCRouteTables(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListRouteTablesRequest{FolderId: conf.FolderID}
	client := vpcsdk.NewRouteTableClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCRouteTable(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC Route Table %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepVPCRouteTable(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepVPCRouteTableOnce, conf, "VPC Route Table", id)
}

func sweepVPCRouteTableOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_vpc.YandexVPCRouteTableDefaultTimeout)
	defer cancel()

	client := vpcsdk.NewRouteTableClient(conf.SDK)

	op, err := client.Delete(ctx, &vpc.DeleteRouteTableRequest{
		RouteTableId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
