package trino_cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func eventListenersTestValue(t *testing.T, enabled bool) EventListenersValue {
	t.Helper()
	dataCatalog := types.ObjectNull(DataCatalogValue{}.AttributeTypes(context.Background()))
	if enabled {
		value, d := (DataCatalogValue{state: attr.ValueStateKnown}).ToObjectValue(context.Background())
		require.False(t, d.HasError(), "%v", d)
		dataCatalog = value
	}
	return EventListenersValue{DataCatalog: dataCatalog, state: attr.ValueStateKnown}
}

func eventListenersTestModel(t *testing.T, listeners EventListenersValue) ClusterModel {
	t.Helper()
	fixedScale, diags := (FixedScaleValue{Count: types.Int64Value(1), state: attr.ValueStateKnown}).ToObjectValue(context.Background())
	require.False(t, diags.HasError(), "%v", diags)
	return ClusterModel{
		Id:               types.StringValue("cluster-id"),
		Name:             types.StringValue("cluster"),
		FolderId:         types.StringValue("folder-id"),
		ServiceAccountId: types.StringValue("service-account-id"),
		Labels:           types.MapNull(types.StringType),
		SubnetIds:        types.SetValueMust(types.StringType, []attr.Value{types.StringValue("subnet-id")}),
		SecurityGroupIds: types.SetNull(types.StringType),
		QueryProperties:  types.MapNull(types.StringType),
		Coordinator:      CoordinatorValue{ResourcePresetId: types.StringValue("c4-m16"), state: attr.ValueStateKnown},
		Worker: WorkerValue{
			ResourcePresetId: types.StringValue("c4-m16"),
			FixedScale:       fixedScale,
			AutoScale:        types.ObjectNull(AutoScaleValue{}.AttributeTypes(context.Background())),
			state:            attr.ValueStateKnown,
		},
		EventListeners: listeners,
	}
}

func TestEventListenersCreateRequest(t *testing.T) {
	for _, tc := range []struct {
		name        string
		value       EventListenersValue
		wantConfig  bool
		wantEnabled bool
	}{
		{"omitted", NewEventListenersValueNull(), false, false},
		{"empty", eventListenersTestValue(t, false), true, false},
		{"data_catalog", eventListenersTestValue(t, true), true, true},
		{"unknown", NewEventListenersValueUnknown(), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := eventListenersTestModel(t, tc.value)
			req, diags := BuildCreateClusterRequest(context.Background(), &model, &config.State{})
			require.False(t, diags.HasError(), "%v", diags)
			require.Equal(t, tc.wantConfig, req.GetTrino().GetEventListeners() != nil)
			require.Equal(t, tc.wantEnabled, req.GetTrino().GetEventListeners().GetDataCatalog() != nil)
		})
	}
}

func TestEventListenersUpdateRequest(t *testing.T) {
	absent := NewEventListenersValueNull()
	empty := eventListenersTestValue(t, false)
	enabled := eventListenersTestValue(t, true)
	for _, tc := range []struct {
		name                    string
		before, after           EventListenersValue
		wantChange, wantEnabled bool
	}{
		{"enable", absent, enabled, true, true},
		{"remove_block", enabled, absent, true, false},
		{"remove_data_catalog", enabled, empty, true, false},
		{"unchanged_enabled", enabled, enabled, false, true},
		{"unchanged_omitted", absent, absent, false, false},
		{"empty_to_null", empty, absent, false, false},
		{"null_to_empty", absent, empty, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := eventListenersTestModel(t, tc.before)
			plan := eventListenersTestModel(t, tc.after)
			// Keep resource groups unchanged so the mask isolates the fields under test.
			state.ResourceGroupsJson = types.StringValue("{}")
			plan.ResourceGroupsJson = state.ResourceGroupsJson
			// Listener changes can be submitted alongside regular cluster changes.
			plan.Name = types.StringValue("renamed-cluster")
			req, diags := BuildUpdateClusterRequest(context.Background(), &state, &plan)
			require.False(t, diags.HasError(), "%v", diags)
			wantPaths := []string{"name"}
			if tc.wantChange {
				wantPaths = append(wantPaths, "trino.event_listeners")
			}
			require.ElementsMatch(t, wantPaths, req.GetUpdateMask().GetPaths())
			require.NotNil(t, req.GetTrino().GetEventListeners())
			require.Equal(t, tc.wantEnabled, req.GetTrino().GetEventListeners().GetDataCatalog() != nil)
		})
	}
}

func TestEventListenersRefresh(t *testing.T) {
	absent := NewEventListenersValueNull()
	empty := eventListenersTestValue(t, false)
	enabled := eventListenersTestValue(t, true)
	apiEnabled := &trino.EventListenersConfig{DataCatalog: &trino.DataCatalogEventListener{}}
	for _, tc := range []struct {
		name   string
		before EventListenersValue
		api    *trino.EventListenersConfig
		want   EventListenersValue
	}{
		{"omitted", absent, nil, absent},
		{"empty_api", absent, &trino.EventListenersConfig{}, absent},
		{"preserve_empty", empty, nil, empty},
		{"preserve_empty_api", empty, &trino.EventListenersConfig{}, empty},
		{"enabled", enabled, apiEnabled, enabled},
		{"import", absent, apiEnabled, enabled},
		{"data_source", NewEventListenersValueUnknown(), apiEnabled, enabled},
		{"data_source_disabled", NewEventListenersValueUnknown(), nil, absent},
		{"external_enable", empty, apiEnabled, enabled},
		{"external_disable", enabled, nil, absent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := eventListenersTestModel(t, tc.before)
			cluster := &trino.Cluster{Trino: &trino.TrinoConfig{EventListeners: tc.api}}
			diags := ClusterToState(context.Background(), cluster, &state)
			require.False(t, diags.HasError(), "%v", diags)
			require.True(t, tc.want.Equal(state.EventListeners), "want %s, got %s", tc.want, state.EventListeners)
		})
	}
}
