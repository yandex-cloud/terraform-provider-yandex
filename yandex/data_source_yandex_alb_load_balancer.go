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

			"allow_zonal_shift": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: resourceYandexALBLoadBalancer().Schema["allow_zonal_shift"].Description,
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
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Unique set of locations.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Unique set of locations.",
									},
									"subnet_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "ID of the subnet that location is located at.",
									},
									"disable_traffic": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "If set, will disable all L7 instances in the zone for request handling.",
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
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Set to `true` to disable Cloud Logging for the balancer.",
						},

						"discard_rule": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of rules to discard a fraction of logs.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"discard_percent": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The percent of logs which will be discarded.",
									},

									"grpc_codes": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Computed:    true,
										Description: "list of grpc codes by name",
									},

									"http_code_intervals": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Computed:    true,
										Description: "List of http code intervals *1XX*-*5XX* or *ALL*",
									},

									"http_codes": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
										Computed:    true,
										Description: "List of http codes *100*-*599*.",
									},
								},
							},
						},

						"log_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Cloud Logging group ID to send logs to.",
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
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the listener.",
						},
						"endpoint": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Network endpoint (addresses and ports) of the listener.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ports": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "One or more ports to listen on.",
										Elem:        &schema.Schema{Type: schema.TypeInt},
									},
									"address": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "One or more addresses to listen on.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"external_ipv4_address": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "External IPv4 address.",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Provided by the client or computed automatically.",
															},
														},
													},
												},
												"internal_ipv4_address": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "Internal IPv4 address.",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Provided by the client or computed automatically.",
															},
															"subnet_id": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "ID of the subnet that the address belongs to.",
															},
														},
													},
												},
												"external_ipv6_address": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "External IPv6 address.",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Provided by the client or computed automatically.",
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
							Type:        schema.TypeList,
							Optional:    true,
							Description: "HTTP handler that sets plain text HTTP router.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"handler": dataSourceHTTPHandler(),
									"redirects": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "Shortcut for adding http -> https redirects.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_to_https": {
													Type:        schema.TypeBool,
													Computed:    true,
													Description: "Redirects all unencrypted HTTP requests to the same URI with scheme changed to `https`.",
												},
											},
										},
									},
								},
							},
						},
						"stream": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Stream configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"handler": dataSourceStreamHandler(),
								},
							},
						},
						"tls": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "TLS configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_handler": dataSourceTLSHandler(),
									"sni_handler": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Settings for handling requests with Server Name Indication (SNI)",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Name of the SNI handler",
												},
												"server_names": {
													Type:        schema.TypeSet,
													Computed:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
													Description: "Server names that are matched by the SNI handler",
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
		Type:        schema.TypeList,
		Computed:    true,
		Description: "TLS handler resource.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_handler":   dataSourceHTTPHandler(),
				"stream_handler": dataSourceStreamHandler(),
				"certificate_ids": {
					Type:        schema.TypeSet,
					Computed:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Set:         schema.HashString,
					Description: "Certificate IDs in the Certificate Manager",
				},
			},
		},
	}
}

func dataSourceHTTPHandler() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "HTTP handler.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_router_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "HTTP router id.",
				},
				"http2_options": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "If set, will enable HTTP2 protocol for the handler.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_concurrent_streams": {
								Type:        schema.TypeInt,
								Computed:    true,
								Description: "Maximum number of concurrent streams.",
							},
						},
					},
				},
				"allow_http10": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "If set, will enable only HTTP1 protocol with HTTP1.0 support.",
				},
				"rewrite_request_id": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "When unset, will preserve the incoming `x-request-id` header, otherwise would rewrite it with a new value.",
				},
			},
		},
	}
}

func dataSourceStreamHandler() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Stream handler resource.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"backend_group_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Backend Group ID.",
				},
				"idle_timeout": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The idle timeout is the interval after which the connection will be forcibly closed if no data has been transmitted or received on either the upstream or downstream connection. If not configured, the default idle timeout is 1 hour. Setting it to 0 disables the timeout.",
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
	d.Set("allow_zonal_shift", alb.AllowZonalShift)

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
