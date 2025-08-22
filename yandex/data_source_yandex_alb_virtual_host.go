package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const (
	allRequestsSchemaKey            = "all_requests"
	perMinuteSchemaKey              = "per_minute"
	perSecondSchemaKey              = "per_second"
	prefixRewriteSchemaKey          = "prefix_rewrite"
	rateLimitSchemaKey              = "rate_limit"
	regexSchemaKey                  = "regex"
	regexRewriteSchemaKey           = "regex_rewrite"
	requestsPerIPSchemaKey          = "requests_per_ip"
	substituteSchemaKey             = "substitute"
	disableSecurityProfileSchemaKey = "disable_security_profile"
)

const (
	allRequestsSchemaDescription   = "Rate limit configuration applied to all incoming requests"
	perMinuteSchemaDescription     = "Limit value specified with per minute time unit"
	perSecondSchemaDescription     = "Limit value specified with per second time unit"
	regexSchemaDescription         = "RE2 regular expression"
	regexRewriteSchemaDescription  = "Replacement for path substrings that match the pattern"
	rateLimitSchemaDescription     = "Rate limit configuration applied for a whole virtual host"
	requestsPerIPSchemaDescription = "Rate limit configuration applied separately for each set of requests grouped by client IP address"
	routeSchemaDescription         = "A Route resource. Routes are matched *in-order*. Be careful when adding them to the end. For instance, having http '/' match first makes all other routes unused.\n\n~> Exactly one type of routes `http_route` or `grpc_route` should be specified.\n"

	routeNameSchemaDescription = "Name of the route."

	routeHTTPRouteActionBackendGroupIDSchemaDescription  = "Backend group to route requests."
	routeHTTPRouteActionSchemaDescription                = "HTTP route action resource.\n\n~> Only one type of host rewrite specifiers `host_rewrite` or `auto_host_rewrite` should be specified.\n"
	routeHTTPRouteActionTimeoutSchemaDescription         = "Specifies the request timeout (overall time request processing is allowed to take) for the route. If not set, default is 60 seconds."
	routeHTTPRouteActionIdleTimeoutSchemaDescription     = "Specifies the idle timeout (time without any data transfer for the active request) for the route. It is useful for streaming scenarios (i.e. long-polling, server-sent events) - one should set idle_timeout to something meaningful and timeout to the maximum time the stream is allowed to be alive. If not specified, there is no per-route idle timeout."
	routeHTTPRouteActionPrefixRewriteSchemaDescription   = "If not empty, matched path prefix will be replaced by this value."
	routeHTTPRouteActionUpgradeTypesSchemaDescription    = "List of upgrade types. Only specified upgrade types will be allowed. For example, `websocket`."
	routeHTTPRouteActionHostRewriteSchemaDescription     = "Host rewrite specifier."
	routeHTTPRouteActionAutoHostRewriteSchemaDescription = "If set, will automatically rewrite host."

	routeHTTPRedirectActionSchemaDescription              = "Redirect action resource.\n\n~> Only one type of paths `replace_path` or `replace_prefix` should be specified.\n"
	routeHTTPRedirectActionReplaceSchemeSchemaDescription = "Replaces scheme. If the original scheme is `http` or `https`, will also remove the 80 or 443 port, if present."
	routeHTTPRedirectActionReplaceHostSchemaDescription   = "Replaces hostname."
	routeHTTPRedirectActionReplacePortSchemaDescription   = "Replaces port."
	routeHTTPRedirectActionRemoveQuerySchemaDescription   = "If set, remove query part."
	routeHTTPRedirectActionResponseCodeSchemaDescription  = "The HTTP status code to use in the redirect response. Supported values are: `moved_permanently`, `found`, `see_other`, `temporary_redirect`, `permanent_redirect`."
	routeHTTPRedirectActionReplacePathSchemaDescription   = "Replace path."
	routeHTTPRedirectActionReplacePrefixSchemaDescription = "Replace only matched prefix. Example:<br/> match:{ prefix_match: `/some` } <br/> redirect: { replace_prefix: `/other` } <br/> will redirect `/something` to `/otherthing`."
	routeHTTPDirectResponseActionSchemaDescription        = "Direct response action resource."
	routeHTTPDirectResponseActionStatusSchemaDescription  = "HTTP response status. Should be between `100` and `599`."
	routeHTTPDirectResponseActionBodySchemaDescription    = "Response body text."

	routeHTTPMatchSchemaDescription       = "Checks `/` prefix by default."
	routeHTTPMatchMethodSchemaDescription = "List of methods (strings)."
	routeHTTPRouteSchemaDescription       = "HTTP route resource.\n\n~> Exactly one type of actions `http_route_action` or `redirect_action` or `direct_response_action` should be specified.\n"

	routeGRPCRouteSchemaDescription                      = "gRPC route resource.\n\n~> Exactly one type of actions `grpc_route_action` or `grpc_status_response_action` should be specified.\n"
	routeGRPCRouteMatchSchemaDescription                 = "Checks `/` prefix by default."
	routeGRPCRouteActionSchemaDescription                = "gRPC route action resource.\n\n~> Only one type of host rewrite specifiers `host_rewrite` or `auto_host_rewrite` should be specified.\n"
	routeGRPCRouteActionBackendGroupIDSchemaDescription  = "Backend group to route requests."
	routeGRPCRouteActionMaxTimeoutSchemaDescription      = "Lower timeout may be specified by the client (using grpc-timeout header). If not set, default is 60 seconds."
	routeGRPCRouteActionIdleTimeoutSchemaDescription     = "Specifies the idle timeout (time without any data transfer for the active request) for the route. It is useful for streaming scenarios - one should set idle_timeout to something meaningful and max_timeout to the maximum time the stream is allowed to be alive. If not specified, there is no per-route idle timeout."
	routeGRPCRouteActionHostRewriteSchemaDescription     = "Host rewrite specifier."
	routeGRPCRouteActionAutoHostRewriteSchemaDescription = "If set, will automatically rewrite host."
	routeGRPCStatusResponseActionSchemaDescription       = "gRPC status response action resource."
	routeGRPCStatusResponseActionStatusSchemaDescription = "The status of the response. Supported values are: ok, invalid_argumet, not_found, permission_denied, unauthenticated, unimplemented, internal, unavailable."

	headerModificationSchemaDescription        = "Apply the following modifications to the Request/Response header.\n\n~> Only one type of actions `append` or `replace` or `remove` should be specified.\n"
	headerModificationNameSchemaDescription    = "Name of the header to modify."
	headerModificationAppendSchemaDescription  = "Append string to the header value."
	headerModificationReplaceSchemaDescription = "New value for a header. Header values support the following [formatters](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#custom-request-response-headers)."
	headerModificationRemoveSchemaDescription  = "If set, remove the header."

	stringMatchSchemaDescription       = "The `path` and `fqmn` blocks.\n\n~> Exactly one type of string matches `exact`, `prefix` or `regex` should be specified.\n"
	stringMatchExactSchemaDescription  = "Match exactly."
	stringMatchPrefixSchemaDescription = "Match prefix."
	stringMatchRegexSchemaDescription  = "Match regex."

	substituteSchemaDescription = "The string which should be used to substitute matched substrings"

	disableSecurityProfileSchemaDescription = "Disables security profile for the route"
)

func dataSourceYandexALBVirtualHost() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex ALB Virtual Host. For more information, see [Yandex Cloud Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/quickstart).\n\nThis data source is used to define [Application Load Balancer Virtual Host](https://yandex.cloud/docs/application-load-balancer/concepts/http-router) that can be used by other resources.\n\n~> One of `virtual_host_id` or `name` with `http_router_id` should be specified.\n",
		Read:        dataSourceYandexALBVirtualHostRead,

		Schema: map[string]*schema.Schema{
			"virtual_host_id": {
				Type:        schema.TypeString,
				Description: "The ID of a specific Virtual Host. Virtual Host ID is concatenation of HTTP Router ID and Virtual Host name with `/` symbol between them. For Example, `http_router_id/vhost_name`.",
				Optional:    true,
				Computed:    true,
			},
			"http_router_id": {
				Type:        schema.TypeString,
				Description: resourceYandexALBVirtualHost().Schema["http_router_id"].Description,
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: resourceYandexALBVirtualHost().Schema["name"].Description,
				Optional:    true,
				Computed:    true,
			},
			"authority": {
				Type:        schema.TypeSet,
				Description: resourceYandexALBVirtualHost().Schema["authority"].Description,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Computed:    true,
			},
			"modify_request_headers":  dataSourceHeaderModification("modify_request_headers."),
			"modify_response_headers": dataSourceHeaderModification("modify_response_headers."),
			"route_options":           dataSourceRouteOptions(),
			rateLimitSchemaKey:        dataSourceRateLimit(),
			"route": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: routeSchemaDescription,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: routeNameSchemaDescription,
							Computed:    true,
						},
						"route_options": dataSourceRouteOptions(),
						disableSecurityProfileSchemaKey: {
							Type:        schema.TypeBool,
							Description: disableSecurityProfileSchemaDescription,
							Computed:    true,
						},
						"http_route": {
							Type:        schema.TypeList,
							Description: routeHTTPRouteSchemaDescription,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_route_action": {
										Type:        schema.TypeList,
										Description: routeHTTPRouteActionSchemaDescription,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:        schema.TypeString,
													Description: routeHTTPRouteActionBackendGroupIDSchemaDescription,
													Computed:    true,
												},
												"timeout": {
													Type:        schema.TypeString,
													Description: routeHTTPRouteActionTimeoutSchemaDescription,
													Computed:    true,
												},
												"idle_timeout": {
													Type:        schema.TypeString,
													Description: routeHTTPRouteActionIdleTimeoutSchemaDescription,
													Computed:    true,
												},
												"prefix_rewrite": {
													Type:        schema.TypeString,
													Description: routeHTTPRouteActionPrefixRewriteSchemaDescription,
													Computed:    true,
												},
												regexRewriteSchemaKey: dataSourceRegexRewrite(),
												"upgrade_types": {
													Type:        schema.TypeSet,
													Description: routeHTTPRouteActionUpgradeTypesSchemaDescription,
													Computed:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
												},
												"host_rewrite": {
													Type:        schema.TypeString,
													Description: routeHTTPRouteActionHostRewriteSchemaDescription,
													Computed:    true,
												},
												"auto_host_rewrite": {
													Type:        schema.TypeBool,
													Description: routeHTTPRouteActionAutoHostRewriteSchemaDescription,
													Computed:    true,
												},
												rateLimitSchemaKey: dataSourceRateLimit(),
											},
										},
									},
									"redirect_action": {
										Type:        schema.TypeList,
										Description: routeHTTPRedirectActionSchemaDescription,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"replace_scheme": {
													Type:        schema.TypeString,
													Description: routeHTTPRedirectActionReplaceSchemeSchemaDescription,
													Computed:    true,
												},
												"replace_host": {
													Type:        schema.TypeString,
													Description: routeHTTPRedirectActionReplaceHostSchemaDescription,
													Computed:    true,
												},
												"replace_port": {
													Type:        schema.TypeInt,
													Description: routeHTTPRedirectActionReplacePortSchemaDescription,
													Computed:    true,
												},
												"remove_query": {
													Type:        schema.TypeBool,
													Description: routeHTTPRedirectActionRemoveQuerySchemaDescription,
													Computed:    true,
												},
												"response_code": {
													Type:        schema.TypeString,
													Description: routeHTTPRedirectActionResponseCodeSchemaDescription,
													Computed:    true,
												},
												"replace_path": {
													Type:        schema.TypeString,
													Description: routeHTTPRedirectActionReplacePathSchemaDescription,
													Computed:    true,
												},
												"replace_prefix": {
													Type:        schema.TypeString,
													Description: routeHTTPRedirectActionReplacePrefixSchemaDescription,
													Computed:    true,
												},
											},
										},
									},
									"direct_response_action": {
										Type:        schema.TypeList,
										Description: routeHTTPDirectResponseActionSchemaDescription,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:        schema.TypeInt,
													Description: routeHTTPDirectResponseActionStatusSchemaDescription,
													Computed:    true,
												},
												"body": {
													Type:        schema.TypeString,
													Description: routeHTTPDirectResponseActionBodySchemaDescription,
													Computed:    true,
												},
											},
										},
									},
									"http_match": {
										Type:        schema.TypeList,
										Description: routeHTTPMatchSchemaDescription,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"http_method": {
													Type:        schema.TypeSet,
													Description: routeHTTPMatchMethodSchemaDescription,
													Computed:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
												},
												"path": dataSourceStringMatch(),
											},
										},
									},
								},
							},
						},
						"grpc_route": {
							Type:        schema.TypeList,
							Description: routeGRPCRouteSchemaDescription,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"grpc_match": {
										Type:        schema.TypeList,
										Description: routeGRPCRouteMatchSchemaDescription,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqmn": dataSourceStringMatch(),
											},
										},
									},
									"grpc_route_action": {
										Type:        schema.TypeList,
										Description: routeGRPCRouteActionSchemaDescription,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"backend_group_id": {
													Type:        schema.TypeString,
													Description: routeGRPCRouteActionBackendGroupIDSchemaDescription,
													Computed:    true,
												},
												"max_timeout": {
													Type:        schema.TypeString,
													Description: routeGRPCRouteActionMaxTimeoutSchemaDescription,
													Computed:    true,
												},
												"idle_timeout": {
													Type:        schema.TypeString,
													Description: routeGRPCRouteActionIdleTimeoutSchemaDescription,
													Computed:    true,
												},
												"host_rewrite": {
													Type:        schema.TypeString,
													Description: routeGRPCRouteActionHostRewriteSchemaDescription,
													Computed:    true,
												},
												"auto_host_rewrite": {
													Type:        schema.TypeBool,
													Description: routeGRPCRouteActionAutoHostRewriteSchemaDescription,
													Computed:    true,
												},
												rateLimitSchemaKey: dataSourceRateLimit(),
											},
										},
									},
									"grpc_status_response_action": {
										Type:        schema.TypeList,
										Description: routeGRPCStatusResponseActionSchemaDescription,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:        schema.TypeString,
													Description: routeGRPCStatusResponseActionStatusSchemaDescription,
													Computed:    true,
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
		Type:        schema.TypeList,
		Description: headerModificationSchemaDescription,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Description: headerModificationNameSchemaDescription,
					Computed:    true,
				},
				"append": {
					Type:        schema.TypeString,
					Description: headerModificationAppendSchemaDescription,
					Computed:    true,
				},
				"replace": {
					Type:        schema.TypeString,
					Description: headerModificationReplaceSchemaDescription,
					Computed:    true,
				},
				"remove": {
					Type:        schema.TypeBool,
					Description: headerModificationRemoveSchemaDescription,
					Computed:    true,
				},
			},
		},
	}
}

func dataSourceStringMatch() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: stringMatchSchemaDescription,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"exact": {
					Type:        schema.TypeString,
					Description: stringMatchExactSchemaDescription,
					Computed:    true,
				},
				"prefix": {
					Type:        schema.TypeString,
					Description: stringMatchPrefixSchemaDescription,
					Computed:    true,
				},
				"regex": {
					Type:        schema.TypeString,
					Description: stringMatchRegexSchemaDescription,
					Computed:    true,
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

	ro, err := flattenALBRouteOptions(virtualHost.GetRouteOptions())
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

	if err := d.Set("route_options", ro); err != nil {
		return err
	}

	d.SetId(virtualHostID.(string))

	return nil

}

func dataSourceRegexRewrite() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: regexRewriteSchemaDescription,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				regexSchemaKey: {
					Type:        schema.TypeString,
					Computed:    true,
					Description: regexSchemaDescription,
				},
				substituteSchemaKey: {
					Type:        schema.TypeString,
					Computed:    true,
					Description: substituteSchemaDescription,
				},
			},
		},
	}
}

func dataSourceRateLimit() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: rateLimitSchemaDescription,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				allRequestsSchemaKey: {
					Type:        schema.TypeList,
					Computed:    true,
					Description: allRequestsSchemaDescription,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							perSecondSchemaKey: {
								Type:        schema.TypeInt,
								Computed:    true,
								Description: perSecondSchemaDescription,
							},
							perMinuteSchemaKey: {
								Type:        schema.TypeInt,
								Computed:    true,
								Description: perMinuteSchemaDescription,
							},
						},
					},
				},
				requestsPerIPSchemaKey: {
					Type:        schema.TypeList,
					Computed:    true,
					Description: requestsPerIPSchemaDescription,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							perSecondSchemaKey: {
								Type:        schema.TypeInt,
								Computed:    true,
								Description: perSecondSchemaDescription,
							},
							perMinuteSchemaKey: {
								Type:        schema.TypeInt,
								Computed:    true,
								Description: perMinuteSchemaDescription,
							},
						},
					},
				},
			},
		},
	}
}
