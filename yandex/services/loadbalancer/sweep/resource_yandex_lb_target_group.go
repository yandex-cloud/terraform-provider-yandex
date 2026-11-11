package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	loadbalancersdk "github.com/yandex-cloud/go-sdk/services/loadbalancer/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_iam "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/pluginsdk"
	"google.golang.org/grpc/codes"
)

func registerResourceYandexLbTargetGroup(add func(string, *resource.Sweeper)) {
	add("yandex_lb_target_group", &resource.Sweeper{
		Name:         "yandex_lb_target_group",
		F:            sweepAllLBTargetGroups,
		Dependencies: []string{},
	})
}

func sweepAllLBTargetGroups(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &loadbalancer.ListNetworkLoadBalancersRequest{FolderId: conf.FolderID}
	nlbClient := loadbalancersdk.NewNetworkLoadBalancerClient(conf.SDK)

	targetGroupClient := loadbalancersdk.NewTargetGroupClient(conf.SDK)

	nlbIt := nlbClient.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for nlbIt.Next() {
		nlbId := nlbIt.Value().GetId()
		for _, tg := range nlbIt.Value().GetAttachedTargetGroups() {
			tgId := tg.TargetGroupId
			if !sweepLBNetworkLoadBalancerAttachments(conf, nlbId, tgId) {
				result = multierror.Append(
					result, fmt.Errorf("failed to sweep Attached Target Group %q for Network Load Balancer %q", nlbId, tgId),
				)
			}
		}
	}
	if err := nlbIt.Error(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := result.ErrorOrNil(); err != nil {
		return err
	}

	reqTg := &loadbalancer.ListTargetGroupsRequest{FolderId: conf.FolderID}
	it := targetGroupClient.Iterator(conf.Context(), reqTg)
	for it.Next() {
		id := it.Value().GetId()
		if !sweepLBTargetGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep LB Target Group %q", id))
		}
	}
	if err := it.Error(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func sweepLBNetworkLoadBalancerAttachments(conf *providerconfig.Config, nlbId, tgId string) bool {
	return sweep.SweepWithRetryByFunc(
		conf, fmt.Sprintf("Attached Target Group %q for Network Load Balancer %q", nlbId, tgId),
		func(conf *providerconfig.Config) error {
			return sweepLBNetworkLoadBalancerAttachmentsOnce(conf, nlbId, tgId)
		},
	)
}

func sweepLBNetworkLoadBalancerAttachmentsOnce(conf *providerconfig.Config, nlbId, tgId string) error {
	ctx, cancel := conf.ContextWithTimeout(5 * time.Minute)
	defer cancel()

	client := loadbalancersdk.NewNetworkLoadBalancerClient(conf.SDK)

	op, err := client.DetachTargetGroup(
		ctx,
		&loadbalancer.DetachNetworkLoadBalancerTargetGroupRequest{
			NetworkLoadBalancerId: nlbId,
			TargetGroupId:         tgId,
		},
	)
	if err != nil {
		if sdkcommon.IsStatusWithCode(err, codes.NotFound) {
			return nil
		}
		return err
	}
	_, err = op.Wait(ctx)
	return err
}

func sweepLBTargetGroup(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepLBTargetGroupOnce, conf, "LB Target Group", id)
}

func sweepLBTargetGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_iam.YandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	client := loadbalancersdk.NewTargetGroupClient(conf.SDK)

	op, err := client.Delete(ctx, &loadbalancer.DeleteTargetGroupRequest{
		TargetGroupId: id,
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
