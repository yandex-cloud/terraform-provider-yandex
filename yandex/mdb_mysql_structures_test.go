package yandex

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"

	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
)

func TestFlattenMySQLSettingsEmpty(t *testing.T) {
	t.Parallel()

	config := &mysql.ClusterConfig{}

	m, err := flattenMySQLSettings(config)

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

	m, err := flattenMySQLSettings(config)

	if err != nil {
		t.Errorf("FlattenMySQLSettings fail: flatten 5_7 error: %v", err)
	}

	ethalon := map[string]string{
		"max_connections":            "555",
		"sql_mode":                   "NO_BACKSLASH_ESCAPES,STRICT_ALL_TABLES",
		"innodb_print_all_deadlocks": "true",
	}

	if !reflect.DeepEqual(m, ethalon) {
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

	m, err := flattenMySQLSettings(config)

	if err != nil {
		t.Errorf("FlattenMySQLSettings fail: flatten 8_0 error: %v", err)
	}

	ethalon := map[string]string{
		"max_connections":            "555",
		"sql_mode":                   "NO_BACKSLASH_ESCAPES,STRICT_ALL_TABLES",
		"innodb_print_all_deadlocks": "true",
	}

	if !reflect.DeepEqual(m, ethalon) {
		t.Errorf("FlattenMySQLSettings fail: flatten 8_0 should return %v map but map is: %v", ethalon, m)
	}
}
