package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"
)

type MergeTreeConfig struct {
	ReplicatedDeduplicationWindow                  types.Int64  `tfsdk:"replicated_deduplication_window"`
	ReplicatedDeduplicationWindowSeconds           types.Int64  `tfsdk:"replicated_deduplication_window_seconds"`
	PartsToDelayInsert                             types.Int64  `tfsdk:"parts_to_delay_insert"`
	PartsToThrowInsert                             types.Int64  `tfsdk:"parts_to_throw_insert"`
	MaxReplicatedMergesInQueue                     types.Int64  `tfsdk:"max_replicated_merges_in_queue"`
	NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge types.Int64  `tfsdk:"number_of_free_entries_in_pool_to_lower_max_size_of_merge"`
	MaxBytesToMergeAtMinSpaceInPool                types.Int64  `tfsdk:"max_bytes_to_merge_at_min_space_in_pool"`
	MaxBytesToMergeAtMaxSpaceInPool                types.Int64  `tfsdk:"max_bytes_to_merge_at_max_space_in_pool"`
	InactivePartsToDelayInsert                     types.Int64  `tfsdk:"inactive_parts_to_delay_insert"`
	InactivePartsToThrowInsert                     types.Int64  `tfsdk:"inactive_parts_to_throw_insert"`
	MinBytesForWidePart                            types.Int64  `tfsdk:"min_bytes_for_wide_part"`
	MinRowsForWidePart                             types.Int64  `tfsdk:"min_rows_for_wide_part"`
	TtlOnlyDropParts                               types.Bool   `tfsdk:"ttl_only_drop_parts"`
	MergeWithTtlTimeout                            types.Int64  `tfsdk:"merge_with_ttl_timeout"`
	MergeWithRecompressionTtlTimeout               types.Int64  `tfsdk:"merge_with_recompression_ttl_timeout"`
	MaxPartsInTotal                                types.Int64  `tfsdk:"max_parts_in_total"`
	MaxNumberOfMergesWithTtlInPool                 types.Int64  `tfsdk:"max_number_of_merges_with_ttl_in_pool"`
	CleanupDelayPeriod                             types.Int64  `tfsdk:"cleanup_delay_period"`
	NumberOfFreeEntriesInPoolToExecuteMutation     types.Int64  `tfsdk:"number_of_free_entries_in_pool_to_execute_mutation"`
	MaxAvgPartSizeForTooManyParts                  types.Int64  `tfsdk:"max_avg_part_size_for_too_many_parts"`
	MinAgeToForceMergeSeconds                      types.Int64  `tfsdk:"min_age_to_force_merge_seconds"`
	MinAgeToForceMergeOnPartitionOnly              types.Bool   `tfsdk:"min_age_to_force_merge_on_partition_only"`
	MergeSelectingSleepMs                          types.Int64  `tfsdk:"merge_selecting_sleep_ms"`
	CheckSampleColumnIsCorrect                     types.Bool   `tfsdk:"check_sample_column_is_correct"`
	MergeMaxBlockSize                              types.Int64  `tfsdk:"merge_max_block_size"`
	MaxMergeSelectingSleepMs                       types.Int64  `tfsdk:"max_merge_selecting_sleep_ms"`
	MaxCleanupDelayPeriod                          types.Int64  `tfsdk:"max_cleanup_delay_period"`
	DeduplicateMergeProjectionMode                 types.String `tfsdk:"deduplicate_merge_projection_mode"`
	LightweightMutationProjectionMode              types.String `tfsdk:"lightweight_mutation_projection_mode"`
	MaterializeTtlRecalculateOnly                  types.Bool   `tfsdk:"materialize_ttl_recalculate_only"`
	FsyncAfterInsert                               types.Bool   `tfsdk:"fsync_after_insert"`
	FsyncPartDirectory                             types.Bool   `tfsdk:"fsync_part_directory"`
	MinCompressedBytesToFsyncAfterFetch            types.Int64  `tfsdk:"min_compressed_bytes_to_fsync_after_fetch"`
	MinCompressedBytesToFsyncAfterMerge            types.Int64  `tfsdk:"min_compressed_bytes_to_fsync_after_merge"`
	MinRowsToFsyncAfterMerge                       types.Int64  `tfsdk:"min_rows_to_fsync_after_merge"`
}

var MergeTreeConfigAttrTypes = map[string]attr.Type{
	"replicated_deduplication_window":                           types.Int64Type,
	"replicated_deduplication_window_seconds":                   types.Int64Type,
	"parts_to_delay_insert":                                     types.Int64Type,
	"parts_to_throw_insert":                                     types.Int64Type,
	"max_replicated_merges_in_queue":                            types.Int64Type,
	"number_of_free_entries_in_pool_to_lower_max_size_of_merge": types.Int64Type,
	"max_bytes_to_merge_at_min_space_in_pool":                   types.Int64Type,
	"max_bytes_to_merge_at_max_space_in_pool":                   types.Int64Type,
	"inactive_parts_to_delay_insert":                            types.Int64Type,
	"inactive_parts_to_throw_insert":                            types.Int64Type,
	"min_bytes_for_wide_part":                                   types.Int64Type,
	"min_rows_for_wide_part":                                    types.Int64Type,
	"ttl_only_drop_parts":                                       types.BoolType,
	"merge_with_ttl_timeout":                                    types.Int64Type,
	"merge_with_recompression_ttl_timeout":                      types.Int64Type,
	"max_parts_in_total":                                        types.Int64Type,
	"max_number_of_merges_with_ttl_in_pool":                     types.Int64Type,
	"cleanup_delay_period":                                      types.Int64Type,
	"number_of_free_entries_in_pool_to_execute_mutation":        types.Int64Type,
	"max_avg_part_size_for_too_many_parts":                      types.Int64Type,
	"min_age_to_force_merge_seconds":                            types.Int64Type,
	"min_age_to_force_merge_on_partition_only":                  types.BoolType,
	"merge_selecting_sleep_ms":                                  types.Int64Type,
	"check_sample_column_is_correct":                            types.BoolType,
	"merge_max_block_size":                                      types.Int64Type,
	"max_merge_selecting_sleep_ms":                              types.Int64Type,
	"max_cleanup_delay_period":                                  types.Int64Type,
	"deduplicate_merge_projection_mode":                         types.StringType,
	"lightweight_mutation_projection_mode":                      types.StringType,
	"materialize_ttl_recalculate_only":                          types.BoolType,
	"fsync_after_insert":                                        types.BoolType,
	"fsync_part_directory":                                      types.BoolType,
	"min_compressed_bytes_to_fsync_after_fetch":                 types.Int64Type,
	"min_compressed_bytes_to_fsync_after_merge":                 types.Int64Type,
	"min_rows_to_fsync_after_merge":                             types.Int64Type,
}

func flattenMergeTree(ctx context.Context, config *clickhouseConfig.ClickhouseConfig_MergeTree, diags *diag.Diagnostics) types.Object {
	if config == nil {
		return types.ObjectNull(MergeTreeConfigAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, MergeTreeConfigAttrTypes, MergeTreeConfig{
			ReplicatedDeduplicationWindow:                  mdbcommon.FlattenInt64Wrapper(ctx, config.ReplicatedDeduplicationWindow, diags),
			ReplicatedDeduplicationWindowSeconds:           mdbcommon.FlattenInt64Wrapper(ctx, config.ReplicatedDeduplicationWindowSeconds, diags),
			PartsToDelayInsert:                             mdbcommon.FlattenInt64Wrapper(ctx, config.PartsToDelayInsert, diags),
			PartsToThrowInsert:                             mdbcommon.FlattenInt64Wrapper(ctx, config.PartsToThrowInsert, diags),
			MaxReplicatedMergesInQueue:                     mdbcommon.FlattenInt64Wrapper(ctx, config.MaxReplicatedMergesInQueue, diags),
			NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge: mdbcommon.FlattenInt64Wrapper(ctx, config.NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge, diags),
			MaxBytesToMergeAtMinSpaceInPool:                mdbcommon.FlattenInt64Wrapper(ctx, config.MaxBytesToMergeAtMinSpaceInPool, diags),
			MaxBytesToMergeAtMaxSpaceInPool:                mdbcommon.FlattenInt64Wrapper(ctx, config.MaxBytesToMergeAtMaxSpaceInPool, diags),
			InactivePartsToDelayInsert:                     mdbcommon.FlattenInt64Wrapper(ctx, config.InactivePartsToDelayInsert, diags),
			InactivePartsToThrowInsert:                     mdbcommon.FlattenInt64Wrapper(ctx, config.InactivePartsToThrowInsert, diags),
			MinBytesForWidePart:                            mdbcommon.FlattenInt64Wrapper(ctx, config.MinBytesForWidePart, diags),
			MinRowsForWidePart:                             mdbcommon.FlattenInt64Wrapper(ctx, config.MinRowsForWidePart, diags),
			TtlOnlyDropParts:                               mdbcommon.FlattenBoolWrapper(ctx, config.TtlOnlyDropParts, diags),
			MergeWithTtlTimeout:                            mdbcommon.FlattenInt64Wrapper(ctx, config.MergeWithTtlTimeout, diags),
			MergeWithRecompressionTtlTimeout:               mdbcommon.FlattenInt64Wrapper(ctx, config.MergeWithRecompressionTtlTimeout, diags),
			MaxPartsInTotal:                                mdbcommon.FlattenInt64Wrapper(ctx, config.MaxPartsInTotal, diags),
			MaxNumberOfMergesWithTtlInPool:                 mdbcommon.FlattenInt64Wrapper(ctx, config.MaxNumberOfMergesWithTtlInPool, diags),
			CleanupDelayPeriod:                             mdbcommon.FlattenInt64Wrapper(ctx, config.CleanupDelayPeriod, diags),
			NumberOfFreeEntriesInPoolToExecuteMutation:     mdbcommon.FlattenInt64Wrapper(ctx, config.NumberOfFreeEntriesInPoolToExecuteMutation, diags),
			MaxAvgPartSizeForTooManyParts:                  mdbcommon.FlattenInt64Wrapper(ctx, config.MaxAvgPartSizeForTooManyParts, diags),
			MinAgeToForceMergeSeconds:                      mdbcommon.FlattenInt64Wrapper(ctx, config.MinAgeToForceMergeSeconds, diags),
			MinAgeToForceMergeOnPartitionOnly:              mdbcommon.FlattenBoolWrapper(ctx, config.MinAgeToForceMergeOnPartitionOnly, diags),
			MergeSelectingSleepMs:                          mdbcommon.FlattenInt64Wrapper(ctx, config.MergeSelectingSleepMs, diags),
			CheckSampleColumnIsCorrect:                     mdbcommon.FlattenBoolWrapper(ctx, config.CheckSampleColumnIsCorrect, diags),
			MergeMaxBlockSize:                              mdbcommon.FlattenInt64Wrapper(ctx, config.MergeMaxBlockSize, diags),
			MaxMergeSelectingSleepMs:                       mdbcommon.FlattenInt64Wrapper(ctx, config.MaxMergeSelectingSleepMs, diags),
			MaxCleanupDelayPeriod:                          mdbcommon.FlattenInt64Wrapper(ctx, config.MaxCleanupDelayPeriod, diags),
			DeduplicateMergeProjectionMode:                 types.StringValue(config.DeduplicateMergeProjectionMode.Enum().String()),
			LightweightMutationProjectionMode:              types.StringValue(config.LightweightMutationProjectionMode.Enum().String()),
			MaterializeTtlRecalculateOnly:                  mdbcommon.FlattenBoolWrapper(ctx, config.MaterializeTtlRecalculateOnly, diags),
			FsyncAfterInsert:                               mdbcommon.FlattenBoolWrapper(ctx, config.FsyncAfterInsert, diags),
			FsyncPartDirectory:                             mdbcommon.FlattenBoolWrapper(ctx, config.FsyncPartDirectory, diags),
			MinCompressedBytesToFsyncAfterFetch:            mdbcommon.FlattenInt64Wrapper(ctx, config.MinCompressedBytesToFsyncAfterFetch, diags),
			MinCompressedBytesToFsyncAfterMerge:            mdbcommon.FlattenInt64Wrapper(ctx, config.MinCompressedBytesToFsyncAfterMerge, diags),
			MinRowsToFsyncAfterMerge:                       mdbcommon.FlattenInt64Wrapper(ctx, config.MinRowsToFsyncAfterMerge, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func expandMergeTree(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_MergeTree {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var mergeTree MergeTreeConfig
	diags.Append(c.As(ctx, &mergeTree, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	deduplicateMergeProjectionModeValue := utils.ExpandEnum("deduplicate_merge_projection_mode", mergeTree.DeduplicateMergeProjectionMode.ValueString(), clickhouseConfig.ClickhouseConfig_MergeTree_DeduplicateMergeProjectionMode_value, diags)
	if diags.HasError() {
		return nil
	}

	lightweightMutationProjectionModeValue := utils.ExpandEnum("lightweight_mutation_projection_mode", mergeTree.LightweightMutationProjectionMode.ValueString(), clickhouseConfig.ClickhouseConfig_MergeTree_LightweightMutationProjectionMode_value, diags)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_MergeTree{
		PartsToDelayInsert:                             mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.PartsToDelayInsert, diags),
		PartsToThrowInsert:                             mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.PartsToThrowInsert, diags),
		InactivePartsToDelayInsert:                     mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.InactivePartsToDelayInsert, diags),
		InactivePartsToThrowInsert:                     mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.InactivePartsToThrowInsert, diags),
		MaxAvgPartSizeForTooManyParts:                  mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MaxAvgPartSizeForTooManyParts, diags),
		MaxPartsInTotal:                                mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MaxPartsInTotal, diags),
		MaxReplicatedMergesInQueue:                     mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MaxReplicatedMergesInQueue, diags),
		NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge: mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge, diags),
		NumberOfFreeEntriesInPoolToExecuteMutation:     mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.NumberOfFreeEntriesInPoolToExecuteMutation, diags),
		MaxBytesToMergeAtMinSpaceInPool:                mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MaxBytesToMergeAtMinSpaceInPool, diags),
		MaxBytesToMergeAtMaxSpaceInPool:                mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MaxBytesToMergeAtMaxSpaceInPool, diags),
		MinBytesForWidePart:                            mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MinBytesForWidePart, diags),
		MinRowsForWidePart:                             mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MinRowsForWidePart, diags),
		CleanupDelayPeriod:                             mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.CleanupDelayPeriod, diags),
		MaxCleanupDelayPeriod:                          mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MaxCleanupDelayPeriod, diags),
		MergeSelectingSleepMs:                          mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MergeSelectingSleepMs, diags),
		MaxMergeSelectingSleepMs:                       mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MaxMergeSelectingSleepMs, diags),
		MinAgeToForceMergeSeconds:                      mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MinAgeToForceMergeSeconds, diags),
		MinAgeToForceMergeOnPartitionOnly:              mdbcommon.ExpandBoolWrapper(ctx, mergeTree.MinAgeToForceMergeOnPartitionOnly, diags),
		MergeMaxBlockSize:                              mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MergeMaxBlockSize, diags),
		DeduplicateMergeProjectionMode:                 clickhouseConfig.ClickhouseConfig_MergeTree_DeduplicateMergeProjectionMode(*deduplicateMergeProjectionModeValue),
		LightweightMutationProjectionMode:              clickhouseConfig.ClickhouseConfig_MergeTree_LightweightMutationProjectionMode(*lightweightMutationProjectionModeValue),
		ReplicatedDeduplicationWindow:                  mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.ReplicatedDeduplicationWindow, diags),
		ReplicatedDeduplicationWindowSeconds:           mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.ReplicatedDeduplicationWindowSeconds, diags),
		FsyncAfterInsert:                               mdbcommon.ExpandBoolWrapper(ctx, mergeTree.FsyncAfterInsert, diags),
		FsyncPartDirectory:                             mdbcommon.ExpandBoolWrapper(ctx, mergeTree.FsyncPartDirectory, diags),
		MinCompressedBytesToFsyncAfterFetch:            mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MinCompressedBytesToFsyncAfterFetch, diags),
		MinCompressedBytesToFsyncAfterMerge:            mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MinCompressedBytesToFsyncAfterMerge, diags),
		MinRowsToFsyncAfterMerge:                       mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MinRowsToFsyncAfterMerge, diags),
		TtlOnlyDropParts:                               mdbcommon.ExpandBoolWrapper(ctx, mergeTree.TtlOnlyDropParts, diags),
		MergeWithTtlTimeout:                            mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MergeWithTtlTimeout, diags),
		MergeWithRecompressionTtlTimeout:               mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MergeWithRecompressionTtlTimeout, diags),
		MaxNumberOfMergesWithTtlInPool:                 mdbcommon.ExpandInt64Wrapper(ctx, mergeTree.MaxNumberOfMergesWithTtlInPool, diags),
		MaterializeTtlRecalculateOnly:                  mdbcommon.ExpandBoolWrapper(ctx, mergeTree.MaterializeTtlRecalculateOnly, diags),
		CheckSampleColumnIsCorrect:                     mdbcommon.ExpandBoolWrapper(ctx, mergeTree.CheckSampleColumnIsCorrect, diags),
	}
}
