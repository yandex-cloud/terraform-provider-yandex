package mdbcommon

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestYandexProvider_MDBClusterMaintenanceWindowFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *MaintenanceWindowMock
		expectedVal types.Object
		hasErr      bool
	}{
		{
			testname: "CheckWeeklyMaintenanceWindow",
			reqVal: &MaintenanceWindowMock{
				Policy: &WeeklyMaintenanceWindowMock{
					Day:  WeekDayMockMon,
					Hour: 10,
				},
			},
			expectedVal: types.ObjectValueMust(MaintenanceWindowType.AttrTypes, map[string]attr.Value{
				"type": types.StringValue("WEEKLY"),
				"day":  types.StringValue("MON"),
				"hour": types.Int64Value(10),
			}),
		},
		{
			testname: "CheckAnytimeMaintenanceWindow",
			reqVal: &MaintenanceWindowMock{
				Policy: &AnytimePolicyMock{},
			},
			expectedVal: types.ObjectValueMust(MaintenanceWindowType.AttrTypes, map[string]attr.Value{
				"type": types.StringValue("ANYTIME"),
				"day":  types.StringNull(),
				"hour": types.Int64Null(),
			}),
		},
		{
			testname:    "CheckNullMaintenanceWindow",
			reqVal:      nil,
			expectedVal: types.ObjectNull(MaintenanceWindowType.AttrTypes),
			hasErr:      true,
		},
		{
			testname:    "CheckEmptyMaintenanceWindow",
			reqVal:      &MaintenanceWindowMock{},
			expectedVal: types.ObjectNull(MaintenanceWindowType.AttrTypes),
			hasErr:      true,
		},
		{
			testname: "CheckPolicyNilMaintenanceWindow",
			reqVal: &MaintenanceWindowMock{
				Policy: nil,
			},
			expectedVal: types.ObjectNull(MaintenanceWindowType.AttrTypes),
			hasErr:      true,
		},
	}

	for _, c := range cases {
		var diags diag.Diagnostics
		res := FlattenMaintenanceWindow[
			MaintenanceWindowMock,
			WeeklyMaintenanceWindowMock,
			AnytimePolicyMock,
			WeekDayMock,
		](ctx, c.reqVal, &diags)

		if c.hasErr {
			if !diags.HasError() {
				t.Errorf("Unexpected flatten error status: expected %v, actual %v", c.hasErr, diags.HasError())
			}
			continue
		}

		if !c.expectedVal.Equal(res) {
			t.Errorf("Unexpected flatten object result %s: expected %v, actual %v", c.testname, c.expectedVal, res)
		}
	}
}

func TestYandexProvider_MDBClusterResourcesFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        *resouresMock
		expectedVal   types.Object
		expectedError bool
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &resouresMock{
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
		r := FlattenResources(ctx, c.reqVal, &diags)
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

func TestYandexProvider_MDBTimeOfDayFlatten(t *testing.T) {
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
				BackupWindowType.AttrTypes, map[string]attr.Value{
					"hours":   types.Int64Value(30),
					"minutes": types.Int64Value(30),
				},
			),
		},
		{
			testname: "CheckAllAttributesWithDefaultValues",
			reqVal:   &timeofday.TimeOfDay{},
			expectedVal: types.ObjectValueMust(
				BackupWindowType.AttrTypes, map[string]attr.Value{
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
				BackupWindowType.AttrTypes, map[string]attr.Value{
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
				BackupWindowType.AttrTypes, map[string]attr.Value{
					"hours":   types.Int64Value(0),
					"minutes": types.Int64Value(30),
				},
			),
		},
		{
			testname:    "CheckNullObject",
			reqVal:      nil,
			expectedVal: types.ObjectNull(BackupWindowType.AttrTypes),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		bws := FlattenBackupWindowStart(ctx, c.reqVal, &diags)
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
		m := FlattenMapString(ctx, c.reqVal, &diags)
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

func TestYandexProvider_SetStringFlatten(t *testing.T) {
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
			testname:    "CheckNullAttribute",
			reqVal:      nil,
			expectedVal: types.SetNull(types.StringType),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		m := FlattenSetString(ctx, c.reqVal, &diags)
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

func TestYandexProvider_BoolWrapperFlatten(t *testing.T) {
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
		m := FlattenBoolWrapper(ctx, c.reqVal, &diags)
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

func TestYandexProvider_MDBInt64WrapperFlatten(t *testing.T) {
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
		brPd := FlattenInt64Wrapper(ctx, c.reqVal, &diags)
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
