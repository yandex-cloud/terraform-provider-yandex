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
				Description: "A Route resource. Routes are matched *in-order*. Be careful when adding them to the end. For instance, having http '/' match first makes all other routes unused.\n\n~> Exactly one type of routes `http_route` or `grpc_route` should be specified.\n",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the route.",
							Optional:    true,
						},
						"http_route": {
							Type:        schema.TypeList,
							Description: "HTTP route resource.\n\n~> Exactly one type of actions `http_route_action` or `redirect_action` or `direct_response_action` should be specified.\n",
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_route_action": {
										Type:        schema.TypeList,
										Description: "HTTP route action resource.\n\n~> Only one type of host rewrite specifiers `host_rewrite` or `auto_host_rewrite` should be specified.\n",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:        schema.TypeString,
													Description: "Backend group to route requests.",
													Required:    true,
												},
												"timeout": {
													Type:             schema.TypeString,
													Description:      "Specifies the request timeout (overall time request processing is allowed to take) for the route. If not set, default is 60 seconds.",
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"idle_timeout": {
													Type:             schema.TypeString,
													Description:      "Specifies the idle timeout (time without any data transfer for the active request) for the route. It is useful for streaming scenarios (i.e. long-polling, server-sent events) - one should set idle_timeout to something meaningful and timeout to the maximum time the stream is allowed to be alive. If not specified, there is no per-route idle timeout.",
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"prefix_rewrite": {
													Type:        schema.TypeString,
													Description: "If not empty, matched path prefix will be replaced by this value.",
													Optional:    true,
												},
												"upgrade_types": {
													Type:        schema.TypeSet,
													Description: "List of upgrade types. Only specified upgrade types will be allowed. For example, `websocket`.",
													Optional:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
												},
												"host_rewrite": {
													Type:        schema.TypeString,
													Description: "Host rewrite specifier.",
													Optional:    true,
												},
												"auto_host_rewrite": {
													Type:        schema.TypeBool,
													Description: "If set, will automatically rewrite host.",
													Optional:    true,
												},
												rateLimitSchemaKey: rateLimit(),
											},
										},
									},
									"redirect_action": {
										Type:        schema.TypeList,
										Description: "Redirect action resource.\n\n~> Only one type of paths `replace_path` or `replace_prefix` should be specified.\n",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"replace_scheme": {
													Type:        schema.TypeString,
													Description: "Replaces scheme. If the original scheme is `http` or `https`, will also remove the 80 or 443 port, if present.",
													Optional:    true,
												},
												"replace_host": {
													Type:        schema.TypeString,
													Description: "Replaces hostname.",
													Optional:    true,
												},
												"replace_port": {
													Type:        schema.TypeInt,
													Description: "Replaces port.",
													Optional:    true,
												},
												"remove_query": {
													Type:        schema.TypeBool,
													Description: "If set, remove query part.",
													Optional:    true,
												},
												"response_code": {
													Type:             schema.TypeString,
													Description:      "The HTTP status code to use in the redirect response. Supported values are: `moved_permanently`, `found`, `see_other`, `temporary_redirect`, `permanent_redirect`.",
													Default:          "moved_permanently",
													Optional:         true,
													DiffSuppressFunc: CaseInsensitive,
												},
												"replace_path": {
													Type:        schema.TypeString,
													Description: "Replace path.",
													Optional:    true,
												},
												"replace_prefix": {
													Type:        schema.TypeString,
													Description: "Replace only matched prefix. Example:<br/> match:{ prefix_match: `/some` } <br/> redirect: { replace_prefix: `/other` } <br/> will redirect `/something` to `/otherthing`.",
													Optional:    true,
												},
											},
										},
									},
									"direct_response_action": {
										Type:        schema.TypeList,
										Description: "Direct response action resource.",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:         schema.TypeInt,
													Description:  "HTTP response status. Should be between `100` and `599`.",
													ValidateFunc: validation.IntBetween(100, 599),
													Optional:     true,
												},
												"body": {
													Type:        schema.TypeString,
													Description: "Response body text.",
													Optional:    true,
												},
											},
										},
									},
									"http_match": {
										Type:        schema.TypeList,
										Description: "Checks `/` prefix by default.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_method": {
													Type:        schema.TypeSet,
													Description: "List of methods (strings).",
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
							Description: "gRPC route resource.\n\n~> Exactly one type of actions `grpc_route_action` or `grpc_status_response_action` should be specified.\n",
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"grpc_match": {
										Type:        schema.TypeList,
										Description: "Checks `/` prefix by default.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqmn": stringMatch(),
											},
										},
									},
									"grpc_route_action": {
										Type:        schema.TypeList,
										Description: "gRPC route action resource.\n\n~> Only one type of host rewrite specifiers `host_rewrite` or `auto_host_rewrite` should be specified.\n",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:        schema.TypeString,
													Description: "Backend group to route requests.",
													Required:    true,
												},
												"max_timeout": {
													Type:             schema.TypeString,
													Description:      "Lower timeout may be specified by the client (using grpc-timeout header). If not set, default is 60 seconds.",
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"idle_timeout": {
													Type:             schema.TypeString,
													Description:      "Specifies the idle timeout (time without any data transfer for the active request) for the route. It is useful for streaming scenarios - one should set idle_timeout to something meaningful and max_timeout to the maximum time the stream is allowed to be alive. If not specified, there is no per-route idle timeout.",
													Optional:         true,
													ValidateFunc:     validateParsableValue(parseDuration),
													DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
												},
												"host_rewrite": {
													Type:        schema.TypeString,
													Description: "Host rewrite specifier.",
													Optional:    true,
												},
												"auto_host_rewrite": {
													Type:        schema.TypeBool,
													Description: "If set, will automatically rewrite host.",
													Optional:    true,
												},
												rateLimitSchemaKey: rateLimit(),
											},
										},
									},
									"grpc_status_response_action": {
										Type:        schema.TypeList,
										Description: "gRPC status response action resource.",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:             schema.TypeString,
													Description:      "The status of the response. Supported values are: ok, invalid_argumet, not_found, permission_denied, unauthenticated, unimplemented, internal, unavailable.",
													Optional:         true,
													DiffSuppressFunc: CaseInsensitive,
												},
											},
										},
									},
								},
							},
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
		Description: "The `path` and `fqmn` blocks.\n\n~> Exactly one type of string matches `exact`, `prefix` or `regex` should be specified.\n",
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"exact": {
					Type:        schema.TypeString,
					Description: "Match exactly.",
					Optional:    true,
				},
				"prefix": {
					Type:        schema.TypeString,
					Description: "Match prefix.",
					Optional:    true,
				},
				"regex": {
					Type:        schema.TypeString,
					Description: "Match regex.",
					Optional:    true,
				},
			},
		},
	}
}

func headerModification() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Apply the following modifications to the Request/Response header.\n\n~> Only one type of actions `append` or `replace` or `remove` should be specified.\n",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Description: "Name of the header to modify.",
					Required:    true,
				},
				"append": {
					Type:        schema.TypeString,
					Description: "Append string to the header value.",
					Optional:    true,
				},
				"replace": {
					Type:        schema.TypeString,
					Description: "New value for a header. Header values support the following [formatters](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#custom-request-response-headers).",
					Optional:    true,
				},
				"remove": {
					Type:        schema.TypeBool,
					Description: "If set, remove the header.",
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
