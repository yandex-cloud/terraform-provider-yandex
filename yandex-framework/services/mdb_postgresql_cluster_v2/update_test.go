package mdb_postgresql_cluster_v2

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var baseCluster = Cluster{
	Id:          types.StringValue("test-id"),
	FolderId:    types.StringValue("test-folder"),
	NetworkId:   types.StringValue("test-network"),
	Name:        types.StringValue("test-cluster"),
	Description: types.StringValue("test-description"),
	Environment: types.StringValue("PRODUCTION"),
	Labels: types.MapValueMust(types.StringType, map[string]attr.Value{
		"key": types.StringValue("value"),
	}),
	MaintenanceWindow: types.ObjectValueMust(
		mdbcommon.MaintenanceWindowType.AttrTypes,
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

func TestYandexProvider_MDBPostgresClusterPrepateUpdateRequestBasic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cluster := baseCluster

	newCfg := baseConfig.Attributes()

	cluster.Config = types.ObjectValueMust(expectedConfigAttrs, newCfg)
	cluster.Name = types.StringValue("test-cluster-new")
	cluster.DeletionProtection = types.BoolValue(false)
	cluster.SecurityGroupIds = types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("test-sg-new"),
	})
	cluster.MaintenanceWindow = types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes)

	req, diags := prepareUpdateRequest(ctx, &baseCluster, &cluster)
	if diags.HasError() {
		t.Fatalf(
			"Unexpected expand diagnostics status: expected without error, actual with errors: %v",
			diags.Errors(),
		)
	}

	expectedUpdateReq := &postgresql.UpdateClusterRequest{
		ClusterId:          "test-id",
		Name:               "test-cluster-new",
		MaintenanceWindow:  nil,
		SecurityGroupIds:   []string{"test-sg-new"},
		DeletionProtection: false,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"name", "security_group_ids", "deletion_protection", "maintenance_window"},
		},
	}

	sort.Strings(req.UpdateMask.Paths)
	sort.Strings(expectedUpdateReq.UpdateMask.Paths)

	for i, path := range req.UpdateMask.Paths {
		if path != expectedUpdateReq.UpdateMask.Paths[i] {
			t.Fatalf("Unexpected update mask paths: expected %s, actual %s", req.UpdateMask, expectedUpdateReq.UpdateMask)
		}
	}

	if !reflect.DeepEqual(req, expectedUpdateReq) {
		t.Fatalf("Unexpected update request:\nexpected %s\nactual %s", req, expectedUpdateReq)
	}
}
