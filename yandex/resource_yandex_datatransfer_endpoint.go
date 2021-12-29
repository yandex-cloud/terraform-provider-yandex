package yandex

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1"
)

func resourceYandexDatatransferEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexDatatransferEndpointCreate,
		Read:   resourceYandexDatatransferEndpointRead,
		Update: resourceYandexDatatransferEndpointUpdate,
		Delete: resourceYandexDatatransferEndpointDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"author": {
				Type:     schema.TypeString,
				Computed: true,
			},

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

			"settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mysql_source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mdb_cluster_id": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"settings.mysql_source.connection.on_premise"},
												},

												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ca_certificate": {
																Type:     schema.TypeString,
																Optional: true,
															},

															"hosts": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},

															"port": {
																Type:     schema.TypeInt,
																Optional: true,
															},

															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
															},

															"tls_mode": {
																Type:     schema.TypeList,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"disabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.mysql_source.connection.on_premise.tls_mode.enabled"},
																		},

																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.mysql_source.connection.on_premise.tls_mode.disabled"},
																		},
																	},
																},
																Optional: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.mysql_source.connection.mdb_cluster_id"},
												},
											},
										},
										Optional: true,
									},

									"database": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"exclude_tables_regex": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
									},
									"include_tables_regex": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
									},

									"object_transfer_settings": {
										Type:     schema.TypeList,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"routine": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"trigger": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"view": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},
											},
										},
										Optional: true,
									},

									"password": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"raw": {
													Sensitive: true,
													Type:      schema.TypeString,
													Optional:  true,
												},
											},
										},
										Optional: true,
									},

									"timezone": {
										Type:     schema.TypeString,
										Computed: true,
										Optional: true,
									},

									"user": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.postgres_source", "settings.mysql_target", "settings.postgres_target"},
						},
						"mysql_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mdb_cluster_id": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"settings.mysql_target.connection.on_premise"},
												},

												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ca_certificate": {
																Type:     schema.TypeString,
																Optional: true,
															},

															"hosts": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},

															"port": {
																Type:     schema.TypeInt,
																Optional: true,
															},

															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
															},

															"tls_mode": {
																Type:     schema.TypeList,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"disabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.mysql_target.connection.on_premise.tls_mode.enabled"},
																		},

																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.mysql_target.connection.on_premise.tls_mode.disabled"},
																		},
																	},
																},
																Optional: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.mysql_target.connection.mdb_cluster_id"},
												},
											},
										},
										Optional: true,
									},

									"database": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"password": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"raw": {
													Sensitive: true,
													Type:      schema.TypeString,
													Optional:  true,
												},
											},
										},
										Optional: true,
									},

									"service_schema": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"skip_constraint_checks": {
										Type:     schema.TypeBool,
										Optional: true,
									},

									"sql_mode": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"timezone": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"user": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.mysql_source", "settings.postgres_source", "settings.postgres_target"},
						},
						"postgres_source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"collapse_inherit_table": {
										Type:     schema.TypeBool,
										Optional: true,
									},

									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mdb_cluster_id": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"settings.postgres_source.connection.on_premise"},
												},

												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ca_certificate": {
																Type:     schema.TypeString,
																Optional: true,
															},

															"hosts": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},

															"port": {
																Type:     schema.TypeInt,
																Optional: true,
															},

															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
															},

															"tls_mode": {
																Type:     schema.TypeList,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"disabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.postgres_source.connection.on_premise.tls_mode.enabled"},
																		},

																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.postgres_source.connection.on_premise.tls_mode.disabled"},
																		},
																	},
																},
																Optional: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.postgres_source.connection.mdb_cluster_id"},
												},
											},
										},
										Optional: true,
									},

									"database": {
										Type:     schema.TypeString,
										Optional: true,
									},

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

									"object_transfer_settings": {
										Type:     schema.TypeList,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cast": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"collation": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"constraint": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"default_values": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"fk_constraint": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"function": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"index": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"policy": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"primary_key": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"rule": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"sequence": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"sequence_owned_by": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"table": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"trigger": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},

												"view": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseEndpointObjectTransferStage),
												},
											},
										},
										Optional: true,
									},

									"password": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"raw": {
													Sensitive: true,
													Type:      schema.TypeString,
													Optional:  true,
												},
											},
										},
										Optional: true,
									},

									"service_schema": {
										Type:     schema.TypeString,
										Computed: true,
										Optional: true,
									},

									"slot_gigabyte_lag_limit": {
										Type:     schema.TypeInt,
										Computed: true,
										Optional: true,
									},

									"user": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.mysql_source", "settings.mysql_target", "settings.postgres_target"},
						},
						"postgres_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mdb_cluster_id": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"settings.postgres_target.connection.on_premise"},
												},

												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ca_certificate": {
																Type:     schema.TypeString,
																Optional: true,
															},

															"hosts": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
															},

															"port": {
																Type:     schema.TypeInt,
																Optional: true,
															},

															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
															},

															"tls_mode": {
																Type:     schema.TypeList,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"disabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.postgres_target.connection.on_premise.tls_mode.enabled"},
																		},

																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.postgres_target.connection.on_premise.tls_mode.disabled"},
																		},
																	},
																},
																Optional: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.postgres_target.connection.mdb_cluster_id"},
												},
											},
										},
										Optional: true,
									},

									"database": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"password": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"raw": {
													Sensitive: true,
													Type:      schema.TypeString,
													Optional:  true,
												},
											},
										},
										Optional: true,
									},

									"user": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.mysql_source", "settings.postgres_source", "settings.mysql_target"},
						},
					},
				},
				Optional: true,
			},
		},
	}
}

func prepareCreateDatatransferEndpointRequest(d *schema.ResourceData, config *Config) (*datatransfer.CreateEndpointRequest, error) {
	folderId, err := getFolderID(d, config)
	if err != nil {
		return nil, err
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, err
	}

	settings, err := expandEndpointSettings(d)
	if err != nil {
		return nil, err
	}

	return &datatransfer.CreateEndpointRequest{
		FolderId:    folderId,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		Settings:    settings,
	}, nil
}

func resourceYandexDatatransferEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	req, err := prepareCreateDatatransferEndpointRequest(d, config)
	if err != nil {
		return fmt.Errorf("could not prepare request: %s", err)
	}

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.DataTransfer().Endpoint().Create(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create Endpoint x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create Endpoint x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return err
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while getting EndpointService.Create operation metadata: %s", err)
	}
	createEndpointMetadata, ok := protoMetadata.(*datatransfer.CreateEndpointMetadata)
	if !ok {
		return fmt.Errorf("expected EndpointService.Create response metadata to have type CreateEndpointMetadata but got %T", protoMetadata)
	}

	d.SetId(createEndpointMetadata.EndpointId)

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	return resourceYandexDatatransferEndpointRead(d, meta)
}

func resourceYandexDatatransferEndpointRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx := config.Context()

	req := &datatransfer.GetEndpointRequest{
		EndpointId: d.Id(),
	}

	md := new(metadata.MD)
	resp, err := config.sdk.DataTransfer().Endpoint().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read Endpoint x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read Endpoint x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("endpoint %q", d.Id()))
	}

	settings, err := flattenDatatransferSettings(d, resp.GetSettings())
	if err != nil {
		return err
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
	if err := d.Set("settings", settings); err != nil {
		log.Printf("[ERROR] failed set field settings: %s", err)
		return err
	}

	return nil
}

func resourceYandexDatatransferEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx := config.Context()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	settings, err := expandEndpointSettings(d)
	if err != nil {
		return err
	}

	req := &datatransfer.UpdateEndpointRequest{
		EndpointId:  d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		Settings:    settings,
	}

	updatePath := generateFieldMasks(d, resourceYandexDatatransferEndpointUpdateFieldsMap)
	req.UpdateMask = &fieldmaskpb.FieldMask{Paths: updatePath}

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.DataTransfer().Endpoint().Update(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update Endpoint x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update Endpoint x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return err
	}

	if err := op.Wait(ctx); err != nil {
		return err
	}

	return resourceYandexDatatransferEndpointRead(d, meta)
}

func resourceYandexDatatransferEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx := config.Context()

	req := &datatransfer.DeleteEndpointRequest{
		EndpointId: d.Id(),
	}

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.DataTransfer().Endpoint().Delete(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete Endpoint x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete Endpoint x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("endpoint %q", d.Id()))
	}

	if err := op.Wait(ctx); err != nil {
		return err
	}

	return nil
}

var resourceYandexDatatransferEndpointUpdateFieldsMap = map[string]string{
	"name":        "name",
	"description": "description",
	"labels":      "labels",
	"settings.0.mysql_source.0.connection.0.mdb_cluster_id":                                      "settings.mysql_source.connection.mdb_cluster_id",
	"settings.0.mysql_source.0.connection.0.on_premise.0.hosts":                                  "settings.mysql_source.connection.on_premise.hosts",
	"settings.0.mysql_source.0.connection.0.on_premise.0.port":                                   "settings.mysql_source.connection.on_premise.port",
	"settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate":    "settings.mysql_source.connection.on_premise.tls_mode.enabled.ca_certificate",
	"settings.0.mysql_source.0.connection.0.on_premise.0.subnet_id":                              "settings.mysql_source.connection.on_premise.subnet_id",
	"settings.0.mysql_source.0.database":                                                         "settings.mysql_source.database",
	"settings.0.mysql_source.0.user":                                                             "settings.mysql_source.user",
	"settings.0.mysql_source.0.password.0.raw":                                                   "settings.mysql_source.password.raw",
	"settings.0.mysql_source.0.include_tables_regex":                                             "settings.mysql_source.include_tables_regex",
	"settings.0.mysql_source.0.exclude_tables_regex":                                             "settings.mysql_source.exclude_tables_regex",
	"settings.0.mysql_source.0.timezone":                                                         "settings.mysql_source.timezone",
	"settings.0.mysql_source.0.object_transfer_settings.0.view":                                  "settings.mysql_source.object_transfer_settings.view",
	"settings.0.mysql_source.0.object_transfer_settings.0.routine":                               "settings.mysql_source.object_transfer_settings.routine",
	"settings.0.mysql_source.0.object_transfer_settings.0.trigger":                               "settings.mysql_source.object_transfer_settings.trigger",
	"settings.0.mysql_target.0.connection.0.mdb_cluster_id":                                      "settings.mysql_target.connection.mdb_cluster_id",
	"settings.0.mysql_target.0.connection.0.on_premise.0.hosts":                                  "settings.mysql_target.connection.on_premise.hosts",
	"settings.0.mysql_target.0.connection.0.on_premise.0.port":                                   "settings.mysql_target.connection.on_premise.port",
	"settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate":    "settings.mysql_target.connection.on_premise.tls_mode.enabled.ca_certificate",
	"settings.0.mysql_target.0.connection.0.on_premise.0.subnet_id":                              "settings.mysql_target.connection.on_premise.subnet_id",
	"settings.0.mysql_target.0.database":                                                         "settings.mysql_target.database",
	"settings.0.mysql_target.0.user":                                                             "settings.mysql_target.user",
	"settings.0.mysql_target.0.password.0.raw":                                                   "settings.mysql_target.password.raw",
	"settings.0.mysql_target.0.sql_mode":                                                         "settings.mysql_target.sql_mode",
	"settings.0.mysql_target.0.skip_constraint_checks":                                           "settings.mysql_target.skip_constraint_checks",
	"settings.0.mysql_target.0.timezone":                                                         "settings.mysql_target.timezone",
	"settings.0.postgres_source.0.connection.0.mdb_cluster_id":                                   "settings.postgres_source.connection.mdb_cluster_id",
	"settings.0.postgres_source.0.connection.0.on_premise.0.hosts":                               "settings.postgres_source.connection.on_premise.hosts",
	"settings.0.postgres_source.0.connection.0.on_premise.0.port":                                "settings.postgres_source.connection.on_premise.port",
	"settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate": "settings.postgres_source.connection.on_premise.tls_mode.enabled.ca_certificate",
	"settings.0.postgres_source.0.connection.0.on_premise.0.subnet_id":                           "settings.postgres_source.connection.on_premise.subnet_id",
	"settings.0.postgres_source.0.database":                                                      "settings.postgres_source.database",
	"settings.0.postgres_source.0.user":                                                          "settings.postgres_source.user",
	"settings.0.postgres_source.0.password.0.raw":                                                "settings.postgres_source.password.raw",
	"settings.0.postgres_source.0.include_tables":                                                "settings.postgres_source.include_tables",
	"settings.0.postgres_source.0.exclude_tables":                                                "settings.postgres_source.exclude_tables",
	"settings.0.postgres_source.0.slot_gigabyte_lag_limit":                                       "settings.postgres_source.slot_byte_lag_limit",
	"settings.0.postgres_source.0.service_schema":                                                "settings.postgres_source.service_schema",
	"settings.0.postgres_source.0.object_transfer_settings.0.sequence":                           "settings.postgres_source.object_transfer_settings.sequence",
	"settings.0.postgres_source.0.object_transfer_settings.0.sequence_owned_by":                  "settings.postgres_source.object_transfer_settings.sequence_owned_by",
	"settings.0.postgres_source.0.object_transfer_settings.0.table":                              "settings.postgres_source.object_transfer_settings.table",
	"settings.0.postgres_source.0.object_transfer_settings.0.primary_key":                        "settings.postgres_source.object_transfer_settings.primary_key",
	"settings.0.postgres_source.0.object_transfer_settings.0.fk_constraint":                      "settings.postgres_source.object_transfer_settings.fk_constraint",
	"settings.0.postgres_source.0.object_transfer_settings.0.default_values":                     "settings.postgres_source.object_transfer_settings.default_values",
	"settings.0.postgres_source.0.object_transfer_settings.0.constraint":                         "settings.postgres_source.object_transfer_settings.constraint",
	"settings.0.postgres_source.0.object_transfer_settings.0.index":                              "settings.postgres_source.object_transfer_settings.index",
	"settings.0.postgres_source.0.object_transfer_settings.0.view":                               "settings.postgres_source.object_transfer_settings.view",
	"settings.0.postgres_source.0.object_transfer_settings.0.function":                           "settings.postgres_source.object_transfer_settings.function",
	"settings.0.postgres_source.0.object_transfer_settings.0.trigger":                            "settings.postgres_source.object_transfer_settings.trigger",
	"settings.0.postgres_source.0.object_transfer_settings.0.type":                               "settings.postgres_source.object_transfer_settings.type",
	"settings.0.postgres_source.0.object_transfer_settings.0.rule":                               "settings.postgres_source.object_transfer_settings.rule",
	"settings.0.postgres_source.0.object_transfer_settings.0.collation":                          "settings.postgres_source.object_transfer_settings.collation",
	"settings.0.postgres_source.0.object_transfer_settings.0.policy":                             "settings.postgres_source.object_transfer_settings.policy",
	"settings.0.postgres_source.0.object_transfer_settings.0.cast":                               "settings.postgres_source.object_transfer_settings.cast",
	"settings.0.postgres_target.0.connection.0.mdb_cluster_id":                                   "settings.postgres_target.connection.mdb_cluster_id",
	"settings.0.postgres_target.0.connection.0.on_premise.0.hosts":                               "settings.postgres_target.connection.on_premise.hosts",
	"settings.0.postgres_target.0.connection.0.on_premise.0.port":                                "settings.postgres_target.connection.on_premise.port",
	"settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate": "settings.postgres_target.connection.on_premise.tls_mode.enabled.ca_certificate",
	"settings.0.postgres_target.0.connection.0.on_premise.0.subnet_id":                           "settings.postgres_target.connection.on_premise.subnet_id",
	"settings.0.postgres_target.0.database":                                                      "settings.postgres_target.database",
	"settings.0.postgres_target.0.user":                                                          "settings.postgres_target.user",
	"settings.0.postgres_target.0.password.0.raw":                                                "settings.postgres_target.password.raw",
}
