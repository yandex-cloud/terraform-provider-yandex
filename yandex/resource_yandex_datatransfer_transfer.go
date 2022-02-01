package yandex

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	traceIDMetadataKey   = "x-server-trace-id"
	requestIDMetadataKey = "x-server-request-id"
)

func resourceYandexDatatransferTransfer() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexDatatransferTransferCreateAndActivate,
		Read:   resourceYandexDatatransferTransferRead,
		Update: resourceYandexDatatransferTransferUpdate,
		Delete: resourceYandexDatatransferTransferDeactivateAndDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"source_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"target_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseDatatransferTransferType),
			},

			"warning": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func createTransfer(config *Config, d *schema.ResourceData) (*datatransfer.Transfer, error) {
	ctx := config.Context()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, err
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return nil, err
	}

	transferType, err := parseDatatransferTransferType(d.Get("type").(string))
	if err != nil {
		return nil, err
	}

	req := &datatransfer.CreateTransferRequest{
		SourceId:    d.Get("source_id").(string),
		TargetId:    d.Get("target_id").(string),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		FolderId:    folderID,
		Type:        transferType,
	}

	createTransferMetadata := new(metadata.MD)
	createOp, err := config.sdk.WrapOperation(config.sdk.DataTransfer().Transfer().Create(ctx, req, grpc.Header(createTransferMetadata)))
	if traceHeader := createTransferMetadata.Get(traceIDMetadataKey); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create Transfer %s: %s", traceIDMetadataKey, traceHeader[0])
	}
	if traceHeader := createTransferMetadata.Get(requestIDMetadataKey); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create Transfer %s: %s", requestIDMetadataKey, traceHeader[0])
	}
	if err != nil {
		return nil, err
	}

	protoMetadata, err := createOp.Metadata()
	if err != nil {
		return nil, fmt.Errorf("error while getting TransferService.Create operation metadata: %s", err)
	}
	createOpMetadata, ok := protoMetadata.(*datatransfer.CreateTransferMetadata)
	if !ok {
		return nil, fmt.Errorf("expected TransferService.Create response metadata to have type CreateTransferMetadata but got %T", protoMetadata)
	}
	d.SetId(createOpMetadata.TransferId)

	if err := createOp.Wait(ctx); err != nil {
		return nil, fmt.Errorf("error while waiting operation to complete: %s", err)
	}

	response, err := createOp.Response()
	if err != nil {
		return nil, fmt.Errorf("cannot get result of the operation: %s", err)
	}
	transfer, ok := response.(*datatransfer.Transfer)
	if !ok {
		return nil, fmt.Errorf("expected TransferService.Create operation response to have type Transfer but got %T", response)
	}
	return transfer, nil
}

func activateTransfer(config *Config, transferID string) error {
	ctx := config.Context()

	req := &datatransfer.ActivateTransferRequest{TransferId: transferID}

	activateTransferMetadata := new(metadata.MD)
	activateOp, err := config.sdk.WrapOperation(config.sdk.DataTransfer().Transfer().Activate(ctx, req, grpc.Header(activateTransferMetadata)))
	if traceHeader := activateTransferMetadata.Get(traceIDMetadataKey); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Activate Transfer %s: %s", traceIDMetadataKey, traceHeader[0])
	}
	if traceHeader := activateTransferMetadata.Get(requestIDMetadataKey); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Activate Transfer %s: %s", requestIDMetadataKey, traceHeader[0])
	}
	if err != nil {
		return err
	}

	if err := activateOp.Wait(ctx); err != nil {
		return fmt.Errorf("error while waiting operation to complete: %s", err)
	}
	return nil
}

func deactivateTransfer(config *Config, transferID string) error {
	ctx := config.Context()

	req := &datatransfer.DeactivateTransferRequest{TransferId: transferID}

	deactivateTransferMetadata := new(metadata.MD)
	deactivateOp, err := config.sdk.WrapOperation(config.sdk.DataTransfer().Transfer().Deactivate(ctx, req, grpc.Header(deactivateTransferMetadata)))
	if traceHeader := deactivateTransferMetadata.Get(traceIDMetadataKey); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Deactivate Transfer %s: %s", traceIDMetadataKey, traceHeader[0])
	}
	if traceHeader := deactivateTransferMetadata.Get(requestIDMetadataKey); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Deactivate Transfer %s: %s", requestIDMetadataKey, traceHeader[0])
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
				log.Printf("[DEBUG] Deactivate operation is not applicable for transfer %q since the status of the transfer is %q", transferID, currentStatus)
				return nil
			}
		}
		return err
	}
	if err := deactivateOp.Wait(ctx); err != nil {
		return fmt.Errorf("error while waiting operation to complete: %s", err)
	}

	return nil
}

func deleteTransfer(config *Config, transferID string) error {
	ctx := config.Context()

	req := &datatransfer.DeleteTransferRequest{TransferId: transferID}

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.DataTransfer().Transfer().Delete(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get(traceIDMetadataKey); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete Transfer %s: %s", traceIDMetadataKey, traceHeader[0])
	}
	if traceHeader := md.Get(requestIDMetadataKey); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete Transfer %s: %s", requestIDMetadataKey, traceHeader[0])
	}
	if err != nil {
		return err
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("error while waiting operation to complete: %s", err)
	}

	return nil
}

func resourceYandexDatatransferTransferCreateAndActivate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	transfer, err := createTransfer(config, d)
	if err != nil {
		return fmt.Errorf("cannot create transfer: %w", err)
	}

	if transfer.Type != datatransfer.TransferType_SNAPSHOT_ONLY {
		err := activateTransfer(config, transfer.Id)
		if err != nil {
			return fmt.Errorf("cannot activate transfer %q: %w", transfer.Id, err)
		}
	}

	return resourceYandexDatatransferTransferRead(d, meta)
}

func resourceYandexDatatransferTransferDeactivateAndDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	transferType, err := parseDatatransferTransferType(d.Get("type").(string))
	if err != nil {
		return err
	}

	if transferType != datatransfer.TransferType_SNAPSHOT_ONLY {
		if err := deactivateTransfer(config, d.Id()); err != nil {
			return handleNotFoundError(err, d, fmt.Sprintf("transfer %q", d.Id()))
		}
	}

	if err := deleteTransfer(config, d.Id()); err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("transfer %q", d.Id()))
	}

	return nil
}

func resourceYandexDatatransferTransferRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx := config.Context()

	req := &datatransfer.GetTransferRequest{
		TransferId: d.Id(),
	}

	md := new(metadata.MD)
	resp, err := config.sdk.DataTransfer().Transfer().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read Transfer x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read Transfer x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("transfer %q", d.Id()))
	}

	if err := d.Set("description", resp.GetDescription()); err != nil {
		log.Printf("[ERROR] failed set field description: %s", err)
		return err
	}
	if err := d.Set("folder_id", resp.GetFolderId()); err != nil {
		log.Printf("[ERROR] failed set field folder_id: %s", err)
		return err
	}
	if err := d.Set("labels", resp.GetLabels()); err != nil {
		log.Printf("[ERROR] failed set field labels: %s", err)
		return err
	}
	if err := d.Set("name", resp.GetName()); err != nil {
		log.Printf("[ERROR] failed set field name: %s", err)
		return err
	}
	if err := d.Set("type", resp.GetType().String()); err != nil {
		log.Printf("[ERROR] failed set field type: %s", err)
		return err
	}
	if err := d.Set("warning", resp.GetWarning()); err != nil {
		log.Printf("[ERROR] failed set field warning: %s", err)
		return err
	}

	return nil
}

func resourceYandexDatatransferTransferUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx := config.Context()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	req := &datatransfer.UpdateTransferRequest{
		TransferId:  d.Id(),
		Description: d.Get("description").(string),
		Labels:      labels,
		Name:        d.Get("name").(string),
	}

	updatePath := generateFieldMasks(d, resourceYandexDatatransferTransferUpdateFieldsMap)
	req.UpdateMask = &fieldmaskpb.FieldMask{Paths: updatePath}

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.DataTransfer().Transfer().Update(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update Transfer x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update Transfer x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return err
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("error while waiting operation to complete: %s", err)
	}

	return resourceYandexDatatransferTransferRead(d, meta)
}

var resourceYandexDatatransferTransferUpdateFieldsMap = map[string]string{
	"description": "description",
	"labels":      "labels",
	"name":        "name",
}
