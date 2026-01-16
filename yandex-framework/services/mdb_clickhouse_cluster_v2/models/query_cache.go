package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type QueryCache struct {
	MaxSizeInBytes      types.Int64 `tfsdk:"max_size_in_bytes"`
	MaxEntries          types.Int64 `tfsdk:"max_entries"`
	MaxEntrySizeInBytes types.Int64 `tfsdk:"max_entry_size_in_bytes"`
	MaxEntrySizeInRows  types.Int64 `tfsdk:"max_entry_size_in_rows"`
}

var QueryCacheAttrTypes = map[string]attr.Type{
	"max_size_in_bytes":       types.Int64Type,
	"max_entries":             types.Int64Type,
	"max_entry_size_in_bytes": types.Int64Type,
	"max_entry_size_in_rows":  types.Int64Type,
}

func flattenQueryCache(ctx context.Context, cache *clickhouseConfig.ClickhouseConfig_QueryCache, diags *diag.Diagnostics) types.Object {
	if cache == nil {
		return types.ObjectNull(QueryCacheAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, QueryCacheAttrTypes, QueryCache{
			MaxSizeInBytes:      mdbcommon.FlattenInt64Wrapper(ctx, cache.MaxSizeInBytes, diags),
			MaxEntries:          mdbcommon.FlattenInt64Wrapper(ctx, cache.MaxEntries, diags),
			MaxEntrySizeInBytes: mdbcommon.FlattenInt64Wrapper(ctx, cache.MaxEntrySizeInBytes, diags),
			MaxEntrySizeInRows:  mdbcommon.FlattenInt64Wrapper(ctx, cache.MaxEntrySizeInRows, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func expandQueryCache(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_QueryCache {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var queryCache QueryCache
	diags.Append(c.As(ctx, &queryCache, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_QueryCache{
		MaxSizeInBytes:      mdbcommon.ExpandInt64Wrapper(ctx, queryCache.MaxSizeInBytes, diags),
		MaxEntries:          mdbcommon.ExpandInt64Wrapper(ctx, queryCache.MaxEntries, diags),
		MaxEntrySizeInBytes: mdbcommon.ExpandInt64Wrapper(ctx, queryCache.MaxEntrySizeInBytes, diags),
		MaxEntrySizeInRows:  mdbcommon.ExpandInt64Wrapper(ctx, queryCache.MaxEntrySizeInRows, diags),
	}
}
