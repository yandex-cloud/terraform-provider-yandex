package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"strings"
)

func dataSourceYandexALBVirtualHost() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexALBVirtualHostRead,

		Schema: map[string]*schema.Schema{
			"virtual_host_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"http_router_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"authority": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"modify_request_headers":  dataSourceHeaderModification("modify_request_headers."),
			"modify_response_headers": dataSourceHeaderModification("modify_response_headers."),
			"route": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type: schema.TypeString,

							Computed: true,
						},
						"http_route": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Computed:      true,
							ConflictsWith: []string{"route.grpc_route"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_route_action": {
										Type:          schema.TypeList,
										MaxItems:      1,
										Computed:      true,
										ConflictsWith: []string{"route.http_route.redirect_action", "route.http_route.direct_response_action"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"timeout": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"idle_timeout": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"prefix_rewrite": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"upgrade_types": {
													Type: schema.TypeSet,

													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
												},
												"host_rewrite": {
													Type:          schema.TypeString,
													Computed:      true,
													ConflictsWith: []string{"route.http_route.http_route_action.auto_host_rewrite"},
												},
												"auto_host_rewrite": {
													Type:          schema.TypeBool,
													Computed:      true,
													ConflictsWith: []string{"route.http_route.http_route_action.host_rewrite"},
												},
											},
										},
									},
									"redirect_action": {
										Type:          schema.TypeList,
										MaxItems:      1,
										Computed:      true,
										ConflictsWith: []string{"route.http_route.http_route_action", "route.http_route.direct_response_action"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"replace_scheme": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"replace_host": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"replace_port": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"remove_query": {
													Type:     schema.TypeBool,
													Computed: true,
												},
												"response_code": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"replace_path": {
													Type:          schema.TypeString,
													Computed:      true,
													ConflictsWith: []string{"route.http_route.redirect_action.replace_prefix"},
												},
												"replace_prefix": {
													Type:          schema.TypeString,
													Computed:      true,
													ConflictsWith: []string{"route.http_route.redirect_action.replace_path"},
												},
											},
										},
									},
									"direct_response_action": {
										Type:          schema.TypeList,
										MaxItems:      1,
										Computed:      true,
										ConflictsWith: []string{"route.http_route.redirect_action", "route.http_route.http_route_action"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"body": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"http_match": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_method": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem:     schema.TypeString,
													Set:      schema.HashString,
												},
												"path": dataSourceStringMatch("route.http_route.http_match.path."),
											},
										},
									},
								},
							},
						},
						"grpc_route": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Computed:      true,
							ConflictsWith: []string{"route.http_route"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"grpc_match": {
										Type: schema.TypeList,

										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqmn": dataSourceStringMatch("route.grpc_route.grpc_match.fqmn."),
											},
										},
									},
									"grpc_route_action": {
										Type:          schema.TypeList,
										MaxItems:      1,
										Computed:      true,
										ConflictsWith: []string{"route.grpc_route.grpc_status_response_action"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"max_timeout": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"idle_timeout": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"host_rewrite": {
													Type:          schema.TypeString,
													Computed:      true,
													ConflictsWith: []string{"route.grpc_route.grpc_route_action.auto_host_rewrite"},
												},
												"auto_host_rewrite": {
													Type:          schema.TypeBool,
													Computed:      true,
													ConflictsWith: []string{"route.grpc_route.grpc_route_action.host_rewrite"},
												},
											},
										},
									},
									"grpc_status_response_action": {
										Type:          schema.TypeList,
										MaxItems:      1,
										Computed:      true,
										ConflictsWith: []string{"route.grpc_route.grpc_route_action"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
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
		},
	}
}

func dataSourceHeaderModification(path string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"append": {
					Type:          schema.TypeString,
					Computed:      true,
					ConflictsWith: []string{path + "replace", path + "remove"},
				},
				"replace": {
					Type:          schema.TypeString,
					Computed:      true,
					ConflictsWith: []string{path + "append", path + "remove"},
				},
				"remove": {
					Type:          schema.TypeBool,
					Computed:      true,
					ConflictsWith: []string{path + "replace", path + "append"},
				},
			},
		},
		Set: resourceALBVirtualHostHeaderModificationHash,
	}
}

func dataSourceStringMatch(path string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"exact": {
					Type:          schema.TypeString,
					Computed:      true,
					ConflictsWith: []string{path + "prefix"},
				},
				"prefix": {
					Type:          schema.TypeString,
					Computed:      true,
					ConflictsWith: []string{path + "exact"},
				},
			},
		},
	}
}

func retrieveDataFromVirtualHostID(id string) (string, string) {
	attrs := strings.Split(id, "/")
	return attrs[0], attrs[1]
}

func dataSourceYandexALBVirtualHostRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "virtual_host_id", "name")
	if err != nil {
		return err
	}

	err = checkOneOf(d, "virtual_host_id", "http_router_id")
	if err != nil {
		return err
	}

	virtualHostName := d.Get("name").(string)
	httpRouterID := d.Get("http_router_id").(string)
	virtualHostID, virtualHostIDOk := d.GetOk("virtual_host_id")

	if virtualHostIDOk {
		httpRouterID, virtualHostName = retrieveDataFromVirtualHostID(virtualHostID.(string))
	} else {
		virtualHostID = httpRouterID + "/" + virtualHostName
	}

	virtualHost, err := config.sdk.ApplicationLoadBalancer().VirtualHost().Get(ctx, &apploadbalancer.GetVirtualHostRequest{
		HttpRouterId:    httpRouterID,
		VirtualHostName: virtualHostName,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Virtual Host %q", virtualHostName))
	}

	requestHeaderModification, err := flattenALBHeaderModification(virtualHost.ModifyRequestHeaders)
	if err != nil {
		return err
	}

	responseHeaderModification, err := flattenALBHeaderModification(virtualHost.ModifyResponseHeaders)
	if err != nil {
		return err
	}

	routes, err := flattenALBRoutes(virtualHost.Routes)
	if err != nil {
		return err
	}

	d.Set("virtual_host_id", virtualHostID.(string))
	d.Set("name", virtualHost.Name)
	d.Set("authority", virtualHost.Authority)

	if err := d.Set("modify_request_headers", requestHeaderModification); err != nil {
		return err
	}

	if err := d.Set("modify_response_headers", responseHeaderModification); err != nil {
		return err
	}

	if err := d.Set("route", routes); err != nil {
		return err
	}

	d.SetId(virtualHostID.(string))

	return nil

}
