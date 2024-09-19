package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	smartcaptcha "github.com/yandex-cloud/go-genproto/yandex/cloud/smartcaptcha/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dataSourceYandexSmartcaptchaCaptcha() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexSmartcaptchaCaptchaRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"allowed_sites": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},

			"captcha_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"challenge_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"client_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cloud_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"complexity": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"override_variant": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"challenge_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"complexity": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"pre_check_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"uuid": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},

			"pre_check_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_rule": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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

									"host": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"hosts": {
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

									"uri": {
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
								},
							},
							Computed: true,
						},

						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"override_variant_uuid": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"priority": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Computed: true,
			},

			"style_json": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"suspend": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"turn_off_hostname_check": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexSmartcaptchaCaptchaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	err := checkOneOf(d, "captcha_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	resourceID := d.Get("captcha_id").(string)
	_, nameOk := d.GetOk("name")
	if nameOk {
		resourceID, err = resolveObjectID(config.Context(), config, d, sdkresolvers.CaptchaResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source Captcha by name: %v", err)
		}
	}

	req := &smartcaptcha.GetCaptchaRequest{
		CaptchaId: resourceID,
	}

	md := new(metadata.MD)

	resp, err := config.sdk.SmartCaptcha().Captcha().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read Captcha x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read Captcha x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("captcha %q", d.Get("captcha_id").(string))))
	}

	d.SetId(resp.Id)

	log.Printf("[DEBUG] Read Captcha response: %s", protoDump(resp))

	createdAt := getTimestamp(resp.GetCreatedAt())

	overrideVariant, err := flattenSmartcaptchaOverrideVariantSlice(resp.GetOverrideVariants())
	if err != nil { // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	securityRule, err := flattenSmartcaptchaSecurityRuleSlice(resp.GetSecurityRules())
	if err != nil { // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	if err := d.Set("allowed_sites", resp.GetAllowedSites()); err != nil {
		log.Printf("[ERROR] failed set field allowed_sites: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("captcha_id", resp.GetId()); err != nil {
		log.Printf("[ERROR] failed set field captcha_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("challenge_type", resp.GetChallengeType().String()); err != nil {
		log.Printf("[ERROR] failed set field challenge_type: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("client_key", resp.GetClientKey()); err != nil {
		log.Printf("[ERROR] failed set field client_key: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("cloud_id", resp.GetCloudId()); err != nil {
		log.Printf("[ERROR] failed set field cloud_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("complexity", resp.GetComplexity().String()); err != nil {
		log.Printf("[ERROR] failed set field complexity: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", createdAt); err != nil {
		log.Printf("[ERROR] failed set field created_at: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("deletion_protection", resp.GetDeletionProtection()); err != nil {
		log.Printf("[ERROR] failed set field deletion_protection: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("folder_id", resp.GetFolderId()); err != nil {
		log.Printf("[ERROR] failed set field folder_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("name", resp.GetName()); err != nil {
		log.Printf("[ERROR] failed set field name: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("override_variant", overrideVariant); err != nil {
		log.Printf("[ERROR] failed set field override_variant: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("pre_check_type", resp.GetPreCheckType().String()); err != nil {
		log.Printf("[ERROR] failed set field pre_check_type: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("security_rule", securityRule); err != nil {
		log.Printf("[ERROR] failed set field security_rule: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("style_json", resp.GetStyleJson()); err != nil {
		log.Printf("[ERROR] failed set field style_json: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("suspend", resp.GetSuspend()); err != nil {
		log.Printf("[ERROR] failed set field suspend: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("turn_off_hostname_check", resp.GetTurnOffHostnameCheck()); err != nil {
		log.Printf("[ERROR] failed set field turn_off_hostname_check: %s", err)
		return diag.FromErr(err)
	}

	return nil
}
