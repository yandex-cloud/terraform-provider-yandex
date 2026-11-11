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

func registerResourceYandexAlbTargetGroup(add func(string, *resource.Sweeper)) {
	add("yandex_alb_target_group", &resource.Sweeper{
		Name: "yandex_alb_target_group",
		F:    sweepAllALBTargetGroups,
		Dependencies: []string{
			"yandex_alb_backend_group",
		},
	})
}

func sweepAllALBTargetGroups(_ string) error {
	log.Printf("[DEBUG] Sweeping TargetGroup")
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}

	req := &apploadbalancer.ListTargetGroupsRequest{FolderId: conf.FolderID}
	client := albsdk.NewTargetGroupClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	for it.Next() {
		id := it.Value().GetId()

		if !sweepALBTargetGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ALB Target Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepALBTargetGroup(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepALBTargetGroupOnce, conf, "ALB Target Group", id)
}

func sweepALBTargetGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_iam.YandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	client := albsdk.NewTargetGroupClient(conf.SDK)

	op, err := client.Delete(ctx, &apploadbalancer.DeleteTargetGroupRequest{
		TargetGroupId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
