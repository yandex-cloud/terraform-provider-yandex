package cdn_resource

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

// CDNResourceSchema returns the schema for yandex_cdn_resource
func CDNResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version:             1, // Incremented for edge_cache_settings and browser_cache_settings refactoring
		MarkdownDescription: "Allows management of [Yandex Cloud CDN Resource](https://yandex.cloud/docs/cdn/concepts/resource).\n\n~> CDN provider must be activated prior usage of CDN resources, either via UI console or via yc cli command: `yc cdn provider activate --folder-id <folder-id> --type gcore`.",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cname": schema.StringAttribute{
				MarkdownDescription: "CDN endpoint CNAME, must be unique among resources.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"provider_type": schema.StringAttribute{
				MarkdownDescription: `CDN provider is a content delivery service provider. Possible values: "ourcdn" (default) or "gcore"`,
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("ourcdn", "gcore"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"folder_id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["folder_id"],
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: common.ResourceDescriptions["labels"],
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Flag to create Resource either in active or disabled state. `True` - the content from CDN is available to clients.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Default: booldefault.StaticBool(true),
			},
			"secondary_hostnames": schema.SetAttribute{
				MarkdownDescription: "List of secondary hostname strings.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"origin_protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol of origin resource. `http` or `https`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("http"),
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["created_at"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Last update timestamp. Computed value for read and update operations.",
				Computed:            true,
				// No plan modifier - this field changes on every update
			},
			"origin_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of a specific origin group.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"origin_group_name": schema.StringAttribute{
				MarkdownDescription: "The name of a specific origin group.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"shielding": schema.StringAttribute{
				MarkdownDescription: "Shielding is a Cloud CDN feature that helps reduce the load on content origins from CDN servers.\nSpecify location id to enable shielding. See https://yandex.cloud/en/docs/cdn/operations/resources/enable-shielding",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("1", "130"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provider_cname": schema.StringAttribute{
				MarkdownDescription: "Provider CNAME of CDN resource, computed value for read and update operations.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"ssl_certificate": SSLCertificateSchema(),
			"options":         CDNOptionsSchema(),
		},
	}
}

// SSLCertificateSchema returns the schema for ssl_certificate block
func SSLCertificateSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: "SSL certificate of CDN resource.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					MarkdownDescription: "SSL certificate type.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("not_used", "certificate_manager", "lets_encrypt"),
					},
				},
				"status": schema.StringAttribute{
					MarkdownDescription: "SSL certificate status.",
					Computed:            true,
				},
				"certificate_manager_id": schema.StringAttribute{
					MarkdownDescription: "Certificate Manager ID.",
					Optional:            true,
				},
			},
		},
		Validators: []validator.Set{
			setvalidator.SizeAtMost(1),
		},
	}
}

// CDNOptionsSchema returns the schema for options block
func CDNOptionsSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "CDN Resource settings and options to tune CDN edge behavior.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				// Boolean options - Optional+Computed WITHOUT Default for tristate support
				"ignore_query_params": schema.BoolAttribute{
					MarkdownDescription: "Files with different query parameters are cached as objects with the same key regardless of the parameter value. selected by default.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
					Validators: []validator.Bool{
						// Conflicts handled in CustomizeDiff at resource level
					},
				},
				"slice": schema.BoolAttribute{
					MarkdownDescription: "Files larger than 10 MB will be requested and cached in parts (no larger than 10 MB each part). It reduces time to first byte. The origin must support HTTP Range requests.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"fetched_compressed": schema.BoolAttribute{
					MarkdownDescription: "Option helps you to reduce the bandwidth between origin and CDN servers. Also, content delivery speed becomes higher because of reducing the time for compressing files in a CDN.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"gzip_on": schema.BoolAttribute{
					MarkdownDescription: "GZip compression at CDN servers reduces file size by 70% and can be as high as 90%.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"redirect_http_to_https": schema.BoolAttribute{
					MarkdownDescription: "Set up a redirect from HTTP to HTTPS.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"redirect_https_to_http": schema.BoolAttribute{
					MarkdownDescription: "Set up a redirect from HTTPS to HTTP.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				// CRITICAL: forward_host_header - the field that caused the bug in SDKv2
				"forward_host_header": schema.BoolAttribute{
					MarkdownDescription: "Choose the Forward Host header option if is important to send in the request to the Origin the same Host header as was sent in the request to CDN server.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"proxy_cache_methods_set": schema.BoolAttribute{
					MarkdownDescription: "Allows caching for GET, HEAD and POST requests.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"disable_proxy_force_ranges": schema.BoolAttribute{
					MarkdownDescription: "Disabling proxy force ranges.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"ignore_cookie": schema.BoolAttribute{
					MarkdownDescription: "Set for ignoring cookie.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"enable_ip_url_signing": schema.BoolAttribute{
					MarkdownDescription: "Enable access limiting by IP addresses, option available only with setting secure_key.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},

				// String options
				"custom_host_header": schema.StringAttribute{
					MarkdownDescription: "Custom value for the Host header. Your server must be able to process requests with the chosen header.",
					Optional:            true,
					Computed:            true,
				},
				"custom_server_name": schema.StringAttribute{
					MarkdownDescription: "Wildcard additional CNAME. If a resource has a wildcard additional CNAME, you can use your own certificate for content delivery via HTTPS.",
					Optional:            true,
					Computed:            true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`),
							"must be a valid domain name",
						),
					},
				},
				"secure_key": schema.StringAttribute{
					MarkdownDescription: "Set secure key for url encoding to protect content and limit access by IP addresses and time limits.",
					Optional:            true,
					Computed:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(6, 32),
					},
				},

				// List options
				"cache_http_headers": schema.ListAttribute{
					MarkdownDescription: "List HTTP headers that must be included in responses to clients.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
					ElementType: types.StringType,
				},
				"query_params_whitelist": schema.ListAttribute{
					MarkdownDescription: "Files with the specified query parameters are cached as objects with different keys, files with other parameters are cached as objects with the same key.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
					ElementType: types.StringType,
				},
				"query_params_blacklist": schema.ListAttribute{
					MarkdownDescription: "Files with the specified query parameters are cached as objects with the same key, files with other parameters are cached as objects with different keys.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
					ElementType: types.StringType,
				},
				"cors": schema.ListAttribute{
					MarkdownDescription: "Parameter that lets browsers get access to selected resources from a domain different to a domain from which the request is received.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
					ElementType: types.StringType,
				},
				"allowed_http_methods": schema.ListAttribute{
					MarkdownDescription: "HTTP methods for your CDN content. By default the following methods are allowed: GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS. In case some methods are not allowed to the user, they will get the 405 (Method Not Allowed) response. If the method is not supported, the user gets the 501 (Not Implemented) response.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
					ElementType: types.StringType,
				},

				// Map options
				"static_response_headers": schema.MapAttribute{
					MarkdownDescription: "Set up a static response header. The header name must be lowercase.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.UseStateForUnknown(),
					},
					ElementType: types.StringType,
					Validators: []validator.Map{
						NewStaticHeadersValidator(),
					},
				},
				"static_request_headers": schema.MapAttribute{
					MarkdownDescription: "Set up custom headers that CDN servers will send in requests to origins.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.UseStateForUnknown(),
					},
					ElementType: types.StringType,
					Validators: []validator.Map{
						NewStaticHeadersValidator(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"edge_cache_settings":    EdgeCacheSettingsSchema(),
				"browser_cache_settings": BrowserCacheSettingsSchema(),
				"ip_address_acl":         IPAddressACLSchema(),
				"rewrite":                RewriteSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// IPAddressACLSchema returns the schema for ip_address_acl block
func IPAddressACLSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "IP address access control list. The list of specified IP addresses to be allowed or denied depending on acl policy type.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"policy_type": schema.StringAttribute{
					MarkdownDescription: "The policy type for ACL. One of `allow` or `deny` values.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("allow", "deny"),
					},
				},
				"excepted_values": schema.ListAttribute{
					MarkdownDescription: "The list of specified IP addresses to be allowed or denied depending on acl policy type.",
					Required:            true,
					ElementType:         types.StringType,
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 200),
						listvalidator.ValueStringsAre(NewIPAddressOrCIDRValidator()),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// RewriteSchema returns the schema for rewrite block
func RewriteSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "An option for changing or redirecting query paths.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					MarkdownDescription: "True - the rewrite option is enabled and its flag is applied to the resource. False - the rewrite option is disabled. Default is false.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"body": schema.StringAttribute{
					MarkdownDescription: "Pattern for rewrite. The value must have the following format: `<source path> <destination path>`, where both paths are regular expressions which use at least one group. E.g., `/foo/(.*) /bar/$1`.",
					Required:            true,
					Validators: []validator.String{
						NewRewriteBodyValidator(),
					},
				},
				"flag": schema.StringAttribute{
					MarkdownDescription: "Rewrite flag. Available values: 'last', 'break', 'redirect', 'permanent'. Default is 'break'.",
					Optional:            true,
					Computed:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("last", "break", "redirect", "permanent"),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// EdgeCacheSettingsSchema returns the schema for edge_cache_settings block
func EdgeCacheSettingsSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "Content will be cached according to origin cache settings. Use either `default_value` for simple caching or `custom_values` for per-HTTP-code caching. The value applies for response codes 200, 201, 204, 206, 301, 302, 303, 304, 307, 308 if origin server does not have caching HTTP headers. **By default, edge caching is enabled in Yandex CDN.** To explicitly disable it, set `enabled = false` (provider will send `cache_time = 0` to API). Alternatively, you can set `enabled = true` with `cache_time = {\"*\" = 0}`. To remove the configuration entirely, omit this block.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					MarkdownDescription: "True - caching is enabled with `cache_time` settings. False - caching is disabled (provider sends `cache_time = 0` to API). Use `enabled = false` to explicitly disable edge caching (which is enabled by default in Yandex CDN). Cannot be used together with `cache_time`.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"cache_time": schema.MapAttribute{
					MarkdownDescription: "Cache time in seconds. Use `\"*\"` as key for default cache time for all HTTP codes (200, 201, 204, 206, 301, 302, 303, 304, 307, 308), or specify cache times per HTTP code (e.g., `{\"200\" = 3600, \"404\" = 300}`). Use `{\"*\" = 0}` to explicitly disable caching. Required when `enabled = true`, must not be set when `enabled = false`.",
					ElementType:         types.Int64Type,
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			NewEdgeCacheSettingsValidator(),
		},
	}
}

// BrowserCacheSettingsSchema returns the schema for browser_cache_settings block
func BrowserCacheSettingsSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "Set up a cache period for the end-users browser. Content will be cached due to origin settings. If there are no cache settings on your origin, the content will not be cached. The list of HTTP response codes that can be cached in browsers: 200, 201, 204, 206, 301, 302, 303, 304, 307, 308. Other response codes will not be cached. The default value is 4 days. **By default, browser caching is enabled in Yandex CDN.** To explicitly disable it, set `enabled = false` (provider will send `cache_time = 0` to API). Alternatively, you can set `enabled = true` with `cache_time = 0`. To remove the configuration entirely, omit this block.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					MarkdownDescription: "True - browser caching is enabled with `cache_time` setting. False - browser caching is disabled (provider sends `cache_time = 0` to API). Use `enabled = false` to explicitly disable browser caching (which is enabled by default in Yandex CDN). Cannot be used together with `cache_time`.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"cache_time": schema.Int64Attribute{
					MarkdownDescription: "Cache time in seconds for browsers. Must be between 0 and 31536000 (1 year). Use `0` to explicitly disable caching. Required when `enabled = true`, must not be set when `enabled = false`.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
					Validators: []validator.Int64{
						int64validator.Between(0, 31536000),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			NewBrowserCacheSettingsValidator(),
		},
	}
}
