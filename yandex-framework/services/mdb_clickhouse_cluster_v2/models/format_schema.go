package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

type FormatSchema struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
	Uri  types.String `tfsdk:"uri"`
}

var FormatSchemaAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"type": types.StringType,
	"uri":  types.StringType,
}

func flattenFormatSchema(ctx context.Context, schema *clickhouse.FormatSchema, diags *diag.Diagnostics) types.Object {
	if schema == nil {
		return types.ObjectNull(FormatSchemaAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, FormatSchemaAttrTypes, FormatSchema{
			Name: types.StringValue(schema.Name),
			Type: types.StringValue(schema.Type.Enum().String()),
			Uri:  types.StringValue(schema.Uri),
		},
	)
	diags.Append(d...)

	return obj
}

func FlattenListFormatSchema(ctx context.Context, schemas []*clickhouse.FormatSchema, diags *diag.Diagnostics) types.Set {
	if schemas == nil {
		return types.SetNull(types.ObjectType{AttrTypes: FormatSchemaAttrTypes})
	}

	tfSchemas := make([]types.Object, len(schemas))
	for i, r := range schemas {
		tfSchemas[i] = flattenFormatSchema(ctx, r, diags)
	}

	set, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: FormatSchemaAttrTypes}, tfSchemas)
	diags.Append(d...)

	return set
}

func ExpandListFormatSchema(ctx context.Context, c types.Set, cid string, diags *diag.Diagnostics) []*clickhouse.FormatSchema {
	emptyList := []*clickhouse.FormatSchema{}

	if c.IsNull() || c.IsUnknown() {
		return emptyList
	}

	result := make([]*clickhouse.FormatSchema, 0, len(c.Elements()))
	schemas := make([]FormatSchema, 0, len(c.Elements()))
	diags.Append(c.ElementsAs(ctx, &schemas, false)...)
	if diags.HasError() {
		return emptyList
	}

	for _, schema := range schemas {
		typeValue := utils.ExpandEnum("type", schema.Type.ValueString(), clickhouse.FormatSchemaType_value, diags)
		if diags.HasError() {
			return emptyList
		}

		result = append(result, &clickhouse.FormatSchema{
			Name:      schema.Name.ValueString(),
			ClusterId: cid,
			Type:      clickhouse.FormatSchemaType(*typeValue),
			Uri:       schema.Uri.ValueString(),
		})
	}

	return result
}
