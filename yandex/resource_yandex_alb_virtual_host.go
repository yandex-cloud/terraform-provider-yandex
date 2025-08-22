package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const yandexALBVirtualHostDefaultTimeout = 5 * time.Minute

func resourceYandexALBVirtualHost() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a virtual host that belongs to specified HTTP router and adds the specified routes to it. For more information, see [the official documentation](https://yandex.cloud/docs/application-load-balancer/concepts/http-router).\n",
		Create:      resourceYandexALBVirtualHostCreate,
		Read:        resourceYandexALBVirtualHostRead,
		Update:      resourceYandexALBVirtualHostUpdate,
		Delete:      resourceYandexALBVirtualHostDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexALBVirtualHostDefaultTimeout),
			Update: schema.DefaultTimeout(yandexALBVirtualHostDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexALBVirtualHostDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"http_router_id": {
				Type:        schema.TypeString,
				Description: "The ID of the HTTP router to which the virtual host belongs.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},
			"authority": {
				Type:        schema.TypeSet,
				Description: "A list of domains (host/authority header) that will be matched to this virtual host. Wildcard hosts are supported in the form of '*.foo.com' or '*-bar.foo.com'. If not specified, all domains will be matched.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"modify_request_headers":  headerModification(),
			"modify_response_headers": headerModification(),
			rateLimitSchemaKey:        rateLimit(),
			"route": {
				Type:        schema.TypeList,
				Description: routeSchemaDescription,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: routeNameSchemaDescription,
							Optional:    true,
						},
						"http_route": {
							Type:        schema.TypeList,
							Description: routeHTTPRouteSchemaDescription,
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_route_action": {
										Type:        schema.TypeList,
										Description: routeHTTPRouteActionSchemaDescription,
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:        schema.TypeString,
													Description: routeHTTPRouteActionBackendGroupIDSchemaDescription,
													Required:    true,
												},
												"timeout": {
													Type:             schema.TypeString,
													Description:      routeHTTPRouteActionTimeoutSchemaDescription,
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"idle_timeout": {
													Type:             schema.TypeString,
													Description:      routeHTTPRouteActionIdleTimeoutSchemaDescription,
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"prefix_rewrite": {
													Type:        schema.TypeString,
													Description: routeHTTPRouteActionPrefixRewriteSchemaDescription,
													Optional:    true,
												},
												regexRewriteSchemaKey: regexRewrite(),
												"upgrade_types": {
													Type:        schema.TypeSet,
													Description: routeHTTPRouteActionUpgradeTypesSchemaDescription,
													Optional:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
												},
												"host_rewrite": {
													Type:        schema.TypeString,
													Description: routeHTTPRouteActionHostRewriteSchemaDescription,
													Optional:    true,
												},
												"auto_host_rewrite": {
													Type:        schema.TypeBool,
													Description: routeHTTPRouteActionAutoHostRewriteSchemaDescription,
													Optional:    true,
												},
												rateLimitSchemaKey: rateLimit(),
											},
										},
									},
									"redirect_action": {
										Type:        schema.TypeList,
										Description: routeHTTPRedirectActionSchemaDescription,
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"replace_scheme": {
													Type:        schema.TypeString,
													Description: routeHTTPRedirectActionReplaceSchemeSchemaDescription,
													Optional:    true,
												},
												"replace_host": {
													Type:        schema.TypeString,
													Description: routeHTTPRedirectActionReplaceHostSchemaDescription,
													Optional:    true,
												},
												"replace_port": {
													Type:        schema.TypeInt,
													Description: routeHTTPRedirectActionReplacePortSchemaDescription,
													Optional:    true,
												},
												"remove_query": {
													Type:        schema.TypeBool,
													Description: routeHTTPRedirectActionRemoveQuerySchemaDescription,
													Optional:    true,
												},
												"response_code": {
													Type:             schema.TypeString,
													Description:      routeHTTPRedirectActionResponseCodeSchemaDescription,
													Default:          "moved_permanently",
													Optional:         true,
													DiffSuppressFunc: CaseInsensitive,
												},
												"replace_path": {
													Type:        schema.TypeString,
													Description: routeHTTPRedirectActionReplacePathSchemaDescription,
													Optional:    true,
												},
												"replace_prefix": {
													Type:        schema.TypeString,
													Description: routeHTTPRedirectActionReplacePrefixSchemaDescription,
													Optional:    true,
												},
											},
										},
									},
									"direct_response_action": {
										Type:        schema.TypeList,
										Description: routeHTTPDirectResponseActionSchemaDescription,
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:         schema.TypeInt,
													Description:  routeHTTPDirectResponseActionStatusSchemaDescription,
													ValidateFunc: validation.IntBetween(100, 599),
													Optional:     true,
												},
												"body": {
													Type:        schema.TypeString,
													Description: routeHTTPDirectResponseActionBodySchemaDescription,
													Optional:    true,
												},
											},
										},
									},
									"http_match": {
										Type:        schema.TypeList,
										Description: routeHTTPMatchSchemaDescription,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_method": {
													Type:        schema.TypeSet,
													Description: routeHTTPMatchMethodSchemaDescription,
													Optional:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
												},
												"path": stringMatch(),
											},
										},
									},
								},
							},
						},
						"grpc_route": {
							Type:        schema.TypeList,
							Description: routeGRPCRouteSchemaDescription,
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"grpc_match": {
										Type:        schema.TypeList,
										Description: routeGRPCRouteMatchSchemaDescription,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqmn": stringMatch(),
											},
										},
									},
									"grpc_route_action": {
										Type:        schema.TypeList,
										Description: routeGRPCRouteActionSchemaDescription,
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:        schema.TypeString,
													Description: routeGRPCRouteActionBackendGroupIDSchemaDescription,
													Required:    true,
												},
												"max_timeout": {
													Type:             schema.TypeString,
													Description:      routeGRPCRouteActionMaxTimeoutSchemaDescription,
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"idle_timeout": {
													Type:             schema.TypeString,
													Description:      routeGRPCRouteActionIdleTimeoutSchemaDescription,
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"host_rewrite": {
													Type:        schema.TypeString,
													Description: routeGRPCRouteActionHostRewriteSchemaDescription,
													Optional:    true,
												},
												"auto_host_rewrite": {
													Type:        schema.TypeBool,
													Description: routeGRPCRouteActionAutoHostRewriteSchemaDescription,
													Optional:    true,
												},
												rateLimitSchemaKey: rateLimit(),
											},
										},
									},
									"grpc_status_response_action": {
										Type:        schema.TypeList,
										Description: routeGRPCStatusResponseActionSchemaDescription,
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:             schema.TypeString,
													Description:      routeGRPCStatusResponseActionStatusSchemaDescription,
													Optional:         true,
													DiffSuppressFunc: CaseInsensitive,
												},
											},
										},
									},
								},
							},
						},
						disableSecurityProfileSchemaKey: {
							Type:        schema.TypeBool,
							Description: disableSecurityProfileSchemaDescription,
							Optional:    true,
						},
						"route_options": routeOptions(),
					},
				},
			},
			"route_options": routeOptions(),
		},
	}
}

func stringMatch() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: stringMatchSchemaDescription,
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"exact": {
					Type:        schema.TypeString,
					Description: stringMatchExactSchemaDescription,
					Optional:    true,
				},
				"prefix": {
					Type:        schema.TypeString,
					Description: "Match prefix.",
					Optional:    true,
				},
				"regex": {
					Type:        schema.TypeString,
					Description: stringMatchRegexSchemaDescription,
					Optional:    true,
				},
			},
		},
	}
}

func headerModification() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: headerModificationSchemaDescription,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Description: headerModificationNameSchemaDescription,
					Required:    true,
				},
				"append": {
					Type:        schema.TypeString,
					Description: headerModificationAppendSchemaDescription,
					Optional:    true,
				},
				"replace": {
					Type:        schema.TypeString,
					Description: headerModificationReplaceSchemaDescription,
					Optional:    true,
				},
				"remove": {
					Type:        schema.TypeBool,
					Description: headerModificationRemoveSchemaDescription,
					Optional:    true,
				},
			},
		},
	}
}

func regexRewrite() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: regexRewriteSchemaDescription,
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				regexSchemaKey: {
					Type:        schema.TypeString,
					Description: regexSchemaDescription,
					Optional:    true,
				},
				substituteSchemaKey: {
					Type:        schema.TypeString,
					Description: substituteSchemaDescription,
					Optional:    true,
				},
			},
		},
	}
}

func rateLimit() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: rateLimitSchemaDescription,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				allRequestsSchemaKey: {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: allRequestsSchemaDescription,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							perSecondSchemaKey: {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: perSecondSchemaDescription,
							},
							perMinuteSchemaKey: {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: perMinuteSchemaDescription,
							},
						},
					},
				},
				requestsPerIPSchemaKey: {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: requestsPerIPSchemaDescription,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							perSecondSchemaKey: {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: perSecondSchemaDescription,
							},
							perMinuteSchemaKey: {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: perMinuteSchemaDescription,
							},
						},
					},
				},
			},
		},
	}
}

func resourceYandexALBVirtualHostCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Creating Application Virtual Host %q", d.Get("name"))

	req, err := buildALBVirtualHostCreateRequest(d)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().VirtualHost().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Application Virtual Host: %w", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Application Virtual Host create operation metadata: %w", err)
	}

	md, ok := protoMetadata.(*apploadbalancer.CreateVirtualHostMetadata)
	if !ok {
		return fmt.Errorf("could not get Application Virtual Host ID from create operation metadata")
	}

	d.SetId(md.HttpRouterId + "/" + md.VirtualHostName)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Application Virtual Host: %w", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Application Virtual Host creation failed: %w", err)
	}

	log.Printf("[DEBUG] Finished creating Application Virtual Host %q", d.Id())
	return resourceYandexALBVirtualHostRead(d, meta)
}

func buildALBVirtualHostCreateRequest(d *schema.ResourceData) (*apploadbalancer.CreateVirtualHostRequest, error) {
	authority, err := expandALBStringListFromSchemaSet(d.Get("authority"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding authority while updating Application Virtual Host: %w", err)
	}

	routes, err := expandALBRoutes(d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding routes while updating Application Virtual Host: %w", err)
	}

	requestHeaders, err := expandALBHeaderModification(d, "modify_request_headers")
	if err != nil {
		return nil, fmt.Errorf("Error expanding modify request headers while updating Application Virtual Host: %w", err)
	}

	responseHeaders, err := expandALBHeaderModification(d, "modify_response_headers")
	if err != nil {
		return nil, fmt.Errorf("Error expanding modify response headers while updating Application Virtual Host: %w", err)
	}

	rateLimit, err := expandALBRateLimit("", d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding rate limit while updating Application Virtual Host: %w", err)
	}

	req := &apploadbalancer.CreateVirtualHostRequest{
		HttpRouterId:          d.Get("http_router_id").(string),
		Name:                  d.Get("name").(string),
		Authority:             authority,
		Routes:                routes,
		ModifyResponseHeaders: responseHeaders,
		ModifyRequestHeaders:  requestHeaders,
		RateLimit:             rateLimit,
	}

	if _, ok := d.GetOk("route_options"); ok {
		ro, err := expandALBRouteOptions(d, "route_options.0.")
		if err != nil {
			return nil, fmt.Errorf("Error expanding route options while creating Application Virtual Host: %w", err)
		}
		req.SetRouteOptions(ro)
	}

	return req, nil
}

func resourceYandexALBVirtualHostRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading Application Virtual Host %q", d.Id())
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	virtualHost, err := config.sdk.ApplicationLoadBalancer().VirtualHost().Get(ctx, &apploadbalancer.GetVirtualHostRequest{
		HttpRouterId:    d.Get("http_router_id").(string),
		VirtualHostName: d.Get("name").(string),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Virtual Host %q", d.Get("name").(string)))
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

	ro, err := flattenALBRouteOptions(virtualHost.GetRouteOptions())
	if err != nil {
		return err
	}

	rateLimit := flattenALBRateLimit(virtualHost.GetRateLimit())

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

	if err := d.Set("route_options", ro); err != nil {
		return err
	}

	if err := d.Set(rateLimitSchemaKey, rateLimit); err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished reading Application Virtual Host %q", d.Id())
	return nil
}

func resourceYandexALBVirtualHostUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Updating Application Virtual Host %q", d.Id())

	req, err := buildALBVirtualHostUpdateRequest(d)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().VirtualHost().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Application Virtual Host %q: %w", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Application Virtual Host %q: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating Application Virtual Host %q", d.Id())
	return resourceYandexALBVirtualHostRead(d, meta)
}

func buildALBVirtualHostUpdateRequest(d *schema.ResourceData) (*apploadbalancer.UpdateVirtualHostRequest, error) {
	authority, err := expandALBStringListFromSchemaSet(d.Get("authority"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding authority while updating Application Virtual Host: %w", err)
	}

	routes, err := expandALBRoutes(d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding routes while updating Application Virtual Host: %w", err)
	}

	requestHeaders, err := expandALBHeaderModification(d, "modify_request_headers")
	if err != nil {
		return nil, fmt.Errorf("Error expanding modify request headers while updating Application Virtual Host: %w", err)
	}

	responseHeaders, err := expandALBHeaderModification(d, "modify_response_headers")
	if err != nil {
		return nil, fmt.Errorf("Error expanding modify response headers while updating Application Virtual Host: %w", err)
	}

	rateLimit, err := expandALBRateLimit("", d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding rate limit while updating Application Virtual Host: %w", err)
	}

	req := &apploadbalancer.UpdateVirtualHostRequest{
		VirtualHostName:       d.Get("name").(string),
		HttpRouterId:          d.Get("http_router_id").(string),
		Authority:             authority,
		Routes:                routes,
		ModifyResponseHeaders: responseHeaders,
		ModifyRequestHeaders:  requestHeaders,
		RateLimit:             rateLimit,
	}

	if _, ok := d.GetOk("route_options"); ok {
		ro, err := expandALBRouteOptions(d, "route_options.0.")
		if err != nil {
			return nil, fmt.Errorf("Error expanding route options while updating Application Virtual Host: %w", err)
		}
		req.SetRouteOptions(ro)
	}

	return req, nil
}

func resourceYandexALBVirtualHostDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Application Virtual Host %q", d.Id())

	req := &apploadbalancer.DeleteVirtualHostRequest{
		VirtualHostName: d.Get("name").(string),
		HttpRouterId:    d.Get("http_router_id").(string),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().VirtualHost().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Virtual Host %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Application Virtual Host %q", d.Id())
	return nil
}
