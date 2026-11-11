package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	loadbalancersdk "github.com/yandex-cloud/go-sdk/services/loadbalancer/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_network_load_balancer "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/loadbalancer/pluginsdk"
	"google.golang.org/grpc/codes"
)

func registerResourceYandexLbNetworkLoadBalancer(add func(string, *resource.Sweeper)) {
	add("yandex_lb_network_load_balancer", &resource.Sweeper{
		Name: "yandex_lb_network_load_balancer",
		F:    sweepAllLBNetworkLoadBalancers,
		Dependencies: []string{
			"yandex_lb_target_group",
		},
	})
}

func sweepAllLBNetworkLoadBalancers(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &loadbalancer.ListNetworkLoadBalancersRequest{FolderId: conf.FolderID}
	client := loadbalancersdk.NewNetworkLoadBalancerClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepLBNetworkLoadBalancer(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Network Load Balancer %q", id))
		}
	}
	if err := it.Error(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func sweepLBNetworkLoadBalancer(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepLBNetworkLoadBalancerOnce, conf, "Network Load Balancer", id)
}

func sweepLBNetworkLoadBalancerOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_network_load_balancer.YandexLBNetworkLoadBalancerDefaultTimeout)
	defer cancel()

	client := loadbalancersdk.NewNetworkLoadBalancerClient(conf.SDK)

	op, err := client.Delete(ctx, &loadbalancer.DeleteNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: id,
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
