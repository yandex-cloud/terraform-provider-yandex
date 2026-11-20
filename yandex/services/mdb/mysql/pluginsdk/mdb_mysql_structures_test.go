package pluginsdk

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"

	ctyjson "github.com/hashicorp/go-cty/cty/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestFlattenMySQLSettingsEmpty(t *testing.T) {
	t.Parallel()

	config := &mysql.ClusterConfig{}

	m, err := FlattenMySQLConfig(config)

	if err != nil {
		t.Errorf("FlattenMySQLSettings fail: flatten empty error: %v", err)
	}

	if m != nil {
		t.Errorf("FlattenMySQLSettings fail: flatten empty should return nil map but map is: %v", m)
	}
}

func TestFlattenMySQLSettings_5_7(t *testing.T) {
	t.Parallel()

	config := &mysql.ClusterConfig{
		MysqlConfig: &mysql.ClusterConfig_MysqlConfig_5_7{
			MysqlConfig_5_7: &config.MysqlConfigSet5_7{
				UserConfig: &config.MysqlConfig5_7{
					MaxConnections: &wrappers.Int64Value{
						Value: 555,
					},
					InnodbPrintAllDeadlocks: &wrappers.BoolValue{
						Value: true,
					},
					AuditLogPolicy:        config.MysqlConfig5_7_LOGINS,
					InnodbChangeBuffering: config.MysqlConfig5_7_INNODB_CHANGE_BUFFERING_INSERTS,
					LogSlowRateType:       config.MysqlConfig5_7_SESSION,
				},
				EffectiveConfig: &config.MysqlConfig5_7{
					SqlMode: []config.MysqlConfig5_7_SQLMode{
						config.MysqlConfig5_7_NO_BACKSLASH_ESCAPES,
						config.MysqlConfig5_7_STRICT_ALL_TABLES,
					},
				},
			},
		},
	}

	m, err := FlattenMySQLConfig(config)

	if err != nil {
		t.Errorf("FlattenMySQLSettings fail: flatten 5_7 error: %v", err)
	}

	ethalon := map[string]string{
		"audit_log_policy":                       "LOGINS",
		"innodb_change_buffering":                "INNODB_CHANGE_BUFFERING_INSERTS",
		"binlog_transaction_dependency_tracking": "0",
		"max_connections":                        "555",
		"sql_mode":                               "NO_BACKSLASH_ESCAPES,STRICT_ALL_TABLES",
		"innodb_print_all_deadlocks":             "true",
		"log_slow_rate_type":                     "SESSION",
	}

	if !reflect.DeepEqual(ethalon, m) {
		t.Errorf("FlattenMySQLSettings fail: flatten 5_7 should return %v map but map is: %v", ethalon, m)
	}
}

func TestFlattenMySQLSettings_8_0(t *testing.T) {
	t.Parallel()

	config := &mysql.ClusterConfig{
		MysqlConfig: &mysql.ClusterConfig_MysqlConfig_8_0{
			MysqlConfig_8_0: &config.MysqlConfigSet8_0{
				UserConfig: &config.MysqlConfig8_0{
					MaxConnections: &wrappers.Int64Value{
						Value: 555,
					},
					InnodbPrintAllDeadlocks: &wrappers.BoolValue{
						Value: true,
					},
					AuditLogPolicy:        config.MysqlConfig8_0_LOGINS,
					InnodbChangeBuffering: config.MysqlConfig8_0_INNODB_CHANGE_BUFFERING_INSERTS,
					LogSlowRateType:       config.MysqlConfig8_0_SESSION,
				},
				EffectiveConfig: &config.MysqlConfig8_0{
					SqlMode: []config.MysqlConfig8_0_SQLMode{
						config.MysqlConfig8_0_NO_BACKSLASH_ESCAPES,
						config.MysqlConfig8_0_STRICT_ALL_TABLES,
					},
				},
			},
		},
	}

	m, err := FlattenMySQLConfig(config)

	if err != nil {
		t.Errorf("FlattenMySQLSettings fail: flatten 8_0 error: %v", err)
	}

	// TODO remove deep equal
	ethalon := map[string]string{
		"audit_log_policy":                       "LOGINS",
		"innodb_change_buffering":                "INNODB_CHANGE_BUFFERING_INSERTS",
		"binlog_transaction_dependency_tracking": "0",
		"max_connections":                        "555",
		"sql_mode":                               "NO_BACKSLASH_ESCAPES,STRICT_ALL_TABLES",
		"innodb_print_all_deadlocks":             "true",
		"log_slow_rate_type":                     "SESSION",
	}

	if !reflect.DeepEqual(ethalon, m) {
		t.Errorf("FlattenMySQLSettings fail: flatten 8_0 should return %v map but map is: %v", ethalon, m)
	}
}

func TestFlattenMySQLSettings_8_4(t *testing.T) {
	t.Parallel()

	config := &mysql.ClusterConfig{
		MysqlConfig: &mysql.ClusterConfig_MysqlConfig_8_4{
			MysqlConfig_8_4: &config.MysqlConfigSet8_4{
				UserConfig: &config.MysqlConfig8_4{
					MaxConnections: &wrappers.Int64Value{
						Value: 555,
					},
					InnodbPrintAllDeadlocks: &wrappers.BoolValue{
						Value: true,
					},
					AuditLogPolicy:        config.MysqlConfig8_4_LOGINS,
					InnodbChangeBuffering: config.MysqlConfig8_4_INNODB_CHANGE_BUFFERING_INSERTS,
					LogSlowRateType:       config.MysqlConfig8_4_SESSION,
				},
				EffectiveConfig: &config.MysqlConfig8_4{
					SqlMode: []config.MysqlConfig8_4_SQLMode{
						config.MysqlConfig8_4_NO_BACKSLASH_ESCAPES,
						config.MysqlConfig8_4_STRICT_ALL_TABLES,
					},
				},
			},
		},
	}

	m, err := FlattenMySQLConfig(config)

	if err != nil {
		t.Errorf("FlattenMySQLSettings fail: flatten 8_4 error: %v", err)
	}

	// TODO remove deep equal
	ethalon := map[string]string{
		"audit_log_policy":           "LOGINS",
		"innodb_change_buffering":    "INNODB_CHANGE_BUFFERING_INSERTS",
		"max_connections":            "555",
		"sql_mode":                   "NO_BACKSLASH_ESCAPES,STRICT_ALL_TABLES",
		"innodb_print_all_deadlocks": "true",
		"log_slow_rate_type":         "SESSION",
	}

	if !reflect.DeepEqual(ethalon, m) {
		t.Errorf("FlattenMySQLSettings fail: flatten 8_4 should return %v map but map is: %v", ethalon, m)
	}
}

func TestMySQLNamedHostMatcher(t *testing.T) {
	t.Parallel()

	// loaded from YandexCloud API
	existingHostsInfo := map[string]*MyHostInfo{
		"rc1a-myhost.yandex.net": {
			fqdn:              "rc1a-myhost.yandex.net",
			zone:              "rc1a",
			subnetID:          "rc1a-subnet",
			oldAssignPublicIP: false,
		},
		"rc1b-myhost.yandex.net": {
			fqdn:              "rc1b-myhost.yandex.net",
			zone:              "rc1b",
			subnetID:          "rc1b-subnet",
			oldAssignPublicIP: false,
		},
		"rc1с-myhost.yandex.net": {
			fqdn:              "rc1с-myhost.yandex.net",
			zone:              "rc1с",
			subnetID:          "rc1с-subnet",
			oldAssignPublicIP: true,
		},
	}

	// Configuration from '.tf' file
	newHostsInfo := []*MyHostInfo{
		{
			name: "host-in-b",
			zone: "rc1b",
		}, {
			name: "host-in-c",
			zone: "rc1c",
			// it is not possible to change oldAssignPublicIP flag, so such hosts should be re-created
			// (in other words such hosts shouldn't match - and shouldn't be in compareMap)
			oldAssignPublicIP: false,
		}, {
			name:     "host-in-a",
			zone:     "rc1a",
			subnetID: "rc1a-subnet",
		},
	}

	compareMap := CompareMySQLNamedHostsInfo(existingHostsInfo, newHostsInfo)

	compareMapExp := map[int]string{
		0: "rc1b-myhost.yandex.net",
		2: "rc1a-myhost.yandex.net",
	}
	if !reflect.DeepEqual(compareMap, compareMapExp) {
		t.Errorf("TestMySQLHostMatcher() compareMap expected = %v, actual = %v", compareMapExp, compareMap)
	}

}

func TestMySQLNoNamedHostMatcher(t *testing.T) {
	t.Parallel()

	// loaded from YandexCloud API
	existingHostsInfo := map[string]*MyHostInfo{
		"rc1a-myhost.yandex.net": {
			fqdn:              "rc1a-myhost.yandex.net",
			zone:              "rc1a",
			subnetID:          "rc1a-subnet",
			oldAssignPublicIP: false,
		},
		"rc1b-myhost.yandex.net": {
			fqdn:              "rc1b-myhost.yandex.net",
			zone:              "rc1b",
			subnetID:          "rc1b-subnet",
			oldAssignPublicIP: false,
		},
		"rc1с-myhost.yandex.net": {
			fqdn:              "rc1с-myhost.yandex.net",
			zone:              "rc1с",
			subnetID:          "rc1с-subnet",
			oldAssignPublicIP: true,
		},
	}

	// Configuration from '.tf' file
	newHostsInfo := []*MyHostInfo{
		{
			zone: "rc1b",
		}, {
			zone: "rc1c",
			// it is not possible to change oldAssignPublicIP flag, so such hosts should be re-created
			// (in other words such hosts shouldn't match - and shouldn't be in compareMap)
			oldAssignPublicIP: false,
		}, {
			zone:     "rc1a",
			subnetID: "rc1a-subnet",
		},
	}

	compareMap := CompareMySQLNoNamedHostsInfo(existingHostsInfo, newHostsInfo)

	compareMapExp := map[int]string{
		0: "rc1b-myhost.yandex.net",
		2: "rc1a-myhost.yandex.net",
	}
	if !reflect.DeepEqual(compareMap, compareMapExp) {
		t.Errorf("TestMySQLNoNamedHostMatcher() compareMap expected = %v, actual = %+v", compareMapExp, compareMap)
	}
}

func TestMySQLLoopDetector(t *testing.T) {
	t.Parallel()

	err := ValidateMysqlReplicationReferences([]*MySQLHostSpec{
		{
			Name: "",
		}, {
			Name: "",
		},
	})
	assert.NoError(t, err)

	err = ValidateMysqlReplicationReferences([]*MySQLHostSpec{
		{
			Name: "rc1a-one",
		}, {
			Name: "rc1a-two",
		}, {
			Name: "rc1a-three",
		},
	})
	assert.NoError(t, err)

	err = ValidateMysqlReplicationReferences([]*MySQLHostSpec{
		{
			Name: "rc1a-one",
		}, {
			Name: "rc1a-two",
		}, {
			Name:                  "rc1a-three",
			ReplicationSourceName: "rc1a-one",
		},
	})
	assert.NoError(t, err)

	err = ValidateMysqlReplicationReferences([]*MySQLHostSpec{
		{
			Name: "rc1a-one",
		}, {
			Name:                  "rc1a-two",
			ReplicationSourceName: "rc1a-three",
		}, {
			Name:                  "rc1a-three",
			ReplicationSourceName: "rc1a-two",
		},
	})
	assert.EqualError(t, err, "there is no replication chain from HA-hosts to following hosts: 'rc1a-two, rc1a-three' (probably, there is a loop in replication_source chain)")
}

func mysql97TestConfig(version string, settings map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":                  "mysql97-test",
		"environment":           "PRESTABLE",
		"network_id":            "test-network",
		"version":               version,
		"access":                []interface{}{map[string]interface{}{}},
		"backup_window_start":   []interface{}{map[string]interface{}{"hours": 0, "minutes": 0}},
		"connection_manager":    []interface{}{map[string]interface{}{"enabled": true}},
		"disk_size_autoscaling": []interface{}{map[string]interface{}{}},
		"maintenance_window":    []interface{}{map[string]interface{}{"type": "ANYTIME"}},
		"performance_diagnostics": []interface{}{map[string]interface{}{
			"enabled": false, "sessions_sampling_interval": 60, "statements_sampling_interval": 600,
		}},
		"resources": []interface{}{map[string]interface{}{
			"resource_preset_id": "s2.micro",
			"disk_type_id":       "network-ssd",
			"disk_size":          16,
		}},
		"host":         []interface{}{map[string]interface{}{"zone": "ru-central1-a"}},
		"mysql_config": settings,
	}
}

func mysql97ResourceConfig(t *testing.T, raw map[string]interface{}) *terraform.ResourceConfig {
	t.Helper()
	cfg := terraform.NewResourceConfigRaw(raw)
	// ResourceConfigRaw omits the raw cty value that Terraform supplies to diff suppression.
	data, err := json.Marshal(raw)
	require.NoError(t, err)
	var value ctyjson.SimpleJSONValue
	require.NoError(t, json.Unmarshal(data, &value))
	cfg.CtyValue = value.Value
	return cfg
}

func TestMySQL97ConfigRoundTrip(t *testing.T) {
	t.Parallel()
	r := ResourceYandexMDBMySQLCluster()
	settings := map[string]interface{}{
		"max_connections":                             "100",
		"replica_parallel_workers":                    "4",
		"rpl_semi_sync_source_wait_for_replica_count": "1",
		"innodb_redo_log_capacity":                    "268435456",
		"innodb_print_all_deadlocks":                  "false",
		"long_query_time":                             "0",
		"sql_mode":                                    "STRICT_TRANS_TABLES,INTERPRET_UTF8_AS_UTF8MB4",
	}
	_, errs := r.Schema["mysql_config"].ValidateFunc(settings, "mysql_config")
	require.Empty(t, errs)

	d := schema.TestResourceDataRaw(t, r.Schema, mysql97TestConfig("9.7", settings))
	spec, err := expandMySQLConfigSpec(d)
	require.NoError(t, err)
	require.Equal(t, "9.7", spec.Version)
	want := &config.MysqlConfig9_7{
		MaxConnections:                       wrapperspb.Int64(100),
		ReplicaParallelWorkers:               wrapperspb.Int64(4),
		RplSemiSyncSourceWaitForReplicaCount: wrapperspb.Int64(1),
		InnodbRedoLogCapacity:                wrapperspb.Int64(268435456),
		InnodbPrintAllDeadlocks:              wrapperspb.Bool(false),
		LongQueryTime:                        wrapperspb.Double(0),
		SqlMode: []config.MysqlConfig9_7_SQLMode{
			config.MysqlConfig9_7_STRICT_TRANS_TABLES,
			config.MysqlConfig9_7_INTERPRET_UTF8_AS_UTF8MB4,
		},
	}
	require.True(t, proto.Equal(want, spec.GetMysqlConfig_9_7()), "got %v", spec)

	flat, err := FlattenMySQLConfig(&mysql.ClusterConfig{
		MysqlConfig: &mysql.ClusterConfig_MysqlConfig_9_7{
			MysqlConfig_9_7: &config.MysqlConfigSet9_7{UserConfig: want, EffectiveConfig: want},
		},
	})
	require.NoError(t, err)
	for key, value := range settings {
		require.Equal(t, value, flat[key], key)
	}
	require.Len(t, flat, len(settings))

	// A refreshed state must not introduce a new configuration diff.
	d.SetId("test-cluster")
	require.NoError(t, d.Set("mysql_config", flat))
	diff, err := r.Diff(context.Background(), d.State(), mysql97ResourceConfig(t, mysql97TestConfig("9.7", settings)), nil)
	require.NoError(t, err)
	require.True(t, diff.Empty(), "unexpected diff: %v", diff)
}

func TestMySQL97ConfigWithoutUserSettings(t *testing.T) {
	t.Parallel()
	r := ResourceYandexMDBMySQLCluster()
	d := schema.TestResourceDataRaw(t, r.Schema, mysql97TestConfig("9.7", nil))
	spec, err := expandMySQLConfigSpec(d)
	require.NoError(t, err)
	require.Equal(t, "9.7", spec.Version)

	flat, err := FlattenMySQLConfig(&mysql.ClusterConfig{
		MysqlConfig: &mysql.ClusterConfig_MysqlConfig_9_7{
			MysqlConfig_9_7: &config.MysqlConfigSet9_7{
				EffectiveConfig: &config.MysqlConfig9_7{
					SqlMode: []config.MysqlConfig9_7_SQLMode{config.MysqlConfig9_7_INTERPRET_UTF8_AS_UTF8MB4},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Empty(t, flat, "default and effective settings must not become user settings")
}

func TestMySQL97SQLModeRejectedByOlderVersions(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"5.7", "8.0", "8.4"} {
		t.Run(version, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, ResourceYandexMDBMySQLCluster().Schema,
				mysql97TestConfig(version, map[string]interface{}{"sql_mode": "INTERPRET_UTF8_AS_UTF8MB4"}))
			_, err := expandMySQLConfigSpec(d)
			require.Error(t, err)
		})
	}
}

func TestMySQL97RejectsRemovedSettings(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"slave_parallel_workers", "rpl_semi_sync_master_wait_for_slave_count", "innodb_log_file_size"} {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, ResourceYandexMDBMySQLCluster().Schema,
				mysql97TestConfig("9.7", map[string]interface{}{name: "4"}))
			_, err := expandMySQLConfigSpec(d)
			require.ErrorContains(t, err, name)
		})
	}
}

func TestMySQL97SettingsRejectedByOlderVersions(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"5.7", "8.0", "8.4"} {
		for name, value := range map[string]string{
			"replica_parallel_workers":                    "4",
			"rpl_semi_sync_source_wait_for_replica_count": "1",
			"innodb_redo_log_capacity":                    "268435456",
		} {
			t.Run(version+"/"+name, func(t *testing.T) {
				r := ResourceYandexMDBMySQLCluster()
				settings := map[string]interface{}{name: value}
				_, errors := r.Schema["mysql_config"].ValidateFunc(settings, "mysql_config")
				require.Empty(t, errors, "the union schema alone cannot validate a version-specific setting")
				d := schema.TestResourceDataRaw(t, r.Schema, mysql97TestConfig(version, settings))
				_, err := expandMySQLConfigSpec(d)
				require.ErrorContains(t, err, "unsupported MySQL "+version+" setting")
				require.ErrorContains(t, err, name)

				data := mysqlUpdateData(t,
					mysql97TestConfig(version, map[string]interface{}{"max_connections": "100"}),
					mysql97TestConfig(version, settings))
				_, err = prepareMySQLClusterUpdateRequest(data, &providerconfig.Config{})
				require.ErrorContains(t, err, "unsupported MySQL "+version+" setting")
				require.ErrorContains(t, err, name)
			})
		}
	}
}
