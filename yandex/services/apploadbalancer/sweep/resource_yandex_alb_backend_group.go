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

func registerResourceYandexAlbBackendGroup(add func(string, *resource.Sweeper)) {
	add("yandex_alb_backend_group", &resource.Sweeper{
		Name: "yandex_alb_backend_group",
		F:    sweepAllALBBackendGroups,
		Dependencies: []string{
			"yandex_alb_http_router",
		},
	})
}

func sweepAllALBBackendGroups(_ string) error {
	log.Printf("[DEBUG] Sweeping Backend Group")
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}

	client := albsdk.NewBackendGroupClient(conf.SDK)

	req := &apploadbalancer.ListBackendGroupsRequest{FolderId: conf.FolderID}
	it := client.Iterator(conf.Context(), req)
	for it.Next() {
		id := it.Value().GetId()

		if !sweepALBBackendGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ALB Backend Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepALBBackendGroup(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepALBBackendGroupOnce, conf, "ALB Backend Group", id)
}

func sweepALBBackendGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_iam.YandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	client := albsdk.NewBackendGroupClient(conf.SDK)

	op, err := client.Delete(ctx, &apploadbalancer.DeleteBackendGroupRequest{
		BackendGroupId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
