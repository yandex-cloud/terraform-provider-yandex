package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"

	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexComputeCDNResourceDefaultTimeout = 5 * time.Minute
)

const (
	cdnSSLCertificateTypeNotUsed = "not_used"
	cdnSSLCertificateTypeLE      = "lets_encrypt_gcore"
	cdnSSLCertificateTypeCM      = "certificate_manager"
)

const (
	cdnSSLCertificateStatusReady    = "ready"
	cdnSSLCertificateStatusCreating = "creating"
)

const (
	cdnACLPolicyTypeAllow = "allow"
	cdnACLPolicyTypeDeny  = "deny"
)

const (
	cdnProviderOurcdn = "ourcdn"
	cdnProviderGcore  = "gcore"
)

func resourceYandexCDNResourceSchema() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of [Yandex Cloud CDN Resource](https://yandex.cloud/docs/cdn/concepts/resource).",

		SchemaVersion: 0,
		CustomizeDiff: func(ctx context.Context, rd *schema.ResourceDiff, v any) error {
			if err := customizeDiffCDN_EdgeCacheSettings(ctx, rd, v); err != nil {
				return err
			}
			if err := customizeDiffCDN_RewriteFlag(ctx, rd, v); err != nil {
				return err
			}
			return nil
		},

		Schema: map[string]*schema.Schema{
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"cname": {
				Type:        schema.TypeString,
				Description: "CDN endpoint CNAME, must be unique among resources.",
				Required:    true,
				ForceNew:    true,

				ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Last update timestamp. Computed value for read and update operations.",
				Optional:    true,
				Computed:    true,
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Flag to create Resource either in active or disabled state. `True` - the content from CDN is available to clients.",
				Default:     true,
				Optional:    true,
			},
			"secondary_hostnames": {
				Type:        schema.TypeSet,
				Description: "List of secondary hostname strings.",
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"origin_group_id": {
				Type:        schema.TypeString,
				Description: "The ID of a specific origin group.",
				Optional:    true,
			},
			"origin_group_name": {
				Type:        schema.TypeString,
				Description: "The name of a specific origin group.",
				Optional:    true,
			},
			"origin_protocol": {
				Type:        schema.TypeString,
				Description: "Protocol of origin resource. `http` or `https`.",
				Optional:    true,
				Default:     "http",
			},

			"ssl_certificate": {
				Type:        schema.TypeSet,
				Description: "SSL certificate of CDN resource.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:             schema.TypeString,
							Description:      "SSL certificate type.",
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validateCDNResourceSSLCertType),
						},
						"status": {
							Type:        schema.TypeString,
							Description: "SSL certificate status.",
							Computed:    true,
						},
						"certificate_manager_id": {
							Type:        schema.TypeString,
							Description: "Certificate Manager ID.",
							Optional:    true,
						},
					},
				},
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"provider_type": {
				Type:             schema.TypeString,
				Description:      `CDN provider is a content delivery service provider. Possible values: "ourcdn" (default) or "gcore"`,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validateCDNProvider),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					return newValue == ""
				},
			},
			"provider_cname": {
				Type:        schema.TypeString,
				Description: "Provider CNAME of CDN resource, computed value for read and update operations.",
				Computed:    true,
			},

			"shielding": {
				Type: schema.TypeString,
				Description: "Shielding is a Cloud CDN feature that helps reduce the load on content origins from CDN servers.\n" +
					"Specify location id to enable shielding. See https://yandex.cloud/en/docs/cdn/operations/resources/enable-shielding",
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validateCDNShieldingLocation),
			},
			"options": {
				Type:        schema.TypeList,
				Description: "CDN Resource settings and options to tune CDN edge behavior.",
				Optional:    true,
				Computed:    true,

				MaxItems: 1,

				Elem: resourceYandexCDNResourceSchema_Options(),
			},
		},
	}
}

func resourceYandexCDNResourceSchema_Options() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"disable_cache": {
				Type:        schema.TypeBool,
				Deprecated:  "This attribute does not affect anything. You can safely delete it.",
				Description: "Setup a cache status.",
				Optional:    true,
				Computed:    true,
			},

			"edge_cache_settings": {
				Type:          schema.TypeInt,
				Description:   "Content will be cached according to origin cache settings. The value applies for a response with codes 200, 201, 204, 206, 301, 302, 303, 304, 307, 308 if an origin server does not have caching HTTP headers. Responses with other codes will not be cached.",
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"options.0.edge_cache_settings_codes"},
			},
			"edge_cache_settings_codes": {
				Type:          schema.TypeList,
				Description:   "Set the cache expiration time for CDN servers",
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"options.0.edge_cache_settings"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type: schema.TypeInt,
							Description: "Caching time for a response with codes 200, 206, 301, 302. " +
								"Responses with codes 4xx, 5xx will not be cached. Use `0` disable to caching. " +
								"Use `custom_values` field to specify a custom caching time for a response with specific codes.",
							Optional:     true,
							AtLeastOneOf: []string{"options.0.edge_cache_settings_codes.0.value", "options.0.edge_cache_settings_codes.0.custom_values"},
						},
						"custom_values": {
							Type: schema.TypeMap,
							Description: "Caching time for a response with specific codes. These settings have a higher priority than the `value` field. " +
								"Response code (`304`, `404` for example). Use `any` to specify caching time for all response codes. ",
							Optional: true,
							Elem: &schema.Schema{
								Type:        schema.TypeInt,
								Description: "Caching time in seconds. Use `0` to disable caching for a specific response code.",
							},
							AtLeastOneOf: []string{"options.0.edge_cache_settings_codes.0.value", "options.0.edge_cache_settings_codes.0.custom_values"},
						},
					},
				},
			},

			"browser_cache_settings": {
				Type:        schema.TypeInt,
				Description: "Set up a cache period for the end-users browser. Content will be cached due to origin settings. If there are no cache settings on your origin, the content will not be cached. The list of HTTP response codes that can be cached in browsers: 200, 201, 204, 206, 301, 302, 303, 304, 307, 308. Other response codes will not be cached. The default value is 4 days.",
				Optional:    true,
				Default:     0,
			},
			"cache_http_headers": {
				Type:        schema.TypeList,
				Deprecated:  "This attribute does not affect anything. You can safely delete it.",
				Description: "List HTTP headers that must be included in responses to clients.",
				Computed:    true,
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"ignore_query_params": {
				Type:        schema.TypeBool,
				Description: "Files with different query parameters are cached as objects with the same key regardless of the parameter value. selected by default.",
				Computed:    true,
				Optional:    true,
			},
			"query_params_whitelist": {
				Type:        schema.TypeList,
				Description: "Files with the specified query parameters are cached as objects with different keys, files with other parameters are cached as objects with the same key.",
				Computed:    true,
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"query_params_blacklist": {
				Type:        schema.TypeList,
				Description: "Files with the specified query parameters are cached as objects with the same key, files with other parameters are cached as objects with different keys.",
				Computed:    true,
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"slice": {
				Type:        schema.TypeBool,
				Description: "Files larger than 10 MB will be requested and cached in parts (no larger than 10 MB each part). It reduces time to first byte. The origin must support HTTP Range requests.",
				Optional:    true,
				Default:     false,
			},

			"fetched_compressed": {
				Type:        schema.TypeBool,
				Description: "Option helps you to reduce the bandwidth between origin and CDN servers. Also, content delivery speed becomes higher because of reducing the time for compressing files in a CDN.",
				Computed:    true,
				Optional:    true,
			},
			"gzip_on": {
				Type:        schema.TypeBool,
				Description: "GZip compression at CDN servers reduces file size by 70% and can be as high as 90%.",
				Computed:    true,
				Optional:    true,
			},
			// TODO: brotli

			"redirect_http_to_https": {
				Type:        schema.TypeBool,
				Description: "Set up a redirect from HTTP to HTTPS.",
				Computed:    true,
				Optional:    true,
			},
			"redirect_https_to_http": {
				Type:        schema.TypeBool,
				Description: "Set up a redirect from HTTPS to HTTP.",
				Computed:    true,
				Optional:    true,
			},

			"custom_host_header": {
				Type:        schema.TypeString,
				Description: "Custom value for the Host header. Your server must be able to process requests with the chosen header.",
				Computed:    true,
				Optional:    true,
			},
			"forward_host_header": {
				Type:        schema.TypeBool,
				Description: "Choose the Forward Host header option if is important to send in the request to the Origin the same Host header as was sent in the request to CDN server.",
				Computed:    true,
				Optional:    true,
			},

			"static_response_headers": {
				Type:        schema.TypeMap,
				Description: "Set up a static response header. The header name must be lowercase.",
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cors": {
				Type:        schema.TypeList,
				Description: "Parameter that lets browsers get access to selected resources from a domain different to a domain from which the request is received.",
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"stale": {
				Type: schema.TypeList,
				Description: "List of errors which instruct CDN servers to serve stale content to clients. " +
					"Possible values: `error`, `http_403`, `http_404`, `http_429`, `http_500`, `http_502`, `http_503`, `http_504`, `invalid_header`, `timeout`, `updating`.",
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validateCDNStale),
				},
			},
			"allowed_http_methods": {
				Type:        schema.TypeList,
				Description: "HTTP methods for your CDN content. By default the following methods are allowed: GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS. In case some methods are not allowed to the user, they will get the 405 (Method Not Allowed) response. If the method is not supported, the user gets the 501 (Not Implemented) response.",
				Computed:    true,
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"proxy_cache_methods_set": {
				Type:        schema.TypeBool,
				Description: "Allows caching for GET, HEAD and POST requests.",
				Computed:    true,
				Optional:    true,
			},
			"disable_proxy_force_ranges": {
				Type:        schema.TypeBool,
				Description: "Disabling proxy force ranges.",
				Computed:    true,
				Optional:    true,
			},
			"static_request_headers": {
				Type:        schema.TypeMap,
				Description: "Set up custom headers that CDN servers will send in requests to origins.",
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"custom_server_name": {
				Type:        schema.TypeString,
				Description: "Wildcard additional CNAME. If a resource has a wildcard additional CNAME, you can use your own certificate for content delivery via HTTPS.",
				Computed:    true,
				Optional:    true,
			},
			"ignore_cookie": {
				Type:        schema.TypeBool,
				Description: "Set for ignoring cookie.",
				Optional:    true,
				Default:     true,
			},

			"rewrite_pattern": {
				Type: schema.TypeString,
				Description: "An option for changing or redirecting query paths. " +
					"The value must have the following format: `<source path> <destination path>`, where both paths are regular expressions which use at least one group. E.g., `/foo/(.*) /bar/$1`.",
				Optional: true,
			},
			"rewrite_flag": {
				Type: schema.TypeString,
				Description: "Defines flag for the Rewrite option (default: `BREAK`).\n" +
					"`LAST` - Stops processing of the current set of ngx_http_rewrite_module directives and starts a search for a new location matching changed URI.\n" +
					"`BREAK` - Stops processing of the current set of the Rewrite option.\n" +
					"`REDIRECT` - Returns a temporary redirect with the 302 code; It is used when a replacement string does not start with \"http://\", \"https://\", or \"$scheme\"\n" +
					"`PERMANENT` - Returns a permanent redirect with the 301 code.",
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validateCDNRewriteFlag),
			},

			"secure_key": {
				Type:        schema.TypeString,
				Description: "Set secure key for url encoding to protect contect and limit access by IP addresses and time limits.",
				Computed:    true,
				Optional:    true,
			},
			"enable_ip_url_signing": {
				Type:         schema.TypeBool,
				Description:  "Enable access limiting by IP addresses, option available only with setting secure_key.",
				Computed:     true,
				Optional:     true,
				RequiredWith: []string{"options.0.secure_key"},
			},
			"ip_address_acl": {
				Type:        schema.TypeList,
				Description: "IP address access control list. The list of specified IP addresses to be allowed or denied depending on acl policy type.",
				Optional:    true,
				Computed:    true,

				MaxItems: 1,

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy_type": {
							Type:        schema.TypeString,
							Description: "The policy type for ACL. One of `allow` or `deny` values.",
							Optional:    true,
							Computed:    true,

							ValidateDiagFunc: validation.ToDiagFunc(validateCDNResourceACLPolicyType),
						},
						"excepted_values": {
							Type:        schema.TypeList,
							Description: "The list of specified IP addresses to be allowed or denied depending on acl policy type.",
							Optional:    true,
							Computed:    true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceYandexCDNResource() *schema.Resource {
	resource := &cdnResource{}

	resourceSchema := resourceYandexCDNResourceSchema()

	resourceSchema.ReadContext = resource.Read
	resourceSchema.CreateContext = resource.Create
	resourceSchema.UpdateContext = resource.Update
	resourceSchema.DeleteContext = resource.Delete

	resourceSchema.Importer = &schema.ResourceImporter{
		StateContext: schema.ImportStatePassthroughContext,
	}

	resourceSchema.Timeouts = &schema.ResourceTimeout{
		Create: schema.DefaultTimeout(yandexComputeCDNResourceDefaultTimeout),
		Read:   schema.DefaultTimeout(yandexComputeCDNResourceDefaultTimeout),
		Update: schema.DefaultTimeout(yandexComputeCDNResourceDefaultTimeout),
		Delete: schema.DefaultTimeout(yandexComputeCDNResourceDefaultTimeout),
	}

	return resourceSchema
}

type cdnResource struct{}

func (c *cdnResource) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Crating CDN Resource %q", d.Get("cname").(string))

	request, err := prepareCDNCreateResourceRequest(ctx, d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	// check whether origin_group.provider matches resource.provider or not
	if err := cdnCheckProviderMatching(ctx, request, config); err != nil {
		return diag.FromErr(err)
	}

	operation, err := config.sdk.WrapOperation(
		config.sdk.CDN().Resource().Create(ctx, request),
	)

	if err != nil {
		return diag.Errorf("error while requesting API to create CDN Resource: %s", err)
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return diag.Errorf("error while obtaining response metadata for create CDN Resource operation: %s", err)
	}

	pm, ok := protoMetadata.(*cdn.CreateResourceMetadata)
	if !ok {
		return diag.Errorf("resource metadata type mismatch")
	}

	d.SetId(pm.ResourceId)

	err = operation.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while requesting API to create CDN Resource: %s", err)
	}

	if _, err = operation.Response(); err != nil {
		return diag.FromErr(err)
	}

	if err := updateShielding(ctx, d, config); err != nil {
		return diag.FromErr(fmt.Errorf("updating shielding: %w", err))
	}

	return c.Read(ctx, d, meta)
}

func (*cdnResource) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Reading CDN Resource: %q", d.Id())

	resource, err := config.sdk.CDN().Resource().Get(ctx, &cdn.GetResourceRequest{
		ResourceId: d.Id(),
	})
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("cdn resource %q", d.Id())))
	}

	log.Printf("[DEBUG] Completed Reading CDN Resource %q", d.Id())

	shielding, err := getShieldingLocation(ctx, d.Id(), config.sdk)
	if err != nil {
		return diag.FromErr(fmt.Errorf("reading shielding: %w", err))
	}

	res, err := flattenCDNResource(resource, shielding)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("origin_group_name"); !ok {
		delete(res, "origin_group_name")
	}
	if _, ok := d.GetOk("origin_group_id"); !ok {
		delete(res, "origin_group_id")
	}

	for k, v := range res {
		if err := d.Set(k, v); err != nil {
			return diag.Errorf("error setting %s for CDN Resource (%s): %s", k, d.Id(), err)
		}
	}
	d.SetId(resource.Id)
	return nil
}

func (c *cdnResource) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Updating CDN Resource %q", d.Id())

	request, err := prepareCDNUpdateResourceRequest(ctx, d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	operation, err := config.sdk.WrapOperation(config.sdk.CDN().Resource().Update(ctx, request))
	if err != nil {
		return diag.FromErr(err)
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return diag.Errorf("error while obtaining response metadate for CDN Resource update: %s", err)
	}

	pm, ok := protoMetadata.(*cdn.UpdateResourceMetadata)
	if !ok {
		return diag.Errorf("cdn resource metadata type mismatch")
	}

	if err = operation.Wait(ctx); err != nil {
		return diag.Errorf("error while requesting API to update CDN Resource: %s", err)
	}

	if _, err := operation.Response(); err != nil {
		return diag.FromErr(err)
	}

	if err := updateShielding(ctx, d, config); err != nil {
		return diag.FromErr(fmt.Errorf("updating shielding: %w", err))
	}

	d.SetId(pm.ResourceId)

	log.Printf("[DEBUG] Completed updating CDN Resource %q", d.Id())
	return c.Read(ctx, d, meta)
}

func (c *cdnResource) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting CDN Resource %q", d.Id())

	operation, err := config.sdk.WrapOperation(
		config.sdk.CDN().Resource().Delete(ctx, &cdn.DeleteResourceRequest{
			ResourceId: d.Id(),
		}),
	)

	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("CDN Resource ID: %q", d.Id())))
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return diag.Errorf("error while obtaining response metadata for CDN Resource: %s", err)
	}

	pm, ok := protoMetadata.(*cdn.DeleteResourceMetadata)
	if !ok {
		return diag.Errorf("resource metadata type mismatch")
	}

	log.Printf("[DEBUG] Waiting Deleting of CDN Resource operation completion %q", d.Id())

	if err = operation.Wait(ctx); err != nil {
		return diag.FromErr(err)
	}

	if _, err := operation.Response(); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting of CDN Resource %q: %#v", d.Id(), pm.ResourceId)
	return nil
}
