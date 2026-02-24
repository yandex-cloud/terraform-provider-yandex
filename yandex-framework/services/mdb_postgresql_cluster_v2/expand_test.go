package mdb_postgresql_cluster_v2

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

var mwAttrsTestExpand = map[string]attr.Type{
	"type": types.StringType,
	"day":  types.StringType,
	"hour": types.Int64Type,
}

func buildMWTestBlockObj(mwType, mwDay *string, mwHour *int64) types.Object {
	testBlock, _ := types.ObjectValue(mwAttrsTestExpand, map[string]attr.Value{
		"type": types.StringPointerValue(mwType),
		"day":  types.StringPointerValue(mwDay),
		"hour": types.Int64PointerValue(mwHour),
	})

	return testBlock
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

func TestYandexProvider_MDBPostgresClusterConfigPerfomanceDiagnosticsExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	rInt64 := rand.Int63n(86401)
	rBool := rand.Intn(2) == 1

	cases := []struct {
		testname  string
		testBlock types.Object
		expected  *postgresql.PerformanceDiagnostics
		hasErr    bool
	}{

		{
			testname:  "CheckNullBlock",
			testBlock: types.ObjectNull(mwAttrsTestExpand),
			expected:  nil,
		},
		{
			testname:  "CheckFullBlock",
			testBlock: buildPDTestBlockObj(&rBool, &rInt64, &rInt64),
			expected: &postgresql.PerformanceDiagnostics{
				Enabled:                    rBool,
				SessionsSamplingInterval:   rInt64,
				StatementsSamplingInterval: rInt64,
			},
		},
		{
			testname:  "CheckPartialBlock",
			testBlock: buildPDTestBlockObj(nil, &rInt64, &rInt64),
			expected: &postgresql.PerformanceDiagnostics{
				Enabled:                    false,
				SessionsSamplingInterval:   rInt64,
				StatementsSamplingInterval: rInt64,
			},
		},
		{
			testname:  "CheckEmptyBlock",
			testBlock: buildPDTestBlockObj(nil, nil, nil),
			expected: &postgresql.PerformanceDiagnostics{
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

func TestYandexProvider_MDBPostgresClusterConfigBackupRetainPeriodDaysExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      types.Int64
		expectedVal *wrapperspb.Int64Value
	}{
		{
			testname: "ExplicitCheck",
			reqVal:   types.Int64Value(5),
			expectedVal: &wrapperspb.Int64Value{
				Value: 5,
			},
		},
		{
			testname:    "NullCheck",
			reqVal:      types.Int64Null(),
			expectedVal: nil,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		pgBrpd := expandBackupRetainPeriodDays(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected expansion diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !reflect.DeepEqual(pgBrpd, c.expectedVal) {
			t.Errorf(
				"Unexpected expansion result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				pgBrpd,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterBoolWrapperExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      types.Bool
		expectedVal *wrapperspb.BoolValue
	}{
		{
			testname:    "CheckValidAttribute",
			reqVal:      types.BoolValue(true),
			expectedVal: wrapperspb.Bool(true),
		},
		{
			testname:    "CheckNullAttribute",
			reqVal:      types.BoolNull(),
			expectedVal: nil,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		b := expandBoolWrapper(ctx, c.reqVal, &diags)

		if !reflect.DeepEqual(b, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				b,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterConfigPgConfigExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pgSettingsType := NewPgSettingsMapType()
	correctPGMap, d := pgSettingsType.ValueFromMap(
		ctx, types.MapValueMust(
			types.StringType,
			map[string]attr.Value{
				"max_connections":                types.StringValue("395"),
				"enable_parallel_hash":           types.StringValue("true"),
				"autovacuum_vacuum_scale_factor": types.StringValue("0.34"),
				"default_transaction_isolation":  types.StringValue("TRANSACTION_ISOLATION_READ_COMMITTED"),
				"shared_preload_libraries":       types.StringValue("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"),
				"auto_explain_log_format":        types.StringValue("AUTO_EXPLAIN_LOG_FORMAT_XML"),
			},
		),
	)

	if d.HasError() {
		t.Errorf("Unexpected error: %s", d.Errors())
	}

	randomPGMap, diags := pgSettingsType.ValueFromMap(
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

	cases := []struct {
		testname      string
		version       string
		reqVal        mdbcommon.SettingsMapValue
		expectedVal   postgresql.ConfigSpec_PostgresqlConfig
		expectedError bool
	}{
		{
			testname: "CheckPartlyAttributes",
			version:  "15",
			reqVal:   correctPGMap.(mdbcommon.SettingsMapValue),
			expectedVal: &postgresql.ConfigSpec_PostgresqlConfig_15{
				PostgresqlConfig_15: &config.PostgresqlConfig15{
					MaxConnections:              wrapperspb.Int64(395),
					EnableParallelHash:          wrapperspb.Bool(true),
					AutovacuumVacuumScaleFactor: wrapperspb.Double(0.34),
					DefaultTransactionIsolation: config.PostgresqlConfig15_TRANSACTION_ISOLATION_READ_COMMITTED,
					SharedPreloadLibraries: []config.PostgresqlConfig15_SharedPreloadLibraries{
						config.PostgresqlConfig15_SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN, config.PostgresqlConfig15_SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN,
					},
					AutoExplainLogFormat: config.PostgresqlConfig15_AUTO_EXPLAIN_LOG_FORMAT_XML,
				},
			},
		},
		{
			testname:      "CheckAccessWithRandomAttributes",
			reqVal:        randomPGMap.(mdbcommon.SettingsMapValue),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		conf := expandPostgresqlConfig(ctx, c.version, c.reqVal, &diags)
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
				"Unexpected expand result value %s test:\n expected %s\n actual %s",
				c.testname,
				c.expectedVal,
				conf,
			)
		}
	}
}

func buildTestPCObj(pd *bool, pm *string) types.Object {
	return types.ObjectValueMust(
		expectedPCAttrTypes, map[string]attr.Value{
			"pool_discard": types.BoolPointerValue(pd),
			"pooling_mode": types.StringPointerValue(pm),
		},
	)
}

func TestYandexProvider_MDBPostgresClusterConfigPoolerConfigExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testBoolTrue := true
	testBoolFalse := false
	validPMs := []string{"SESSION", "TRANSACTION", "STATEMENT"}

	invalidPM := "INVALID1"

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *postgresql.ConnectionPoolerConfig
		expectedError bool
	}{
		{
			testname:    "CheckInvalidPoolingMode",
			reqVal:      buildTestPCObj(nil, &invalidPM),
			expectedVal: &postgresql.ConnectionPoolerConfig{},
		},
		{
			testname: "CheckPartlyAttributesWithPD",
			reqVal:   buildTestPCObj(&testBoolFalse, nil),
			expectedVal: &postgresql.ConnectionPoolerConfig{
				PoolDiscard: wrapperspb.Bool(false),
				PoolingMode: postgresql.ConnectionPoolerConfig_PoolingMode(0),
			},
		},
		{
			testname: "CheckPartlyAttributesWithPM",
			reqVal:   buildTestPCObj(nil, &validPMs[0]),
			expectedVal: &postgresql.ConnectionPoolerConfig{
				PoolDiscard: nil,
				PoolingMode: postgresql.ConnectionPoolerConfig_PoolingMode(1),
			},
		},
		{
			testname: "CheckWithoutAttributes",
			reqVal:   buildTestPCObj(nil, nil),
			expectedVal: &postgresql.ConnectionPoolerConfig{
				PoolDiscard: nil,
				PoolingMode: postgresql.ConnectionPoolerConfig_PoolingMode(0),
			},
		},
		{
			testname:    "CheckNullObj",
			reqVal:      types.ObjectNull(expectedPCAttrTypes),
			expectedVal: &postgresql.ConnectionPoolerConfig{},
		},
		{
			testname: "CheckWithRandomAttributes",
			reqVal: types.ObjectValueMust(
				map[string]attr.Type{"random": types.StringType},
				map[string]attr.Value{"random": types.StringValue("s1")},
			),
			expectedError: true,
		},
	}

	for idx, pm := range validPMs {
		pm := pm
		cases = append(cases, struct {
			testname      string
			reqVal        types.Object
			expectedVal   *postgresql.ConnectionPoolerConfig
			expectedError bool
		}{
			testname: fmt.Sprintf("CheckAllExplicitWithValidPoolingMode%s", pm),
			reqVal:   buildTestPCObj(&testBoolTrue, &pm),
			expectedVal: &postgresql.ConnectionPoolerConfig{
				PoolDiscard: wrapperspb.Bool(true),
				PoolingMode: postgresql.ConnectionPoolerConfig_PoolingMode(idx + 1),
			},
		})
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		pc := expandPoolerConfig(ctx, c.reqVal, &diags)
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

		if !reflect.DeepEqual(pc, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				pc,
			)
		}
	}
}

func buildTestDiskSizeAutoscalingObject(diskSizeLimit, plannedUsageThreshold, emergencyUsageThreshold *int64) types.Object {
	return types.ObjectValueMust(
		expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
			"disk_size_limit":           types.Int64PointerValue(diskSizeLimit),
			"planned_usage_threshold":   types.Int64PointerValue(plannedUsageThreshold),
			"emergency_usage_threshold": types.Int64PointerValue(emergencyUsageThreshold),
		},
	)
}

func TestYandexProvider_MDBPostgresClusterConfigDiskSizeAutoscalingExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var gB, pThreshold, emThreshold int64 = 3, 50, 30

	cases := []struct {
		testname      string
		dsaObj        types.Object
		expectedVal   *postgresql.DiskSizeAutoscaling
		expectedError bool
	}{
		{
			testname: "CheckAllExplicitAttributes",
			dsaObj:   buildTestDiskSizeAutoscalingObject(&gB, &pThreshold, &emThreshold),
			expectedVal: &postgresql.DiskSizeAutoscaling{
				DiskSizeLimit:           datasize.ToBytes(gB),
				PlannedUsageThreshold:   pThreshold,
				EmergencyUsageThreshold: emThreshold,
			},
			expectedError: false,
		},
		{
			testname: "CheckExplicitAttributesWithPlannedUsageThreshold",
			dsaObj:   buildTestDiskSizeAutoscalingObject(&gB, &pThreshold, nil),
			expectedVal: &postgresql.DiskSizeAutoscaling{
				DiskSizeLimit:         datasize.ToBytes(gB),
				PlannedUsageThreshold: pThreshold,
			},
			expectedError: false,
		},
		{
			testname: "CheckExplicitAttributesWithEmergencyUsageThreshold",
			dsaObj:   buildTestDiskSizeAutoscalingObject(&gB, nil, &emThreshold),
			expectedVal: &postgresql.DiskSizeAutoscaling{
				DiskSizeLimit:           datasize.ToBytes(gB),
				EmergencyUsageThreshold: emThreshold,
			},
			expectedError: false,
		},
		{
			testname: "CheckWithoutOptionalAttributes",
			dsaObj:   buildTestDiskSizeAutoscalingObject(&gB, nil, nil),
			expectedVal: &postgresql.DiskSizeAutoscaling{
				DiskSizeLimit: datasize.ToBytes(gB),
			},
			expectedError: false,
		},
		{
			testname:      "CheckNullObject",
			dsaObj:        types.ObjectNull(expectedDiskSizeAutoscalingAttrs),
			expectedVal:   nil,
			expectedError: false,
		},
		{
			testname: "CheckAccessWithRandomAttributes",
			dsaObj: types.ObjectValueMust(
				map[string]attr.Type{"random": types.StringType},
				map[string]attr.Value{"random": types.StringValue("s1")},
			),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		pgDSA := expandDiskSizeAutoscaling(ctx, c.dsaObj, &diags)
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

		if !reflect.DeepEqual(pgDSA, c.expectedVal) {
			t.Errorf(
				"Unexpected expansion result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				pgDSA,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterConfigExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *postgresql.ConfigSpec
		expectedError bool
	}{
		{
			testname: "CheckPartlyAttributes",
			reqVal: types.ObjectValueMust(
				expectedConfigAttrs,
				map[string]attr.Value{
					"version": types.StringValue("15"),
					"resources": types.ObjectValueMust(
						mdbcommon.ResourceType.AttrTypes,
						map[string]attr.Value{
							"resource_preset_id": types.StringValue("s1.micro"),
							"disk_type_id":       types.StringValue("network-ssd"),
							"disk_size":          types.Int64Value(10),
						},
					),
					"backup_window_start":       types.ObjectNull(mdbcommon.BackupWindowType.AttrTypes),
					"backup_retain_period_days": types.Int64Null(),
					"access":                    types.ObjectNull(accessAttrTypes),
					"performance_diagnostics":   types.ObjectNull(expectedPDAttrs),
					"postgresql_config":         NewPgSettingsMapNull(),
					"pooler_config":             types.ObjectNull(expectedPCAttrTypes),
					"disk_size_autoscaling":     types.ObjectNull(expectedDiskSizeAutoscalingAttrs),
				},
			),
			expectedVal: &postgresql.ConfigSpec{
				Version: "15",
				Resources: &postgresql.Resources{
					ResourcePresetId: "s1.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         datasize.ToBytes(10),
				},
				BackupWindowStart:      &timeofday.TimeOfDay{},
				BackupRetainPeriodDays: nil,
				Access:                 &postgresql.Access{},
				PerformanceDiagnostics: nil,
				PostgresqlConfig: &postgresql.ConfigSpec_PostgresqlConfig_15{
					PostgresqlConfig_15: &config.PostgresqlConfig15{},
				},
				PoolerConfig: &postgresql.ConnectionPoolerConfig{},
			},
		},
		{
			testname: "CheckFullAttributes",
			reqVal: types.ObjectValueMust(
				expectedConfigAttrs,
				map[string]attr.Value{
					"version": types.StringValue("15"),
					"resources": types.ObjectValueMust(
						mdbcommon.ResourceType.AttrTypes,
						map[string]attr.Value{
							"resource_preset_id": types.StringValue("s1.micro"),
							"disk_type_id":       types.StringValue("network-ssd"),
							"disk_size":          types.Int64Value(10),
						},
					),
					"backup_window_start": types.ObjectValueMust(
						mdbcommon.BackupWindowType.AttrTypes,
						map[string]attr.Value{
							"hours":   types.Int64Value(23),
							"minutes": types.Int64Value(0),
						},
					),
					"backup_retain_period_days": types.Int64Value(7),
					"access": types.ObjectValueMust(
						accessAttrTypes,
						map[string]attr.Value{
							"web_sql":       types.BoolValue(true),
							"serverless":    types.BoolValue(false),
							"data_transfer": types.BoolValue(false),
							"data_lens":     types.BoolValue(true),
							"yandex_query":  types.BoolValue(true),
						},
					),
					"performance_diagnostics": types.ObjectValueMust(
						expectedPDAttrs,
						map[string]attr.Value{
							"enabled":                      types.BoolValue(true),
							"statements_sampling_interval": types.Int64Value(600),
							"sessions_sampling_interval":   types.Int64Value(60),
						},
					),
					"postgresql_config": NewPgSettingsMapValueMust(
						map[string]attr.Value{
							"max_connections": types.Int64Value(100),
						},
					),
					"pooler_config": types.ObjectValueMust(expectedPCAttrTypes, map[string]attr.Value{
						"pool_discard": types.BoolValue(false),
						"pooling_mode": types.StringValue(postgresql.ConnectionPoolerConfig_STATEMENT.String()),
					}),
					"disk_size_autoscaling": types.ObjectValueMust(expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
						"disk_size_limit":           types.Int64Value(5),
						"emergency_usage_threshold": types.Int64Value(20),
						"planned_usage_threshold":   types.Int64Value(30),
					}),
				},
			),
			expectedVal: &postgresql.ConfigSpec{
				Version: "15",
				Resources: &postgresql.Resources{
					ResourcePresetId: "s1.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         datasize.ToBytes(10),
				},
				BackupWindowStart: &timeofday.TimeOfDay{
					Hours:   23,
					Minutes: 0,
				},
				BackupRetainPeriodDays: wrapperspb.Int64(7),
				Access: &postgresql.Access{
					WebSql:      true,
					DataLens:    true,
					YandexQuery: true,
				},
				PerformanceDiagnostics: &postgresql.PerformanceDiagnostics{
					Enabled:                    true,
					StatementsSamplingInterval: 600,
					SessionsSamplingInterval:   60,
				},
				PostgresqlConfig: &postgresql.ConfigSpec_PostgresqlConfig_15{
					PostgresqlConfig_15: &config.PostgresqlConfig15{
						MaxConnections: wrapperspb.Int64(100),
					},
				},
				PoolerConfig: &postgresql.ConnectionPoolerConfig{
					PoolingMode: postgresql.ConnectionPoolerConfig_STATEMENT,
					PoolDiscard: wrapperspb.Bool(false),
				},
				DiskSizeAutoscaling: &postgresql.DiskSizeAutoscaling{
					DiskSizeLimit:           datasize.ToBytes(5),
					EmergencyUsageThreshold: 20,
					PlannedUsageThreshold:   30,
				},
			},
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

		if !proto.Equal(conf, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test:\n expected %s\n actual %s",
				c.testname,
				c.expectedVal,
				conf,
			)
		}
	}
}
