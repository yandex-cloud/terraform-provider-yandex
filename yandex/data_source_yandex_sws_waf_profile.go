package yandex

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	waf "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/waf"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dataSourceYandexSmartwebsecurityWafWafProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexSmartwebsecurityWafWafProfileRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"analyze_request_body": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"size_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"size_limit_action": {
							Type:     schema.TypeString,
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

			"core_rule_set": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inbound_anomaly_score": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"paranoia_level": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"rule_set": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},

									"version": {
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

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"exclusion_rule": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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

						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"exclude_rules": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"exclude_all": {
										Type:     schema.TypeBool,
										Computed: true,
									},

									"rule_ids": {
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

						"log_excluded": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

			"rule": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_blocking": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"is_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"rule_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},

			"waf_profile_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceYandexSmartwebsecurityWafWafProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	err := checkOneOf(d, "waf_profile_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	resourceID := d.Get("waf_profile_id").(string)
	_, nameOk := d.GetOk("name")
	if nameOk {
		resourceID, err = resolveObjectID(config.Context(), config, d, sdkresolvers.SWSWafProfileResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source WafProfile by name: %v", err)
		}
	}

	req := &waf.GetWafProfileRequest{
		WafProfileId: resourceID,
	}

	md := new(metadata.MD)

	resp, err := config.sdk.SmartWebSecurityWaf().WafProfile().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read WafProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read WafProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("waf_profile %q", d.Get("waf_profile_id").(string))))
	}

	d.SetId(resp.Id)

	log.Printf("[DEBUG] Read WafProfile response: %s", protoDump(resp))

	analyzeRequestBody, err := flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfile_AnalyzeRequestBody(resp.GetAnalyzeRequestBody())
	if err != nil {
		// B // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	coreRuleSet, err := flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfile_CoreRuleSet(resp.GetCoreRuleSet())
	if err != nil {
		// B // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	createdAt := getTimestamp(resp.GetCreatedAt())

	exclusionRule, err := flattenWafExclusionRuleSlice(resp.GetExclusionRules())
	if err != nil { // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	rule, err := flattenWafRuleSlice(resp.GetRules())
	if err != nil { // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	if err := d.Set("analyze_request_body", analyzeRequestBody); err != nil {
		log.Printf("[ERROR] failed set field analyze_request_body: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("cloud_id", resp.GetCloudId()); err != nil {
		log.Printf("[ERROR] failed set field cloud_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("core_rule_set", coreRuleSet); err != nil {
		log.Printf("[ERROR] failed set field core_rule_set: %s", err)
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
	if err := d.Set("exclusion_rule", exclusionRule); err != nil {
		log.Printf("[ERROR] failed set field exclusion_rule: %s", err)
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
	if err := d.Set("rule", rule); err != nil {
		log.Printf("[ERROR] failed set field rule: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("waf_profile_id", resp.GetId()); err != nil {
		log.Printf("[ERROR] failed set field waf_profile_id: %s", err)
		return diag.FromErr(err)
	}

	return nil
}
