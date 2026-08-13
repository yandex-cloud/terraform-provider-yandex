package usersettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon"
)

var defaultObjectOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

func expandJoinAlgorithmsFromState(ctx context.Context, algoState types.Set, diags *diag.Diagnostics) []clickhouse.UserSettings_JoinAlgorithm {
	if algoState.IsNull() || algoState.IsUnknown() {
		return nil
	}

	if len(algoState.Elements()) == 0 {
		return []clickhouse.UserSettings_JoinAlgorithm{}
	}

	algoValues := make([]clickhouse.UserSettings_JoinAlgorithm, 0, len(algoState.Elements()))
	algoTypes := make([]string, 0, len(algoState.Elements()))
	diag := algoState.ElementsAs(ctx, &algoTypes, false)
	diags.Append(diag...)
	if diag.HasError() {
		return nil
	}

	for _, algoName := range algoTypes {
		algoValues = append(algoValues, getJoinAlgorithmValue(types.StringValue(algoName)))
	}

	return algoValues
}

func Expand(ctx context.Context, settingsState types.Object, diags *diag.Diagnostics) *clickhouse.UserSettings {
	if settingsState.IsNull() || settingsState.IsUnknown() {
		return nil
	}

	var settings Setting

	diags.Append(settingsState.As(ctx, &settings, defaultObjectOptions)...)
	if diags.HasError() {
		return nil
	}

	joinAlgorithms := expandJoinAlgorithmsFromState(ctx, settings.JoinAlgorithm, diags)
	if diags.HasError() {
		return nil
	}

	return &clickhouse.UserSettings{
		Readonly:                            chcommon.WrapInt64(settings.Readonly),
		AllowDdl:                            chcommon.WrapBool(settings.AllowDdl),
		AllowIntrospectionFunctions:         chcommon.WrapBool(settings.AllowIntrospectionFunctions),
		ConnectTimeout:                      chcommon.WrapInt64(settings.ConnectTimeout),
		ConnectTimeoutWithFailover:          chcommon.WrapInt64(settings.ConnectTimeoutWithFailover),
		ReceiveTimeout:                      chcommon.WrapInt64(settings.ReceiveTimeout),
		SendTimeout:                         chcommon.WrapInt64(settings.SendTimeout),
		TimeoutBeforeCheckingExecutionSpeed: chcommon.WrapInt64(settings.TimeoutBeforeCheckingExecutionSpeed),
		InsertQuorum:                        chcommon.WrapInt64(settings.InsertQuorum),
		InsertQuorumTimeout:                 chcommon.WrapInt64(settings.InsertQuorumTimeout),
		InsertQuorumParallel:                chcommon.WrapBool(settings.InsertQuorumParallel),
		InsertNullAsDefault:                 chcommon.WrapBool(settings.InsertNullAsDefault),
		SelectSequentialConsistency:         chcommon.WrapBool(settings.SelectSequentialConsistency),
		DeduplicateBlocksInDependentMaterializedViews: chcommon.WrapBool(settings.DeduplicateBlocksInDependentMaterializedViews),
		ReplicationAlterPartitionsSync:                chcommon.WrapInt64(settings.ReplicationAlterPartitionsSync),
		MaxReplicaDelayForDistributedQueries:          chcommon.WrapInt64(settings.MaxReplicaDelayForDistributedQueries),
		FallbackToStaleReplicasForDistributedQueries:  chcommon.WrapBool(settings.FallbackToStaleReplicasForDistributedQueries),
		DistributedProductMode:                        getDistributedProductModeValue(settings.DistributedProductMode),
		DistributedAggregationMemoryEfficient:         chcommon.WrapBool(settings.DistributedAggregationMemoryEfficient),
		DistributedDdlTaskTimeout:                     chcommon.WrapInt64(settings.DistributedDdlTaskTimeout),
		SkipUnavailableShards:                         chcommon.WrapBool(settings.SkipUnavailableShards),
		CompileExpressions:                            chcommon.WrapBool(settings.CompileExpressions),
		MinCountToCompileExpression:                   chcommon.WrapInt64(settings.MinCountToCompileExpression),
		MaxBlockSize:                                  chcommon.WrapInt64(settings.MaxBlockSize),
		MinInsertBlockSizeRows:                        chcommon.WrapInt64(settings.MinInsertBlockSizeRows),
		MinInsertBlockSizeBytes:                       chcommon.WrapInt64(settings.MinInsertBlockSizeBytes),
		MaxInsertBlockSize:                            chcommon.WrapInt64(settings.MaxInsertBlockSize),
		MinBytesToUseDirectIo:                         chcommon.WrapInt64(settings.MinBytesToUseDirectIo),
		UseUncompressedCache:                          chcommon.WrapBool(settings.UseUncompressedCache),
		MergeTreeMaxRowsToUseCache:                    chcommon.WrapInt64(settings.MergeTreeMaxRowsToUseCache),
		MergeTreeMaxBytesToUseCache:                   chcommon.WrapInt64(settings.MergeTreeMaxBytesToUseCache),
		MergeTreeMinRowsForConcurrentRead:             chcommon.WrapInt64(settings.MergeTreeMinRowsForConcurrentRead),
		MergeTreeMinBytesForConcurrentRead:            chcommon.WrapInt64(settings.MergeTreeMinBytesForConcurrentRead),
		MaxBytesBeforeExternalGroupBy:                 chcommon.WrapInt64(settings.MaxBytesBeforeExternalGroupBy),
		MaxBytesBeforeExternalSort:                    chcommon.WrapInt64(settings.MaxBytesBeforeExternalSort),
		GroupByTwoLevelThreshold:                      chcommon.WrapInt64(settings.GroupByTwoLevelThreshold),
		GroupByTwoLevelThresholdBytes:                 chcommon.WrapInt64(settings.GroupByTwoLevelThresholdBytes),
		Priority:                                      chcommon.WrapInt64(settings.Priority),
		MaxThreads:                                    chcommon.WrapInt64(settings.MaxThreads),
		MaxMemoryUsage:                                chcommon.WrapInt64(settings.MaxMemoryUsage),
		MaxMemoryUsageForUser:                         chcommon.WrapInt64(settings.MaxMemoryUsageForUser),
		MaxNetworkBandwidth:                           chcommon.WrapInt64(settings.MaxNetworkBandwidth),
		MaxNetworkBandwidthForUser:                    chcommon.WrapInt64(settings.MaxNetworkBandwidthForUser),
		MaxPartitionsPerInsertBlock:                   chcommon.WrapInt64(settings.MaxPartitionsPerInsertBlock),
		MaxConcurrentQueriesForUser:                   chcommon.WrapInt64(settings.MaxConcurrentQueriesForUser),
		ForceIndexByDate:                              chcommon.WrapBool(settings.ForceIndexByDate),
		ForcePrimaryKey:                               chcommon.WrapBool(settings.ForcePrimaryKey),
		MaxRowsToRead:                                 chcommon.WrapInt64(settings.MaxRowsToRead),
		MaxBytesToRead:                                chcommon.WrapInt64(settings.MaxBytesToRead),
		ReadOverflowMode:                              getOverflowModeValue(settings.ReadOverflowMode),
		MaxRowsToGroupBy:                              chcommon.WrapInt64(settings.MaxRowsToGroupBy),
		GroupByOverflowMode:                           getGroupByOverflowModeValue(settings.GroupByOverflowMode),
		MaxRowsToSort:                                 chcommon.WrapInt64(settings.MaxRowsToSort),
		MaxBytesToSort:                                chcommon.WrapInt64(settings.MaxBytesToSort),
		SortOverflowMode:                              getOverflowModeValue(settings.SortOverflowMode),
		MaxResultRows:                                 chcommon.WrapInt64(settings.MaxResultRows),
		MaxResultBytes:                                chcommon.WrapInt64(settings.MaxResultBytes),
		ResultOverflowMode:                            getOverflowModeValue(settings.ResultOverflowMode),
		MaxRowsInDistinct:                             chcommon.WrapInt64(settings.MaxRowsInDistinct),
		MaxBytesInDistinct:                            chcommon.WrapInt64(settings.MaxBytesInDistinct),
		DistinctOverflowMode:                          getOverflowModeValue(settings.DistinctOverflowMode),
		MaxRowsToTransfer:                             chcommon.WrapInt64(settings.MaxRowsToTransfer),
		MaxBytesToTransfer:                            chcommon.WrapInt64(settings.MaxBytesToTransfer),
		TransferOverflowMode:                          getOverflowModeValue(settings.TransferOverflowMode),
		MaxExecutionTime:                              chcommon.WrapInt64(settings.MaxExecutionTime),
		TimeoutOverflowMode:                           getOverflowModeValue(settings.TimeoutOverflowMode),
		MaxRowsInSet:                                  chcommon.WrapInt64(settings.MaxRowsInSet),
		MaxBytesInSet:                                 chcommon.WrapInt64(settings.MaxBytesInSet),
		SetOverflowMode:                               getOverflowModeValue(settings.SetOverflowMode),
		MaxRowsInJoin:                                 chcommon.WrapInt64(settings.MaxRowsInJoin),
		MaxBytesInJoin:                                chcommon.WrapInt64(settings.MaxBytesInJoin),
		JoinOverflowMode:                              getOverflowModeValue(settings.JoinOverflowMode),
		JoinAlgorithm:                                 joinAlgorithms,
		AnyJoinDistinctRightTableKeys:                 chcommon.WrapBool(settings.AnyJoinDistinctRightTableKeys),
		MaxColumnsToRead:                              chcommon.WrapInt64(settings.MaxColumnsToRead),
		MaxTemporaryColumns:                           chcommon.WrapInt64(settings.MaxTemporaryColumns),
		MaxTemporaryNonConstColumns:                   chcommon.WrapInt64(settings.MaxTemporaryNonConstColumns),
		MaxQuerySize:                                  chcommon.WrapInt64(settings.MaxQuerySize),
		MaxAstDepth:                                   chcommon.WrapInt64(settings.MaxAstDepth),
		MaxAstElements:                                chcommon.WrapInt64(settings.MaxAstElements),
		MaxExpandedAstElements:                        chcommon.WrapInt64(settings.MaxExpandedAstElements),
		MinExecutionSpeed:                             chcommon.WrapInt64(settings.MinExecutionSpeed),
		MinExecutionSpeedBytes:                        chcommon.WrapInt64(settings.MinExecutionSpeedBytes),
		CountDistinctImplementation:                   getCountDistinctImplementationValue(settings.CountDistinctImplementation),
		InputFormatValuesInterpretExpressions:         chcommon.WrapBool(settings.InputFormatValuesInterpretExpressions),
		InputFormatDefaultsForOmittedFields:           chcommon.WrapBool(settings.InputFormatDefaultsForOmittedFields),
		InputFormatNullAsDefault:                      chcommon.WrapBool(settings.InputFormatNullAsDefault),
		DateTimeInputFormat:                           getDateTimeInputFormatValue(settings.DateTimeInputFormat),
		InputFormatWithNamesUseHeader:                 chcommon.WrapBool(settings.InputFormatWithNamesUseHeader),
		OutputFormatJsonQuote_64BitIntegers:           chcommon.WrapBool(settings.OutputFormatJsonQuote_64BitIntegers),
		OutputFormatJsonQuoteDenormals:                chcommon.WrapBool(settings.OutputFormatJsonQuoteDenormals),
		DateTimeOutputFormat:                          getDateTimeOutputFormatValue(settings.DateTimeOutputFormat),
		LowCardinalityAllowInNativeFormat:             chcommon.WrapBool(settings.LowCardinalityAllowInNativeFormat),
		AllowSuspiciousLowCardinalityTypes:            chcommon.WrapBool(settings.AllowSuspiciousLowCardinalityTypes),
		EmptyResultForAggregationByEmptySet:           chcommon.WrapBool(settings.EmptyResultForAggregationByEmptySet),
		HttpConnectionTimeout:                         chcommon.WrapInt64(settings.HttpConnectionTimeout),
		HttpReceiveTimeout:                            chcommon.WrapInt64(settings.HttpReceiveTimeout),
		HttpSendTimeout:                               chcommon.WrapInt64(settings.HttpSendTimeout),
		EnableHttpCompression:                         chcommon.WrapBool(settings.EnableHttpCompression),
		SendProgressInHttpHeaders:                     chcommon.WrapBool(settings.SendProgressInHttpHeaders),
		HttpHeadersProgressInterval:                   chcommon.WrapInt64(settings.HttpHeadersProgressInterval),
		AddHttpCorsHeader:                             chcommon.WrapBool(settings.AddHttpCorsHeader),
		CancelHttpReadonlyQueriesOnClientClose:        chcommon.WrapBool(settings.CancelHttpReadonlyQueriesOnClientClose),
		MaxHttpGetRedirects:                           chcommon.WrapInt64(settings.MaxHttpGetRedirects),
		JoinedSubqueryRequiresAlias:                   chcommon.WrapBool(settings.JoinedSubqueryRequiresAlias),
		JoinUseNulls:                                  chcommon.WrapBool(settings.JoinUseNulls),
		TransformNullIn:                               chcommon.WrapBool(settings.TransformNullIn),
		QuotaMode:                                     getQuotaModeValue(settings.QuotaMode),
		FlattenNested:                                 chcommon.WrapBool(settings.FlattenNested),
		FormatRegexp:                                  chcommon.WrapString(settings.FormatRegexp),
		FormatRegexpSkipUnmatched:                     chcommon.WrapBool(settings.FormatRegexpSkipUnmatched),
		AsyncInsert:                                   chcommon.WrapBool(settings.AsyncInsert),
		AsyncInsertThreads:                            chcommon.WrapInt64(settings.AsyncInsertThreads),
		WaitForAsyncInsert:                            chcommon.WrapBool(settings.WaitForAsyncInsert),
		WaitForAsyncInsertTimeout:                     chcommon.WrapInt64(settings.WaitForAsyncInsertTimeout),
		AsyncInsertMaxDataSize:                        chcommon.WrapInt64(settings.AsyncInsertMaxDataSize),
		AsyncInsertBusyTimeout:                        chcommon.WrapInt64(settings.AsyncInsertBusyTimeout),
		AsyncInsertStaleTimeout:                       chcommon.WrapInt64(settings.AsyncInsertStaleTimeout),
		MemoryProfilerStep:                            chcommon.WrapInt64(settings.MemoryProfilerStep),
		MemoryProfilerSampleProbability:               chcommon.WrapDouble(settings.MemoryProfilerSampleProbability),
		MaxFinalThreads:                               chcommon.WrapInt64(settings.MaxFinalThreads),
		InputFormatParallelParsing:                    chcommon.WrapBool(settings.InputFormatParallelParsing),
		InputFormatImportNestedJson:                   chcommon.WrapBool(settings.InputFormatImportNestedJson),
		LocalFilesystemReadMethod:                     getLocalFilesystemReadMethodValue(settings.LocalFilesystemReadMethod),
		MaxReadBufferSize:                             chcommon.WrapInt64(settings.MaxReadBufferSize),
		InsertKeeperMaxRetries:                        chcommon.WrapInt64(settings.InsertKeeperMaxRetries),
		MaxTemporaryDataOnDiskSizeForUser:             chcommon.WrapInt64(settings.MaxTemporaryDataOnDiskSizeForUser),
		MaxTemporaryDataOnDiskSizeForQuery:            chcommon.WrapInt64(settings.MaxTemporaryDataOnDiskSizeForQuery),
		MaxParserDepth:                                chcommon.WrapInt64(settings.MaxParserDepth),
		RemoteFilesystemReadMethod:                    getRemoteFilesystemReadMethodValue(settings.RemoteFilesystemReadMethod),
		MemoryOvercommitRatioDenominator:              chcommon.WrapInt64(settings.MemoryOvercommitRatioDenominator),
		MemoryOvercommitRatioDenominatorForUser:       chcommon.WrapInt64(settings.MemoryOvercommitRatioDenominatorForUser),
		MemoryUsageOvercommitMaxWaitMicroseconds:      chcommon.WrapInt64(settings.MemoryUsageOvercommitMaxWaitMicroseconds),
		LogQueryThreads:                               chcommon.WrapBool(settings.LogQueryThreads),
		MaxInsertThreads:                              chcommon.WrapInt64(settings.MaxInsertThreads),
		UseHedgedRequests:                             chcommon.WrapBool(settings.UseHedgedRequests),
		IdleConnectionTimeout:                         chcommon.WrapInt64(settings.IdleConnectionTimeout),
		HedgedConnectionTimeoutMs:                     chcommon.WrapInt64(settings.HedgedConnectionTimeoutMs),
		LoadBalancing:                                 getLoadBalancingValue(settings.LoadBalancing),
		PreferLocalhostReplica:                        chcommon.WrapBool(settings.PreferLocalhostReplica), // FormatRegexpEscapingRule:                      0,
		DistributedDdlOutputMode:                      getDistributedDdlOutputModeValue(settings.DistributedDdlOutputMode),
		FormatAvroSchemaRegistryUrl:                   chcommon.WrapString(settings.FormatAvroSchemaRegistryUrl),
		DataTypeDefaultNullable:                       chcommon.WrapBool(settings.DataTypeDefaultNullable),
		HttpMaxFieldNameSize:                          chcommon.WrapInt64(settings.HttpMaxFieldNameSize),
		HttpMaxFieldValueSize:                         chcommon.WrapInt64(settings.HttpMaxFieldValueSize),
		AsyncInsertUseAdaptiveBusyTimeout:             chcommon.WrapBool(settings.AsyncInsertUseAdaptiveBusyTimeout),
		LogQueryViews:                                 chcommon.WrapBool(settings.LogQueryViews),
		LogQueriesProbability:                         chcommon.WrapDouble(settings.LogQueriesProbability),
		LogProcessorsProfiles:                         chcommon.WrapBool(settings.LogProcessorsProfiles),
		UseQueryCache:                                 chcommon.WrapBool(settings.UseQueryCache),
		EnableReadsFromQueryCache:                     chcommon.WrapBool(settings.EnableReadsFromQueryCache),
		EnableWritesToQueryCache:                      chcommon.WrapBool(settings.EnableWritesToQueryCache),
		QueryCacheMinQueryRuns:                        chcommon.WrapInt64(settings.QueryCacheMinQueryRuns),
		QueryCacheMinQueryDuration:                    chcommon.WrapInt64(settings.QueryCacheMinQueryDuration),
		QueryCacheTtl:                                 chcommon.WrapInt64(settings.QueryCacheTtl),
		QueryCacheMaxEntries:                          chcommon.WrapInt64(settings.QueryCacheMaxEntries),
		QueryCacheMaxSizeInBytes:                      chcommon.WrapInt64(settings.QueryCacheMaxSizeInBytes),
		QueryCacheTag:                                 chcommon.WrapString(settings.QueryCacheTag),
		QueryCacheShareBetweenUsers:                   chcommon.WrapBool(settings.QueryCacheShareBetweenUsers),
		QueryCacheNondeterministicFunctionHandling:    getQueryCacheNondeterministicFunctionHandlingValue(settings.QueryCacheNondeterministicFunctionHandling),
		QueryCacheSystemTableHandling:                 getQueryCacheSystemTableHandlingValue(settings.QueryCacheSystemTableHandling),
		DoNotMergeAcrossPartitionsSelectFinal:         chcommon.WrapBool(settings.DoNotMergeAcrossPartitionsSelectFinal),
		IgnoreMaterializedViewsWithDroppedTargetTable: chcommon.WrapBool(settings.IgnoreMaterializedViewsWithDroppedTargetTable),
		EnableAnalyzer:                                chcommon.WrapBool(settings.EnableAnalyzer),
		S3UseAdaptiveTimeouts:                         chcommon.WrapBool(settings.S3UseAdaptiveTimeouts),
	}
}
