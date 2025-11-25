package cdn_resource

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// rewriteBodyValidator validates the format of rewrite body pattern
type rewriteBodyValidator struct{}

func NewRewriteBodyValidator() validator.String {
	return &rewriteBodyValidator{}
}

func (v *rewriteBodyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	// Check if the string contains at least two parts separated by space
	parts := strings.Fields(value)
	if len(parts) != 2 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			"Invalid rewrite body format",
			fmt.Sprintf("Must have format '<source path> <destination path>' (e.g., '/foo/(.*) /bar/$1'), got: %q", value),
		))
		return
	}

	// Basic validation that source and destination paths start with / or are regex patterns
	source, destination := parts[0], parts[1]

	// Check if source starts with ^ (regex anchor) or / (path)
	if !strings.HasPrefix(source, "^") && !strings.HasPrefix(source, "/") {
		resp.Diagnostics.AddAttributeWarning(
			req.Path,
			"Potentially incorrect rewrite source",
			fmt.Sprintf("Source path %q should start with '^' for regex or '/' for path", source),
		)
	}

	// Check if destination starts with / or $ (for variables like $scheme)
	if !strings.HasPrefix(destination, "/") && !strings.HasPrefix(destination, "$") &&
		!strings.HasPrefix(destination, "http://") && !strings.HasPrefix(destination, "https://") {
		resp.Diagnostics.AddAttributeWarning(
			req.Path,
			"Potentially incorrect rewrite destination",
			fmt.Sprintf("Destination path %q should start with '/', '$', 'http://' or 'https://'", destination),
		)
	}
}

func (v *rewriteBodyValidator) Description(_ context.Context) string {
	return "Validates rewrite body format: '<source path> <destination path>'"
}

func (v *rewriteBodyValidator) MarkdownDescription(_ context.Context) string {
	return "Validates rewrite body format: `<source path> <destination path>`"
}

// ipAddressOrCIDRValidator validates that a string is a valid IP address or CIDR notation
type ipAddressOrCIDRValidator struct{}

func NewIPAddressOrCIDRValidator() validator.String {
	return &ipAddressOrCIDRValidator{}
}

func (v *ipAddressOrCIDRValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	// Try parsing as CIDR first
	if _, _, err := net.ParseCIDR(value); err == nil {
		return
	}

	// Try parsing as IP address
	if ip := net.ParseIP(value); ip != nil {
		return
	}

	resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
		req.Path,
		"Invalid IP address or CIDR",
		fmt.Sprintf("Must be a valid IP address (e.g., 192.168.1.1) or CIDR notation (e.g., 192.168.1.0/24), got: %q", value),
	))
}

func (v *ipAddressOrCIDRValidator) Description(_ context.Context) string {
	return "Validates that string is a valid IP address or CIDR notation"
}

func (v *ipAddressOrCIDRValidator) MarkdownDescription(_ context.Context) string {
	return "Validates that string is a valid IP address or CIDR notation"
}

// staticHeadersValidator validates static HTTP headers
type staticHeadersValidator struct{}

func NewStaticHeadersValidator() validator.Map {
	return &staticHeadersValidator{}
}

func (v *staticHeadersValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Forbidden headers according to CDN standards - using map for O(1) lookup
	forbiddenHeaders := map[string]bool{
		"host":                true,
		"content-length":      true,
		"transfer-encoding":   true,
		"connection":          true,
		"keep-alive":          true,
		"proxy-authenticate":  true,
		"proxy-authorization": true,
		"te":                  true,
		"trailer":             true,
		"upgrade":             true,
	}

	headers := make(map[string]string)
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &headers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	headerNameRegex := regexp.MustCompile(`^[A-Za-z0-9\-]+$`)

	for key := range headers {
		// Check for forbidden headers - O(1) lookup
		if forbiddenHeaders[strings.ToLower(key)] {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				"Forbidden header",
				fmt.Sprintf("Header '%s' cannot be set as static header", key),
			))
		}

		// Validate header name format
		if !headerNameRegex.MatchString(key) {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				"Invalid header name",
				fmt.Sprintf("Header name '%s' contains invalid characters (must be alphanumeric or hyphen)", key),
			))
		}
	}
}

func (v *staticHeadersValidator) Description(_ context.Context) string {
	return "Validates static HTTP headers (forbidden headers and name format)"
}

func (v *staticHeadersValidator) MarkdownDescription(_ context.Context) string {
	return "Validates static HTTP headers (forbidden headers and name format)"
}

// cdnOptionsValidator validates CDN options logic
type cdnOptionsValidator struct{}

func NewCDNOptionsValidator() validator.Object {
	return &cdnOptionsValidator{}
}

func (v *cdnOptionsValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var edgeCacheSettingsList types.List
	var browserCacheSettingsList types.List

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("edge_cache_settings"), &edgeCacheSettingsList)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("browser_cache_settings"), &browserCacheSettingsList)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Both settings must be configured for validation
	if edgeCacheSettingsList.IsNull() || browserCacheSettingsList.IsNull() ||
		len(edgeCacheSettingsList.Elements()) == 0 || len(browserCacheSettingsList.Elements()) == 0 {
		return
	}

	// Extract edge_cache_settings
	var edgeSettings []EdgeCacheSettingsModel
	diags := edgeCacheSettingsList.ElementsAs(ctx, &edgeSettings, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || len(edgeSettings) == 0 {
		return
	}

	// Extract browser_cache_settings
	var browserSettings []BrowserCacheSettingsModel
	diags = browserCacheSettingsList.ElementsAs(ctx, &browserSettings, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || len(browserSettings) == 0 {
		return
	}

	edgeSetting := edgeSettings[0]
	browserSetting := browserSettings[0]

	// Skip validation if either is disabled
	if (!edgeSetting.Enabled.IsNull() && !edgeSetting.Enabled.ValueBool()) ||
		(!browserSetting.Enabled.IsNull() && !browserSetting.Enabled.ValueBool()) {
		return
	}

	// Get browser cache time
	if browserSetting.CacheTime.IsNull() {
		return
	}
	browserCacheTime := browserSetting.CacheTime.ValueInt64()

	// Get edge cache time - only validate if using simple value mode
	// For custom per-code settings (custom_values), validation is not meaningful
	if !edgeSetting.Value.IsNull() && !edgeSetting.Value.IsUnknown() {
		edgeCacheTime := edgeSetting.Value.ValueInt64()
		if browserCacheTime > edgeCacheTime {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid CDN options logic",
				fmt.Sprintf("browser_cache_settings.cache_time (%d) cannot be greater than edge_cache_settings.value (%d)", browserCacheTime, edgeCacheTime),
			)
		}
	}
}

func (v *cdnOptionsValidator) Description(_ context.Context) string {
	return "Validates CDN options logic (edge_cache_settings, browser_cache_settings relationships)"
}

func (v *cdnOptionsValidator) MarkdownDescription(_ context.Context) string {
	return "Validates CDN options logic (edge_cache_settings, browser_cache_settings relationships)"
}

// edgeCacheSettingsValidator validates that exactly one of default_value or custom_values is set
type edgeCacheSettingsValidator struct{}

func NewEdgeCacheSettingsValidator() validator.List {
	return &edgeCacheSettingsValidator{}
}

func (v *edgeCacheSettingsValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}

	// Get the single element (MaxItems: 1)
	var elements []types.Object
	diags := req.ConfigValue.ElementsAs(ctx, &elements, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(elements) == 0 {
		return
	}

	elem := elements[0]
	if elem.IsNull() || elem.IsUnknown() {
		return
	}

	// Extract enabled and cache_time from the element
	var edgeSettings EdgeCacheSettingsModel
	diags = elem.As(ctx, &edgeSettings, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasValue := !edgeSettings.Value.IsNull() && !edgeSettings.Value.IsUnknown()
	hasCustomValues := !edgeSettings.CustomValues.IsNull() && !edgeSettings.CustomValues.IsUnknown() && len(edgeSettings.CustomValues.Elements()) > 0
	isEnabled := !edgeSettings.Enabled.IsNull() && !edgeSettings.Enabled.IsUnknown() && edgeSettings.Enabled.ValueBool()
	hasDefaultValue := !edgeSettings.DefaultValue.IsNull() && !edgeSettings.DefaultValue.IsUnknown()

	// Validation logic:
	// 1. If enabled=false → value, custom_values or default_value should not be set
	// 2. If enabled=true or not set → at least value, custom_values or default_value must be set
	if !isEnabled && (hasValue || hasCustomValues || hasDefaultValue) {
		resp.Diagnostics.AddError(
			"Invalid edge_cache_settings configuration",
			"When enabled=false, value, custom_values or default_value should not be specified",
		)
	} else if isEnabled {
		if !hasValue && !hasCustomValues && !hasDefaultValue {
			resp.Diagnostics.AddError(
				"Invalid edge_cache_settings configuration",
				"When enabled=true, at least value, custom_values or default_value must be specified",
			)
		}
	}
}

func (v *edgeCacheSettingsValidator) Description(_ context.Context) string {
	return "Validates edge_cache_settings: value and custom_values fields logic"
}

func (v *edgeCacheSettingsValidator) MarkdownDescription(_ context.Context) string {
	return "Validates `edge_cache_settings`: `value` and `custom_values` fields logic"
}

// browserCacheSettingsValidator validates browser_cache_settings logic
type browserCacheSettingsValidator struct{}

func NewBrowserCacheSettingsValidator() validator.List {
	return &browserCacheSettingsValidator{}
}

func (v *browserCacheSettingsValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}

	// Get the single element (MaxItems: 1)
	var elements []types.Object
	diags := req.ConfigValue.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(elements) == 0 {
		return
	}

	elem := elements[0]
	if elem.IsNull() || elem.IsUnknown() {
		return
	}

	// Extract enabled and cache_time from the element
	var browserSettings BrowserCacheSettingsModel
	diags = elem.As(ctx, &browserSettings, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasCacheTime := !browserSettings.CacheTime.IsNull() && !browserSettings.CacheTime.IsUnknown()
	isEnabled := browserSettings.Enabled.IsNull() || (!browserSettings.Enabled.IsUnknown() && browserSettings.Enabled.ValueBool())

	// Validation logic:
	// 1. If enabled=false → cache_time should not be set
	// 2. If enabled=true or not set → cache_time must be set
	if !browserSettings.Enabled.IsNull() && !browserSettings.Enabled.IsUnknown() && !browserSettings.Enabled.ValueBool() {
		// enabled = false
		if hasCacheTime {
			resp.Diagnostics.AddError(
				"Invalid browser_cache_settings configuration",
				"When enabled=false, cache_time should not be specified",
			)
		}
	} else if isEnabled {
		// enabled = true or not set
		if !hasCacheTime {
			resp.Diagnostics.AddError(
				"Invalid browser_cache_settings configuration",
				"When enabled=true, cache_time must be specified",
			)
		}
	}
}

func (v *browserCacheSettingsValidator) Description(_ context.Context) string {
	return "Validates browser_cache_settings: cache_time must be set when enabled=true"
}

func (v *browserCacheSettingsValidator) MarkdownDescription(_ context.Context) string {
	return "Validates `browser_cache_settings`: `cache_time` must be set when `enabled=true`"
}

// sliceFetchedCompressedValidator validates that slice and fetched_compressed are not both enabled
type sliceFetchedCompressedValidator struct{}

func NewSliceFetchedCompressedValidator() validator.List {
	return &sliceFetchedCompressedValidator{}
}

func (v *sliceFetchedCompressedValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}

	// Get the single element (MaxItems: 1)
	var elements []types.Object
	diags := req.ConfigValue.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(elements) == 0 {
		return
	}

	elem := elements[0]
	if elem.IsNull() || elem.IsUnknown() {
		return
	}

	// Extract slice and fetched_compressed from the element
	var options CDNOptionsModel
	diags = elem.As(ctx, &options, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if both are set to true
	if !options.Slice.IsNull() && !options.Slice.IsUnknown() && options.Slice.ValueBool() &&
		!options.FetchedCompressed.IsNull() && !options.FetchedCompressed.IsUnknown() && options.FetchedCompressed.ValueBool() {
		resp.Diagnostics.AddError(
			"Incompatible CDN options",
			"slice and fetched_compressed cannot both be enabled simultaneously. Set one of them to false.",
		)
	}
}

func (v *sliceFetchedCompressedValidator) Description(_ context.Context) string {
	return "Validates that slice and fetched_compressed are not both enabled"
}

func (v *sliceFetchedCompressedValidator) MarkdownDescription(_ context.Context) string {
	return "Validates that `slice` and `fetched_compressed` are not both enabled"
}

// gzipOnFetchedCompressedValidator validates that gzip_on and fetched_compressed are not both enabled
type gzipOnFetchedCompressedValidator struct{}

func NewGzipOnFetchedCompressedValidator() validator.List {
	return &gzipOnFetchedCompressedValidator{}
}

func (v *gzipOnFetchedCompressedValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}

	// Get the single element (MaxItems: 1)
	var elements []types.Object
	diags := req.ConfigValue.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(elements) == 0 {
		return
	}

	elem := elements[0]
	if elem.IsNull() || elem.IsUnknown() {
		return
	}

	// Extract gzip_on and fetched_compressed from the element
	var options CDNOptionsModel
	diags = elem.As(ctx, &options, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if both are set to true
	if !options.GzipOn.IsNull() && !options.GzipOn.IsUnknown() && options.GzipOn.ValueBool() &&
		!options.FetchedCompressed.IsNull() && !options.FetchedCompressed.IsUnknown() && options.FetchedCompressed.ValueBool() {
		resp.Diagnostics.AddError(
			"Incompatible CDN compression options",
			"gzip_on and fetched_compressed cannot both be enabled simultaneously. These are mutually exclusive compression methods. Set one of them to false.",
		)
	}
}

func (v *gzipOnFetchedCompressedValidator) Description(_ context.Context) string {
	return "Validates that gzip_on and fetched_compressed are not both enabled"
}

func (v *gzipOnFetchedCompressedValidator) MarkdownDescription(_ context.Context) string {
	return "Validates that `gzip_on` and `fetched_compressed` are not both enabled"
}

// mutuallyExclusiveBoolsValidator validates that two boolean options are not both enabled (generic validator)
type mutuallyExclusiveBoolsValidator struct {
	field1Name     string
	field2Name     string
	errorSummary   string
	errorDetail    string
	field1Accessor func(*CDNOptionsModel) types.Bool
	field2Accessor func(*CDNOptionsModel) types.Bool
}

// NewMutuallyExclusiveBoolsValidator creates a generic validator for mutually exclusive boolean fields
func NewMutuallyExclusiveBoolsValidator(
	field1Name, field2Name string,
	field1Accessor, field2Accessor func(*CDNOptionsModel) types.Bool,
	errorSummary, errorDetail string,
) validator.List {
	return &mutuallyExclusiveBoolsValidator{
		field1Name:     field1Name,
		field2Name:     field2Name,
		errorSummary:   errorSummary,
		errorDetail:    errorDetail,
		field1Accessor: field1Accessor,
		field2Accessor: field2Accessor,
	}
}

func (v *mutuallyExclusiveBoolsValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}

	// Get the single element (MaxItems: 1)
	var elements []types.Object
	diags := req.ConfigValue.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || len(elements) == 0 {
		return
	}

	elem := elements[0]
	if elem.IsNull() || elem.IsUnknown() {
		return
	}

	// Extract fields from the element
	var options CDNOptionsModel
	diags = elem.As(ctx, &options, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get field values using accessors
	field1 := v.field1Accessor(&options)
	field2 := v.field2Accessor(&options)

	// Check if both are set to true
	if !field1.IsNull() && field1.ValueBool() && !field2.IsNull() && field2.ValueBool() {
		resp.Diagnostics.AddError(v.errorSummary, v.errorDetail)
	}
}

func (v *mutuallyExclusiveBoolsValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Validates that %s and %s are not both enabled", v.field1Name, v.field2Name)
}

func (v *mutuallyExclusiveBoolsValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Validates that `%s` and `%s` are not both enabled", v.field1Name, v.field2Name)
}
