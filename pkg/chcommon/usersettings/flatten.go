package usersettings

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon"
)

func flattenJoinAlgorithmSettings(ctx context.Context, algoEnumSlice []clickhouse.UserSettings_JoinAlgorithm, diags *diag.Diagnostics) types.Set {
	if algoEnumSlice == nil {
		return types.SetNull(types.StringType)
	}

	algoNamesSlice := make([]string, 0, len(algoEnumSlice))

	for _, algoEnum := range algoEnumSlice {
		algoNamesSlice = append(algoNamesSlice, getJoinAlgorithmName(algoEnum).ValueString())
	}

	value, diag := types.SetValueFrom(ctx, types.StringType, algoNamesSlice)
	diags.Append(diag...)

	return value

}

func Flatten(ctx context.Context, settings *clickhouse.UserSettings, diags *diag.Diagnostics) types.Object {
	if settings == nil {
		return types.ObjectNull(AttrTypes)
	}

	joinAlgorithms := flattenJoinAlgorithmSettings(ctx, settings.JoinAlgorithm, diags)
	if diags.HasError() {
		return types.ObjectNull(AttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, AttrTypes, Setting{
			Readonly:                            chcommon.Int64FromWrapper(settings.Readonly),
			AllowDdl:                            chcommon.BoolFromWrapper(settings.AllowDdl),
			AllowIntrospectionFunctions:         chcommon.BoolFromWrapper(settings.AllowIntrospectionFunctions),
			ConnectTimeout:                      chcommon.Int64FromWrapper(settings.ConnectTimeout),
			ConnectTimeoutWithFailover:          chcommon.Int64FromWrapper(settings.ConnectTimeoutWithFailover),
			ReceiveTimeout:                      chcommon.Int64FromWrapper(settings.ReceiveTimeout),
			SendTimeout:                         chcommon.Int64FromWrapper(settings.SendTimeout),
			TimeoutBeforeCheckingExecutionSpeed: chcommon.Int64FromWrapper(settings.TimeoutBeforeCheckingExecutionSpeed),
			InsertQuorum:                        chcommon.Int64FromWrapper(settings.InsertQuorum),
			InsertQuorumTimeout:                 chcommon.Int64FromWrapper(settings.InsertQuorumTimeout),
			InsertQuorumParallel:                chcommon.BoolFromWrapper(settings.InsertQuorumParallel),
			InsertNullAsDefault:                 chcommon.BoolFromWrapper(settings.InputFormatNullAsDefault),
			SelectSequentialConsistency:         chcommon.BoolFromWrapper(settings.SelectSequentialConsistency),
			DeduplicateBlocksInDependentMaterializedViews: chcommon.BoolFromWrapper(settings.DeduplicateBlocksInDependentMaterializedViews),
			ReplicationAlterPartitionsSync:                chcommon.Int64FromWrapper(settings.ReplicationAlterPartitionsSync),
			MaxReplicaDelayForDistributedQueries:          chcommon.Int64FromWrapper(settings.MaxReplicaDelayForDistributedQueries),
			FallbackToStaleReplicasForDistributedQueries:  chcommon.BoolFromWrapper(settings.FallbackToStaleReplicasForDistributedQueries),
			DistributedProductMode:                        getDistributedProductModeName(settings.DistributedProductMode),
			DistributedAggregationMemoryEfficient:         chcommon.BoolFromWrapper(settings.DistributedAggregationMemoryEfficient),
			DistributedDdlTaskTimeout:                     chcommon.Int64FromWrapper(settings.DistributedDdlTaskTimeout),
			SkipUnavailableShards:                         chcommon.BoolFromWrapper(settings.SkipUnavailableShards),
			CompileExpressions:                            chcommon.BoolFromWrapper(settings.CompileExpressions),
			MinCountToCompileExpression:                   chcommon.Int64FromWrapper(settings.MinCountToCompileExpression),
			MaxBlockSize:                                  chcommon.Int64FromWrapper(settings.MaxBlockSize),
			MinInsertBlockSizeRows:                        chcommon.Int64FromWrapper(settings.MinInsertBlockSizeRows),
			MinInsertBlockSizeBytes:                       chcommon.Int64FromWrapper(settings.MinInsertBlockSizeBytes),
			MaxInsertBlockSize:                            chcommon.Int64FromWrapper(settings.MaxInsertBlockSize),
			MinBytesToUseDirectIo:                         chcommon.Int64FromWrapper(settings.MinBytesToUseDirectIo),
			UseUncompressedCache:                          chcommon.BoolFromWrapper(settings.UseUncompressedCache),
			MergeTreeMaxRowsToUseCache:                    chcommon.Int64FromWrapper(settings.MergeTreeMaxRowsToUseCache),
			MergeTreeMaxBytesToUseCache:                   chcommon.Int64FromWrapper(settings.MergeTreeMaxBytesToUseCache),
			MergeTreeMinRowsForConcurrentRead:             chcommon.Int64FromWrapper(settings.MergeTreeMinRowsForConcurrentRead),
			MergeTreeMinBytesForConcurrentRead:            chcommon.Int64FromWrapper(settings.MergeTreeMinBytesForConcurrentRead),
			MaxBytesBeforeExternalGroupBy:                 chcommon.Int64FromWrapper(settings.MaxBytesBeforeExternalGroupBy),
			MaxBytesBeforeExternalSort:                    chcommon.Int64FromWrapper(settings.MaxBytesBeforeExternalSort),
			GroupByTwoLevelThreshold:                      chcommon.Int64FromWrapper(settings.GroupByTwoLevelThreshold),
			GroupByTwoLevelThresholdBytes:                 chcommon.Int64FromWrapper(settings.GroupByTwoLevelThresholdBytes),
			Priority:                                      chcommon.Int64FromWrapper(settings.Priority),
			MaxThreads:                                    chcommon.Int64FromWrapper(settings.MaxThreads),
			MaxMemoryUsage:                                chcommon.Int64FromWrapper(settings.MaxMemoryUsage),
			MaxMemoryUsageForUser:                         chcommon.Int64FromWrapper(settings.MaxMemoryUsageForUser),
			MaxNetworkBandwidth:                           chcommon.Int64FromWrapper(settings.MaxNetworkBandwidth),
			MaxNetworkBandwidthForUser:                    chcommon.Int64FromWrapper(settings.MaxNetworkBandwidthForUser),
			MaxPartitionsPerInsertBlock:                   chcommon.Int64FromWrapper(settings.MaxPartitionsPerInsertBlock),
			MaxConcurrentQueriesForUser:                   chcommon.Int64FromWrapper(settings.MaxConcurrentQueriesForUser),
			ForceIndexByDate:                              chcommon.BoolFromWrapper(settings.ForceIndexByDate),
			ForcePrimaryKey:                               chcommon.BoolFromWrapper(settings.ForcePrimaryKey),
			MaxRowsToRead:                                 chcommon.Int64FromWrapper(settings.MaxRowsToRead),
			MaxBytesToRead:                                chcommon.Int64FromWrapper(settings.MaxBytesToRead),
			ReadOverflowMode:                              getOverflowModeName(settings.ReadOverflowMode),
			MaxRowsToGroupBy:                              chcommon.Int64FromWrapper(settings.MaxRowsToGroupBy),
			GroupByOverflowMode:                           getGroupByOverflowModeName(settings.GroupByOverflowMode),
			MaxRowsToSort:                                 chcommon.Int64FromWrapper(settings.MaxRowsToSort),
			MaxBytesToSort:                                chcommon.Int64FromWrapper(settings.MaxBytesToSort),
			SortOverflowMode:                              getOverflowModeName(settings.SortOverflowMode),
			MaxResultRows:                                 chcommon.Int64FromWrapper(settings.MaxResultRows),
			MaxResultBytes:                                chcommon.Int64FromWrapper(settings.MaxResultBytes),
			ResultOverflowMode:                            getOverflowModeName(settings.ResultOverflowMode),
			MaxRowsInDistinct:                             chcommon.Int64FromWrapper(settings.MaxRowsInDistinct),
			MaxBytesInDistinct:                            chcommon.Int64FromWrapper(settings.MaxBytesInDistinct),
			DistinctOverflowMode:                          getOverflowModeName(settings.DistinctOverflowMode),
			MaxRowsToTransfer:                             chcommon.Int64FromWrapper(settings.MaxRowsToTransfer),
			MaxBytesToTransfer:                            chcommon.Int64FromWrapper(settings.MaxBytesToTransfer),
			TransferOverflowMode:                          getOverflowModeName(settings.TransferOverflowMode),
			MaxExecutionTime:                              chcommon.Int64FromWrapper(settings.MaxExecutionTime),
			TimeoutOverflowMode:                           getOverflowModeName(settings.TimeoutOverflowMode),
			MaxRowsInSet:                                  chcommon.Int64FromWrapper(settings.MaxRowsInSet),
			MaxBytesInSet:                                 chcommon.Int64FromWrapper(settings.MaxBytesInSet),
			SetOverflowMode:                               getOverflowModeName(settings.SetOverflowMode),
			MaxRowsInJoin:                                 chcommon.Int64FromWrapper(settings.MaxRowsInJoin),
			MaxBytesInJoin:                                chcommon.Int64FromWrapper(settings.MaxBytesInJoin),
			JoinOverflowMode:                              getOverflowModeName(settings.JoinOverflowMode),
			AnyJoinDistinctRightTableKeys:                 chcommon.BoolFromWrapper(settings.AnyJoinDistinctRightTableKeys),
			MaxColumnsToRead:                              chcommon.Int64FromWrapper(settings.MaxColumnsToRead),
			MaxTemporaryColumns:                           chcommon.Int64FromWrapper(settings.MaxTemporaryColumns),
			MaxTemporaryNonConstColumns:                   chcommon.Int64FromWrapper(settings.MaxTemporaryNonConstColumns),
			MaxQuerySize:                                  chcommon.Int64FromWrapper(settings.MaxQuerySize),
			MaxAstDepth:                                   chcommon.Int64FromWrapper(settings.MaxAstDepth),
			MaxAstElements:                                chcommon.Int64FromWrapper(settings.MaxAstElements),
			MaxExpandedAstElements:                        chcommon.Int64FromWrapper(settings.MaxExpandedAstElements),
			MinExecutionSpeed:                             chcommon.Int64FromWrapper(settings.MinExecutionSpeed),
			MinExecutionSpeedBytes:                        chcommon.Int64FromWrapper(settings.MinExecutionSpeedBytes),
			CountDistinctImplementation:                   getCountDistinctImplementationName(settings.CountDistinctImplementation),
			InputFormatValuesInterpretExpressions:         chcommon.BoolFromWrapper(settings.InputFormatValuesInterpretExpressions),
			InputFormatDefaultsForOmittedFields:           chcommon.BoolFromWrapper(settings.InputFormatDefaultsForOmittedFields),
			InputFormatNullAsDefault:                      chcommon.BoolFromWrapper(settings.InputFormatNullAsDefault),
			DateTimeInputFormat:                           getDateTimeInputFormatName(settings.DateTimeInputFormat),
			InputFormatWithNamesUseHeader:                 chcommon.BoolFromWrapper(settings.InputFormatWithNamesUseHeader),
			OutputFormatJsonQuote_64BitIntegers:           chcommon.BoolFromWrapper(settings.OutputFormatJsonQuote_64BitIntegers),
			OutputFormatJsonQuoteDenormals:                chcommon.BoolFromWrapper(settings.OutputFormatJsonQuoteDenormals),
			DateTimeOutputFormat:                          getDateTimeOutputFormatName(settings.DateTimeOutputFormat),
			LowCardinalityAllowInNativeFormat:             chcommon.BoolFromWrapper(settings.LowCardinalityAllowInNativeFormat),
			AllowSuspiciousLowCardinalityTypes:            chcommon.BoolFromWrapper(settings.AllowSuspiciousLowCardinalityTypes),
			EmptyResultForAggregationByEmptySet:           chcommon.BoolFromWrapper(settings.EmptyResultForAggregationByEmptySet),
			HttpConnectionTimeout:                         chcommon.Int64FromWrapper(settings.HttpConnectionTimeout),
			HttpReceiveTimeout:                            chcommon.Int64FromWrapper(settings.HttpReceiveTimeout),
			HttpSendTimeout:                               chcommon.Int64FromWrapper(settings.HttpSendTimeout),
			EnableHttpCompression:                         chcommon.BoolFromWrapper(settings.EnableHttpCompression),
			SendProgressInHttpHeaders:                     chcommon.BoolFromWrapper(settings.SendProgressInHttpHeaders),
			HttpHeadersProgressInterval:                   chcommon.Int64FromWrapper(settings.HttpHeadersProgressInterval),
			AddHttpCorsHeader:                             chcommon.BoolFromWrapper(settings.AddHttpCorsHeader),
			CancelHttpReadonlyQueriesOnClientClose:        chcommon.BoolFromWrapper(settings.CancelHttpReadonlyQueriesOnClientClose),
			MaxHttpGetRedirects:                           chcommon.Int64FromWrapper(settings.MaxHttpGetRedirects),
			JoinedSubqueryRequiresAlias:                   chcommon.BoolFromWrapper(settings.JoinedSubqueryRequiresAlias),
			JoinUseNulls:                                  chcommon.BoolFromWrapper(settings.JoinUseNulls),
			TransformNullIn:                               chcommon.BoolFromWrapper(settings.TransformNullIn),
			QuotaMode:                                     getQuotaModeName(settings.QuotaMode),
			FlattenNested:                                 chcommon.BoolFromWrapper(settings.FlattenNested),
			FormatRegexp:                                  chcommon.NullableString(settings.FormatRegexp),
			FormatRegexpSkipUnmatched:                     chcommon.BoolFromWrapper(settings.FormatRegexpSkipUnmatched),
			AsyncInsert:                                   chcommon.BoolFromWrapper(settings.AsyncInsert),
			AsyncInsertThreads:                            chcommon.Int64FromWrapper(settings.AsyncInsertThreads),
			WaitForAsyncInsert:                            chcommon.BoolFromWrapper(settings.WaitForAsyncInsert),
			WaitForAsyncInsertTimeout:                     chcommon.Int64FromWrapper(settings.WaitForAsyncInsertTimeout),
			AsyncInsertMaxDataSize:                        chcommon.Int64FromWrapper(settings.AsyncInsertMaxDataSize),
			AsyncInsertBusyTimeout:                        chcommon.Int64FromWrapper(settings.AsyncInsertBusyTimeout),
			AsyncInsertStaleTimeout:                       chcommon.Int64FromWrapper(settings.AsyncInsertStaleTimeout),
			MemoryProfilerStep:                            chcommon.Int64FromWrapper(settings.MemoryProfilerStep),
			MemoryProfilerSampleProbability:               chcommon.Float64FromWrapper(settings.MemoryProfilerSampleProbability),
			MaxFinalThreads:                               chcommon.Int64FromWrapper(settings.MaxFinalThreads),
			InputFormatParallelParsing:                    chcommon.BoolFromWrapper(settings.InputFormatParallelParsing),
			InputFormatImportNestedJson:                   chcommon.BoolFromWrapper(settings.InputFormatImportNestedJson),
			LocalFilesystemReadMethod:                     getLocalFilesystemReadMethodName(settings.LocalFilesystemReadMethod),
			MaxReadBufferSize:                             chcommon.Int64FromWrapper(settings.MaxReadBufferSize),
			InsertKeeperMaxRetries:                        chcommon.Int64FromWrapper(settings.InsertKeeperMaxRetries),
			DoNotMergeAcrossPartitionsSelectFinal:         chcommon.BoolFromWrapper(settings.DoNotMergeAcrossPartitionsSelectFinal),
			MaxTemporaryDataOnDiskSizeForUser:             chcommon.Int64FromWrapper(settings.MaxTemporaryDataOnDiskSizeForUser),
			MaxTemporaryDataOnDiskSizeForQuery:            chcommon.Int64FromWrapper(settings.MaxTemporaryDataOnDiskSizeForQuery),
			MaxParserDepth:                                chcommon.Int64FromWrapper(settings.MaxParserDepth),
			RemoteFilesystemReadMethod:                    getRemoteFilesystemReadMethodName(settings.RemoteFilesystemReadMethod),
			MemoryOvercommitRatioDenominator:              chcommon.Int64FromWrapper(settings.MemoryOvercommitRatioDenominator),
			MemoryOvercommitRatioDenominatorForUser:       chcommon.Int64FromWrapper(settings.MemoryOvercommitRatioDenominatorForUser),
			MemoryUsageOvercommitMaxWaitMicroseconds:      chcommon.Int64FromWrapper(settings.MemoryUsageOvercommitMaxWaitMicroseconds),
			LogQueryThreads:                               chcommon.BoolFromWrapper(settings.LogQueryThreads),
			LogQueryViews:                                 chcommon.BoolFromWrapper(settings.LogQueryViews),
			MaxInsertThreads:                              chcommon.Int64FromWrapper(settings.MaxInsertThreads),
			UseHedgedRequests:                             chcommon.BoolFromWrapper(settings.UseHedgedRequests),
			IdleConnectionTimeout:                         chcommon.Int64FromWrapper(settings.IdleConnectionTimeout),
			HedgedConnectionTimeoutMs:                     chcommon.Int64FromWrapper(settings.HedgedConnectionTimeoutMs),
			LoadBalancing:                                 getLoadBalancingName(settings.LoadBalancing),
			PreferLocalhostReplica:                        chcommon.BoolFromWrapper(settings.PreferLocalhostReplica),
			JoinAlgorithm:                                 joinAlgorithms,
			// FormatRegexpEscapingRule:                      (settings.)),
			FormatAvroSchemaRegistryUrl:                   chcommon.NullableString(settings.FormatAvroSchemaRegistryUrl),
			DataTypeDefaultNullable:                       chcommon.BoolFromWrapper(settings.DataTypeDefaultNullable),
			HttpMaxFieldNameSize:                          chcommon.Int64FromWrapper(settings.HttpMaxFieldNameSize),
			HttpMaxFieldValueSize:                         chcommon.Int64FromWrapper(settings.HttpMaxFieldValueSize),
			AsyncInsertUseAdaptiveBusyTimeout:             chcommon.BoolFromWrapper(settings.AsyncInsertUseAdaptiveBusyTimeout),
			LogQueriesProbability:                         chcommon.Float64FromWrapper(settings.LogQueriesProbability),
			LogProcessorsProfiles:                         chcommon.BoolFromWrapper(settings.LogProcessorsProfiles),
			UseQueryCache:                                 chcommon.BoolFromWrapper(settings.UseQueryCache),
			EnableReadsFromQueryCache:                     chcommon.BoolFromWrapper(settings.EnableReadsFromQueryCache),
			EnableWritesToQueryCache:                      chcommon.BoolFromWrapper(settings.EnableWritesToQueryCache),
			QueryCacheMinQueryRuns:                        chcommon.Int64FromWrapper(settings.QueryCacheMinQueryRuns),
			QueryCacheMinQueryDuration:                    chcommon.Int64FromWrapper(settings.QueryCacheMinQueryDuration),
			QueryCacheTtl:                                 chcommon.Int64FromWrapper(settings.QueryCacheTtl),
			QueryCacheMaxEntries:                          chcommon.Int64FromWrapper(settings.QueryCacheMaxEntries),
			QueryCacheMaxSizeInBytes:                      chcommon.Int64FromWrapper(settings.QueryCacheMaxSizeInBytes),
			QueryCacheTag:                                 chcommon.NullableString(settings.QueryCacheTag),
			QueryCacheShareBetweenUsers:                   chcommon.BoolFromWrapper(settings.QueryCacheShareBetweenUsers),
			QueryCacheNondeterministicFunctionHandling:    getQueryCacheNondeterministicFunctionHandlingName(settings.QueryCacheNondeterministicFunctionHandling),
			QueryCacheSystemTableHandling:                 getQueryCacheSystemTableHandlingName(settings.QueryCacheSystemTableHandling),
			IgnoreMaterializedViewsWithDroppedTargetTable: chcommon.BoolFromWrapper(settings.IgnoreMaterializedViewsWithDroppedTargetTable),
			EnableAnalyzer:                                chcommon.BoolFromWrapper(settings.EnableAnalyzer),
			DistributedDdlOutputMode:                      getDistributedDdlOutputModeName(settings.DistributedDdlOutputMode),
			S3UseAdaptiveTimeouts:                         chcommon.BoolFromWrapper(settings.S3UseAdaptiveTimeouts),
		},
	)
	log.Printf("[TRACE] usersettings: flatten settings to state: %+v\n", obj)
	diags.Append(d...)
	return obj
}
