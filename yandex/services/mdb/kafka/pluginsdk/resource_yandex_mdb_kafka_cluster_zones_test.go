package pluginsdk_test

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_kafka "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/kafka/pluginsdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type kafkaRoleZoneAPI struct {
	kafkaSubnetClusterAPI
	role    string
	zones   []string
	updates chan *kafka.UpdateClusterRequest
}

func (s *kafkaRoleZoneAPI) Get(ctx context.Context, req *kafka.GetClusterRequest) (*kafka.Cluster, error) {
	cluster, err := s.kafkaSubnetClusterAPI.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	switch s.role {
	case "kraft":
		cluster.Config.Kraft = &kafka.ConfigSpec_KRaft{Resources: cluster.Config.Kafka.Resources, ZoneIds: s.zones}
	case "zookeeper":
		cluster.Config.Zookeeper = &kafka.ConfigSpec_Zookeeper{Resources: cluster.Config.Kafka.Resources, ZoneIds: s.zones}
	}
	return cluster, nil
}

func (s *kafkaRoleZoneAPI) Update(_ context.Context, req *kafka.UpdateClusterRequest) (*operation.Operation, error) {
	s.updates <- req
	return nil, status.Error(codes.FailedPrecondition, "placement denied by fixture")
}

func kafkaTestZoneList(zones ...string) cty.Value {
	if len(zones) == 0 {
		return cty.ListValEmpty(cty.String)
	}
	values := make([]cty.Value, len(zones))
	for i, zone := range zones {
		values[i] = cty.StringVal(zone)
	}
	return cty.ListVal(values)
}

// Coerce the config the same way for legacy SDK diffs and their raw cty input.
func kafkaRoleConfig(t *testing.T, r *schema.Resource, role string, zones cty.Value, description string) cty.Value {
	t.Helper()
	resources := cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
		"resource_preset_id": cty.StringVal("s2.micro"), "disk_type_id": cty.StringVal("network-ssd"), "disk_size": cty.NumberIntVal(16),
	})})
	roleConfig := map[string]cty.Value{"resources": resources}
	if zones != cty.NilVal {
		roleConfig["zone_ids"] = zones
	}
	raw := cty.ObjectVal(map[string]cty.Value{
		"folder_id": cty.StringVal("test-folder"), "network_id": cty.StringVal("test-network"),
		"name": cty.StringVal("test-cluster"), "description": cty.StringVal(description),
		// Keep unrelated optional/computed collections known in both state and plan.
		"labels":     cty.MapVal(map[string]cty.Value{"team": cty.StringVal("test")}),
		"subnet_ids": cty.ListVal([]cty.Value{cty.StringVal("test-subnet")}),
		"maintenance_window": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
			"type": cty.StringVal("ANYTIME"),
		})}),
		"config": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
			"zones": kafkaTestZoneList("broker"), "version": cty.StringVal("4.1"),
			"kafka": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"resources": resources})}),
			role:    cty.ListVal([]cty.Value{cty.ObjectVal(roleConfig)}),
		})}),
	})
	value, err := r.CoreConfigSchema().CoerceValue(raw)
	require.NoError(t, err)
	return value
}

func TestKafkaClusterRoleZonesCreate(t *testing.T) {
	for _, role := range []string{"zookeeper", "kraft"} {
		t.Run(role, func(t *testing.T) {
			r := svc_kafka.ResourceYandexMDBKafkaCluster()
			raw := kafkaRoleConfig(t, r, role, kafkaTestZoneList("e", "d", "e"), "")
			config := terraform.NewResourceConfigShimmed(raw, r.CoreConfigSchema())
			d := schema.TestResourceDataRaw(t, r.Schema, config.Raw)
			req, err := svc_kafka.PrepareKafkaCreateRequest(d, &providerconfig.Config{})
			require.NoError(t, err)
			require.Equal(t, []string{"broker"}, req.ConfigSpec.ZoneId)
			if role == "kraft" {
				require.Equal(t, []string{"e", "d", "e"}, req.ConfigSpec.GetKraft().GetZoneIds())
				require.Empty(t, req.ConfigSpec.GetZookeeper().GetZoneIds())
			} else {
				require.Equal(t, []string{"e", "d", "e"}, req.ConfigSpec.GetZookeeper().GetZoneIds())
				require.Empty(t, req.ConfigSpec.GetKraft().GetZoneIds())
			}
		})
	}
}

func TestKafkaClusterRoleZonesRead(t *testing.T) {
	for _, role := range []string{"zookeeper", "kraft"} {
		t.Run(role, func(t *testing.T) {
			path := "config.0." + role + ".0.zone_ids"
			r := svc_kafka.ResourceYandexMDBKafkaCluster()
			for _, tc := range []struct {
				name   string
				remote []string
				want   []interface{}
			}{
				{"fresh_import", []string{"d", "e", "e"}, []interface{}{"d", "e", "e"}},
				{"retain_order", []string{"d", "e", "e"}, []interface{}{"e", "d", "e"}},
				{"remote_drift", []string{"d", "d", "e"}, []interface{}{"d", "d", "e"}},
				{"missing_role", nil, nil},
			} {
				t.Run(tc.name, func(t *testing.T) {
					api := &kafkaRoleZoneAPI{role: role, zones: tc.remote}
					if tc.name == "missing_role" {
						api.role = ""
					}
					meta := kafkaClusterTestConfig(t, api)
					state := kafkaSubnetState(nil)
					if tc.name != "fresh_import" {
						raw := kafkaRoleConfig(t, r, role, kafkaTestZoneList("e", "d", "e"), "")
						d := schema.TestResourceDataRaw(t, r.Schema, terraform.NewResourceConfigShimmed(raw, r.CoreConfigSchema()).Raw)
						d.SetId(state.ID)
						state = d.State()
					}
					refreshed, diags := r.RefreshWithoutUpgrade(context.Background(), state, meta)
					require.False(t, diags.HasError(), "%v", diags)
					ds := svc_kafka.DataSourceYandexMDBKafkaCluster()
					reader := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{"cluster_id": state.ID})
					require.NoError(t, ds.Read(reader, meta))
					for _, d := range []*schema.ResourceData{r.Data(refreshed), reader} {
						other := "zookeeper"
						if role == other {
							other = "kraft"
						}
						require.Empty(t, d.Get("config.0."+other))
						if api.role == "" {
							require.Empty(t, d.Get("config.0."+role))
						}
					}
					if api.role != "" {
						require.Equal(t, tc.want, r.Data(refreshed).Get(path))
						values := reader.Get(path).([]interface{})
						require.Len(t, values, len(tc.remote))
						for i, zone := range tc.remote {
							require.Equal(t, zone, values[i])
						}
					}
				})
			}
		})
	}
}

func TestKafkaClusterRoleZonesPlanAndUpdate(t *testing.T) {
	for _, role := range []string{"zookeeper", "kraft"} {
		t.Run(role, func(t *testing.T) {
			r := svc_kafka.ResourceYandexMDBKafkaCluster()
			before := kafkaRoleConfig(t, r, role, kafkaTestZoneList("d", "e", "e"), "")
			initial := schema.TestResourceDataRaw(t, r.Schema, terraform.NewResourceConfigShimmed(before, r.CoreConfigSchema()).Raw)
			initial.SetId("test-cluster")
			for _, tc := range []struct {
				name        string
				config      cty.Value
				description string
				wantMask    []string
				unknown     bool
			}{
				{"reorder_after_import", kafkaTestZoneList("e", "d", "e"), "", nil, false},
				{"omitted", cty.NilVal, "mixed", []string{"description"}, false},
				{"null", cty.NullVal(cty.List(cty.String)), "mixed", []string{"description"}, false},
				{"explicit_empty", kafkaTestZoneList(), "", []string{"config_spec." + role + ".zone_ids"}, false},
				{"changed_counts", kafkaTestZoneList("d", "d", "e"), "mixed", []string{"config_spec." + role + ".zone_ids", "description"}, false},
				{"unknown_list", cty.UnknownVal(cty.List(cty.String)), "", nil, true},
				{"unknown_element", cty.ListVal([]cty.Value{cty.StringVal("e"), cty.UnknownVal(cty.String), cty.StringVal("e")}), "", nil, true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					raw := kafkaRoleConfig(t, r, role, tc.config, tc.description)
					state := initial.State()
					state.RawConfig = raw
					diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigShimmed(raw, r.CoreConfigSchema()), nil)
					require.NoError(t, err)
					if tc.wantMask == nil && !tc.unknown {
						require.True(t, diff.Empty(), "%v", diff)
						return
					}
					require.False(t, diff.Empty())
					require.False(t, diff.RequiresNew())
					if tc.unknown {
						return // The SDK must leave this as a diff until the value is known.
					}
					// Apply uses the raw config from the protocol, not just the legacy typed getters.
					diff.RawConfig = raw
					api := &kafkaRoleZoneAPI{updates: make(chan *kafka.UpdateClusterRequest, 1)}
					offline := kafkaClusterTestConfig(t, api)
					meta := &providerconfig.Config{Token: "t1.fixture.only", Endpoint: "127.0.0.1:0", Plaintext: true}
					require.NoError(t, meta.InitAndValidate(context.Background(), "test", false))
					require.NoError(t, meta.SDK.Shutdown(context.Background()))
					meta.SDK = offline.SDK
					after, diags := r.Apply(context.Background(), state, diff, meta)
					require.True(t, diags.HasError())
					require.Contains(t, diags[0].Summary, "placement denied by fixture")
					require.Equal(t, state.ID, after.ID)
					require.Len(t, api.updates, 1)
					req := <-api.updates
					require.Equal(t, tc.wantMask, req.UpdateMask.Paths)
					require.Equal(t, tc.description, req.Description)
					zones := req.ConfigSpec.GetZookeeper().GetZoneIds()
					if role == "kraft" {
						zones = req.ConfigSpec.GetKraft().GetZoneIds()
					}
					switch tc.name {
					case "explicit_empty":
						require.Empty(t, zones)
					case "changed_counts":
						require.Equal(t, []string{"d", "d", "e"}, zones)
					}
				})
			}
		})
	}
}
