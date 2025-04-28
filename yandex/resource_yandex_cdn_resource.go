package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"

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

func defineYandexCDNResourceBaseSchema() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of [Yandex Cloud CDN Resource](https://yandex.cloud/docs/cdn/concepts/resource).\n\n~> CDN provider must be activated prior usage of CDN resources, either via UI console or via yc cli command: `yc cdn provider activate --folder-id <folder-id> --type gcore`.",

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"cname": {
				Type:        schema.TypeString,
				Description: "CDN endpoint CNAME, must be unique among resources.",
				Computed:    true,
				Optional:    true,

				ValidateFunc: validation.NoZeroValues,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
			"origin_protocol": {
				Type:        schema.TypeString,
				Description: "Protocol of origin resource. `http` or `https`.",
				Optional:    true,
				Default:     "http",
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
			"origin_group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of a specific origin group.",
				Optional:    true,
			},
			"origin_group_name": {
				Type:        schema.TypeString,
				Description: "The name of a specific origin group.",
				Optional:    true,
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
							Type:         schema.TypeString,
							Description:  "SSL certificate type.",
							Required:     true,
							ValidateFunc: validateResourceSSLCertTypeFunc(),
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
			"provider_cname": {
				Type:        schema.TypeString,
				Description: "Provider CNAME of CDN resource, computed value for read and update operations.",
				Computed:    true,
			},
			"options": {
				Type:        schema.TypeList,
				Description: "CDN Resource settings and options to tune CDN edge behavior.",
				Optional:    true,
				Computed:    true,

				MaxItems: 1,

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disable_cache": {
							Type:        schema.TypeBool,
							Description: "Setup a cache status.",
							Optional:    true,
							Computed:    true,
						},
						// TODO: use CDN Provider custom values for response codes.
						"edge_cache_settings": {
							Type:        schema.TypeInt,
							Description: "Content will be cached according to origin cache settings. The value applies for a response with codes 200, 201, 204, 206, 301, 302, 303, 304, 307, 308 if an origin server does not have caching HTTP headers. Responses with other codes will not be cached.",
							Computed:    true,
							Optional:    true,
						},
						"browser_cache_settings": {
							Type:        schema.TypeInt,
							Description: "Set up a cache period for the end-users browser. Content will be cached due to origin settings. If there are no cache settings on your origin, the content will not be cached. The list of HTTP response codes that can be cached in browsers: 200, 201, 204, 206, 301, 302, 303, 304, 307, 308. Other response codes will not be cached. The default value is 4 days.",
							Computed:    true,
							Optional:    true,
						},
						"cache_http_headers": {
							Type:        schema.TypeList,
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
							Computed:    true,
							Optional:    true,
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

							Computed: true,
							Optional: true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"cors": {
							Type:        schema.TypeList,
							Description: "Parameter that lets browsers get access to selected resources from a domain different to a domain from which the request is received.",
							Computed:    true,
							Optional:    true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
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
							Computed:    true,
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
							Computed:    true,
							Optional:    true,
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

										ValidateFunc: validateACLPolicyTypeFunc(),
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
				},
			},
		},
	}
}

func validateResourceSSLCertTypeFunc() schema.SchemaValidateFunc {
	return validation.StringInSlice(
		[]string{
			cdnSSLCertificateTypeNotUsed,
			cdnSSLCertificateTypeCM,
			cdnSSLCertificateTypeLE,
		},
		false,
	)
}

func validateACLPolicyTypeFunc() schema.SchemaValidateFunc {
	return validation.StringInSlice(
		[]string{
			cdnACLPolicyTypeAllow,
			cdnACLPolicyTypeDeny,
		},
		false,
	)
}

func aclPolicyTypeFromString(policyType string) cdn.PolicyType {
	switch policyType {
	case cdnACLPolicyTypeAllow:
		return cdn.PolicyType_POLICY_TYPE_ALLOW
	case cdnACLPolicyTypeDeny:
		return cdn.PolicyType_POLICY_TYPE_DENY
	}

	return cdn.PolicyType_POLICY_TYPE_ALLOW
}

func aclPolicyTypeToString(policyType cdn.PolicyType) string {
	switch policyType {
	case cdn.PolicyType_POLICY_TYPE_ALLOW:
		return cdnACLPolicyTypeAllow
	case cdn.PolicyType_POLICY_TYPE_DENY:
		return cdnACLPolicyTypeDeny
	}

	return cdnACLPolicyTypeAllow
}

func resourceYandexCDNResource() *schema.Resource {
	resourceSchema := defineYandexCDNResourceBaseSchema()

	resourceSchema.Create = resourceYandexCDNResourceCreate
	resourceSchema.Read = resourceYandexCDNResourceRead
	resourceSchema.Update = resourceYandexCDNResourceUpdate
	resourceSchema.Delete = resourceYandexCDNResourceDelete

	resourceSchema.Importer = &schema.ResourceImporter{
		StateContext: schema.ImportStatePassthroughContext,
	}

	resourceSchema.Timeouts = &schema.ResourceTimeout{
		Create: schema.DefaultTimeout(yandexComputeCDNResourceDefaultTimeout),
		Update: schema.DefaultTimeout(yandexComputeCDNResourceDefaultTimeout),
		Delete: schema.DefaultTimeout(yandexComputeCDNResourceDefaultTimeout),
	}

	return resourceSchema
}

func expandCDNResourceOptions(d *schema.ResourceData) *cdn.ResourceOptions {
	_, ok := d.GetOk("options")
	if !ok {
		log.Printf("[DEBUG] empty cdn resource options list")
		return nil
	}

	size := d.Get("options.#").(int)
	if size < 1 {
		log.Printf("[DEBUG] resource options list is empty")
		return nil
	}

	result := &cdn.ResourceOptions{}
	var optionsSet bool

	if rawOption, ok := d.GetOk("options.0.disable_cache"); ok {
		optionsSet = true

		result.DisableCache = &cdn.ResourceOptions_BoolOption{
			Enabled: rawOption.(bool),
			Value:   rawOption.(bool),
		}
	}

	if rawOption, ok := d.GetOk("options.0.edge_cache_settings"); ok {
		optionsSet = true
		result.EdgeCacheSettings = &cdn.ResourceOptions_EdgeCacheSettings{
			Enabled: true,
			ValuesVariant: &cdn.ResourceOptions_EdgeCacheSettings_DefaultValue{
				DefaultValue: int64(rawOption.(int)),
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.browser_cache_settings"); ok {
		optionsSet = true
		result.BrowserCacheSettings = &cdn.ResourceOptions_Int64Option{
			Enabled: true,
			Value:   int64(rawOption.(int)),
		}
	}

	if rawOption, ok := d.GetOk("options.0.cache_http_headers"); ok {
		optionsSet = true

		var values []string
		for _, v := range rawOption.([]interface{}) {
			values = append(values, v.(string))
		}

		if len(values) != 0 {
			result.CacheHttpHeaders = &cdn.ResourceOptions_StringsListOption{
				Enabled: true,
				Value:   values,
			}
		}
	}

	if rawOption, ok := d.GetOk("options.0.ignore_query_params"); ok {
		optionsSet = true

		result.QueryParamsOptions = &cdn.ResourceOptions_QueryParamsOptions{
			QueryParamsVariant: &cdn.ResourceOptions_QueryParamsOptions_IgnoreQueryString{
				IgnoreQueryString: &cdn.ResourceOptions_BoolOption{
					Enabled: rawOption.(bool),
					Value:   rawOption.(bool),
				},
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.query_params_whitelist"); ok {
		optionsSet = true

		var values []string
		for _, v := range rawOption.([]interface{}) {
			values = append(values, v.(string))
		}

		if len(values) != 0 {
			result.QueryParamsOptions = &cdn.ResourceOptions_QueryParamsOptions{
				QueryParamsVariant: &cdn.ResourceOptions_QueryParamsOptions_QueryParamsWhitelist{
					QueryParamsWhitelist: &cdn.ResourceOptions_StringsListOption{
						Enabled: true,
						Value:   values,
					},
				},
			}
		}
	}

	if rawOption, ok := d.GetOk("options.0.query_params_blacklist"); ok {
		optionsSet = true

		var values []string
		for _, v := range rawOption.([]interface{}) {
			values = append(values, v.(string))
		}

		if len(values) != 0 {
			result.QueryParamsOptions = &cdn.ResourceOptions_QueryParamsOptions{
				QueryParamsVariant: &cdn.ResourceOptions_QueryParamsOptions_QueryParamsBlacklist{
					QueryParamsBlacklist: &cdn.ResourceOptions_StringsListOption{
						Enabled: true,
						Value:   values,
					},
				},
			}
		}
	}

	if rawOption, ok := d.GetOk("options.0.slice"); ok {
		optionsSet = true

		result.Slice = &cdn.ResourceOptions_BoolOption{
			Enabled: rawOption.(bool),
			Value:   rawOption.(bool),
		}
	}

	if rawOption, ok := d.GetOk("options.0.fetched_compressed"); ok {
		optionsSet = true

		result.CompressionOptions = &cdn.ResourceOptions_CompressionOptions{
			CompressionVariant: &cdn.ResourceOptions_CompressionOptions_FetchCompressed{
				FetchCompressed: &cdn.ResourceOptions_BoolOption{
					Enabled: rawOption.(bool),
					Value:   rawOption.(bool),
				},
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.gzip_on"); ok {
		optionsSet = true

		result.CompressionOptions = &cdn.ResourceOptions_CompressionOptions{
			CompressionVariant: &cdn.ResourceOptions_CompressionOptions_GzipOn{
				GzipOn: &cdn.ResourceOptions_BoolOption{
					Enabled: rawOption.(bool),
					Value:   rawOption.(bool),
				},
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.redirect_http_to_https"); ok {
		optionsSet = true

		result.RedirectOptions = &cdn.ResourceOptions_RedirectOptions{
			RedirectVariant: &cdn.ResourceOptions_RedirectOptions_RedirectHttpToHttps{
				RedirectHttpToHttps: &cdn.ResourceOptions_BoolOption{
					Enabled: rawOption.(bool),
					Value:   rawOption.(bool),
				},
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.redirect_https_to_http"); ok {
		optionsSet = true

		result.RedirectOptions = &cdn.ResourceOptions_RedirectOptions{
			RedirectVariant: &cdn.ResourceOptions_RedirectOptions_RedirectHttpsToHttp{
				RedirectHttpsToHttp: &cdn.ResourceOptions_BoolOption{
					Enabled: rawOption.(bool),
					Value:   rawOption.(bool),
				},
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.ignore_cookie"); ok {
		optionsSet = true

		result.IgnoreCookie = &cdn.ResourceOptions_BoolOption{
			Enabled: rawOption.(bool),
			Value:   rawOption.(bool),
		}
	}

	makeHostOption := func() *cdn.ResourceOptions_HostOptions {
		if rawOption, ok := d.GetOk("options.0.custom_host_header"); ok && rawOption.(string) != "" {
			optionsSet = true

			return &cdn.ResourceOptions_HostOptions{
				HostVariant: &cdn.ResourceOptions_HostOptions_Host{
					Host: &cdn.ResourceOptions_StringOption{
						Enabled: true,
						Value:   rawOption.(string),
					},
				},
			}
		}

		if rawOption, ok := d.GetOk("options.0.forward_host_header"); ok && rawOption.(bool) {
			optionsSet = true

			return &cdn.ResourceOptions_HostOptions{
				HostVariant: &cdn.ResourceOptions_HostOptions_ForwardHostHeader{
					ForwardHostHeader: &cdn.ResourceOptions_BoolOption{
						Enabled: rawOption.(bool),
						Value:   rawOption.(bool),
					},
				},
			}
		}

		return nil
	}

	result.HostOptions = makeHostOption()

	if rawOption, ok := d.GetOk("options.0.static_response_headers"); ok {
		optionsSet = true

		result.StaticHeaders = &cdn.ResourceOptions_StringsMapOption{
			Enabled: true,
		}

		result.StaticHeaders.Value = make(map[string]string)
		for k, v := range rawOption.(map[string]interface{}) {
			result.StaticHeaders.Value[k] = v.(string)
		}
	}

	if rawOption, ok := d.GetOk("options.0.cors"); ok {
		optionsSet = true

		var values []string
		for _, v := range rawOption.([]interface{}) {
			values = append(values, v.(string))
		}

		if len(values) != 0 {
			result.Cors = &cdn.ResourceOptions_StringsListOption{
				Enabled: true,
				Value:   values,
			}
		}
	}

	if rawOption, ok := d.GetOk("options.0.allowed_http_methods"); ok {
		optionsSet = true

		var values []string
		for _, v := range rawOption.([]interface{}) {
			values = append(values, v.(string))
		}

		if len(values) != 0 {
			result.AllowedHttpMethods = &cdn.ResourceOptions_StringsListOption{
				Enabled: true,
				Value:   values,
			}
		}
	}

	if rawOption, ok := d.GetOk("options.0.proxy_cache_method_set"); ok {
		optionsSet = true

		result.ProxyCacheMethodsSet = &cdn.ResourceOptions_BoolOption{
			Enabled: rawOption.(bool),
			Value:   rawOption.(bool),
		}
	}

	if rawOption, ok := d.GetOk("options.0.disable_proxy_force_ranges"); ok {
		optionsSet = true

		result.DisableProxyForceRanges = &cdn.ResourceOptions_BoolOption{
			Enabled: rawOption.(bool),
			Value:   rawOption.(bool),
		}
	}

	if rawOption, ok := d.GetOk("options.0.static_request_headers"); ok {
		optionsSet = true

		result.StaticRequestHeaders = &cdn.ResourceOptions_StringsMapOption{
			Enabled: true,
		}

		result.StaticRequestHeaders.Value = make(map[string]string)
		for k, v := range rawOption.(map[string]interface{}) {
			result.StaticRequestHeaders.Value[k] = v.(string)
		}
	}

	if rawOption, ok := d.GetOk("options.0.custom_server_name"); ok {
		optionsSet = true

		result.CustomServerName = &cdn.ResourceOptions_StringOption{
			Enabled: true,
			Value:   rawOption.(string),
		}
	}

	if rawOption, ok := d.GetOk("options.0.ignore_cookie"); ok {
		optionsSet = true

		result.IgnoreCookie = &cdn.ResourceOptions_BoolOption{
			Enabled: rawOption.(bool),
			Value:   rawOption.(bool),
		}
	}

	if rawOption, ok := d.GetOk("options.0.secure_key"); ok {
		optionsSet = true

		urlType := cdn.SecureKeyURLType_DISABLE_IP_SIGNING
		if rawUrlType, ok := d.GetOk("options.0.enable_ip_url_signing"); ok && rawUrlType.(bool) {
			urlType = cdn.SecureKeyURLType_ENABLE_IP_SIGNING
		}

		result.SecureKey = &cdn.ResourceOptions_SecureKeyOption{
			Enabled: true,
			Key:     rawOption.(string),
			Type:    urlType,
		}
	}

	if _, ok := d.GetOk("options.0.ip_address_acl"); ok {
		if size := d.Get("options.0.ip_address_acl.#").(int); size > 0 {
			if rawPolicyType, ok := d.GetOk("options.0.ip_address_acl.0.policy_type"); ok {
				optionsSet = true

				var values []string
				if rawExceptedValues, ok := d.GetOk("options.0.ip_address_acl.0.excepted_values"); ok {
					for _, v := range rawExceptedValues.([]interface{}) {
						values = append(values, v.(string))
					}
				}

				result.IpAddressAcl = &cdn.ResourceOptions_IPAddressACLOption{
					Enabled:        true,
					PolicyType:     aclPolicyTypeFromString(rawPolicyType.(string)),
					ExceptedValues: values,
				}
			}
		}
	}

	if !optionsSet {
		return nil
	}

	return result
}

func prepareCDNResourceOptions(d *schema.ResourceData) *cdn.ResourceOptions {
	if options := expandCDNResourceOptions(d); options != nil {
		return options
	}

	return nil
}

func prepareCDNResourceLabels(d *schema.ResourceData) map[string]string {
	labels := make(map[string]string)
	if rawOption, ok := d.GetOk("labels"); ok {
		for k, v := range rawOption.(map[string]interface{}) {
			labels[k] = v.(string)
		}
	}

	return labels
}

func prepareCDNCreateResourceRequest(ctx context.Context, d *schema.ResourceData, meta *Config) (*cdn.CreateResourceRequest, error) {
	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating instance: %s", err)
	}

	prepareResourceOriginVariant := func() (*cdn.CreateResourceRequest_Origin, error) {
		result := &cdn.CreateResourceRequest_Origin{}

		if v, ok := d.GetOk("origin_group_id"); ok {
			groupID := int64(v.(int))

			result.OriginVariant = &cdn.CreateResourceRequest_Origin_OriginGroupId{
				OriginGroupId: groupID,
			}

			return result, nil
		}

		if v, ok := d.GetOk("origin_group_name"); ok {
			groupName := v.(string)

			groupID, err := resolveCDNOriginGroupID(ctx, meta, folderID, groupName)
			if err != nil {
				return nil, err
			}

			result.OriginVariant = &cdn.CreateResourceRequest_Origin_OriginGroupId{
				OriginGroupId: groupID,
			}

			return result, nil
		}

		return nil, nil
	}

	originVariant, err := prepareResourceOriginVariant()
	if err != nil {
		return nil, err
	}

	result := &cdn.CreateResourceRequest{
		FolderId: folderID,
		Cname:    d.Get("cname").(string),

		SecondaryHostnames: prepareCDNResourceSecondaryHostnames(d),

		Origin: originVariant,

		Active: &wrappers.BoolValue{
			Value: d.Get("active").(bool),
		},

		Options: prepareCDNResourceOptions(d),
		Labels:  prepareCDNResourceLabels(d),
	}

	if _, ok := d.GetOk("origin_protocol"); ok {
		result.OriginProtocol = prepareCDNResourceOriginProtocol(d)
	}

	if _, ok := d.GetOk("ssl_certificate"); ok {
		var err error
		if result.SslCertificate, err = prepareCDNResourceNewSSLCertificate(d); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func resourceYandexCDNResourceCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Crating CDN Resource %q", d.Get("cname").(string))

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	request, err := prepareCDNCreateResourceRequest(ctx, d, config)
	if err != nil {
		return err
	}

	operation, err := config.sdk.WrapOperation(
		config.sdk.CDN().Resource().Create(ctx, request),
	)

	if err != nil {
		return fmt.Errorf("error while requesting API to create CDN Resource: %s", err)
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return fmt.Errorf("error while obtaining response metadata for create CDN Resource operation: %s", err)
	}

	pm, ok := protoMetadata.(*cdn.CreateResourceMetadata)
	if !ok {
		return fmt.Errorf("resource metadata type mismatch")
	}

	d.SetId(pm.ResourceId)

	err = operation.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while requesting API to create CDN Resource: %s", err)
	}

	if _, err = operation.Response(); err != nil {
		return err
	}

	return resourceYandexCDNResourceRead(d, meta)
}

func flattenYandexCDNResourceOptions(options *cdn.ResourceOptions) []map[string]interface{} {
	if options == nil {
		log.Printf("[DEBUG] empty cdn resource options set")
		return nil
	}

	item := make(map[string]interface{})

	setIfEnabled := func(optionName string, enabled bool, value interface{}) {
		if !enabled {
			return
		}

		item[optionName] = value
	}

	if options.DisableCache != nil {
		setIfEnabled("disable_cache", options.DisableCache.Enabled, options.DisableCache.Value)
	}

	if options.EdgeCacheSettings != nil && options.EdgeCacheSettings.Enabled {
		switch v := options.EdgeCacheSettings.ValuesVariant.(type) {
		case *cdn.ResourceOptions_EdgeCacheSettings_DefaultValue:
			item["edge_cache_settings"] = v.DefaultValue
		default:
			log.Printf("[WARN] custom timings for cdn edge_cache_setting option are not implemented")
		}
	}

	if options.BrowserCacheSettings != nil {
		setIfEnabled("browser_cache_settings", options.BrowserCacheSettings.Enabled, options.BrowserCacheSettings.Value)
	}

	if options.CacheHttpHeaders != nil {
		setIfEnabled("cache_http_headers", options.CacheHttpHeaders.Enabled, options.CacheHttpHeaders.Value)
	}

	if options.QueryParamsOptions != nil {
		switch val := options.QueryParamsOptions.QueryParamsVariant.(type) {
		case *cdn.ResourceOptions_QueryParamsOptions_IgnoreQueryString:
			setIfEnabled("ignore_query_params", val.IgnoreQueryString.Enabled, val.IgnoreQueryString.Value)
		case *cdn.ResourceOptions_QueryParamsOptions_QueryParamsBlacklist:
			setIfEnabled("query_params_blacklist", val.QueryParamsBlacklist.Enabled, val.QueryParamsBlacklist.Value)
		case *cdn.ResourceOptions_QueryParamsOptions_QueryParamsWhitelist:
			setIfEnabled("query_params_whitelist", val.QueryParamsWhitelist.Enabled, val.QueryParamsWhitelist.Value)
		}
	}

	if options.Slice != nil {
		setIfEnabled("slice", options.Slice.Enabled, options.Slice.Value)
	}

	if options.CompressionOptions != nil {
		switch val := options.CompressionOptions.CompressionVariant.(type) {
		case *cdn.ResourceOptions_CompressionOptions_FetchCompressed:
			setIfEnabled("fetched_compressed", val.FetchCompressed.Enabled, val.FetchCompressed.Value)
		case *cdn.ResourceOptions_CompressionOptions_GzipOn:
			setIfEnabled("gzip_on", val.GzipOn.Enabled, val.GzipOn.Value)
		}
	}

	if options.RedirectOptions != nil {
		switch val := options.RedirectOptions.RedirectVariant.(type) {
		case *cdn.ResourceOptions_RedirectOptions_RedirectHttpToHttps:
			setIfEnabled("redirect_http_to_https", val.RedirectHttpToHttps.Enabled, val.RedirectHttpToHttps.Value)
		case *cdn.ResourceOptions_RedirectOptions_RedirectHttpsToHttp:
			setIfEnabled("redirect_https_to_http", val.RedirectHttpsToHttp.Enabled, val.RedirectHttpsToHttp.Value)
		}
	}

	if options.HostOptions != nil {
		switch val := options.HostOptions.HostVariant.(type) {
		case *cdn.ResourceOptions_HostOptions_ForwardHostHeader:
			setIfEnabled("forward_host_header", val.ForwardHostHeader.Enabled, val.ForwardHostHeader.Value)
		case *cdn.ResourceOptions_HostOptions_Host:
			setIfEnabled("custom_host_header", val.Host.Enabled, val.Host.Value)
		}
	}

	if options.Cors != nil {
		setIfEnabled("cors", options.Cors.Enabled, options.Cors.Value)
	}

	if options.AllowedHttpMethods != nil {
		setIfEnabled("allowed_http_methods", options.AllowedHttpMethods.Enabled, options.AllowedHttpMethods.Value)
	}

	if options.ProxyCacheMethodsSet != nil {
		setIfEnabled("proxy_cache_methods_set", options.ProxyCacheMethodsSet.Enabled, options.ProxyCacheMethodsSet.Value)
	}

	if options.DisableProxyForceRanges != nil {
		setIfEnabled("disable_proxy_force_ranges", options.DisableProxyForceRanges.Enabled, options.DisableProxyForceRanges.Value)
	}

	if options.StaticHeaders != nil {
		setIfEnabled("static_response_headers", options.StaticHeaders.Enabled, options.StaticHeaders.Value)
	}

	if options.StaticRequestHeaders != nil {
		setIfEnabled("static_request_headers", options.StaticRequestHeaders.Enabled, options.StaticRequestHeaders.Value)
	}

	if options.CustomServerName != nil {
		setIfEnabled("custom_server_name", options.CustomServerName.Enabled, options.CustomServerName.Value)
	}

	if options.IgnoreCookie != nil {
		setIfEnabled("ignore_cookie", options.IgnoreCookie.Enabled, options.IgnoreCookie.Value)
	}

	if options.SecureKey != nil {
		setIfEnabled("secure_key", options.SecureKey.Enabled, options.SecureKey.Key)

		if options.SecureKey.Type == cdn.SecureKeyURLType_ENABLE_IP_SIGNING {
			setIfEnabled("enable_ip_url_signing", options.SecureKey.Enabled, true)
		} else {
			setIfEnabled("enable_ip_url_signing", options.SecureKey.Enabled, false)
		}
	}

	if options.IpAddressAcl != nil {
		ipAddrACL := make(map[string]interface{})
		ipAddrACL["policy_type"] = aclPolicyTypeToString(options.IpAddressAcl.PolicyType)
		ipAddrACL["excepted_values"] = options.IpAddressAcl.ExceptedValues

		setIfEnabled("ip_address_acl", options.IpAddressAcl.Enabled, []map[string]interface{}{ipAddrACL})
	}

	return []map[string]interface{}{
		item,
	}
}

func flattenYandexCDNResource(d *schema.ResourceData, resource *cdn.Resource) error {
	d.SetId(resource.Id)

	_ = d.Set("folder_id", resource.FolderId)
	_ = d.Set("cname", resource.Cname)
	_ = d.Set("labels", resource.Labels)

	_ = d.Set("created_at", getTimestamp(resource.CreatedAt))
	_ = d.Set("updated_at", getTimestamp(resource.UpdatedAt))

	_ = d.Set("active", resource.Active)

	if err := flattenYandexCDNResourceSecondaryNames(d, resource.SecondaryHostnames); err != nil {
		return err
	}

	flattenYandexCDNResourceOriginGroup(d, resource)

	if err := flattenYandexCDNResourceOriginProtocol(d, resource.OriginProtocol); err != nil {
		return err
	}

	if err := flattenYandexCDNResourceSSLCertificate(d, resource.SslCertificate); err != nil {
		return err
	}

	return nil
}

func flattenYandexCDNResourceSecondaryNames(d *schema.ResourceData, secondaryHostnames []string) error {
	if len(secondaryHostnames) == 0 {
		return nil
	}

	var result []interface{}
	for i := range secondaryHostnames {
		result = append(result, secondaryHostnames[i])
	}

	return d.Set("secondary_hostnames", result)
}

func flattenYandexCDNResourceOriginGroup(d *schema.ResourceData, resource *cdn.Resource) {
	if _, ok := d.GetOk("origin_group_name"); ok {
		_ = d.Set("origin_group_name", resource.OriginGroupName)
	}

	if _, ok := d.GetOk("origin_group_id"); ok {
		_ = d.Set("origin_group_id", resource.OriginGroupId)
	}
}

func flattenYandexCDNResourceOriginProtocol(d *schema.ResourceData, protocol cdn.OriginProtocol) error {
	switch protocol {
	case cdn.OriginProtocol_HTTP:
		_ = d.Set("origin_protocol", "http")
	case cdn.OriginProtocol_HTTPS:
		_ = d.Set("origin_protocol", "https")
	case cdn.OriginProtocol_MATCH:
		_ = d.Set("origin_protocol", "match")
	default:
		return fmt.Errorf("unexpected origin protocol value in API response")
	}
	return nil
}

func flattenYandexCDNResourceSSLCertificate(d *schema.ResourceData, cert *cdn.SSLCertificate) error {
	if cert == nil {
		return nil
	}

	result := make(map[string]interface{})

	var typeStr string
	switch cert.Type {
	case cdn.SSLCertificateType_DONT_USE:
		typeStr = cdnSSLCertificateTypeNotUsed
	case cdn.SSLCertificateType_LETS_ENCRYPT_GCORE:
		typeStr = cdnSSLCertificateTypeLE
	case cdn.SSLCertificateType_CM:
		typeStr = cdnSSLCertificateTypeCM
	default:
		return fmt.Errorf("unexpected ssl certificate type in API response")
	}
	result["type"] = typeStr

	var statusStr string
	switch cert.Status {
	case cdn.SSLCertificateStatus_READY:
		statusStr = cdnSSLCertificateStatusReady
	case cdn.SSLCertificateStatus_CREATING:
		statusStr = cdnSSLCertificateStatusCreating
	}
	result["status"] = statusStr

	if cert.Type == cdn.SSLCertificateType_CM {
		if cert.Data == nil || cert.Data.GetCm() == nil {
			return fmt.Errorf("certificate manager data is absent in API response")
		}
		result["certificate_manager_id"] = cert.Data.GetCm().GetId()
	}

	return d.Set("ssl_certificate", []interface{}{result})
}

func resourceYandexCDNResourceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Reading CDN Resource: %q", d.Id())

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	resource, err := config.sdk.CDN().Resource().Get(ctx, &cdn.GetResourceRequest{
		ResourceId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("cdn resource %q", d.Id()))
	}

	log.Printf("[DEBUG] Completed Reading CDN Resource %q", d.Id())

	if err = flattenYandexCDNResource(d, resource); err != nil {
		return err
	}

	if err = d.Set("options", flattenYandexCDNResourceOptions(resource.Options)); err != nil {
		return err
	}

	cname, err := config.sdk.CDN().Resource().GetProviderCName(ctx, &cdn.GetProviderCNameRequest{
		FolderId: resource.FolderId,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("get provider cname: cdn resource %q, folder id %q", d.Id(), resource.FolderId))
	}

	if err = d.Set("provider_cname", cname.Cname); err != nil {
		return err
	}

	return nil
}

func prepareCDNUpdateResourceRequest(ctx context.Context, d *schema.ResourceData, config *Config) (*cdn.UpdateResourceRequest, error) {
	request := &cdn.UpdateResourceRequest{
		ResourceId: d.Id(),
	}

	if d.HasChange("origin_group_id") {
		groupID := d.Get("origin_group_id").(int)
		if groupID > 0 {
			request.OriginGroupId = &wrappers.Int64Value{
				Value: int64(groupID),
			}
		}
	}

	if d.HasChange("origin_group_name") {
		groupName := d.Get("origin_group_name").(string)
		if groupName != "" {
			folderID, err := getFolderID(d, config)
			if err != nil {
				return nil, fmt.Errorf("error getting folder ID while creating instance: %s", err)
			}

			groupID, err := resolveCDNOriginGroupID(ctx, config, folderID, groupName)
			if err != nil {
				return nil, err
			}

			request.OriginGroupId = &wrappers.Int64Value{
				Value: groupID,
			}
		}
	}

	if d.HasChange("secondary_hostnames") {
		request.SecondaryHostnames = prepareCDNResourceSecondaryHostnames(d)
	}

	if d.HasChange("origin_protocol") {
		request.OriginProtocol = prepareCDNResourceOriginProtocol(d)
	}

	if d.HasChange("active") {
		request.Active = &wrappers.BoolValue{
			Value: d.Get("active").(bool),
		}
	}

	if d.HasChange("ssl_certificate") {
		var err error
		if request.SslCertificate, err = prepareCDNResourceNewSSLCertificate(d); err != nil {
			return nil, err
		}
	}

	if d.HasChange("options") {
		request.Options = prepareCDNResourceOptions(d)
	}

	if d.HasChange("labels") {
		request.Labels = prepareCDNResourceLabels(d)
		if len(request.Labels) == 0 {
			request.RemoveLabels = true
		}
	}

	return request, nil
}

func resourceYandexCDNResourceUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Updating CDN Resource %q", d.Id())

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	request, err := prepareCDNUpdateResourceRequest(ctx, d, config)
	if err != nil {
		return err
	}

	operation, err := config.sdk.WrapOperation(config.sdk.CDN().Resource().Update(ctx, request))
	if err != nil {
		return err
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return fmt.Errorf("error while obtaining response metadate for CDN Resource update: %s", err)
	}

	pm, ok := protoMetadata.(*cdn.UpdateResourceMetadata)
	if !ok {
		return fmt.Errorf("cdn resource metadata type mismatch")
	}

	d.SetId(pm.ResourceId)

	err = operation.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while requesting API to update CDN Resource: %s", err)
	}

	if _, err := operation.Response(); err != nil {
		return err
	}

	log.Printf("[DEBUG] Completed updating CDN Resource %q", d.Id())

	resource, err := config.sdk.CDN().Resource().Get(ctx, &cdn.GetResourceRequest{
		ResourceId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("cdn resource %q", d.Id()))
	}

	cname, err := config.sdk.CDN().Resource().GetProviderCName(ctx, &cdn.GetProviderCNameRequest{
		FolderId: resource.FolderId,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("get provider cname: cdn resource %q, folder id %q", d.Id(), resource.FolderId))
	}

	if err = d.Set("provider_cname", cname.Cname); err != nil {
		return err
	}

	return resourceYandexCDNResourceRead(d, meta)
}

func resourceYandexCDNResourceDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting CDN Resource %q", d.Id())

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	operation, err := config.sdk.WrapOperation(
		config.sdk.CDN().Resource().Delete(ctx, &cdn.DeleteResourceRequest{
			ResourceId: d.Id(),
		}),
	)

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("CDN Resource ID: %q", d.Id()))
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return fmt.Errorf("error while obtaining response metadata for CDN Resource: %s", err)
	}

	pm, ok := protoMetadata.(*cdn.DeleteResourceMetadata)
	if !ok {
		return fmt.Errorf("resource metadata type mismatch")
	}

	log.Printf("[DEBUG] Waiting Deleting of CDN Resource operation completion %q", d.Id())

	if err = operation.Wait(ctx); err != nil {
		return err
	}

	if _, err := operation.Response(); err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting of CDN Resource %q: %#v", d.Id(), pm.ResourceId)

	return nil
}
