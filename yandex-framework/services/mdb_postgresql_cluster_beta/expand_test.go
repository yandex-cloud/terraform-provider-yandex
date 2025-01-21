package mdb_postgresql_cluster_beta

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
