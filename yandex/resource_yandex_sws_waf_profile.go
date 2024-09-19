package yandex

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	waf "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/waf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func resourceYandexSmartwebsecurityWafWafProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexSmartwebsecurityWafWafProfileCreate,
		ReadContext:   resourceYandexSmartwebsecurityWafWafProfileRead,
		UpdateContext: resourceYandexSmartwebsecurityWafWafProfileUpdate,
		DeleteContext: resourceYandexSmartwebsecurityWafWafProfileDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Read:   schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"analyze_request_body": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"size_limit": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"size_limit_action": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateParsableValue(parseWafWafProfileXAnalyzeRequestBodyXAction),
						},
					},
				},
				Optional: true,
			},

			"cloud_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"core_rule_set": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inbound_anomaly_score": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 10000),
						},

						"paranoia_level": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"rule_set": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"version": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
				Optional: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},

			"exclusion_rule": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authority": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"authorities": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"exact_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},

									"headers": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 255),
												},

												"value": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"exact_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},
														},
													},
													Required: true,
												},
											},
										},
										Optional: true,
									},

									"http_method": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_methods": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"exact_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},

									"request_uri": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"path": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"exact_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_not_match": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},
														},
													},
													Optional: true,
												},

												"queries": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"key": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 255),
															},

															"value": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"exact_not_match": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"pire_regex_match": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"pire_regex_not_match": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"prefix_match": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"prefix_not_match": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},
																	},
																},
																Required: true,
															},
														},
													},
													Optional: true,
												},
											},
										},
										Optional: true,
									},

									"source_ip": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"geo_ip_match": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"locations": {
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

												"geo_ip_not_match": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"locations": {
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

												"ip_ranges_match": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ip_ranges": {
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

												"ip_ranges_not_match": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ip_ranges": {
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

						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 512),
						},

						"exclude_rules": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"exclude_all": {
										Type:     schema.TypeBool,
										Optional: true,
									},

									"rule_ids": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
									},
								},
							},
							Required: true,
						},

						"log_excluded": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
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
					Type:         schema.TypeString,
					ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([-_0-9a-z]*)$"), ""), validation.StringLenBetween(0, 63)),
				},
				Set:      schema.HashString,
				Optional: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([a-zA-Z0-9][a-zA-Z0-9-_.]*)$"), ""), validation.StringLenBetween(1, 50)),
			},

			"rule": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_blocking": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"is_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"rule_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Optional: true,
			},
		},
	}
}

func resourceYandexSmartwebsecurityWafWafProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	folderId, err := getFolderID(d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	labels := expandStringStringMap(d.Get("labels").(map[string]interface{}))
	rules, err := expandWafProfileRulesSlice(d)
	if err != nil {
		return diag.FromErr(err)
	}

	exclusionRules, err := expandWafProfileExclusionRulesSlice(d)
	if err != nil {
		return diag.FromErr(err)
	}

	analyzeRequestBody, err := expandWafProfileAnalyzeRequestBody(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &waf.CreateWafProfileRequest{
		FolderId:           folderId,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		Rules:              rules,
		ExclusionRules:     exclusionRules,
		AnalyzeRequestBody: analyzeRequestBody,
	}

	if _, ok := d.GetOk("core_rule_set"); ok {
		coreRuleSet, err := expandWafProfileCoreRuleSet(d)
		if err != nil {
			return diag.FromErr(err)
		}

		req.SetCoreRuleSet(coreRuleSet)
	}

	log.Printf("[DEBUG] Create WafProfile request: %s", protoDump(req))

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.SmartWebSecurityWaf().WafProfile().Create(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create WafProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create WafProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while get waf.WafProfile create operation metadata: %v", err)
	}

	createMetadata, ok := protoMetadata.(*waf.CreateWafProfileMetadata)
	if !ok {
		return diag.Errorf("could not get WafProfile ID from create operation metadata")
	}

	d.SetId(createMetadata.WafProfileId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexSmartwebsecurityWafWafProfileRead(ctx, d, meta)
}

func resourceYandexSmartwebsecurityWafWafProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &waf.GetWafProfileRequest{
		WafProfileId: d.Id(),
	}

	log.Printf("[DEBUG] Read WafProfile request: %s", protoDump(req))

	md := new(metadata.MD)
	resp, err := config.sdk.SmartWebSecurityWaf().WafProfile().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read WafProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read WafProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("waf_profile %q", d.Id())))
	}

	log.Printf("[DEBUG] Read WafProfile response: %s", protoDump(resp))

	analyzeRequestBody, err := flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfile_AnalyzeRequestBody(resp.GetAnalyzeRequestBody())
	if err != nil {
		// B // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	coreRuleSet_, err := flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfile_CoreRuleSet(resp.GetCoreRuleSet())
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
	if err := d.Set("core_rule_set", coreRuleSet_); err != nil {
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

	return nil
}

func resourceYandexSmartwebsecurityWafWafProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	labels := expandStringStringMap(d.Get("labels").(map[string]interface{}))
	rules, err := expandWafProfileRulesSlice_(d)
	if err != nil {
		return diag.FromErr(err)
	}

	exclusionRules, err := expandWafProfileExclusionRulesSlice_(d)
	if err != nil {
		return diag.FromErr(err)
	}

	analyzeRequestBody, err := expandWafProfileAnalyzeRequestBody_(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &waf.UpdateWafProfileRequest{
		WafProfileId:       d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		Rules:              rules,
		ExclusionRules:     exclusionRules,
		AnalyzeRequestBody: analyzeRequestBody,
	}

	if _, ok := d.GetOk("core_rule_set"); ok {
		updateWafProfileRequestCoreRuleSet, err := expandWafProfileCoreRuleSet_(d)
		if err != nil {
			return diag.FromErr(err)
		}

		req.SetCoreRuleSet(updateWafProfileRequestCoreRuleSet)
	}

	updatePath := generateFieldMasks(d, resourceYandexSmartwebsecurityWafWafProfileUpdateFieldsMap)
	req.UpdateMask = &fieldmaskpb.FieldMask{Paths: updatePath}

	log.Printf("[DEBUG] Update WafProfile request: %s", protoDump(req))

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.SmartWebSecurityWaf().WafProfile().Update(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update WafProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update WafProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexSmartwebsecurityWafWafProfileRead(ctx, d, meta)
}

func resourceYandexSmartwebsecurityWafWafProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &waf.DeleteWafProfileRequest{
		WafProfileId: d.Id(),
	}

	log.Printf("[DEBUG] Delete WafProfile request: %s", protoDump(req))

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.SmartWebSecurityWaf().WafProfile().Delete(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete WafProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete WafProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("waf_profile %q", d.Id())))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

var resourceYandexSmartwebsecurityWafWafProfileUpdateFieldsMap = map[string]string{
	"name":                                     "name",
	"description":                              "description",
	"labels":                                   "labels",
	"rule":                                     "rules",
	"exclusion_rule":                           "exclusion_rules",
	"core_rule_set.0.inbound_anomaly_score":    "core_rule_set.inbound_anomaly_score",
	"core_rule_set.0.paranoia_level":           "core_rule_set.paranoia_level",
	"core_rule_set.0.rule_set.0.name":          "core_rule_set.rule_set.name",
	"core_rule_set.0.rule_set.0.version":       "core_rule_set.rule_set.version",
	"analyze_request_body.0.is_enabled":        "analyze_request_body.is_enabled",
	"analyze_request_body.0.size_limit":        "analyze_request_body.size_limit",
	"analyze_request_body.0.size_limit_action": "analyze_request_body.size_limit_action",
}
