package cdn_origin_group

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

// TestExpandOriginParams_IgnoresOriginGroupID is a CRITICAL test that ensures
// origin_group_id is NOT sent to the API during CREATE/UPDATE operations.
//
// CONTEXT: origin_group_id is a Computed field that comes from the API response.
// It should NEVER be included in OriginParams sent TO the API.
//
// If we accidentally include it, it could cause API errors or unexpected behavior.
func TestExpandOriginParams_IgnoresOriginGroupID(t *testing.T) {
	ctx := context.Background()

	origin := &OriginModel{
		Source:        types.StringValue("example.com:443"),
		OriginGroupID: types.StringValue("12345"), // This should be IGNORED
		Enabled:       types.BoolValue(true),
		Backup:        types.BoolValue(false),
	}

	result := expandOriginParams(ctx, origin)

	require.NotNil(t, result, "expandOriginParams should return non-nil result")

	// Verify correct fields are included
	assert.Equal(t, "example.com:443", result.Source)
	assert.True(t, result.Enabled)
	assert.False(t, result.Backup)

	// CRITICAL: OriginParams should NOT have OriginGroupId field set
	// The protobuf definition shows OriginParams doesn't have this field,
	// but we verify the expanded struct only contains the correct fields
	assert.Equal(t, "example.com:443", result.Source, "Source should be set")
	assert.True(t, result.Enabled, "Enabled should be set")
	assert.False(t, result.Backup, "Backup should be set")

	// Note: We can't directly check that OriginGroupId is absent because
	// OriginParams struct doesn't have this field at all (by design).
	// This test documents that behavior and ensures we don't accidentally
	// try to use origin_group_id during expand.
}

// TestExpandOriginParams_AllFields verifies all fields are correctly expanded
func TestExpandOriginParams_AllFields(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name            string
		origin          *OriginModel
		expectedSource  string
		expectedEnabled bool
		expectedBackup  bool
	}{
		{
			name: "enabled primary origin",
			origin: &OriginModel{
				Source:        types.StringValue("primary.example.com:80"),
				OriginGroupID: types.StringValue("999"), // Should be ignored
				Enabled:       types.BoolValue(true),
				Backup:        types.BoolValue(false),
			},
			expectedSource:  "primary.example.com:80",
			expectedEnabled: true,
			expectedBackup:  false,
		},
		{
			name: "disabled backup origin",
			origin: &OriginModel{
				Source:        types.StringValue("backup.example.com:443"),
				OriginGroupID: types.StringValue("888"), // Should be ignored
				Enabled:       types.BoolValue(false),
				Backup:        types.BoolValue(true),
			},
			expectedSource:  "backup.example.com:443",
			expectedEnabled: false,
			expectedBackup:  true,
		},
		{
			name: "origin with port",
			origin: &OriginModel{
				Source:        types.StringValue("192.168.1.1:8080"),
				OriginGroupID: types.StringValue("777"), // Should be ignored
				Enabled:       types.BoolValue(true),
				Backup:        types.BoolValue(false),
			},
			expectedSource:  "192.168.1.1:8080",
			expectedEnabled: true,
			expectedBackup:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := expandOriginParams(ctx, tc.origin)

			require.NotNil(t, result)
			assert.Equal(t, tc.expectedSource, result.Source)
			assert.Equal(t, tc.expectedEnabled, result.Enabled)
			assert.Equal(t, tc.expectedBackup, result.Backup)
		})
	}
}

// TestExpandOriginParams_NilInput verifies correct handling of nil input
func TestExpandOriginParams_NilInput(t *testing.T) {
	ctx := context.Background()

	result := expandOriginParams(ctx, nil)

	assert.Nil(t, result, "expandOriginParams with nil input should return nil")
}

// TestExpandOrigins_MultipleOrigins verifies correct expansion of multiple origins
func TestExpandOrigins_MultipleOrigins(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	originsSet := []OriginModel{
		{
			Source:        types.StringValue("origin1.example.com:443"),
			OriginGroupID: types.StringValue("111"), // Should be ignored
			Enabled:       types.BoolValue(true),
			Backup:        types.BoolValue(false),
		},
		{
			Source:        types.StringValue("origin2.example.com:443"),
			OriginGroupID: types.StringValue("222"), // Should be ignored
			Enabled:       types.BoolValue(true),
			Backup:        types.BoolValue(false),
		},
		{
			Source:        types.StringValue("backup.example.com:443"),
			OriginGroupID: types.StringValue("333"), // Should be ignored
			Enabled:       types.BoolValue(false),
			Backup:        types.BoolValue(true),
		},
	}

	result := expandOrigins(ctx, originsSet, &diags)

	require.False(t, diags.HasError(), "expandOrigins should not produce errors")
	require.NotNil(t, result)
	require.Len(t, result, 3, "should expand all 3 origins")

	// Verify first origin
	assert.Equal(t, "origin1.example.com:443", result[0].Source)
	assert.True(t, result[0].Enabled)
	assert.False(t, result[0].Backup)

	// Verify second origin
	assert.Equal(t, "origin2.example.com:443", result[1].Source)
	assert.True(t, result[1].Enabled)
	assert.False(t, result[1].Backup)

	// Verify third origin (backup)
	assert.Equal(t, "backup.example.com:443", result[2].Source)
	assert.False(t, result[2].Enabled)
	assert.True(t, result[2].Backup)
}

// TestExpandOrigins_EmptySet verifies correct handling of empty origin set
func TestExpandOrigins_EmptySet(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	result := expandOrigins(ctx, []OriginModel{}, &diags)

	require.False(t, diags.HasError())
	assert.Nil(t, result, "empty origin set should return nil")
}

// TestExpandOrigins_NilSet verifies correct handling of nil origin set
func TestExpandOrigins_NilSet(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	result := expandOrigins(ctx, nil, &diags)

	require.False(t, diags.HasError())
	assert.Nil(t, result, "nil origin set should return nil")
}

// TestExpandCollapse_RoundTrip verifies that expand->API->flatten produces consistent results.
// This is a critical integration test that simulates the full lifecycle.
func TestExpandCollapse_RoundTrip(t *testing.T) {
	ctx := context.Background()

	// Original Terraform config
	originalOrigins := []OriginModel{
		{
			Source:  types.StringValue("example.com:443"),
			Enabled: types.BoolValue(true),
			Backup:  types.BoolValue(false),
			// Note: OriginGroupID is NOT set by user, it's Computed
		},
	}

	// Step 1: Expand (Terraform -> API request)
	var expandDiags diag.Diagnostics
	expandedParams := expandOrigins(ctx, originalOrigins, &expandDiags)
	require.False(t, expandDiags.HasError())
	require.Len(t, expandedParams, 1)

	// Verify expanded params don't have origin_group_id
	assert.Equal(t, "example.com:443", expandedParams[0].Source)
	assert.True(t, expandedParams[0].Enabled)
	assert.False(t, expandedParams[0].Backup)

	// Step 2: Simulate API response (API adds ID fields)
	// In real scenario, API would return Origin (with OriginGroupId) instead of OriginParams
	simulatedAPIResponse := []*cdn.Origin{
		{
			Id:            1,
			OriginGroupId: 12345, // API sets this
			Source:        expandedParams[0].Source,
			Enabled:       expandedParams[0].Enabled,
			Backup:        expandedParams[0].Backup,
		},
	}

	// Step 3: Flatten (API response -> Terraform state)
	var flattenDiags diag.Diagnostics
	parentGroupID := "12345" // This comes from the resource ID
	flattenedSet := flattenOrigins(ctx, simulatedAPIResponse, parentGroupID, &flattenDiags)
	require.False(t, flattenDiags.HasError())

	// Step 4: Extract and verify
	var flattenedOrigins []OriginModel
	flattenDiags = flattenedSet.ElementsAs(ctx, &flattenedOrigins, false)
	require.False(t, flattenDiags.HasError())
	require.Len(t, flattenedOrigins, 1)

	// Verify flattened data matches original + computed fields
	assert.Equal(t, "example.com:443", flattenedOrigins[0].Source.ValueString())
	assert.True(t, flattenedOrigins[0].Enabled.ValueBool())
	assert.False(t, flattenedOrigins[0].Backup.ValueBool())

	// CRITICAL: origin_group_id should be parentGroupID (12345), not API's value
	assert.Equal(t, "12345", flattenedOrigins[0].OriginGroupID.ValueString(),
		"origin_group_id must match parent resource ID for Set hash consistency")
}
