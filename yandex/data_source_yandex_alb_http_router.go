package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexALBHTTPRouter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexALBHTTPRouterRead,
		Schema: map[string]*schema.Schema{
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

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"route_options": dataSourceRouteOptions(),
		},
	}
}

func dataSourceRouteOptions() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"rbac": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"action": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"principals": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"and_principals": {
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"headers": {
														Type:     schema.TypeList,
														Computed: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:     schema.TypeString,
																	Computed: true,
																},
																"value": dataSourceStringMatch(),
															},
														},
													},
													"remote_ip": {
														Type:     schema.TypeString,
														Computed: true,
													},
													"any": {
														Type:     schema.TypeBool,
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
				"security_profile_id": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func dataSourceYandexALBHTTPRouterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "http_router_id", "name")
	if err != nil {
		return err
	}

	routerID := d.Get("http_router_id").(string)
	_, routerNameOk := d.GetOk("name")

	if routerNameOk {
		routerID, err = resolveObjectID(ctx, config, d, sdkresolvers.ALBHTTPRouterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Http Router by name: %v", err)
		}
	}

	router, err := config.sdk.ApplicationLoadBalancer().HttpRouter().Get(ctx, &apploadbalancer.GetHttpRouterRequest{
		HttpRouterId: routerID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Http Router with ID %q", routerID))
	}

	ro, err := flattenALBRouteOptions(router.GetRouteOptions())
	if err != nil {
		return err
	}

	d.Set("http_router_id", router.Id)
	d.Set("name", router.Name)
	d.Set("description", router.Description)
	d.Set("created_at", getTimestamp(router.CreatedAt))
	d.Set("folder_id", router.FolderId)
	d.Set("route_options", ro)

	if err := d.Set("labels", router.Labels); err != nil {
		return err
	}

	d.SetId(router.Id)

	return nil
}
