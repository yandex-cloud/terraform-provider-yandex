package sweep

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	connectionmanager "github.com/yandex-cloud/go-genproto/yandex/cloud/connectionmanager/v1"
	connectionmanagerv1sdk "github.com/yandex-cloud/go-sdk/services/connectionmanager/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func registerYandexConnectionmanagerConnectionResource(add func(string, *resource.Sweeper)) {
	add("yandex_connectionmanager_connection", &resource.Sweeper{
		Name:         "yandex_connectionmanager_connection",
		F:            sweepAllConnection,
		Dependencies: []string{},
	})
}

func sweepAllConnection(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	resp, err := connectionmanagerv1sdk.NewConnectionClient(conf.SDKv2).List(context.Background(), &connectionmanager.ListConnectionRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error getting Connection Manager connection: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Connection {
		if c.GetIsManaged() {
			continue
		}
		if !sweepConnection(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Connection Manager connection %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepConnection(conf *provider_config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepConnectionOnce, conf, "yandex_connectionmanager_connection", id)
}

func sweepConnectionOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	op, err := connectionmanagerv1sdk.NewConnectionClient(conf.SDKv2).Delete(ctx, &connectionmanager.DeleteConnectionRequest{
		ConnectionId: id,
	})
	_, err = op.Wait(context.Background())
	return err
}
