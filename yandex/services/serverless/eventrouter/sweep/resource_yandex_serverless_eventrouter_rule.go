package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	eventroutersdk "github.com/yandex-cloud/go-sdk/services/serverless/eventrouter/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_eventrouter "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/eventrouter/pluginsdk"
)

func registerResourceYandexServerlessEventrouterRule(add func(string, *resource.Sweeper)) {
	add("yandex_serverless_eventrouter_rule", &resource.Sweeper{
		Name: "yandex_serverless_eventrouter_rule",
		F:    sweepAllEventrouterRule,
	})
}

func sweepAllEventrouterRule(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &eventrouter.ListRulesRequest{
		ContainerId: &eventrouter.ListRulesRequest_FolderId{
			FolderId: conf.FolderID,
		},
	}

	client := eventroutersdk.NewRuleClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepEventrouterRule(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep sweep Event Router rule %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepEventrouterRule(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepEventrouterRuleOnce, conf, "Event Router rule", id)
}

func sweepEventrouterRuleOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_eventrouter.YandexEventrouterRuleDefaultTimeout)
	defer cancel()

	client := eventroutersdk.NewRuleClient(conf.SDK)

	op, err := client.Delete(ctx, &eventrouter.DeleteRuleRequest{
		RuleId: id,
	})
	if err != nil {
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
