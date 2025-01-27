package mdb_postgresql_cluster_beta

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
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

var mwAttrsTestFlatten = map[string]attr.Type{
	"type": types.StringType,
	"day":  types.StringType,
	"hour": types.Int64Type,
}

func TestYandexProvider_MDBPostgresClusterMaintenanceWindowFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *postgresql.MaintenanceWindow
		expectedVal types.Object
		hasErr      bool
	}{
		{
			testname: "CheckWeeklyMaintenanceWindow",
			reqVal: &postgresql.MaintenanceWindow{
				Policy: &postgresql.MaintenanceWindow_WeeklyMaintenanceWindow{
					WeeklyMaintenanceWindow: &postgresql.WeeklyMaintenanceWindow{
						Hour: 10,
						Day:  postgresql.WeeklyMaintenanceWindow_WeekDay(1),
					},
				},
			},
			expectedVal: types.ObjectValueMust(mwAttrsTestFlatten, map[string]attr.Value{
				"type": types.StringValue("WEEKLY"),
				"day":  types.StringValue("MON"),
				"hour": types.Int64Value(10),
			}),
		},
		{
			testname: "CheckAnytimeMaintenanceWindow",
			reqVal: &postgresql.MaintenanceWindow{
				Policy: &postgresql.MaintenanceWindow_Anytime{
					Anytime: &postgresql.AnytimeMaintenanceWindow{},
				},
			},
			expectedVal: types.ObjectValueMust(mwAttrsTestFlatten, map[string]attr.Value{
				"type": types.StringValue("ANYTIME"),
				"day":  types.StringNull(),
				"hour": types.Int64Null(),
			}),
		},
		{
			testname:    "CheckNullMaintenanceWindow",
			reqVal:      nil,
			expectedVal: types.ObjectNull(mwAttrsTestFlatten),
			hasErr:      true,
		},
		{
			testname:    "CheckEmptyMaintenanceWindow",
			reqVal:      &postgresql.MaintenanceWindow{},
			expectedVal: types.ObjectNull(mwAttrsTestFlatten),
			hasErr:      true,
		},
		{
			testname: "CheckPolicyNilMaintenanceWindow",
			reqVal: &postgresql.MaintenanceWindow{
				Policy: nil,
			},
			expectedVal: types.ObjectNull(mwAttrsTestFlatten),
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

var pdTestFlatten = map[string]attr.Type{
	"enabled":                      types.BoolType,
	"sessions_sampling_interval":   types.Int64Type,
	"statements_sampling_interval": types.Int64Type,
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
			expectedObject: types.ObjectNull(pdTestFlatten),
		},
		{
			testname: "CheckAllAttributes",
			testData: &postgresql.PerformanceDiagnostics{
				Enabled:                    true,
				SessionsSamplingInterval:   10,
				StatementsSamplingInterval: 5,
			},
			expectedObject: types.ObjectValueMust(pdTestFlatten, map[string]attr.Value{
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
			expectedObject: types.ObjectValueMust(pdTestFlatten, map[string]attr.Value{
				"enabled":                      types.BoolValue(true),
				"sessions_sampling_interval":   types.Int64Value(10),
				"statements_sampling_interval": types.Int64Value(0),
			}),
		},
		{
			testname: "CheckEmptyAttributes",
			testData: &postgresql.PerformanceDiagnostics{},
			expectedObject: types.ObjectValueMust(pdTestFlatten, map[string]attr.Value{
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

func TestYandexProvider_MDBPostgresClusterConfigBackupWindowStartFlattener(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expectedAccessAttrs := map[string]attr.Type{
		"hours":   types.Int64Type,
		"minutes": types.Int64Type,
	}

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
				expectedAccessAttrs, map[string]attr.Value{
					"hours":   types.Int64Value(30),
					"minutes": types.Int64Value(30),
				},
			),
		},
		{
			testname: "CheckAllAttributesWithDefaultValues",
			reqVal:   &timeofday.TimeOfDay{},
			expectedVal: types.ObjectValueMust(
				expectedAccessAttrs, map[string]attr.Value{
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
				expectedAccessAttrs, map[string]attr.Value{
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
				expectedAccessAttrs, map[string]attr.Value{
					"hours":   types.Int64Value(0),
					"minutes": types.Int64Value(30),
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
