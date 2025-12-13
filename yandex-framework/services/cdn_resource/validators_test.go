package cdn_resource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSliceFetchedCompressedValidator_BothTrue verifies that validation fails when both are true
func TestSliceFetchedCompressedValidator_BothTrue(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolValue(true), types.BoolValue(true))

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewSliceFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "Validation should fail when both slice and fetched_compressed are true")
	if resp.Diagnostics.HasError() {
		assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Incompatible CDN options",
			"Error message should indicate incompatible options")
	}
}

// Helper function to create a minimal CDNOptionsModel
func createCDNOptionsModel(slice, fetchedCompressed types.Bool) CDNOptionsModel {
	return CDNOptionsModel{
		Slice:                   slice,
		FetchedCompressed:       fetchedCompressed,
		IgnoreQueryParams:       types.BoolNull(),
		GzipOn:                  types.BoolNull(),
		RedirectHttpToHttps:     types.BoolNull(),
		RedirectHttpsToHttp:     types.BoolNull(),
		ForwardHostHeader:       types.BoolNull(),
		ProxyCacheMethodsSet:    types.BoolNull(),
		DisableProxyForceRanges: types.BoolNull(),
		IgnoreCookie:            types.BoolNull(),
		EnableIPURLSigning:      types.BoolNull(),
		CustomHostHeader:        types.StringNull(),
		CustomServerName:        types.StringNull(),
		SecureKey:               types.StringNull(),
		CacheHTTPHeaders:        types.ListNull(types.StringType),
		QueryParamsWhitelist:    types.ListNull(types.StringType),
		QueryParamsBlacklist:    types.ListNull(types.StringType),
		AllowedHTTPMethods:      types.ListNull(types.StringType),
		Cors:                    types.ListNull(types.StringType),
		Stale:                   types.ListNull(types.StringType),
		StaticResponseHeaders:   types.MapNull(types.StringType),
		StaticRequestHeaders:    types.MapNull(types.StringType),
		EdgeCacheSettings: types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled":       types.BoolType,
				"value":         types.Int64Type,
				"custom_values": types.MapType{ElemType: types.Int64Type},
				"default_value": types.Int64Type,
			},
		}),
		BrowserCacheSettings: types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled":    types.BoolType,
				"cache_time": types.Int64Type,
			},
		}),
		IPAddressACL: types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"policy_type":     types.StringType,
				"excepted_values": types.ListType{ElemType: types.StringType},
			},
		}),
		Rewrite: types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled": types.BoolType,
				"body":    types.StringType,
				"flag":    types.StringType,
			},
		}),
	}
}

// TestSliceFetchedCompressedValidator_SliceTrueCompressedFalse verifies that validation passes
func TestSliceFetchedCompressedValidator_SliceTrueCompressedFalse(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolValue(true), types.BoolValue(false))

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewSliceFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Validation should pass when slice=true and fetched_compressed=false")
}

// TestSliceFetchedCompressedValidator_SliceFalseCompressedTrue verifies that validation passes
func TestSliceFetchedCompressedValidator_SliceFalseCompressedTrue(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolValue(false), types.BoolValue(true))

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewSliceFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Validation should pass when slice=false and fetched_compressed=true")
}

// TestSliceFetchedCompressedValidator_BothFalse verifies that validation passes when both are false
func TestSliceFetchedCompressedValidator_BothFalse(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolValue(false), types.BoolValue(false))

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewSliceFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Validation should pass when both slice and fetched_compressed are false")
}

// TestSliceFetchedCompressedValidator_NullValues verifies that validation passes with null values
func TestSliceFetchedCompressedValidator_NullValues(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolNull(), types.BoolNull())

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewSliceFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Validation should pass when values are null")
}

// TestGzipOnFetchedCompressedValidator_BothTrue verifies that validation fails when both are true
func TestGzipOnFetchedCompressedValidator_BothTrue(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolNull(), types.BoolNull())
	optionsModel.GzipOn = types.BoolValue(true)
	optionsModel.FetchedCompressed = types.BoolValue(true)

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewGzipOnFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "Validation should fail when both gzip_on and fetched_compressed are true")
	if resp.Diagnostics.HasError() {
		assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Incompatible CDN compression options",
			"Error message should indicate incompatible compression options")
	}
}

// TestGzipOnFetchedCompressedValidator_GzipTrueCompressedFalse verifies that validation passes
func TestGzipOnFetchedCompressedValidator_GzipTrueCompressedFalse(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolNull(), types.BoolNull())
	optionsModel.GzipOn = types.BoolValue(true)
	optionsModel.FetchedCompressed = types.BoolValue(false)

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewGzipOnFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Validation should pass when gzip_on=true and fetched_compressed=false")
}

// TestGzipOnFetchedCompressedValidator_GzipFalseCompressedTrue verifies that validation passes
func TestGzipOnFetchedCompressedValidator_GzipFalseCompressedTrue(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolNull(), types.BoolNull())
	optionsModel.GzipOn = types.BoolValue(false)
	optionsModel.FetchedCompressed = types.BoolValue(true)

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewGzipOnFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Validation should pass when gzip_on=false and fetched_compressed=true")
}

// TestGzipOnFetchedCompressedValidator_BothFalse verifies that validation passes when both are false
func TestGzipOnFetchedCompressedValidator_BothFalse(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolNull(), types.BoolNull())
	optionsModel.GzipOn = types.BoolValue(false)
	optionsModel.FetchedCompressed = types.BoolValue(false)

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewGzipOnFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Validation should pass when both gzip_on and fetched_compressed are false")
}

// TestGzipOnFetchedCompressedValidator_BothNull verifies that validation passes with null values
func TestGzipOnFetchedCompressedValidator_BothNull(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolNull(), types.BoolNull())
	optionsModel.GzipOn = types.BoolNull()
	optionsModel.FetchedCompressed = types.BoolNull()

	optionsValue, diags := types.ObjectValueFrom(ctx, GetCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: GetCDNOptionsAttrTypes(),
	}, []attr.Value{optionsValue})
	if diags.HasError() {
		t.Fatalf("Creating test list produced errors: %v", diags)
	}

	req := validator.ListRequest{
		Path:        path.Root("options"),
		ConfigValue: listValue,
	}

	resp := &validator.ListResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v := NewGzipOnFetchedCompressedValidator()
	v.ValidateList(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Validation should pass when both values are null")
}

func TestEdgeCacheSettingsValidator(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		enable       basetypes.BoolValue
		value        basetypes.Int64Value
		defaultValue basetypes.Int64Value
		customValues basetypes.MapValue
		diagnostics  []diag.Diagnostic
	}{
		{
			name:         "disable cache with no values",
			enable:       types.BoolValue(false),
			value:        types.Int64Null(),
			defaultValue: types.Int64Unknown(),
			customValues: types.MapUnknown(types.Int64Type),
			diagnostics:  []diag.Diagnostic{},
		},
		{
			name:         "disable cache with some value",
			enable:       types.BoolValue(false),
			value:        types.Int64Value(100),
			defaultValue: types.Int64Unknown(),
			customValues: types.MapUnknown(types.Int64Type),
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic("Invalid edge_cache_settings configuration", "When enabled=false, value, custom_values or default_value should not be specified"),
			},
		},
		{
			name:         "enable cache without specify usage type",
			enable:       types.BoolValue(true),
			value:        types.Int64Null(),
			defaultValue: types.Int64Unknown(),
			customValues: types.MapUnknown(types.Int64Type),
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic("Invalid edge_cache_settings configuration", "When enabled=true, at least value, custom_values or default_value must be specified"),
			},
		},
		{
			name:         "enable cache with value",
			enable:       types.BoolValue(true),
			value:        types.Int64Value(100),
			defaultValue: types.Int64Unknown(),
			customValues: types.MapUnknown(types.Int64Type),
			diagnostics:  []diag.Diagnostic{},
		},
		{
			name:         "enable cache with value which implicitly disabling",
			enable:       types.BoolValue(true),
			value:        types.Int64Value(0),
			defaultValue: types.Int64Unknown(),
			customValues: types.MapUnknown(types.Int64Type),
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic("Invalid edge_cache_settings configuration", "When enabled=true, value cannot be 0, because it will disable cache implicitly"),
			},
		},
		{
			name:         "enable cache with both types of caching",
			enable:       types.BoolValue(true),
			value:        types.Int64Value(100),
			defaultValue: types.Int64Value(100),
			customValues: types.MapUnknown(types.Int64Type),
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic("Invalid edge_cache_settings configuration", "When enabled=true, value and default_value cannot be used together"),
			},
		},
		{
			name:         "enable cache with default value and custom_values",
			enable:       types.BoolValue(true),
			value:        types.Int64Value(100),
			defaultValue: types.Int64Value(100),
			customValues: types.MapValueMust(types.Int64Type, map[string]attr.Value{
				"400": types.Int64Value(200),
			}),
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic("Invalid edge_cache_settings configuration", "When enabled=true, value and default_value cannot be used together"),
				diag.NewErrorDiagnostic("Invalid edge_cache_settings configuration", "When enabled=true, custom_values can be used only with value"),
			},
		},
		{
			name:         "enable cache with default value",
			enable:       types.BoolValue(true),
			value:        types.Int64Null(),
			defaultValue: types.Int64Value(100),
			customValues: types.MapUnknown(types.Int64Type),
			diagnostics:  []diag.Diagnostic{},
		},
		{
			name:         "enable cache with default value which implicitly disabling",
			enable:       types.BoolValue(true),
			value:        types.Int64Null(),
			defaultValue: types.Int64Value(0),
			customValues: types.MapUnknown(types.Int64Type),
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic("Invalid edge_cache_settings configuration", "When enabled=true, default_value cannot be 0, because it will disable cache implicitly"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			edgeCacheSettings := types.ListValueMust(types.ObjectType{
				AttrTypes: GetEdgeCacheSettingsAttrTypes(),
			}, []attr.Value{
				types.ObjectValueMust(GetEdgeCacheSettingsAttrTypes(),
					map[string]attr.Value{
						"enabled":       testCase.enable,
						"value":         testCase.value,
						"default_value": testCase.defaultValue,
						"custom_values": testCase.customValues,
					}),
			})

			req := validator.ListRequest{
				Path:        path.Root("edge_cache_settings"),
				ConfigValue: edgeCacheSettings,
			}

			resp := &validator.ListResponse{Diagnostics: diag.Diagnostics{}}

			v := NewEdgeCacheSettingsValidator()
			v.ValidateList(ctx, req, resp)

			if len(testCase.diagnostics) == 0 {
				assert.False(t, resp.Diagnostics.HasError(), "Expected no errors, but found some")
			} else {
				require.True(t, resp.Diagnostics.HasError(), "Expected errors, but no errors actual")
				require.Equal(t, len(testCase.diagnostics), len(resp.Diagnostics.Errors()), "Count of expected errors does not equal to actual")
				for _, diagnostic := range testCase.diagnostics {
					assert.Truef(t, resp.Diagnostics.Contains(diagnostic), "Diagnostic %s: %s does not contains in errors", diagnostic.Summary(), diagnostic.Detail())
				}
			}
		})
	}
}
