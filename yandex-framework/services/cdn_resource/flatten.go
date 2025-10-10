package cdn_resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

// FlattenCDNResourceOptions converts CDN API ResourceOptions to Terraform state
// planOptions: optional plan options to preserve disabled cache blocks
// When API returns nil for cache settings but plan has enabled=false,
// we preserve the disabled block in state to prevent plan/apply inconsistency
// Exported for reuse in cdn_rule package
func FlattenCDNResourceOptions(ctx context.Context, options *cdn.ResourceOptions, planOptions types.List, diags *diag.Diagnostics) types.List {
	if options == nil {
		return types.ListNull(types.ObjectType{
			AttrTypes: getCDNOptionsAttrTypes(),
		})
	}

	// Extract plan options model if available
	// Errors are logged but don't fail the operation (graceful degradation)
	var planOptionsModel *CDNOptionsModel
	if !planOptions.IsNull() && len(planOptions.Elements()) > 0 {
		var planOptionsModels []CDNOptionsModel
		d := planOptions.ElementsAs(ctx, &planOptionsModels, false)
		if d.HasError() {
			tflog.Warn(ctx, "Failed to extract plan options", map[string]interface{}{
				"error": d.Errors(),
			})
		} else if len(planOptionsModels) > 0 {
			planOptionsModel = &planOptionsModels[0]
		}
	}

	opt := CDNOptionsModel{}

	// Boolean options - CRITICAL: Set null when Enabled=false to prevent state drift
	opt.Slice = flattenBoolOption(options.Slice)
	opt.IgnoreCookie = flattenBoolOption(options.IgnoreCookie)
	opt.ProxyCacheMethodsSet = flattenBoolOption(options.ProxyCacheMethodsSet)
	opt.DisableProxyForceRanges = flattenBoolOption(options.DisableProxyForceRanges)

	// Cache settings - nested blocks (pass plan to preserve disabled blocks)
	opt.EdgeCacheSettings = flattenEdgeCacheSettings(ctx, options.EdgeCacheSettings, planOptionsModel, diags)
	opt.BrowserCacheSettings = flattenBrowserCacheSettings(ctx, options.BrowserCacheSettings, planOptionsModel, diags)

	// String options - CORRECT SEMANTICS: null when not configured
	if options.CustomServerName != nil && options.CustomServerName.Enabled {
		opt.CustomServerName = types.StringValue(options.CustomServerName.Value)
	} else {
		opt.CustomServerName = types.StringNull()
	}

	// SecureKey - combines secure_key and enable_ip_url_signing
	// CORRECT SEMANTICS: Both null when secure_key is not configured
	if options.SecureKey != nil && options.SecureKey.Enabled {
		opt.SecureKey = types.StringValue(options.SecureKey.Key)
		// EnableIPURLSigning is derived from SecureKey.Type
		if options.SecureKey.Type == cdn.SecureKeyURLType_ENABLE_IP_SIGNING {
			opt.EnableIPURLSigning = types.BoolValue(true)
		} else {
			opt.EnableIPURLSigning = types.BoolValue(false)
		}
	} else {
		opt.SecureKey = types.StringNull()
		opt.EnableIPURLSigning = types.BoolNull()
	}

	// List options - CORRECT SEMANTICS: null when not configured
	if options.CacheHttpHeaders != nil && options.CacheHttpHeaders.Enabled {
		listVal, d := types.ListValueFrom(ctx, types.StringType, options.CacheHttpHeaders.Value)
		diags.Append(d...)
		opt.CacheHTTPHeaders = listVal
	} else {
		opt.CacheHTTPHeaders = types.ListNull(types.StringType)
	}

	if options.Cors != nil && options.Cors.Enabled {
		listVal, d := types.ListValueFrom(ctx, types.StringType, options.Cors.Value)
		diags.Append(d...)
		opt.Cors = listVal
	} else {
		opt.Cors = types.ListNull(types.StringType)
	}

	if options.AllowedHttpMethods != nil && options.AllowedHttpMethods.Enabled {
		listVal, d := types.ListValueFrom(ctx, types.StringType, options.AllowedHttpMethods.Value)
		diags.Append(d...)
		opt.AllowedHTTPMethods = listVal
	} else {
		opt.AllowedHTTPMethods = types.ListNull(types.StringType)
	}

	// Map options - CORRECT SEMANTICS: null when not configured
	if options.StaticHeaders != nil && options.StaticHeaders.Enabled {
		mapVal, d := types.MapValueFrom(ctx, types.StringType, options.StaticHeaders.Value)
		diags.Append(d...)
		opt.StaticResponseHeaders = mapVal
	} else {
		opt.StaticResponseHeaders = types.MapNull(types.StringType)
	}

	if options.StaticRequestHeaders != nil && options.StaticRequestHeaders.Enabled {
		mapVal, d := types.MapValueFrom(ctx, types.StringType, options.StaticRequestHeaders.Value)
		diags.Append(d...)
		opt.StaticRequestHeaders = mapVal
	} else {
		opt.StaticRequestHeaders = types.MapNull(types.StringType)
	}

	// Mutually exclusive options groups
	flattenHostOptions(options.HostOptions, &opt)
	flattenQueryParamsOptions(ctx, options.QueryParamsOptions, &opt, diags)
	flattenCompressionOptions(options.CompressionOptions, &opt)
	flattenRedirectOptions(options.RedirectOptions, &opt)

	// Nested blocks
	flattenIPAddressACL(ctx, options.IpAddressAcl, &opt, diags)
	flattenRewrite(ctx, options.Rewrite, &opt, diags)

	optionsList, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
	}, []CDNOptionsModel{opt})
	diags.Append(d...)

	return optionsList
}

// flattenBoolOption converts CDN API BoolOption to types.Bool with proper null handling
// CORRECT SEMANTICS: Enabled=false means "not configured by user" → return null
// This is the proper Framework way - null = "provider doesn't manage this field"
func flattenBoolOption(option *cdn.ResourceOptions_BoolOption) types.Bool {
	if option == nil || !option.Enabled {
		// Not configured in API = not managed by provider
		return types.BoolNull()
	}
	return types.BoolValue(option.Value)
}

// flattenHostOptions handles mutually exclusive forward_host_header and custom_host_header
// IMPORTANT: Returns zero values for inactive fields to work with plan modifiers
// expand.go will check if ALL fields are zero values before sending to API
func flattenHostOptions(hostOptions *cdn.ResourceOptions_HostOptions, opt *CDNOptionsModel) {
	if hostOptions == nil {
		// No host options configured → zero values for both
		opt.ForwardHostHeader = types.BoolValue(false)
		opt.CustomHostHeader = types.StringValue("")
		return
	}

	switch variant := hostOptions.HostVariant.(type) {
	case *cdn.ResourceOptions_HostOptions_ForwardHostHeader:
		// forward_host_header is active → set its value, custom_host_header gets zero value
		if variant.ForwardHostHeader != nil && variant.ForwardHostHeader.Enabled {
			opt.ForwardHostHeader = types.BoolValue(variant.ForwardHostHeader.Value)
		} else {
			opt.ForwardHostHeader = types.BoolValue(false)
		}
		opt.CustomHostHeader = types.StringValue("") // Inactive field → zero value
	case *cdn.ResourceOptions_HostOptions_Host:
		// custom_host_header is active → set its value, forward_host_header gets zero value
		if variant.Host != nil && variant.Host.Enabled {
			opt.CustomHostHeader = types.StringValue(variant.Host.Value)
		} else {
			opt.CustomHostHeader = types.StringValue("")
		}
		opt.ForwardHostHeader = types.BoolValue(false) // Inactive field → zero value
	default:
		// Unknown variant → zero values for both
		opt.ForwardHostHeader = types.BoolValue(false)
		opt.CustomHostHeader = types.StringValue("")
	}
}

// flattenQueryParamsOptions handles mutually exclusive query params options
// IMPORTANT: Returns zero values for inactive fields to work with plan modifiers
// expand.go will check if ALL fields are zero values before sending to API
func flattenQueryParamsOptions(ctx context.Context, queryOptions *cdn.ResourceOptions_QueryParamsOptions, opt *CDNOptionsModel, diags *diag.Diagnostics) {
	// Initialize all to zero values
	opt.IgnoreQueryParams = types.BoolValue(false)
	opt.QueryParamsWhitelist = types.ListNull(types.StringType)
	opt.QueryParamsBlacklist = types.ListNull(types.StringType)

	if queryOptions == nil {
		return // All remain at zero values
	}

	switch variant := queryOptions.QueryParamsVariant.(type) {
	case *cdn.ResourceOptions_QueryParamsOptions_IgnoreQueryString:
		// ignore_query_params is active
		if variant.IgnoreQueryString != nil && variant.IgnoreQueryString.Enabled {
			opt.IgnoreQueryParams = types.BoolValue(variant.IgnoreQueryString.Value)
		}
		// whitelist and blacklist remain null (inactive fields)
	case *cdn.ResourceOptions_QueryParamsOptions_QueryParamsWhitelist:
		// query_params_whitelist is active
		if variant.QueryParamsWhitelist != nil && variant.QueryParamsWhitelist.Enabled {
			listVal, d := types.ListValueFrom(ctx, types.StringType, variant.QueryParamsWhitelist.Value)
			diags.Append(d...)
			opt.QueryParamsWhitelist = listVal
		}
		// ignore_query_params remains false, blacklist remains null (inactive fields)
	case *cdn.ResourceOptions_QueryParamsOptions_QueryParamsBlacklist:
		// query_params_blacklist is active
		if variant.QueryParamsBlacklist != nil && variant.QueryParamsBlacklist.Enabled {
			listVal, d := types.ListValueFrom(ctx, types.StringType, variant.QueryParamsBlacklist.Value)
			diags.Append(d...)
			opt.QueryParamsBlacklist = listVal
		}
		// ignore_query_params remains false, whitelist remains null (inactive fields)
	}
}

// flattenCompressionOptions handles mutually exclusive gzip_on and fetched_compressed
// IMPORTANT: Returns zero values for inactive fields to work with plan modifiers
// expand.go will check if ALL fields are zero values before sending to API
func flattenCompressionOptions(compressionOptions *cdn.ResourceOptions_CompressionOptions, opt *CDNOptionsModel) {
	// Initialize both to false (zero value for bool)
	opt.GzipOn = types.BoolValue(false)
	opt.FetchedCompressed = types.BoolValue(false)

	if compressionOptions == nil {
		return // Both remain false
	}

	switch variant := compressionOptions.CompressionVariant.(type) {
	case *cdn.ResourceOptions_CompressionOptions_GzipOn:
		// gzip_on is active
		if variant.GzipOn != nil && variant.GzipOn.Enabled {
			opt.GzipOn = types.BoolValue(variant.GzipOn.Value)
		}
		// fetched_compressed remains false (inactive field)
	case *cdn.ResourceOptions_CompressionOptions_FetchCompressed:
		// fetched_compressed is active
		if variant.FetchCompressed != nil && variant.FetchCompressed.Enabled {
			opt.FetchedCompressed = types.BoolValue(variant.FetchCompressed.Value)
		}
		// gzip_on remains false (inactive field)
	}
}

// flattenRedirectOptions handles mutually exclusive redirect options
// IMPORTANT: Returns zero values for inactive fields to work with plan modifiers
// expand.go will check if ALL fields are zero values before sending to API
func flattenRedirectOptions(redirectOptions *cdn.ResourceOptions_RedirectOptions, opt *CDNOptionsModel) {
	// Initialize both to false (zero value for bool)
	opt.RedirectHttpToHttps = types.BoolValue(false)
	opt.RedirectHttpsToHttp = types.BoolValue(false)

	if redirectOptions == nil {
		return // Both remain false
	}

	switch variant := redirectOptions.RedirectVariant.(type) {
	case *cdn.ResourceOptions_RedirectOptions_RedirectHttpToHttps:
		// redirect_http_to_https is active
		if variant.RedirectHttpToHttps != nil && variant.RedirectHttpToHttps.Enabled {
			opt.RedirectHttpToHttps = types.BoolValue(variant.RedirectHttpToHttps.Value)
		}
		// redirect_https_to_http remains false (inactive field)
	case *cdn.ResourceOptions_RedirectOptions_RedirectHttpsToHttp:
		// redirect_https_to_http is active
		if variant.RedirectHttpsToHttp != nil && variant.RedirectHttpsToHttp.Enabled {
			opt.RedirectHttpsToHttp = types.BoolValue(variant.RedirectHttpsToHttp.Value)
		}
		// redirect_http_to_https remains false (inactive field)
	}
}

// flattenIPAddressACL converts API IP address ACL to Terraform state
func flattenIPAddressACL(ctx context.Context, acl *cdn.ResourceOptions_IPAddressACLOption, opt *CDNOptionsModel, diags *diag.Diagnostics) {
	if acl == nil {
		opt.IPAddressACL = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"policy_type":     types.StringType,
				"excepted_values": types.ListType{ElemType: types.StringType},
			},
		})
		return
	}

	var policyType string
	switch acl.PolicyType {
	case cdn.PolicyType_POLICY_TYPE_ALLOW:
		policyType = "allow"
	case cdn.PolicyType_POLICY_TYPE_DENY:
		policyType = "deny"
	default:
		policyType = "allow"
	}

	exceptedList, d := types.ListValueFrom(ctx, types.StringType, acl.ExceptedValues)
	diags.Append(d...)

	aclModel := IPAddressACLModel{
		PolicyType:     types.StringValue(policyType),
		ExceptedValues: exceptedList,
	}

	aclList, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"policy_type":     types.StringType,
			"excepted_values": types.ListType{ElemType: types.StringType},
		},
	}, []IPAddressACLModel{aclModel})
	diags.Append(d...)

	opt.IPAddressACL = aclList
}

// flattenRewrite converts API rewrite option to Terraform state
func flattenRewrite(ctx context.Context, rewrite *cdn.ResourceOptions_RewriteOption, opt *CDNOptionsModel, diags *diag.Diagnostics) {
	if rewrite == nil || !rewrite.Enabled {
		opt.Rewrite = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled": types.BoolType,
				"body":    types.StringType,
				"flag":    types.StringType,
			},
		})
		return
	}

	var flag string
	switch rewrite.Flag {
	case cdn.RewriteFlag_LAST:
		flag = "last"
	case cdn.RewriteFlag_BREAK:
		flag = "break"
	case cdn.RewriteFlag_REDIRECT:
		flag = "redirect"
	case cdn.RewriteFlag_PERMANENT:
		flag = "permanent"
	default:
		flag = "break"
	}

	rewriteModel := RewriteModel{
		Enabled: types.BoolValue(rewrite.Enabled),
		Body:    types.StringValue(rewrite.Body),
		Flag:    types.StringValue(flag),
	}

	rewriteList, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled": types.BoolType,
			"body":    types.StringType,
			"flag":    types.StringType,
		},
	}, []RewriteModel{rewriteModel})
	diags.Append(d...)
	opt.Rewrite = rewriteList
}

// flattenEdgeCacheSettings converts API EdgeCacheSettings to Terraform state
// Handles two API variants:
// 1. DefaultValue → cache_time = {"*" = value}
// 2. Value.CustomValues → cache_time = map
// planOptionsModel: optional plan options model to preserve disabled blocks
func flattenEdgeCacheSettings(ctx context.Context, edgeCache *cdn.ResourceOptions_EdgeCacheSettings, planOptionsModel *CDNOptionsModel, diags *diag.Diagnostics) types.List {
	edgeCacheAttrTypes := map[string]attr.Type{
		"enabled":    types.BoolType,
		"cache_time": types.MapType{ElemType: types.Int64Type},
	}

	// If API returns nil or disabled, check if plan has enabled=false
	if edgeCache == nil || !edgeCache.Enabled {
		// If plan has a block with enabled=false, preserve it in state
		if planOptionsModel != nil && !planOptionsModel.EdgeCacheSettings.IsNull() && len(planOptionsModel.EdgeCacheSettings.Elements()) > 0 {
			var planEdgeCache []EdgeCacheSettingsModel
			d := planOptionsModel.EdgeCacheSettings.ElementsAs(ctx, &planEdgeCache, false)
			if d.HasError() {
				tflog.Warn(ctx, "Failed to extract plan edge cache settings", map[string]interface{}{
					"error": d.Errors(),
				})
			} else if len(planEdgeCache) > 0 {
				// Check if enabled field is explicitly false (not null/unknown)
				if !planEdgeCache[0].Enabled.IsNull() && !planEdgeCache[0].Enabled.IsUnknown() && !planEdgeCache[0].Enabled.ValueBool() {
					// Plan has enabled=false, preserve the disabled block in state
					disabledModel := EdgeCacheSettingsModel{
						Enabled:   types.BoolValue(false),
						CacheTime: types.MapNull(types.Int64Type),
					}
					disabledList, d := types.ListValueFrom(ctx, types.ObjectType{
						AttrTypes: edgeCacheAttrTypes,
					}, []EdgeCacheSettingsModel{disabledModel})
					diags.Append(d...)
					return disabledList
				}
			}
		}
		// Otherwise return null (truly not configured)
		return types.ListNull(types.ObjectType{AttrTypes: edgeCacheAttrTypes})
	}

	edgeCacheModel := EdgeCacheSettingsModel{
		Enabled: types.BoolValue(true),
	}

	// Handle cache_time based on API response
	if edgeCache.ValuesVariant != nil {
		switch v := edgeCache.ValuesVariant.(type) {
		case *cdn.ResourceOptions_EdgeCacheSettings_DefaultValue:
			// DefaultValue variant → create cache_time = {"*" = value}
			cacheTimeMap := map[string]int64{
				"*": v.DefaultValue,
			}
			mapVal, d := types.MapValueFrom(ctx, types.Int64Type, cacheTimeMap)
			diags.Append(d...)
			edgeCacheModel.CacheTime = mapVal
		case *cdn.ResourceOptions_EdgeCacheSettings_Value:
			// Value variant with CustomValues → create cache_time = map
			if v.Value != nil && len(v.Value.CustomValues) > 0 {
				mapVal, d := types.MapValueFrom(ctx, types.Int64Type, v.Value.CustomValues)
				diags.Append(d...)
				edgeCacheModel.CacheTime = mapVal
			} else {
				edgeCacheModel.CacheTime = types.MapNull(types.Int64Type)
			}
		default:
			edgeCacheModel.CacheTime = types.MapNull(types.Int64Type)
		}
	} else {
		edgeCacheModel.CacheTime = types.MapNull(types.Int64Type)
	}

	edgeCacheList, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: edgeCacheAttrTypes,
	}, []EdgeCacheSettingsModel{edgeCacheModel})
	diags.Append(d...)

	return edgeCacheList
}

// flattenBrowserCacheSettings converts API BrowserCacheSettings to Terraform state
// planOptionsModel: optional plan options model to preserve disabled blocks
func flattenBrowserCacheSettings(ctx context.Context, browserCache *cdn.ResourceOptions_Int64Option, planOptionsModel *CDNOptionsModel, diags *diag.Diagnostics) types.List {
	browserCacheAttrTypes := map[string]attr.Type{
		"enabled":    types.BoolType,
		"cache_time": types.Int64Type,
	}

	// If API returns nil or disabled, check if plan has enabled=false
	if browserCache == nil || !browserCache.Enabled {
		// If plan has a block with enabled=false, preserve it in state
		if planOptionsModel != nil && !planOptionsModel.BrowserCacheSettings.IsNull() && len(planOptionsModel.BrowserCacheSettings.Elements()) > 0 {
			var planBrowserCache []BrowserCacheSettingsModel
			d := planOptionsModel.BrowserCacheSettings.ElementsAs(ctx, &planBrowserCache, false)
			if d.HasError() {
				tflog.Warn(ctx, "Failed to extract plan browser cache settings", map[string]interface{}{
					"error": d.Errors(),
				})
			} else if len(planBrowserCache) > 0 {
				// Check if enabled field is explicitly false (not null/unknown)
				if !planBrowserCache[0].Enabled.IsNull() && !planBrowserCache[0].Enabled.IsUnknown() && !planBrowserCache[0].Enabled.ValueBool() {
					// Plan has enabled=false, preserve the disabled block in state
					disabledModel := BrowserCacheSettingsModel{
						Enabled:   types.BoolValue(false),
						CacheTime: types.Int64Null(),
					}
					disabledList, d := types.ListValueFrom(ctx, types.ObjectType{
						AttrTypes: browserCacheAttrTypes,
					}, []BrowserCacheSettingsModel{disabledModel})
					diags.Append(d...)
					return disabledList
				}
			}
		}
		// Otherwise return null (truly not configured)
		return types.ListNull(types.ObjectType{AttrTypes: browserCacheAttrTypes})
	}

	// Only create block when enabled=true
	browserCacheModel := BrowserCacheSettingsModel{
		Enabled:   types.BoolValue(true),
		CacheTime: types.Int64Value(browserCache.Value),
	}

	browserCacheList, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: browserCacheAttrTypes,
	}, []BrowserCacheSettingsModel{browserCacheModel})
	diags.Append(d...)

	return browserCacheList
}

// flattenSSLCertificate converts API SSL certificate to Terraform state
func flattenSSLCertificate(ctx context.Context, cert *cdn.SSLCertificate, diags *diag.Diagnostics) types.Set {
	if cert == nil {
		return types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":                   types.StringType,
				"status":                 types.StringType,
				"certificate_manager_id": types.StringType,
			},
		})
	}

	var certType string
	switch cert.Type {
	case cdn.SSLCertificateType_DONT_USE:
		certType = "not_used"
	case cdn.SSLCertificateType_CM:
		certType = "certificate_manager"
	case cdn.SSLCertificateType_LETS_ENCRYPT_GCORE:
		certType = "lets_encrypt"
	default:
		certType = "not_used"
	}

	var status string
	switch cert.Status {
	case cdn.SSLCertificateStatus_READY:
		status = "ready"
	case cdn.SSLCertificateStatus_CREATING:
		status = "creating"
	default:
		status = ""
	}

	// Get certificate manager ID if available
	var cmID string
	if cert.Data != nil && cert.Data.GetCm() != nil {
		cmID = cert.Data.GetCm().Id
	}

	certModel := SSLCertificateModel{
		Type:                 types.StringValue(certType),
		Status:               types.StringValue(status),
		CertificateManagerID: types.StringValue(cmID),
	}

	certSet, d := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type":                   types.StringType,
			"status":                 types.StringType,
			"certificate_manager_id": types.StringType,
		},
	}, []SSLCertificateModel{certModel})
	diags.Append(d...)

	return certSet
}

// getCDNOptionsAttrTypes returns the attribute types for CDNOptionsModel
// This is used for creating types.List from CDNOptionsModel
func getCDNOptionsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		// Boolean options
		"ignore_query_params":        types.BoolType,
		"slice":                      types.BoolType,
		"fetched_compressed":         types.BoolType,
		"gzip_on":                    types.BoolType,
		"redirect_http_to_https":     types.BoolType,
		"redirect_https_to_http":     types.BoolType,
		"forward_host_header":        types.BoolType,
		"proxy_cache_methods_set":    types.BoolType,
		"disable_proxy_force_ranges": types.BoolType,
		"ignore_cookie":              types.BoolType,
		"enable_ip_url_signing":      types.BoolType,

		// Cache settings - nested blocks
		"edge_cache_settings": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"enabled":    types.BoolType,
					"cache_time": types.MapType{ElemType: types.Int64Type},
				},
			},
		},
		"browser_cache_settings": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"enabled":    types.BoolType,
					"cache_time": types.Int64Type,
				},
			},
		},

		// String options
		"custom_host_header": types.StringType,
		"custom_server_name": types.StringType,
		"secure_key":         types.StringType,

		// List options
		"cache_http_headers":     types.ListType{ElemType: types.StringType},
		"query_params_whitelist": types.ListType{ElemType: types.StringType},
		"query_params_blacklist": types.ListType{ElemType: types.StringType},
		"cors":                   types.ListType{ElemType: types.StringType},
		"allowed_http_methods":   types.ListType{ElemType: types.StringType},

		// Map options
		"static_response_headers": types.MapType{ElemType: types.StringType},
		"static_request_headers":  types.MapType{ElemType: types.StringType},

		// Nested objects
		"ip_address_acl": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"policy_type":     types.StringType,
					"excepted_values": types.ListType{ElemType: types.StringType},
				},
			},
		},
		"rewrite": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"enabled": types.BoolType,
					"body":    types.StringType,
					"flag":    types.StringType,
				},
			},
		},
	}
}
