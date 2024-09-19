package yandex

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	advanced_rate_limiter "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/advanced_rate_limiter"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dataSourceYandexSmartwebsecurityAdvancedRateLimiterAdvancedRateLimiterProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexSmartwebsecurityAdvancedRateLimiterAdvancedRateLimiterProfileRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"advanced_rate_limiter_profile_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"advanced_rate_limiter_rule": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"dry_run": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"dynamic_quota": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"characteristic": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"case_insensitive": {
													Type:     schema.TypeBool,
													Computed: true,
												},

												"key_characteristic": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"type": {
																Type:     schema.TypeString,
																Computed: true,
															},

															"value": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
													Computed: true,
												},

												"simple_characteristic": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"type": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
													Computed: true,
												},
											},
										},
										Computed: true,
									},

									"condition": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"authority": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"authorities": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"exact_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},

												"headers": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},

															"value": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"exact_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},

												"http_method": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"http_methods": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"exact_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},

												"request_uri": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"path": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"exact_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},

															"queries": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"key": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"value": {
																			Type: schema.TypeList,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"exact_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"exact_not_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"pire_regex_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"pire_regex_not_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"prefix_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"prefix_not_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},
																				},
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},

												"source_ip": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"geo_ip_match": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"locations": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},

															"geo_ip_not_match": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"locations": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},

															"ip_ranges_match": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"ip_ranges": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},

															"ip_ranges_not_match": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"ip_ranges": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},
											},
										},
										Computed: true,
									},

									"limit": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"period": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"priority": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"static_quota": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"condition": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"authority": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"authorities": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"exact_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},

												"headers": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:     schema.TypeString,
																Computed: true,
															},

															"value": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"exact_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},

												"http_method": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"http_methods": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"exact_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},

												"request_uri": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"path": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"exact_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"pire_regex_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix_not_match": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},

															"queries": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"key": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"value": {
																			Type: schema.TypeList,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"exact_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"exact_not_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"pire_regex_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"pire_regex_not_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"prefix_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"prefix_not_match": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},
																				},
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},

												"source_ip": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"geo_ip_match": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"locations": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},

															"geo_ip_not_match": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"locations": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},

															"ip_ranges_match": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"ip_ranges": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},

															"ip_ranges_not_match": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"ip_ranges": {
																			Type: schema.TypeList,
																			Elem: &schema.Schema{
																				Type: schema.TypeString,
																			},
																			Computed: true,
																		},
																	},
																},
																Computed: true,
															},
														},
													},
													Computed: true,
												},
											},
										},
										Computed: true,
									},

									"limit": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"period": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
							Computed: true,
						},
					},
				},
				Computed: true,
			},

			"cloud_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([-_0-9a-z]*)$"), ""), validation.StringLenBetween(0, 63)),
				},
				Set:      schema.HashString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexSmartwebsecurityAdvancedRateLimiterAdvancedRateLimiterProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	err := checkOneOf(d, "advanced_rate_limiter_profile_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	resourceID := d.Get("advanced_rate_limiter_profile_id").(string)
	_, nameOk := d.GetOk("name")
	if nameOk {
		resourceID, err = resolveObjectID(config.Context(), config, d, sdkresolvers.SWSAdvancedRateLimiterProfileResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source AdvancedRateLimiterProfile by name: %v", err)
		}
	}

	req := &advanced_rate_limiter.GetAdvancedRateLimiterProfileRequest{
		AdvancedRateLimiterProfileId: resourceID,
	}

	md := new(metadata.MD)

	resp, err := config.sdk.SmartWebSecurityArl().AdvancedRateLimiterProfile().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read AdvancedRateLimiterProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read AdvancedRateLimiterProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("advanced_rate_limiter_profile %q", d.Get("advanced_rate_limiter_profile_id").(string))))
	}

	d.SetId(resp.Id)

	log.Printf("[DEBUG] Read AdvancedRateLimiterProfile response: %s", protoDump(resp))

	advancedRateLimiterRule, err := flattenAdvancedXrateXlimiterAdvancedRateLimiterRuleSlice(resp.GetAdvancedRateLimiterRules())
	if err != nil { // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	createdAt := getTimestamp(resp.GetCreatedAt())

	if err := d.Set("advanced_rate_limiter_profile_id", resp.GetId()); err != nil {
		log.Printf("[ERROR] failed set field advanced_rate_limiter_profile_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("advanced_rate_limiter_rule", advancedRateLimiterRule); err != nil {
		log.Printf("[ERROR] failed set field advanced_rate_limiter_rule: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("cloud_id", resp.GetCloudId()); err != nil {
		log.Printf("[ERROR] failed set field cloud_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", createdAt); err != nil {
		log.Printf("[ERROR] failed set field created_at: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("description", resp.GetDescription()); err != nil {
		log.Printf("[ERROR] failed set field description: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("folder_id", resp.GetFolderId()); err != nil {
		log.Printf("[ERROR] failed set field folder_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("labels", resp.GetLabels()); err != nil {
		log.Printf("[ERROR] failed set field labels: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("name", resp.GetName()); err != nil {
		log.Printf("[ERROR] failed set field name: %s", err)
		return diag.FromErr(err)
	}

	return nil
}
