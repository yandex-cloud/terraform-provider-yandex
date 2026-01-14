package mdb_sharded_postgresql_cluster

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	baseCluster = Cluster{
		Id:          types.StringValue("test-id"),
		FolderId:    types.StringValue("test-folder"),
		NetworkId:   types.StringValue("test-network"),
		Name:        types.StringValue("test-name"),
		Description: types.StringValue("test-description"),
		Environment: types.StringValue("PRODUCTION"),
		Labels: types.MapValueMust(types.StringType, map[string]attr.Value{
			"key": types.StringValue("value"),
		}),
		MaintenanceWindow: types.ObjectValueMust(
			expectedMWAttrs,
			map[string]attr.Value{
				"type": types.StringValue("WEEKLY"),
				"day":  types.StringValue("MON"),
				"hour": types.Int64Value(1),
			},
		),
		Config:             baseConfig,
		DeletionProtection: types.BoolValue(true),
		SecurityGroupIds: types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("test-sg"),
		}),
	}
)

func TestYandexProvider_MDBSPQRClusterPrepareUpdateRequestBasic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cluster := baseCluster
	newCfg := types.ObjectValueMust(
		expectedConfigAttrs,
		map[string]attr.Value{
			"backup_window_start": types.ObjectNull(
				expectedBWSAttrs,
			),
			"backup_retain_period_days": types.Int64Value(7),
			"access":                    types.ObjectNull(accessAttrTypes),
			"sharded_postgresql_config": types.ObjectValueMust(
				ShardedPostgreSQLConfigAttrTypes,
				map[string]attr.Value{
					"router": types.ObjectValueMust(
						ComponentsAttrTypes,
						map[string]attr.Value{
							"resources": types.ObjectValueMust(
								ResourcesAttrTypes,
								map[string]attr.Value{
									"disk_type_id":       types.StringValue("local-ssd"),
									"resource_preset_id": types.StringValue("s2.micro"),
									"disk_size":          types.Int64Value(20),
								},
							),
							"config": mdbcommon.NewSettingsMapValueMust(map[string]attr.Value{
								"show_notice_messages": types.BoolValue(true),
							}, attrProvider),
						},
					),
					"coordinator": types.ObjectNull(ComponentsAttrTypes),
					"infra":       types.ObjectNull(InfraAttrTypes),
					"balancer":    mdbcommon.NewSettingsMapNull(),
					"common": mdbcommon.NewSettingsMapValueMust(map[string]attr.Value{
						"log_level": types.Int64Value(1),
					}, attrProvider),
				},
			),
		},
	)

	cluster.Config = types.ObjectValueMust(expectedConfigAttrs, newCfg.Attributes())
	cluster.Name = types.StringValue("test-cluster-new")
	cluster.DeletionProtection = types.BoolValue(false)
	cluster.SecurityGroupIds = types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("test-sg-new"),
	})
	cluster.MaintenanceWindow = types.ObjectNull(expectedMWAttrs)

	req, diags := prepareUpdateRequest(ctx, &baseCluster, &cluster)
	if diags.HasError() {
		t.Fatalf(
			"Unexpected expand diagnostics status: expected without error, actual with errors: %v",
			diags.Errors(),
		)
	}

	expectedUpdateReq := &spqr.UpdateClusterRequest{
		ClusterId: "test-id",
		Name:      "test-cluster-new",
		ConfigSpec: &spqr.ConfigSpec{
			SpqrSpec: &spqr.SpqrSpec{
				LogLevel: spqr.LogLevel_DEBUG,
				Router: &spqr.SpqrSpec_Router{
					Config: &spqr.RouterSettings{
						ShowNoticeMessages: wrapperspb.Bool(true),
					},
					Resources: &spqr.Resources{
						ResourcePresetId: "s2.micro",
						DiskSize:         21474836480,
						DiskTypeId:       "local-ssd",
					},
				},
			},
		},
		MaintenanceWindow:  nil,
		SecurityGroupIds:   []string{"test-sg-new"},
		DeletionProtection: false,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"name",
				"security_group_ids",
				"deletion_protection",
				"maintenance_window",
				"config_spec.spqr_spec.log_level",
				"config_spec.spqr_spec.router.config.show_notice_messages",
				"config_spec.spqr_spec.router.resources",
			},
		},
	}

	sort.Strings(req.UpdateMask.Paths)
	sort.Strings(expectedUpdateReq.UpdateMask.Paths)

	for i, path := range req.UpdateMask.Paths {
		if path != expectedUpdateReq.UpdateMask.Paths[i] {
			t.Fatalf("Unexpected update mask paths: expected %s, actual %s", expectedUpdateReq.UpdateMask, req.UpdateMask)
		}
	}

	if !reflect.DeepEqual(req, expectedUpdateReq) {
		t.Fatalf("Unexpected update request:\nexpected %s\nactual %s", expectedUpdateReq, req)
	}
}
