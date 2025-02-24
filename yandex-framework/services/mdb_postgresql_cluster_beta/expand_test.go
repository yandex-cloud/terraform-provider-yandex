package mdb_postgresql_cluster_beta

import (
	"context"
	"math/rand"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

var expectedAccessAttrTypes = map[string]attr.Type{
	"data_lens":     types.BoolType,
	"web_sql":       types.BoolType,
	"serverless":    types.BoolType,
	"data_transfer": types.BoolType,
}

func buildTestAccessObj(dataLens, dataTransfer, webSql, serverless *bool) types.Object {
	return types.ObjectValueMust(
		expectedAccessAttrTypes, map[string]attr.Value{
			"data_transfer": types.BoolPointerValue(dataTransfer),
			"data_lens":     types.BoolPointerValue(dataLens),
			"serverless":    types.BoolPointerValue(serverless),
			"web_sql":       types.BoolPointerValue(webSql),
		},
	)
}

func TestYandexProvider_MDBPostgresClusterConfigAccessExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	trueAttr := true
	falseAttr := false

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *postgresql.Access
		expectedError bool
	}{
		{
			testname: "CheckAllExplicitAttributes",
			reqVal:   buildTestAccessObj(&trueAttr, &trueAttr, &falseAttr, &falseAttr),
			expectedVal: &postgresql.Access{
				DataLens:     trueAttr,
				DataTransfer: trueAttr,
			},
			expectedError: false,
		},
		{
			testname: "CheckPartlyAttributes",
			reqVal:   buildTestAccessObj(&trueAttr, &falseAttr, nil, nil),
			expectedVal: &postgresql.Access{
				DataLens:     trueAttr,
				DataTransfer: falseAttr,
			},
			expectedError: false,
		},
		{
			testname:      "CheckWithoutAttributes",
			reqVal:        buildTestAccessObj(nil, nil, nil, nil),
			expectedVal:   &postgresql.Access{},
			expectedError: false,
		},
		{
			testname:      "CheckNullAccess",
			reqVal:        types.ObjectNull(expectedAccessAttrTypes),
			expectedVal:   &postgresql.Access{},
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

func TestYandexProvider_MDBPostgresClusterMaintenanceWindowExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	anytimeType := "ANYTIME"
	weeklyType := "WEEKLY"

	day := "MON"
	var hour int64 = 1

	cases := []struct {
		testname       string
		reqVal         types.Object
		expectedPolicy postgresql.MaintenanceWindow_Policy
		expectedError  bool
	}{
		{
			testname:       "CheckNullObject",
			reqVal:         types.ObjectNull(mwAttrsTestExpand),
			expectedPolicy: nil,
		},
		{
			testname: "CheckAnytimeMaintenanceWindow",
			reqVal:   buildMWTestBlockObj(&anytimeType, nil, nil),
			expectedPolicy: &postgresql.MaintenanceWindow_Anytime{
				Anytime: &postgresql.AnytimeMaintenanceWindow{},
			},
		},
		{
			testname: "CheckWeeklyMaintenanceWindow",
			reqVal:   buildMWTestBlockObj(&weeklyType, &day, &hour),
			expectedPolicy: &postgresql.MaintenanceWindow_WeeklyMaintenanceWindow{
				WeeklyMaintenanceWindow: &postgresql.WeeklyMaintenanceWindow{
					Hour: hour,
					Day:  postgresql.WeeklyMaintenanceWindow_WeekDay(1),
				},
			},
		},
		{
			testname:      "CheckBlockWithRandomAttributes",
			reqVal:        types.ObjectValueMust(map[string]attr.Type{"random": types.StringType}, map[string]attr.Value{"random": types.StringValue("s1")}),
			expectedError: true,
		},
	}

	for _, c := range cases {
		var diags diag.Diagnostics
		res := expandClusterMaintenanceWindow(ctx, c.reqVal, &diags)
		if c.expectedError {
			if !diags.HasError() {
				t.Errorf("Unexpected expancion error status: expected %v, actual %v", c.expectedError, diags.HasError())
			}
			continue
		}

		if c.expectedPolicy != nil && !reflect.DeepEqual(res.Policy, c.expectedPolicy) {
			t.Errorf("Unexpected expancion result policy: expected %v, actual %v", c.expectedPolicy, res.Policy)
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

var expectedBwsAttrTypes = map[string]attr.Type{
	"hours":   types.Int64Type,
	"minutes": types.Int64Type,
}

func buildTestBwsObj(h, m *int64) types.Object {
	return types.ObjectValueMust(
		expectedBwsAttrTypes, map[string]attr.Value{
			"hours":   types.Int64PointerValue(h),
			"minutes": types.Int64PointerValue(m),
		},
	)
}

func TestYandexProvider_MDBPostgresClusterConfigBackupWindowStartExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testInt64 := int64(30)

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *timeofday.TimeOfDay
		expectedError bool
	}{
		{
			testname: "CheckAllExplicitAttributes",
			reqVal:   buildTestBwsObj(&testInt64, &testInt64),
			expectedVal: &timeofday.TimeOfDay{
				Hours:   30,
				Minutes: 30,
			},
		},
		{
			testname: "CheckPartlyAttributesWithHours",
			reqVal:   buildTestBwsObj(&testInt64, nil),
			expectedVal: &timeofday.TimeOfDay{
				Hours: 30,
			},
		},
		{
			testname: "CheckPartlyAttributesWithMinutes",
			reqVal:   buildTestBwsObj(nil, &testInt64),
			expectedVal: &timeofday.TimeOfDay{
				Minutes: 30,
			},
		},
		{
			testname:    "CheckWithoutAttributes",
			reqVal:      buildTestBwsObj(nil, nil),
			expectedVal: &timeofday.TimeOfDay{},
		},
		{
			testname:    "CheckNullObj",
			reqVal:      types.ObjectNull(expectedBwsAttrTypes),
			expectedVal: &timeofday.TimeOfDay{},
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

	for _, c := range cases {
		diags := diag.Diagnostics{}
		pgBws := expandBackupWindowStart(ctx, c.reqVal, &diags)
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

		if !reflect.DeepEqual(pgBws, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				pgBws,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterLabelsExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Map
		expectedVal   map[string]string
		expectedError bool
	}{
		{
			testname: "CheckSeveralAttributes",
			reqVal: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{"key1": types.StringValue("value1"), "key2": types.StringValue("value2")},
			),
			expectedVal: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			testname: "CheckOneAttribute",
			reqVal: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{"key3": types.StringValue("value3")},
			),
			expectedVal: map[string]string{"key3": "value3"},
		},
		{
			testname: "CheckEmpty",
			reqVal: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{},
			),
			expectedVal: map[string]string{},
		},
		{
			testname:    "CheckNull",
			reqVal:      types.MapNull(types.StringType),
			expectedVal: nil,
		},
		{
			testname:      "CheckNonExpectedStructure",
			reqVal:        types.MapValueMust(types.Int64Type, map[string]attr.Value{"key": types.Int64Value(1)}),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		lbls := expandLabels(ctx, c.reqVal, &diags)
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

		if !reflect.DeepEqual(lbls, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				lbls,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterEnvironmentExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	validEnvs := []string{"PRODUCTION", "PRESTABLE"}
	randValid := validEnvs[rand.Intn(len(validEnvs))]

	cases := []struct {
		testname      string
		reqVal        types.String
		expectedVal   postgresql.Cluster_Environment
		expectedError bool
	}{
		{
			testname:    "CheckValidAttribute",
			reqVal:      types.StringValue(randValid),
			expectedVal: postgresql.Cluster_Environment(postgresql.Cluster_Environment_value[randValid]),
		},
		{
			testname:      "CheckInvalidAttribute",
			reqVal:        types.StringValue("INVALID"),
			expectedError: true,
		},
		{
			testname:    "ChecNullAttribute",
			reqVal:      types.StringNull(),
			expectedVal: postgresql.Cluster_ENVIRONMENT_UNSPECIFIED,
		},
		{
			testname:      "CheckExplicitUnspecifiedAttribute",
			reqVal:        types.StringValue("ENVIRONMENT_UNSPECIFIED"),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		lbls := expandEnvironment(ctx, c.reqVal, &diags)
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

		if !reflect.DeepEqual(lbls, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				lbls,
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

func TestYandexProvider_MDBPostgresClusterSecurityGroupIdsExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Set
		expectedVal   []string
		expectedError bool
	}{
		{
			testname:    "CheckSeveralAttributes",
			reqVal:      types.SetValueMust(types.StringType, []attr.Value{types.StringValue("sg-1"), types.StringValue("sg-2")}),
			expectedVal: []string{"sg-1", "sg-2"},
		},
		{
			testname:    "CheckOneAttribute",
			reqVal:      types.SetValueMust(types.StringType, []attr.Value{types.StringValue("sg")}),
			expectedVal: []string{"sg"},
		},
		{
			testname:    "CheckEmptyAttribute",
			reqVal:      types.SetValueMust(types.StringType, []attr.Value{}),
			expectedVal: []string{},
		},
		{
			testname:    "CheckNullAttribute",
			reqVal:      types.SetNull(types.StringType),
			expectedVal: nil,
		},
		{
			testname:      "CheckInvalidAttribute",
			reqVal:        types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(1)}),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		sg := expandSecurityGroupIds(ctx, c.reqVal, &diags)
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

		if !reflect.DeepEqual(sg, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				sg,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterResourcesExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *postgresql.Resources
		expectedError bool
	}{
		{
			testname: "CheckFullAttribute",
			reqVal: types.ObjectValueMust(
				expectedResourcesAttrs,
				map[string]attr.Value{
					"resource_preset_id": types.StringValue("s1.micro"),
					"disk_type_id":       types.StringValue("network-hdd"),
					"disk_size":          types.Int64Value(13),
				},
			),
			expectedVal: &postgresql.Resources{
				ResourcePresetId: "s1.micro",
				DiskTypeId:       "network-hdd",
				DiskSize:         datasize.ToBytes(13),
			},
		},
		{
			testname: "CheckNullAttribute",
			reqVal: types.ObjectNull(
				expectedResourcesAttrs,
			),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		r := expandResources(ctx, c.reqVal, &diags)
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

		if !reflect.DeepEqual(r, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test:\nexpected %s\nactual %s",
				c.testname,
				c.expectedVal,
				r,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterConfigPgConfigExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		version       string
		reqVal        PgSettingsMapValue
		expectedVal   postgresql.ConfigSpec_PostgresqlConfig
		expectedError bool
	}{
		{
			testname: "CheckPartlyAttributes",
			version:  "15",
			reqVal: PgSettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"max_connections":                types.StringValue("395"),
						"enable_parallel_hash":           types.StringValue("true"),
						"autovacuum_vacuum_scale_factor": types.StringValue("0.34"),
						"default_transaction_isolation":  types.StringValue("TRANSACTION_ISOLATION_READ_COMMITTED"),
						"shared_preload_libraries":       types.StringValue("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"),
					},
				),
			},

			expectedVal: &postgresql.ConfigSpec_PostgresqlConfig_15{
				PostgresqlConfig_15: &config.PostgresqlConfig15{
					MaxConnections:              wrapperspb.Int64(395),
					EnableParallelHash:          wrapperspb.Bool(true),
					AutovacuumVacuumScaleFactor: wrapperspb.Double(0.34),
					DefaultTransactionIsolation: config.PostgresqlConfig15_TRANSACTION_ISOLATION_READ_COMMITTED,
					SharedPreloadLibraries: []config.PostgresqlConfig15_SharedPreloadLibraries{
						config.PostgresqlConfig15_SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN, config.PostgresqlConfig15_SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN,
					},
				},
			},
		},
		{
			testname: "CheckAccessWithRandomAttributes",
			reqVal: PgSettingsMapValue{
				MapValue: types.MapValueMust(
					types.Int64Type,
					map[string]attr.Value{
						"random": types.Int64Value(11),
					},
				),
			},

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
						expectedResourcesAttrs,
						map[string]attr.Value{
							"resource_preset_id": types.StringValue("s1.micro"),
							"disk_type_id":       types.StringValue("network-ssd"),
							"disk_size":          types.Int64Value(10),
						},
					),
					"backup_window_start":       types.ObjectNull(expectedBwsAttrTypes),
					"backup_retain_period_days": types.Int64Null(),
					"autofailover":              types.BoolNull(),
					"access":                    types.ObjectNull(expectedAccessAttrTypes),
					"performance_diagnostics":   types.ObjectNull(expectedPDAttrs),
					"postgresql_config":         NewPgSettingsMapNull(),
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
				Autofailover:           nil,
				Access:                 &postgresql.Access{},
				PerformanceDiagnostics: nil,
				PostgresqlConfig: &postgresql.ConfigSpec_PostgresqlConfig_15{
					PostgresqlConfig_15: &config.PostgresqlConfig15{},
				},
			},
		},
		{
			testname: "CheckFullAttributes",
			reqVal: types.ObjectValueMust(
				expectedConfigAttrs,
				map[string]attr.Value{
					"version": types.StringValue("15"),
					"resources": types.ObjectValueMust(
						expectedResourcesAttrs,
						map[string]attr.Value{
							"resource_preset_id": types.StringValue("s1.micro"),
							"disk_type_id":       types.StringValue("network-ssd"),
							"disk_size":          types.Int64Value(10),
						},
					),
					"backup_window_start": types.ObjectValueMust(
						expectedBwsAttrTypes,
						map[string]attr.Value{
							"hours":   types.Int64Value(23),
							"minutes": types.Int64Value(0),
						},
					),
					"backup_retain_period_days": types.Int64Value(7),
					"autofailover":              types.BoolValue(true),
					"access": types.ObjectValueMust(
						expectedAccessAttrTypes,
						map[string]attr.Value{
							"web_sql":       types.BoolValue(true),
							"serverless":    types.BoolValue(false),
							"data_transfer": types.BoolValue(false),
							"data_lens":     types.BoolValue(true),
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
				Autofailover:           wrapperspb.Bool(true),
				Access: &postgresql.Access{
					WebSql:   true,
					DataLens: true,
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
