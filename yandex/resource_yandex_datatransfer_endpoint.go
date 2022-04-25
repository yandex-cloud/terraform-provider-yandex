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
							ConflictsWith: []string{"settings.0.clickhouse_target", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target"},
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
													ConflictsWith: []string{"settings.0.clickhouse_target.0.sharding.0.transfer_id"},
												},
												"transfer_id": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
													Optional:      true,
													ConflictsWith: []string{"settings.0.clickhouse_target.0.sharding.0.column_value_hash"},
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
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target"},
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
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target"},
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
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.mongo_source", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target"},
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
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_target", "settings.0.postgres_source", "settings.0.postgres_target"},
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
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.postgres_source", "settings.0.postgres_target"},
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
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_target"},
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
							ConflictsWith: []string{"settings.0.clickhouse_source", "settings.0.clickhouse_target", "settings.0.mongo_source", "settings.0.mongo_target", "settings.0.mysql_source", "settings.0.mysql_target", "settings.0.postgres_source"},
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

	updatePath := generateEndpointFieldMasks(d, datatransferUpdateEndpointRequestFieldsRoot)
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
								{
									protobufFieldName:      "materialized_view",
									terraformAttributeName: "materialized_view",
									children:               nil,
								},
							},
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
			},
		},
	},
}
