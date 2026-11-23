package pluginsdk

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func greenplumPoolDiscardTestConfig(pool map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		"name": "test-cluster", "folder_id": "folder-id", "network_id": "network-id",
		"environment": "PRESTABLE", "zone": "ru-central1-a", "subnet_id": "subnet-id",
		"user_name": "user", "user_password": "test-password", "version": "6.29",
	}
	if pool != nil {
		raw["pooler_config"] = []interface{}{pool}
	}
	return raw
}

func greenplumPoolDiscardTestData(t *testing.T, state *terraform.InstanceState, raw map[string]interface{}) *schema.ResourceData {
	t.Helper()
	r := ResourceYandexMDBGreenplumCluster()
	resourceSchema := schema.InternalMap(r.Schema)
	diff, err := resourceSchema.Diff(context.Background(), state, terraform.NewResourceConfigRaw(raw), nil, nil, true)
	require.NoError(t, err)
	if diff == nil {
		diff = &terraform.InstanceDiff{}
	}
	prior := cty.NullVal(r.CoreConfigSchema().ImpliedType())
	if state != nil {
		prior, err = state.AttrsAsObjectValue(r.CoreConfigSchema().ImpliedType())
		require.NoError(t, err)
	}
	// TestResourceDataRaw does not populate the raw plan supplied by Terraform during apply.
	diff.RawPlan, err = schema.ApplyDiff(prior, diff, r.CoreConfigSchema())
	require.NoError(t, err)
	d, err := resourceSchema.Data(state, diff)
	require.NoError(t, err)
	return d
}

func TestMDBGreenplumClusterPoolDiscardCreate(t *testing.T) {
	for _, tt := range []struct {
		name string
		pool map[string]interface{}
		want *wrapperspb.BoolValue
	}{
		{name: "no pooler config"},
		{name: "omitted", pool: map[string]interface{}{"pooling_mode": "SESSION"}},
		{name: "null", pool: map[string]interface{}{"pool_discard": nil}},
		{name: "true", pool: map[string]interface{}{"pool_discard": true}, want: wrapperspb.Bool(true)},
		{name: "false", pool: map[string]interface{}{"pool_discard": false}, want: wrapperspb.Bool(false)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := greenplumPoolDiscardTestData(t, nil, greenplumPoolDiscardTestConfig(tt.pool))
			req, err := PrepareCreateGreenplumClusterRequest(d, &providerconfig.Config{})
			require.NoError(t, err)
			assert.Equal(t, tt.want, req.ConfigSpec.Pool.PoolDiscard)
		})
	}
}

func TestMDBGreenplumClusterPoolDiscardDiff(t *testing.T) {
	const field = "pooler_config.0.pool_discard"
	const path = "config_spec.pool.pool_discard"
	for _, previous := range []bool{true, false} {
		for _, next := range []string{"omitted", "no block", "true", "false", "change sibling"} {
			t.Run(fmt.Sprintf("%t to %s", previous, next), func(t *testing.T) {
				r := ResourceYandexMDBGreenplumCluster()
				pool := map[string]interface{}{"pooling_mode": "SESSION", "pool_discard": previous, "pool_size": 10}
				d := schema.TestResourceDataRaw(t, r.Schema, greenplumPoolDiscardTestConfig(pool))
				d.SetId("cluster-id")
				state := d.State()
				delete(pool, "pool_discard")
				if next == "true" || next == "false" {
					pool["pool_discard"] = next == "true"
				}
				if next == "no block" {
					pool = nil
				}
				if next == "change sibling" {
					pool["pool_size"] = 20
				}
				updated := greenplumPoolDiscardTestData(t, state, greenplumPoolDiscardTestConfig(pool))
				want := previous
				if next == "true" || next == "false" {
					want = next == "true"
				}
				assert.Equal(t, want != previous, updated.HasChange(field))
				req, err := prepareUpdateGreenplumClusterRequest(updated, &providerconfig.Config{})
				require.NoError(t, err)
				require.NotNil(t, req.ConfigSpec.Pool.PoolDiscard)
				assert.Equal(t, want, req.ConfigSpec.Pool.PoolDiscard.Value)
				if want != previous {
					assert.Contains(t, req.UpdateMask.Paths, path)
				} else {
					assert.NotContains(t, req.UpdateMask.Paths, path)
				}
				if next == "change sibling" {
					assert.Contains(t, req.UpdateMask.Paths, "config_spec.pool.size")
				}
				if want == previous && next != "change sibling" {
					assert.False(t, updated.HasChange("pooler_config"), "unchanged pooler config must not produce a diff")
				}
			})
		}
	}
}

func TestMDBGreenplumClusterPoolDiscardRead(t *testing.T) {
	for _, value := range []bool{true, false} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			pool, err := flattenGreenplumPoolerConfig(&greenplum.ConnectionPoolerConfigSet{
				// Read must use the effective setting, including server defaults.
				UserConfig: &greenplum.ConnectionPoolerConfig{PoolDiscard: wrapperspb.Bool(!value)},
				EffectiveConfig: &greenplum.ConnectionPoolerConfig{
					Mode: greenplum.ConnectionPoolerConfig_SESSION, PoolDiscard: wrapperspb.Bool(value),
				},
			})
			require.NoError(t, err)
			r := ResourceYandexMDBGreenplumCluster()
			d := r.Data(nil)
			d.SetId("cluster-id")
			imported, err := r.Importer.State(d, nil)
			require.NoError(t, err)
			require.Len(t, imported, 1)
			require.NoError(t, imported[0].Set("pooler_config", pool))
			assert.Equal(t, fmt.Sprint(value), imported[0].State().Attributes["pooler_config.0.pool_discard"])

			ds := DataSourceYandexMDBGreenplumCluster().Data(nil)
			ds.SetId("cluster-id")
			require.NoError(t, ds.Set("pooler_config", pool))
			assert.Equal(t, fmt.Sprint(value), ds.State().Attributes["pooler_config.0.pool_discard"])
		})
	}
	for _, config := range []*greenplum.ConnectionPoolerConfigSet{nil, {}, {EffectiveConfig: &greenplum.ConnectionPoolerConfig{}}} {
		_, err := flattenGreenplumPoolerConfig(config)
		require.NoError(t, err)
	}
}
