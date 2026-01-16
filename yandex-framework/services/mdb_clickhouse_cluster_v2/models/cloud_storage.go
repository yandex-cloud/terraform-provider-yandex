package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type CloudStorage struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	MoveFactor       types.Number `tfsdk:"move_factor"`
	DataCacheEnabled types.Bool   `tfsdk:"data_cache_enabled"`
	DataCacheMaxSize types.Int64  `tfsdk:"data_cache_max_size"`
	PreferNotToMerge types.Bool   `tfsdk:"prefer_not_to_merge"`
}

var CloudStorageAttrTypes = map[string]attr.Type{
	"enabled":             types.BoolType,
	"move_factor":         types.NumberType,
	"data_cache_enabled":  types.BoolType,
	"data_cache_max_size": types.Int64Type,
	"prefer_not_to_merge": types.BoolType,
}

func FlattenCloudStorage(ctx context.Context, cloudStorage *clickhouse.CloudStorage, diags *diag.Diagnostics) types.Object {
	if cloudStorage == nil {
		return types.ObjectNull(CloudStorageAttrTypes)
	}

	enabled := cloudStorage.Enabled
	cloudStorageModel := CloudStorage{
		Enabled: types.BoolValue(enabled),
	}

	if enabled {
		moveFactor := types.NumberNull()
		if cloudStorage.MoveFactor != nil {
			moveFactor = types.Number(types.Float64Value(cloudStorage.GetMoveFactor().GetValue()))
		}

		cloudStorageModel.MoveFactor = moveFactor
		cloudStorageModel.DataCacheEnabled = mdbcommon.FlattenBoolWrapper(ctx, cloudStorage.DataCacheEnabled, diags)
		cloudStorageModel.DataCacheMaxSize = mdbcommon.FlattenInt64Wrapper(ctx, cloudStorage.DataCacheMaxSize, diags)
		cloudStorageModel.PreferNotToMerge = mdbcommon.FlattenBoolWrapper(ctx, cloudStorage.PreferNotToMerge, diags)
	}

	obj, d := types.ObjectValueFrom(
		ctx, CloudStorageAttrTypes, cloudStorageModel,
	)
	diags.Append(d...)

	return obj
}

func ExpandCloudStorage(ctx context.Context, cloudStorage types.Object, diags *diag.Diagnostics) *clickhouse.CloudStorage {
	if cloudStorage.IsNull() || cloudStorage.IsUnknown() {
		return nil
	}

	var cs CloudStorage
	if diags.Append(cloudStorage.As(ctx, &cs, datasize.DefaultOpts)...); diags.HasError() {
		return nil
	}

	enabled := cs.Enabled.ValueBool()
	result := &clickhouse.CloudStorage{
		Enabled: enabled,
	}

	if enabled {
		moveFactor, _ := cs.MoveFactor.ValueBigFloat().Float64()

		result.MoveFactor = &wrapperspb.DoubleValue{Value: moveFactor}
		result.DataCacheEnabled = &wrapperspb.BoolValue{Value: cs.DataCacheEnabled.ValueBool()}
		result.DataCacheMaxSize = &wrapperspb.Int64Value{Value: cs.DataCacheMaxSize.ValueInt64()}
		result.PreferNotToMerge = &wrapperspb.BoolValue{Value: cs.PreferNotToMerge.ValueBool()}
	}

	return result
}
