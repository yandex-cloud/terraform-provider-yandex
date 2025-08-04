package mdb_postgresql_cluster_v2

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func TestYandexProvider_MDBPostgresClusterConfigAccessFlattener(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expectedAccessAttrs := map[string]attr.Type{
		"data_lens":     types.BoolType,
		"data_transfer": types.BoolType,
		"serverless":    types.BoolType,
		"web_sql":       types.BoolType,
	}

	cases := []struct {
		testname    string
		reqVal      *postgresql.Access
		expectedVal types.Object
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &postgresql.Access{
				WebSql:   true,
				DataLens: true,
			},
			expectedVal: types.ObjectValueMust(
				expectedAccessAttrs, map[string]attr.Value{
					"data_lens":     types.BoolValue(true),
					"data_transfer": types.BoolValue(false),
					"serverless":    types.BoolValue(false),
					"web_sql":       types.BoolValue(true),
				},
			),
		},
		{
			testname:    "CheckNullObject",
			reqVal:      nil,
			expectedVal: types.ObjectNull(expectedAccessAttrs),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		access := flattenAccess(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected flatten diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !c.expectedVal.Equal(access) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				access,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterConfigPerfomanceDiagnosticsFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cases := []struct {
		testname       string
		testData       *postgresql.PerformanceDiagnostics
		expectedObject types.Object
		hasErr         bool
	}{
		{
			testname:       "CheckNullObject",
			testData:       nil,
			expectedObject: types.ObjectNull(expectedPDAttrs),
		},
		{
			testname: "CheckAllAttributes",
			testData: &postgresql.PerformanceDiagnostics{
				Enabled:                    true,
				SessionsSamplingInterval:   10,
				StatementsSamplingInterval: 5,
			},
			expectedObject: types.ObjectValueMust(expectedPDAttrs, map[string]attr.Value{
				"enabled":                      types.BoolValue(true),
				"sessions_sampling_interval":   types.Int64Value(10),
				"statements_sampling_interval": types.Int64Value(5),
			}),
		},
		{
			testname: "CheckPartialAttributes",
			testData: &postgresql.PerformanceDiagnostics{
				Enabled:                  true,
				SessionsSamplingInterval: 10,
			},
			expectedObject: types.ObjectValueMust(expectedPDAttrs, map[string]attr.Value{
				"enabled":                      types.BoolValue(true),
				"sessions_sampling_interval":   types.Int64Value(10),
				"statements_sampling_interval": types.Int64Value(0),
			}),
		},
		{
			testname: "CheckEmptyAttributes",
			testData: &postgresql.PerformanceDiagnostics{},
			expectedObject: types.ObjectValueMust(expectedPDAttrs, map[string]attr.Value{
				"enabled":                      types.BoolValue(false),
				"sessions_sampling_interval":   types.Int64Value(0),
				"statements_sampling_interval": types.Int64Value(0),
			}),
		},
	}

	for _, c := range cases {
		var diags diag.Diagnostics
		res := flattenPerformanceDiagnostics(ctx, c.testData, &diags)
		if c.hasErr {
			if !diags.HasError() {
				t.Errorf("Unexpected flatten error status: expected %v, actual %v", c.hasErr, diags.HasError())
			}
			continue
		}

		if !c.expectedObject.Equal(res) {
			t.Errorf("Unexpected flatten object result: expected %v, actual %v", c.expectedObject, res)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterConfigBackupRetainPeriodDaysFlattener(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *wrapperspb.Int64Value
		expectedVal types.Int64
	}{
		{
			testname: "ExplicitCheck",
			reqVal: &wrapperspb.Int64Value{
				Value: 5,
			},
			expectedVal: types.Int64Value(5),
		},
		{
			testname:    "NullCheck",
			reqVal:      nil,
			expectedVal: types.Int64Null(),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		brPd := flattenBackupRetainPeriodDays(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected flatten diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !c.expectedVal.Equal(brPd) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				brPd,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterMapStringFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      map[string]string
		expectedVal types.Map
	}{
		{
			testname: "CheckSeveralAttributes",
			reqVal: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectedVal: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			),
		},
		{
			testname: "CheckOnelAttribute",
			reqVal: map[string]string{
				"key": "value",
			},
			expectedVal: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"key": types.StringValue("value"),
				},
			),
		},
		{
			testname: "CheckEmptyAttribute",
			reqVal:   map[string]string{},
			expectedVal: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{},
			),
		},
		{
			testname: "CheckNullAttribute",
			reqVal:   nil,
			expectedVal: types.MapNull(
				types.StringType,
			),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		m := flattenMapString(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected flatten diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !c.expectedVal.Equal(m) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				m,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterSetStringFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      []string
		expectedVal types.Set
	}{
		{
			testname: "CheckSeveralAttributes",
			reqVal:   []string{"key1", "key2"},
			expectedVal: types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("key1"),
					types.StringValue("key2"),
				},
			),
		},
		{
			testname: "CheckOneAttribute",
			reqVal:   []string{"key"},
			expectedVal: types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("key"),
				},
			),
		},
		{
			testname: "CheckNullAttribute",
			reqVal:   nil,
			expectedVal: types.SetValueMust(
				types.StringType,
				[]attr.Value{},
			),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		m := flattenSetString(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected flatten diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !c.expectedVal.Equal(m) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				m,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterBoolWrapperFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *wrapperspb.BoolValue
		expectedVal types.Bool
	}{
		{
			testname:    "CheckExplicitAttribute",
			reqVal:      wrapperspb.Bool(true),
			expectedVal: types.BoolValue(true),
		},
		{
			testname:    "CheckNullAttribute",
			reqVal:      nil,
			expectedVal: types.BoolNull(),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		m := flattenBoolWrapper(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected flatten diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !c.expectedVal.Equal(m) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				m,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterConfigDiskSizeAutoscalingFlattener(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *postgresql.DiskSizeAutoscaling
		expectedVal types.Object
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &postgresql.DiskSizeAutoscaling{
				DiskSizeLimit:           datasize.ToBytes(3),
				PlannedUsageThreshold:   50,
				EmergencyUsageThreshold: 70,
			},
			expectedVal: types.ObjectValueMust(
				expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
					"disk_size_limit":           types.Int64Value(3),
					"planned_usage_threshold":   types.Int64Value(50),
					"emergency_usage_threshold": types.Int64Value(70),
				},
			),
		},
		{
			testname: "CheckRequiredAttribute",
			reqVal: &postgresql.DiskSizeAutoscaling{
				DiskSizeLimit: datasize.ToBytes(5),
			},
			expectedVal: types.ObjectValueMust(
				expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
					"disk_size_limit":           types.Int64Value(5),
					"planned_usage_threshold":   types.Int64Value(0),
					"emergency_usage_threshold": types.Int64Value(0),
				},
			),
		},
		{
			testname: "CheckAttributesWithPlannedUsageThresold",
			reqVal: &postgresql.DiskSizeAutoscaling{
				DiskSizeLimit:         datasize.ToBytes(5),
				PlannedUsageThreshold: 10,
			},
			expectedVal: types.ObjectValueMust(
				expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
					"disk_size_limit":           types.Int64Value(5),
					"planned_usage_threshold":   types.Int64Value(10),
					"emergency_usage_threshold": types.Int64Value(0),
				},
			),
		},
		{
			testname: "CheckAttributesWithEmergencyUsageThresold",
			reqVal: &postgresql.DiskSizeAutoscaling{
				DiskSizeLimit:           datasize.ToBytes(6),
				EmergencyUsageThreshold: 10,
			},
			expectedVal: types.ObjectValueMust(
				expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
					"disk_size_limit":           types.Int64Value(6),
					"planned_usage_threshold":   types.Int64Value(0),
					"emergency_usage_threshold": types.Int64Value(10),
				},
			),
		},
		{
			testname: "CheckNullObject",
			reqVal:   nil,
			expectedVal: types.ObjectValueMust(expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
				"disk_size_limit":           types.Int64Value(0),
				"planned_usage_threshold":   types.Int64Value(0),
				"emergency_usage_threshold": types.Int64Value(0),
			}),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		dsaObj := flattenDiskSizeAutoscaling(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected flatten diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !c.expectedVal.Equal(dsaObj) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				dsaObj,
			)
		}
	}
}

type invalidPgConfig struct {
	postgresql.ClusterConfig_PostgresqlConfig
}

func TestYandexProvider_MDBPostgresClusterConfigPostgresqlConfigFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        postgresql.ClusterConfig_PostgresqlConfig
		expectedVal   mdbcommon.SettingsMapValue
		expectedError bool
	}{
		{
			testname: "CheckFullAttributes",
			reqVal: &postgresql.ClusterConfig_PostgresqlConfig_14{
				PostgresqlConfig_14: &config.PostgresqlConfigSet14{
					UserConfig: &config.PostgresqlConfig14{
						MaxConnections:      wrapperspb.Int64(14),
						WorkMem:             wrapperspb.Int64(50),
						BgwriterLruMaxpages: wrapperspb.Int64(100),
						WalLevel:            config.PostgresqlConfig14_WAL_LEVEL_LOGICAL,
						LogConnections:      wrapperspb.Bool(true),
						CursorTupleFraction: wrapperspb.Double(1.2),
						SharedPreloadLibraries: []config.PostgresqlConfig14_SharedPreloadLibraries{
							config.PostgresqlConfig14_SHARED_PRELOAD_LIBRARIES_PGLOGICAL, config.PostgresqlConfig14_SHARED_PRELOAD_LIBRARIES_PGLOGICAL,
						},
						SearchPath:     "path",
						BackslashQuote: config.PostgresqlConfig14_BACKSLASH_QUOTE_ON,
					},
				},
			},
			expectedVal: mdbcommon.SettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"max_connections":          types.StringValue("14"),
						"work_mem":                 types.StringValue("50"),
						"bgwriter_lru_maxpages":    types.StringValue("100"),
						"wal_level":                types.StringValue(config.PostgresqlConfig14_WAL_LEVEL_LOGICAL.String()),
						"log_connections":          types.StringValue("true"),
						"cursor_tuple_fraction":    types.StringValue("1.20"),
						"shared_preload_libraries": types.StringValue(fmt.Sprintf("%s,%s", config.PostgresqlConfig14_SHARED_PRELOAD_LIBRARIES_PGLOGICAL.String(), config.PostgresqlConfig14_SHARED_PRELOAD_LIBRARIES_PGLOGICAL.String())),
						"search_path":              types.StringValue("path"),
						"backslash_quote":          types.StringValue(config.PostgresqlConfig14_BACKSLASH_QUOTE_ON.String()),
					},
				),
			},
		},
		{
			testname: "CheckFullAttributes2",
			reqVal: &postgresql.ClusterConfig_PostgresqlConfig_14_1C{
				PostgresqlConfig_14_1C: &config.PostgresqlConfigSet14_1C{
					UserConfig: &config.PostgresqlConfig14_1C{
						WalLevel:            config.PostgresqlConfig14_1C_WAL_LEVEL_REPLICA,
						LogConnections:      wrapperspb.Bool(false),
						CursorTupleFraction: wrapperspb.Double(144.53425),
						SharedPreloadLibraries: []config.PostgresqlConfig14_1C_SharedPreloadLibraries{
							config.PostgresqlConfig14_1C_SHARED_PRELOAD_LIBRARIES_PGAUDIT,
						},
					},
				},
			},
			expectedVal: mdbcommon.SettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"wal_level":                types.StringValue(config.PostgresqlConfig14_1C_WAL_LEVEL_REPLICA.String()),
						"log_connections":          types.StringValue("false"),
						"cursor_tuple_fraction":    types.StringValue("144.53"),
						"shared_preload_libraries": types.StringValue(config.PostgresqlConfig14_1C_SHARED_PRELOAD_LIBRARIES_PGAUDIT.String()),
					},
				),
			},
		},
		{
			testname:    "CheckNull",
			reqVal:      nil,
			expectedVal: mdbcommon.SettingsMapValue{MapValue: types.MapValueMust(types.StringType, map[string]attr.Value{})},
		},
		{
			testname:      "CheckInvalidStructure",
			reqVal:        invalidPgConfig{},
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}

		conf := flattenPostgresqlConfig(ctx, c.reqVal, &diags)
		if diags.HasError() != c.expectedError {
			if !c.expectedError {
				t.Errorf(
					"Unexpected flatten diagnostics status %s test: errors: %v",
					c.testname,
					diags.Errors(),
				)
			} else {
				t.Errorf(
					"Unexpected flatten diagnostics status %s test: expected error, actual not",
					c.testname,
				)
			}

			continue
		}

		if !c.expectedError && !c.expectedVal.Equal(conf) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				conf,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterConfigPoolerConfigFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *postgresql.ConnectionPoolerConfig
		expectedVal types.Object
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &postgresql.ConnectionPoolerConfig{
				PoolDiscard: wrapperspb.Bool(true),
				PoolingMode: postgresql.ConnectionPoolerConfig_SESSION,
			},
			expectedVal: types.ObjectValueMust(
				expectedPCAttrTypes, map[string]attr.Value{
					"pooling_mode": types.StringValue("SESSION"),
					"pool_discard": types.BoolValue(true),
				},
			),
		},
		{
			testname: "CheckAllAttributesWithDefaultValues",
			reqVal:   &postgresql.ConnectionPoolerConfig{},
			expectedVal: types.ObjectValueMust(
				expectedPCAttrTypes, map[string]attr.Value{
					"pooling_mode": types.StringValue("POOLING_MODE_UNSPECIFIED"),
					"pool_discard": types.BoolNull(),
				},
			),
		},
		{
			testname: "CheckPartlyAttributesWithPM",
			reqVal: &postgresql.ConnectionPoolerConfig{
				PoolingMode: postgresql.ConnectionPoolerConfig_SESSION,
			},
			expectedVal: types.ObjectValueMust(
				expectedPCAttrTypes, map[string]attr.Value{
					"pooling_mode": types.StringValue("SESSION"),
					"pool_discard": types.BoolNull(),
				},
			),
		},
		{
			testname: "CheckPartlyAttributesWithPD",
			reqVal: &postgresql.ConnectionPoolerConfig{
				PoolDiscard: wrapperspb.Bool(true),
			},
			expectedVal: types.ObjectValueMust(
				expectedPCAttrTypes, map[string]attr.Value{
					"pooling_mode": types.StringValue("POOLING_MODE_UNSPECIFIED"),
					"pool_discard": types.BoolValue(true),
				},
			),
		},
		{
			testname: "CheckNullObject",
			reqVal:   nil,
			expectedVal: types.ObjectValueMust(expectedPCAttrTypes, map[string]attr.Value{
				"pooling_mode": types.StringValue("POOLING_MODE_UNSPECIFIED"),
				"pool_discard": types.BoolNull(),
			}),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		pc := flattenPoolerConfig(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected flatten diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !c.expectedVal.Equal(pc) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				pc,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterConfigFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        *postgresql.ClusterConfig
		expectedVal   types.Object
		expectedError bool
	}{
		{
			testname: "CheckFullAttributes",
			reqVal: &postgresql.ClusterConfig{
				Version: "9.6",
				Resources: &postgresql.Resources{
					ResourcePresetId: "s1.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         datasize.ToBytes(10),
				},
				Autofailover: wrapperspb.Bool(true),
				Access: &postgresql.Access{
					DataLens:     true,
					DataTransfer: true,
				},
				PerformanceDiagnostics: &postgresql.PerformanceDiagnostics{
					Enabled:                    true,
					SessionsSamplingInterval:   60,
					StatementsSamplingInterval: 600,
				},
				BackupWindowStart: &timeofday.TimeOfDay{
					Hours:   10,
					Minutes: 0,
				},
				BackupRetainPeriodDays: wrapperspb.Int64(7),
				PostgresqlConfig: &postgresql.ClusterConfig_PostgresqlConfig_12{
					PostgresqlConfig_12: &config.PostgresqlConfigSet12{
						UserConfig: &config.PostgresqlConfig12{
							ExitOnError: wrapperspb.Bool(true),
						},
					},
				},
				PoolerConfig: &postgresql.ConnectionPoolerConfig{
					PoolDiscard: wrapperspb.Bool(true),
					PoolingMode: postgresql.ConnectionPoolerConfig_STATEMENT,
				},
				DiskSizeAutoscaling: &postgresql.DiskSizeAutoscaling{
					DiskSizeLimit:           datasize.ToBytes(5),
					EmergencyUsageThreshold: 20,
					PlannedUsageThreshold:   30,
				},
			},
			expectedVal: types.ObjectValueMust(
				expectedConfigAttrs, map[string]attr.Value{
					"version": types.StringValue("9.6"),
					"resources": types.ObjectValueMust(mdbcommon.ResourceType.AttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					}),
					"autofailover": types.BoolValue(true),
					"access": types.ObjectValueMust(expectedAccessAttrTypes, map[string]attr.Value{
						"data_lens":     types.BoolValue(true),
						"data_transfer": types.BoolValue(true),
						"serverless":    types.BoolValue(false),
						"web_sql":       types.BoolValue(false),
					}),
					"performance_diagnostics": types.ObjectValueMust(expectedPDAttrs, map[string]attr.Value{
						"enabled":                      types.BoolValue(true),
						"sessions_sampling_interval":   types.Int64Value(60),
						"statements_sampling_interval": types.Int64Value(600),
					}),
					"backup_window_start": types.ObjectValueMust(mdbcommon.BackupWindowType.AttrTypes, map[string]attr.Value{
						"hours":   types.Int64Value(10),
						"minutes": types.Int64Value(0),
					}),
					"backup_retain_period_days": types.Int64Value(7),
					"postgresql_config": NewPgSettingsMapValueMust(map[string]attr.Value{
						"exit_on_error": types.StringValue("true"),
					}),
					"pooler_config": types.ObjectValueMust(expectedPCAttrTypes, map[string]attr.Value{
						"pool_discard": types.BoolValue(true),
						"pooling_mode": types.StringValue(postgresql.ConnectionPoolerConfig_STATEMENT.String()),
					}),
					"disk_size_autoscaling": types.ObjectValueMust(expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
						"disk_size_limit":           types.Int64Value(5),
						"emergency_usage_threshold": types.Int64Value(20),
						"planned_usage_threshold":   types.Int64Value(30),
					}),
				},
			),
		},
		{
			testname: "CheckPartlyAttributes",
			reqVal: &postgresql.ClusterConfig{
				Version: "15",
				Resources: &postgresql.Resources{
					ResourcePresetId: "s2.nano",
					DiskTypeId:       "network-hdd",
					DiskSize:         datasize.ToBytes(15),
				},
				PostgresqlConfig: nil,
			},
			expectedVal: types.ObjectValueMust(
				expectedConfigAttrs, map[string]attr.Value{
					"version": types.StringValue("15"),
					"resources": types.ObjectValueMust(mdbcommon.ResourceType.AttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s2.nano"),
						"disk_type_id":       types.StringValue("network-hdd"),
						"disk_size":          types.Int64Value(15),
					}),
					"autofailover":              types.BoolNull(),
					"access":                    types.ObjectNull(expectedAccessAttrTypes),
					"performance_diagnostics":   types.ObjectNull(expectedPDAttrs),
					"backup_window_start":       types.ObjectNull(mdbcommon.BackupWindowType.AttrTypes),
					"backup_retain_period_days": types.Int64Null(),
					"postgresql_config":         NewPgSettingsMapValueMust(map[string]attr.Value{}),
					"pooler_config": types.ObjectValueMust(expectedPCAttrTypes, map[string]attr.Value{
						"pool_discard": types.BoolNull(),
						"pooling_mode": types.StringValue(postgresql.ConnectionPoolerConfig_POOLING_MODE_UNSPECIFIED.String()),
					}),
					"disk_size_autoscaling": types.ObjectValueMust(expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
						"disk_size_limit":           types.Int64Value(0),
						"planned_usage_threshold":   types.Int64Value(0),
						"emergency_usage_threshold": types.Int64Value(0),
					}),
				},
			),
		},
		{
			testname:      "CheckNull",
			reqVal:        nil,
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}

		conf := flattenConfig(ctx, mdbcommon.SettingsMapValue{MapValue: types.MapNull(types.StringType)}, c.reqVal, &diags)
		if diags.HasError() != c.expectedError {
			if !c.expectedError {
				t.Errorf(
					"Unexpected flatten diagnostics status %s test: errors: %v",
					c.testname,
					diags.Errors(),
				)
			} else {
				t.Errorf(
					"Unexpected flatten diagnostics status %s test: expected error, actual not",
					c.testname,
				)
			}

			continue
		}

		if !c.expectedVal.Equal(conf) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				conf,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterConfigFlattenPgConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	reqVal := &postgresql.ClusterConfig{
		Version: "9.6",
		Resources: &postgresql.Resources{
			ResourcePresetId: "s1.micro",
			DiskTypeId:       "network-ssd",
			DiskSize:         datasize.ToBytes(10),
		},
		Autofailover: wrapperspb.Bool(true),
		Access: &postgresql.Access{
			DataLens:     true,
			DataTransfer: true,
		},
		PerformanceDiagnostics: &postgresql.PerformanceDiagnostics{
			Enabled:                    true,
			SessionsSamplingInterval:   60,
			StatementsSamplingInterval: 600,
		},
		BackupWindowStart: &timeofday.TimeOfDay{
			Hours:   10,
			Minutes: 0,
		},
		BackupRetainPeriodDays: wrapperspb.Int64(7),
		PostgresqlConfig: &postgresql.ClusterConfig_PostgresqlConfig_12{
			PostgresqlConfig_12: &config.PostgresqlConfigSet12{
				UserConfig: &config.PostgresqlConfig12{
					MaxConnections: wrapperspb.Int64(200),
					ExitOnError:    wrapperspb.Bool(true),
				},
			},
		},
	}

	expectedVal := mdbcommon.SettingsMapValue{
		MapValue: types.MapValueMust(
			types.StringType,
			map[string]attr.Value{
				"max_connections": types.StringValue("100"),
			},
		),
	}

	var diags diag.Diagnostics
	conf := flattenConfig(
		ctx, expectedVal,
		reqVal, &diags,
	)

	if diags.HasError() {
		t.Errorf(
			"Unexpected flatten diagnostics status test: errors: %v",
			diags.Errors(),
		)
	}

	pgConfResult, ok := conf.Attributes()["postgresql_config"]
	if !ok {
		t.Error("Unexpected flatten result value test: expected postgresql_confing in config, actual not foudnd")
	}

	if !expectedVal.Equal(pgConfResult) {
		t.Errorf(
			"Unexpected flatten result value test: expected %s, actual %s",
			expectedVal,
			pgConfResult,
		)
	}
}
