package sweep

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1/privatelink"
	privatelinksdk "github.com/yandex-cloud/go-sdk/services/vpc/v1/privatelink"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
	"google.golang.org/grpc/codes"
)

func registerResourceYandexVpcPrivateEndpoint(add func(string, *resource.Sweeper)) {
	add("yandex_vpc_private_endpoint", &resource.Sweeper{
		Name:         "yandex_vpc_private_endpoint",
		F:            sweepAllVPCPrivateEndpoints,
		Dependencies: []string{},
	})
}

func sweepAllVPCPrivateEndpoints(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &privatelink.ListPrivateEndpointsRequest{
		Container: &privatelink.ListPrivateEndpointsRequest_FolderId{
			FolderId: conf.FolderID,
		},
	}
	client := privatelinksdk.NewPrivateEndpointClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCPrivateEndpoint(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC Private Endpoint %q", id))
		}
	}
	if err := it.Error(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func sweepVPCPrivateEndpoint(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepVPCPrivateEndpointOnce, conf, "VPC Private Endpoint", id)
}

func sweepVPCPrivateEndpointOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_vpc.YandexVPCPrivateEndpointDefaultTimeout)
	defer cancel()

	client := privatelinksdk.NewPrivateEndpointClient(conf.SDK)

	op, err := client.Delete(ctx, &privatelink.DeletePrivateEndpointRequest{
		PrivateEndpointId: id,
	})
	if err != nil {
		if sdkcommon.IsStatusWithCode(err, codes.NotFound) {
			return nil
		}
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
