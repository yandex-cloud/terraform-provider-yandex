package cdn_resource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func DataSourceCDNResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Get information about a Yandex CDN Resource. For more information, see " +
			"[the official documentation](https://yandex.cloud/docs/cdn/concepts/).\n\n" +
			"~> **Note:** One of `resource_id` or `cname` should be specified.",
		MarkdownDescription: "Get information about a Yandex CDN Resource. For more information, see " +
			"[the official documentation](https://yandex.cloud/docs/cdn/concepts/).\n\n" +
			"~> **Note:** One of `resource_id` or `cname` should be specified.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The ID of the CDN resource.",
				MarkdownDescription: "The ID of the CDN resource.",
				Computed:            true,
			},
			"resource_id": schema.StringAttribute{
				Description:         "The ID of a specific CDN resource.",
				MarkdownDescription: "The ID of a specific CDN resource.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("cname")),
				},
			},
			"cname": schema.StringAttribute{
				Description:         "CNAME of the CDN resource. Can be used to find resource by its CNAME.",
				MarkdownDescription: "CNAME of the CDN resource. Can be used to find resource by its CNAME.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("resource_id")),
				},
			},
			"folder_id": schema.StringAttribute{
				Description:         common.ResourceDescriptions["folder_id"],
				MarkdownDescription: common.ResourceDescriptions["folder_id"],
				Optional:            true,
				Computed:            true,
			},
			"provider_type": schema.StringAttribute{
				Description:         "CDN provider type.",
				MarkdownDescription: "CDN provider type.",
				Computed:            true,
			},
			"provider_cname": schema.StringAttribute{
				Description:         "Provider CNAME of the CDN resource.",
				MarkdownDescription: "Provider CNAME of the CDN resource.",
				Computed:            true,
			},
			"active": schema.BoolAttribute{
				Description:         "Flag to create Resource either in active or disabled state.",
				MarkdownDescription: "Flag to create Resource either in active or disabled state.",
				Computed:            true,
			},
			"labels": schema.MapAttribute{
				Description:         common.ResourceDescriptions["labels"],
				MarkdownDescription: common.ResourceDescriptions["labels"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"secondary_hostnames": schema.SetAttribute{
				Description:         "List of secondary hostname strings.",
				MarkdownDescription: "List of secondary hostname strings.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"origin_protocol": schema.StringAttribute{
				Description:         "Origin protocol. Possible values: `http`, `https`, `match` (match client protocol).",
				MarkdownDescription: "Origin protocol. Possible values: `http`, `https`, `match` (match client protocol).",
				Computed:            true,
			},
			"origin_group_id": schema.StringAttribute{
				Description:         "ID of the origin group.",
				MarkdownDescription: "ID of the origin group.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				Description:         "Creation timestamp.",
				MarkdownDescription: "Creation timestamp.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				Description:         "Last update timestamp.",
				MarkdownDescription: "Last update timestamp.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"ssl_certificate": SSLCertificateDataSourceSchema(),
			"options":         CDNOptionsDataSourceSchema(),
		},
	}
}

// SSLCertificateDataSourceSchema returns the schema for SSL certificate block in data source
func SSLCertificateDataSourceSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Description:         "SSL certificate configuration block.",
		MarkdownDescription: "SSL certificate configuration block.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Description: "Type of the SSL certificate. Possible values: " +
						"`not_used` - do not use SSL, `certificate_manager` - certificate from Yandex Certificate Manager, " +
						"`lets_encrypt_gcore` - Let's Encrypt certificate.",
					MarkdownDescription: "Type of the SSL certificate. Possible values: " +
						"`not_used` - do not use SSL, `certificate_manager` - certificate from Yandex Certificate Manager, " +
						"`lets_encrypt_gcore` - Let's Encrypt certificate.",
					Computed: true,
				},
				"status": schema.StringAttribute{
					Description:         "Status of the SSL certificate.",
					MarkdownDescription: "Status of the SSL certificate.",
					Computed:            true,
				},
				"certificate_manager_id": schema.StringAttribute{
					Description:         "ID of certificate from Yandex Certificate Manager (required if type is `certificate_manager`).",
					MarkdownDescription: "ID of certificate from Yandex Certificate Manager (required if type is `certificate_manager`).",
					Computed:            true,
				},
			},
		},
	}
}

// CDNOptionsDataSourceSchema returns the schema for options block in data source
func CDNOptionsDataSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description:         "CDN resource options configuration.",
		MarkdownDescription: "CDN resource options configuration.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				// Boolean options - all computed
				"ignore_query_params": schema.BoolAttribute{
					Description:         "Ignore query parameters.",
					MarkdownDescription: "Ignore query parameters.",
					Computed:            true,
				},
				"slice": schema.BoolAttribute{
					Description:         "Enable slicing.",
					MarkdownDescription: "Enable slicing.",
					Computed:            true,
				},
				"fetched_compressed": schema.BoolAttribute{
					Description:         "Fetch compressed content from origin.",
					MarkdownDescription: "Fetch compressed content from origin.",
					Computed:            true,
				},
				"gzip_on": schema.BoolAttribute{
					Description:         "Enable gzip compression.",
					MarkdownDescription: "Enable gzip compression.",
					Computed:            true,
				},
				"redirect_http_to_https": schema.BoolAttribute{
					Description:         "Redirect HTTP requests to HTTPS.",
					MarkdownDescription: "Redirect HTTP requests to HTTPS.",
					Computed:            true,
				},
				"redirect_https_to_http": schema.BoolAttribute{
					Description:         "Redirect HTTPS requests to HTTP.",
					MarkdownDescription: "Redirect HTTPS requests to HTTP.",
					Computed:            true,
				},
				"forward_host_header": schema.BoolAttribute{
					Description:         "Forward Host header to origin.",
					MarkdownDescription: "Forward Host header to origin.",
					Computed:            true,
				},
				"proxy_cache_methods_set": schema.BoolAttribute{
					Description:         "Enable caching for POST/PUT/PATCH methods.",
					MarkdownDescription: "Enable caching for POST/PUT/PATCH methods.",
					Computed:            true,
				},
				"disable_proxy_force_ranges": schema.BoolAttribute{
					Description:         "Disable proxy force ranges.",
					MarkdownDescription: "Disable proxy force ranges.",
					Computed:            true,
				},
				"ignore_cookie": schema.BoolAttribute{
					Description:         "Ignore Set-Cookie header from origin.",
					MarkdownDescription: "Ignore Set-Cookie header from origin.",
					Computed:            true,
				},
				"enable_ip_url_signing": schema.BoolAttribute{
					Description:         "Enable IP/URL signing.",
					MarkdownDescription: "Enable IP/URL signing.",
					Computed:            true,
				},

				// String options
				"custom_host_header": schema.StringAttribute{
					Description:         "Custom Host header value.",
					MarkdownDescription: "Custom Host header value.",
					Computed:            true,
				},
				"custom_server_name": schema.StringAttribute{
					Description:         "Custom server name for TLS SNI.",
					MarkdownDescription: "Custom server name for TLS SNI.",
					Computed:            true,
				},
				"secure_key": schema.StringAttribute{
					Description:         "Secure key for URL signing.",
					MarkdownDescription: "Secure key for URL signing.",
					Computed:            true,
					Sensitive:           true,
				},

				// List/Set options
				"cache_http_headers": schema.ListAttribute{
					Description:         "HTTP headers to include in cache key.",
					MarkdownDescription: "HTTP headers to include in cache key.",
					ElementType:         types.StringType,
					Computed:            true,
				},
				"query_params_whitelist": schema.ListAttribute{
					Description:         "Whitelist of query parameters to include in cache key.",
					MarkdownDescription: "Whitelist of query parameters to include in cache key.",
					ElementType:         types.StringType,
					Computed:            true,
				},
				"query_params_blacklist": schema.ListAttribute{
					Description:         "Blacklist of query parameters to exclude from cache key.",
					MarkdownDescription: "Blacklist of query parameters to exclude from cache key.",
					ElementType:         types.StringType,
					Computed:            true,
				},
				"cors": schema.ListAttribute{
					Description:         "CORS origins.",
					MarkdownDescription: "CORS origins.",
					ElementType:         types.StringType,
					Computed:            true,
				},
				"allowed_http_methods": schema.ListAttribute{
					Description:         "Allowed HTTP methods.",
					MarkdownDescription: "Allowed HTTP methods.",
					ElementType:         types.StringType,
					Computed:            true,
				},

				// Map options
				"static_response_headers": schema.MapAttribute{
					Description:         "Static response headers.",
					MarkdownDescription: "Static response headers.",
					ElementType:         types.StringType,
					Computed:            true,
				},
				"static_request_headers": schema.MapAttribute{
					Description:         "Static request headers to origin.",
					MarkdownDescription: "Static request headers to origin.",
					ElementType:         types.StringType,
					Computed:            true,
				},
			},
			Blocks: map[string]schema.Block{
				"edge_cache_settings":    EdgeCacheSettingsDataSourceSchema(),
				"browser_cache_settings": BrowserCacheSettingsDataSourceSchema(),
				"ip_address_acl":         IPAddressACLDataSourceSchema(),
				"rewrite":                RewriteDataSourceSchema(),
			},
		},
	}
}

// EdgeCacheSettingsDataSourceSchema returns the schema for edge cache settings
func EdgeCacheSettingsDataSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description:         "Edge cache settings.",
		MarkdownDescription: "Edge cache settings.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Description:         "Enable edge caching.",
					MarkdownDescription: "Enable edge caching.",
					Computed:            true,
				},
				"cache_time": schema.MapAttribute{
					Description:         "Cache time in seconds for different HTTP status codes.",
					MarkdownDescription: "Cache time in seconds for different HTTP status codes.",
					ElementType:         types.Int64Type,
					Computed:            true,
				},
			},
		},
	}
}

// BrowserCacheSettingsDataSourceSchema returns the schema for browser cache settings
func BrowserCacheSettingsDataSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description:         "Browser cache settings.",
		MarkdownDescription: "Browser cache settings.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Description:         "Enable browser caching.",
					MarkdownDescription: "Enable browser caching.",
					Computed:            true,
				},
				"cache_time": schema.Int64Attribute{
					Description:         "Browser cache time in seconds.",
					MarkdownDescription: "Browser cache time in seconds.",
					Computed:            true,
				},
			},
		},
	}
}

// IPAddressACLDataSourceSchema returns the schema for IP address ACL
func IPAddressACLDataSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description:         "IP address ACL settings.",
		MarkdownDescription: "IP address ACL settings.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"policy_type": schema.StringAttribute{
					Description:         "Policy type: `allow` or `deny`.",
					MarkdownDescription: "Policy type: `allow` or `deny`.",
					Computed:            true,
				},
				"excepted_values": schema.ListAttribute{
					Description:         "List of IP addresses or CIDR blocks.",
					MarkdownDescription: "List of IP addresses or CIDR blocks.",
					ElementType:         types.StringType,
					Computed:            true,
				},
			},
		},
	}
}

// RewriteDataSourceSchema returns the schema for rewrite rules
func RewriteDataSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description:         "URL rewrite rules.",
		MarkdownDescription: "URL rewrite rules.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Description:         "Enable rewrite.",
					MarkdownDescription: "Enable rewrite.",
					Computed:            true,
				},
				"body": schema.StringAttribute{
					Description:         "Rewrite pattern.",
					MarkdownDescription: "Rewrite pattern.",
					Computed:            true,
				},
				"flag": schema.StringAttribute{
					Description:         "Rewrite flag: `last`, `break`, `redirect`, `permanent`.",
					MarkdownDescription: "Rewrite flag: `last`, `break`, `redirect`, `permanent`.",
					Computed:            true,
				},
			},
		},
	}
}
