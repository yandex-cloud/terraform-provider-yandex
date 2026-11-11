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
	svc_iam "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/pluginsdk"
)

func registerResourceYandexAlbHttpRouter(add func(string, *resource.Sweeper)) {
	add("yandex_alb_http_router", &resource.Sweeper{
		Name: "yandex_alb_http_router",
		F:    sweepAllALBHTTPRouters,
		Dependencies: []string{
			"yandex_alb_load_balancer",
		},
	})
}

func sweepAllALBHTTPRouters(_ string) error {
	log.Printf("[DEBUG] Sweeping Http Router")
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}

	req := &apploadbalancer.ListHttpRoutersRequest{FolderId: conf.FolderID}
	client := albsdk.NewHttpRouterClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	for it.Next() {
		id := it.Value().GetId()

		if !sweepALBHTTPRouter(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ALB Http Router %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepALBHTTPRouter(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepALBHTTPRouterOnce, conf, "ALB Http Router", id)
}

func sweepALBHTTPRouterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_iam.YandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	client := albsdk.NewHttpRouterClient(conf.SDK)

	op, err := client.Delete(ctx, &apploadbalancer.DeleteHttpRouterRequest{
		HttpRouterId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
