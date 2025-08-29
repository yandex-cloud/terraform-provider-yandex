package mdb_clickhouse_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var defaultObjectOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

func expandPermissionsFromState(ctx context.Context, permissionsSet types.Set, diags *diag.Diagnostics) []*clickhouse.Permission {
	if permissionsSet.IsNull() || permissionsSet.IsUnknown() {
		return nil
	}

	permissionsRes := make([]*clickhouse.Permission, 0, len(permissionsSet.Elements()))
	permissionsType := make([]Permission, 0, len(permissionsSet.Elements()))
	diag := permissionsSet.ElementsAs(ctx, &permissionsType, false)
	diags.Append(diag...)
	if diag.HasError() {
		return nil
	}

	for _, permission := range permissionsType {
		permissionsRes = append(permissionsRes, &clickhouse.Permission{
			DatabaseName: permission.DatabaseName.ValueString(),
		})
	}

	return permissionsRes
}

func expandQuotasFromState(ctx context.Context, quotasState types.Set, diags *diag.Diagnostics) []*clickhouse.UserQuota {
	if quotasState.IsNull() || quotasState.IsUnknown() {
		return nil
	}

	quotasRes := make([]*clickhouse.UserQuota, 0, len(quotasState.Elements()))
	quotasTypes := make([]Quota, 0, len(quotasState.Elements()))
	diag := quotasState.ElementsAs(ctx, &quotasTypes, false)
	diags.Append(diag...)
	if diag.HasError() {
		return nil
	}

	for _, quota := range quotasTypes {
		quotasRes = append(quotasRes, &clickhouse.UserQuota{
			IntervalDuration: wrapInt64(quota.IntervalDuration),
			Queries:          wrapInt64(quota.Queries),
			Errors:           wrapInt64(quota.Errors),
			ResultRows:       wrapInt64(quota.ResultRows),
			ReadRows:         wrapInt64(quota.ReadRows),
			ExecutionTime:    wrapInt64(quota.ExecutionTime),
		})
	}

	return quotasRes
}

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

func expandSettingsFromState(ctx context.Context, settingsState types.Object, diags *diag.Diagnostics) *clickhouse.UserSettings {
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
		Readonly:                            wrapInt64(settings.Readonly),
		AllowDdl:                            wrapBool(settings.AllowDdl),
		AllowIntrospectionFunctions:         wrapBool(settings.AllowIntrospectionFunctions),
		ConnectTimeout:                      wrapInt64(settings.ConnectTimeout),
		ConnectTimeoutWithFailover:          wrapInt64(settings.ConnectTimeoutWithFailover),
		ReceiveTimeout:                      wrapInt64(settings.ReceiveTimeout),
		SendTimeout:                         wrapInt64(settings.SendTimeout),
		TimeoutBeforeCheckingExecutionSpeed: wrapInt64(settings.TimeoutBeforeCheckingExecutionSpeed),
		InsertQuorum:                        wrapInt64(settings.InsertQuorum),
		InsertQuorumTimeout:                 wrapInt64(settings.InsertQuorumTimeout),
		InsertQuorumParallel:                wrapBool(settings.InsertQuorumParallel),
		InsertNullAsDefault:                 wrapBool(settings.InsertNullAsDefault),
		SelectSequentialConsistency:         wrapBool(settings.SelectSequentialConsistency),
		DeduplicateBlocksInDependentMaterializedViews: wrapBool(settings.DeduplicateBlocksInDependentMaterializedViews),
		ReplicationAlterPartitionsSync:                wrapInt64(settings.ReplicationAlterPartitionsSync),
		MaxReplicaDelayForDistributedQueries:          wrapInt64(settings.MaxReplicaDelayForDistributedQueries),
		FallbackToStaleReplicasForDistributedQueries:  wrapBool(settings.FallbackToStaleReplicasForDistributedQueries),
		DistributedProductMode:                        getDistributedProductModeValue(settings.DistributedProductMode),
		DistributedAggregationMemoryEfficient:         wrapBool(settings.DistributedAggregationMemoryEfficient),
		DistributedDdlTaskTimeout:                     wrapInt64(settings.DistributedDdlTaskTimeout),
		SkipUnavailableShards:                         wrapBool(settings.SkipUnavailableShards),
		CompileExpressions:                            wrapBool(settings.CompileExpressions),
		MinCountToCompileExpression:                   wrapInt64(settings.MinCountToCompileExpression),
		MaxBlockSize:                                  wrapInt64(settings.MaxBlockSize),
		MinInsertBlockSizeRows:                        wrapInt64(settings.MinInsertBlockSizeRows),
		MinInsertBlockSizeBytes:                       wrapInt64(settings.MinInsertBlockSizeBytes),
		MaxInsertBlockSize:                            wrapInt64(settings.MaxInsertBlockSize),
		MinBytesToUseDirectIo:                         wrapInt64(settings.MinBytesToUseDirectIo),
		UseUncompressedCache:                          wrapBool(settings.UseUncompressedCache),
		MergeTreeMaxRowsToUseCache:                    wrapInt64(settings.MergeTreeMaxRowsToUseCache),
		MergeTreeMaxBytesToUseCache:                   wrapInt64(settings.MergeTreeMaxBytesToUseCache),
		MergeTreeMinRowsForConcurrentRead:             wrapInt64(settings.MergeTreeMinRowsForConcurrentRead),
		MergeTreeMinBytesForConcurrentRead:            wrapInt64(settings.MergeTreeMinBytesForConcurrentRead),
		MaxBytesBeforeExternalGroupBy:                 wrapInt64(settings.MaxBytesBeforeExternalGroupBy),
		MaxBytesBeforeExternalSort:                    wrapInt64(settings.MaxBytesBeforeExternalSort),
		GroupByTwoLevelThreshold:                      wrapInt64(settings.GroupByTwoLevelThreshold),
		GroupByTwoLevelThresholdBytes:                 wrapInt64(settings.GroupByTwoLevelThresholdBytes),
		Priority:                                      wrapInt64(settings.Priority),
		MaxThreads:                                    wrapInt64(settings.MaxThreads),
		MaxMemoryUsage:                                wrapInt64(settings.MaxMemoryUsage),
		MaxMemoryUsageForUser:                         wrapInt64(settings.MaxMemoryUsageForUser),
		MaxNetworkBandwidth:                           wrapInt64(settings.MaxNetworkBandwidth),
		MaxNetworkBandwidthForUser:                    wrapInt64(settings.MaxNetworkBandwidthForUser),
		MaxPartitionsPerInsertBlock:                   wrapInt64(settings.MaxPartitionsPerInsertBlock),
		MaxConcurrentQueriesForUser:                   wrapInt64(settings.MaxConcurrentQueriesForUser),
		ForceIndexByDate:                              wrapBool(settings.ForceIndexByDate),
		ForcePrimaryKey:                               wrapBool(settings.ForcePrimaryKey),
		MaxRowsToRead:                                 wrapInt64(settings.MaxRowsToRead),
		MaxBytesToRead:                                wrapInt64(settings.MaxBytesToRead),
		ReadOverflowMode:                              getOverflowModeValue(settings.ReadOverflowMode),
		MaxRowsToGroupBy:                              wrapInt64(settings.MaxRowsToGroupBy),
		GroupByOverflowMode:                           getGroupByOverflowModeValue(settings.GroupByOverflowMode),
		MaxRowsToSort:                                 wrapInt64(settings.MaxRowsToSort),
		MaxBytesToSort:                                wrapInt64(settings.MaxBytesToSort),
		SortOverflowMode:                              getOverflowModeValue(settings.SortOverflowMode),
		MaxResultRows:                                 wrapInt64(settings.MaxResultRows),
		MaxResultBytes:                                wrapInt64(settings.MaxResultBytes),
		ResultOverflowMode:                            getOverflowModeValue(settings.ResultOverflowMode),
		MaxRowsInDistinct:                             wrapInt64(settings.MaxRowsInDistinct),
		MaxBytesInDistinct:                            wrapInt64(settings.MaxBytesInDistinct),
		DistinctOverflowMode:                          getOverflowModeValue(settings.DistinctOverflowMode),
		MaxRowsToTransfer:                             wrapInt64(settings.MaxRowsToTransfer),
		MaxBytesToTransfer:                            wrapInt64(settings.MaxBytesToTransfer),
		TransferOverflowMode:                          getOverflowModeValue(settings.TransferOverflowMode),
		MaxExecutionTime:                              wrapInt64(settings.MaxExecutionTime),
		TimeoutOverflowMode:                           getOverflowModeValue(settings.TimeoutOverflowMode),
		MaxRowsInSet:                                  wrapInt64(settings.MaxRowsInSet),
		MaxBytesInSet:                                 wrapInt64(settings.MaxBytesInSet),
		SetOverflowMode:                               getOverflowModeValue(settings.SetOverflowMode),
		MaxRowsInJoin:                                 wrapInt64(settings.MaxRowsInJoin),
		MaxBytesInJoin:                                wrapInt64(settings.MaxBytesInJoin),
		JoinOverflowMode:                              getOverflowModeValue(settings.JoinOverflowMode),
		JoinAlgorithm:                                 joinAlgorithms,
		AnyJoinDistinctRightTableKeys:                 wrapBool(settings.AnyJoinDistinctRightTableKeys),
		MaxColumnsToRead:                              wrapInt64(settings.MaxColumnsToRead),
		MaxTemporaryColumns:                           wrapInt64(settings.MaxTemporaryColumns),
		MaxTemporaryNonConstColumns:                   wrapInt64(settings.MaxTemporaryNonConstColumns),
		MaxQuerySize:                                  wrapInt64(settings.MaxQuerySize),
		MaxAstDepth:                                   wrapInt64(settings.MaxAstDepth),
		MaxAstElements:                                wrapInt64(settings.MaxAstElements),
		MaxExpandedAstElements:                        wrapInt64(settings.MaxExpandedAstElements),
		MinExecutionSpeed:                             wrapInt64(settings.MinExecutionSpeed),
		MinExecutionSpeedBytes:                        wrapInt64(settings.MinExecutionSpeedBytes),
		CountDistinctImplementation:                   getCountDistinctImplementationValue(settings.CountDistinctImplementation),
		InputFormatValuesInterpretExpressions:         wrapBool(settings.InputFormatValuesInterpretExpressions),
		InputFormatDefaultsForOmittedFields:           wrapBool(settings.InputFormatDefaultsForOmittedFields),
		InputFormatNullAsDefault:                      wrapBool(settings.InputFormatNullAsDefault),
		DateTimeInputFormat:                           getDateTimeInputFormatValue(settings.DateTimeInputFormat),
		InputFormatWithNamesUseHeader:                 wrapBool(settings.InputFormatWithNamesUseHeader),
		OutputFormatJsonQuote_64BitIntegers:           wrapBool(settings.OutputFormatJsonQuote_64BitIntegers),
		OutputFormatJsonQuoteDenormals:                wrapBool(settings.OutputFormatJsonQuoteDenormals),
		DateTimeOutputFormat:                          getDateTimeOutputFormatValue(settings.DateTimeOutputFormat),
		LowCardinalityAllowInNativeFormat:             wrapBool(settings.LowCardinalityAllowInNativeFormat),
		AllowSuspiciousLowCardinalityTypes:            wrapBool(settings.AllowSuspiciousLowCardinalityTypes),
		EmptyResultForAggregationByEmptySet:           wrapBool(settings.EmptyResultForAggregationByEmptySet),
		HttpConnectionTimeout:                         wrapInt64(settings.HttpConnectionTimeout),
		HttpReceiveTimeout:                            wrapInt64(settings.HttpReceiveTimeout),
		HttpSendTimeout:                               wrapInt64(settings.HttpSendTimeout),
		EnableHttpCompression:                         wrapBool(settings.EnableHttpCompression),
		SendProgressInHttpHeaders:                     wrapBool(settings.SendProgressInHttpHeaders),
		HttpHeadersProgressInterval:                   wrapInt64(settings.HttpHeadersProgressInterval),
		AddHttpCorsHeader:                             wrapBool(settings.AddHttpCorsHeader),
		CancelHttpReadonlyQueriesOnClientClose:        wrapBool(settings.CancelHttpReadonlyQueriesOnClientClose),
		MaxHttpGetRedirects:                           wrapInt64(settings.MaxHttpGetRedirects),
		JoinedSubqueryRequiresAlias:                   wrapBool(settings.JoinedSubqueryRequiresAlias),
		JoinUseNulls:                                  wrapBool(settings.JoinUseNulls),
		TransformNullIn:                               wrapBool(settings.TransformNullIn),
		QuotaMode:                                     getQuotaModeValue(settings.QuotaMode),
		FlattenNested:                                 wrapBool(settings.FlattenNested),
		FormatRegexp:                                  wrapString(settings.FormatRegexp),
		FormatRegexpSkipUnmatched:                     wrapBool(settings.FormatRegexpSkipUnmatched),
		AsyncInsert:                                   wrapBool(settings.AsyncInsert),
		AsyncInsertThreads:                            wrapInt64(settings.AsyncInsertThreads),
		WaitForAsyncInsert:                            wrapBool(settings.WaitForAsyncInsert),
		WaitForAsyncInsertTimeout:                     wrapInt64(settings.WaitForAsyncInsertTimeout),
		AsyncInsertMaxDataSize:                        wrapInt64(settings.AsyncInsertMaxDataSize),
		AsyncInsertBusyTimeout:                        wrapInt64(settings.AsyncInsertBusyTimeout),
		AsyncInsertStaleTimeout:                       wrapInt64(settings.AsyncInsertStaleTimeout),
		MemoryProfilerStep:                            wrapInt64(settings.MemoryProfilerStep),
		MemoryProfilerSampleProbability:               wrapDouble(settings.MemoryProfilerSampleProbability),
		MaxFinalThreads:                               wrapInt64(settings.MaxFinalThreads),
		InputFormatParallelParsing:                    wrapBool(settings.InputFormatParallelParsing),
		InputFormatImportNestedJson:                   wrapBool(settings.InputFormatImportNestedJson),
		LocalFilesystemReadMethod:                     getLocalFilesystemReadMethodValue(settings.LocalFilesystemReadMethod),
		MaxReadBufferSize:                             wrapInt64(settings.MaxReadBufferSize),
		InsertKeeperMaxRetries:                        wrapInt64(settings.InsertKeeperMaxRetries),
		MaxTemporaryDataOnDiskSizeForUser:             wrapInt64(settings.MaxTemporaryDataOnDiskSizeForUser),
		MaxTemporaryDataOnDiskSizeForQuery:            wrapInt64(settings.MaxTemporaryDataOnDiskSizeForQuery),
		MaxParserDepth:                                wrapInt64(settings.MaxParserDepth),
		RemoteFilesystemReadMethod:                    getRemoteFilesystemReadMethodValue(settings.RemoteFilesystemReadMethod),
		MemoryOvercommitRatioDenominator:              wrapInt64(settings.MemoryOvercommitRatioDenominator),
		MemoryOvercommitRatioDenominatorForUser:       wrapInt64(settings.MemoryOvercommitRatioDenominatorForUser),
		MemoryUsageOvercommitMaxWaitMicroseconds:      wrapInt64(settings.MemoryUsageOvercommitMaxWaitMicroseconds),
		LogQueryThreads:                               wrapBool(settings.LogQueryThreads),
		MaxInsertThreads:                              wrapInt64(settings.MaxInsertThreads),
		UseHedgedRequests:                             wrapBool(settings.UseHedgedRequests),
		IdleConnectionTimeout:                         wrapInt64(settings.IdleConnectionTimeout),
		HedgedConnectionTimeoutMs:                     wrapInt64(settings.HedgedConnectionTimeoutMs),
		LoadBalancing:                                 getLoadBalancingValue(settings.LoadBalancing),
		PreferLocalhostReplica:                        wrapBool(settings.PreferLocalhostReplica), // FormatRegexpEscapingRule:                      0,
		DistributedDdlOutputMode:                      getDistributedDdlOutputModeValue(settings.DistributedDdlOutputMode),
		FormatAvroSchemaRegistryUrl:                   wrapString(settings.FormatAvroSchemaRegistryUrl),
		DataTypeDefaultNullable:                       wrapBool(settings.DataTypeDefaultNullable),
		HttpMaxFieldNameSize:                          wrapInt64(settings.HttpMaxFieldNameSize),
		HttpMaxFieldValueSize:                         wrapInt64(settings.HttpMaxFieldValueSize),
		AsyncInsertUseAdaptiveBusyTimeout:             wrapBool(settings.AsyncInsertUseAdaptiveBusyTimeout),
		LogQueryViews:                                 wrapBool(settings.LogQueryViews),
		LogQueriesProbability:                         wrapDouble(settings.LogQueriesProbability),
		LogProcessorsProfiles:                         wrapBool(settings.LogProcessorsProfiles),
		UseQueryCache:                                 wrapBool(settings.UseQueryCache),
		EnableReadsFromQueryCache:                     wrapBool(settings.EnableReadsFromQueryCache),
		EnableWritesToQueryCache:                      wrapBool(settings.EnableWritesToQueryCache),
		QueryCacheMinQueryRuns:                        wrapInt64(settings.QueryCacheMinQueryRuns),
		QueryCacheMinQueryDuration:                    wrapInt64(settings.QueryCacheMinQueryDuration),
		QueryCacheTtl:                                 wrapInt64(settings.QueryCacheTtl),
		QueryCacheMaxEntries:                          wrapInt64(settings.QueryCacheMaxEntries),
		QueryCacheMaxSizeInBytes:                      wrapInt64(settings.QueryCacheMaxSizeInBytes),
		QueryCacheTag:                                 wrapString(settings.QueryCacheTag),
		QueryCacheShareBetweenUsers:                   wrapBool(settings.QueryCacheShareBetweenUsers),
		QueryCacheNondeterministicFunctionHandling:    getQueryCacheNondeterministicFunctionHandlingValue(settings.QueryCacheNondeterministicFunctionHandling),
		QueryCacheSystemTableHandling:                 getQueryCacheSystemTableHandlingValue(settings.QueryCacheSystemTableHandling),
		DoNotMergeAcrossPartitionsSelectFinal:         wrapBool(settings.DoNotMergeAcrossPartitionsSelectFinal),
		IgnoreMaterializedViewsWithDroppedTargetTable: wrapBool(settings.IgnoreMaterializedViewsWithDroppedTargetTable),
		EnableAnalyzer:                                wrapBool(settings.EnableAnalyzer),
		S3UseAdaptiveTimeouts:                         wrapBool(settings.S3UseAdaptiveTimeouts),
	}
}

func wrapInt64(value types.Int64) *wrapperspb.Int64Value {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return &wrapperspb.Int64Value{Value: value.ValueInt64()}
}

func wrapBool(value types.Bool) *wrapperspb.BoolValue {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return &wrapperspb.BoolValue{Value: value.ValueBool()}
}

func wrapDouble(value types.Float64) *wrapperspb.DoubleValue {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return &wrapperspb.DoubleValue{Value: value.ValueFloat64()}
}

func wrapString(value types.String) string {
	return value.ValueString()
}
