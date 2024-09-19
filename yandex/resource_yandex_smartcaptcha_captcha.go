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
	smartcaptcha "github.com/yandex-cloud/go-genproto/yandex/cloud/smartcaptcha/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func resourceYandexSmartcaptchaCaptcha() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexSmartcaptchaCaptchaCreate,
		ReadContext:   resourceYandexSmartcaptchaCaptchaRead,
		UpdateContext: resourceYandexSmartcaptchaCaptchaUpdate,
		DeleteContext: resourceYandexSmartcaptchaCaptchaDelete,

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
			"allowed_sites": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},

			"challenge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateParsableValue(parseSmartcaptchaCaptchaChallengeType),
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateParsableValue(parseSmartcaptchaCaptchaComplexity),
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"folder_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^(|[a-z]([-a-z0-9]{0,61}[a-z0-9])?)$"), ""),
			},

			"override_variant": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"challenge_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateParsableValue(parseSmartcaptchaCaptchaChallengeType),
						},

						"complexity": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateParsableValue(parseSmartcaptchaCaptchaComplexity),
						},

						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 512),
						},

						"pre_check_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateParsableValue(parseSmartcaptchaCaptchaPreCheckType),
						},

						"uuid": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([a-zA-Z0-9][a-zA-Z0-9-_.]*)$"), ""), validation.StringLenBetween(0, 64)),
						},
					},
				},
				Optional: true,
			},

			"pre_check_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateParsableValue(parseSmartcaptchaCaptchaPreCheckType),
			},

			"security_rule": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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

									"host": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"hosts": {
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

									"uri": {
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
								},
							},
							Optional: true,
						},

						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 512),
						},

						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([a-zA-Z0-9][a-zA-Z0-9-_.]*)$"), ""), validation.StringLenBetween(1, 50)),
						},

						"override_variant_uuid": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"priority": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 999999),
						},
					},
				},
				Optional: true,
			},

			"style_json": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"suspend": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"turn_off_hostname_check": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceYandexSmartcaptchaCaptchaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	folderId, err := getFolderID(d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	allowedSites := expandStringSlice(d.Get("allowed_sites").([]interface{}))
	captchaComplexity, err := parseSmartcaptchaCaptchaComplexity(d.Get("complexity").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	captchaPreCheckType, err := parseSmartcaptchaCaptchaPreCheckType(d.Get("pre_check_type").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	captchaChallengeType, err := parseSmartcaptchaCaptchaChallengeType(d.Get("challenge_type").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	securityRules, err := expandCaptchaSecurityRulesSlice(d)
	if err != nil {
		return diag.FromErr(err)
	}

	overrideVariants, err := expandCaptchaOverrideVariantsSlice(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &smartcaptcha.CreateCaptchaRequest{
		FolderId:             folderId,
		Name:                 d.Get("name").(string),
		AllowedSites:         allowedSites,
		Complexity:           captchaComplexity,
		StyleJson:            d.Get("style_json").(string),
		TurnOffHostnameCheck: d.Get("turn_off_hostname_check").(bool),
		PreCheckType:         captchaPreCheckType,
		ChallengeType:        captchaChallengeType,
		SecurityRules:        securityRules,
		DeletionProtection:   d.Get("deletion_protection").(bool),
		OverrideVariants:     overrideVariants,
	}

	log.Printf("[DEBUG] Create Captcha request: %s", protoDump(req))

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.SmartCaptcha().Captcha().Create(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create Captcha x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Create Captcha x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while get smartcaptcha.Captcha create operation metadata: %v", err)
	}

	createMetadata, ok := protoMetadata.(*smartcaptcha.CreateCaptchaMetadata)
	if !ok {
		return diag.Errorf("could not get Captcha ID from create operation metadata")
	}

	d.SetId(createMetadata.CaptchaId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexSmartcaptchaCaptchaRead(ctx, d, meta)
}

func resourceYandexSmartcaptchaCaptchaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &smartcaptcha.GetCaptchaRequest{
		CaptchaId: d.Id(),
	}

	log.Printf("[DEBUG] Read Captcha request: %s", protoDump(req))

	md := new(metadata.MD)
	resp, err := config.sdk.SmartCaptcha().Captcha().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read Captcha x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read Captcha x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("captcha %q", d.Id())))
	}

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

func resourceYandexSmartcaptchaCaptchaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	allowedSites := expandStringSlice(d.Get("allowed_sites").([]interface{}))
	captchaComplexity, err := parseSmartcaptchaCaptchaComplexity(d.Get("complexity").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	captchaPreCheckType, err := parseSmartcaptchaCaptchaPreCheckType(d.Get("pre_check_type").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	captchaChallengeType, err := parseSmartcaptchaCaptchaChallengeType(d.Get("challenge_type").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	securityRules, err := expandCaptchaSecurityRulesSlice_(d)
	if err != nil {
		return diag.FromErr(err)
	}

	overrideVariants, err := expandCaptchaOverrideVariantsSlice_(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &smartcaptcha.UpdateCaptchaRequest{
		CaptchaId:            d.Id(),
		Name:                 d.Get("name").(string),
		AllowedSites:         allowedSites,
		Complexity:           captchaComplexity,
		StyleJson:            d.Get("style_json").(string),
		TurnOffHostnameCheck: d.Get("turn_off_hostname_check").(bool),
		PreCheckType:         captchaPreCheckType,
		ChallengeType:        captchaChallengeType,
		SecurityRules:        securityRules,
		DeletionProtection:   d.Get("deletion_protection").(bool),
		OverrideVariants:     overrideVariants,
	}

	updatePath := generateFieldMasks(d, resourceYandexSmartcaptchaCaptchaUpdateFieldsMap)
	req.UpdateMask = &fieldmaskpb.FieldMask{Paths: updatePath}

	log.Printf("[DEBUG] Update Captcha request: %s", protoDump(req))

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.SmartCaptcha().Captcha().Update(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update Captcha x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Update Captcha x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexSmartcaptchaCaptchaRead(ctx, d, meta)
}

func resourceYandexSmartcaptchaCaptchaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &smartcaptcha.DeleteCaptchaRequest{
		CaptchaId: d.Id(),
	}

	log.Printf("[DEBUG] Delete Captcha request: %s", protoDump(req))

	md := new(metadata.MD)
	op, err := config.sdk.WrapOperation(config.sdk.SmartCaptcha().Captcha().Delete(ctx, req, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete Captcha x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Delete Captcha x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("captcha %q", d.Id())))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

var resourceYandexSmartcaptchaCaptchaUpdateFieldsMap = map[string]string{
	"name":                    "name",
	"allowed_sites":           "allowed_sites",
	"complexity":              "complexity",
	"style_json":              "style_json",
	"turn_off_hostname_check": "turn_off_hostname_check",
	"pre_check_type":          "pre_check_type",
	"challenge_type":          "challenge_type",
	"security_rule":           "security_rules",
	"deletion_protection":     "deletion_protection",
	"override_variant":        "override_variants",
}
