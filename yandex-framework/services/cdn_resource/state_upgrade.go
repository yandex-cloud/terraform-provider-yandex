package cdn_resource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// CDNOptionsModelV0 represents the old schema version 0 structure
type CDNOptionsModelV0 struct {
	// Boolean options
	IgnoreQueryParams       types.Bool `tfsdk:"ignore_query_params"`
	Slice                   types.Bool `tfsdk:"slice"`
	FetchedCompressed       types.Bool `tfsdk:"fetched_compressed"`
	GzipOn                  types.Bool `tfsdk:"gzip_on"`
	RedirectHttpToHttps     types.Bool `tfsdk:"redirect_http_to_https"`
	RedirectHttpsToHttp     types.Bool `tfsdk:"redirect_https_to_http"`
	ForwardHostHeader       types.Bool `tfsdk:"forward_host_header"`
	ProxyCacheMethodsSet    types.Bool `tfsdk:"proxy_cache_methods_set"`
	DisableProxyForceRanges types.Bool `tfsdk:"disable_proxy_force_ranges"`
	IgnoreCookie            types.Bool `tfsdk:"ignore_cookie"`
	EnableIPURLSigning      types.Bool `tfsdk:"enable_ip_url_signing"`

	// OLD: Integer options as simple Int64
	EdgeCacheSettings    types.Int64 `tfsdk:"edge_cache_settings"`
	BrowserCacheSettings types.Int64 `tfsdk:"browser_cache_settings"`

	// String options
	CustomHostHeader types.String `tfsdk:"custom_host_header"`
	CustomServerName types.String `tfsdk:"custom_server_name"`
	SecureKey        types.String `tfsdk:"secure_key"`

	// List options
	CacheHTTPHeaders     types.List `tfsdk:"cache_http_headers"`
	QueryParamsWhitelist types.List `tfsdk:"query_params_whitelist"`
	QueryParamsBlacklist types.List `tfsdk:"query_params_blacklist"`
	Cors                 types.List `tfsdk:"cors"`
	AllowedHTTPMethods   types.List `tfsdk:"allowed_http_methods"`

	// Map options
	StaticResponseHeaders types.Map `tfsdk:"static_response_headers"`
	StaticRequestHeaders  types.Map `tfsdk:"static_request_headers"`

	// Nested objects
	IPAddressACL types.List `tfsdk:"ip_address_acl"`
	Rewrite      types.List `tfsdk:"rewrite"`
}

// CDNResourceModelV0 represents the old schema version 0
type CDNResourceModelV0 struct {
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
	ID                 types.String   `tfsdk:"id"`
	Cname              types.String   `tfsdk:"cname"`
	ProviderType       types.String   `tfsdk:"provider_type"`
	FolderID           types.String   `tfsdk:"folder_id"`
	Labels             types.Map      `tfsdk:"labels"`
	Active             types.Bool     `tfsdk:"active"`
	SecondaryHostnames types.Set      `tfsdk:"secondary_hostnames"`
	OriginProtocol     types.String   `tfsdk:"origin_protocol"`
	CreatedAt          types.String   `tfsdk:"created_at"`
	UpdatedAt          types.String   `tfsdk:"updated_at"`
	OriginGroupID      types.String   `tfsdk:"origin_group_id"`
	OriginGroupName    types.String   `tfsdk:"origin_group_name"`
	SSLCertificate     types.Set      `tfsdk:"ssl_certificate"`
	ProviderCname      types.String   `tfsdk:"provider_cname"`
	Options            types.List     `tfsdk:"options"` // List of CDNOptionsModelV0
}

// upgradeStateV0ToV1 migrates schema version 0 to version 1
// Changes:
// - edge_cache_settings: Int64 -> List[{enabled, cache_time: Map}]
// - browser_cache_settings: Int64 -> List[{enabled, cache_time: Int64}]
// - disable_cache: removed (deprecated attribute)
func upgradeStateV0ToV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	tflog.Debug(ctx, "Upgrading CDN resource state from v0 to v1")

	// Parse RawState JSON
	type rawState map[string]interface{}
	var state rawState

	err := json.Unmarshal(req.RawState.JSON, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse Prior State",
			fmt.Sprintf("Error parsing state JSON: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Parsed state: %+v", state))

	// Transform edge_cache_settings and browser_cache_settings within options
	if optionsRaw, ok := state["options"].([]interface{}); ok && len(optionsRaw) > 0 {
		if options, ok := optionsRaw[0].(map[string]interface{}); ok {
			// Upgrade edge_cache_settings
			if val, exists := options["edge_cache_settings"]; exists && val != nil {
				switch v := val.(type) {
				case float64:
					tflog.Debug(ctx, fmt.Sprintf("Migrating edge_cache_settings from %v to nested block", v))
					options["edge_cache_settings"] = []interface{}{
						map[string]interface{}{
							"enabled":    true,
							"cache_time": map[string]interface{}{"*": v},
						},
					}
				}
			}

			// Upgrade browser_cache_settings
			// Note: 0 means disabled in old schema, >0 means enabled with value
			if val, exists := options["browser_cache_settings"]; exists && val != nil {
				switch v := val.(type) {
				case float64:
					if v == 0 {
						// 0 = disabled in old schema → remove block in new schema
						tflog.Debug(ctx, "browser_cache_settings was 0 (disabled), removing from state")
						delete(options, "browser_cache_settings")
					} else {
						// >0 = enabled with value
						tflog.Debug(ctx, fmt.Sprintf("Migrating browser_cache_settings from %v to nested block", v))
						options["browser_cache_settings"] = []interface{}{
							map[string]interface{}{
								"enabled":    true,
								"cache_time": v,
							},
						}
					}
				}
			}

			// Remove deprecated disable_cache attribute
			// This attribute was removed in schema v1
			if _, exists := options["disable_cache"]; exists {
				tflog.Debug(ctx, "Removing deprecated disable_cache attribute from state")
				delete(options, "disable_cache")
			}
		}
	}

	// Marshal back to JSON
	upgradedJSON, err := json.Marshal(state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Marshal Upgraded State",
			fmt.Sprintf("Error marshaling upgraded state: %s", err.Error()),
		)
		return
	}

	// Unmarshal into tftypes.Value using the new schema
	schema := CDNResourceSchema(ctx)
	schemaType := schema.Type().TerraformType(ctx)

	upgradedValue, err := tftypes.ValueFromJSON(upgradedJSON, schemaType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create State Value",
			fmt.Sprintf("Error creating state value: %s", err.Error()),
		)
		return
	}

	// Set the new state
	resp.State = tfsdk.State{
		Raw:    upgradedValue,
		Schema: schema,
	}

	tflog.Debug(ctx, "Successfully upgraded state from v0 to v1")
}

// upgradeStateV1ToV2 migrates schema version 1 to version 2
// Changes:
// - edge_cache_settings: cache_time (Map) → value (Int64) + custom_values (Map)
//   - cache_time = {"*" = X} → value = X (SimpleValue for success codes)
//   - cache_time = {specific codes} → custom_values = {same}
//
// This migration is required for full transition to new API matching master
// (commit 042b2e91: edge_cache_settings with caching by http code)
func upgradeStateV1ToV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	tflog.Debug(ctx, "Upgrading CDN resource state from v1 to v2")

	// Parse RawState JSON
	type rawState map[string]interface{}
	var state rawState

	err := json.Unmarshal(req.RawState.JSON, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse Prior State",
			fmt.Sprintf("Error parsing state JSON: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Parsed state for v1→v2 migration")

	// Navigate to options.edge_cache_settings and convert cache_time → value + custom_values
	if optionsRaw, ok := state["options"].([]interface{}); ok && len(optionsRaw) > 0 {
		if options, ok := optionsRaw[0].(map[string]interface{}); ok {
			if edgeCacheRaw, exists := options["edge_cache_settings"]; exists && edgeCacheRaw != nil {
				if edgeCacheList, ok := edgeCacheRaw.([]interface{}); ok && len(edgeCacheList) > 0 {
					if edgeCache, ok := edgeCacheList[0].(map[string]interface{}); ok {
						if cacheTimeRaw, exists := edgeCache["cache_time"]; exists && cacheTimeRaw != nil {
							if cacheTimeMap, ok := cacheTimeRaw.(map[string]interface{}); ok {
								tflog.Debug(ctx, "Migrating edge_cache_settings: cache_time → value + custom_values")

								// Check if "*" key exists (legacy format)
								if starValue, hasStarKey := cacheTimeMap["*"]; hasStarKey {
									// "*" → value (SimpleValue)
									edgeCache["value"] = starValue
									delete(cacheTimeMap, "*")

									// If there are other keys, put them in custom_values
									if len(cacheTimeMap) > 0 {
										edgeCache["custom_values"] = cacheTimeMap
									} else {
										edgeCache["custom_values"] = nil
									}
								} else {
									// No "*" key → all goes to custom_values
									edgeCache["value"] = nil
									edgeCache["custom_values"] = cacheTimeMap
								}

								// Remove old cache_time field
								delete(edgeCache, "cache_time")
							}
						}
					}
				}
			}
		}
	}

	// Marshal back to JSON
	upgradedJSON, err := json.Marshal(state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Marshal Upgraded State",
			fmt.Sprintf("Error marshaling upgraded state: %s", err.Error()),
		)
		return
	}

	// Unmarshal into tftypes.Value using the new schema
	schema := CDNResourceSchema(ctx)
	schemaType := schema.Type().TerraformType(ctx)

	upgradedValue, err := tftypes.ValueFromJSON(upgradedJSON, schemaType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create State Value",
			fmt.Sprintf("Error creating state value: %s", err.Error()),
		)
		return
	}

	// Set the new state
	resp.State = tfsdk.State{
		Raw:    upgradedValue,
		Schema: schema,
	}

	tflog.Debug(ctx, "Successfully upgraded state from v1 to v2")
}
