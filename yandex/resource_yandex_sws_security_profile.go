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
	smartwebsecurity "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func resourceYandexSmartwebsecuritySecurityProfile() *schema.Resource {
	return &schema.Resource{
		Description: "With security profiles you can protect your infrastructure from DDoS attacks at the application level (L7).\n\nCreates a Security Profile in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/profiles).\n",

		CreateContext: resourceYandexSmartwebsecuritySecurityProfileCreate,
		ReadContext:   resourceYandexSmartwebsecuritySecurityProfileRead,
		UpdateContext: resourceYandexSmartwebsecuritySecurityProfileUpdate,
		DeleteContext: resourceYandexSmartwebsecuritySecurityProfileDelete,

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
			"advanced_rate_limiter_profile_id": {
				Type:        schema.TypeString,
				Description: "Advanced rate limiter profile ID to use with this security profile. Set empty to use default.",
				Optional:    true,
			},

			"captcha_id": {
				Type:        schema.TypeString,
				Description: "Captcha ID to use with this security profile. Set empty to use default.",
				Optional:    true,
			},

			"cloud_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["cloud_id"],
				Computed:    true,
				Optional:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"default_action": {
				Type:         schema.TypeString,
				Description:  "Action to perform if none of rules matched. Possible values: `ALLOW` or `DENY`.",
				Optional:     true,
				ValidateFunc: validateParsableValue(parseSmartwebsecuritySecurityProfileXDefaultAction),
			},

			"description": {
				Type:         schema.TypeString,
				Description:  common.ResourceDescriptions["description"],
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
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

			"name": {
				Type:         schema.TypeString,
				Description:  common.ResourceDescriptions["name"],
				Optional:     true,
				ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([a-zA-Z0-9][a-zA-Z0-9-_.]*)$"), ""), validation.StringLenBetween(1, 50)),
			},

			"security_rule": {
				Type:        schema.TypeList,
				Description: "List of security rules.\n\n~> Exactly one rule specifier: `smart_protection` or `rule_condition` or `waf` should be specified.\n",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:         schema.TypeString,
							Description:  "Optional description of the rule. 0-512 characters long.",
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 512),
						},

						"dry_run": {
							Type:        schema.TypeBool,
							Description: "This mode allows you to test your security profile or a single rule.",
							Optional:    true,
						},

						"name": {
							Type:         schema.TypeString,
							Description:  "Name of the rule. The name is unique within the security profile. 1-50 characters long.",
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([a-zA-Z0-9][a-zA-Z0-9-_.]*)$"), ""), validation.StringLenBetween(1, 50)),
						},

						"priority": {
							Type:         schema.TypeInt,
							Description:  "Determines the priority for checking the incoming traffic.",
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 999999),
						},

						"rule_condition": {
							Type:        schema.TypeList,
							Description: "Rule actions, see [Rule actions](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#rule-action).",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:         schema.TypeString,
										Description:  "Action to perform if this rule matched. Possible values: `ALLOW` or `DENY`.",
										Optional:     true,
										ValidateFunc: validateParsableValue(parseSmartwebsecuritySecurityRuleXRuleConditionXAction),
									},

									"condition": {
										Type:        schema.TypeList,
										Description: "The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).",
										MaxItems:    1,
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
								},
							},
							Optional: true,
						},

						"smart_protection": {
							Type:        schema.TypeList,
							Description: "Smart Protection rule, see [Smart Protection rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#smart-protection-rules).",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"condition": {
										Type:        schema.TypeList,
										Description: "The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).",
										MaxItems:    1,
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

									"mode": {
										Type:         schema.TypeString,
										Description:  "Mode of protection. Possible values: `FULL` (full protection means that the traffic will be checked based on ML models and behavioral analysis, with suspicious requests being sent to SmartCaptcha) or `API` (API protection means checking the traffic based on ML models and behavioral analysis without sending suspicious requests to SmartCaptcha. The suspicious requests will be blocked).",
										Optional:     true,
										ValidateFunc: validateParsableValue(parseSmartwebsecuritySecurityRuleXSmartProtectionXMode),
									},
								},
							},
							Optional: true,
						},

						"waf": {
							Type:        schema.TypeList,
							Description: "Web Application Firewall (WAF) rule, see [WAF rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#waf-rules).",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"condition": {
										Type:        schema.TypeList,
										Description: "The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).",
										MaxItems:    1,
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

									"mode": {
										Type:         schema.TypeString,
										Description:  "Mode of protection. Possible values: `FULL` (full protection means that the traffic will be checked based on ML models and behavioral analysis, with suspicious requests being sent to SmartCaptcha) or `API` (API protection means checking the traffic based on ML models and behavioral analysis without sending suspicious requests to SmartCaptcha. The suspicious requests will be blocked).",
										Optional:     true,
										ValidateFunc: validateParsableValue(parseSmartwebsecuritySecurityRuleXWafXMode),
									},

									"waf_profile_id": {
										Type:        schema.TypeString,
										Description: "ID of WAF profile to use in this rule.",
										Required:    true,
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

func resourceYandexSmartwebsecuritySecurityProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	folderId, err := getFolderID(d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	labels := expandStringStringMap(d.Get("labels").(map[string]interface{}))
	defaultAction, err := parseSmartwebsecuritySecurityProfileXDefaultAction(d.Get("default_action").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	securityRules, err := expandSecurityProfileSecurityRulesSlice(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &smartwebsecurity.CreateSecurityProfileRequest{
		FolderId:                     folderId,
		Labels:                       labels,
		Name:                         d.Get("name").(string),
		Description:                  d.Get("description").(string),
		DefaultAction:                defaultAction,
		SecurityRules:                securityRules,
		CaptchaId:                    d.Get("captcha_id").(string),
		AdvancedRateLimiterProfileId: d.Get("advanced_rate_limiter_profile_id").(string),
	}

	log.Printf("[DEBUG] Create SecurityProfile request: %s", protoDump(req))

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.SmartWebSecurity().SecurityProfile().Create(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create SecurityProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create SecurityProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while get smartwebsecurity.SecurityProfile create operation metadata: %v", err)
	}

	createMetadata, ok := protoMetadata.(*smartwebsecurity.CreateSecurityProfileMetadata)
	if !ok {
		return diag.Errorf("could not get SecurityProfile ID from create operation metadata")
	}

	d.SetId(createMetadata.SecurityProfileId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexSmartwebsecuritySecurityProfileRead(ctx, d, meta)
}

func resourceYandexSmartwebsecuritySecurityProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &smartwebsecurity.GetSecurityProfileRequest{
		SecurityProfileId: d.Id(),
	}

	log.Printf("[DEBUG] Read SecurityProfile request: %s", protoDump(req))

	md := new(metadata.MD)
	resp, err := config.sdk.SmartWebSecurity().SecurityProfile().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read SecurityProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read SecurityProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("security_profile %q", d.Id())))
	}

	log.Printf("[DEBUG] Read SecurityProfile response: %s", protoDump(resp))

	createdAt := getTimestamp(resp.GetCreatedAt())

	securityRule, err := flattenSmartwebsecuritySecurityRuleSlice(resp.GetSecurityRules())
	if err != nil { // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	if err := d.Set("advanced_rate_limiter_profile_id", resp.GetAdvancedRateLimiterProfileId()); err != nil {
		log.Printf("[ERROR] failed set field advanced_rate_limiter_profile_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("captcha_id", resp.GetCaptchaId()); err != nil {
		log.Printf("[ERROR] failed set field captcha_id: %s", err)
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
	if err := d.Set("default_action", resp.GetDefaultAction().String()); err != nil {
		log.Printf("[ERROR] failed set field default_action: %s", err)
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
	if err := d.Set("security_rule", securityRule); err != nil {
		log.Printf("[ERROR] failed set field security_rule: %s", err)
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexSmartwebsecuritySecurityProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	labels := expandStringStringMap(d.Get("labels").(map[string]interface{}))
	defaultAction, err := parseSmartwebsecuritySecurityProfileXDefaultAction(d.Get("default_action").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	securityRules, err := expandSecurityProfileSecurityRulesSlice_(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &smartwebsecurity.UpdateSecurityProfileRequest{
		SecurityProfileId:            d.Id(),
		Labels:                       labels,
		Name:                         d.Get("name").(string),
		Description:                  d.Get("description").(string),
		DefaultAction:                defaultAction,
		SecurityRules:                securityRules,
		CaptchaId:                    d.Get("captcha_id").(string),
		AdvancedRateLimiterProfileId: d.Get("advanced_rate_limiter_profile_id").(string),
	}

	updatePath := generateFieldMasks(d, resourceYandexSmartwebsecuritySecurityProfileUpdateFieldsMap)
	req.UpdateMask = &fieldmaskpb.FieldMask{Paths: updatePath}

	log.Printf("[DEBUG] Update SecurityProfile request: %s", protoDump(req))

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.SmartWebSecurity().SecurityProfile().Update(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update SecurityProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update SecurityProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexSmartwebsecuritySecurityProfileRead(ctx, d, meta)
}

func resourceYandexSmartwebsecuritySecurityProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &smartwebsecurity.DeleteSecurityProfileRequest{
		SecurityProfileId: d.Id(),
	}

	log.Printf("[DEBUG] Delete SecurityProfile request: %s", protoDump(req))

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.SmartWebSecurity().SecurityProfile().Delete(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete SecurityProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete SecurityProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("security_profile %q", d.Id())))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

var resourceYandexSmartwebsecuritySecurityProfileUpdateFieldsMap = map[string]string{
	"labels":                           "labels",
	"name":                             "name",
	"description":                      "description",
	"default_action":                   "default_action",
	"security_rule":                    "security_rules",
	"captcha_id":                       "captcha_id",
	"advanced_rate_limiter_profile_id": "advanced_rate_limiter_profile_id",
}
