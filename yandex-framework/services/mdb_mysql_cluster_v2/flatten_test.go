package mdb_mysql_cluster_v2

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"

	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func TestYandexProvider_MDBMySQLClusterConfigAccessFlattener(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expectedAccessAttrs := map[string]attr.Type{
		"data_lens":     types.BoolType,
		"data_transfer": types.BoolType,
		"web_sql":       types.BoolType,
		"yandex_query":  types.BoolType,
	}

	cases := []struct {
		testname    string
		reqVal      *mysql.Access
		expectedVal types.Object
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &mysql.Access{
				WebSql:      true,
				DataLens:    true,
				YandexQuery: true,
			},
			expectedVal: types.ObjectValueMust(
				expectedAccessAttrs, map[string]attr.Value{
					"data_lens":     types.BoolValue(true),
					"data_transfer": types.BoolValue(false),
					"web_sql":       types.BoolValue(true),
					"yandex_query":  types.BoolValue(true),
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

func TestYandexProvider_MDBMySQLClusterConfigPerfomanceDiagnosticsFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cases := []struct {
		testname       string
		testData       *mysql.PerformanceDiagnostics
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
			testData: &mysql.PerformanceDiagnostics{
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
			testData: &mysql.PerformanceDiagnostics{
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
			testData: &mysql.PerformanceDiagnostics{},
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

func TestYandexProvider_MDBMySQLClusterConfigDiskSizeAutoscaleFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cases := []struct {
		testname       string
		testData       *mysql.DiskSizeAutoscaling
		expectedObject types.Object
		hasErr         bool
	}{
		{
			testname:       "CheckNullObject",
			testData:       nil,
			expectedObject: types.ObjectNull(expectedDSAAttrs),
		},
		{
			testname: "CheckAllAttributes",
			testData: &mysql.DiskSizeAutoscaling{
				DiskSizeLimit:           datasize.ToBytes(20),
				PlannedUsageThreshold:   30,
				EmergencyUsageThreshold: 60,
			},
			expectedObject: types.ObjectValueMust(expectedDSAAttrs, map[string]attr.Value{
				"disk_size_limit":           types.Int64Value(20),
				"planned_usage_threshold":   types.Int64Value(30),
				"emergency_usage_threshold": types.Int64Value(60),
			}),
		},
		{
			testname: "CheckPartialAttributes",
			testData: &mysql.DiskSizeAutoscaling{
				DiskSizeLimit:           datasize.ToBytes(20),
				EmergencyUsageThreshold: 60,
			},
			expectedObject: types.ObjectValueMust(expectedDSAAttrs, map[string]attr.Value{
				"disk_size_limit":           types.Int64Value(20),
				"planned_usage_threshold":   types.Int64Value(0),
				"emergency_usage_threshold": types.Int64Value(60),
			}),
		},
		{
			testname: "CheckEmptyAttributes",
			testData: &mysql.DiskSizeAutoscaling{},
			expectedObject: types.ObjectValueMust(expectedDSAAttrs, map[string]attr.Value{
				"disk_size_limit":           types.Int64Value(0),
				"planned_usage_threshold":   types.Int64Value(0),
				"emergency_usage_threshold": types.Int64Value(0),
			}),
		},
	}

	for _, c := range cases {
		var diags diag.Diagnostics
		res := flattenDiskSizeAutoscaling(ctx, c.testData, &diags)
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

type invalidMsConfig struct {
	mysql.ClusterConfig_MysqlConfig
}

func TestYandexProvider_MDBMysqlClusterConfigMysqlConfigFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        mysql.ClusterConfig_MysqlConfig
		expectedVal   mdbcommon.SettingsMapValue
		expectedError bool
	}{
		{
			testname: "CheckFullAttributes",
			reqVal: &mysql.ClusterConfig_MysqlConfig_5_7{
				MysqlConfig_5_7: &config.MysqlConfigSet5_7{
					UserConfig: &config.MysqlConfig5_7{
						SqlMode: []config.MysqlConfig5_7_SQLMode{
							config.MysqlConfig5_7_ONLY_FULL_GROUP_BY,
							config.MysqlConfig5_7_STRICT_TRANS_TABLES,
							config.MysqlConfig5_7_NO_ZERO_IN_DATE,
						},
						MaxConnections:              wrapperspb.Int64(100),
						DefaultAuthenticationPlugin: config.MysqlConfig5_7_MYSQL_NATIVE_PASSWORD,
						InnodbPrintAllDeadlocks:     wrapperspb.Bool(true),
						LongQueryTime:               wrapperspb.Double(5.24),
						DefaultTimeZone:             "UTC",
					},
				},
			},
			expectedVal: mdbcommon.SettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"sql_mode":                      types.StringValue("ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE"),
						"max_connections":               types.StringValue("100"),
						"default_authentication_plugin": types.StringValue("MYSQL_NATIVE_PASSWORD"),
						"innodb_print_all_deadlocks":    types.StringValue("true"),
						"long_query_time":               types.StringValue("5.24"),
						"default_time_zone":             types.StringValue("UTC"),
					},
				),
			},
		},
		{
			testname: "CheckFullAttributes2",
			reqVal: &mysql.ClusterConfig_MysqlConfig_8_0{
				MysqlConfig_8_0: &config.MysqlConfigSet8_0{
					UserConfig: &config.MysqlConfig8_0{
						SqlMode: []config.MysqlConfig8_0_SQLMode{
							config.MysqlConfig8_0_NO_UNSIGNED_SUBTRACTION,
						},
						DefaultAuthenticationPlugin: config.MysqlConfig8_0_SHA256_PASSWORD,
						InnodbPrintAllDeadlocks:     wrapperspb.Bool(false),
					},
				},
			},
			expectedVal: mdbcommon.SettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"sql_mode":                      types.StringValue(config.MysqlConfig8_0_NO_UNSIGNED_SUBTRACTION.String()),
						"default_authentication_plugin": types.StringValue(config.MysqlConfig8_0_SHA256_PASSWORD.String()),
						"innodb_print_all_deadlocks":    types.StringValue("false"),
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
			reqVal:        invalidMsConfig{},
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}

		conf := flattenMySQLConfig(ctx, c.reqVal, &diags)
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
				"Unexpected flatten result value %s test: \nexpected %v\n, actual %v\n",
				c.testname,
				c.expectedVal,
				conf,
			)
		}
	}
}

func TestYandexProvider_MDBMySQLClusterConfigFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        *mysql.ClusterConfig
		expectedVal   Config
		expectedError bool
	}{
		{
			testname: "CheckFullAttributes",
			reqVal: &mysql.ClusterConfig{
				Version: "5.7",
				Resources: &mysql.Resources{
					ResourcePresetId: "s1.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         datasize.ToBytes(10),
				},
				Access: &mysql.Access{
					DataLens:     true,
					DataTransfer: true,
					YandexQuery:  true,
				},
				PerformanceDiagnostics: &mysql.PerformanceDiagnostics{
					Enabled:                    true,
					SessionsSamplingInterval:   60,
					StatementsSamplingInterval: 600,
				},
				DiskSizeAutoscaling: &mysql.DiskSizeAutoscaling{
					DiskSizeLimit:           datasize.ToBytes(20),
					PlannedUsageThreshold:   30,
					EmergencyUsageThreshold: 60,
				},
				BackupWindowStart: &timeofday.TimeOfDay{
					Hours:   10,
					Minutes: 0,
				},
				BackupRetainPeriodDays: wrapperspb.Int64(7),
			},
			expectedVal: Config{
				Version: types.StringValue("5.7"),
				Resources: types.ObjectValueMust(expectedResourcesAttrs, map[string]attr.Value{
					"resource_preset_id": types.StringValue("s1.micro"),
					"disk_type_id":       types.StringValue("network-ssd"),
					"disk_size":          types.Int64Value(10),
				}),
				Access: types.ObjectValueMust(expectedAccessAttrTypes, map[string]attr.Value{
					"data_lens":     types.BoolValue(true),
					"data_transfer": types.BoolValue(true),
					"web_sql":       types.BoolValue(false),
					"yandex_query":  types.BoolValue(true),
				}),
				PerformanceDiagnostics: types.ObjectValueMust(expectedPDAttrs, map[string]attr.Value{
					"enabled":                      types.BoolValue(true),
					"sessions_sampling_interval":   types.Int64Value(60),
					"statements_sampling_interval": types.Int64Value(600),
				}),
				DiskSizeAutoscaling: types.ObjectValueMust(expectedDSAAttrs, map[string]attr.Value{
					"disk_size_limit":           types.Int64Value(20),
					"emergency_usage_threshold": types.Int64Value(60),
					"planned_usage_threshold":   types.Int64Value(30),
				}),
				BackupWindowStart: types.ObjectValueMust(expectedBwsAttrTypes, map[string]attr.Value{
					"hours":   types.Int64Value(10),
					"minutes": types.Int64Value(0),
				}),
				BackupRetainPeriodDays: types.Int64Value(7),
				MySQLConfig:            NewMsSettingsMapValueMust(map[string]attr.Value{}),
			},
		},
		{
			testname: "CheckPartlyAttributes",
			reqVal: &mysql.ClusterConfig{
				Version: "8.0",
				Resources: &mysql.Resources{
					ResourcePresetId: "s2.nano",
					DiskTypeId:       "network-hdd",
					DiskSize:         datasize.ToBytes(15),
				},
			},
			expectedVal: Config{
				Version: types.StringValue("8.0"),
				Resources: types.ObjectValueMust(expectedResourcesAttrs, map[string]attr.Value{
					"resource_preset_id": types.StringValue("s2.nano"),
					"disk_type_id":       types.StringValue("network-hdd"),
					"disk_size":          types.Int64Value(15),
				}),

				Access:                 types.ObjectNull(expectedAccessAttrTypes),
				PerformanceDiagnostics: types.ObjectNull(expectedPDAttrs),
				DiskSizeAutoscaling:    types.ObjectNull(expectedDSAAttrs),
				BackupWindowStart:      types.ObjectNull(expectedBwsAttrTypes),
				BackupRetainPeriodDays: types.Int64Null(),
				MySQLConfig:            NewMsSettingsMapValueMust(map[string]attr.Value{}),
			},
		},
		{
			testname:      "CheckNull",
			reqVal:        nil,
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		conf := flattenConfig(ctx, NewMsSettingsMapNull(), c.reqVal, &diags)
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

		if diags.HasError() {
			continue
		}

		confVal := reflect.ValueOf(conf)
		expectedConfVal := reflect.ValueOf(c.expectedVal)

		for i := 0; i < confVal.NumField(); i++ {
			confField := confVal.Field(i)
			expectedConfField := expectedConfVal.Field(i)
			if !confField.Interface().(attr.Value).Equal(expectedConfField.Interface().(attr.Value)) {
				t.Errorf(
					"Unexpected flatten result value %s test: expected %s, actual %s",
					c.testname,
					expectedConfField,
					confField,
				)
			}
		}
	}
}
