package mdb_sharded_postgresql_cluster

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestYandexProvider_MDBSPQRClusterBackupRetainPeriodDaysExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		reqVal      types.Int64
		expectedVal *wrapperspb.Int64Value
	}{
		{
			"CheckExplicit",
			types.Int64Value(10),
			wrapperspb.Int64(10),
		},
		{
			"CheckNull",
			types.Int64Null(),
			nil,
		},
	}

	for _, c := range cases {
		diags := &diag.Diagnostics{}
		res := expandBackupRetainPeriodDays(ctx, c.reqVal, diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected expansion diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !reflect.DeepEqual(res, c.expectedVal) {
			t.Errorf(
				"Unexpected expansion result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				res,
			)
		}
	}
}

func TestYandexProvider_MDBSPQRClusterConfigExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        *ShardedPostgreSQLConfig
		expectedVal   *spqr.SpqrSpec
		expectedError bool
	}{
		{
			"CheckExplicitAttributes",
			&ShardedPostgreSQLConfig{
				Common: NewSettingsMapValueMust(map[string]attr.Value{
					"log_level": types.Int64Value(int64(spqr.LogLevel_DEBUG)),
				}),
				Router: &ComponentConfig{
					Config: NewSettingsMapValueMust(map[string]attr.Value{
						"show_notice_messages": types.StringValue("true"),
						//"default_route_behavior": types.StringValue("ALLOW"),
						//"time_quantiles":                types.StringValue("0.95,0.99"),
						"prefer_same_availability_zone": types.StringValue("true"),
					}),
					Resources: types.ObjectValueMust(ResourcesAttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					}),
				},
				Coordinator: &ComponentConfig{
					Config: NewSettingsMapValueMust(map[string]attr.Value{}),
					Resources: types.ObjectValueMust(ResourcesAttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					}),
				},
				Infra: &InfraConfig{
					Router: NewSettingsMapValueMust(map[string]attr.Value{
						"prefer_same_availability_zone": types.StringValue("true"),
					}),
					Resources: types.ObjectValueMust(ResourcesAttrTypes, map[string]attr.Value{
						"resource_preset_id": types.StringValue("s1.micro"),
						"disk_type_id":       types.StringValue("network-ssd"),
						"disk_size":          types.Int64Value(10),
					}),
				},
				Balancer: NewSettingsMapValueMust(map[string]attr.Value{
					"cpu_threshold":     types.StringValue("0.5"),
					"space_threshold":   types.StringValue("0.5"),
					"stat_interval_sec": types.StringValue("15"),
					"max_move_count":    types.StringValue("10"),
					"keys_per_move":     types.StringValue("10"),
				}),
			},
			&spqr.SpqrSpec{
				Router: &spqr.SpqrSpec_Router{
					Config: &spqr.RouterSettings{
						ShowNoticeMessages: wrapperspb.Bool(true),
						//DefaultRouteBehavior: sharded_postgresql.RouterSettings_ALLOW,
						//TimeQuantiles:              []float64{0.95, 0.99},
						PreferSameAvailabilityZone: wrapperspb.Bool(true),
					},
					Resources: &spqr.Resources{
						ResourcePresetId: "s1.micro",
						DiskSize:         datasize.ToBytes(10),
						DiskTypeId:       "network-ssd",
					},
				},
				Coordinator: &spqr.SpqrSpec_Coordinator{
					Config: &spqr.CoordinatorSettings{},
					Resources: &spqr.Resources{
						ResourcePresetId: "s1.micro",
						DiskSize:         datasize.ToBytes(10),
						DiskTypeId:       "network-ssd",
					},
				},
				Infra: &spqr.SpqrSpec_Infra{
					Router: &spqr.RouterSettings{
						PreferSameAvailabilityZone: wrapperspb.Bool(true),
					},
					Coordinator: &spqr.CoordinatorSettings{},
					Resources: &spqr.Resources{
						ResourcePresetId: "s1.micro",
						DiskSize:         datasize.ToBytes(10),
						DiskTypeId:       "network-ssd",
					},
				},
				LogLevel: spqr.LogLevel_DEBUG,
				Balancer: &spqr.BalancerSettings{
					CpuThreshold:    wrapperspb.Double(0.5),
					SpaceThreshold:  wrapperspb.Double(0.5),
					StatIntervalSec: wrapperspb.Int64(15),
					MaxMoveCount:    wrapperspb.Int64(10),
					KeysPerMove:     wrapperspb.Int64(10),
				},
			},
			false,
		},
		{
			testname: "CheckWithRandomAttributes",
			reqVal: &ShardedPostgreSQLConfig{
				Common: NewSettingsMapValueMust(
					map[string]attr.Value{
						"random": types.Int64Value(11),
					},
				),
			},
			expectedVal: &spqr.SpqrSpec{
				Balancer: &spqr.BalancerSettings{},
			},
			expectedError: true,
		},
	}

	for _, c := range cases {
		diags := &diag.Diagnostics{}
		res := expandSPQRConfig(ctx, *c.reqVal, diags)
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

		if !reflect.DeepEqual(res, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test:\n expected %s\n actual %s",
				c.testname,
				c.expectedVal,
				res,
			)
		}
	}
}
