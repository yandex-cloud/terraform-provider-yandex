// Code generated with gentf. DO NOT EDIT.
package yandex

import (
	fmt "fmt"
	log "log"

	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	datatransfer "github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1"
	grpc "google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
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
			"settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"clickhouse_source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"clickhouse_cluster_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"connection_options": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"database": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"mdb_cluster_id": {
																Type:          schema.TypeString,
																Optional:      true,
																ConflictsWith: []string{"settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise"},
															},
															"on_premise": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"http_port": {
																			Type:     schema.TypeInt,
																			Optional: true,
																			Computed: true,
																		},
																		"native_port": {
																			Type:     schema.TypeInt,
																			Optional: true,
																			Computed: true,
																		},
																		"shards": {
																			Type: schema.TypeList,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"hosts": {
																						Type: schema.TypeList,
																						Elem: &schema.Schema{
																							Type: schema.TypeString,
																						},
																						Optional: true,
																						Computed: true,
																					},
																					"name": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional: true,
																			Computed: true,
																		},
																		"tls_mode": {
																			Type:     schema.TypeList,
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
																						ConflictsWith: []string{"settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled"},
																					},
																					"enabled": {
																						Type:     schema.TypeList,
																						MaxItems: 1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"ca_certificate": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																							},
																						},
																						Optional:      true,
																						ConflictsWith: []string{"settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.disabled"},
																					},
																				},
																			},
																			Optional: true,
																			Computed: true,
																		},
																	},
																},
																Optional:      true,
																ConflictsWith: []string{"settings.0.clickhouse_source.0.connection.0.connection_options.0.mdb_cluster_id"},
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
																			Computed:  true,
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"user": {
																Type:     schema.TypeString,
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
									"exclude_tables": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"include_tables": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"clickhouse_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"alt_names": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"from_name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"to_name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"cleanup_policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateParsableValue(parseDatatransferEndpointClickhouseCleanupPolicy),
										Computed:     true,
									},
									"clickhouse_cluster_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"connection_options": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"database": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"mdb_cluster_id": {
																Type:          schema.TypeString,
																Optional:      true,
																ConflictsWith: []string{"settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise"},
															},
															"on_premise": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"http_port": {
																			Type:     schema.TypeInt,
																			Optional: true,
																			Computed: true,
																		},
																		"native_port": {
																			Type:     schema.TypeInt,
																			Optional: true,
																			Computed: true,
																		},
																		"shards": {
																			Type: schema.TypeList,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"hosts": {
																						Type: schema.TypeList,
																						Elem: &schema.Schema{
																							Type: schema.TypeString,
																						},
																						Optional: true,
																						Computed: true,
																					},
																					"name": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional: true,
																			Computed: true,
																		},
																		"tls_mode": {
																			Type:     schema.TypeList,
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
																						ConflictsWith: []string{"settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled"},
																					},
																					"enabled": {
																						Type:     schema.TypeList,
																						MaxItems: 1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"ca_certificate": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																							},
																						},
																						Optional:      true,
																						ConflictsWith: []string{"settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.disabled"},
																					},
																				},
																			},
																			Optional: true,
																			Computed: true,
																		},
																	},
																},
																Optional:      true,
																ConflictsWith: []string{"settings.0.clickhouse_target.0.connection.0.connection_options.0.mdb_cluster_id"},
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
																			Computed:  true,
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"user": {
																Type:     schema.TypeString,
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
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"sharding": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"column_value_hash": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"column_name": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.clickhouse_target.0.sharding.0.custom_mapping", "settings.0.clickhouse_target.0.sharding.0.round_robin", "settings.0.clickhouse_target.0.sharding.0.transfer_id"},
												},
												"custom_mapping": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"column_name": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"mapping": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"column_value": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"string_value": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional: true,
																			Computed: true,
																		},
																		"shard_name": {
																			Type:     schema.TypeString,
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
													Optional:      true,
													ConflictsWith: []string{"settings.0.clickhouse_target.0.sharding.0.column_value_hash", "settings.0.clickhouse_target.0.sharding.0.round_robin", "settings.0.clickhouse_target.0.sharding.0.transfer_id"},
												},
												"round_robin": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.clickhouse_target.0.sharding.0.column_value_hash", "settings.0.clickhouse_target.0.sharding.0.custom_mapping", "settings.0.clickhouse_target.0.sharding.0.transfer_id"},
												},
												"transfer_id": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.clickhouse_target.0.sharding.0.column_value_hash", "settings.0.clickhouse_target.0.sharding.0.custom_mapping", "settings.0.clickhouse_target.0.sharding.0.round_robin"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"kafka_source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auth": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"no_auth": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_source.0.auth.0.sasl"},
												},
												"sasl": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"mechanism": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validateParsableValue(parseDatatransferEndpointKafkaMechanism),
																Computed:     true,
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
																			Computed:  true,
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"user": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_source.0.auth.0.no_auth"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cluster_id": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_source.0.connection.0.on_premise"},
												},
												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"broker_urls": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
																Computed: true,
															},
															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"tls_mode": {
																Type:     schema.TypeList,
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
																			ConflictsWith: []string{"settings.0.kafka_source.0.connection.0.on_premise.0.tls_mode.0.enabled"},
																		},
																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.kafka_source.0.connection.0.on_premise.0.tls_mode.0.disabled"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_source.0.connection.0.cluster_id"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"parser": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"audit_trails_v1_parser": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_source.0.parser.0.cloud_logging_parser", "settings.0.kafka_source.0.parser.0.json_parser", "settings.0.kafka_source.0.parser.0.tskv_parser"},
												},
												"cloud_logging_parser": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_source.0.parser.0.audit_trails_v1_parser", "settings.0.kafka_source.0.parser.0.json_parser", "settings.0.kafka_source.0.parser.0.tskv_parser"},
												},
												"json_parser": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"add_rest_column": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
															"data_schema": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"fields": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"fields": {
																						Type: schema.TypeList,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"key": {
																									Type:     schema.TypeBool,
																									Optional: true,
																									Computed: true,
																								},
																								"name": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																								"path": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																								"required": {
																									Type:     schema.TypeBool,
																									Optional: true,
																									Computed: true,
																								},
																								"type": {
																									Type:         schema.TypeString,
																									Optional:     true,
																									ValidateFunc: validateParsableValue(parseDatatransferEndpointColumnType),
																									Computed:     true,
																								},
																							},
																						},
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.json_fields"},
																		},
																		"json_fields": {
																			Type:          schema.TypeString,
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.fields"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"null_keys_allowed": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
															"unescape_string_values": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_source.0.parser.0.audit_trails_v1_parser", "settings.0.kafka_source.0.parser.0.cloud_logging_parser", "settings.0.kafka_source.0.parser.0.tskv_parser"},
												},
												"tskv_parser": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"add_rest_column": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
															"data_schema": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"fields": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"fields": {
																						Type: schema.TypeList,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"key": {
																									Type:     schema.TypeBool,
																									Optional: true,
																									Computed: true,
																								},
																								"name": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																								"path": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																								"required": {
																									Type:     schema.TypeBool,
																									Optional: true,
																									Computed: true,
																								},
																								"type": {
																									Type:         schema.TypeString,
																									Optional:     true,
																									ValidateFunc: validateParsableValue(parseDatatransferEndpointColumnType),
																									Computed:     true,
																								},
																							},
																						},
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.json_fields"},
																		},
																		"json_fields": {
																			Type:          schema.TypeString,
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.fields"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"null_keys_allowed": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
															"unescape_string_values": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_source.0.parser.0.audit_trails_v1_parser", "settings.0.kafka_source.0.parser.0.cloud_logging_parser", "settings.0.kafka_source.0.parser.0.json_parser"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"topic_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"topic_names": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"transformer": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"buffer_flush_interval": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"buffer_size": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"cloud_function": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"invocation_timeout": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"number_of_retries": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"service_account_id": {
													Type:     schema.TypeString,
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
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"kafka_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auth": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"no_auth": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_target.0.auth.0.sasl"},
												},
												"sasl": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"mechanism": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validateParsableValue(parseDatatransferEndpointKafkaMechanism),
																Computed:     true,
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
																			Computed:  true,
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"user": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_target.0.auth.0.no_auth"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cluster_id": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_target.0.connection.0.on_premise"},
												},
												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"broker_urls": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
																Computed: true,
															},
															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"tls_mode": {
																Type:     schema.TypeList,
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
																			ConflictsWith: []string{"settings.0.kafka_target.0.connection.0.on_premise.0.tls_mode.0.enabled"},
																		},
																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.kafka_target.0.connection.0.on_premise.0.tls_mode.0.disabled"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_target.0.connection.0.cluster_id"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"serializer": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"serializer_auto": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_target.0.serializer.0.serializer_debezium", "settings.0.kafka_target.0.serializer.0.serializer_json"},
												},
												"serializer_debezium": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"serializer_parameters": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"key": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																		"value": {
																			Type:     schema.TypeString,
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
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_target.0.serializer.0.serializer_auto", "settings.0.kafka_target.0.serializer.0.serializer_json"},
												},
												"serializer_json": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_target.0.serializer.0.serializer_auto", "settings.0.kafka_target.0.serializer.0.serializer_debezium"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"topic_settings": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"topic": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"save_tx_order": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
															"topic_name": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_target.0.topic_settings.0.topic_prefix"},
												},
												"topic_prefix": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"settings.0.kafka_target.0.topic_settings.0.topic"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"metrika_source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"counter_ids": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
										Optional: true,
										Computed: true,
									},
									"streams": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"columns": {
													Type: schema.TypeList,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Optional: true,
													Computed: true,
												},
												"type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointMetrikaStreamType),
													Computed:     true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"token": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"raw": {
													Sensitive: true,
													Type:      schema.TypeString,
													Optional:  true,
													Computed:  true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"mongo_source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"collections": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"collection_name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"database_name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"connection_options": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_source": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"mdb_cluster_id": {
																Type:          schema.TypeString,
																Optional:      true,
																ConflictsWith: []string{"settings.0.mongo_source.0.connection.0.connection_options.0.on_premise"},
															},
															"on_premise": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"hosts": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Optional: true,
																			Computed: true,
																		},
																		"port": {
																			Type:     schema.TypeInt,
																			Optional: true,
																			Computed: true,
																		},
																		"replica_set": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																		"tls_mode": {
																			Type:     schema.TypeList,
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
																						ConflictsWith: []string{"settings.0.mongo_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled"},
																					},
																					"enabled": {
																						Type:     schema.TypeList,
																						MaxItems: 1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"ca_certificate": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																							},
																						},
																						Optional:      true,
																						ConflictsWith: []string{"settings.0.mongo_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.disabled"},
																					},
																				},
																			},
																			Optional: true,
																			Computed: true,
																		},
																	},
																},
																Optional:      true,
																ConflictsWith: []string{"settings.0.mongo_source.0.connection.0.connection_options.0.mdb_cluster_id"},
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
																			Computed:  true,
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"user": {
																Type:     schema.TypeString,
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
									"excluded_collections": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"collection_name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"database_name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"secondary_preferred_mode": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"mongo_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cleanup_policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateParsableValue(parseDatatransferEndpointCleanupPolicy),
										Computed:     true,
									},
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"connection_options": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_source": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"mdb_cluster_id": {
																Type:          schema.TypeString,
																Optional:      true,
																ConflictsWith: []string{"settings.0.mongo_target.0.connection.0.connection_options.0.on_premise"},
															},
															"on_premise": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"hosts": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Optional: true,
																			Computed: true,
																		},
																		"port": {
																			Type:     schema.TypeInt,
																			Optional: true,
																			Computed: true,
																		},
																		"replica_set": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																		"tls_mode": {
																			Type:     schema.TypeList,
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
																						ConflictsWith: []string{"settings.0.mongo_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled"},
																					},
																					"enabled": {
																						Type:     schema.TypeList,
																						MaxItems: 1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"ca_certificate": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																							},
																						},
																						Optional:      true,
																						ConflictsWith: []string{"settings.0.mongo_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.disabled"},
																					},
																				},
																			},
																			Optional: true,
																			Computed: true,
																		},
																	},
																},
																Optional:      true,
																ConflictsWith: []string{"settings.0.mongo_target.0.connection.0.connection_options.0.mdb_cluster_id"},
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
																			Computed:  true,
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"user": {
																Type:     schema.TypeString,
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
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
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
													ConflictsWith: []string{"settings.0.mysql_source.0.connection.0.on_premise"},
												},
												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"hosts": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
																Computed: true,
															},
															"port": {
																Type:     schema.TypeInt,
																Optional: true,
																Computed: true,
															},
															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"tls_mode": {
																Type:     schema.TypeList,
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
																			ConflictsWith: []string{"settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode.0.enabled"},
																		},
																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode.0.disabled"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.mysql_source.0.connection.0.mdb_cluster_id"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"exclude_tables_regex": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"include_tables_regex": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"object_transfer_settings": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"routine": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"tables": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"trigger": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"view": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
											},
										},
										Optional: true,
										Computed: true,
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
													Computed:  true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"service_database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"timezone": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"user": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"mysql_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cleanup_policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateParsableValue(parseDatatransferEndpointCleanupPolicy),
										Computed:     true,
									},
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mdb_cluster_id": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"settings.0.mysql_target.0.connection.0.on_premise"},
												},
												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"hosts": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
																Computed: true,
															},
															"port": {
																Type:     schema.TypeInt,
																Optional: true,
																Computed: true,
															},
															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"tls_mode": {
																Type:     schema.TypeList,
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
																			ConflictsWith: []string{"settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode.0.enabled"},
																		},
																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode.0.disabled"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.mysql_target.0.connection.0.mdb_cluster_id"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
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
													Computed:  true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"service_database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"skip_constraint_checks": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"sql_mode": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"timezone": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"user": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"postgres_source": {
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
													ConflictsWith: []string{"settings.0.postgres_source.0.connection.0.on_premise"},
												},
												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"hosts": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
																Computed: true,
															},
															"port": {
																Type:     schema.TypeInt,
																Optional: true,
																Computed: true,
															},
															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"tls_mode": {
																Type:     schema.TypeList,
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
																			ConflictsWith: []string{"settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode.0.enabled"},
																		},
																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode.0.disabled"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.postgres_source.0.connection.0.mdb_cluster_id"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"exclude_tables": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"include_tables": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"object_transfer_settings": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cast": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"collation": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"constraint": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"default_values": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"fk_constraint": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"function": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"index": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"materialized_view": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"policy": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"primary_key": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"rule": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"sequence": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"sequence_owned_by": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"sequence_set": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"table": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"trigger": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
												"view": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateParsableValue(parseDatatransferEndpointObjectTransferStage),
													Computed:     true,
												},
											},
										},
										Optional: true,
										Computed: true,
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
													Computed:  true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"service_schema": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"slot_gigabyte_lag_limit": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"user": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"postgres_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cleanup_policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateParsableValue(parseDatatransferEndpointCleanupPolicy),
										Computed:     true,
									},
									"connection": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mdb_cluster_id": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"settings.0.postgres_target.0.connection.0.on_premise"},
												},
												"on_premise": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"hosts": {
																Type: schema.TypeList,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
																Optional: true,
																Computed: true,
															},
															"port": {
																Type:     schema.TypeInt,
																Optional: true,
																Computed: true,
															},
															"subnet_id": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"tls_mode": {
																Type:     schema.TypeList,
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
																			ConflictsWith: []string{"settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode.0.enabled"},
																		},
																		"enabled": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"ca_certificate": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode.0.disabled"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.postgres_target.0.connection.0.mdb_cluster_id"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
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
													Computed:  true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"user": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"ydb_source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"changefeed_custom_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"instance": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"paths": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"sa_key_content": {
										Sensitive: true,
										Type:      schema.TypeString,
										Optional:  true,
										Computed:  true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"service_account_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_target", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"ydb_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cleanup_policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateParsableValue(parseDatatransferEndpointYdbCleanupPolicy),
										Computed:     true,
									},
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"default_compression": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateParsableValue(parseDatatransferEndpointYdbDefaultCompression),
										Computed:     true,
									},
									"instance": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"is_table_column_oriented": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"sa_key_content": {
										Sensitive: true,
										Type:      schema.TypeString,
										Optional:  true,
										Computed:  true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"service_account_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.yds_source", "settings.0.yds_target"},
						},
						"yds_source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allow_ttl_rewind": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"consumer": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"endpoint": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"parser": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"audit_trails_v1_parser": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.yds_source.0.parser.0.cloud_logging_parser", "settings.0.yds_source.0.parser.0.json_parser", "settings.0.yds_source.0.parser.0.tskv_parser"},
												},
												"cloud_logging_parser": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.yds_source.0.parser.0.audit_trails_v1_parser", "settings.0.yds_source.0.parser.0.json_parser", "settings.0.yds_source.0.parser.0.tskv_parser"},
												},
												"json_parser": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"add_rest_column": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
															"data_schema": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"fields": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"fields": {
																						Type: schema.TypeList,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"key": {
																									Type:     schema.TypeBool,
																									Optional: true,
																									Computed: true,
																								},
																								"name": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																								"path": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																								"required": {
																									Type:     schema.TypeBool,
																									Optional: true,
																									Computed: true,
																								},
																								"type": {
																									Type:         schema.TypeString,
																									Optional:     true,
																									ValidateFunc: validateParsableValue(parseDatatransferEndpointColumnType),
																									Computed:     true,
																								},
																							},
																						},
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.json_fields"},
																		},
																		"json_fields": {
																			Type:          schema.TypeString,
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.fields"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"null_keys_allowed": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
															"unescape_string_values": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.yds_source.0.parser.0.audit_trails_v1_parser", "settings.0.yds_source.0.parser.0.cloud_logging_parser", "settings.0.yds_source.0.parser.0.tskv_parser"},
												},
												"tskv_parser": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"add_rest_column": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
															"data_schema": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"fields": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"fields": {
																						Type: schema.TypeList,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"key": {
																									Type:     schema.TypeBool,
																									Optional: true,
																									Computed: true,
																								},
																								"name": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																								"path": {
																									Type:     schema.TypeString,
																									Optional: true,
																									Computed: true,
																								},
																								"required": {
																									Type:     schema.TypeBool,
																									Optional: true,
																									Computed: true,
																								},
																								"type": {
																									Type:         schema.TypeString,
																									Optional:     true,
																									ValidateFunc: validateParsableValue(parseDatatransferEndpointColumnType),
																									Computed:     true,
																								},
																							},
																						},
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.json_fields"},
																		},
																		"json_fields": {
																			Type:          schema.TypeString,
																			Optional:      true,
																			ConflictsWith: []string{"settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.fields"},
																		},
																	},
																},
																Optional: true,
																Computed: true,
															},
															"null_keys_allowed": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
															"unescape_string_values": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
														},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.yds_source.0.parser.0.audit_trails_v1_parser", "settings.0.yds_source.0.parser.0.cloud_logging_parser", "settings.0.yds_source.0.parser.0.json_parser"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"service_account_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"stream": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"supported_codecs": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_target"},
						},
						"yds_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"endpoint": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"save_tx_order": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"security_groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										Computed: true,
									},
									"serializer": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"serializer_auto": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.yds_target.0.serializer.0.serializer_debezium", "settings.0.yds_target.0.serializer.0.serializer_json"},
												},
												"serializer_debezium": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"serializer_parameters": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"key": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																		"value": {
																			Type:     schema.TypeString,
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
													Optional:      true,
													ConflictsWith: []string{"settings.0.yds_target.0.serializer.0.serializer_auto", "settings.0.yds_target.0.serializer.0.serializer_json"},
												},
												"serializer_json": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.yds_target.0.serializer.0.serializer_auto", "settings.0.yds_target.0.serializer.0.serializer_debezium"},
												},
											},
										},
										Optional: true,
										Computed: true,
									},
									"service_account_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"stream": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional:      true,
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.kafka_source", "settings.0.kafka_target", "settings.0.metrika_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target", "settings.0.ydb_source", "settings.0.ydb_target", "settings.0.yds_source"},
						},
					},
				},
				Optional: true,
				Computed: true,
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

	settings, err := expandDatatransferEndpointSettings(d)
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

	settings, err := flattenDatatransferEndpointSettings(d, resp.GetSettings())
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

	settings, err := expandDatatransferEndpointSettings(d)
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

	updatePath := generateDatatransferFieldMasks(d, datatransferUpdateEndpointRequestFieldsRoot)
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

var datatransferUpdateEndpointRequestFieldsRoot = &fieldTreeNode{
	protobufFieldName:      "",
	terraformAttributeName: "",
	children: []*fieldTreeNode{
		{
			protobufFieldName:      "name",
			terraformAttributeName: "name",
			children:               nil,
		},
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
			protobufFieldName:      "settings",
			terraformAttributeName: "settings",
			children: []*fieldTreeNode{
				{
					protobufFieldName:      "mysql_source",
					terraformAttributeName: "mysql_source",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "mdb_cluster_id",
									terraformAttributeName: "mdb_cluster_id",
									children:               nil,
								},
								{
									protobufFieldName:      "on_premise",
									terraformAttributeName: "on_premise",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "hosts",
											terraformAttributeName: "hosts",
											children:               nil,
										},
										{
											protobufFieldName:      "port",
											terraformAttributeName: "port",
											children:               nil,
										},
										{
											protobufFieldName:      "tls_mode",
											terraformAttributeName: "tls_mode",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "disabled",
													terraformAttributeName: "disabled",
													children:               nil,
												},
												{
													protobufFieldName:      "enabled",
													terraformAttributeName: "enabled",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "ca_certificate",
															terraformAttributeName: "ca_certificate",
															children:               nil,
														},
													},
												},
											},
										},
										{
											protobufFieldName:      "subnet_id",
											terraformAttributeName: "subnet_id",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "database",
							terraformAttributeName: "database",
							children:               nil,
						},
						{
							protobufFieldName:      "service_database",
							terraformAttributeName: "service_database",
							children:               nil,
						},
						{
							protobufFieldName:      "user",
							terraformAttributeName: "user",
							children:               nil,
						},
						{
							protobufFieldName:      "password",
							terraformAttributeName: "password",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "raw",
									terraformAttributeName: "raw",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "include_tables_regex",
							terraformAttributeName: "include_tables_regex",
							children:               nil,
						},
						{
							protobufFieldName:      "exclude_tables_regex",
							terraformAttributeName: "exclude_tables_regex",
							children:               nil,
						},
						{
							protobufFieldName:      "timezone",
							terraformAttributeName: "timezone",
							children:               nil,
						},
						{
							protobufFieldName:      "object_transfer_settings",
							terraformAttributeName: "object_transfer_settings",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "view",
									terraformAttributeName: "view",
									children:               nil,
								},
								{
									protobufFieldName:      "routine",
									terraformAttributeName: "routine",
									children:               nil,
								},
								{
									protobufFieldName:      "trigger",
									terraformAttributeName: "trigger",
									children:               nil,
								},
								{
									protobufFieldName:      "tables",
									terraformAttributeName: "tables",
									children:               nil,
								},
							},
						},
					},
				},
				{
					protobufFieldName:      "postgres_source",
					terraformAttributeName: "postgres_source",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "mdb_cluster_id",
									terraformAttributeName: "mdb_cluster_id",
									children:               nil,
								},
								{
									protobufFieldName:      "on_premise",
									terraformAttributeName: "on_premise",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "hosts",
											terraformAttributeName: "hosts",
											children:               nil,
										},
										{
											protobufFieldName:      "port",
											terraformAttributeName: "port",
											children:               nil,
										},
										{
											protobufFieldName:      "tls_mode",
											terraformAttributeName: "tls_mode",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "disabled",
													terraformAttributeName: "disabled",
													children:               nil,
												},
												{
													protobufFieldName:      "enabled",
													terraformAttributeName: "enabled",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "ca_certificate",
															terraformAttributeName: "ca_certificate",
															children:               nil,
														},
													},
												},
											},
										},
										{
											protobufFieldName:      "subnet_id",
											terraformAttributeName: "subnet_id",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "database",
							terraformAttributeName: "database",
							children:               nil,
						},
						{
							protobufFieldName:      "user",
							terraformAttributeName: "user",
							children:               nil,
						},
						{
							protobufFieldName:      "password",
							terraformAttributeName: "password",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "raw",
									terraformAttributeName: "raw",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "include_tables",
							terraformAttributeName: "include_tables",
							children:               nil,
						},
						{
							protobufFieldName:      "exclude_tables",
							terraformAttributeName: "exclude_tables",
							children:               nil,
						},
						{
							protobufFieldName:      "slot_byte_lag_limit",
							terraformAttributeName: "slot_gigabyte_lag_limit",
							children:               nil,
						},
						{
							protobufFieldName:      "service_schema",
							terraformAttributeName: "service_schema",
							children:               nil,
						},
						{
							protobufFieldName:      "object_transfer_settings",
							terraformAttributeName: "object_transfer_settings",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "sequence",
									terraformAttributeName: "sequence",
									children:               nil,
								},
								{
									protobufFieldName:      "sequence_owned_by",
									terraformAttributeName: "sequence_owned_by",
									children:               nil,
								},
								{
									protobufFieldName:      "sequence_set",
									terraformAttributeName: "sequence_set",
									children:               nil,
								},
								{
									protobufFieldName:      "table",
									terraformAttributeName: "table",
									children:               nil,
								},
								{
									protobufFieldName:      "primary_key",
									terraformAttributeName: "primary_key",
									children:               nil,
								},
								{
									protobufFieldName:      "fk_constraint",
									terraformAttributeName: "fk_constraint",
									children:               nil,
								},
								{
									protobufFieldName:      "default_values",
									terraformAttributeName: "default_values",
									children:               nil,
								},
								{
									protobufFieldName:      "constraint",
									terraformAttributeName: "constraint",
									children:               nil,
								},
								{
									protobufFieldName:      "index",
									terraformAttributeName: "index",
									children:               nil,
								},
								{
									protobufFieldName:      "view",
									terraformAttributeName: "view",
									children:               nil,
								},
								{
									protobufFieldName:      "materialized_view",
									terraformAttributeName: "materialized_view",
									children:               nil,
								},
								{
									protobufFieldName:      "function",
									terraformAttributeName: "function",
									children:               nil,
								},
								{
									protobufFieldName:      "trigger",
									terraformAttributeName: "trigger",
									children:               nil,
								},
								{
									protobufFieldName:      "type",
									terraformAttributeName: "type",
									children:               nil,
								},
								{
									protobufFieldName:      "rule",
									terraformAttributeName: "rule",
									children:               nil,
								},
								{
									protobufFieldName:      "collation",
									terraformAttributeName: "collation",
									children:               nil,
								},
								{
									protobufFieldName:      "policy",
									terraformAttributeName: "policy",
									children:               nil,
								},
								{
									protobufFieldName:      "cast",
									terraformAttributeName: "cast",
									children:               nil,
								},
							},
						},
					},
				},
				{
					protobufFieldName:      "ydb_source",
					terraformAttributeName: "ydb_source",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "database",
							terraformAttributeName: "database",
							children:               nil,
						},
						{
							protobufFieldName:      "instance",
							terraformAttributeName: "instance",
							children:               nil,
						},
						{
							protobufFieldName:      "service_account_id",
							terraformAttributeName: "service_account_id",
							children:               nil,
						},
						{
							protobufFieldName:      "paths",
							terraformAttributeName: "paths",
							children:               nil,
						},
						{
							protobufFieldName:      "subnet_id",
							terraformAttributeName: "subnet_id",
							children:               nil,
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "sa_key_content",
							terraformAttributeName: "sa_key_content",
							children:               nil,
						},
						{
							protobufFieldName:      "changefeed_custom_name",
							terraformAttributeName: "changefeed_custom_name",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "yds_source",
					terraformAttributeName: "yds_source",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "endpoint",
							terraformAttributeName: "endpoint",
							children:               nil,
						},
						{
							protobufFieldName:      "database",
							terraformAttributeName: "database",
							children:               nil,
						},
						{
							protobufFieldName:      "stream",
							terraformAttributeName: "stream",
							children:               nil,
						},
						{
							protobufFieldName:      "consumer",
							terraformAttributeName: "consumer",
							children:               nil,
						},
						{
							protobufFieldName:      "service_account_id",
							terraformAttributeName: "service_account_id",
							children:               nil,
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "subnet_id",
							terraformAttributeName: "subnet_id",
							children:               nil,
						},
						{
							protobufFieldName:      "parser",
							terraformAttributeName: "parser",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "json_parser",
									terraformAttributeName: "json_parser",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "data_schema",
											terraformAttributeName: "data_schema",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "fields",
													terraformAttributeName: "fields",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "fields",
															terraformAttributeName: "fields",
															children:               nil,
														},
													},
												},
												{
													protobufFieldName:      "json_fields",
													terraformAttributeName: "json_fields",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "null_keys_allowed",
											terraformAttributeName: "null_keys_allowed",
											children:               nil,
										},
										{
											protobufFieldName:      "add_rest_column",
											terraformAttributeName: "add_rest_column",
											children:               nil,
										},
										{
											protobufFieldName:      "unescape_string_values",
											terraformAttributeName: "unescape_string_values",
											children:               nil,
										},
									},
								},
								{
									protobufFieldName:      "audit_trails_v1_parser",
									terraformAttributeName: "audit_trails_v1_parser",
									children:               nil,
								},
								{
									protobufFieldName:      "cloud_logging_parser",
									terraformAttributeName: "cloud_logging_parser",
									children:               nil,
								},
								{
									protobufFieldName:      "tskv_parser",
									terraformAttributeName: "tskv_parser",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "data_schema",
											terraformAttributeName: "data_schema",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "fields",
													terraformAttributeName: "fields",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "fields",
															terraformAttributeName: "fields",
															children:               nil,
														},
													},
												},
												{
													protobufFieldName:      "json_fields",
													terraformAttributeName: "json_fields",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "null_keys_allowed",
											terraformAttributeName: "null_keys_allowed",
											children:               nil,
										},
										{
											protobufFieldName:      "add_rest_column",
											terraformAttributeName: "add_rest_column",
											children:               nil,
										},
										{
											protobufFieldName:      "unescape_string_values",
											terraformAttributeName: "unescape_string_values",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "supported_codecs",
							terraformAttributeName: "supported_codecs",
							children:               nil,
						},
						{
							protobufFieldName:      "allow_ttl_rewind",
							terraformAttributeName: "allow_ttl_rewind",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "kafka_source",
					terraformAttributeName: "kafka_source",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "cluster_id",
									terraformAttributeName: "cluster_id",
									children:               nil,
								},
								{
									protobufFieldName:      "on_premise",
									terraformAttributeName: "on_premise",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "broker_urls",
											terraformAttributeName: "broker_urls",
											children:               nil,
										},
										{
											protobufFieldName:      "tls_mode",
											terraformAttributeName: "tls_mode",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "disabled",
													terraformAttributeName: "disabled",
													children:               nil,
												},
												{
													protobufFieldName:      "enabled",
													terraformAttributeName: "enabled",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "ca_certificate",
															terraformAttributeName: "ca_certificate",
															children:               nil,
														},
													},
												},
											},
										},
										{
											protobufFieldName:      "subnet_id",
											terraformAttributeName: "subnet_id",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "auth",
							terraformAttributeName: "auth",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "sasl",
									terraformAttributeName: "sasl",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "user",
											terraformAttributeName: "user",
											children:               nil,
										},
										{
											protobufFieldName:      "password",
											terraformAttributeName: "password",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "raw",
													terraformAttributeName: "raw",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "mechanism",
											terraformAttributeName: "mechanism",
											children:               nil,
										},
									},
								},
								{
									protobufFieldName:      "no_auth",
									terraformAttributeName: "no_auth",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "topic_name",
							terraformAttributeName: "topic_name",
							children:               nil,
						},
						{
							protobufFieldName:      "transformer",
							terraformAttributeName: "transformer",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "cloud_function",
									terraformAttributeName: "cloud_function",
									children:               nil,
								},
								{
									protobufFieldName:      "service_account_id",
									terraformAttributeName: "service_account_id",
									children:               nil,
								},
								{
									protobufFieldName:      "number_of_retries",
									terraformAttributeName: "number_of_retries",
									children:               nil,
								},
								{
									protobufFieldName:      "buffer_size",
									terraformAttributeName: "buffer_size",
									children:               nil,
								},
								{
									protobufFieldName:      "buffer_flush_interval",
									terraformAttributeName: "buffer_flush_interval",
									children:               nil,
								},
								{
									protobufFieldName:      "invocation_timeout",
									terraformAttributeName: "invocation_timeout",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "parser",
							terraformAttributeName: "parser",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "json_parser",
									terraformAttributeName: "json_parser",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "data_schema",
											terraformAttributeName: "data_schema",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "fields",
													terraformAttributeName: "fields",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "fields",
															terraformAttributeName: "fields",
															children:               nil,
														},
													},
												},
												{
													protobufFieldName:      "json_fields",
													terraformAttributeName: "json_fields",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "null_keys_allowed",
											terraformAttributeName: "null_keys_allowed",
											children:               nil,
										},
										{
											protobufFieldName:      "add_rest_column",
											terraformAttributeName: "add_rest_column",
											children:               nil,
										},
										{
											protobufFieldName:      "unescape_string_values",
											terraformAttributeName: "unescape_string_values",
											children:               nil,
										},
									},
								},
								{
									protobufFieldName:      "audit_trails_v1_parser",
									terraformAttributeName: "audit_trails_v1_parser",
									children:               nil,
								},
								{
									protobufFieldName:      "cloud_logging_parser",
									terraformAttributeName: "cloud_logging_parser",
									children:               nil,
								},
								{
									protobufFieldName:      "tskv_parser",
									terraformAttributeName: "tskv_parser",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "data_schema",
											terraformAttributeName: "data_schema",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "fields",
													terraformAttributeName: "fields",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "fields",
															terraformAttributeName: "fields",
															children:               nil,
														},
													},
												},
												{
													protobufFieldName:      "json_fields",
													terraformAttributeName: "json_fields",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "null_keys_allowed",
											terraformAttributeName: "null_keys_allowed",
											children:               nil,
										},
										{
											protobufFieldName:      "add_rest_column",
											terraformAttributeName: "add_rest_column",
											children:               nil,
										},
										{
											protobufFieldName:      "unescape_string_values",
											terraformAttributeName: "unescape_string_values",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "topic_names",
							terraformAttributeName: "topic_names",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "mongo_source",
					terraformAttributeName: "mongo_source",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "connection_options",
									terraformAttributeName: "connection_options",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "mdb_cluster_id",
											terraformAttributeName: "mdb_cluster_id",
											children:               nil,
										},
										{
											protobufFieldName:      "on_premise",
											terraformAttributeName: "on_premise",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "hosts",
													terraformAttributeName: "hosts",
													children:               nil,
												},
												{
													protobufFieldName:      "port",
													terraformAttributeName: "port",
													children:               nil,
												},
												{
													protobufFieldName:      "tls_mode",
													terraformAttributeName: "tls_mode",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "disabled",
															terraformAttributeName: "disabled",
															children:               nil,
														},
														{
															protobufFieldName:      "enabled",
															terraformAttributeName: "enabled",
															children: []*fieldTreeNode{
																{
																	protobufFieldName:      "ca_certificate",
																	terraformAttributeName: "ca_certificate",
																	children:               nil,
																},
															},
														},
													},
												},
												{
													protobufFieldName:      "replica_set",
													terraformAttributeName: "replica_set",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "user",
											terraformAttributeName: "user",
											children:               nil,
										},
										{
											protobufFieldName:      "password",
											terraformAttributeName: "password",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "raw",
													terraformAttributeName: "raw",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "auth_source",
											terraformAttributeName: "auth_source",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "subnet_id",
							terraformAttributeName: "subnet_id",
							children:               nil,
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "collections",
							terraformAttributeName: "collections",
							children:               nil,
						},
						{
							protobufFieldName:      "excluded_collections",
							terraformAttributeName: "excluded_collections",
							children:               nil,
						},
						{
							protobufFieldName:      "secondary_preferred_mode",
							terraformAttributeName: "secondary_preferred_mode",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "clickhouse_source",
					terraformAttributeName: "clickhouse_source",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "connection_options",
									terraformAttributeName: "connection_options",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "mdb_cluster_id",
											terraformAttributeName: "mdb_cluster_id",
											children:               nil,
										},
										{
											protobufFieldName:      "on_premise",
											terraformAttributeName: "on_premise",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "shards",
													terraformAttributeName: "shards",
													children:               nil,
												},
												{
													protobufFieldName:      "http_port",
													terraformAttributeName: "http_port",
													children:               nil,
												},
												{
													protobufFieldName:      "native_port",
													terraformAttributeName: "native_port",
													children:               nil,
												},
												{
													protobufFieldName:      "tls_mode",
													terraformAttributeName: "tls_mode",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "disabled",
															terraformAttributeName: "disabled",
															children:               nil,
														},
														{
															protobufFieldName:      "enabled",
															terraformAttributeName: "enabled",
															children: []*fieldTreeNode{
																{
																	protobufFieldName:      "ca_certificate",
																	terraformAttributeName: "ca_certificate",
																	children:               nil,
																},
															},
														},
													},
												},
											},
										},
										{
											protobufFieldName:      "database",
											terraformAttributeName: "database",
											children:               nil,
										},
										{
											protobufFieldName:      "user",
											terraformAttributeName: "user",
											children:               nil,
										},
										{
											protobufFieldName:      "password",
											terraformAttributeName: "password",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "raw",
													terraformAttributeName: "raw",
													children:               nil,
												},
											},
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "subnet_id",
							terraformAttributeName: "subnet_id",
							children:               nil,
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "clickhouse_cluster_name",
							terraformAttributeName: "clickhouse_cluster_name",
							children:               nil,
						},
						{
							protobufFieldName:      "include_tables",
							terraformAttributeName: "include_tables",
							children:               nil,
						},
						{
							protobufFieldName:      "exclude_tables",
							terraformAttributeName: "exclude_tables",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "mysql_target",
					terraformAttributeName: "mysql_target",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "mdb_cluster_id",
									terraformAttributeName: "mdb_cluster_id",
									children:               nil,
								},
								{
									protobufFieldName:      "on_premise",
									terraformAttributeName: "on_premise",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "hosts",
											terraformAttributeName: "hosts",
											children:               nil,
										},
										{
											protobufFieldName:      "port",
											terraformAttributeName: "port",
											children:               nil,
										},
										{
											protobufFieldName:      "tls_mode",
											terraformAttributeName: "tls_mode",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "disabled",
													terraformAttributeName: "disabled",
													children:               nil,
												},
												{
													protobufFieldName:      "enabled",
													terraformAttributeName: "enabled",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "ca_certificate",
															terraformAttributeName: "ca_certificate",
															children:               nil,
														},
													},
												},
											},
										},
										{
											protobufFieldName:      "subnet_id",
											terraformAttributeName: "subnet_id",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "database",
							terraformAttributeName: "database",
							children:               nil,
						},
						{
							protobufFieldName:      "user",
							terraformAttributeName: "user",
							children:               nil,
						},
						{
							protobufFieldName:      "password",
							terraformAttributeName: "password",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "raw",
									terraformAttributeName: "raw",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "sql_mode",
							terraformAttributeName: "sql_mode",
							children:               nil,
						},
						{
							protobufFieldName:      "skip_constraint_checks",
							terraformAttributeName: "skip_constraint_checks",
							children:               nil,
						},
						{
							protobufFieldName:      "timezone",
							terraformAttributeName: "timezone",
							children:               nil,
						},
						{
							protobufFieldName:      "cleanup_policy",
							terraformAttributeName: "cleanup_policy",
							children:               nil,
						},
						{
							protobufFieldName:      "service_database",
							terraformAttributeName: "service_database",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "postgres_target",
					terraformAttributeName: "postgres_target",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "mdb_cluster_id",
									terraformAttributeName: "mdb_cluster_id",
									children:               nil,
								},
								{
									protobufFieldName:      "on_premise",
									terraformAttributeName: "on_premise",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "hosts",
											terraformAttributeName: "hosts",
											children:               nil,
										},
										{
											protobufFieldName:      "port",
											terraformAttributeName: "port",
											children:               nil,
										},
										{
											protobufFieldName:      "tls_mode",
											terraformAttributeName: "tls_mode",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "disabled",
													terraformAttributeName: "disabled",
													children:               nil,
												},
												{
													protobufFieldName:      "enabled",
													terraformAttributeName: "enabled",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "ca_certificate",
															terraformAttributeName: "ca_certificate",
															children:               nil,
														},
													},
												},
											},
										},
										{
											protobufFieldName:      "subnet_id",
											terraformAttributeName: "subnet_id",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "database",
							terraformAttributeName: "database",
							children:               nil,
						},
						{
							protobufFieldName:      "user",
							terraformAttributeName: "user",
							children:               nil,
						},
						{
							protobufFieldName:      "password",
							terraformAttributeName: "password",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "raw",
									terraformAttributeName: "raw",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "cleanup_policy",
							terraformAttributeName: "cleanup_policy",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "clickhouse_target",
					terraformAttributeName: "clickhouse_target",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "connection_options",
									terraformAttributeName: "connection_options",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "mdb_cluster_id",
											terraformAttributeName: "mdb_cluster_id",
											children:               nil,
										},
										{
											protobufFieldName:      "on_premise",
											terraformAttributeName: "on_premise",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "shards",
													terraformAttributeName: "shards",
													children:               nil,
												},
												{
													protobufFieldName:      "http_port",
													terraformAttributeName: "http_port",
													children:               nil,
												},
												{
													protobufFieldName:      "native_port",
													terraformAttributeName: "native_port",
													children:               nil,
												},
												{
													protobufFieldName:      "tls_mode",
													terraformAttributeName: "tls_mode",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "disabled",
															terraformAttributeName: "disabled",
															children:               nil,
														},
														{
															protobufFieldName:      "enabled",
															terraformAttributeName: "enabled",
															children: []*fieldTreeNode{
																{
																	protobufFieldName:      "ca_certificate",
																	terraformAttributeName: "ca_certificate",
																	children:               nil,
																},
															},
														},
													},
												},
											},
										},
										{
											protobufFieldName:      "database",
											terraformAttributeName: "database",
											children:               nil,
										},
										{
											protobufFieldName:      "user",
											terraformAttributeName: "user",
											children:               nil,
										},
										{
											protobufFieldName:      "password",
											terraformAttributeName: "password",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "raw",
													terraformAttributeName: "raw",
													children:               nil,
												},
											},
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "subnet_id",
							terraformAttributeName: "subnet_id",
							children:               nil,
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "clickhouse_cluster_name",
							terraformAttributeName: "clickhouse_cluster_name",
							children:               nil,
						},
						{
							protobufFieldName:      "alt_names",
							terraformAttributeName: "alt_names",
							children:               nil,
						},
						{
							protobufFieldName:      "sharding",
							terraformAttributeName: "sharding",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "column_value_hash",
									terraformAttributeName: "column_value_hash",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "column_name",
											terraformAttributeName: "column_name",
											children:               nil,
										},
									},
								},
								{
									protobufFieldName:      "custom_mapping",
									terraformAttributeName: "custom_mapping",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "column_name",
											terraformAttributeName: "column_name",
											children:               nil,
										},
										{
											protobufFieldName:      "mapping",
											terraformAttributeName: "mapping",
											children:               nil,
										},
									},
								},
								{
									protobufFieldName:      "transfer_id",
									terraformAttributeName: "transfer_id",
									children:               nil,
								},
								{
									protobufFieldName:      "round_robin",
									terraformAttributeName: "round_robin",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "cleanup_policy",
							terraformAttributeName: "cleanup_policy",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "ydb_target",
					terraformAttributeName: "ydb_target",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "database",
							terraformAttributeName: "database",
							children:               nil,
						},
						{
							protobufFieldName:      "instance",
							terraformAttributeName: "instance",
							children:               nil,
						},
						{
							protobufFieldName:      "service_account_id",
							terraformAttributeName: "service_account_id",
							children:               nil,
						},
						{
							protobufFieldName:      "path",
							terraformAttributeName: "path",
							children:               nil,
						},
						{
							protobufFieldName:      "subnet_id",
							terraformAttributeName: "subnet_id",
							children:               nil,
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "sa_key_content",
							terraformAttributeName: "sa_key_content",
							children:               nil,
						},
						{
							protobufFieldName:      "cleanup_policy",
							terraformAttributeName: "cleanup_policy",
							children:               nil,
						},
						{
							protobufFieldName:      "is_table_column_oriented",
							terraformAttributeName: "is_table_column_oriented",
							children:               nil,
						},
						{
							protobufFieldName:      "default_compression",
							terraformAttributeName: "default_compression",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "kafka_target",
					terraformAttributeName: "kafka_target",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "cluster_id",
									terraformAttributeName: "cluster_id",
									children:               nil,
								},
								{
									protobufFieldName:      "on_premise",
									terraformAttributeName: "on_premise",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "broker_urls",
											terraformAttributeName: "broker_urls",
											children:               nil,
										},
										{
											protobufFieldName:      "tls_mode",
											terraformAttributeName: "tls_mode",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "disabled",
													terraformAttributeName: "disabled",
													children:               nil,
												},
												{
													protobufFieldName:      "enabled",
													terraformAttributeName: "enabled",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "ca_certificate",
															terraformAttributeName: "ca_certificate",
															children:               nil,
														},
													},
												},
											},
										},
										{
											protobufFieldName:      "subnet_id",
											terraformAttributeName: "subnet_id",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "auth",
							terraformAttributeName: "auth",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "sasl",
									terraformAttributeName: "sasl",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "user",
											terraformAttributeName: "user",
											children:               nil,
										},
										{
											protobufFieldName:      "password",
											terraformAttributeName: "password",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "raw",
													terraformAttributeName: "raw",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "mechanism",
											terraformAttributeName: "mechanism",
											children:               nil,
										},
									},
								},
								{
									protobufFieldName:      "no_auth",
									terraformAttributeName: "no_auth",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "topic_settings",
							terraformAttributeName: "topic_settings",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "topic",
									terraformAttributeName: "topic",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "topic_name",
											terraformAttributeName: "topic_name",
											children:               nil,
										},
										{
											protobufFieldName:      "save_tx_order",
											terraformAttributeName: "save_tx_order",
											children:               nil,
										},
									},
								},
								{
									protobufFieldName:      "topic_prefix",
									terraformAttributeName: "topic_prefix",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "serializer",
							terraformAttributeName: "serializer",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "serializer_auto",
									terraformAttributeName: "serializer_auto",
									children:               nil,
								},
								{
									protobufFieldName:      "serializer_json",
									terraformAttributeName: "serializer_json",
									children:               nil,
								},
								{
									protobufFieldName:      "serializer_debezium",
									terraformAttributeName: "serializer_debezium",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "serializer_parameters",
											terraformAttributeName: "serializer_parameters",
											children:               nil,
										},
									},
								},
							},
						},
					},
				},
				{
					protobufFieldName:      "mongo_target",
					terraformAttributeName: "mongo_target",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "connection",
							terraformAttributeName: "connection",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "connection_options",
									terraformAttributeName: "connection_options",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "mdb_cluster_id",
											terraformAttributeName: "mdb_cluster_id",
											children:               nil,
										},
										{
											protobufFieldName:      "on_premise",
											terraformAttributeName: "on_premise",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "hosts",
													terraformAttributeName: "hosts",
													children:               nil,
												},
												{
													protobufFieldName:      "port",
													terraformAttributeName: "port",
													children:               nil,
												},
												{
													protobufFieldName:      "tls_mode",
													terraformAttributeName: "tls_mode",
													children: []*fieldTreeNode{
														{
															protobufFieldName:      "disabled",
															terraformAttributeName: "disabled",
															children:               nil,
														},
														{
															protobufFieldName:      "enabled",
															terraformAttributeName: "enabled",
															children: []*fieldTreeNode{
																{
																	protobufFieldName:      "ca_certificate",
																	terraformAttributeName: "ca_certificate",
																	children:               nil,
																},
															},
														},
													},
												},
												{
													protobufFieldName:      "replica_set",
													terraformAttributeName: "replica_set",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "user",
											terraformAttributeName: "user",
											children:               nil,
										},
										{
											protobufFieldName:      "password",
											terraformAttributeName: "password",
											children: []*fieldTreeNode{
												{
													protobufFieldName:      "raw",
													terraformAttributeName: "raw",
													children:               nil,
												},
											},
										},
										{
											protobufFieldName:      "auth_source",
											terraformAttributeName: "auth_source",
											children:               nil,
										},
									},
								},
							},
						},
						{
							protobufFieldName:      "subnet_id",
							terraformAttributeName: "subnet_id",
							children:               nil,
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "database",
							terraformAttributeName: "database",
							children:               nil,
						},
						{
							protobufFieldName:      "cleanup_policy",
							terraformAttributeName: "cleanup_policy",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "metrika_source",
					terraformAttributeName: "metrika_source",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "counter_ids",
							terraformAttributeName: "counter_ids",
							children:               nil,
						},
						{
							protobufFieldName:      "token",
							terraformAttributeName: "token",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "raw",
									terraformAttributeName: "raw",
									children:               nil,
								},
							},
						},
						{
							protobufFieldName:      "streams",
							terraformAttributeName: "streams",
							children:               nil,
						},
					},
				},
				{
					protobufFieldName:      "yds_target",
					terraformAttributeName: "yds_target",
					children: []*fieldTreeNode{
						{
							protobufFieldName:      "endpoint",
							terraformAttributeName: "endpoint",
							children:               nil,
						},
						{
							protobufFieldName:      "database",
							terraformAttributeName: "database",
							children:               nil,
						},
						{
							protobufFieldName:      "stream",
							terraformAttributeName: "stream",
							children:               nil,
						},
						{
							protobufFieldName:      "service_account_id",
							terraformAttributeName: "service_account_id",
							children:               nil,
						},
						{
							protobufFieldName:      "security_groups",
							terraformAttributeName: "security_groups",
							children:               nil,
						},
						{
							protobufFieldName:      "subnet_id",
							terraformAttributeName: "subnet_id",
							children:               nil,
						},
						{
							protobufFieldName:      "save_tx_order",
							terraformAttributeName: "save_tx_order",
							children:               nil,
						},
						{
							protobufFieldName:      "serializer",
							terraformAttributeName: "serializer",
							children: []*fieldTreeNode{
								{
									protobufFieldName:      "serializer_auto",
									terraformAttributeName: "serializer_auto",
									children:               nil,
								},
								{
									protobufFieldName:      "serializer_json",
									terraformAttributeName: "serializer_json",
									children:               nil,
								},
								{
									protobufFieldName:      "serializer_debezium",
									terraformAttributeName: "serializer_debezium",
									children: []*fieldTreeNode{
										{
											protobufFieldName:      "serializer_parameters",
											terraformAttributeName: "serializer_parameters",
											children:               nil,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
