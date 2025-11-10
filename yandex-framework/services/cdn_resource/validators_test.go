package cdn_resource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// TestSliceFetchedCompressedValidator_BothTrue verifies that validation fails when both are true
func TestSliceFetchedCompressedValidator_BothTrue(t *testing.T) {
	ctx := context.Background()

	optionsModel := createCDNOptionsModel(types.BoolValue(true), types.BoolValue(true))

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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

	optionsValue, diags := types.ObjectValueFrom(ctx, getCDNOptionsAttrTypes(), optionsModel)
	if diags.HasError() {
		t.Fatalf("Creating test object produced errors: %v", diags)
	}

	listValue, diags := types.ListValue(types.ObjectType{
		AttrTypes: getCDNOptionsAttrTypes(),
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
