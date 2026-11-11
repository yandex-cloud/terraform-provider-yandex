package sweep

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	albsdk "github.com/yandex-cloud/go-sdk/services/apploadbalancer/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_alb "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/apploadbalancer/pluginsdk"
)

func registerResourceYandexAlbLoadBalancer(add func(string, *resource.Sweeper)) {
	add("yandex_alb_load_balancer", &resource.Sweeper{
		Name:         "yandex_alb_load_balancer",
		F:            sweepAllALBLoadBalancers,
		Dependencies: []string{},
	})
}

func sweepAllALBLoadBalancers(_ string) error {
	log.Printf("[DEBUG] Sweeping ALB Load Balancer")
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}

	client := albsdk.NewLoadBalancerClient(conf.SDK)

	req := &apploadbalancer.ListLoadBalancersRequest{FolderId: conf.FolderID}
	it := client.Iterator(conf.Context(), req)
	for it.Next() {
		id := it.Value().GetId()

		if !sweepALBLoadBalancer(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ALB Load Balancer %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepALBLoadBalancer(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepALBLoadBalancerOnce, conf, "ALB Load Balancer", id)
}

func sweepALBLoadBalancerOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_alb.YandexALBLoadBalancerDefaultTimeout)
	defer cancel()

	client := albsdk.NewLoadBalancerClient(conf.SDK)

	op, err := client.Delete(ctx, &apploadbalancer.DeleteLoadBalancerRequest{
		LoadBalancerId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
