package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

type MLModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
	Uri  types.String `tfsdk:"uri"`
}

var MLModelAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"type": types.StringType,
	"uri":  types.StringType,
}

func flattenMLModel(ctx context.Context, model *clickhouse.MlModel, diags *diag.Diagnostics) types.Object {
	if model == nil {
		return types.ObjectNull(MLModelAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, MLModelAttrTypes, MLModel{
			Name: types.StringValue(model.Name),
			Type: types.StringValue(model.Type.Enum().String()),
			Uri:  types.StringValue(model.Uri),
		},
	)
	diags.Append(d...)

	return obj
}

func FlattenListMLModel(ctx context.Context, models []*clickhouse.MlModel, diags *diag.Diagnostics) types.Set {
	if models == nil {
		return types.SetNull(types.ObjectType{AttrTypes: MLModelAttrTypes})
	}

	tfModels := make([]types.Object, len(models))
	for i, r := range models {
		tfModels[i] = flattenMLModel(ctx, r, diags)
	}

	set, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: MLModelAttrTypes}, tfModels)
	diags.Append(d...)

	return set
}

func ExpandListMLModel(ctx context.Context, c types.Set, cid string, diags *diag.Diagnostics) []*clickhouse.MlModel {
	emptyList := []*clickhouse.MlModel{}

	if c.IsNull() || c.IsUnknown() {
		return emptyList
	}

	result := make([]*clickhouse.MlModel, 0, len(c.Elements()))
	models := make([]MLModel, 0, len(c.Elements()))
	diags.Append(c.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return emptyList
	}

	for _, model := range models {
		typeValue := utils.ExpandEnum("type", model.Type.ValueString(), clickhouse.MlModelType_value, diags)
		if diags.HasError() {
			return emptyList
		}

		result = append(result, &clickhouse.MlModel{
			Name:      model.Name.ValueString(),
			ClusterId: cid,
			Type:      clickhouse.MlModelType(*typeValue),
			Uri:       model.Uri.ValueString(),
		})
	}

	return result
}
