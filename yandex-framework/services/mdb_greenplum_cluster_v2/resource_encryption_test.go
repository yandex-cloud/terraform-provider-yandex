package mdb_greenplum_cluster_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func greenplumEncryptionTestPlan(t *testing.T) yandexMdbGreenplumClusterV2Model {
	t.Helper()
	ctx := context.Background()
	plan := NewYandexMdbGreenplumClusterV2Model()
	plan.FolderId = types.StringValue("folder-id")
	plan.Environment = types.StringValue("PRESTABLE")
	plan.NetworkId = types.StringValue("network-id")
	var diags diag.Diagnostics
	plan.Config, diags = types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ConfigModelType.AttrTypes, NewYandexMdbGreenplumClusterV2ConfigModel())
	if diags.HasError() {
		t.Fatal(diags)
	}
	plan.MasterConfig, diags = types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2MasterConfigModelType.AttrTypes, NewYandexMdbGreenplumClusterV2MasterConfigModel())
	if diags.HasError() {
		t.Fatal(diags)
	}
	plan.SegmentConfig, diags = types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2SegmentConfigModelType.AttrTypes, NewYandexMdbGreenplumClusterV2SegmentConfigModel())
	if diags.HasError() {
		t.Fatal(diags)
	}
	plan.Restore = types.ObjectValueMust(map[string]attr.Type{
		"backup_id": types.StringType, "time": types.StringType, "restore_only": types.SetType{ElemType: types.StringType},
	}, map[string]attr.Value{
		"backup_id": types.StringValue("backup-id"), "time": types.StringNull(), "restore_only": types.SetNull(types.StringType),
	})
	return plan
}

func TestGreenplumClusterDiskEncryptionRequestsAndRefresh(t *testing.T) {
	for _, tt := range []struct {
		name string
		key  types.String
	}{
		{"without encryption", types.StringNull()},
		{"with encryption", types.StringValue("kms-key-id")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			plan := greenplumEncryptionTestPlan(t)
			plan.DiskEncryptionKeyId = tt.key
			create, diags := prepareCreateRequest(ctx, &plan, types.StringNull(), &providerconfig.Config{})
			if diags.HasError() {
				t.Fatal(diags)
			}
			if (create.DiskEncryptionKeyId == nil) != tt.key.IsNull() || create.GetDiskEncryptionKeyId().GetValue() != tt.key.ValueString() {
				t.Fatalf("unexpected create encryption key: %v", create.DiskEncryptionKeyId)
			}
			restore, diags := prepareRestoreRequest(ctx, &plan, &providerconfig.State{})
			if diags.HasError() {
				t.Fatal(diags)
			}
			if restore.DiskEncryptionKeyId == nil || restore.DiskEncryptionKeyId.GetValue() != tt.key.ValueString() {
				t.Fatalf("restore must explicitly set encryption key, got %v", restore.DiskEncryptionKeyId)
			}
			if restore.BackupId != "backup-id" || restore.NetworkId != "network-id" {
				t.Fatal("restore options were not preserved")
			}
			state := flattenYandexMdbGreenplumClusterV2(ctx, &greenplum.Cluster{}, plan, plan.Timeouts, &diags)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if !state.DiskEncryptionKeyId.Equal(tt.key) {
				t.Fatalf("refresh lost encryption key: %v", state.DiskEncryptionKeyId)
			}
		})
	}
}

func TestGreenplumClusterDiskEncryptionRequiresReplacement(t *testing.T) {
	ctx := context.Background()
	field := YandexMdbGreenplumClusterV2ResourceSchema(ctx).Attributes["disk_encryption_key_id"].(schema.StringAttribute)
	if !field.IsOptional() || field.IsComputed() {
		t.Fatal("encryption key must be optional")
	}
	for _, key := range []types.String{types.StringValue("another-key"), types.StringNull()} {
		req := planmodifier.StringRequest{StateValue: types.StringValue("old-key"), PlanValue: key}
		// RequiresReplace also checks that the resource exists in both state and plan.
		object := types.ObjectValueMust(map[string]attr.Type{"disk_encryption_key_id": types.StringType}, map[string]attr.Value{"disk_encryption_key_id": req.StateValue})
		value, diags := object.ToTerraformValue(ctx)
		if diags != nil {
			t.Fatal(diags)
		}
		req.State.Raw = value
		req.Plan.Raw = value
		replaced := false
		for _, modifier := range field.PlanModifiers {
			var resp planmodifier.StringResponse
			modifier.PlanModifyString(ctx, req, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatal(resp.Diagnostics)
			}
			replaced = replaced || resp.RequiresReplace
		}
		if !replaced {
			t.Fatalf("changing the encryption key to %v must replace the cluster", key)
		}
	}
}
