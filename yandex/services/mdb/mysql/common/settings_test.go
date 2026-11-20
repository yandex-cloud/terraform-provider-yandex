package common

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestRemoveUnchangedSettings(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		field   string
		actual  *config.MysqlConfig9_7
		desired *config.MysqlConfig9_7
		changed bool
	}{
		{"migrated value", "replica_parallel_workers", &config.MysqlConfig9_7{ReplicaParallelWorkers: wrapperspb.Int64(4)}, &config.MysqlConfig9_7{ReplicaParallelWorkers: wrapperspb.Int64(4)}, false},
		{"changed value", "replica_parallel_workers", &config.MysqlConfig9_7{ReplicaParallelWorkers: wrapperspb.Int64(4)}, &config.MysqlConfig9_7{ReplicaParallelWorkers: wrapperspb.Int64(8)}, true},
		{"reset", "replica_parallel_workers", &config.MysqlConfig9_7{ReplicaParallelWorkers: wrapperspb.Int64(4)}, &config.MysqlConfig9_7{}, true},
		{"already reset", "replica_parallel_workers", &config.MysqlConfig9_7{}, &config.MysqlConfig9_7{}, false},
		{"missing user config", "replica_parallel_workers", nil, &config.MysqlConfig9_7{}, false},
		{"explicit zero", "long_query_time", &config.MysqlConfig9_7{}, &config.MysqlConfig9_7{LongQueryTime: wrapperspb.Double(0)}, true},
		{"unchanged zero", "long_query_time", &config.MysqlConfig9_7{LongQueryTime: wrapperspb.Double(0)}, &config.MysqlConfig9_7{LongQueryTime: wrapperspb.Double(0)}, false},
		{"reset zero", "long_query_time", &config.MysqlConfig9_7{LongQueryTime: wrapperspb.Double(0)}, &config.MysqlConfig9_7{}, true},
		{"explicit false", "innodb_print_all_deadlocks", &config.MysqlConfig9_7{}, &config.MysqlConfig9_7{InnodbPrintAllDeadlocks: wrapperspb.Bool(false)}, true},
		{"reordered SQL modes", "sql_mode", &config.MysqlConfig9_7{SqlMode: []config.MysqlConfig9_7_SQLMode{config.MysqlConfig9_7_STRICT_TRANS_TABLES, config.MysqlConfig9_7_NO_ENGINE_SUBSTITUTION}}, &config.MysqlConfig9_7{SqlMode: []config.MysqlConfig9_7_SQLMode{config.MysqlConfig9_7_NO_ENGINE_SUBSTITUTION, config.MysqlConfig9_7_STRICT_TRANS_TABLES}}, false},
		{"reset SQL mode", "sql_mode", &config.MysqlConfig9_7{SqlMode: []config.MysqlConfig9_7_SQLMode{config.MysqlConfig9_7_STRICT_TRANS_TABLES}}, &config.MysqlConfig9_7{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := "config_spec.mysql_config_9_7." + tc.field
			request := &mysql.UpdateClusterRequest{
				ConfigSpec:  &mysql.ConfigSpec{MysqlConfig: &mysql.ConfigSpec_MysqlConfig_9_7{MysqlConfig_9_7: tc.desired}},
				UpdateMask:  &fieldmaskpb.FieldMask{Paths: []string{path, "description"}},
				Description: "a separate change",
			}
			current := &mysql.ClusterConfig{Version: "9.7", MysqlConfig: &mysql.ClusterConfig_MysqlConfig_9_7{MysqlConfig_9_7: &config.MysqlConfigSet9_7{
				UserConfig: tc.actual,
				// Matching effective/default values do not mean an explicit setting
				// already exists. In particular, zero and false must still be set.
				EffectiveConfig: tc.desired, DefaultConfig: tc.desired,
			}}}
			require.NoError(t, RemoveUnchangedSettings(request, current))
			want := []string{"description"}
			if tc.changed {
				want = append(want, path)
			}
			require.ElementsMatch(t, want, request.UpdateMask.Paths)
		})
	}
}

func TestRemoveUnchangedSettingsVersions(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"5.7", "8.0", "8.4", "9.7"} {
		t.Run(version, func(t *testing.T) {
			field := "mysql_config_" + strings.ReplaceAll(version, ".", "_")
			spec := &mysql.ConfigSpec{}
			current := &mysql.ClusterConfig{}
			require.NoError(t, protojson.Unmarshal([]byte(fmt.Sprintf(`{"%s":{"max_connections":"100"}}`, field)), spec))
			require.NoError(t, protojson.Unmarshal([]byte(fmt.Sprintf(`{"version":"%s","%s":{"user_config":{"max_connections":"100"}}}`, version, field)), current))
			request := &mysql.UpdateClusterRequest{ConfigSpec: spec, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"config_spec." + field + ".max_connections"}}}
			require.NoError(t, RemoveUnchangedSettings(request, current))
			require.Empty(t, request.UpdateMask.Paths)
		})
	}
}
