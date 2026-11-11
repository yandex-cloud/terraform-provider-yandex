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

func registerResourceYandexServerlessEventrouterConnector(add func(string, *resource.Sweeper)) {
	add("yandex_serverless_eventrouter_connector", &resource.Sweeper{
		Name: "yandex_serverless_eventrouter_connector",
		F:    sweepAllEventrouterConnector,
	})
}

func sweepAllEventrouterConnector(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &eventrouter.ListConnectorsRequest{
		ContainerId: &eventrouter.ListConnectorsRequest_FolderId{
			FolderId: conf.FolderID,
		},
	}

	client := eventroutersdk.NewConnectorClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepEventrouterConnector(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep sweep Event Router connector %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepEventrouterConnector(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepEventrouterConnectorOnce, conf, "Event Router connector", id)
}

func sweepEventrouterConnectorOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_eventrouter.YandexEventrouterConnectorDefaultTimeout)
	defer cancel()

	client := eventroutersdk.NewConnectorClient(conf.SDK)

	op, err := client.Delete(ctx, &eventrouter.DeleteConnectorRequest{
		ConnectorId: id,
	})
	if err != nil {
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
