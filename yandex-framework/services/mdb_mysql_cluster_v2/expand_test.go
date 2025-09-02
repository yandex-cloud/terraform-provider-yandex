package mdb_mysql_cluster_v2

import (
	"context"
	"math/rand"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

var expectedAccessAttrTypes = map[string]attr.Type{
	"data_lens":     types.BoolType,
	"web_sql":       types.BoolType,
	"data_transfer": types.BoolType,
}

func buildTestAccessObj(dataLens, dataTransfer, webSql *bool) types.Object {
	return types.ObjectValueMust(
		expectedAccessAttrTypes, map[string]attr.Value{
			"data_transfer": types.BoolPointerValue(dataTransfer),
			"data_lens":     types.BoolPointerValue(dataLens),
			"web_sql":       types.BoolPointerValue(webSql),
		},
	)
}

func TestYandexProvider_MDBMySQLClusterConfigAccessExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	trueAttr := true
	falseAttr := false

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *mysql.Access
		expectedError bool
	}{
		{
			testname: "CheckAllExplicitAttributes",
			reqVal:   buildTestAccessObj(&trueAttr, &trueAttr, &falseAttr),
			expectedVal: &mysql.Access{
				DataLens:     trueAttr,
				DataTransfer: trueAttr,
			},
			expectedError: false,
		},
		{
			testname: "CheckPartlyAttributes",
			reqVal:   buildTestAccessObj(&trueAttr, &falseAttr, nil),
			expectedVal: &mysql.Access{
				DataLens:     trueAttr,
				DataTransfer: falseAttr,
			},
			expectedError: false,
		},
		{
			testname:      "CheckWithoutAttributes",
			reqVal:        buildTestAccessObj(nil, nil, nil),
			expectedVal:   &mysql.Access{},
			expectedError: false,
		},
		{
			testname:      "CheckNullAccess",
			reqVal:        types.ObjectNull(expectedAccessAttrTypes),
			expectedVal:   &mysql.Access{},
			expectedError: false,
		},
		{
			testname: "CheckAccessWithRandomAttributes",
			reqVal: types.ObjectValueMust(
				map[string]attr.Type{"random": types.StringType},
				map[string]attr.Value{"random": types.StringValue("s1")},
			),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		pgAccess := expandAccess(ctx, c.reqVal, &diags)
		if diags.HasError() != c.expectedError {
			t.Errorf(
				"Unexpected expansion diagnostics status %s test: expected %t, actual %t with errors: %v",
				c.testname,
				c.expectedError,
				diags.HasError(),
				diags.Errors(),
			)
			continue
		}

		if !reflect.DeepEqual(pgAccess, c.expectedVal) {
			t.Errorf(
				"Unexpected expansion result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				pgAccess,
			)
		}
	}
}

var pdTestExpand = map[string]attr.Type{
	"enabled":                      types.BoolType,
	"sessions_sampling_interval":   types.Int64Type,
	"statements_sampling_interval": types.Int64Type,
}

func buildPDTestBlockObj(enabled *bool, sessionsSi, statementsSi *int64) types.Object {
	return types.ObjectValueMust(pdTestExpand, map[string]attr.Value{
		"enabled":                      types.BoolPointerValue(enabled),
		"sessions_sampling_interval":   types.Int64PointerValue(sessionsSi),
		"statements_sampling_interval": types.Int64PointerValue(statementsSi),
	})
}

func TestYandexProvider_MDBMySQLClusterConfigPerfomanceDiagnosticsExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	rInt64 := rand.Int63n(86401)
	rBool := rand.Intn(2) == 1

	cases := []struct {
		testname  string
		testBlock types.Object
		expected  *mysql.PerformanceDiagnostics
		hasErr    bool
	}{
		{
			testname:  "CheckNullBlock",
			testBlock: types.ObjectNull(pdTestExpand),
			expected:  nil,
		},
		{
			testname:  "CheckFullBlock",
			testBlock: buildPDTestBlockObj(&rBool, &rInt64, &rInt64),
			expected: &mysql.PerformanceDiagnostics{
				Enabled:                    rBool,
				SessionsSamplingInterval:   rInt64,
				StatementsSamplingInterval: rInt64,
			},
		},
		{
			testname:  "CheckPartialBlock",
			testBlock: buildPDTestBlockObj(nil, &rInt64, &rInt64),
			expected: &mysql.PerformanceDiagnostics{
				Enabled:                    false,
				SessionsSamplingInterval:   rInt64,
				StatementsSamplingInterval: rInt64,
			},
		},
		{
			testname:  "CheckEmptyBlock",
			testBlock: buildPDTestBlockObj(nil, nil, nil),
			expected: &mysql.PerformanceDiagnostics{
				Enabled:                    false,
				SessionsSamplingInterval:   0,
				StatementsSamplingInterval: 0,
			},
		},
		{
			testname: "CheckWithRandomAttributes",
			testBlock: types.ObjectValueMust(map[string]attr.Type{
				"attr1": types.Int64Type,
			}, map[string]attr.Value{
				"attr1": types.Int64Value(10),
			}),
			hasErr: true,
		},
	}

	for _, c := range cases {
		var diags diag.Diagnostics
		res := expandPerformanceDiagnostics(ctx, c.testBlock, &diags)
		if c.hasErr {
			if !diags.HasError() {
				t.Errorf("Unexpected expand error status: expected %v, actual %v", c.hasErr, diags.HasError())
			}
			continue
		}

		if !reflect.DeepEqual(res, c.expected) {
			t.Errorf("Unexpected expancion result policy: expected %v, actual %v", c.expected, res)
		}
	}
}

var expectedBwsAttrTypes = map[string]attr.Type{
	"hours":   types.Int64Type,
	"minutes": types.Int64Type,
}

func TestYandexProvider_MDBMySQLClusterConfigExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        Config
		expectedVal   *mysql.ConfigSpec
		expectedError bool
	}{
		{
			testname: "CheckPartlyAttributes",
			reqVal: Config{
				Version: types.StringValue("5.7"),
				Resources: types.ObjectValueMust(
					expectedResourcesAttrs,
					map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					},
				),
				BackupWindowStart:      types.ObjectNull(expectedBwsAttrTypes),
				BackupRetainPeriodDays: types.Int64Null(),
				Access:                 types.ObjectNull(expectedAccessAttrTypes),
				PerformanceDiagnostics: types.ObjectNull(expectedPDAttrs),
				DiskSizeAutoscaling:    types.ObjectNull(expectedDSAAttrs),
				MySQLConfig:            NewMsSettingsMapNull(),
			},
			expectedVal: &mysql.ConfigSpec{
				Version: "5.7",
				Resources: &mysql.Resources{
					ResourcePresetId: "s1.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         datasize.ToBytes(10),
				},
				BackupWindowStart: &timeofday.TimeOfDay{},
				Access:            &mysql.Access{},
				MysqlConfig: &mysql.ConfigSpec_MysqlConfig_5_7{
					MysqlConfig_5_7: &config.MysqlConfig5_7{},
				},
			},
		},
		{
			testname: "CheckFullAttributes",
			reqVal: Config{
				Version: types.StringValue("8.0"),
				Resources: types.ObjectValueMust(
					expectedResourcesAttrs,
					map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					},
				),
				BackupWindowStart: types.ObjectValueMust(
					expectedBwsAttrTypes,
					map[string]attr.Value{
						"hours":   types.Int64Value(23),
						"minutes": types.Int64Value(0),
					},
				),
				BackupRetainPeriodDays: types.Int64Value(7),
				Access: types.ObjectValueMust(
					expectedAccessAttrTypes,
					map[string]attr.Value{
						"web_sql":       types.BoolValue(true),
						"data_transfer": types.BoolValue(false),
						"data_lens":     types.BoolValue(true),
					},
				),
				PerformanceDiagnostics: types.ObjectValueMust(
					expectedPDAttrs,
					map[string]attr.Value{
						"enabled":                      types.BoolValue(true),
						"statements_sampling_interval": types.Int64Value(600),
						"sessions_sampling_interval":   types.Int64Value(60),
					},
				),
				DiskSizeAutoscaling: types.ObjectValueMust(
					expectedDSAAttrs,
					map[string]attr.Value{
						"disk_size_limit":           types.Int64Value(20),
						"planned_usage_threshold":   types.Int64Value(30),
						"emergency_usage_threshold": types.Int64Value(60),
					},
				),
				MySQLConfig: NewMsSettingsMapValueMust(map[string]attr.Value{
					"max_connections": types.Int64Value(100),
				}),
			},
			expectedVal: &mysql.ConfigSpec{
				Version: "8.0",
				Resources: &mysql.Resources{
					ResourcePresetId: "s1.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         datasize.ToBytes(10),
				},
				BackupWindowStart: &timeofday.TimeOfDay{
					Hours:   23,
					Minutes: 0,
				},
				BackupRetainPeriodDays: wrapperspb.Int64(7),

				Access: &mysql.Access{
					WebSql:   true,
					DataLens: true,
				},
				PerformanceDiagnostics: &mysql.PerformanceDiagnostics{
					Enabled:                    true,
					StatementsSamplingInterval: 600,
					SessionsSamplingInterval:   60,
				},
				DiskSizeAutoscaling: &mysql.DiskSizeAutoscaling{
					DiskSizeLimit:           datasize.ToBytes(20),
					PlannedUsageThreshold:   30,
					EmergencyUsageThreshold: 60,
				},
				MysqlConfig: &mysql.ConfigSpec_MysqlConfig_8_0{
					MysqlConfig_8_0: &config.MysqlConfig8_0{
						MaxConnections: wrapperspb.Int64(100),
					},
				},
			},
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		conf := expandConfig(ctx, c.reqVal, &diags)
		if diags.HasError() != c.expectedError {
			t.Errorf(
				"Unexpected expand diagnostics status %s test: expected %t, actual %t with errors: %v",
				c.testname,
				c.expectedError,
				diags.HasError(),
				diags.Errors(),
			)
			continue
		}

		if !reflect.DeepEqual(conf, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test:\n expected %v\n actual %v",
				c.testname,
				c.expectedVal,
				conf,
			)
		}
	}
}

func TestYandexProvider_MDBMySQLClusterConfigMsConfigExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	msSettingsType := NewMsSettingsMapType()

	partlyMSMap, diags := msSettingsType.ValueFromMap(
		ctx, types.MapValueMust(
			types.StringType,
			map[string]attr.Value{
				"sql_mode":                      types.StringValue("ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE"),
				"max_connections":               types.StringValue("100"),
				"default_authentication_plugin": types.StringValue("MYSQL_NATIVE_PASSWORD"),
				"innodb_print_all_deadlocks":    types.StringValue("true"),
			},
		),
	)

	if diags.HasError() {
		t.Fatal(diags)
	}

	randomMSMap, diags := msSettingsType.ValueFromMap(
		ctx, types.MapValueMust(
			types.Int64Type,
			map[string]attr.Value{
				"random": types.Int64Value(11),
			},
		),
	)

	if diags.HasError() {
		t.Fatal(diags)
	}

	incorrectTupleMSMap, diags := msSettingsType.ValueFromMap(
		ctx, types.MapValueMust(
			types.StringType,
			map[string]attr.Value{
				"sql_mode": types.StringValue("ONLY_FULL_GROUP_BY,,,,"),
			},
		),
	)

	if diags.HasError() {
		t.Fatal(diags)
	}

	cases := []struct {
		testname      string
		version       string
		reqVal        mdbcommon.SettingsMapValue
		expectedVal   mysql.ConfigSpec_MysqlConfig
		expectedError bool
	}{
		{
			testname: "CheckPartlyAttributes",
			version:  "5.7",
			reqVal:   partlyMSMap.(mdbcommon.SettingsMapValue),
			expectedVal: &mysql.ConfigSpec_MysqlConfig_5_7{
				MysqlConfig_5_7: &config.MysqlConfig5_7{
					MaxConnections: wrapperspb.Int64(100),
					SqlMode: []config.MysqlConfig5_7_SQLMode{
						config.MysqlConfig5_7_ONLY_FULL_GROUP_BY,
						config.MysqlConfig5_7_STRICT_TRANS_TABLES,
						config.MysqlConfig5_7_NO_ZERO_IN_DATE,
					},
					DefaultAuthenticationPlugin: config.MysqlConfig5_7_MYSQL_NATIVE_PASSWORD,
					InnodbPrintAllDeadlocks:     wrapperspb.Bool(true),
				},
			},
		},
		{
			testname:      "CheckIncorrectTuple",
			version:       "5.7",
			reqVal:        incorrectTupleMSMap.(mdbcommon.SettingsMapValue),
			expectedError: true,
		},
		{
			testname:      "CheckAccessWithRandomAttributes",
			reqVal:        randomMSMap.(mdbcommon.SettingsMapValue),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		conf := expandMySQLConfig(ctx, c.version, c.reqVal, &diags)
		if diags.HasError() != c.expectedError {
			t.Errorf(
				"Unexpected expand diagnostics status %s test: expected %t, actual %t with errors: %v",
				c.testname,
				c.expectedError,
				diags.HasError(),
				diags.Errors(),
			)
			continue
		}

		if diags.HasError() {
			continue
		}

		if !reflect.DeepEqual(conf, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test:\n expected %s\n actual %s",
				c.testname,
				c.expectedVal,
				conf,
			)
		}
	}
}
