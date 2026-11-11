package sweep

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
)

// NOTE(dxan): function may return non-empty string and non-nil error. Example:
// Resource is successfully created, but wait fails: the function returns id and wait error
func CreateVPCNetworkForSweeper(conf *providerconfig.Config) (string, error) {
	ctx, cancel := conf.ContextWithTimeout(svc_vpc.YandexVPCNetworkDefaultTimeout)
	defer cancel()
	client := vpcsdk.NewNetworkClient(conf.SDK)

	op, err := client.Create(ctx, &vpc.CreateNetworkRequest{
		FolderId:    conf.FolderID,
		Name:        acctest.RandomWithPrefix("sweeper"),
		Description: "created by sweeper",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create network: %v", err)
	}

	md := op.Metadata()
	sweep.DebugLog("Network '%s' was created, waiting for complete operation '%s'", md.GetNetworkId(), op.ID())

	_, err = op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("error while waiting for create subnet operation: %v", err)
	}

	return md.NetworkId, nil
}

func SweepVPCNetwork(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepVPCNetworkOnce, conf, "VPC Network", id)
}

func sweepVPCNetworkOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_vpc.YandexVPCNetworkDefaultTimeout)
	defer cancel()

	req := &vpc.ListNetworkSubnetsRequest{NetworkId: id}
	client := vpcsdk.NewNetworkClient(conf.SDK)

	subIt := client.SubnetsIterator(conf.Context(), req)
	for subIt.Next() {
		subID := subIt.Value().GetId()
		err := sweepVPCSubnetOnce(conf, subID)
		if err != nil {
			return err
		}
	}

	op, err := client.Delete(ctx, &vpc.DeleteNetworkRequest{
		NetworkId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}

func registerResourceYandexVpcNetwork(add func(string, *resource.Sweeper)) {
	add("yandex_vpc_network", &resource.Sweeper{
		Name: "yandex_vpc_network",
		F:    sweepAllVPCNetworks,
		Dependencies: []string{
			"yandex_vpc_subnet",
			"yandex_vpc_route_table",
			"yandex_vpc_security_group",
		},
	})
}

func sweepAllVPCNetworks(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListNetworksRequest{FolderId: conf.FolderID}
	client := vpcsdk.NewNetworkClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !SweepVPCNetwork(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC network %q", id))
		}
	}

	return result.ErrorOrNil()
}
