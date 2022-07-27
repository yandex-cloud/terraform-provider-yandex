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

func defineYandexCDNResourceBaseSchema() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"cname": {
				Type: schema.TypeString,

				Computed: true,
				Optional: true,

				ValidateFunc: validation.NoZeroValues,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"secondary_hostnames": {
				Type:     schema.TypeSet,
				Optional: true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"origin_protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "http",
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"origin_group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"origin_group_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssl_certificate": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateResourceSSLCertTypeFunc(),
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"certificate_manager_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"options": {
				Type: schema.TypeList,

				Optional: true,
				Computed: true,

				MaxItems: 1,

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disable_cache": {
							Type: schema.TypeBool,

							Optional: true,
							Computed: true,
						},
						// TODO: use CDN Provider custom values for response codes.
						"edge_cache_settings": {
							Type: schema.TypeInt,

							Computed: true,
							Optional: true,
						},
						"browser_cache_settings": {
							Type: schema.TypeInt,

							Computed: true,
							Optional: true,
						},
						"cache_http_headers": {
							Type: schema.TypeList,

							Computed: true,
							Optional: true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ignore_query_params": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
						},
						"query_params_whitelist": {
							Type: schema.TypeList,

							Computed: true,
							Optional: true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"query_params_blacklist": {
							Type: schema.TypeList,

							Computed: true,
							Optional: true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"slice": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
						},
						"fetched_compressed": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
						},
						"gzip_on": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
						},
						"redirect_http_to_https": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
						},
						"redirect_https_to_http": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
						},
						"custom_host_header": {
							Type: schema.TypeString,

							Computed: true,
							Optional: true,
						},
						"forward_host_header": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
						},
						"static_response_headers": {
							Type: schema.TypeMap,

							Computed: true,
							Optional: true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"cors": {
							Type: schema.TypeList,

							Computed: true,
							Optional: true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"allowed_http_methods": {
							Type: schema.TypeList,

							Computed: true,
							Optional: true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"proxy_cache_methods_set": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
						},
						"disable_proxy_force_ranges": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
						},
						"static_request_headers": {
							Type: schema.TypeList,

							Computed: true,
							Optional: true,

							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"custom_server_name": {
							Type: schema.TypeString,

							Computed: true,
							Optional: true,
						},
						"ignore_cookie": {
							Type: schema.TypeBool,

							Computed: true,
							Optional: true,
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
			Enabled: true,
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
					Enabled: true,
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
			Enabled: true,
			Value:   rawOption.(bool),
		}
	}

	if rawOption, ok := d.GetOk("options.0.fetched_compressed"); ok {
		optionsSet = true

		result.CompressionOptions = &cdn.ResourceOptions_CompressionOptions{
			CompressionVariant: &cdn.ResourceOptions_CompressionOptions_FetchCompressed{
				FetchCompressed: &cdn.ResourceOptions_BoolOption{
					Enabled: true,
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
					Enabled: true,
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
					Enabled: true,
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
					Enabled: true,
					Value:   rawOption.(bool),
				},
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.ignore_cookie"); ok {
		optionsSet = true

		result.IgnoreCookie = &cdn.ResourceOptions_BoolOption{
			Enabled: true,
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
						Enabled: true,
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
			Enabled: true,
			Value:   rawOption.(bool),
		}
	}

	if rawOption, ok := d.GetOk("options.0.disable_proxy_force_ranges"); ok {
		optionsSet = true

		result.DisableProxyForceRanges = &cdn.ResourceOptions_BoolOption{
			Enabled: true,
			Value:   rawOption.(bool),
		}
	}

	// TODO: Add option `static_request_headers`.

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
			Enabled: true,
			Value:   rawOption.(bool),
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

	if options.CustomServerName != nil {
		setIfEnabled("custom_server_name", options.CustomServerName.Enabled, options.CustomServerName.Value)
	}

	if options.IgnoreCookie != nil {
		setIfEnabled("ignore_cookie", options.IgnoreCookie.Enabled, options.IgnoreCookie.Value)
	}

	return []map[string]interface{}{
		item,
	}
}

func flattenYandexCDNResource(d *schema.ResourceData, resource *cdn.Resource) error {
	d.SetId(resource.Id)

	_ = d.Set("folder_id", resource.FolderId)
	_ = d.Set("cname", resource.Cname)

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
				return nil, fmt.Errorf("Error getting folder ID while creating instance: %s", err)
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
