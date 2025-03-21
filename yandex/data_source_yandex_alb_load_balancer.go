package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexALBLoadBalancer() *schema.Resource {
	return &schema.Resource{
		//Description: resourceYandexALBLoadBalancer().Description,
		Description: "Get information about a Yandex Application Load Balancer. For more information, see [Yandex Cloud Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/quickstart).\n\nThis data source is used to define [Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/concepts/application-load-balancer) that can be used by other resources.\n\n~> One of `load_balancer_id` or `name` should be specified.\n",
		Read:        dataSourceYandexALBLoadBalancerRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough, // TODO: SA1019: schema.ImportStatePassthrough is deprecated: Please use the context aware ImportStatePassthroughContext instead (staticcheck)
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexALBLoadBalancerDefaultTimeout),
			Update: schema.DefaultTimeout(yandexALBLoadBalancerDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexALBLoadBalancerDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"load_balancer_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: common.ResourceDescriptions["id"],
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: common.ResourceDescriptions["name"],
			},

			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: common.ResourceDescriptions["description"],
			},

			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: common.ResourceDescriptions["folder_id"],
			},

			"labels": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: common.ResourceDescriptions["labels"],
			},

			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: resourceYandexALBLoadBalancer().Schema["region_id"].Description,
			},

			"network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: resourceYandexALBLoadBalancer().Schema["network_id"].Description,
			},

			"log_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: resourceYandexALBLoadBalancer().Schema["log_group_id"].Description,
			},

			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: resourceYandexALBLoadBalancer().Schema["status"].Description,
			},

			"security_group_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: resourceYandexALBLoadBalancer().Schema["security_group_ids"].Description,
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: common.ResourceDescriptions["created_at"],
			},

			"allocation_policy": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: resourceYandexALBLoadBalancer().Schema["allocation_policy"].Description,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disable_traffic": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
							Set: resourceALBAllocationPolicyLocationHash,
						},
					},
				},
			},

			"log_options": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: resourceYandexALBLoadBalancer().Schema["log_options"].Description,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disable": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"discard_rule": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"discard_percent": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"grpc_codes": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Computed: true,
									},

									"http_code_intervals": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Computed: true,
									},

									"http_codes": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
										Computed: true,
									},
								},
							},
						},

						"log_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"listener": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: resourceYandexALBLoadBalancer().Schema["listener"].Description,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoint": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ports": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeInt},
									},
									"address": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"external_ipv4_address": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
												},
												"internal_ipv4_address": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:     schema.TypeString,
																Computed: true,
															},
															"subnet_id": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
												},
												"external_ipv6_address": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:     schema.TypeString,
																Computed: true,
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
						"http": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"handler": dataSourceHTTPHandler(),
									"redirects": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_to_https": {
													Type:     schema.TypeBool,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"stream": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"handler": dataSourceStreamHandler(),
								},
							},
						},
						"tls": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_handler": dataSourceTLSHandler(),
									"sni_handler": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"server_names": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
												},
												"handler": dataSourceTLSHandler(),
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
}

func dataSourceTLSHandler() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_handler":   dataSourceHTTPHandler(),
				"stream_handler": dataSourceStreamHandler(),
				"certificate_ids": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Set:      schema.HashString,
				},
			},
		},
	}
}

func dataSourceHTTPHandler() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_router_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"http2_options": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_concurrent_streams": {
								Type:     schema.TypeInt,
								Computed: true,
							},
						},
					},
				},
				"allow_http10": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"rewrite_request_id": {
					Type:     schema.TypeBool,
					Computed: true,
				},
			},
		},
	}
}

func dataSourceStreamHandler() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"backend_group_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"idle_timeout": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func dataSourceYandexALBLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "load_balancer_id", "name")
	if err != nil {
		return err
	}

	albID := d.Get("load_balancer_id").(string)
	_, albNameOk := d.GetOk("name")

	if albNameOk {
		albID, err = resolveObjectID(ctx, config, d, sdkresolvers.ApplicationLoadBalancerResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Load Balancerby name: %v", err)
		}
	}

	alb, err := config.sdk.ApplicationLoadBalancer().LoadBalancer().Get(ctx, &apploadbalancer.GetLoadBalancerRequest{
		LoadBalancerId: albID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("ALB Load Balancer %q", d.Get("name").(string)))
	}

	d.Set("load_balancer_id", alb.Id)
	d.Set("created_at", getTimestamp(alb.CreatedAt))
	d.Set("name", alb.Name)
	d.Set("folder_id", alb.FolderId)
	d.Set("description", alb.Description)
	d.Set("region_id", alb.RegionId)
	d.Set("network_id", alb.NetworkId)
	d.Set("log_group_id", alb.LogGroupId)
	d.Set("status", strings.ToLower(alb.Status.String()))

	allocationPolicy, err := flattenALBAllocationPolicy(alb)
	if err != nil {
		return err
	}
	if err := d.Set("allocation_policy", allocationPolicy); err != nil {
		return err
	}

	listeners, err := flattenALBListeners(alb)
	if err != nil {
		return err
	}
	if err := d.Set("listener", listeners); err != nil {
		return err
	}

	logOptions, err := flattenALBLogOptions(alb)
	if err != nil {
		return err
	}
	if err = d.Set("log_options", logOptions); err != nil {
		return err
	}

	if err := d.Set("labels", alb.Labels); err != nil {
		return err
	}

	d.SetId(alb.Id)

	return nil
}
