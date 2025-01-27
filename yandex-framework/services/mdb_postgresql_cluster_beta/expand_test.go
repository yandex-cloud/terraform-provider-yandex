package mdb_postgresql_cluster_beta

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

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
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
