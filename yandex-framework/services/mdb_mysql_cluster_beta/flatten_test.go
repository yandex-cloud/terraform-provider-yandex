package mdb_mysql_cluster_beta

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
	}

	cases := []struct {
		testname    string
		reqVal      *mysql.Access
		expectedVal types.Object
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &mysql.Access{
				WebSql:   true,
				DataLens: true,
			},
			expectedVal: types.ObjectValueMust(
				expectedAccessAttrs, map[string]attr.Value{
					"data_lens":     types.BoolValue(true),
					"data_transfer": types.BoolValue(false),
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

func TestYandexProvider_MDBMySQLClusterMaintenanceWindowFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *mysql.MaintenanceWindow
		expectedVal types.Object
		hasErr      bool
	}{
		{
			testname: "CheckWeeklyMaintenanceWindow",
			reqVal: &mysql.MaintenanceWindow{
				Policy: &mysql.MaintenanceWindow_WeeklyMaintenanceWindow{
					WeeklyMaintenanceWindow: &mysql.WeeklyMaintenanceWindow{
						Hour: 10,
						Day:  mysql.WeeklyMaintenanceWindow_WeekDay(1),
					},
				},
			},
			expectedVal: types.ObjectValueMust(expectedMWAttrs, map[string]attr.Value{
				"type": types.StringValue("WEEKLY"),
				"day":  types.StringValue("MON"),
				"hour": types.Int64Value(10),
			}),
		},
		{
			testname: "CheckAnytimeMaintenanceWindow",
			reqVal: &mysql.MaintenanceWindow{
				Policy: &mysql.MaintenanceWindow_Anytime{
					Anytime: &mysql.AnytimeMaintenanceWindow{},
				},
			},
			expectedVal: types.ObjectValueMust(expectedMWAttrs, map[string]attr.Value{
				"type": types.StringValue("ANYTIME"),
				"day":  types.StringNull(),
				"hour": types.Int64Null(),
			}),
		},
		{
			testname:    "CheckNullMaintenanceWindow",
			reqVal:      nil,
			expectedVal: types.ObjectNull(expectedMWAttrs),
			hasErr:      true,
		},
		{
			testname:    "CheckEmptyMaintenanceWindow",
			reqVal:      &mysql.MaintenanceWindow{},
			expectedVal: types.ObjectNull(expectedMWAttrs),
			hasErr:      true,
		},
		{
			testname: "CheckPolicyNilMaintenanceWindow",
			reqVal: &mysql.MaintenanceWindow{
				Policy: nil,
			},
			expectedVal: types.ObjectNull(expectedMWAttrs),
			hasErr:      true,
		},
	}

	for _, c := range cases {
		var diags diag.Diagnostics
		res := flattenMaintenanceWindow(ctx, c.reqVal, &diags)
		if c.hasErr {
			if !diags.HasError() {
				t.Errorf("Unexpected flatten error status: expected %v, actual %v", c.hasErr, diags.HasError())
			}
			continue
		}

		if !c.expectedVal.Equal(res) {
			t.Errorf("Unexpected flatten object result: expected %v, actual %v", c.expectedVal, res)
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

func TestYandexProvider_MDBMySQLClusterConfigBackupRetainPeriodDaysFlattener(t *testing.T) {
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

func TestYandexProvider_MDBMySQLClusterConfigBackupWindowStartFlattener(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *timeofday.TimeOfDay
		expectedVal types.Object
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &timeofday.TimeOfDay{
				Hours:   30,
				Minutes: 30,
			},
			expectedVal: types.ObjectValueMust(
				expectedBWSAttrs, map[string]attr.Value{
					"hours":   types.Int64Value(30),
					"minutes": types.Int64Value(30),
				},
			),
		},
		{
			testname: "CheckAllAttributesWithDefaultValues",
			reqVal:   &timeofday.TimeOfDay{},
			expectedVal: types.ObjectValueMust(
				expectedBWSAttrs, map[string]attr.Value{
					"hours":   types.Int64Value(0),
					"minutes": types.Int64Value(0),
				},
			),
		},
		{
			testname: "CheckPartlyAttributesWithHours",
			reqVal: &timeofday.TimeOfDay{
				Hours: 30,
			},
			expectedVal: types.ObjectValueMust(
				expectedBWSAttrs, map[string]attr.Value{
					"hours":   types.Int64Value(30),
					"minutes": types.Int64Value(0),
				},
			),
		},
		{
			testname: "CheckPartlyAttributesWithMinutes",
			reqVal: &timeofday.TimeOfDay{
				Minutes: 30,
			},
			expectedVal: types.ObjectValueMust(
				expectedBWSAttrs, map[string]attr.Value{
					"hours":   types.Int64Value(0),
					"minutes": types.Int64Value(30),
				},
			),
		},
		{
			testname:    "CheckNullObject",
			reqVal:      nil,
			expectedVal: types.ObjectNull(expectedBWSAttrs),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		bws := flattenBackupWindowStart(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected flatten diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !c.expectedVal.Equal(bws) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				bws,
			)
		}
	}
}

func TestYandexProvider_MDBMySQLClusterMapStringFlatten(t *testing.T) {
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

func TestYandexProvider_MDBMySQLClusterSetStringFlatten(t *testing.T) {
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

func TestYandexProvider_MDBMySQLClusterBoolWrapperFlatten(t *testing.T) {
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

func TestYandexProvider_MDBMySQLClusterResourcesFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        *mysql.Resources
		expectedVal   types.Object
		expectedError bool
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &mysql.Resources{
				ResourcePresetId: "s1.micro",
				DiskTypeId:       "network-ssd",
				DiskSize:         datasize.ToBytes(10),
			},
			expectedVal: types.ObjectValueMust(
				expectedResourcesAttrs, map[string]attr.Value{
					"resource_preset_id": types.StringValue("s1.micro"),
					"disk_type_id":       types.StringValue("network-ssd"),
					"disk_size":          types.Int64Value(10),
				},
			),
		},
		{
			testname:      "CheckNullAttributes",
			reqVal:        nil,
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		r := flattenResources(ctx, c.reqVal, &diags)
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

		if !c.expectedVal.Equal(r) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				r,
			)
		}
	}
}

type invalidPgConfig struct {
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
			reqVal:        invalidPgConfig{},
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
				},
				PerformanceDiagnostics: &mysql.PerformanceDiagnostics{
					Enabled:                    true,
					SessionsSamplingInterval:   60,
					StatementsSamplingInterval: 600,
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
				}),
				PerformanceDiagnostics: types.ObjectValueMust(expectedPDAttrs, map[string]attr.Value{
					"enabled":                      types.BoolValue(true),
					"sessions_sampling_interval":   types.Int64Value(60),
					"statements_sampling_interval": types.Int64Value(600),
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
