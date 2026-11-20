package pluginsdk

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	msconfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/testfixtures"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestMySQL97UpdateOperations(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name             string
		settings         map[string]interface{}
		previousSettings map[string]interface{}
		failOperation    string
		wantEvents       []string
		wantError        bool
	}{
		{
			name: "failed upgrade", settings: map[string]interface{}{"max_connections": "200"},
			failOperation: "version", wantError: true,
			wantEvents: []string{"update:version", "wait:version"},
		},
		{
			name: "failed settings", settings: map[string]interface{}{"max_connections": "200"},
			failOperation: "settings", wantError: true,
			wantEvents: []string{"update:version", "wait:version", "get", "update:settings", "wait:settings"},
		},
		{
			name: "invalid setting", settings: map[string]interface{}{"slave_parallel_workers": "4"},
			wantError: true,
		},
		{
			name: "unchanged removed setting", settings: map[string]interface{}{"slave_parallel_workers": "4"},
			previousSettings: map[string]interface{}{"slave_parallel_workers": "4"}, wantError: true,
		},
		{
			name: "invalid value", settings: map[string]interface{}{"max_connections": "invalid"},
			wantError: true,
		},
		{
			name: "unchanged settings already match", settings: map[string]interface{}{"max_connections": "100"},
			wantEvents: []string{"update:version", "wait:version", "get"},
		},
		{
			name:       "unmanaged settings",
			wantEvents: []string{"update:version", "wait:version"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previous := tc.previousSettings
			if previous == nil {
				previous = map[string]interface{}{"max_connections": "100"}
			}
			data := mysqlUpdateData(t, mysql97TestConfig("8.4", previous), mysql97TestConfig("9.7", tc.settings))
			api := testfixtures.UpgradeCluster()
			api.FailOperation = tc.failOperation
			api.UpgradedConfig.GetMysqlConfig_9_7().UserConfig = &msconfig.MysqlConfig9_7{MaxConnections: wrapperspb.Int64(100)}
			err := updateMysqlClusterParams(data, mysqlUpdateProvider(t, api))
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.wantEvents, api.Events)
			if len(tc.wantEvents) == 0 {
				require.Empty(t, api.Requests, "invalid settings must be rejected before any Update RPC")
			}
		})
	}
}

func TestMySQL97UpgradeUsesActualSettings(t *testing.T) {
	t.Parallel()
	for _, tc := range testfixtures.UpgradeCases() {
		t.Run(tc.Name, func(t *testing.T) {
			api := testfixtures.UpgradeCluster()
			api.GetError = tc.GetError
			if tc.AlreadyReset {
				api.UpgradedConfig.GetMysqlConfig_9_7().UserConfig = &msconfig.MysqlConfig9_7{MaxConnections: wrapperspb.Int64(100)}
			}
			planned := mysql97TestConfig("9.7", mysqlSettingsMap(tc.Settings))
			planned["description"] = tc.Description
			previous := mysqlSettingsMap(testfixtures.UpgradeSettings("8.4"))
			previous["slave_parallel_type"] = "LOGICAL_CLOCK" // No counterpart in 9.7.
			data := mysqlUpdateData(t, mysql97TestConfig("8.4", previous), planned)
			err := updateMysqlClusterParams(data, mysqlUpdateProvider(t, api))
			wantEvents := []string{"update:version", "wait:version", "get"}
			if tc.GetError != nil {
				require.ErrorContains(t, err, tc.GetError.Error())
			} else {
				require.NoError(t, err)
				actual, err := FlattenMySQLConfig(api.Cluster.Config)
				require.NoError(t, err)
				require.Equal(t, tc.Settings, actual)
			}
			if len(tc.WantPaths) == 0 {
				require.Len(t, api.Requests, 1, "a settings no-op must not reach the API")
			} else {
				require.Len(t, api.Requests, 2)
				require.Empty(t, api.Requests[1].ConfigSpec.Version)
				require.ElementsMatch(t, tc.WantPaths, api.Requests[1].UpdateMask.Paths)
				wantEvents = append(wantEvents, "update:settings", "wait:settings")
			}
			require.Equal(t, wantEvents, api.Events)
			require.Equal(t, "9.7", api.Cluster.Config.Version)
		})
	}
}

func TestMySQL97UpgradeWithUnchangedSettings(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name           string
		settings       map[string]interface{}
		maxConnections *wrapperspb.Int64Value
	}{
		{"max connections", map[string]interface{}{"max_connections": "100"}, wrapperspb.Int64(100)},
		{"explicit empty map", map[string]interface{}{}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := testfixtures.UpgradeWithDefaultRedoCapacity(tc.maxConnections)
			api.Hosts = []*mysql.Host{{Name: "test-host", ZoneId: "ru-central1-d"}}
			target := map[string]interface{}{
				"name":        "test-cluster",
				"environment": "PRESTABLE",
				"network_id":  "test-network",
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s1.micro", "disk_type_id": "network-ssd", "disk_size": 10,
				}},
				"host":         []interface{}{map[string]interface{}{"zone": "ru-central1-d"}},
				"version":      "9.7",
				"mysql_config": tc.settings,
			}
			previous := maps.Clone(target)
			previous["version"] = "8.4"
			data := mysqlUpdateData(t, previous, target)
			require.False(t, data.HasChange("mysql_config"))
			require.NoError(t, resourceYandexMDBMySQLClusterUpdate(data, mysqlUpdateProvider(t, api)))
			require.Len(t, api.Requests, 2)
			require.Equal(t, []string{"config_spec.mysql_config_9_7.innodb_redo_log_capacity"}, api.Requests[1].UpdateMask.Paths)
			require.Nil(t, api.Cluster.Config.GetMysqlConfig_9_7().GetUserConfig().GetInnodbRedoLogCapacity())
			require.Equal(t, tc.settings, data.Get("mysql_config"))

			diff, err := ResourceYandexMDBMySQLCluster().Diff(context.Background(), data.State(), mysql97ResourceConfig(t, target), nil)
			require.NoError(t, err)
			require.True(t, diff.Empty(), "unexpected second plan: %s", diff)
		})
	}
}

func TestMySQL97ResetSettings(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		settings map[string]interface{}
		want     *msconfig.MysqlConfig9_7
	}{
		{
			name: "remove setting", settings: map[string]interface{}{"max_connections": "100", "sql_mode": "INTERPRET_UTF8_AS_UTF8MB4"},
			want: &msconfig.MysqlConfig9_7{MaxConnections: wrapperspb.Int64(100), SqlMode: []msconfig.MysqlConfig9_7_SQLMode{msconfig.MysqlConfig9_7_INTERPRET_UTF8_AS_UTF8MB4}},
		},
		{
			name: "reset SQL mode", settings: map[string]interface{}{"max_connections": "100", "replica_parallel_workers": "4"},
			want: &msconfig.MysqlConfig9_7{MaxConnections: wrapperspb.Int64(100), ReplicaParallelWorkers: wrapperspb.Int64(4)},
		},
		{name: "reset all", settings: map[string]interface{}{}, want: &msconfig.MysqlConfig9_7{}},
		{name: "unmanaged"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previous := map[string]interface{}{"max_connections": "100", "replica_parallel_workers": "4", "sql_mode": "INTERPRET_UTF8_AS_UTF8MB4"}
			api := &testfixtures.ClusterUpdates{Cluster: testfixtures.MySQLCluster()}
			api.Cluster.Config.GetMysqlConfig_9_7().UserConfig = &msconfig.MysqlConfig9_7{
				MaxConnections: wrapperspb.Int64(100), ReplicaParallelWorkers: wrapperspb.Int64(4),
				SqlMode: []msconfig.MysqlConfig9_7_SQLMode{msconfig.MysqlConfig9_7_INTERPRET_UTF8_AS_UTF8MB4},
			}
			data := mysqlUpdateData(t, mysql97TestConfig("9.7", previous), mysql97TestConfig("9.7", tc.settings))
			require.NoError(t, updateMysqlClusterParams(data, mysqlUpdateProvider(t, api)))
			if tc.settings == nil {
				require.Empty(t, api.Requests, "omitted configuration remains unmanaged")
				return
			}
			require.Equal(t, []string{"update:settings", "wait:settings"}, api.Events)
			require.True(t, proto.Equal(tc.want, api.Cluster.Config.GetMysqlConfig_9_7().UserConfig))
		})
	}
}

func mysqlUpdateData(t *testing.T, previous, planned map[string]interface{}) *schema.ResourceData {
	t.Helper()
	r := ResourceYandexMDBMySQLCluster()
	old := schema.TestResourceDataRaw(t, r.Schema, previous)
	old.SetId("test-id")
	raw := mysql97ResourceConfig(t, planned)
	diff, err := r.Diff(context.Background(), old.State(), raw, nil)
	require.NoError(t, err)
	require.False(t, diff.RequiresNew(), "updating settings/version must preserve the cluster")
	if diff != nil {
		diff.RawConfig = raw.CtyValue
	}
	data, err := schema.InternalMap(r.Schema).Data(old.State(), diff)
	require.NoError(t, err)
	return data
}

func mysqlUpdateProvider(t *testing.T, api *testfixtures.ClusterUpdates) *providerconfig.Config {
	t.Helper()
	cfg := &providerconfig.Config{Token: "test-token"}
	// Initialize the provider context without credentials files or discovery calls.
	require.NoError(t, cfg.InitAndValidate(context.Background(), "test", false))
	require.NoError(t, cfg.SDK.Shutdown(context.Background()))
	cfg.SDK = api.SDK(t)
	return cfg
}

func mysqlSettingsMap(settings map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(settings))
	for name, value := range settings {
		result[name] = value
	}
	return result
}
