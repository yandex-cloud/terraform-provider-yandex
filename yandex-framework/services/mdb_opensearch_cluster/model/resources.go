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
	DiskTypeID       types.String `tfsdk:"disk_type_id"`
}

var NodeResourceAttrTypes = map[string]attr.Type{
	"resource_preset_id": types.StringType,
	"disk_size":          types.Int64Type,
	"disk_type_id":       types.StringType,
}

func resourcesToObject(ctx context.Context, r *opensearch.Resources) (types.Object, diag.Diagnostics) {
	if isEmptyResources(r) {
		return types.ObjectNull(NodeResourceAttrTypes), diag.Diagnostics{}
	}

	return types.ObjectValueFrom(ctx, NodeResourceAttrTypes, NodeResource{
		ResourcePresetID: types.StringValue(r.GetResourcePresetId()),
		DiskSize:         types.Int64Value(r.GetDiskSize()),
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
