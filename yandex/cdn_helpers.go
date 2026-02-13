package yandex

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/go-cty/cty/gocty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	validateCDNProvider = validation.StringInSlice([]string{cdnProviderOurcdn, cdnProviderGcore}, false)

	validateCDNShieldingLocation = validation.StringInSlice([]string{
		"1",   // ourcdn
		"130", // gcore
	}, false)

	validateCDNStale       = validation.StringInSlice([]string{`error`, `http_403`, `http_404`, `http_429`, `http_500`, `http_502`, `http_503`, `http_504`, `invalid_header`, `timeout`, `updating`}, false)
	validateCDNRewriteFlag = validation.StringInSlice([]string{"LAST", "BREAK", "REDIRECT", "PERMANENT"}, false)

	validateCDNResourceACLPolicyType = validation.StringInSlice(
		[]string{cdnACLPolicyTypeAllow, cdnACLPolicyTypeDeny},
		false,
	)
	validateCDNResourceSSLCertType = validation.StringInSlice(
		[]string{
			cdnSSLCertificateTypeNotUsed,
			cdnSSLCertificateTypeCM,
			cdnSSLCertificateTypeLE,
		},
		false,
	)
)

var (
	customizeDiffCDN_EdgeCacheSettings schema.CustomizeDiffFunc = func(_ context.Context, rd *schema.ResourceDiff, _ any) error {
		if _, ok := rd.GetOk("options.0.edge_cache_settings_codes"); ok {
			opts := rd.Get("options").([]any)[0].(map[string]any)
			opts["edge_cache_settings"] = nil
			if err := rd.SetNew("options", []any{opts}); err != nil {
				return err
			}
		} else if optsConfig := rd.GetRawConfig().GetAttr("options").AsValueSlice(); len(optsConfig) > 0 && optsConfig[0].GetAttr("edge_cache_settings").IsNull() {
			opts := rd.Get("options").([]any)[0].(map[string]any)
			opts["edge_cache_settings"] = 86400
			if err := rd.SetNew("options", []any{opts}); err != nil {
				return err
			}
		}
		return nil
	}
	customizeDiffCDN_RewriteFlag schema.CustomizeDiffFunc = func(_ context.Context, rd *schema.ResourceDiff, _ any) error {
		if _, ok := rd.GetOk("options.0.rewrite_pattern"); !ok {
			return nil
		}
		optsConfig := rd.GetRawConfig().GetAttr("options")
		if optsConfig.LengthInt() == 0 {
			return nil
		}
		flagConfig := optsConfig.AsValueSlice()[0].GetAttr("rewrite_flag")
		if !flagConfig.IsNull() {
			return nil
		}
		opts := rd.Get("options").([]any)[0].(map[string]any)
		opts["rewrite_flag"] = "BREAK"
		return rd.SetNew("options", []any{opts})
	}

	customizeDiffCDN_QueryParams schema.CustomizeDiffFunc = func(_ context.Context, rd *schema.ResourceDiff, _ any) error {
		optsConfig := rd.GetRawConfig().GetAttr("options")
		if optsConfig.LengthInt() != 1 {
			return nil
		}

		ignoreQueryParamsCfg := optsConfig.AsValueSlice()[0].GetAttr("ignore_query_params")
		queryParamsWhitelistCfg := optsConfig.AsValueSlice()[0].GetAttr("query_params_whitelist")
		queryParamsBlacklistCfg := optsConfig.AsValueSlice()[0].GetAttr("query_params_blacklist")
		var (
			ignoreQueryParamsPlan    = false
			queryParamsWhitelistPlan = []string{}
			queryParamsBlacklistPlan = []string{}
		)
		switch {
		case !ignoreQueryParamsCfg.IsNull():
			ignoreQueryParamsPlan = ignoreQueryParamsCfg.True()
		case !queryParamsWhitelistCfg.IsNull():
			gocty.FromCtyValue(queryParamsWhitelistCfg, &queryParamsWhitelistPlan)
		case !queryParamsBlacklistCfg.IsNull():
			gocty.FromCtyValue(queryParamsBlacklistCfg, &queryParamsBlacklistPlan)
		default:
			ignoreQueryParamsPlan = true
		}
		opts := rd.Get("options").([]any)[0].(map[string]any)
		opts["ignore_query_params"] = ignoreQueryParamsPlan
		opts["query_params_whitelist"] = queryParamsWhitelistPlan
		opts["query_params_blacklist"] = queryParamsBlacklistPlan
		return rd.SetNew("options", []any{opts})
	}
)

func cdnCheckProviderMatching(ctx context.Context, req *cdn.CreateResourceRequest, config *Config) error {
	originGroup, err := config.sdk.CDN().OriginGroup().Get(ctx, &cdn.GetOriginGroupRequest{
		FolderId:      req.FolderId,
		OriginGroupId: req.Origin.GetOriginGroupId(),
	})
	if err != nil {
		return fmt.Errorf("cannot check origin group: %w", err)
	}
	if req.ProviderType != originGroup.ProviderType {
		return fmt.Errorf(
			"cdn_resource provider %q does not match cdn_origin_group provider %q",
			req.ProviderType,
			originGroup.ProviderType,
		)
	}
	return nil
}

func prepareCDNUpdateResourceRequest(ctx context.Context, d *schema.ResourceData, config *Config) (*cdn.UpdateResourceRequest, error) {
	request := &cdn.UpdateResourceRequest{
		ResourceId: d.Id(),
	}
	if d.HasChange("origin_group_id") {
		groupID, _ := strconv.ParseInt(d.Get("origin_group_id").(string), 10, 64)
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
		request.SecondaryHostnames = expandCDNResourceSecondaryHostnames(d)
	}
	if d.HasChange("origin_protocol") {
		request.OriginProtocol = expandCDNResourceOriginProtocol(d)
	}
	if d.HasChange("active") {
		request.Active = wrapperspb.Bool(d.Get("active").(bool))
	}
	if d.HasChange("ssl_certificate") {
		var err error
		if request.SslCertificate, err = expandCDNResourceNewSSLCertificate(d); err != nil {
			return nil, err
		}
	}
	if d.HasChange("options") {
		request.Options = expandCDNResourceOptions(d)
	}
	if d.HasChange("labels") {
		request.Labels = expandCDNResourceLabels(d)
		if len(request.Labels) == 0 {
			request.RemoveLabels = true
		}
	}

	return request, nil
}

func prepareCDNCreateResourceRequest(ctx context.Context, d *schema.ResourceData, meta *Config) (*cdn.CreateResourceRequest, error) {
	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating instance: %s", err)
	}

	originVariant, err := expandCDNResourceOriginVariant(ctx, meta, folderID, d)
	if err != nil {
		return nil, err
	}
	provider := cdnProviderOurcdn
	if v := d.Get("provider_type"); v != "" {
		provider = v.(string)
	}

	var originProtocol cdn.OriginProtocol
	if _, ok := d.GetOk("origin_protocol"); ok {
		originProtocol = expandCDNResourceOriginProtocol(d)
	}

	var sslCertificate *cdn.SSLTargetCertificate
	if _, ok := d.GetOk("ssl_certificate"); ok {
		var err error
		if sslCertificate, err = expandCDNResourceNewSSLCertificate(d); err != nil {
			return nil, err
		}
	}

	result := &cdn.CreateResourceRequest{
		FolderId:     folderID,
		Cname:        d.Get("cname").(string),
		ProviderType: provider,

		Origin:         originVariant,
		OriginProtocol: originProtocol,

		Active:             wrapperspb.Bool(d.Get("active").(bool)),
		SecondaryHostnames: expandCDNResourceSecondaryHostnames(d),
		Options:            expandCDNResourceOptions(d),
		Labels:             expandCDNResourceLabels(d),
		SslCertificate:     sslCertificate,
	}

	return result, nil
}

func expandCDNResourceSecondaryHostnames(d *schema.ResourceData) *cdn.SecondaryHostnames {
	hostsSet := d.Get("secondary_hostnames").(*schema.Set)
	hostNames := castSlice[string](hostsSet.List())
	return &cdn.SecondaryHostnames{Values: hostNames}
}

func expandCDNResourceLabels(d *schema.ResourceData) map[string]string {
	labels := make(map[string]string)
	if rawOption, ok := d.GetOk("labels"); ok {
		for k, v := range rawOption.(map[string]any) {
			labels[k] = v.(string)
		}
	}
	return labels
}

func expandCDNResourceOriginVariant(ctx context.Context, meta *Config, folderID string, d *schema.ResourceData) (*cdn.CreateResourceRequest_Origin, error) {
	if v, ok := d.GetOk("origin_group_id"); ok {
		groupId, _ := strconv.ParseInt(v.(string), 10, 64)
		return &cdn.CreateResourceRequest_Origin{
			OriginVariant: &cdn.CreateResourceRequest_Origin_OriginGroupId{
				OriginGroupId: groupId,
			},
		}, nil
	}

	if v, ok := d.GetOk("origin_group_name"); ok {
		groupName := v.(string)

		groupID, err := resolveCDNOriginGroupID(ctx, meta, folderID, groupName)
		if err != nil {
			return nil, err
		}

		return &cdn.CreateResourceRequest_Origin{
			OriginVariant: &cdn.CreateResourceRequest_Origin_OriginGroupId{
				OriginGroupId: int64(groupID),
			},
		}, nil
	}

	return nil, nil
}

func expandCDNResourceOriginProtocol(d *schema.ResourceData) cdn.OriginProtocol {
	switch d.Get("origin_protocol").(string) {
	case "http":
		return cdn.OriginProtocol_HTTP
	case "https":
		return cdn.OriginProtocol_HTTPS
	case "match":
		return cdn.OriginProtocol_MATCH
	default:
		return cdn.OriginProtocol_ORIGIN_PROTOCOL_UNSPECIFIED
	}
}

func expandCDNResourceNewSSLCertificate_Type(certType string) cdn.SSLCertificateType {
	switch certType {
	case cdnSSLCertificateTypeNotUsed:
		return cdn.SSLCertificateType_DONT_USE
	case cdnSSLCertificateTypeCM:
		return cdn.SSLCertificateType_CM
	case cdnSSLCertificateTypeLE:
		return cdn.SSLCertificateType_LETS_ENCRYPT_GCORE
	default:
		return cdn.SSLCertificateType_SSL_CERTIFICATE_TYPE_UNSPECIFIED
	}
}

func expandCDNResourceNewSSLCertificate(d *schema.ResourceData) (*cdn.SSLTargetCertificate, error) {
	certSet, ok := d.Get("ssl_certificate").(*schema.Set)
	if !ok || certSet.Len() == 0 {
		return nil, nil
	}

	certFields := certSet.List()[0].(map[string]any)

	result := &cdn.SSLTargetCertificate{}
	result.Type = expandCDNResourceNewSSLCertificate_Type(certFields["type"].(string))

	if result.Type == cdn.SSLCertificateType_CM {
		cmCertID, exist := certFields["certificate_manager_id"]
		if !exist {
			return nil, fmt.Errorf("certificate_manager_id is mandatory field " +
				"for 'certificate_manager' SSL certificate type") // TODO: requiredWith in schema
		}
		result.Data = &cdn.SSLCertificateData{
			SslCertificateDataVariant: &cdn.SSLCertificateData_Cm{
				Cm: &cdn.SSLCertificateCMData{
					Id: cmCertID.(string),
				},
			},
		}
	}

	return result, nil
}

func expandCDNResourceOptions(d *schema.ResourceData) *cdn.ResourceOptions {
	if _, ok := d.GetOk("options"); !ok {
		log.Printf("[DEBUG] empty cdn resource options list")
		return nil
	}
	if size := d.Get("options.#").(int); size < 1 {
		log.Printf("[DEBUG] resource options list is empty")
		return nil
	}

	bcs := &cdn.ResourceOptions_Int64Option{
		Enabled: false,
		Value:   0,
	}
	if v := int64(d.Get("options.0.browser_cache_settings").(int)); v != 0 {
		bcs = &cdn.ResourceOptions_Int64Option{
			Enabled: true,
			Value:   v,
		}
	}

	result := &cdn.ResourceOptions{
		HostOptions:        expandCDNResourceOptions_HostOptions(d),
		QueryParamsOptions: expandCDNResourceOptions_QueryParamsOptions(d),
		CompressionOptions: expandCDNResourceOptions_CompressionOptions(d),
		RedirectOptions:    expandCDNResourceOptions_RedirectOptions(d),
		IpAddressAcl:       expandCDNResourceOptions_IPAddressACL(d),
		SecureKey:          expandCDNResourceOptions_SecureKey(d),
		EdgeCacheSettings:  expandCDNResourceOptions_EdgeCacheSettings(d),

		Slice:        cdnBoolOption(d.Get("options.0.slice").(bool)),
		IgnoreCookie: cdnBoolOption(d.Get("options.0.ignore_cookie").(bool)),

		Rewrite: expandCDNResourceOptions_Rewrite(d),

		Cors:  cdnStringListOption(d.Get("options.0.cors").([]any)),
		Stale: cdnStringListOption(d.Get("options.0.stale").([]any)),

		StaticHeaders:        cdnStringsMapOption(d.Get("options.0.static_response_headers").(map[string]any)),
		StaticRequestHeaders: cdnStringsMapOption(d.Get("options.0.static_request_headers").(map[string]any)),

		BrowserCacheSettings: bcs,
	}

	// bool options
	if rawOption, ok := d.GetOk("options.0.proxy_cache_method_set"); ok {
		result.ProxyCacheMethodsSet = cdnBoolOption(rawOption.(bool))
	}
	if rawOption, ok := d.GetOk("options.0.disable_proxy_force_ranges"); ok {
		result.DisableProxyForceRanges = cdnBoolOption(rawOption.(bool))
	}

	// stringList options
	if rawOption, ok := d.GetOk("options.0.allowed_http_methods"); ok {
		result.AllowedHttpMethods = cdnStringListOption(rawOption.([]any))
	}

	if rawOption, ok := d.GetOk("options.0.custom_server_name"); ok {
		result.CustomServerName = cdnStringOption(rawOption.(string))
	}

	return result
}

func expandCDNResourceOptions_EdgeCacheSettings(d *schema.ResourceData) *cdn.ResourceOptions_EdgeCacheSettings {
	if _, ok := d.GetOk("options.0.edge_cache_settings_codes"); ok {
		res := new(cdn.ResourceOptions_CachingTimes)
		if valueRaw, ok := d.GetOk("options.0.edge_cache_settings_codes.0.value"); ok {
			res.SimpleValue = int64(valueRaw.(int))
		}
		if custom_values, ok := d.GetOk("options.0.edge_cache_settings_codes.0.custom_values"); ok {
			mp := make(map[string]int64)
			for k, v := range custom_values.(map[string]any) {
				// TODO: validate keys
				mp[k] = int64(v.(int))
			}
			res.CustomValues = mp
		}
		return &cdn.ResourceOptions_EdgeCacheSettings{
			Enabled:       true,
			ValuesVariant: &cdn.ResourceOptions_EdgeCacheSettings_Value{Value: res},
		}
	}
	if rawOption, ok := d.GetOk("options.0.edge_cache_settings"); ok {
		return &cdn.ResourceOptions_EdgeCacheSettings{
			Enabled: true,
			ValuesVariant: &cdn.ResourceOptions_EdgeCacheSettings_DefaultValue{
				DefaultValue: int64(rawOption.(int)),
			},
		}
	}
	return &cdn.ResourceOptions_EdgeCacheSettings{}
}

func expandCDNResourceOptions_Rewrite(d *schema.ResourceData) *cdn.ResourceOptions_RewriteOption {
	pattern, ok := d.GetOk("options.0.rewrite_pattern")
	if !ok {
		return new(cdn.ResourceOptions_RewriteOption)
	}
	flag := d.Get("options.0.rewrite_flag").(string)
	return &cdn.ResourceOptions_RewriteOption{
		Enabled: true,
		Body:    pattern.(string),
		Flag:    cdn.RewriteFlag(cdn.RewriteFlag_value[flag]),
	}
}

func expandCDNResourceOptions_SecureKey(d *schema.ResourceData) *cdn.ResourceOptions_SecureKeyOption {
	rawOption, ok := d.GetOk("options.0.secure_key")
	if !ok {
		return nil
	}

	urlType := cdn.SecureKeyURLType_DISABLE_IP_SIGNING
	if rawUrlType, ok := d.GetOk("options.0.enable_ip_url_signing"); ok && rawUrlType.(bool) {
		urlType = cdn.SecureKeyURLType_ENABLE_IP_SIGNING
	}

	return &cdn.ResourceOptions_SecureKeyOption{
		Enabled: true,
		Key:     rawOption.(string),
		Type:    urlType,
	}
}

func expandCDNResourceOptions_IPAddressACL(d *schema.ResourceData) *cdn.ResourceOptions_IPAddressACLOption {
	if _, ok := d.GetOk("options.0.ip_address_acl"); !ok {
		return nil
	}
	if size := d.Get("options.0.ip_address_acl.#").(int); size <= 0 {
		return nil
	}
	rawPolicyType, ok := d.GetOk("options.0.ip_address_acl.0.policy_type")
	if !ok {
		return nil
	}

	var values []string
	if rawExceptedValues, ok := d.GetOk("options.0.ip_address_acl.0.excepted_values"); ok {
		for _, v := range rawExceptedValues.([]any) {
			values = append(values, v.(string))
		}
	}

	return &cdn.ResourceOptions_IPAddressACLOption{
		Enabled:        true,
		PolicyType:     aclPolicyTypeFromString(rawPolicyType.(string)),
		ExceptedValues: values,
	}
}

func expandCDNResourceOptions_RedirectOptions(d *schema.ResourceData) *cdn.ResourceOptions_RedirectOptions {
	if rawOption, ok := d.GetOk("options.0.redirect_http_to_https"); ok {
		return &cdn.ResourceOptions_RedirectOptions{
			RedirectVariant: &cdn.ResourceOptions_RedirectOptions_RedirectHttpToHttps{
				RedirectHttpToHttps: cdnBoolOption(rawOption.(bool)),
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.redirect_https_to_http"); ok {
		return &cdn.ResourceOptions_RedirectOptions{
			RedirectVariant: &cdn.ResourceOptions_RedirectOptions_RedirectHttpsToHttp{
				RedirectHttpsToHttp: cdnBoolOption(rawOption.(bool)),
			},
		}
	}
	return nil
}

func expandCDNResourceOptions_CompressionOptions(d *schema.ResourceData) *cdn.ResourceOptions_CompressionOptions {
	if rawOption, ok := d.GetOk("options.0.fetched_compressed"); ok {
		return &cdn.ResourceOptions_CompressionOptions{
			CompressionVariant: &cdn.ResourceOptions_CompressionOptions_FetchCompressed{
				FetchCompressed: cdnBoolOption(rawOption.(bool)),
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.gzip_on"); ok {
		return &cdn.ResourceOptions_CompressionOptions{
			CompressionVariant: &cdn.ResourceOptions_CompressionOptions_GzipOn{
				GzipOn: cdnBoolOption(rawOption.(bool)),
			},
		}
	}
	// TODO: brotli
	return nil
}

func expandCDNResourceOptions_QueryParamsOptions(d *schema.ResourceData) *cdn.ResourceOptions_QueryParamsOptions {
	if rawOption, ok := d.GetOk("options.0.query_params_whitelist"); ok {
		option := cdnStringListOption(rawOption.([]any))
		if option != nil {
			return &cdn.ResourceOptions_QueryParamsOptions{
				QueryParamsVariant: &cdn.ResourceOptions_QueryParamsOptions_QueryParamsWhitelist{
					QueryParamsWhitelist: option,
				},
			}
		}
	}

	if rawOption, ok := d.GetOk("options.0.query_params_blacklist"); ok {
		option := cdnStringListOption(rawOption.([]any))
		if option != nil {
			return &cdn.ResourceOptions_QueryParamsOptions{
				QueryParamsVariant: &cdn.ResourceOptions_QueryParamsOptions_QueryParamsBlacklist{
					QueryParamsBlacklist: option,
				},
			}
		}
	}
	rawOption := d.Get("options.0.ignore_query_params")
	return &cdn.ResourceOptions_QueryParamsOptions{
		QueryParamsVariant: &cdn.ResourceOptions_QueryParamsOptions_IgnoreQueryString{
			IgnoreQueryString: cdnBoolOption(rawOption.(bool)),
		},
	}
}

func expandCDNResourceOptions_HostOptions(d *schema.ResourceData) *cdn.ResourceOptions_HostOptions {
	if rawOption, ok := d.GetOk("options.0.custom_host_header"); ok && rawOption.(string) != "" {
		return &cdn.ResourceOptions_HostOptions{
			HostVariant: &cdn.ResourceOptions_HostOptions_Host{
				Host: cdnStringOption(rawOption.(string)),
			},
		}
	}

	if rawOption, ok := d.GetOk("options.0.forward_host_header"); ok && rawOption.(bool) {
		return &cdn.ResourceOptions_HostOptions{
			HostVariant: &cdn.ResourceOptions_HostOptions_ForwardHostHeader{
				ForwardHostHeader: cdnBoolOption(rawOption.(bool)),
			},
		}
	}
	return nil
}

func cdnStringListOption(value []any) *cdn.ResourceOptions_StringsListOption {
	if len(value) == 0 {
		return new(cdn.ResourceOptions_StringsListOption)
	}
	return &cdn.ResourceOptions_StringsListOption{
		Enabled: true,
		Value:   castSlice[string](value),
	}
}

func cdnStringOption(value string) *cdn.ResourceOptions_StringOption {
	return &cdn.ResourceOptions_StringOption{
		Enabled: true,
		Value:   value,
	}
}

func cdnBoolOption(value bool) *cdn.ResourceOptions_BoolOption {
	return &cdn.ResourceOptions_BoolOption{
		Enabled: value,
		Value:   value,
	}
}

func cdnStringsMapOption(rawOption map[string]any) *cdn.ResourceOptions_StringsMapOption {
	if len(rawOption) == 0 {
		return new(cdn.ResourceOptions_StringsMapOption)
	}
	res := &cdn.ResourceOptions_StringsMapOption{
		Enabled: true,
		Value:   make(map[string]string),
	}

	for k, v := range rawOption {
		res.Value[k] = v.(string)
	}
	return res
}

func castSlice[T any](arr []any) []T {
	res := make([]T, len(arr))
	for i := range arr {
		res[i] = arr[i].(T)
	}
	return res
}

func flattenCDNResource(resource *cdn.Resource, shieldingLocation *int64) (map[string]any, error) {
	res := make(map[string]any)
	res["folder_id"] = resource.FolderId
	res["cname"] = resource.Cname
	res["labels"] = resource.Labels

	res["created_at"] = getTimestamp(resource.CreatedAt)
	res["updated_at"] = getTimestamp(resource.UpdatedAt)

	res["active"] = resource.Active
	res["provider_cname"] = resource.ProviderCname
	res["provider_type"] = resource.ProviderType

	res["origin_group_name"] = resource.OriginGroupName
	res["origin_group_id"] = fmt.Sprint(resource.OriginGroupId)

	res["shielding"] = flattenCDNShielding(shieldingLocation)

	if secondaryHostnames := flattenCDNResourceSecondaryNames(resource.SecondaryHostnames); secondaryHostnames != nil {
		res["secondary_hostnames"] = secondaryHostnames
	}
	if protocol, err := flattenCDNResourceOriginProtocol(resource.OriginProtocol); err == nil {
		res["origin_protocol"] = protocol
	} else {
		return nil, err
	}
	if cert, err := flattenCDNResourceSSLCertificate(resource.SslCertificate); err == nil {
		res["ssl_certificate"] = cert
	} else {
		return nil, err
	}
	if opts, err := flattenCDNResourceOptions(resource.Options); err == nil {
		res["options"] = opts
	} else {
		return nil, err
	}
	return res, nil
}

func flattenCDNShielding(shielding *int64) any {
	if shielding == nil {
		return nil
	}
	return fmt.Sprint(*shielding)
}

func flattenCDNResourceSecondaryNames(secondaryHostnames []string) []any {
	if len(secondaryHostnames) == 0 {
		return nil
	}

	result := make([]any, len(secondaryHostnames))
	for i := range secondaryHostnames {
		result[i] = secondaryHostnames[i]
	}

	return result
}

func flattenCDNResourceOriginProtocol(protocol cdn.OriginProtocol) (string, error) {
	switch protocol {
	case cdn.OriginProtocol_HTTP:
		return "http", nil
	case cdn.OriginProtocol_HTTPS:
		return "https", nil
	case cdn.OriginProtocol_MATCH:
		return "match", nil
	default:
		return "", fmt.Errorf("unexpected origin protocol value in API response")
	}
}

func flattenCDNResourceSSLCertificate(cert *cdn.SSLCertificate) ([]map[string]any, error) {
	if cert == nil {
		return nil, nil
	}

	result := make(map[string]any)

	var typeStr string
	switch cert.Type {
	case cdn.SSLCertificateType_DONT_USE:
		typeStr = cdnSSLCertificateTypeNotUsed
	case cdn.SSLCertificateType_LETS_ENCRYPT_GCORE:
		typeStr = cdnSSLCertificateTypeLE
	case cdn.SSLCertificateType_CM:
		typeStr = cdnSSLCertificateTypeCM
	default:
		return nil, fmt.Errorf("unexpected ssl certificate type in API response")
	}

	var statusStr string
	switch cert.Status {
	case cdn.SSLCertificateStatus_READY:
		statusStr = cdnSSLCertificateStatusReady
	case cdn.SSLCertificateStatus_CREATING:
		statusStr = cdnSSLCertificateStatusCreating
	}

	if cert.Type == cdn.SSLCertificateType_CM {
		if cert.Data == nil || cert.Data.GetCm() == nil {
			return nil, fmt.Errorf("certificate manager data is absent in API response")
		}
		result["certificate_manager_id"] = cert.Data.GetCm().GetId()
	}

	result["type"] = typeStr
	result["status"] = statusStr
	return []map[string]any{result}, nil
}

func flattenCDNResourceOptions(options *cdn.ResourceOptions) ([]map[string]any, error) {
	if options == nil {
		log.Printf("[DEBUG] empty cdn resource options set")
		return nil, nil
	}

	item := make(map[string]any)

	setIfEnabled := func(optionName string, enabled bool, value any) {
		if !enabled {
			return
		}
		item[optionName] = value
	}

	flattenCDNResourceOptions_EdgeCacheSettings(options.EdgeCacheSettings, item)

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

	if options.BrowserCacheSettings != nil {
		setIfEnabled("browser_cache_settings", options.BrowserCacheSettings.Enabled, options.BrowserCacheSettings.Value)
	}

	if options.Slice != nil {
		setIfEnabled("slice", options.Slice.Enabled, options.Slice.Value)
	}

	if options.Cors != nil {
		setIfEnabled("cors", options.Cors.Enabled, options.Cors.Value)
	}
	if options.Stale != nil {
		setIfEnabled("stale", options.Stale.Enabled, options.Stale.Value)
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
	if options.Rewrite != nil {
		setIfEnabled("rewrite_pattern", options.Rewrite.Enabled, options.Rewrite.Body)
		setIfEnabled("rewrite_flag", options.Rewrite.Enabled, options.Rewrite.Flag.String())
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
		ipAddrACL := make(map[string]any)
		ipAddrACL["policy_type"] = aclPolicyTypeToString(options.IpAddressAcl.PolicyType)
		ipAddrACL["excepted_values"] = options.IpAddressAcl.ExceptedValues

		setIfEnabled("ip_address_acl", options.IpAddressAcl.Enabled, []map[string]any{ipAddrACL})
	}

	return []map[string]any{item}, nil
}

func flattenCDNResourceOptions_EdgeCacheSettings(opts *cdn.ResourceOptions_EdgeCacheSettings, dest map[string]any) {
	if opts == nil || !opts.Enabled {
		return
	}
	switch v := opts.ValuesVariant.(type) {
	case *cdn.ResourceOptions_EdgeCacheSettings_DefaultValue:
		dest["edge_cache_settings"] = v.DefaultValue
	case *cdn.ResourceOptions_EdgeCacheSettings_Value:
		dest["edge_cache_settings"] = 0
		dest["edge_cache_settings_codes"] = []map[string]any{{
			"value":         v.Value.SimpleValue,
			"custom_values": v.Value.CustomValues,
		}}
	}
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

func getShieldingLocation(ctx context.Context, resourceId string, sdk *ycsdk.SDK) (*int64, error) {
	resp, err := sdk.CDN().Shielding().Get(ctx, &cdn.GetShieldingDetailsRequest{
		ResourceId: resourceId,
	})
	if isStatusWithCode(err, codes.NotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &resp.LocationId, nil
}

func updateShielding(ctx context.Context, d *schema.ResourceData, config *Config) error {
	if !d.HasChange("shielding") {
		return nil
	}
	val, ok := d.GetOk("shielding")
	if ok {
		locationId, _ := strconv.Atoi(val.(string))
		return cdnEnableShielding(ctx, config, d.Id(), int64(locationId))
	}
	return cdnDisableShielding(ctx, config, d.Id())
}

func cdnDisableShielding(ctx context.Context, config *Config, resourceId string) error {
	res, err := config.sdk.WrapOperation(
		config.sdk.CDN().Shielding().Deactivate(ctx, &cdn.DeactivateShieldingRequest{
			ResourceId: resourceId,
		}),
	)
	if err != nil {
		return err
	}
	return res.Wait(ctx)
}

func cdnEnableShielding(ctx context.Context, config *Config, resourceId string, locationId int64) error {
	res, err := config.sdk.WrapOperation(
		config.sdk.CDN().Shielding().Activate(ctx, &cdn.ActivateShieldingRequest{
			ResourceId: resourceId,
			LocationId: locationId,
		}),
	)
	if err != nil {
		return err
	}
	return res.Wait(ctx)
}
