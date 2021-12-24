package yandex

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

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
		"log_slow_rate_type":         "0",
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
		"log_slow_rate_type":         "0",
	}

	if !reflect.DeepEqual(m, ethalon) {
		t.Errorf("FlattenMySQLSettings fail: flatten 8_0 should return %v map but map is: %v", ethalon, m)
	}
}

func TestMySQLNamedHostMatcher(t *testing.T) {
	t.Parallel()

	// loaded from YandexCloud API
	existingHostsInfo := map[string]*myHostInfo{
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
	newHostsInfo := []*myHostInfo{
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

	compareMap := compareMySQLNamedHostsInfo(existingHostsInfo, newHostsInfo)

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
	existingHostsInfo := map[string]*myHostInfo{
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
	newHostsInfo := []*myHostInfo{
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

	compareMap := compareMySQLNoNamedHostsInfo(existingHostsInfo, newHostsInfo)

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

	err := validateMysqlReplicationReferences([]*MySQLHostSpec{
		{
			Name: "",
		}, {
			Name: "",
		},
	})
	assert.NoError(t, err)

	err = validateMysqlReplicationReferences([]*MySQLHostSpec{
		{
			Name: "rc1a-one",
		}, {
			Name: "rc1a-two",
		}, {
			Name: "rc1a-three",
		},
	})
	assert.NoError(t, err)

	err = validateMysqlReplicationReferences([]*MySQLHostSpec{
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

	err = validateMysqlReplicationReferences([]*MySQLHostSpec{
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
