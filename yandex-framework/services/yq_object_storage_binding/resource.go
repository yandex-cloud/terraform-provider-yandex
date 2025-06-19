package yq_object_storage_binding

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/yqcommon"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type objectStorageBindingStrategy struct {
}

func (r *objectStorageBindingStrategy) ExpandSetting(ctx context.Context, plan *tfsdk.Plan, diagnostics *diag.Diagnostics) *Ydb_FederatedQuery.BindingSetting {
	var model objectStorageBindingModel
	diagnostics.Append(plan.Get(ctx, &model)...)
	if diagnostics.HasError() {
		return nil
	}

	projection := make(map[string]string)
	diagnostics.Append(model.Projection.ElementsAs(ctx, &projection, false)...)
	if diagnostics.HasError() {
		return nil
	}

	formatSetting := make(map[string]string)
	diagnostics.Append(model.FormatSetting.ElementsAs(ctx, &formatSetting, false)...)
	if diagnostics.HasError() {
		return nil
	}

	partitionedBy := make([]string, 0, len(model.PartitionedBy.Elements()))
	diagnostics.Append(model.PartitionedBy.ElementsAs(ctx, &partitionedBy, false)...)
	if diagnostics.HasError() {
		return nil
	}

	schema := yqcommon.ParseSchema(ctx, &model.Column, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	return &Ydb_FederatedQuery.BindingSetting{
		Binding: &Ydb_FederatedQuery.BindingSetting_ObjectStorage{
			ObjectStorage: &Ydb_FederatedQuery.ObjectStorageBinding{
				Subset: []*Ydb_FederatedQuery.ObjectStorageBinding_Subset{
					{
						Format:        model.Format.ValueString(),
						Compression:   model.Compression.ValueString(),
						PathPattern:   model.PathPattern.ValueString(),
						Schema:        schema,
						FormatSetting: formatSetting,
						Projection:    projection,
						PartitionedBy: partitionedBy,
					},
				},
			},
		},
	}
}

func (r *objectStorageBindingStrategy) PackToState(ctx context.Context, setting *Ydb_FederatedQuery.BindingSetting, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	var model objectStorageBindingModel
	objectStorageSetting := setting.GetObjectStorage()
	if objectStorageSetting == nil {
		diagnostics.AddError("unexpected null settings", "")
		return
	}

	if objectStorageSetting.Subset == nil {
		diagnostics.AddError("unexpected no subsets", "")
		return
	}

	if len(objectStorageSetting.Subset) != 1 {
		diagnostics.AddError(fmt.Sprintf("unexpected subset count %d", len(objectStorageSetting.Subset)), "")
		return
	}

	subset := objectStorageSetting.Subset[0]
	model.Format = types.StringValue(subset.GetFormat())
	model.Compression = types.StringValue(subset.GetCompression())
	model.PathPattern = types.StringValue(subset.GetPathPattern())

	formatSetting, diag := types.MapValueFrom(ctx, types.StringType, subset.GetFormatSetting())
	if diag.HasError() {
		diagnostics.Append(diag...)
		return
	}
	model.FormatSetting = formatSetting

	projection, diag := types.MapValueFrom(ctx, types.StringType, subset.GetProjection())
	if diag.HasError() {
		diagnostics.Append(diag...)
		return
	}
	model.Projection = projection

	partitionedBy := subset.GetPartitionedBy()
	if partitionedBy != nil {
		pBy, diag := types.ListValueFrom(ctx, types.StringType, partitionedBy)
		if diag.HasError() {
			diagnostics.Append(diag...)
			return
		}
		model.PartitionedBy = pBy
	}

	schema := subset.GetSchema()
	model.Column = yqcommon.FlattenSchema(ctx, schema, diagnostics)
	if diagnostics.HasError() {
		return
	}

	diagnostics.Append(state.Set(ctx, &model)...)
}

func newObjectStorageBindingStrategy() yqcommon.BindingStrategy {
	return &objectStorageBindingStrategy{}
}

func newObjectStorageBindingResourceSchema() (map[string]schema.Attribute, map[string]schema.Block) {
	return yqcommon.NewBindingResourceSchema(
		yqcommon.AttributePathPattern,
		yqcommon.AttributeProjection,
		yqcommon.AttributePartitionedBy,
	)
}

func NewResource() resource.Resource {
	attributes, blocks := newObjectStorageBindingResourceSchema()
	return yqcommon.NewBaseBindingResource(
		attributes,
		blocks,
		newObjectStorageBindingStrategy(),
		"_yq_object_storage_binding",
		"Manages Object Storage binding in Yandex Query service. For more information, see [the official documentation](https://yandex.cloud/docs/query/concepts/glossary#Binding).\n\n")
}
