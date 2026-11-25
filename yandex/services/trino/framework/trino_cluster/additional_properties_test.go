package trino_cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
)

func additionalPropertiesTestModel(t *testing.T, properties types.Map) ClusterModel {
	t.Helper()
	model := eventListenersTestModel(t, NewEventListenersValueNull())
	model.AdditionalProperties = properties
	// Keep resource groups unchanged to isolate the update mask under test.
	model.ResourceGroupsJson = types.StringValue("{}")
	return model
}

func TestAdditionalPropertiesCreateRequest(t *testing.T) {
	for _, tc := range []struct {
		name       string
		properties map[string]string
	}{
		{"omitted", nil},
		{"empty", map[string]string{}},
		{"configured", map[string]string{"sql.default-schema": "default", "sql.default-catalog": "hive"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			properties, diags := types.MapValueFrom(context.Background(), types.StringType, tc.properties)
			require.False(t, diags.HasError(), "%v", diags)
			model := additionalPropertiesTestModel(t, properties)
			req, diags := BuildCreateClusterRequest(context.Background(), &model, &config.State{})
			require.False(t, diags.HasError(), "%v", diags)
			require.Equal(t, tc.properties, req.GetTrino().GetAdditionalProperties())
		})
	}
}

func TestAdditionalPropertiesUpdateRequest(t *testing.T) {
	absent := types.MapNull(types.StringType)
	empty := types.MapValueMust(types.StringType, map[string]attr.Value{})
	configured := types.MapValueMust(types.StringType, map[string]attr.Value{
		"sql.default-schema":  types.StringValue("default"),
		"sql.default-catalog": types.StringValue("hive"),
	})
	changed := types.MapValueMust(types.StringType, map[string]attr.Value{
		"sql.default-schema": types.StringValue("analytics"),
	})
	for _, tc := range []struct {
		name          string
		before, after types.Map
		wantChange    bool
	}{
		{"add", absent, configured, true},
		{"change_value_and_remove_key", configured, changed, true},
		{"remove_attribute", configured, absent, true},
		{"clear_map", configured, empty, true},
		{"unchanged", configured, configured, false},
		{"unchanged_omitted", absent, absent, false},
		{"null_to_empty", absent, empty, false},
		{"empty_to_null", empty, absent, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := additionalPropertiesTestModel(t, tc.before)
			plan := additionalPropertiesTestModel(t, tc.after)
			plan.Name = types.StringValue("renamed-cluster")
			req, diags := BuildUpdateClusterRequest(context.Background(), &state, &plan)
			require.False(t, diags.HasError(), "%v", diags)
			wantPaths := []string{"name"}
			if tc.wantChange {
				wantPaths = append(wantPaths, "trino.additional_properties")
			}
			require.ElementsMatch(t, wantPaths, req.GetUpdateMask().GetPaths())
			var want map[string]string
			diags = tc.after.ElementsAs(context.Background(), &want, false)
			require.False(t, diags.HasError(), "%v", diags)
			require.Equal(t, want, req.GetTrino().GetAdditionalProperties())
		})
	}
}

func TestAdditionalPropertiesUnknownRequest(t *testing.T) {
	for _, properties := range []types.Map{
		types.MapUnknown(types.StringType),
		types.MapValueMust(types.StringType, map[string]attr.Value{"sql.default-schema": types.StringUnknown()}),
	} {
		model := additionalPropertiesTestModel(t, properties)
		req, diags := BuildCreateClusterRequest(context.Background(), &model, &config.State{})
		require.True(t, diags.HasError())
		require.Nil(t, req)
	}
}

func TestAdditionalPropertiesRefresh(t *testing.T) {
	absent := types.MapNull(types.StringType)
	empty := types.MapValueMust(types.StringType, map[string]attr.Value{})
	unknown := types.MapUnknown(types.StringType)
	apiProperties := map[string]string{"sql.default-schema": "default"}
	configured, diags := types.MapValueFrom(context.Background(), types.StringType, apiProperties)
	require.False(t, diags.HasError(), "%v", diags)
	for _, tc := range []struct {
		name   string
		before types.Map
		api    map[string]string
		want   types.Map
	}{
		{"omitted", absent, nil, absent},
		{"preserve_null", absent, map[string]string{}, absent},
		{"preserve_empty", empty, nil, empty},
		{"configured", configured, apiProperties, configured},
		{"import", absent, apiProperties, configured},
		{"data_source", unknown, apiProperties, configured},
		{"data_source_absent", unknown, nil, absent},
		{"data_source_empty", unknown, map[string]string{}, empty},
		{"external_add", empty, apiProperties, configured},
		{"external_remove", configured, nil, absent},
		{"external_clear", configured, map[string]string{}, empty},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := additionalPropertiesTestModel(t, tc.before)
			cluster := &trino.Cluster{Trino: &trino.TrinoConfig{AdditionalProperties: tc.api}}
			diags := ClusterToState(context.Background(), cluster, &state)
			require.False(t, diags.HasError(), "%v", diags)
			require.True(t, tc.want.Equal(state.AdditionalProperties), "want %s, got %s", tc.want, state.AdditionalProperties)
		})
	}
}
