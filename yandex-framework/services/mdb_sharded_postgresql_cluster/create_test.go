package mdb_sharded_postgresql_cluster

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	expectedConfigAttrs = map[string]attr.Type{
		"access":                    types.ObjectType{AttrTypes: AccessAttrTypes},
		"backup_window_start":       types.ObjectType{AttrTypes: BackupWindowStartAttrTypes},
		"backup_retain_period_days": types.Int64Type,
		"sharded_postgresql_config": types.ObjectType{AttrTypes: ShardedPostgreSQLConfigAttrTypes},
	}
	expectedResourcesAttrs = map[string]attr.Type{
		"resource_preset_id": types.StringType,
		"disk_type_id":       types.StringType,
		"disk_size":          types.Int64Type,
	}
	expectedBWSAttrs = map[string]attr.Type{
		"hours":   types.Int64Type,
		"minutes": types.Int64Type,
	}
	expectedMWAttrs = map[string]attr.Type{
		"type": types.StringType,
		"day":  types.StringType,
		"hour": types.Int64Type,
	}
	expectedClusterAttrs = map[string]attr.Type{
		"name":                types.StringType,
		"description":         types.StringType,
		"labels":              types.MapType{ElemType: types.StringType},
		"environment":         types.StringType,
		"network_id":          types.StringType,
		"maintenance_window":  types.ObjectType{AttrTypes: expectedMWAttrs},
		"security_group_ids":  types.SetType{ElemType: types.StringType},
		"config":              types.ObjectType{AttrTypes: expectedConfigAttrs},
		"deletion_protection": types.BoolType,
		"folder_id":           types.StringType,
		"hosts":               types.MapType{ElemType: types.StringType},
		"id":                  types.StringType,
	}
	baseConfig = types.ObjectValueMust(
		expectedConfigAttrs,
		map[string]attr.Value{
			"backup_window_start": types.ObjectNull(
				expectedBWSAttrs,
			),
			"backup_retain_period_days": types.Int64Value(7),
			"access":                    types.ObjectNull(AccessAttrTypes),
			"sharded_postgresql_config": types.ObjectValueMust(
				ShardedPostgreSQLConfigAttrTypes,
				map[string]attr.Value{
					"router": types.ObjectValueMust(
						ComponentsAttrTypes,
						map[string]attr.Value{
							"resources": types.ObjectValueMust(
								ResourcesAttrTypes,
								map[string]attr.Value{
									"disk_type_id":       types.StringValue("network-ssd"),
									"resource_preset_id": types.StringValue("s1.micro"),
									"disk_size":          types.Int64Value(10),
								},
							),
							"config": mdbcommon.NewSettingsMapNull(),
						},
					),
					"coordinator": types.ObjectNull(ComponentsAttrTypes),
					"infra":       types.ObjectNull(InfraAttrTypes),
					"balancer":    mdbcommon.NewSettingsMapNull(),
					"common":      mdbcommon.NewSettingsMapNull(),
				},
			),
		},
	)
)

func TestYandexProvider_MDBSPQRClusterPrepareCreateRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *spqr.CreateClusterRequest
		expectedError bool
	}{
		{
			testname: "CheckFullAttributes",
			reqVal: types.ObjectValueMust(
				expectedClusterAttrs,
				map[string]attr.Value{
					"id":          types.StringUnknown(),
					"folder_id":   types.StringValue("test-folder"),
					"name":        types.StringValue("test-cluster"),
					"description": types.StringNull(),
					"labels":      types.MapNull(types.StringType),
					"environment": types.StringValue("PRODUCTION"),
					"network_id":  types.StringValue("test-network"),
					"config":      baseConfig,
					"hosts": types.MapValueMust(types.StringType, map[string]attr.Value{
						"host1": types.StringValue("host1"),
						"host2": types.StringValue("host2"),
					}),
					"maintenance_window":  types.ObjectNull(expectedMWAttrs),
					"deletion_protection": types.BoolNull(),
					"security_group_ids":  types.SetNull(types.StringType),
				},
			),
			expectedVal: &spqr.CreateClusterRequest{
				FolderId:    "test-folder",
				Name:        "test-cluster",
				Description: "",
				Labels:      nil,
				Environment: spqr.Cluster_PRODUCTION,
				NetworkId:   "test-network",
				ConfigSpec: &spqr.ConfigSpec{
					Access:                 &spqr.Access{},
					BackupRetainPeriodDays: wrapperspb.Int64(7),
					BackupWindowStart:      &timeofday.TimeOfDay{},
					SpqrSpec: &spqr.SpqrSpec{
						Router: &spqr.SpqrSpec_Router{
							Resources: &spqr.Resources{
								ResourcePresetId: "s1.micro",
								DiskSize:         10737418240,
								DiskTypeId:       "network-ssd",
							},
							Config: &spqr.RouterSettings{},
						},
						Coordinator: nil,
						Infra:       nil,
						Balancer:    &spqr.BalancerSettings{},
					},
				},
			},
			expectedError: false,
		},
	}

	for _, c := range cases {
		cluster := &Cluster{}
		diags := c.reqVal.As(ctx, cluster, datasize.DefaultOpts)
		if diags.HasError() {
			t.Errorf(
				"Unexpected prepare create status diagnostics status %s test errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		req, diags := prepareCreateRequest(ctx, cluster, &config.State{})
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

		if !reflect.DeepEqual(req, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test:\nexpected %s\nactual %s",
				c.testname,
				c.expectedVal,
				req,
			)
			continue
		}
	}
}
