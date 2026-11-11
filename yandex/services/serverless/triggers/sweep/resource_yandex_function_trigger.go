package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	triggerssdk "github.com/yandex-cloud/go-sdk/services/serverless/triggers/v1"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_cloud_functions "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/functions/pluginsdk"
)

func registerResourceYandexFunctionTrigger(add func(string, *resource.Sweeper)) {
	add("yandex_function_trigger", &resource.Sweeper{
		Name: "yandex_function_trigger",
		F:    sweepAllFunctionTrigger,
	})
}

func sweepAllFunctionTrigger(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &triggers.ListTriggersRequest{FolderId: conf.FolderID}
	client := triggerssdk.NewTriggerClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepFunctionTrigger(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Function Trigger %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepFunctionTrigger(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepFunctionTriggerOnce, conf, "Function Trigger", id)
}

func sweepFunctionTriggerOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_cloud_functions.YandexFunctionDefaultTimeout)
	defer cancel()

	client := triggerssdk.NewTriggerClient(conf.SDK)

	op, err := client.Delete(ctx, &triggers.DeleteTriggerRequest{
		TriggerId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
