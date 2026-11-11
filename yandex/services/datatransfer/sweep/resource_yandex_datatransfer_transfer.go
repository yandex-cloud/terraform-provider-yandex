package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1"
	datatransfersdk "github.com/yandex-cloud/go-sdk/services/datatransfer/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerResourceYandexDatatransferTransfer(add func(string, *resource.Sweeper)) {
	add("yandex_datatransfer", &resource.Sweeper{
		Name: "yandex_datatransfer",
		F:    sweepAllDataTransfer,
	})
}

func sweepAllDataTransfer(_ string) error {
	config, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting config: %s", err)
	}
	transferClient := datatransfersdk.NewTransferClient(config.SDK)

	endpointClient := datatransfersdk.NewEndpointClient(config.SDK)

	listTransfersResponse, err := transferClient.List(config.Context(), &datatransfer.ListTransfersRequest{FolderId: config.FolderID})
	if err != nil {
		return fmt.Errorf("error getting DataTransfer transfers: %s", err)
	}

	resultError := &multierror.Error{}
	for _, transfer := range listTransfersResponse.Transfers {
		if !sweepDataTransferTransfer(config, transfer.Id) {
			resultError = multierror.Append(resultError, fmt.Errorf("failed to sweep DataTransfer transfer %q", transfer.Id))
		}
	}
	if resultError.Len() > 0 {
		return resultError
	}

	listEndpointsResponse, err := endpointClient.List(config.Context(), &datatransfer.ListEndpointsRequest{FolderId: config.FolderID})
	if err != nil {
		return fmt.Errorf("error getting DataTransfer transfers: %s", err)
	}
	for _, endpoint := range listEndpointsResponse.Endpoints {
		if !sweepDataTransferEndpoint(config, endpoint.Id) {
			resultError = multierror.Append(resultError, fmt.Errorf("failed to sweep "))
		}
	}
	return resultError.ErrorOrNil()
}

func sweepDataTransferTransfer(config *providerconfig.Config, transferID string) bool {
	return sweep.SweepWithRetry(sweepDataTransferTransferOnce, config, "DataTransfer transfer", transferID)
}

func sweepDataTransferTransferOnce(config *providerconfig.Config, transferID string) error {
	ctx, cancel := config.ContextWithTimeout(yandexDataTransferTransferDeleteTimeout)
	defer cancel()

	client := datatransfersdk.NewTransferClient(config.SDK)

	deleteOperation, err := client.Delete(ctx, &datatransfer.DeleteTransferRequest{TransferId: transferID})
	if err != nil {
		return err
	}
	_, err = deleteOperation.Wait(ctx)
	return err
}

const (
	yandexDataTransferTransferDeleteTimeout = 5 * time.Minute

	syncActivateMode  = "sync_activate"
	asyncActivateMode = "async_activate"
	dontActivateMode  = "dont_activate"

	sourceEndpointResourceName = "yandex_datatransfer_endpoint.pg_source"
	targetEndpointResourceName = "yandex_datatransfer_endpoint.pg_target"
	transferResourceName       = "yandex_datatransfer_transfer.pgpg_transfer"
)

func sweepDataTransferEndpoint(config *providerconfig.Config, transferID string) bool {
	return sweep.SweepWithRetry(sweepDataTransferEndpointOnce, config, "DataTransfer endpoint", transferID)
}

func sweepDataTransferEndpointOnce(config *providerconfig.Config, endpointID string) error {
	ctx, cancel := config.ContextWithTimeout(yandexDataTransferTransferDeleteTimeout)
	defer cancel()

	client := datatransfersdk.NewEndpointClient(config.SDK)

	deleteOperation, err := client.Delete(ctx, &datatransfer.DeleteEndpointRequest{EndpointId: endpointID})
	if err != nil {
		return err
	}
	_, err = deleteOperation.Wait(ctx)
	return err
}
