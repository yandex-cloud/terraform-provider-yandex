package mdbcommon

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Maintenance Window test

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

type WeekDayMock int32

const (
	WeekDayMockUnspecidied WeekDayMock = iota
	WeekDayMockMon
	WeekDayMockTue
	WeekDayMockWed
	WeekDayMockThu
	WeekDayMockFri
	WeekDayMockSat
	WeekDayMockSun
)

type WeeklyMaintenanceWindowMock struct {
	Day  WeekDayMock
	Hour int64
}

func (w *WeeklyMaintenanceWindowMock) SetDay(v WeekDayMock) {
	w.Day = v
}

func (w *WeeklyMaintenanceWindowMock) SetHour(v int64) {
	w.Hour = v
}

func (w *WeeklyMaintenanceWindowMock) GetDay() WeekDayMock {
	return w.Day
}

func (w *WeeklyMaintenanceWindowMock) GetHour() int64 {
	return w.Hour
}

type AnytimePolicyMock struct{}

type MaintenanceWindowMock struct {
	Policy any
}

func (m *MaintenanceWindowMock) SetAnytime(v *AnytimePolicyMock) {
	m.Policy = v
}

func (m *MaintenanceWindowMock) SetWeeklyMaintenanceWindow(v *WeeklyMaintenanceWindowMock) {
	m.Policy = v
}

func (m *MaintenanceWindowMock) GetAnytime() *AnytimePolicyMock {
	if p, ok := m.Policy.(*AnytimePolicyMock); ok {
		return p
	}
	return nil
}

func (m *MaintenanceWindowMock) GetWeeklyMaintenanceWindow() *WeeklyMaintenanceWindowMock {
	if p, ok := m.Policy.(*WeeklyMaintenanceWindowMock); ok {
		return p
	}
	return nil
}

func TestYandexProvider_MDBMySQLClusterMaintenanceWindowExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	anytimeType := "ANYTIME"
	weeklyType := "WEEKLY"

	day := "MON"
	var hour int64 = 1

	cases := []struct {
		testname       string
		reqVal         types.Object
		expectedPolicy interface{}
		expectedError  bool
	}{
		{
			testname:       "CheckNullObject",
			reqVal:         types.ObjectNull(mwAttrsTestExpand),
			expectedPolicy: nil,
		},
		{
			testname:       "CheckAnytimeMaintenanceWindow",
			reqVal:         buildMWTestBlockObj(&anytimeType, nil, nil),
			expectedPolicy: &AnytimePolicyMock{},
		},
		{
			testname: "CheckWeeklyMaintenanceWindow",
			reqVal:   buildMWTestBlockObj(&weeklyType, &day, &hour),
			expectedPolicy: &WeeklyMaintenanceWindowMock{
				Hour: hour,
				Day:  WeekDayMockMon,
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
		res := ExpandClusterMaintenanceWindow[
			MaintenanceWindowMock,
			WeeklyMaintenanceWindowMock,
			AnytimePolicyMock,
			WeekDayMock,
		](ctx, c.reqVal, &diags)
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

// Resources

var expectedResourcesAttrs = map[string]attr.Type{
	"resource_preset_id": types.StringType,
	"disk_type_id":       types.StringType,
	"disk_size":          types.Int64Type,
}

type resouresMock struct {
	ResourcePresetId string
	DiskTypeId       string
	DiskSize         int64
}

func (r *resouresMock) SetResourcePresetId(v string) {
	r.ResourcePresetId = v
}

func (r *resouresMock) SetDiskSize(v int64) {
	r.DiskSize = v
}
func (r *resouresMock) SetDiskTypeId(v string) {
	r.DiskTypeId = v
}
func (r *resouresMock) GetResourcePresetId() string {
	return r.ResourcePresetId
}
func (r *resouresMock) GetDiskSize() int64 {
	return r.DiskSize
}
func (r *resouresMock) GetDiskTypeId() string {
	return r.DiskTypeId
}

func TestYandexProvider_MDBMySQLClusterResourcesExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *resouresMock
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
			expectedVal: &resouresMock{
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
			expectedError: false,
			expectedVal:   nil,
		},
		{
			testname:      "CheckBlockWithRandomAttributes",
			reqVal:        types.ObjectValueMust(map[string]attr.Type{"random": types.StringType}, map[string]attr.Value{"random": types.StringValue("s1")}),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		r := ExpandResources[resouresMock](ctx, c.reqVal, &diags)

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
				"Unexpected expand result value %s test:\nexpected %v\nactual %v",
				c.testname,
				c.expectedVal,
				r,
			)
		}
	}
}

func TestYandexProvider_MDBMySQLClusterLabelsExpand(t *testing.T) {
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
		lbls := ExpandLabels(ctx, c.reqVal, &diags)
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

func TestYandexProvider_MDBMySQLClusterSecurityGroupIdsExpand(t *testing.T) {
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
		sg := ExpandSecurityGroupIds(ctx, c.reqVal, &diags)
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

func TestYandexProvider_MDBMySQLClusterBoolWrapperExpand(t *testing.T) {
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
		b := ExpandBoolWrapper(ctx, c.reqVal, &diags)

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

type MockEnvironment int32

const (
	MockEnvironmentUnspecified MockEnvironment = iota
	MockEnvironmentProduction
	MockEnvironmentPrestable
)

func TestYandexProvider_MDBMySQLClusterEnvironmentExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.String
		expectedVal   MockEnvironment
		expectedError bool
	}{
		{
			testname:    "CheckValidAttribute_PRODUCTION",
			reqVal:      types.StringValue("PRODUCTION"),
			expectedVal: MockEnvironmentProduction,
		},
		{
			testname:    "CheckValidAttribute_PRESTABLE",
			reqVal:      types.StringValue("PRESTABLE"),
			expectedVal: MockEnvironmentPrestable,
		},
		{
			testname:      "CheckInvalidAttribute",
			reqVal:        types.StringValue("INVALID"),
			expectedError: true,
		},
		{
			testname:    "ChecNullAttribute",
			reqVal:      types.StringNull(),
			expectedVal: MockEnvironmentUnspecified,
		},
		{
			testname:      "CheckExplicitUnspecifiedAttribute",
			reqVal:        types.StringValue("ENVIRONMENT_UNSPECIFIED"),
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		env := ExpandEnvironment[MockEnvironment](ctx, c.reqVal, &diags)
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

		if !reflect.DeepEqual(env, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %v, actual %v",
				c.testname,
				c.expectedVal,
				env,
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

func TestYandexProvider_MDBTimeOfDayExpand(t *testing.T) {
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
		pgBws := ExpandBackupWindow(ctx, c.reqVal, &diags)
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

func TestYandexProvider_Int64WrapperExpand(t *testing.T) {
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
		pgBrpd := ExpandInt64Wrapper(ctx, c.reqVal, &diags)
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

func buildTestAccessObj(dataLens, dataTransfer, webSql, serverless *bool) types.Object {
	return types.ObjectValueMust(
		AccessAttrTypes, map[string]attr.Value{
			"data_transfer": types.BoolPointerValue(dataTransfer),
			"data_lens":     types.BoolPointerValue(dataLens),
			"serverless":    types.BoolPointerValue(serverless),
			"web_sql":       types.BoolPointerValue(webSql),
		},
	)
}

func TestYandexProvider_MDBCommonAccessExpand(t *testing.T) {
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
			reqVal:        types.ObjectNull(AccessAttrTypes),
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
		access := ExpandAccess[postgresql.Access](ctx, c.reqVal, &diags)
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

		if !reflect.DeepEqual(access, c.expectedVal) {
			t.Errorf(
				"Unexpected expansion result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				access,
			)
		}
	}
}
