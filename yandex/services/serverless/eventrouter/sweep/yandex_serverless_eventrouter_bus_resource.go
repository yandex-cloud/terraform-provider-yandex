package sweep

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	eventrouterv1sdk "github.com/yandex-cloud/go-sdk/services/serverless/eventrouter/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func registerYandexServerlessEventrouterBusResource(add func(string, *resource.Sweeper)) {
	add("yandex_serverless_eventrouter_bus", &resource.Sweeper{
		Name: "yandex_serverless_eventrouter_bus",
		F:    sweepAllEventrouterBus,
	})
}

func sweepAllEventrouterBus(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &eventrouter.ListBusesRequest{FolderId: conf.ProviderState.FolderID.ValueString()}
	resp, err := eventrouterv1sdk.NewBusClient(conf.SDKv2).List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting buses: %s", err)
	}
	result := &multierror.Error{}
	for _, b := range resp.Buses {
		id := b.GetId()
		if !test.SweepWithRetry(sweepEventrouterBusOnce, conf, "Event Router bus", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep sweep Event Router bus %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepEventrouterBusOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	op, err := eventrouterv1sdk.NewBusClient(conf.SDKv2).Delete(ctx, &eventrouter.DeleteBusRequest{
		BusId: id,
	})
	_, err = op.Wait(ctx)
	return err
}
