package validate

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

func TestValidateResourcesDiskXorDefersUnknownValues(t *testing.T) {
	ctx := context.Background()
	resources, diags := types.ObjectValue(model.NodeResourceAttrTypes, map[string]attr.Value{
		"resource_preset_id": types.StringNull(),
		"disk_size":          types.Int64Unknown(),
		"disk_size_gb":       types.Int64Null(),
		"disk_type_id":       types.StringNull(),
	})
	require.False(t, diags.HasError())

	resp := &resource.ValidateConfigResponse{}
	validateResourcesDiskXor(ctx, resources, path.Root("resources"), resp)

	require.False(t, resp.Diagnostics.HasError())
}

func TestValidateAutoscalingDiskLimitXorDefersUnknownValues(t *testing.T) {
	ctx := context.Background()
	autoscaling, diags := types.ObjectValue(model.DiskSizeAutoscalingAttrTypes, map[string]attr.Value{
		"disk_size_limit":           types.Int64Unknown(),
		"disk_size_gb_limit":        types.Int64Null(),
		"planned_usage_threshold":   types.Int64Null(),
		"emergency_usage_threshold": types.Int64Null(),
	})
	require.False(t, diags.HasError())

	resp := &resource.ValidateConfigResponse{}
	validateAutoscalingDiskLimitXor(ctx, autoscaling, path.Root("disk_size_autoscaling"), resp)

	require.False(t, resp.Diagnostics.HasError())
}
