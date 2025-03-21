package yandex

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	smartwebsecurity "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dataSourceYandexSmartwebsecuritySecurityProfile() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about SecurityProfile. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/profiles).\n\nThis data source is used to define SecurityProfile that can be used by other resources.\n\n~> One of `security_profile_id` or `name` should be specified.\n",

		ReadContext: dataSourceYandexSmartwebsecuritySecurityProfileRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"advanced_rate_limiter_profile_id": {
				Type:        schema.TypeString,
				Description: resourceYandexSmartwebsecuritySecurityProfile().Schema["advanced_rate_limiter_profile_id"].Description,
				Computed:    true,
			},

			"captcha_id": {
				Type:        schema.TypeString,
				Description: resourceYandexSmartwebsecuritySecurityProfile().Schema["captcha_id"].Description,
				Computed:    true,
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
				Type:        schema.TypeString,
				Description: resourceYandexSmartwebsecuritySecurityProfile().Schema["default_action"].Description,
				Computed:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([-_0-9a-z]*)$"), ""), validation.StringLenBetween(0, 63)),
				},
				Set:      schema.HashString,
				Computed: true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},

			"security_profile_id": {
				Type:        schema.TypeString,
				Description: "ID of the security profile.",
				Optional:    true,
			},

			"security_rule": {
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

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"priority": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"rule_condition": {
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
								},
							},
							Computed: true,
						},

						"smart_protection": {
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

									"mode": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
							Computed: true,
						},

						"waf": {
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

									"mode": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"waf_profile_id": {
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
	}
}

func dataSourceYandexSmartwebsecuritySecurityProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	err := checkOneOf(d, "security_profile_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	resourceID := d.Get("security_profile_id").(string)
	_, nameOk := d.GetOk("name")
	if nameOk {
		resourceID, err = resolveObjectID(config.Context(), config, d, sdkresolvers.SecurityProfileResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source SecurityProfile by name: %v", err)
		}
	}

	req := &smartwebsecurity.GetSecurityProfileRequest{
		SecurityProfileId: resourceID,
	}

	md := new(metadata.MD)

	resp, err := config.sdk.SmartWebSecurity().SecurityProfile().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read SecurityProfile x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read SecurityProfile x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("security_profile %q", d.Get("security_profile_id").(string))))
	}

	d.SetId(resp.Id)

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
	if err := d.Set("security_profile_id", resp.GetId()); err != nil {
		log.Printf("[ERROR] failed set field security_profile_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("security_rule", securityRule); err != nil {
		log.Printf("[ERROR] failed set field security_rule: %s", err)
		return diag.FromErr(err)
	}

	return nil
}
