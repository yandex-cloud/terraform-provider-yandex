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
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func resourceYandexSmartwebsecurityWafWafProfile() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a WAF Profile in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/quickstart#waf).",

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
				Type:        schema.TypeList,
				Description: "Parameters for request body analyzer.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_enabled": {
							Type:        schema.TypeBool,
							Description: "Possible to turn analyzer on and turn if off.",
							Optional:    true,
						},

						"size_limit": {
							Type:        schema.TypeInt,
							Description: "Maximum size of body to pass to analyzer. In kilobytes.",
							Optional:    true,
						},

						"size_limit_action": {
							Type:         schema.TypeString,
							Description:  "Action to perform if maximum size of body exceeded. Possible values: `IGNORE` and `DENY`.",
							Optional:     true,
							ValidateFunc: validateParsableValue(parseWafWafProfileXAnalyzeRequestBodyXAction),
						},
					},
				},
				Optional: true,
			},

			"cloud_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["cloud_id"],
				Computed:    true,
				Optional:    true,
			},

			"core_rule_set": {
				Type:        schema.TypeList,
				Description: "Core rule set settings. See [Basic rule set](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#rules-set) for details.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inbound_anomaly_score": {
							Type:         schema.TypeInt,
							Description:  "Anomaly score. Enter an integer within the range of 2 and 10000. The higher this value, the more likely it is that the request that satisfies the rule is an attack. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#anomaly) for more details.",
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 10000),
						},

						"paranoia_level": {
							Type:        schema.TypeInt,
							Description: "Paranoia level. Enter an integer within the range of 1 and 4. Paranoia level classifies rules according to their aggression. The higher the paranoia level, the better your protection, but also the higher the probability of WAF false positives. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#paranoia) for more details. NOTE: this option has no effect on enabling or disabling rules, it is used only as recommendation for user to enable all rules with paranoia_level <= this value.",
							Optional:    true,
						},

						"rule_set": {
							Type:        schema.TypeList,
							Description: "Rule set settings. See [Basic rule set](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#rules-set) for details.",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Description: "Name of the rule set.",
										Optional:    true,
									},

									"type": {
										Type:         schema.TypeString,
										Description:  "Type of the rule set.",
										Optional:     true,
										ValidateFunc: validateParsableValue(parseWafRuleSetXRuleSetType),
									},

									"version": {
										Type:        schema.TypeString,
										Description: "Version of the rule set.",
										Required:    true,
									},

									"id": {
										Type:        schema.TypeString,
										Description: "Id of the rule set.",
										Optional:    true,
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
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"description": {
				Type:         schema.TypeString,
				Description:  common.ResourceDescriptions["description"],
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},

			"exclusion_rule": {
				Type:        schema.TypeList,
				Description: "List of exclusion rules. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#exclusion-rules).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:        schema.TypeList,
							Description: "",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authority": {
										Type:        schema.TypeList,
										Description: "",
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"authorities": {
													Type:        schema.TypeList,
													Description: "",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"exact_not_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_not_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_not_match": {
																Type:         schema.TypeString,
																Description:  "",
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
										Type:        schema.TypeList,
										Description: "",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:         schema.TypeString,
													Description:  "",
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 255),
												},

												"value": {
													Type:        schema.TypeList,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"exact_not_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_not_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_not_match": {
																Type:         schema.TypeString,
																Description:  "",
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
										Type:        schema.TypeList,
										Description: "",
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_methods": {
													Type:        schema.TypeList,
													Description: "",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"exact_not_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_not_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_not_match": {
																Type:         schema.TypeString,
																Description:  "",
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
										Type:        schema.TypeList,
										Description: "",
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"path": {
													Type:        schema.TypeList,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"exact_not_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"pire_regex_not_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},

															"prefix_not_match": {
																Type:         schema.TypeString,
																Description:  "",
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 255),
															},
														},
													},
													Optional: true,
												},

												"queries": {
													Type:        schema.TypeList,
													Description: "",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"key": {
																Type:         schema.TypeString,
																Description:  "",
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 255),
															},

															"value": {
																Type:        schema.TypeList,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact_match": {
																			Type:         schema.TypeString,
																			Description:  "",
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"exact_not_match": {
																			Type:         schema.TypeString,
																			Description:  "",
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"pire_regex_match": {
																			Type:         schema.TypeString,
																			Description:  "",
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"pire_regex_not_match": {
																			Type:         schema.TypeString,
																			Description:  "",
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"prefix_match": {
																			Type:         schema.TypeString,
																			Description:  "",
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},

																		"prefix_not_match": {
																			Type:         schema.TypeString,
																			Description:  "",
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
										Type:        schema.TypeList,
										Description: "Source IP.",
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"geo_ip_match": {
													Type:        schema.TypeList,
													Description: "Locations to include.",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"locations": {
																Type:        schema.TypeList,
																Description: "Locations to include.",
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
													Type:        schema.TypeList,
													Description: "Locations to exclude.",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"locations": {
																Type:        schema.TypeList,
																Description: "Locations to exclude.",
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
													Type:        schema.TypeList,
													Description: "IP ranges to include.",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ip_ranges": {
																Type:        schema.TypeList,
																Description: "IP ranges to include.",
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
													Type:        schema.TypeList,
													Description: "IP ranges to exclude.",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ip_ranges": {
																Type:        schema.TypeList,
																Description: "IP ranges to exclude.",
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
							Description:  "Description of the rule. 0-512 characters long.",
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 512),
						},

						"exclude_rules": {
							Type:        schema.TypeList,
							Description: "Exclude rules.",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"exclude_all": {
										Type:        schema.TypeBool,
										Description: "Set this option true to exclude all rules.",
										Optional:    true,
									},

									"rule_ids": {
										Type:        schema.TypeList,
										Description: "List of rules to exclude.",
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
							Type:        schema.TypeBool,
							Description: "Records the fact that an exception rule is triggered.",
							Optional:    true,
						},

						"name": {
							Type:        schema.TypeString,
							Description: "Name of exclusion rule.",
							Optional:    true,
						},
					},
				},
				Optional: true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([-_0-9a-z]*)$"), ""), validation.StringLenBetween(0, 63)),
				},
				Set:      schema.HashString,
				Optional: true,
			},

			"match_all_rule_sets": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["match_all_rule_sets"],
				Optional:    true,
			},

			"name": {
				Type:         schema.TypeString,
				Description:  common.ResourceDescriptions["name"],
				Optional:     true,
				ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([a-zA-Z0-9][a-zA-Z0-9-_.]*)$"), ""), validation.StringLenBetween(1, 50)),
			},

			"rule": {
				Type:        schema.TypeList,
				Description: "Settings for each rule in rule set.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_blocking": {
							Type:        schema.TypeBool,
							Description: "Determines is it rule blocking or not.",
							Optional:    true,
						},

						"is_enabled": {
							Type:        schema.TypeBool,
							Description: "Determines is it rule enabled or not.",
							Optional:    true,
						},

						"rule_id": {
							Type:        schema.TypeString,
							Description: "Rule ID.",
							Required:    true,
						},
					},
				},
				Optional: true,
			},

			"rule_set": {
				Type:        schema.TypeList,
				Description: "Rule set.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:         schema.TypeString,
							Description:  "Action of the rule set.",
							Optional:     true,
							ValidateFunc: validateParsableValue(parseWafWafProfileXWafProfileRuleSetXRuleSetAction),
						},

						"core_rule_set": {
							Type:        schema.TypeList,
							Description: "Core rule set.",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"inbound_anomaly_score": {
										Type:         schema.TypeInt,
										Description:  "Inbound anomaly score of the rule set.",
										Optional:     true,
										ValidateFunc: validation.IntBetween(2, 10000),
									},

									"paranoia_level": {
										Type:        schema.TypeInt,
										Description: "Paranoia level of the rule set.",
										Optional:    true,
									},

									"rule_set": {
										Type:        schema.TypeList,
										Description: "Rule set.",
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Description: "Name of the rule set.",
													Optional:    true,
												},

												"type": {
													Type:         schema.TypeString,
													Description:  "Type of the rule set.",
													Optional:     true,
													ValidateFunc: validateParsableValue(parseWafRuleSetXRuleSetType),
												},

												"version": {
													Type:        schema.TypeString,
													Description: "Version of the rule set.",
													Required:    true,
												},

												"id": {
													Type:        schema.TypeString,
													Description: "ID of the rule set.",
													Optional:    true,
												},
											},
										},
										Required: true,
									},
								},
							},
							Optional: true,
						},

						"is_enabled": {
							Type:        schema.TypeBool,
							Description: "Determines is it rule set enabled or not.",
							Optional:    true,
						},

						"ml_rule_set": {
							Type:        schema.TypeList,
							Description: "List of ML rule sets.",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"rule_group": {
										Type:        schema.TypeList,
										Description: "List of rule groups.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"action": {
													Type:         schema.TypeString,
													Description:  "Action of the rule group.",
													Optional:     true,
													ValidateFunc: validateParsableValue(parseWafWafProfileXWafProfileRuleSetXRuleGroupXAction),
												},

												"id": {
													Type:        schema.TypeString,
													Description: "ID of the rule group.",
													Optional:    true,
												},

												"inbound_anomaly_score": {
													Type:         schema.TypeInt,
													Description:  "Inbound anomaly score.",
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 10000),
												},

												"is_enabled": {
													Type:        schema.TypeBool,
													Description: "Is the rule group enabled.",
													Optional:    true,
												},
											},
										},
										Optional: true,
									},

									"rule_set": {
										Type:        schema.TypeList,
										Description: "Rule set of the ML rule set.",
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Description: "Name of the rule set.",
													Optional:    true,
												},

												"type": {
													Type:         schema.TypeString,
													Description:  "Type of the rule set.",
													Optional:     true,
													ValidateFunc: validateParsableValue(parseWafRuleSetXRuleSetType),
												},

												"version": {
													Type:        schema.TypeString,
													Description: "Version of the rule set.",
													Required:    true,
												},

												"id": {
													Type:        schema.TypeString,
													Description: "ID of the rule set.",
													Optional:    true,
												},
											},
										},
										Required: true,
									},
								},
							},
							Optional: true,
						},

						"priority": {
							Type:         schema.TypeInt,
							Description:  "Priority of the rule set.",
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 999999),
						},

						"ya_rule_set": {
							Type:        schema.TypeList,
							Description: "Yandex rule set.",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"rule_group": {
										Type:        schema.TypeList,
										Description: "List of rule groups.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"action": {
													Type:         schema.TypeString,
													Description:  "Action of the rule group.",
													Optional:     true,
													ValidateFunc: validateParsableValue(parseWafWafProfileXWafProfileRuleSetXRuleGroupXAction),
												},

												"id": {
													Type:        schema.TypeString,
													Description: "ID of the rule group.",
													Optional:    true,
												},

												"inbound_anomaly_score": {
													Type:         schema.TypeInt,
													Description:  "Inbound anomaly score.",
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 10000),
												},

												"is_enabled": {
													Type:        schema.TypeBool,
													Description: "Is the rule group enabled.",
													Optional:    true,
												},
											},
										},
										Optional: true,
									},

									"rule_set": {
										Type:        schema.TypeList,
										Description: "Rule set of the Yandex rule set.",
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Description: "Name of the rule set.",
													Optional:    true,
												},

												"type": {
													Type:         schema.TypeString,
													Description:  "Type of the rule set.",
													Optional:     true,
													ValidateFunc: validateParsableValue(parseWafRuleSetXRuleSetType),
												},

												"version": {
													Type:        schema.TypeString,
													Description: "Version of the rule set.",
													Required:    true,
												},

												"id": {
													Type:        schema.TypeString,
													Description: "ID of the rule set.",
													Optional:    true,
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

	ruleSets, err := expandWafProfileRuleSetsSlice(d)
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
		RuleSets:           ruleSets,
		MatchAllRuleSets:   d.Get("match_all_rule_sets").(bool),
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

	ruleSet, err := flattenWafRuleSetSlice(resp.GetRuleSets())
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
	if err := d.Set("match_all_rule_sets", resp.GetMatchAllRuleSets()); err != nil {
		log.Printf("[ERROR] failed set field match_all_rule_sets: %s", err)
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
	if err := d.Set("rule_set", ruleSet); err != nil {
		log.Printf("[ERROR] failed set field rule_set: %s", err)
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

	ruleSets, err := expandWafProfileRuleSetsSlice_(d)
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
		RuleSets:           ruleSets,
		MatchAllRuleSets:   d.Get("match_all_rule_sets").(bool),
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
	"core_rule_set.0.rule_set.0.type":          "core_rule_set.rule_set.type",
	"analyze_request_body.0.is_enabled":        "analyze_request_body.is_enabled",
	"analyze_request_body.0.size_limit":        "analyze_request_body.size_limit",
	"analyze_request_body.0.size_limit_action": "analyze_request_body.size_limit_action",
	"rule_set":                                 "rule_sets",
	"match_all_rule_sets":                      "match_all_rule_sets",
}
