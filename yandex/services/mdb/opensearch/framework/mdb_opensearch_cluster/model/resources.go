package model

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type NodeResource struct {
	ResourcePresetID types.String `tfsdk:"resource_preset_id"`
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	DiskSizeGb       types.Int64  `tfsdk:"disk_size_gb"`
	DiskTypeID       types.String `tfsdk:"disk_type_id"`
}

var NodeResourceAttrTypes = map[string]attr.Type{
	"resource_preset_id": types.StringType,
	"disk_size":          types.Int64Type,
	"disk_size_gb":       types.Int64Type,
	"disk_type_id":       types.StringType,
}

func resourcesToObject(ctx context.Context, r *opensearch.Resources) (types.Object, diag.Diagnostics) {
	if isEmptyResources(r) {
		return types.ObjectNull(NodeResourceAttrTypes), diag.Diagnostics{}
	}

	bytes := r.GetDiskSize()
	return types.ObjectValueFrom(ctx, NodeResourceAttrTypes, NodeResource{
		ResourcePresetID: types.StringValue(r.GetResourcePresetId()),
		DiskSize:         types.Int64Value(bytes),
		DiskSizeGb:       types.Int64Value(datasize.ToGigabytes(bytes)),
		DiskTypeID:       types.StringValue(r.GetDiskTypeId()),
	})
}

func isEmptyResources(r *opensearch.Resources) bool {
	return r == nil ||
		(r.DiskSize == 0 && r.DiskTypeId == "" && r.ResourcePresetId == "")
}

type WithResources interface {
	GetResources() types.Object
}

func ParseNodeResource(ctx context.Context, ng WithResources) (*NodeResource, diag.Diagnostics) {
	res := &NodeResource{}
	diags := ng.GetResources().As(ctx, res, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}

// EffectiveDiskSizeBytes returns API disk size in bytes from either disk_size or disk_size_gb.
// When both are present (e.g. after Read), disk_size is authoritative.
func EffectiveDiskSizeBytes(nr *NodeResource) (int64, diag.Diagnostics) {
	var diags diag.Diagnostics
	hasBytes := !nr.DiskSize.IsNull() && !nr.DiskSize.IsUnknown()
	hasGb := !nr.DiskSizeGb.IsNull() && !nr.DiskSizeGb.IsUnknown()
	if hasBytes {
		return nr.DiskSize.ValueInt64(), diags
	}
	if hasGb {
		return datasize.ToBytes(nr.DiskSizeGb.ValueInt64()), diags
	}
	diags.AddError(
		"Invalid resource disk configuration",
		"One of `disk_size` (bytes) or `disk_size_gb` (GiB) is required.",
	)
	return 0, diags
}
