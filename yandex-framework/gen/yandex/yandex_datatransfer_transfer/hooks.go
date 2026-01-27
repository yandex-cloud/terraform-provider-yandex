package yandex_datatransfer_transfer

import (
	fmt "fmt"

	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	datatransfer "github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1"
	datatransferv1sdk "github.com/yandex-cloud/go-sdk/services/datatransfer/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// possible scenarios for activating SNAPSHOT_AND_INCREMENT and SNAPSHOT_ONLY
	// transfers when created or re-created through a Terraform provider.
	syncActivateMode  = "sync_activate"
	asyncActivateMode = "async_activate"
	dontActivateMode  = "dont_activate"

	traceIDMetadataKey   = "x-server-trace-id"
	requestIDMetadataKey = "x-server-request-id"
)

func PatchIDsAfterRead(ctx context.Context, config *config.Config, resp *datatransfer.Transfer, model *yandexDatatransferTransferModel) diag.Diagnostics {
	model.SourceId = types.StringValue(resp.GetSource().GetId())
	model.TargetId = types.StringValue(resp.GetTarget().GetId())
	if model.OnCreateActivateMode.ValueString() == "" {
		model.OnCreateActivateMode = types.StringValue(asyncActivateMode) // default value
	}
	return nil
}

func ActivateReplicationAfterCreate(ctx context.Context, config *config.Config, resp *datatransfer.Transfer, model *yandexDatatransferTransferModel) diag.Diagnostics {
	var diags diag.Diagnostics
	transferType := model.GetType().String()
	if transferType != datatransfer.TransferType_SNAPSHOT_ONLY.String() {
		activateType := model.GetOnCreateActivateMode().ValueString()
		if activateType != dontActivateMode {
			syncMode := activateType == syncActivateMode
			tflog.Debug(ctx, fmt.Sprintf("activating transfer due to on_create_activate_mode param: %s", activateType))
			if err := activateTransfer(ctx, config, resp.Id, syncMode); err != nil {
				diags.AddError("failed to activate transfer", fmt.Sprintf("cannot activate transfer %q: %s", resp.Id, err.Error()))
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("activating skipped by on_create_activate_mode param: %s", activateType))
		}
	}
	return diags
}

func activateTransfer(ctx context.Context, config *config.Config, transferID string, waitActivating bool) error {
	req := &datatransfer.ActivateTransferRequest{TransferId: transferID}
	activateTransferMetadata := new(metadata.MD)
	activateOp, err := datatransferv1sdk.NewTransferClient(config.SDKv2).Activate(ctx, req, grpc.Header(activateTransferMetadata))
	if traceHeader := activateTransferMetadata.Get(traceIDMetadataKey); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Activate Transfer %s: %s", traceIDMetadataKey, traceHeader[0]))
	}
	if traceHeader := activateTransferMetadata.Get(requestIDMetadataKey); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Activate Transfer %s: %s", requestIDMetadataKey, traceHeader[0]))
	}
	if err != nil {
		return err
	}
	if waitActivating {
		if _, err := activateOp.Wait(ctx); err != nil {
			tflog.Warn(ctx, fmt.Sprintf("[ERROR] Error while waiting for transfer activation to complete: %s", err.Error()))
			return fmt.Errorf("error while waiting for transfer %s activation to complete:  %s", transferID, err.Error())
		}
	}
	return nil
}

func DeactivateReplicationBeforeDelete(ctx context.Context, config *config.Config, req *datatransfer.DeleteTransferRequest, model *yandexDatatransferTransferModel) diag.Diagnostics {
	transferType := model.GetType().String()
	if transferType != datatransfer.TransferType_SNAPSHOT_ONLY.String() {
		if err := deactivateTransfer(ctx, config, req.GetTransferId()); err != nil {
			tflog.Error(ctx, fmt.Sprintf("[WARN] Deactivate Transfer %s error: %s. Trying to delete", req.GetTransferId(), err.Error()))
			// still proceed with delete
		}
	}
	return nil
}

func deactivateTransfer(ctx context.Context, config *config.Config, transferID string) error {
	req := &datatransfer.DeactivateTransferRequest{TransferId: transferID}

	deactivateTransferMetadata := new(metadata.MD)
	deactivateOp, err := datatransferv1sdk.NewTransferClient(config.SDKv2).Deactivate(ctx, req, grpc.Header(deactivateTransferMetadata))
	if traceHeader := deactivateTransferMetadata.Get(traceIDMetadataKey); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Deactivate Transfer %s: %s", traceIDMetadataKey, traceHeader[0]))
	}
	if traceHeader := deactivateTransferMetadata.Get(requestIDMetadataKey); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Deactivate Transfer %s: %s", requestIDMetadataKey, traceHeader[0]))
	}
	if err != nil {
		grpcStatus, ok := status.FromError(err)
		if !ok {
			return err
		}
		for _, detail := range grpcStatus.Details() {
			errorInfo, ok := detail.(*errdetails.ErrorInfo)
			if !ok {
				continue
			}
			if errorInfo.Domain == "datatransfer" && errorInfo.Reason == "INVALID_TRANSFER_STATUS" {
				currentStatus := errorInfo.Metadata["current_status"]
				tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Deactivate operation is not applicable for transfer %q since the status of the transfer is %q", transferID, currentStatus))
				// this is ok if deactivate is not applicable
				return nil
			}
		}
		return err
	}
	if _, err := deactivateOp.Wait(ctx); err != nil {
		return fmt.Errorf("error while waiting operation to complete: %s", err.Error())
	}

	return nil
}
