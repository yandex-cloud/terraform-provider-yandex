package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"
)

type Compression struct {
	Method           types.String `tfsdk:"method"`
	MinPartSize      types.Int64  `tfsdk:"min_part_size"`
	MinPartSizeRatio types.Number `tfsdk:"min_part_size_ratio"`
	Level            types.Int64  `tfsdk:"level"`
}

var CompressionAttrTypes = map[string]attr.Type{
	"method":              types.StringType,
	"min_part_size":       types.Int64Type,
	"min_part_size_ratio": types.NumberType,
	"level":               types.Int64Type,
}

func flattenCompression(ctx context.Context, compression *clickhouseConfig.ClickhouseConfig_Compression, diags *diag.Diagnostics) types.Object {
	if compression == nil {
		return types.ObjectNull(CompressionAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, CompressionAttrTypes, Compression{
			Method:           types.StringValue(compression.Method.Enum().String()),
			MinPartSize:      types.Int64Value(compression.MinPartSize),
			MinPartSizeRatio: types.Number(types.Float64Value(compression.MinPartSizeRatio)),
			Level:            mdbcommon.FlattenInt64Wrapper(ctx, compression.Level, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenListCompression(ctx context.Context, compressions []*clickhouseConfig.ClickhouseConfig_Compression, diags *diag.Diagnostics) types.List {
	if compressions == nil {
		return types.ListNull(types.ObjectType{AttrTypes: CompressionAttrTypes})
	}

	tfCompressions := make([]types.Object, len(compressions))
	for i, c := range compressions {
		tfCompressions[i] = flattenCompression(ctx, c, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: CompressionAttrTypes}, tfCompressions)
	diags.Append(d...)

	return list
}

func expandListCompression(ctx context.Context, c types.List, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_Compression {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	result := make([]*clickhouseConfig.ClickhouseConfig_Compression, 0, len(c.Elements()))
	compressions := make([]Compression, 0, len(c.Elements()))
	diags.Append(c.ElementsAs(ctx, &compressions, false)...)
	if diags.HasError() {
		return nil
	}

	for _, compression := range compressions {
		minPartSizeRatio, _ := compression.MinPartSizeRatio.ValueBigFloat().Float64()

		method := utils.ExpandEnum("method", compression.Method.ValueString(), clickhouseConfig.ClickhouseConfig_Compression_Method_value, diags)
		if diags.HasError() {
			return nil
		}

		result = append(result, &clickhouseConfig.ClickhouseConfig_Compression{
			Method:           clickhouseConfig.ClickhouseConfig_Compression_Method(*method),
			MinPartSize:      compression.MinPartSize.ValueInt64(),
			MinPartSizeRatio: minPartSizeRatio,
			Level:            mdbcommon.ExpandInt64Wrapper(ctx, compression.Level, diags),
		})
	}

	return result
}
