package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	functionssdk "github.com/yandex-cloud/go-sdk/services/serverless/functions/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_cloud_functions "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/functions/pluginsdk"
)

func registerResourceYandexFunction(add func(string, *resource.Sweeper)) {
	add("yandex_function", &resource.Sweeper{
		Name: "yandex_function",
		F:    sweepAllFunction,
		Dependencies: []string{
			"yandex_function_trigger",
		},
	})
}

func sweepAllFunction(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &functions.ListFunctionsRequest{FolderId: conf.FolderID}
	client := functionssdk.NewFunctionClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepFunction(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Function %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepFunction(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepFunctionOnce, conf, "Function", id)
}

func sweepFunctionOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_cloud_functions.YandexFunctionDefaultTimeout)
	defer cancel()

	client := functionssdk.NewFunctionClient(conf.SDK)

	op, err := client.Delete(ctx, &functions.DeleteFunctionRequest{
		FunctionId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
