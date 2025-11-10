package cdn_origin_group

import (
	"context"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

// TestFlattenOrigins_UsesParentGroupID is the CRITICAL test that prevents the
// "Provider produced inconsistent result after apply" bug.
//
// ROOT CAUSE: If we use origin.OriginGroupId from API response, the Set hash will change
// between plan (where origin_group_id is unknown/null) and apply (where it becomes the actual ID).
//
// CORRECT BEHAVIOR: We must use parentGroupID parameter, which is the resource's own ID,
// ensuring consistency between plan and apply phases.
func TestFlattenOrigins_UsesParentGroupID(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Create API origin with OriginGroupId = 999 (different from parent)
	apiOrigins := []*cdn.Origin{
		{
			Id:            1,
			OriginGroupId: 999, // ← This should be IGNORED
			Source:        "example.com:443",
			Enabled:       true,
			Backup:        false,
		},
	}

	// Flatten with parentGroupID = "12345"
	parentGroupID := "12345"
	result := flattenOrigins(ctx, apiOrigins, parentGroupID, &diags)

	require.False(t, diags.HasError(), "flattenOrigins should not produce errors")
	require.False(t, result.IsNull(), "result should not be null")

	// Extract the origin model
	var origins []OriginModel
	diags = result.ElementsAs(ctx, &origins, false)
	require.False(t, diags.HasError(), "ElementsAs should not produce errors")
	require.Len(t, origins, 1, "should have exactly one origin")

	origin := origins[0]

	// CRITICAL ASSERTION: origin_group_id must equal parentGroupID (12345), NOT api's OriginGroupId (999)
	assert.Equal(t, "12345", origin.OriginGroupID.ValueString(),
		"origin_group_id MUST use parentGroupID parameter, not API's origin.OriginGroupId - this prevents Set hash mismatch bug")

	// Verify other fields are correct
	assert.Equal(t, "example.com:443", origin.Source.ValueString())
	assert.True(t, origin.Enabled.ValueBool())
	assert.False(t, origin.Backup.ValueBool())
}

// TestFlattenOrigins_SetHashConsistency verifies that identical origins produce
// identical Set hashes. This test ensures that the Set hash calculation is stable
// and doesn't introduce inconsistencies.
func TestFlattenOrigins_SetHashConsistency(t *testing.T) {
	ctx := context.Background()
	var diags1, diags2 diag.Diagnostics

	apiOrigins := []*cdn.Origin{
		{
			Id:            1,
			OriginGroupId: 100,
			Source:        "test.example.com:80",
			Enabled:       true,
			Backup:        false,
		},
		{
			Id:            2,
			OriginGroupId: 100,
			Source:        "backup.example.com:80",
			Enabled:       false,
			Backup:        true,
		},
	}

	parentGroupID := "12345"

	// Flatten twice with the same data
	result1 := flattenOrigins(ctx, apiOrigins, parentGroupID, &diags1)
	result2 := flattenOrigins(ctx, apiOrigins, parentGroupID, &diags2)

	require.False(t, diags1.HasError())
	require.False(t, diags2.HasError())

	// Sets should be equal (same hash)
	assert.True(t, result1.Equal(result2),
		"Identical data should produce identical Sets with same hash - this ensures consistency between plan and apply")
}

// TestFlattenOrigins_EmptyOrigins verifies correct handling of empty origin list
func TestFlattenOrigins_EmptyOrigins(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	result := flattenOrigins(ctx, []*cdn.Origin{}, "12345", &diags)

	require.False(t, diags.HasError())
	assert.True(t, result.IsNull(), "empty origins should produce null Set")
}

// TestFlattenOrigins_MultipleOrigins verifies correct handling of multiple origins
func TestFlattenOrigins_MultipleOrigins(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	apiOrigins := []*cdn.Origin{
		{
			Id:            1,
			OriginGroupId: 888, // Different ID - should be ignored
			Source:        "origin1.example.com:443",
			Enabled:       true,
			Backup:        false,
		},
		{
			Id:            2,
			OriginGroupId: 999, // Different ID - should be ignored
			Source:        "origin2.example.com:443",
			Enabled:       true,
			Backup:        false,
		},
		{
			Id:            3,
			OriginGroupId: 777, // Different ID - should be ignored
			Source:        "backup.example.com:443",
			Enabled:       false,
			Backup:        true,
		},
	}

	parentGroupID := "555"
	result := flattenOrigins(ctx, apiOrigins, parentGroupID, &diags)

	require.False(t, diags.HasError())
	require.False(t, result.IsNull())

	var origins []OriginModel
	diags = result.ElementsAs(ctx, &origins, false)
	require.False(t, diags.HasError())
	require.Len(t, origins, 3)

	// Verify ALL origins use parentGroupID
	for i, origin := range origins {
		assert.Equal(t, "555", origin.OriginGroupID.ValueString(),
			"origin[%d].origin_group_id must use parentGroupID (555), got %s", i, origin.OriginGroupID.ValueString())
	}
}

// TestOriginModelSetHash_Stability verifies that Set hash calculation is stable
// and doesn't change unexpectedly. This test creates Sets directly from OriginModel
// to verify hash behavior.
func TestOriginModelSetHash_Stability(t *testing.T) {
	ctx := context.Background()

	// Create the same origin model twice
	createOriginSet := func() types.List {
		originModels := []OriginModel{
			{
				Source:        types.StringValue("test.example.com:443"),
				OriginGroupID: types.StringValue("12345"),
				Enabled:       types.BoolValue(true),
				Backup:        types.BoolValue(false),
			},
		}

		setVal, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"source":          types.StringType,
				"origin_group_id": types.StringType,
				"enabled":         types.BoolType,
				"backup":          types.BoolType,
			},
		}, originModels)

		require.False(t, diags.HasError(), "SetValueFrom should not produce errors")
		return setVal
	}

	set1 := createOriginSet()
	set2 := createOriginSet()

	assert.True(t, set1.Equal(set2),
		"Sets created from identical data must be equal (same hash)")
}

// TestFlattenCDNOriginGroup verifies the full flattenCDNOriginGroup function
// uses correct parentGroupID handling
func TestFlattenCDNOriginGroup(t *testing.T) {
	ctx := context.Background()

	originGroup := &cdn.OriginGroup{
		Id:       12345,
		FolderId: "folder123",
		Name:     "test-group",
		UseNext:  true,
		Origins: []*cdn.Origin{
			{
				Id:            1,
				OriginGroupId: 99999, // Wrong ID - should be ignored
				Source:        "origin.example.com:443",
				Enabled:       true,
				Backup:        false,
			},
		},
	}

	var state CDNOriginGroupModel
	state.ID = types.StringValue(strconv.FormatInt(originGroup.Id, 10))
	var diags diag.Diagnostics

	flattenCDNOriginGroup(ctx, &state, originGroup, &diags)

	require.False(t, diags.HasError(), "flattenCDNOriginGroup should not produce errors")

	// Extract origins
	var origins []OriginModel
	diags = state.Origins.ElementsAs(ctx, &origins, false)
	require.False(t, diags.HasError())
	require.Len(t, origins, 1)

	// CRITICAL: origin_group_id should be "12345" (from state.ID), not "99999" (from API)
	expectedID := strconv.FormatInt(originGroup.Id, 10) // "12345"
	assert.Equal(t, expectedID, origins[0].OriginGroupID.ValueString(),
		"flattenCDNOriginGroup must use resource ID as parentGroupID")
}

// TestFlattenOrigins_NilOriginsList verifies correct handling of nil origins list
func TestFlattenOrigins_NilOriginsList(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	result := flattenOrigins(ctx, nil, "12345", &diags)

	require.False(t, diags.HasError())
	assert.True(t, result.IsNull(), "nil origins should produce null Set")
}

// TestPrePopulateOriginGroupID_SetHashConsistency is the CRITICAL test that verifies
// the pre-populate fix for "Provider produced inconsistent result after apply" bug.
//
// This test simulates the full CREATE lifecycle:
// 1. User plan has origins with origin_group_id = NULL
// 2. After resource creation, we pre-populate origin_group_id = resource ID
// 3. Then flatten is called, which should produce the SAME Set hash
// 4. Terraform compares hashes - they must match!
func TestPrePopulateOriginGroupID_SetHashConsistency(t *testing.T) {
	ctx := context.Background()

	// Step 1: Create initial plan origins (from user config - no origin_group_id)
	planOrigins := []OriginModel{
		{
			Source:        types.StringValue("example.com:443"),
			OriginGroupID: types.StringNull(), // NULL in initial plan
			Enabled:       types.BoolValue(true),
			Backup:        types.BoolValue(false),
		},
	}

	// Create Set from plan origins
	planSet, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"source":          types.StringType,
			"origin_group_id": types.StringType,
			"enabled":         types.BoolType,
			"backup":          types.BoolType,
		},
	}, planOrigins)
	require.False(t, diags.HasError())

	// Step 2: Simulate resource creation - pre-populate origin_group_id
	resourceID := "12345"
	var originsForUpdate []OriginModel
	diags = planSet.ElementsAs(ctx, &originsForUpdate, false)
	require.False(t, diags.HasError())

	for i := range originsForUpdate {
		originsForUpdate[i].OriginGroupID = types.StringValue(resourceID)
	}

	prePopulatedSet, diags := types.ListValueFrom(ctx, planSet.ElementType(ctx), originsForUpdate)
	require.False(t, diags.HasError())

	// Step 3: Simulate API response and flatten
	apiOrigins := []*cdn.Origin{
		{
			Id:            1,
			OriginGroupId: 12345,
			Source:        "example.com:443",
			Enabled:       true,
			Backup:        false,
		},
	}

	var flattenDiags diag.Diagnostics
	flattenedSet := flattenOrigins(ctx, apiOrigins, resourceID, &flattenDiags)
	require.False(t, flattenDiags.HasError())

	// Step 4: CRITICAL ASSERTION - Sets must be equal (same hash)
	assert.True(t, prePopulatedSet.Equal(flattenedSet),
		"Pre-populated Set and flattened Set must have the same hash to prevent 'inconsistent result' error")

	// Verify both have the correct origin_group_id
	var prePopOrigins, flatOrigins []OriginModel
	diags = prePopulatedSet.ElementsAs(ctx, &prePopOrigins, false)
	require.False(t, diags.HasError())
	diags = flattenedSet.ElementsAs(ctx, &flatOrigins, false)
	require.False(t, diags.HasError())

	assert.Equal(t, "12345", prePopOrigins[0].OriginGroupID.ValueString(),
		"Pre-populated origin must have origin_group_id = resource ID")
	assert.Equal(t, "12345", flatOrigins[0].OriginGroupID.ValueString(),
		"Flattened origin must have origin_group_id = resource ID")
}

// TestPrePopulateOriginGroupID_MultipleOrigins verifies pre-populate works with multiple origins
func TestPrePopulateOriginGroupID_MultipleOrigins(t *testing.T) {
	ctx := context.Background()

	// Create plan with 3 origins, all with NULL origin_group_id
	planOrigins := []OriginModel{
		{
			Source:        types.StringValue("origin1.example.com:443"),
			OriginGroupID: types.StringNull(),
			Enabled:       types.BoolValue(true),
			Backup:        types.BoolValue(false),
		},
		{
			Source:        types.StringValue("origin2.example.com:443"),
			OriginGroupID: types.StringNull(),
			Enabled:       types.BoolValue(true),
			Backup:        types.BoolValue(false),
		},
		{
			Source:        types.StringValue("backup.example.com:443"),
			OriginGroupID: types.StringNull(),
			Enabled:       types.BoolValue(false),
			Backup:        types.BoolValue(true),
		},
	}

	planSet, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"source":          types.StringType,
			"origin_group_id": types.StringType,
			"enabled":         types.BoolType,
			"backup":          types.BoolType,
		},
	}, planOrigins)
	require.False(t, diags.HasError())

	// Pre-populate origin_group_id
	resourceID := "99999"
	var originsForUpdate []OriginModel
	diags = planSet.ElementsAs(ctx, &originsForUpdate, false)
	require.False(t, diags.HasError())

	for i := range originsForUpdate {
		originsForUpdate[i].OriginGroupID = types.StringValue(resourceID)
	}

	prePopulatedSet, diags := types.ListValueFrom(ctx, planSet.ElementType(ctx), originsForUpdate)
	require.False(t, diags.HasError())

	// Simulate flatten
	apiOrigins := []*cdn.Origin{
		{Id: 1, OriginGroupId: 99999, Source: "origin1.example.com:443", Enabled: true, Backup: false},
		{Id: 2, OriginGroupId: 99999, Source: "origin2.example.com:443", Enabled: true, Backup: false},
		{Id: 3, OriginGroupId: 99999, Source: "backup.example.com:443", Enabled: false, Backup: true},
	}

	var flattenDiags diag.Diagnostics
	flattenedSet := flattenOrigins(ctx, apiOrigins, resourceID, &flattenDiags)
	require.False(t, flattenDiags.HasError())

	// Sets must be equal
	assert.True(t, prePopulatedSet.Equal(flattenedSet),
		"Pre-populated and flattened Sets with multiple origins must have the same hash")
}
