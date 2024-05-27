// Code generated with gentf. DO NOT EDIT.
package yandex

import (
	fmt "fmt"
	log "log"

	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	validation "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	datatransfer "github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1"
	errdetails "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	traceIDMetadataKey   = "x-server-trace-id"
	requestIDMetadataKey = "x-server-request-id"
)

const (
	// а fake state of the field `on_create_activate_mode` that is set automatically
	// when the transfer has already been created or imported.
	internalMessageActivateMode = "[WARN: works only on create resource]"
	// possible scenarios for activating SNAPSHOT_AND_INCREMENT and SNAPSHOT_ONLY
	// transfers when created or re-created through a Terraform provider.
	syncActivateMode  = "sync_activate"
	asyncActivateMode = "async_activate"
	dontActivateMode  = "dont_activate"
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
			"warning": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"on_create_activate_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      asyncActivateMode,
				ValidateFunc: stringInSliceWithHiddenDefault([]string{syncActivateMode, asyncActivateMode, dontActivateMode}, false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == internalMessageActivateMode
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},

				Set:      schema.HashString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"runtime": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"yc_runtime": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"job_count": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"upload_shard_params": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"job_count": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"process_count": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional: true,
							Computed: true,
						},
					},
				},
				Optional: true,
				Computed: true,
			},
			"source_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"target_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"transformation": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transformers": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"convert_to_string": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"columns": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_columns": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_columns": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
												"tables": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},
									"filter_columns": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"columns": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_columns": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_columns": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
												"tables": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},
									"filter_rows": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"filter": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"tables": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},
									"mask_field": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"columns": {
													Type: schema.TypeList,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Optional: true,
												},
												"function": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"mask_function_hash": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"user_defined_salt": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},
																	},
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
												"tables": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},
									"rename_tables": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"rename_tables": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"new_name": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"name": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},
																		"name_space": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},
																	},
																},
																Optional: true,
															},
															"original_name": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"name": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},
																		"name_space": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},
																	},
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},
									"replace_primary_key": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"keys": {
													Type: schema.TypeList,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Optional: true,
												},
												"tables": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},
									"sharder_transformer": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"columns": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_columns": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_columns": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
												"shards_count": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"tables": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},
									"table_splitter_transformer": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"columns": {
													Type: schema.TypeList,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Optional: true,
												},
												"splitter": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"tables": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exclude_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
															"include_tables": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},
								},
							},
							Optional: true,
						},
					},
				},
				Optional: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseDatatransferTransferTransferType),
				Computed:     true,
			},
		},
	}
}

func stringInSliceWithHiddenDefault(valid []string, ignoreCase bool) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		if k == internalMessageActivateMode {
			return nil, nil
		}
		return validation.StringInSlice(valid, ignoreCase)(i, k)
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

	transferType, err := parseDatatransferTransferTransferType(d.Get("type").(string))
	if err != nil {
		return nil, err
	}

	transformation, err := expandDatatransferTransferTransformation(d)
	if err != nil {
		return nil, err
	}

	runtime, err := expandDatatransferTransferRuntime(d)
	if err != nil {
		return nil, err
	}

	req := &datatransfer.CreateTransferRequest{
		SourceId:       d.Get("source_id").(string),
		TargetId:       d.Get("target_id").(string),
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		Labels:         labels,
		FolderId:       folderID,
		Type:           transferType,
		Runtime:        runtime,
		Transformation: transformation,
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

func activateTransfer(config *Config, transferID string, waitActivating bool) error {
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
	if waitActivating {
		if err := activateOp.Wait(ctx); err != nil {
			return fmt.Errorf("error while waiting operation to complete: %s", err)
		}
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
		activateType := d.Get("on_create_activate_mode").(string)
		if activateType == asyncActivateMode || activateType == syncActivateMode || activateType == internalMessageActivateMode {
			syncMode := activateType == syncActivateMode
			if err := activateTransfer(config, transfer.Id, syncMode); err != nil {
				return fmt.Errorf("cannot activate transfer %q: %w", transfer.Id, err)
			}
		} else {
			log.Printf("activating skipped by on_create_activate_mode param: %s", activateType)
		}
	}

	return resourceYandexDatatransferTransferRead(d, meta)
}

func resourceYandexDatatransferTransferDeactivateAndDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	transferType, err := parseDatatransferTransferTransferType(d.Get("type").(string))
	if err != nil {
		return err
	}

	if transferType != datatransfer.TransferType_SNAPSHOT_ONLY {
		if err := deactivateTransfer(config, d.Id()); err != nil {
			if err := handleNotFoundError(err, d, fmt.Sprintf("transfer %q", d.Id())); err != nil {
				log.Printf("[WARN] Deactivate Transfer %s error: %s. Trying to delete", d.Id(), err)
			} else {
				log.Printf("[INFO] Transfer %s not found", d.Id())
				return nil
			}
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
	if err := d.Set("source_id", resp.GetSource().GetId()); err != nil {
		log.Printf("[ERROR] failed set field source_id: %s", err)
		return err
	}
	if err := d.Set("target_id", resp.GetTarget().GetId()); err != nil {
		log.Printf("[ERROR] failed set field target_id: %s", err)
		return err
	}
	if err := d.Set("on_create_activate_mode", internalMessageActivateMode); err != nil {
		log.Printf("[ERROR] failed set field activate_mode: %s", err)
		return err
	}

	transformation, err := flattenDatatransferTransferTransformation(d, resp.GetTransformation())
	if err != nil {
		log.Printf("[ERROR] failed read field transformation: %s", err)
		return err
	}
	if err := d.Set("transformation", transformation); err != nil {
		log.Printf("[ERROR] failed set field transformation: %s", err)
		return err
	}
	runtime, err := flattenDatatransferTransferRuntime(d, resp.GetRuntime())
	if err != nil {
		log.Printf("[ERROR] failed read field runtime: %s", err)
		return err
	}
	if err := d.Set("runtime", runtime); err != nil {
		log.Printf("[ERROR] failed set field runtime: %s", err)
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

	transformation, err := expandDatatransferTransferTransformation(d)
	if err != nil {
		return err
	}

	runtime, err := expandDatatransferTransferRuntime(d)
	if err != nil {
		return err
	}

	req := &datatransfer.UpdateTransferRequest{
		TransferId:     d.Id(),
		Description:    d.Get("description").(string),
		Labels:         labels,
		Name:           d.Get("name").(string),
		Runtime:        runtime,
		Transformation: transformation,
	}

	updatePath := generateDatatransferFieldMasks(d, datatransferUpdateTransferRequestFieldsRoot)
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

var datatransferUpdateTransferRequestFieldsRoot = &fieldTreeNode{
	protobufFieldName:      "",
	terraformAttributeName: "",
	children: []*fieldTreeNode{
		{
			protobufFieldName:      "description",
			terraformAttributeName: "description",
			children:               nil,
		},
		{
			protobufFieldName:      "labels",
			terraformAttributeName: "labels",
			children:               nil,
		},
		{
			protobufFieldName:      "runtime",
			terraformAttributeName: "runtime",
			children: []*fieldTreeNode{
				{
					protobufFieldName:      "yc_runtime",
					terraformAttributeName: "yc_runtime",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "job_count",
							terraformAttributeName: "job_count",
							children:               nil,
						},
						{
							protobufFieldName:      "upload_shard_params",
							terraformAttributeName: "upload_shard_params",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "job_count",
									terraformAttributeName: "job_count",
									children:               nil,
								},
								{
									protobufFieldName:      "process_count",
									terraformAttributeName: "process_count",
									children:               nil,
								},
							},
						},
					},
				},
			},
		},
		{
			protobufFieldName:      "name",
			terraformAttributeName: "name",
			children:               nil,
		},
		{
			protobufFieldName:      "transformation",
			terraformAttributeName: "transformation",
			children: []*fieldTreeNode{
				{
					protobufFieldName:      "transformers",
					terraformAttributeName: "transformers",
					children:               nil,
				},
			},
		},
	},
}
