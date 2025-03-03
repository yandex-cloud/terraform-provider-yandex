package mdb_mysql_cluster_beta

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestYandexProvider_MDBMySQLClusterPrepateUpdateRequestBasic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cluster := baseCluster

	cluster.Version = types.StringValue("9.0")
	cluster.BackupWindowStart = types.ObjectValueMust(expectedBWSAttrs, map[string]attr.Value{
		"hours":   types.Int64Value(2),
		"minutes": types.Int64Value(0),
	})

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

	expectedUpdateReq := &mysql.UpdateClusterRequest{
		ClusterId: "test-id",
		Name:      "test-cluster-new",
		ConfigSpec: &mysql.ConfigSpec{
			BackupWindowStart: &timeofday.TimeOfDay{
				Hours:   2,
				Minutes: 0,
			},
		},
		MaintenanceWindow:  nil,
		SecurityGroupIds:   []string{"test-sg-new"},
		DeletionProtection: false,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"name",
				"config_spec.backup_window_start.hours",
				"config_spec.backup_window_start.minutes",
				"security_group_ids",
				"deletion_protection",
				"maintenance_window",
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
		t.Fatalf("Unexpected update request:\nexpected %s\nactual %s", req, expectedUpdateReq)
	}
}

func TestYandexProvider_MDBMySQLClusterPrepateUpdateVersionRequest(t *testing.T) {
	t.Parallel()

	cluster := baseCluster

	cluster.Version = types.StringValue("9.0")

	req, diags := prepareVersionUpdateRequest(&baseCluster, &cluster)
	if diags.HasError() {
		t.Fatalf(
			"Unexpected expand diagnostics status: expected without error, actual with errors: %v",
			diags.Errors(),
		)
	}

	expectedUpdateReq := &mysql.UpdateClusterRequest{
		ClusterId: "test-id",
		ConfigSpec: &mysql.ConfigSpec{
			Version: "9.0",
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"config_spec.version",
			},
		},
	}

	for i, path := range req.UpdateMask.Paths {
		if path != expectedUpdateReq.UpdateMask.Paths[i] {
			t.Fatalf("Unexpected update mask paths: expected %s, actual %s", expectedUpdateReq.UpdateMask, req.UpdateMask)
		}
	}

	if !reflect.DeepEqual(req, expectedUpdateReq) {
		t.Fatalf("Unexpected update request:\nexpected %s\nactual %s", req, expectedUpdateReq)
	}
}
