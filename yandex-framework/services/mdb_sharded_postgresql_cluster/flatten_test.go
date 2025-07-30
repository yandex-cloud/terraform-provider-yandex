package mdb_sharded_postgresql_cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestYandexProvider_MDBSPQRClusterConfigAccessFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *spqr.Access
		expectedVal types.Object
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &spqr.Access{
				WebSql:   true,
				DataLens: true,
			},
			expectedVal: types.ObjectValueMust(
				AccessAttrTypes, map[string]attr.Value{
					"data_lens":     types.BoolValue(true),
					"serverless":    types.BoolValue(false),
					"data_transfer": types.BoolValue(false),
					"web_sql":       types.BoolValue(true),
				},
			),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		access := flattenAccess(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf("Unexpected flatten diagnostics status %s test: errors: %v", c.testname, diags.Errors())
			continue
		}

		if !c.expectedVal.Equal(access) {
			t.Errorf("Unexpected flatten result value %s test: expected %s, actual %s", c.testname, c.expectedVal, access)
		}
	}
}

func TestYandexProvider_MDBSPQRClusterMaintenanceWindowFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *spqr.MaintenanceWindow
		expectedVal types.Object
		hasErr      bool
	}{
		{
			testname: "CheckWeeklyMaintenanceWindow",
			reqVal: &spqr.MaintenanceWindow{
				Policy: &spqr.MaintenanceWindow_WeeklyMaintenanceWindow{
					WeeklyMaintenanceWindow: &spqr.WeeklyMaintenanceWindow{
						Hour: 10,
						Day:  spqr.WeeklyMaintenanceWindow_FRI,
					},
				},
			},
			expectedVal: types.ObjectValueMust(MaintenanceWindowAttrTypes, map[string]attr.Value{
				"type": types.StringValue("WEEKLY"),
				"day":  types.StringValue("FRI"),
				"hour": types.Int64Value(10),
			}),
		},
		{
			testname: "CheckAnytimeMaintenanceWindow",
			reqVal: &spqr.MaintenanceWindow{
				Policy: &spqr.MaintenanceWindow_Anytime{
					Anytime: &spqr.AnytimeMaintenanceWindow{},
				},
			},
			expectedVal: types.ObjectValueMust(MaintenanceWindowAttrTypes, map[string]attr.Value{
				"type": types.StringValue("ANYTIME"),
				"day":  types.StringNull(),
				"hour": types.Int64Null(),
			}),
		},
		{
			testname:    "CheckNullMaintenanceWindow",
			reqVal:      nil,
			expectedVal: types.ObjectNull(MaintenanceWindowAttrTypes),
			hasErr:      true,
		},
		{
			testname:    "CheckEmptyMaintenanceWindow",
			reqVal:      &spqr.MaintenanceWindow{},
			expectedVal: types.ObjectNull(MaintenanceWindowAttrTypes),
			hasErr:      true,
		},
		{
			testname: "CheckPolicyNilMaintenanceWindow",
			reqVal: &spqr.MaintenanceWindow{
				Policy: nil,
			},
			expectedVal: types.ObjectNull(MaintenanceWindowAttrTypes),
			hasErr:      true,
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		res := flattenMaintenanceWindow(ctx, c.reqVal, &diags)
		if c.hasErr {
			if !diags.HasError() {
				t.Errorf("Unexpected flatten error status: expected %v, actual %v", c.hasErr, diags.HasError())
				continue
			}
		}

		if !c.expectedVal.Equal(res) {
			t.Errorf("Unexpected flatten object result: expected %v, actual %v", c.expectedVal, res)
		}
	}
}

func TestYandexProvider_MDBSPQRClusterConfigBackupRetainPeriodDays(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      *wrapperspb.Int64Value
		expectedVal types.Int64
	}{
		{
			testname:    "ExplicitCheck",
			reqVal:      wrapperspb.Int64(5),
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
		res := flattenBackupRetainPeriodDays(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf("Unexpected flatten diagnostics status %s test: errors: %v", c.testname, diags.Errors())
			continue
		}

		if !c.expectedVal.Equal(res) {
			t.Errorf("Unexpected flatten result value %s test: expected %s, actual %s", c.testname, c.expectedVal, res)
		}
	}
}

func TestYandexProvider_MDBSPQRClusterConfigBackupWindowStart(t *testing.T) {
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
				Hours:   10,
				Minutes: 10,
			},
			expectedVal: types.ObjectValueMust(
				BackupWindowStartAttrTypes, map[string]attr.Value{
					"hours":   types.Int64Value(10),
					"minutes": types.Int64Value(10),
				},
			),
		},
		{
			testname: "CheckAllAttributesWithDefaultValues",
			reqVal:   &timeofday.TimeOfDay{},
			expectedVal: types.ObjectValueMust(
				BackupWindowStartAttrTypes, map[string]attr.Value{
					"hours":   types.Int64Value(0),
					"minutes": types.Int64Value(0),
				},
			),
		},
		{
			testname: "CheckPartlyAttributesWithHours",
			reqVal: &timeofday.TimeOfDay{
				Hours: 12,
			},
			expectedVal: types.ObjectValueMust(
				BackupWindowStartAttrTypes, map[string]attr.Value{
					"hours":   types.Int64Value(12),
					"minutes": types.Int64Value(0),
				},
			),
		},
		{
			testname: "CheckPartlyAttributesWithMinutes",
			reqVal: &timeofday.TimeOfDay{
				Minutes: 10,
			},
			expectedVal: types.ObjectValueMust(
				BackupWindowStartAttrTypes, map[string]attr.Value{
					"hours":   types.Int64Value(0),
					"minutes": types.Int64Value(10),
				},
			),
		},
		{
			testname:    "CheckNullObject",
			reqVal:      nil,
			expectedVal: types.ObjectNull(BackupWindowStartAttrTypes),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		res := flattenBackupWindowStart(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf("Unexpected flatten diagnostics status %s test: errors: %v", c.testname, diags.Errors())
			continue
		}

		if !c.expectedVal.Equal(res) {
			t.Errorf("Unexpected flatten result value %s test: expected %s, actual %s", c.testname, c.expectedVal, res)
		}
	}
}

func TestYandexProvider_MDBSPQRClusterConfigSPQRConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        *spqr.SPQRConfig
		expectedVal   *ShardedPostgreSQLConfig
		expectedError bool
	}{
		{
			testname: "CheckFullAttributes",
			reqVal: &spqr.SPQRConfig{
				Router: &spqr.RouterConfig{
					Config: &spqr.RouterSettings{
						ShowNoticeMessages: wrapperspb.Bool(true),
						TimeQuantiles:      []float64{0.95, 0.99},
						//DefaultRouteBehavior:       sharded_postgresql.RouterSettings_ALLOW,
						PreferSameAvailabilityZone: wrapperspb.Bool(true),
					},
					Resources: &spqr.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
				},
				Coordinator: &spqr.CoordinatorConfig{
					Config: &spqr.CoordinatorSettings{},
					Resources: &spqr.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
				},
				Infra: &spqr.InfraConfig{
					Router: &spqr.RouterSettings{
						ShowNoticeMessages: wrapperspb.Bool(true),
						TimeQuantiles:      []float64{0.9, 0.95},
						//DefaultRouteBehavior:       sharded_postgresql.RouterSettings_ALLOW,
						PreferSameAvailabilityZone: wrapperspb.Bool(true),
					},
					Coordinator: &spqr.CoordinatorSettings{},
					Resources: &spqr.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
				},
				Balancer: &spqr.BalancerSettings{
					CpuThreshold:    wrapperspb.Double(0.5),
					SpaceThreshold:  wrapperspb.Double(0.5),
					StatIntervalSec: wrapperspb.Int64(15),
					MaxMoveCount:    wrapperspb.Int64(10),
					KeysPerMove:     wrapperspb.Int64(10),
				},
				LogLevel: spqr.LogLevel_DEBUG,
			},
			expectedVal: &ShardedPostgreSQLConfig{
				Common: NewSettingsMapValueMust(
					map[string]attr.Value{
						"log_level": types.Int64Value(int64(spqr.LogLevel_DEBUG)),
					},
				),
				Router: &ComponentConfig{
					Config: mdbcommon.SettingsMapValue{
						MapValue: types.MapValueMust(
							types.StringType,
							map[string]attr.Value{
								"show_notice_messages": types.StringValue("true"),
								"time_quantiles":       types.StringValue("0.95,0.99"),
								//"default_route_behavior":        types.StringValue("2"),
								"prefer_same_availability_zone": types.StringValue("true"),
							},
						),
					},
					Resources: types.ObjectValueMust(ResourcesAttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					}),
				},
				Coordinator: &ComponentConfig{
					Config: NewSettingsMapEmpty(),
					Resources: types.ObjectValueMust(ResourcesAttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					}),
				},
				Infra: &InfraConfig{
					Router: mdbcommon.SettingsMapValue{
						MapValue: types.MapValueMust(
							types.StringType,
							map[string]attr.Value{
								"show_notice_messages": types.StringValue("true"),
								"time_quantiles":       types.StringValue("0.90,0.95"),
								//"default_route_behavior":        types.StringValue("2"),
								"prefer_same_availability_zone": types.StringValue("true"),
							},
						),
					},
					Coordinator: NewSettingsMapEmpty(),
					Resources: types.ObjectValueMust(ResourcesAttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					}),
				},
				Balancer: mdbcommon.SettingsMapValue{
					MapValue: types.MapValueMust(types.StringType, map[string]attr.Value{
						"cpu_threshold":     types.StringValue("0.5"),
						"space_threshold":   types.StringValue("0.5"),
						"stat_interval_sec": types.StringValue("15"),
						"max_move_count":    types.StringValue("10"),
						"keys_per_move":     types.StringValue("10"),
					}),
				},
			},
		},
		{
			testname: "CheckPartialyFilled",
			reqVal: &spqr.SPQRConfig{
				Router: &spqr.RouterConfig{
					Config: &spqr.RouterSettings{
						ShowNoticeMessages: wrapperspb.Bool(true),
						//DefaultRouteBehavior: sharded_postgresql.RouterSettings_BLOCK,
					},
					Resources: &spqr.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
				},
				Coordinator: &spqr.CoordinatorConfig{
					Resources: &spqr.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
				},
			},
			expectedVal: &ShardedPostgreSQLConfig{
				Common: NewSettingsMapValueMust(map[string]attr.Value{}),
				Router: &ComponentConfig{
					Config: mdbcommon.SettingsMapValue{
						MapValue: types.MapValueMust(types.StringType, map[string]attr.Value{
							"show_notice_messages": types.StringValue("true"),
							//"default_route_behavior": types.StringValue("1"),
						}),
					},
					Resources: types.ObjectValueMust(ResourcesAttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					}),
				},
				Coordinator: &ComponentConfig{
					Config: NewSettingsMapEmpty(),
					Resources: types.ObjectValueMust(ResourcesAttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					}),
				},
				Infra:    nil,
				Balancer: NewSettingsMapEmpty(),
			},
		},
		{
			testname: "CheckNull",
			reqVal:   nil,
			expectedVal: &ShardedPostgreSQLConfig{
				Common:   NewSettingsMapEmpty(),
				Balancer: NewSettingsMapEmpty(),
			},
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		res := flattenSPQRConfig(ctx, Config{}, c.reqVal, &diags)
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

		check := func(t *testing.T, cc attr.Value, exp attr.Value) {
			if (cc.IsNull() && !exp.IsNull()) || (!cc.IsNull() && exp.IsNull()) || (!exp.IsNull() && !exp.Equal(cc)) {
				t.Errorf(
					"Unexpected flatten result value %s test: expected %s, actual %s",
					c.testname,
					exp,
					cc,
				)
			}
		}

		resNil := res.Coordinator == nil
		expNil := c.expectedVal.Coordinator == nil
		if !resNil && !expNil {
			check(t, res.Router.Config, c.expectedVal.Router.Config)
			check(t, res.Router.Resources, c.expectedVal.Router.Resources)
		} else if (resNil && !expNil) || (!resNil && expNil) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal.Coordinator,
				res.Coordinator,
			)
		}

		if !resNil && !expNil {
			check(t, res.Coordinator.Config, c.expectedVal.Coordinator.Config)
			check(t, res.Coordinator.Resources, c.expectedVal.Coordinator.Resources)
		} else if (resNil && !expNil) || (!resNil && expNil) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal.Coordinator,
				res.Coordinator,
			)
		}

		resNil = res.Infra == nil
		expNil = c.expectedVal.Infra == nil
		if !resNil && !expNil {
			check(t, res.Infra.Router, c.expectedVal.Infra.Router)
			check(t, res.Infra.Coordinator, c.expectedVal.Infra.Coordinator)
			check(t, res.Infra.Router, c.expectedVal.Infra.Router)
			check(t, res.Infra.Resources, c.expectedVal.Infra.Resources)
		} else if (resNil && !expNil) || (!resNil && expNil) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal.Infra,
				res.Infra,
			)
		}
		check(t, res.Common, c.expectedVal.Common)
	}
}
