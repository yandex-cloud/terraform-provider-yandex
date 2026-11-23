package pluginsdk

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	ctyjson "github.com/hashicorp/go-cty/cty/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const pgPoolDiscardUpdatePath = "config_spec.pooler_config.pool_discard"

// Use the actual nested schema and SDK refresh/plan/apply lifecycle. Only the
// remote API is replaced: reads return api, and updates apply their field mask.
func pgPoolerTestResource(api *postgresql.ConnectionPoolerConfig, update func(*postgresql.ConnectionPoolerConfig, []string)) *schema.Resource {
	production := ResourceYandexMDBPostgreSQLCluster()
	config := *production.Schema["config"]
	configFields := config.Elem.(*schema.Resource).Schema
	config.Elem = &schema.Resource{Schema: map[string]*schema.Schema{
		"version":       configFields["version"],
		"pooler_config": configFields["pooler_config"],
	}}
	name := "pg-test"
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":   production.Schema["name"],
			"config": &config,
		},
		CustomizeDiff: production.CustomizeDiff,
	}
	r.Read = func(d *schema.ResourceData, _ any) error {
		if err := d.Set("name", name); err != nil {
			return err
		}
		return d.Set("config", []any{map[string]any{
			"version":       "17",
			"pooler_config": flattenPGPoolerConfig(api),
		}})
	}
	r.Update = func(d *schema.ResourceData, _ any) error {
		pooler, err := expandPGPoolerConfig(d)
		if err != nil {
			return err
		}
		paths, err := expandPGParamsUpdatePath(d, nil)
		if err != nil {
			return err
		}
		if update != nil {
			update(pooler, paths)
		}
		for _, path := range paths {
			switch path {
			case pgPoolDiscardUpdatePath:
				api.PoolDiscard = pooler.PoolDiscard
			case "name":
				name = d.Get("name").(string)
			}
		}
		return r.Read(d, nil)
	}
	return r
}

func pgPoolerTestConfig(pooler map[string]any) map[string]any {
	config := map[string]any{"version": "17"}
	if pooler != nil {
		config["pooler_config"] = []any{pooler}
	}
	return map[string]any{"name": "pg-test", "config": []any{config}}
}

func pgPoolerTestRefresh(t *testing.T, r *schema.Resource, state *terraform.InstanceState) *terraform.InstanceState {
	t.Helper()
	state = state.DeepCopy()
	// ReadResource receives state only, without the HCL configuration.
	state.RawConfig = cty.NullVal(r.CoreConfigSchema().ImpliedType())
	state.RawPlan = cty.NullVal(r.CoreConfigSchema().ImpliedType())
	var err error
	state.RawState, err = state.AttrsAsObjectValue(r.CoreConfigSchema().ImpliedType())
	require.NoError(t, err)
	refreshed, diags := r.RefreshWithoutUpgrade(context.Background(), state, nil)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, refreshed)
	return refreshed
}

func pgPoolerTestPlan(t *testing.T, r *schema.Resource, state *terraform.InstanceState, config map[string]any) *terraform.InstanceDiff {
	t.Helper()
	encoded, err := json.Marshal(config)
	require.NoError(t, err)
	raw, err := ctyjson.Unmarshal(encoded, r.CoreConfigSchema().ImpliedType())
	require.NoError(t, err)
	state = state.DeepCopy()
	state.RawConfig = raw
	state.RawState, err = state.AttrsAsObjectValue(r.CoreConfigSchema().ImpliedType())
	require.NoError(t, err)
	diff, err := schema.InternalMap(r.Schema).Diff(context.Background(), state,
		terraform.NewResourceConfigShimmed(raw, r.CoreConfigSchema()), r.CustomizeDiff, nil, true)
	require.NoError(t, err)
	return diff
}

func TestPGPoolerImport(t *testing.T) {
	require.NoError(t, ResourceYandexMDBPostgreSQLCluster().InternalValidate(nil, true))
	require.NoError(t, DataSourceYandexMDBPostgreSQLCluster().InternalValidate(nil, false))
	for _, tc := range []struct {
		name  string
		value *wrapperspb.BoolValue
		want  string
	}{
		{"false", wrapperspb.Bool(false), "false"},
		{"true", wrapperspb.Bool(true), "true"},
		{"unspecified", nil, "unspecified"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := pgPoolerTestResource(&postgresql.ConnectionPoolerConfig{
				PoolingMode: postgresql.ConnectionPoolerConfig_TRANSACTION, PoolDiscard: tc.value,
			}, nil)
			state := pgPoolerTestRefresh(t, r, &terraform.InstanceState{ID: "pg-imported"})
			require.Equal(t, tc.want, state.Attributes["config.0.pooler_config.0.pooler_pool_discard"])
			for _, config := range []struct {
				name   string
				pooler map[string]any
			}{
				{"explicit value", map[string]any{"pooling_mode": "TRANSACTION", "pooler_pool_discard": tc.want}},
				{"field omitted", map[string]any{"pooling_mode": "TRANSACTION"}},
				{"block omitted", nil},
			} {
				t.Run(config.name, func(t *testing.T) {
					diff := pgPoolerTestPlan(t, r, state, pgPoolerTestConfig(config.pooler))
					require.True(t, diff == nil || diff.Empty(), "import must not produce an update: %#v", diff)
				})
			}
		})
	}
}

func TestPGPoolerOldStateRefresh(t *testing.T) {
	for _, value := range []bool{false, true} {
		name := "false"
		if value {
			name = "true"
		}
		t.Run(name, func(t *testing.T) {
			r := pgPoolerTestResource(&postgresql.ConnectionPoolerConfig{
				PoolingMode: postgresql.ConnectionPoolerConfig_TRANSACTION, PoolDiscard: wrapperspb.Bool(value),
			}, nil)
			state := pgPoolerTestRefresh(t, r, &terraform.InstanceState{
				ID: "pg-legacy",
				Attributes: map[string]string{
					"name":                                         "pg-test",
					"config.#":                                     "1",
					"config.0.version":                             "17",
					"config.0.pooler_config.#":                     "1",
					"config.0.pooler_config.0.pooling_mode":        "TRANSACTION",
					"config.0.pooler_config.0.pool_discard":        name,
					"config.0.pooler_config.0.pooler_pool_discard": "",
				},
			})
			require.Equal(t, name, state.Attributes["config.0.pooler_config.0.pooler_pool_discard"])
			require.NotContains(t, state.Attributes, "config.0.pooler_config.0.pool_discard")
			config := pgPoolerTestConfig(map[string]any{"pooling_mode": "TRANSACTION", "pooler_pool_discard": name})
			diff := pgPoolerTestPlan(t, r, state, config)
			require.True(t, diff == nil || diff.Empty(), "old state must migrate without an API update: %#v", diff)

			legacyConfig := pgPoolerTestConfig(map[string]any{"pool_discard": value})
			diags := r.Validate(terraform.NewResourceConfigRaw(legacyConfig))
			require.True(t, diags.HasError(), "removed pool_discard argument must be rejected")
		})
	}
}

func TestPGPoolerUpdate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		before    *wrapperspb.BoolValue
		pooler    map[string]any
		rename    bool
		want      *wrapperspb.BoolValue
		wantPaths []string
	}{
		{name: "enable", before: wrapperspb.Bool(false), pooler: map[string]any{"pooler_pool_discard": "true"}, want: wrapperspb.Bool(true), wantPaths: []string{pgPoolDiscardUpdatePath}},
		{name: "disable", before: wrapperspb.Bool(true), pooler: map[string]any{"pooler_pool_discard": "false"}, want: wrapperspb.Bool(false), wantPaths: []string{pgPoolDiscardUpdatePath}},
		{name: "reset", before: wrapperspb.Bool(false), pooler: map[string]any{"pooler_pool_discard": "unspecified"}, wantPaths: []string{pgPoolDiscardUpdatePath}},
		{name: "set explicit false", pooler: map[string]any{"pooler_pool_discard": "false"}, want: wrapperspb.Bool(false), wantPaths: []string{pgPoolDiscardUpdatePath}},
		{name: "unrelated change with field omitted", before: wrapperspb.Bool(false), pooler: map[string]any{}, rename: true, want: wrapperspb.Bool(false), wantPaths: []string{"name"}},
		{name: "unrelated change with block omitted", before: wrapperspb.Bool(true), rename: true, want: wrapperspb.Bool(true), wantPaths: []string{"name"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := &postgresql.ConnectionPoolerConfig{
				PoolingMode: postgresql.ConnectionPoolerConfig_TRANSACTION, PoolDiscard: tc.before,
			}
			updates := 0
			r := pgPoolerTestResource(api, func(_ *postgresql.ConnectionPoolerConfig, paths []string) {
				updates++
				require.ElementsMatch(t, tc.wantPaths, paths)
			})
			state := &terraform.InstanceState{ID: "pg-imported"}
			state = pgPoolerTestRefresh(t, r, state)
			if tc.pooler != nil {
				tc.pooler["pooling_mode"] = "TRANSACTION"
			}
			config := pgPoolerTestConfig(tc.pooler)
			if tc.rename {
				config["name"] = "pg-renamed"
			}
			diff := pgPoolerTestPlan(t, r, state, config)
			if diff != nil && !diff.Empty() {
				applied, diags := r.Apply(context.Background(), state, diff, nil)
				require.False(t, diags.HasError(), "%v", diags)
				state = applied
			}
			if len(tc.wantPaths) > 0 {
				require.Equal(t, 1, updates)
			}
			require.True(t, proto.Equal(tc.want, api.PoolDiscard), "API pool discard: got %v, want %v", api.PoolDiscard, tc.want)
			state = pgPoolerTestRefresh(t, r, state)
			next := pgPoolerTestPlan(t, r, state, config)
			require.True(t, next == nil || next.Empty(), "configuration must converge after apply: %#v", next)
		})
	}
}

func TestPGPoolerIgnoreChanges(t *testing.T) {
	api := &postgresql.ConnectionPoolerConfig{
		PoolingMode: postgresql.ConnectionPoolerConfig_TRANSACTION, PoolDiscard: wrapperspb.Bool(true),
	}
	updates := 0
	r := pgPoolerTestResource(api, func(pooler *postgresql.ConnectionPoolerConfig, paths []string) {
		updates++
		require.Equal(t, []string{"name"}, paths)
		require.True(t, proto.Equal(wrapperspb.Bool(true), pooler.PoolDiscard), "expand must use the effective plan, not the ignored raw value")
	})
	state := pgPoolerTestRefresh(t, r, &terraform.InstanceState{ID: "pg-imported"})
	config := pgPoolerTestConfig(map[string]any{"pooling_mode": "TRANSACTION", "pooler_pool_discard": "false"})
	config["name"] = "pg-renamed"
	encoded, err := json.Marshal(config)
	require.NoError(t, err)
	raw, err := ctyjson.Unmarshal(encoded, r.CoreConfigSchema().ImpliedType())
	require.NoError(t, err)
	prior, err := state.AttrsAsObjectValue(r.CoreConfigSchema().ImpliedType())
	require.NoError(t, err)
	// Terraform applies ignore_changes before calling ApplyResourceChange:
	// the raw config requests false, but the effective plan retains true.
	plannedState := state.DeepCopy()
	plannedState.Attributes["name"] = "pg-renamed"
	planned, err := plannedState.AttrsAsObjectValue(r.CoreConfigSchema().ImpliedType())
	require.NoError(t, err)
	diff, err := schema.DiffFromValues(context.Background(), prior, planned, raw, r)
	require.NoError(t, err)
	_, diags := r.Apply(context.Background(), state, diff, nil)
	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, 1, updates)
	require.True(t, api.PoolDiscard.GetValue())
}
