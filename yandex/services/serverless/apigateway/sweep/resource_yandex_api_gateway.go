package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	apigatewaysdk "github.com/yandex-cloud/go-sdk/services/serverless/apigateway/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_api_gateway "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/apigateway/pluginsdk"
)

func registerResourceYandexApiGateway(add func(string, *resource.Sweeper)) {
	add("yandex_api_gateway", &resource.Sweeper{
		Name:         "yandex_api_gateway",
		F:            sweepAllAPIGateway,
		Dependencies: []string{},
	})
}

func sweepAllAPIGateway(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &apigateway.ListApiGatewayRequest{FolderId: conf.FolderID}
	client := apigatewaysdk.NewApiGatewayClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepAPIGateway(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep API Gateway %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepAPIGateway(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepAPIGatewayOnce, conf, "API Gateway", id)
}

func sweepAPIGatewayOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_api_gateway.YandexApiGatewayDefaultTimeout)
	defer cancel()

	client := apigatewaysdk.NewApiGatewayClient(conf.SDK)

	op, err := client.Delete(ctx, &apigateway.DeleteApiGatewayRequest{
		ApiGatewayId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}

var specParametrized string
