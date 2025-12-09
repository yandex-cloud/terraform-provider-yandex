package cdn_rule

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cdn_resource"
)

func CDNRuleSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a CDN Rule. CDN Rules allow you to apply different CDN settings based on URL patterns.\n\n" +
			"Rules with lower weight values are executed first.",
		MarkdownDescription: "Manages a CDN Rule. CDN Rules allow you to apply different CDN settings based on URL patterns.\n\n" +
			"Rules with lower weight values are executed first.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the CDN rule in the format 'resource_id/rule_id'.",
				MarkdownDescription: "The ID of the CDN rule in the format `resource_id/rule_id`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_id": schema.StringAttribute{
				Required:            true,
				Description:         "CDN Resource ID to attach the rule to.",
				MarkdownDescription: "CDN Resource ID to attach the rule to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_id": schema.StringAttribute{
				Computed:            true,
				Description:         "Rule ID (stored as string to avoid int64 precision loss).",
				MarkdownDescription: "Rule ID (stored as string to avoid int64 precision loss).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Rule name (max 50 characters).",
				MarkdownDescription: "Rule name (max 50 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[^-]*$`), "Rule name cannot contain dash (-) character"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"rule_pattern": schema.StringAttribute{
				Required:            true,
				Description:         "Rule pattern - must be a valid regular expression (max 100 characters).",
				MarkdownDescription: "Rule pattern - must be a valid regular expression (max 100 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^.+$`),
						"must be a valid regular expression",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"weight": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Description:         "Rule weight (0-9999) - rules with lower weights execute first. Default: 0.",
				MarkdownDescription: "Rule weight (0-9999) - rules with lower weights execute first. Default: `0`.",
				Validators: []validator.Int64{
					int64validator.Between(0, 9999),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),

			// Reuse options schema from cdn_resource - identical structure
			"options": CDNOptionsSchema(),
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
					DeprecationMessage:  "This attribute does not affect anything. You can safely delete it.",
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
					Validators: []validator.List{
						listvalidator.ValueStringsAre(
							stringvalidator.OneOf("GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"),
						),
					},
					ElementType: types.StringType,
				},
				"stale": schema.ListAttribute{
					MarkdownDescription: "List of errors which instruct CDN servers to serve stale content to clients. Possible values: `error`, `http_403`, `http_404`, `http_429`, `http_500`, `http_502`, `http_503`, `http_504`, `invalid_header`, `timeout`, `updating`.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
					ElementType: types.StringType,
					Validators: []validator.List{
						listvalidator.ValueStringsAre(
							stringvalidator.OneOf("error", "http_403", "http_404", "http_429", "http_500", "http_502", "http_503", "http_504", "invalid_header", "timeout", "updating"),
						),
					},
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
						cdn_resource.NewStaticHeadersValidator(),
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
						cdn_resource.NewStaticHeadersValidator(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"edge_cache_settings":    cdn_resource.EdgeCacheSettingsSchema(),
				"browser_cache_settings": cdn_resource.BrowserCacheSettingsSchema(),
				"ip_address_acl":         cdn_resource.IPAddressACLSchema(),
				"rewrite":                cdn_resource.RewriteSchema(),
			},
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			cdn_resource.NewMutuallyExclusiveBoolsValidator(
				"slice", "fetched_compressed",
				func(o *cdn_resource.CDNOptionsModel) types.Bool { return o.Slice },
				func(o *cdn_resource.CDNOptionsModel) types.Bool { return o.FetchedCompressed },
				"Incompatible CDN options",
				"slice and fetched_compressed cannot both be enabled simultaneously. Set one of them to false.",
			),
			cdn_resource.NewMutuallyExclusiveBoolsValidator(
				"gzip_on", "fetched_compressed",
				func(o *cdn_resource.CDNOptionsModel) types.Bool { return o.GzipOn },
				func(o *cdn_resource.CDNOptionsModel) types.Bool { return o.FetchedCompressed },
				"Incompatible CDN compression options",
				"gzip_on and fetched_compressed cannot both be enabled simultaneously. These are mutually exclusive compression methods. Set one of them to false.",
			),
		},
	}
}
