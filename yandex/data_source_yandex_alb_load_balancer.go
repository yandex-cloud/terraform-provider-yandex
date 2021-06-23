package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"strings"
)

func dataSourceYandexALBLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexALBLoadBalancerRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexALBLoadBalancerDefaultTimeout),
			Update: schema.DefaultTimeout(yandexALBLoadBalancerDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexALBLoadBalancerDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"load_balancer_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"region_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"log_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"allocation_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
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

			"listener": {
				Type:     schema.TypeList,
				Computed: true,
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
													MaxItems: 1,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
													ConflictsWith: []string{"listener.endpoint.address.internal_ipv4_address", "listener.endpoint.address.external_ipv6_address"},
												},
												"internal_ipv4_address": {
													Type:     schema.TypeList,
													MaxItems: 1,
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
													ConflictsWith: []string{"listener.endpoint.address.external_ipv4_address", "listener.endpoint.address.external_ipv6_address"},
												},
												"external_ipv6_address": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"address": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
													ConflictsWith: []string{"listener.endpoint.address.internal_ipv4_address", "listener.endpoint.address.external_ipv4_address"},
												},
											},
										},
									},
								},
							},
						},
						"http": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Optional:      true,
							ConflictsWith: []string{"listener.tls"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"handler": dataSourceHTTPHandler("listener.http.handler", []string{"listener.http.redirects"}),
									"redirects": {
										Type:          schema.TypeList,
										MaxItems:      1,
										Optional:      true,
										ConflictsWith: []string{"listener.http.handler"},
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
						"tls": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Optional:      true,
							ConflictsWith: []string{"listener.http"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_handler": dataSourceTLSHandler("listener.tls.default_handler"),
									"sni_handler": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"service_name": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
												},
												"handler": dataSourceTLSHandler("listener.tls.sni_handler.handler"),
											},
										},
									},
								},
							},
						},
					},
				},
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTLSHandler(path string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_handler": dataSourceHTTPHandler(path+".http_handler", nil),
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

func dataSourceHTTPHandler(path string, conflictWith []string) *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		MaxItems:      1,
		Optional:      true,
		ConflictsWith: conflictWith,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_router_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"http2_options": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_concurrent_streams": {
								Type:     schema.TypeInt,
								Computed: true,
							},
						},
					},
					ConflictsWith: []string{path + ".allow_http10"},
				},
				"allow_http10": {
					Type:          schema.TypeBool,
					Optional:      true,
					ConflictsWith: []string{path + ".http2_options"},
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

	createdAt, err := getTimestamp(alb.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("load_balancer_id", alb.Id)
	d.Set("created_at", createdAt)
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

	if err := d.Set("labels", alb.Labels); err != nil {
		return err
	}

	d.SetId(alb.Id)

	return nil
}
