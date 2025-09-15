package cdn_resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

// ExpandCDNResourceOptions converts Terraform plan options to CDN API ResourceOptions
// CRITICAL: This function properly handles tristate booleans using types.Bool.IsNull()
// Exported for reuse in cdn_rule package
func ExpandCDNResourceOptions(ctx context.Context, planOptions []CDNOptionsModel, diags *diag.Diagnostics) *cdn.ResourceOptions {
	if len(planOptions) == 0 {
		return nil
	}

	opt := planOptions[0]
	result := &cdn.ResourceOptions{}

	// Boolean options - CRITICAL: Check IsNull() to distinguish unset from false
	if !opt.Slice.IsNull() {
		result.Slice = &cdn.ResourceOptions_BoolOption{
			Enabled: true,
			Value:   opt.Slice.ValueBool(),
		}
	}

	if !opt.IgnoreCookie.IsNull() {
		result.IgnoreCookie = &cdn.ResourceOptions_BoolOption{
			Enabled: true,
			Value:   opt.IgnoreCookie.ValueBool(),
		}
	}

	if !opt.ProxyCacheMethodsSet.IsNull() {
		result.ProxyCacheMethodsSet = &cdn.ResourceOptions_BoolOption{
			Enabled: true,
			Value:   opt.ProxyCacheMethodsSet.ValueBool(),
		}
	}

	if !opt.DisableProxyForceRanges.IsNull() {
		result.DisableProxyForceRanges = &cdn.ResourceOptions_BoolOption{
			Enabled: true,
			Value:   opt.DisableProxyForceRanges.ValueBool(),
		}
	}

	// Cache settings - nested blocks
	expandEdgeCacheSettings(ctx, opt.EdgeCacheSettings, result, diags)
	expandBrowserCacheSettings(ctx, opt.BrowserCacheSettings, result, diags)

	// String options
	if !opt.CustomServerName.IsNull() && opt.CustomServerName.ValueString() != "" {
		result.CustomServerName = &cdn.ResourceOptions_StringOption{
			Enabled: true,
			Value:   opt.CustomServerName.ValueString(),
		}
	}

	// SecureKey - combines secure_key and enable_ip_url_signing
	if !opt.SecureKey.IsNull() && opt.SecureKey.ValueString() != "" {
		urlType := cdn.SecureKeyURLType_DISABLE_IP_SIGNING
		if !opt.EnableIPURLSigning.IsNull() && opt.EnableIPURLSigning.ValueBool() {
			urlType = cdn.SecureKeyURLType_ENABLE_IP_SIGNING
		}

		result.SecureKey = &cdn.ResourceOptions_SecureKeyOption{
			Enabled: true,
			Key:     opt.SecureKey.ValueString(),
			Type:    urlType,
		}
	}

	// List options
	if !opt.CacheHTTPHeaders.IsNull() && !opt.CacheHTTPHeaders.IsUnknown() && len(opt.CacheHTTPHeaders.Elements()) > 0 {
		var headers []string
		diags.Append(opt.CacheHTTPHeaders.ElementsAs(ctx, &headers, false)...)
		if !diags.HasError() && len(headers) > 0 {
			result.CacheHttpHeaders = &cdn.ResourceOptions_StringsListOption{
				Enabled: true,
				Value:   headers,
			}
		}
	}

	if !opt.Cors.IsNull() && !opt.Cors.IsUnknown() && len(opt.Cors.Elements()) > 0 {
		var corsOrigins []string
		diags.Append(opt.Cors.ElementsAs(ctx, &corsOrigins, false)...)
		if !diags.HasError() && len(corsOrigins) > 0 {
			result.Cors = &cdn.ResourceOptions_StringsListOption{
				Enabled: true,
				Value:   corsOrigins,
			}
		}
	}

	if !opt.AllowedHTTPMethods.IsNull() && !opt.AllowedHTTPMethods.IsUnknown() && len(opt.AllowedHTTPMethods.Elements()) > 0 {
		var methods []string
		diags.Append(opt.AllowedHTTPMethods.ElementsAs(ctx, &methods, false)...)
		if !diags.HasError() && len(methods) > 0 {
			result.AllowedHttpMethods = &cdn.ResourceOptions_StringsListOption{
				Enabled: true,
				Value:   methods,
			}
		}
	}

	// Map options
	if !opt.StaticResponseHeaders.IsNull() && !opt.StaticResponseHeaders.IsUnknown() && len(opt.StaticResponseHeaders.Elements()) > 0 {
		headers := make(map[string]string)
		diags.Append(opt.StaticResponseHeaders.ElementsAs(ctx, &headers, false)...)
		if !diags.HasError() && len(headers) > 0 {
			result.StaticHeaders = &cdn.ResourceOptions_StringsMapOption{
				Enabled: true,
				Value:   headers,
			}
		}
	}

	if !opt.StaticRequestHeaders.IsNull() && !opt.StaticRequestHeaders.IsUnknown() && len(opt.StaticRequestHeaders.Elements()) > 0 {
		headers := make(map[string]string)
		diags.Append(opt.StaticRequestHeaders.ElementsAs(ctx, &headers, false)...)
		if !diags.HasError() && len(headers) > 0 {
			result.StaticRequestHeaders = &cdn.ResourceOptions_StringsMapOption{
				Enabled: true,
				Value:   headers,
			}
		}
	}

	// Mutually exclusive options groups
	expandHostOptions(&opt, result, diags)
	expandQueryParamsOptions(ctx, &opt, result, diags)
	expandCompressionOptions(&opt, result, diags)
	expandRedirectOptions(&opt, result, diags)

	// Nested blocks
	expandIPAddressACL(ctx, opt.IPAddressACL, result, diags)
	expandRewrite(ctx, opt.Rewrite, result, diags)

	return result
}

// expandHostOptions handles mutually exclusive forward_host_header and custom_host_header
// CRITICAL FIX: This properly handles forward_host_header tristate using IsNull()
func expandHostOptions(opt *CDNOptionsModel, result *cdn.ResourceOptions, diags *diag.Diagnostics) {
	// custom_host_header takes precedence over forward_host_header
	if !opt.CustomHostHeader.IsNull() && opt.CustomHostHeader.ValueString() != "" {
		result.HostOptions = &cdn.ResourceOptions_HostOptions{
			HostVariant: &cdn.ResourceOptions_HostOptions_Host{
				Host: &cdn.ResourceOptions_StringOption{
					Enabled: true,
					Value:   opt.CustomHostHeader.ValueString(),
				},
			},
		}
	} else if !opt.ForwardHostHeader.IsNull() {
		// CRITICAL: This is the fix for the forward_host_header bug
		// In SDKv2, GetOk couldn't distinguish between "not set" and "false"
		// In Framework, IsNull() perfectly handles this:
		//   - IsNull() == true  → user didn't set it, don't send to API
		//   - IsNull() == false → user set it (true or false), send Enabled=true
		result.HostOptions = &cdn.ResourceOptions_HostOptions{
			HostVariant: &cdn.ResourceOptions_HostOptions_ForwardHostHeader{
				ForwardHostHeader: &cdn.ResourceOptions_BoolOption{
					Enabled: true,
					Value:   opt.ForwardHostHeader.ValueBool(),
				},
			},
		}
	}
}

// expandQueryParamsOptions handles mutually exclusive query params options
func expandQueryParamsOptions(ctx context.Context, opt *CDNOptionsModel, result *cdn.ResourceOptions, diags *diag.Diagnostics) {
	if !opt.IgnoreQueryParams.IsNull() {
		result.QueryParamsOptions = &cdn.ResourceOptions_QueryParamsOptions{
			QueryParamsVariant: &cdn.ResourceOptions_QueryParamsOptions_IgnoreQueryString{
				IgnoreQueryString: &cdn.ResourceOptions_BoolOption{
					Enabled: true,
					Value:   opt.IgnoreQueryParams.ValueBool(),
				},
			},
		}
	} else if !opt.QueryParamsWhitelist.IsNull() && !opt.QueryParamsWhitelist.IsUnknown() && len(opt.QueryParamsWhitelist.Elements()) > 0 {
		var params []string
		diags.Append(opt.QueryParamsWhitelist.ElementsAs(ctx, &params, false)...)
		if !diags.HasError() && len(params) > 0 {
			result.QueryParamsOptions = &cdn.ResourceOptions_QueryParamsOptions{
				QueryParamsVariant: &cdn.ResourceOptions_QueryParamsOptions_QueryParamsWhitelist{
					QueryParamsWhitelist: &cdn.ResourceOptions_StringsListOption{
						Enabled: true,
						Value:   params,
					},
				},
			}
		}
	} else if !opt.QueryParamsBlacklist.IsNull() && !opt.QueryParamsBlacklist.IsUnknown() && len(opt.QueryParamsBlacklist.Elements()) > 0 {
		var params []string
		diags.Append(opt.QueryParamsBlacklist.ElementsAs(ctx, &params, false)...)
		if !diags.HasError() && len(params) > 0 {
			result.QueryParamsOptions = &cdn.ResourceOptions_QueryParamsOptions{
				QueryParamsVariant: &cdn.ResourceOptions_QueryParamsOptions_QueryParamsBlacklist{
					QueryParamsBlacklist: &cdn.ResourceOptions_StringsListOption{
						Enabled: true,
						Value:   params,
					},
				},
			}
		}
	}
}

// expandCompressionOptions handles mutually exclusive gzip_on and fetched_compressed
func expandCompressionOptions(opt *CDNOptionsModel, result *cdn.ResourceOptions, diags *diag.Diagnostics) {
	if !opt.FetchedCompressed.IsNull() {
		result.CompressionOptions = &cdn.ResourceOptions_CompressionOptions{
			CompressionVariant: &cdn.ResourceOptions_CompressionOptions_FetchCompressed{
				FetchCompressed: &cdn.ResourceOptions_BoolOption{
					Enabled: true,
					Value:   opt.FetchedCompressed.ValueBool(),
				},
			},
		}
	} else if !opt.GzipOn.IsNull() {
		result.CompressionOptions = &cdn.ResourceOptions_CompressionOptions{
			CompressionVariant: &cdn.ResourceOptions_CompressionOptions_GzipOn{
				GzipOn: &cdn.ResourceOptions_BoolOption{
					Enabled: true,
					Value:   opt.GzipOn.ValueBool(),
				},
			},
		}
	}
}

// expandRedirectOptions handles mutually exclusive redirect options
func expandRedirectOptions(opt *CDNOptionsModel, result *cdn.ResourceOptions, diags *diag.Diagnostics) {
	if !opt.RedirectHttpToHttps.IsNull() {
		result.RedirectOptions = &cdn.ResourceOptions_RedirectOptions{
			RedirectVariant: &cdn.ResourceOptions_RedirectOptions_RedirectHttpToHttps{
				RedirectHttpToHttps: &cdn.ResourceOptions_BoolOption{
					Enabled: true,
					Value:   opt.RedirectHttpToHttps.ValueBool(),
				},
			},
		}
	} else if !opt.RedirectHttpsToHttp.IsNull() {
		result.RedirectOptions = &cdn.ResourceOptions_RedirectOptions{
			RedirectVariant: &cdn.ResourceOptions_RedirectOptions_RedirectHttpsToHttp{
				RedirectHttpsToHttp: &cdn.ResourceOptions_BoolOption{
					Enabled: true,
					Value:   opt.RedirectHttpsToHttp.ValueBool(),
				},
			},
		}
	}
}

// expandIPAddressACL converts IP address ACL block to API format
func expandIPAddressACL(ctx context.Context, aclList types.List, result *cdn.ResourceOptions, diags *diag.Diagnostics) {
	if aclList.IsNull() || aclList.IsUnknown() || len(aclList.Elements()) == 0 {
		return
	}

	var aclModels []IPAddressACLModel
	diags.Append(aclList.ElementsAs(ctx, &aclModels, false)...)
	if diags.HasError() || len(aclModels) == 0 {
		return
	}

	acl := aclModels[0]
	if acl.ExceptedValues.IsNull() || acl.ExceptedValues.IsUnknown() {
		return
	}

	var exceptedValues []string
	diags.Append(acl.ExceptedValues.ElementsAs(ctx, &exceptedValues, false)...)
	if diags.HasError() {
		return
	}

	var policyType cdn.PolicyType
	switch acl.PolicyType.ValueString() {
	case "allow":
		policyType = cdn.PolicyType_POLICY_TYPE_ALLOW
	case "deny":
		policyType = cdn.PolicyType_POLICY_TYPE_DENY
	default:
		diags.AddError(
			"Invalid ACL policy type",
			fmt.Sprintf("policy_type must be 'allow' or 'deny', got: %s", acl.PolicyType.ValueString()),
		)
		return
	}

	result.IpAddressAcl = &cdn.ResourceOptions_IPAddressACLOption{
		Enabled:        true,
		PolicyType:     policyType,
		ExceptedValues: exceptedValues,
	}
}

// expandRewrite converts rewrite block to API format
func expandRewrite(ctx context.Context, rewriteList types.List, result *cdn.ResourceOptions, diags *diag.Diagnostics) {
	if rewriteList.IsNull() || rewriteList.IsUnknown() || len(rewriteList.Elements()) == 0 {
		return
	}

	var rewriteModels []RewriteModel
	diags.Append(rewriteList.ElementsAs(ctx, &rewriteModels, false)...)
	if diags.HasError() || len(rewriteModels) == 0 {
		return
	}

	rewrite := rewriteModels[0]

	// Determine rewrite flag
	var flag cdn.RewriteFlag
	flagStr := "break" // default
	if !rewrite.Flag.IsNull() && rewrite.Flag.ValueString() != "" {
		flagStr = rewrite.Flag.ValueString()
	}

	switch flagStr {
	case "last":
		flag = cdn.RewriteFlag_LAST
	case "break":
		flag = cdn.RewriteFlag_BREAK
	case "redirect":
		flag = cdn.RewriteFlag_REDIRECT
	case "permanent":
		flag = cdn.RewriteFlag_PERMANENT
	default:
		diags.AddError(
			"Invalid rewrite flag",
			fmt.Sprintf("flag must be one of: last, break, redirect, permanent, got: %s", flagStr),
		)
		return
	}

	// Determine enabled status
	enabled := false
	if !rewrite.Enabled.IsNull() {
		enabled = rewrite.Enabled.ValueBool()
	}

	result.Rewrite = &cdn.ResourceOptions_RewriteOption{
		Enabled: enabled,
		Body:    rewrite.Body.ValueString(),
		Flag:    flag,
	}
}

// expandEdgeCacheSettings converts edge_cache_settings block to API format
// Handles two modes:
// 1. cache_time = {"*" = 3600} → DefaultValue variant
// 2. cache_time = {"200" = 3600, "404" = 300} → Value.CustomValues variant
func expandEdgeCacheSettings(ctx context.Context, edgeCacheList types.List, result *cdn.ResourceOptions, diags *diag.Diagnostics) {
	if edgeCacheList.IsNull() || edgeCacheList.IsUnknown() || len(edgeCacheList.Elements()) == 0 {
		return
	}

	var edgeCacheModels []EdgeCacheSettingsModel
	diags.Append(edgeCacheList.ElementsAs(ctx, &edgeCacheModels, false)...)
	if diags.HasError() || len(edgeCacheModels) == 0 {
		return
	}

	edgeCache := edgeCacheModels[0]

	// Determine enabled status (defaults to true if not set)
	enabled := true
	if !edgeCache.Enabled.IsNull() && !edgeCache.Enabled.IsUnknown() {
		enabled = edgeCache.Enabled.ValueBool()
	}

	// If enabled=false, don't send block to API at all
	// The API doesn't support storing disabled state - it simply omits the block
	// Users should remove the block from config to disable the feature
	if !enabled {
		return
	}

	// enabled=true, process cache_time
	if edgeCache.CacheTime.IsNull() || edgeCache.CacheTime.IsUnknown() || len(edgeCache.CacheTime.Elements()) == 0 {
		// This should not happen due to validator, but handle gracefully
		return
	}

	cacheTimeMap := make(map[string]int64)
	diags.Append(edgeCache.CacheTime.ElementsAs(ctx, &cacheTimeMap, false)...)
	if diags.HasError() {
		return
	}

	// Check if wildcard "*" is present
	if defaultValue, hasWildcard := cacheTimeMap["*"]; hasWildcard {
		// Simple mode: use DefaultValue variant
		result.EdgeCacheSettings = &cdn.ResourceOptions_EdgeCacheSettings{
			Enabled: true,
			ValuesVariant: &cdn.ResourceOptions_EdgeCacheSettings_DefaultValue{
				DefaultValue: defaultValue,
			},
		}
	} else {
		// Advanced mode: use Value.CustomValues variant
		result.EdgeCacheSettings = &cdn.ResourceOptions_EdgeCacheSettings{
			Enabled: true,
			ValuesVariant: &cdn.ResourceOptions_EdgeCacheSettings_Value{
				Value: &cdn.ResourceOptions_CachingTimes{
					CustomValues: cacheTimeMap,
				},
			},
		}
	}
}

// expandBrowserCacheSettings converts browser_cache_settings block to API format
func expandBrowserCacheSettings(ctx context.Context, browserCacheList types.List, result *cdn.ResourceOptions, diags *diag.Diagnostics) {
	if browserCacheList.IsNull() || browserCacheList.IsUnknown() || len(browserCacheList.Elements()) == 0 {
		return
	}

	var browserCacheModels []BrowserCacheSettingsModel
	diags.Append(browserCacheList.ElementsAs(ctx, &browserCacheModels, false)...)
	if diags.HasError() || len(browserCacheModels) == 0 {
		return
	}

	browserCache := browserCacheModels[0]

	// Determine enabled status (defaults to true if not set)
	enabled := true
	if !browserCache.Enabled.IsNull() && !browserCache.Enabled.IsUnknown() {
		enabled = browserCache.Enabled.ValueBool()
	}

	// If enabled=false, don't send block to API at all
	// The API doesn't support storing disabled state - it simply omits the block
	// Users should remove the block from config to disable the feature
	if !enabled {
		return
	}

	// enabled=true, process cache_time
	if browserCache.CacheTime.IsNull() || browserCache.CacheTime.IsUnknown() {
		// This should not happen due to validator, but handle gracefully
		return
	}

	result.BrowserCacheSettings = &cdn.ResourceOptions_Int64Option{
		Enabled: true,
		Value:   browserCache.CacheTime.ValueInt64(),
	}
}
