package mdb_clickhouse_user

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func flattenPermissions(ctx context.Context, permissions []*clickhouse.Permission, diags *diag.Diagnostics) types.Set {
	if permissions == nil {
		return types.SetNull(permissionType)
	}

	var permissionValues []attr.Value

	for _, permission := range permissions {
		permissionValue, diag := types.ObjectValue(permissionType.AttrTypes, map[string]attr.Value{
			"database_name": types.StringValue(permission.DatabaseName),
		})

		permissionValues = append(permissionValues, permissionValue)
		diags.Append(diag...)
	}

	value, diag := types.SetValue(permissionType, permissionValues)
	diags.Append(diag...)

	return value
}

func flattenQuotas(ctx context.Context, quotas []*clickhouse.UserQuota, diags *diag.Diagnostics) types.Set {
	if quotas == nil {
		return types.SetNull(quotaType)
	}

	var quotaValues []attr.Value

	for _, quota := range quotas {
		quotaValue, diag := types.ObjectValue(quotaType.AttrTypes, map[string]attr.Value{
			"interval_duration": typesInt64FromWrapper(quota.IntervalDuration),
			"queries":           typesInt64FromWrapper(quota.Queries),
			"errors":            typesInt64FromWrapper(quota.Errors),
			"result_rows":       typesInt64FromWrapper(quota.ResultRows),
			"read_rows":         typesInt64FromWrapper(quota.ReadRows),
			"execution_time":    typesInt64FromWrapper(quota.ExecutionTime),
		})

		quotaValues = append(quotaValues, quotaValue)
		diags.Append(diag...)
	}

	value, diag := types.SetValue(quotaType, quotaValues)
	diags.Append(diag...)

	return value
}

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

func flattenSettings(ctx context.Context, settings *clickhouse.UserSettings, diags *diag.Diagnostics) types.Object {
	if settings == nil {
		return types.ObjectNull(settingsType)
	}

	joinAlgorithms := flattenJoinAlgorithmSettings(ctx, settings.JoinAlgorithm, diags)
	if diags.HasError() {
		return types.ObjectNull(settingsType)
	}

	obj, d := types.ObjectValueFrom(
		ctx, settingsType, Setting{
			Readonly:                            typesInt64FromWrapper(settings.Readonly),
			AllowDdl:                            typesBoolFromWrapper(settings.AllowDdl),
			AllowIntrospectionFunctions:         typesBoolFromWrapper(settings.AllowIntrospectionFunctions),
			ConnectTimeout:                      typesInt64FromWrapper(settings.ConnectTimeout),
			ConnectTimeoutWithFailover:          typesInt64FromWrapper(settings.ConnectTimeoutWithFailover),
			ReceiveTimeout:                      typesInt64FromWrapper(settings.ReceiveTimeout),
			SendTimeout:                         typesInt64FromWrapper(settings.SendTimeout),
			TimeoutBeforeCheckingExecutionSpeed: typesInt64FromWrapper(settings.TimeoutBeforeCheckingExecutionSpeed),
			InsertQuorum:                        typesInt64FromWrapper(settings.InsertQuorum),
			InsertQuorumTimeout:                 typesInt64FromWrapper(settings.InsertQuorumTimeout),
			InsertQuorumParallel:                typesBoolFromWrapper(settings.InsertQuorumParallel),
			InsertNullAsDefault:                 typesBoolFromWrapper(settings.InputFormatNullAsDefault),
			SelectSequentialConsistency:         typesBoolFromWrapper(settings.SelectSequentialConsistency),
			DeduplicateBlocksInDependentMaterializedViews: typesBoolFromWrapper(settings.DeduplicateBlocksInDependentMaterializedViews),
			ReplicationAlterPartitionsSync:                typesInt64FromWrapper(settings.ReplicationAlterPartitionsSync),
			MaxReplicaDelayForDistributedQueries:          typesInt64FromWrapper(settings.MaxReplicaDelayForDistributedQueries),
			FallbackToStaleReplicasForDistributedQueries:  typesBoolFromWrapper(settings.FallbackToStaleReplicasForDistributedQueries),
			DistributedProductMode:                        getDistributedProductModeName(settings.DistributedProductMode),
			DistributedAggregationMemoryEfficient:         typesBoolFromWrapper(settings.DistributedAggregationMemoryEfficient),
			DistributedDdlTaskTimeout:                     typesInt64FromWrapper(settings.DistributedDdlTaskTimeout),
			SkipUnavailableShards:                         typesBoolFromWrapper(settings.SkipUnavailableShards),
			CompileExpressions:                            typesBoolFromWrapper(settings.CompileExpressions),
			MinCountToCompileExpression:                   typesInt64FromWrapper(settings.MinCountToCompileExpression),
			MaxBlockSize:                                  typesInt64FromWrapper(settings.MaxBlockSize),
			MinInsertBlockSizeRows:                        typesInt64FromWrapper(settings.MinInsertBlockSizeRows),
			MinInsertBlockSizeBytes:                       typesInt64FromWrapper(settings.MinInsertBlockSizeBytes),
			MaxInsertBlockSize:                            typesInt64FromWrapper(settings.MaxInsertBlockSize),
			MinBytesToUseDirectIo:                         typesInt64FromWrapper(settings.MinBytesToUseDirectIo),
			UseUncompressedCache:                          typesBoolFromWrapper(settings.UseUncompressedCache),
			MergeTreeMaxRowsToUseCache:                    typesInt64FromWrapper(settings.MergeTreeMaxRowsToUseCache),
			MergeTreeMaxBytesToUseCache:                   typesInt64FromWrapper(settings.MergeTreeMaxBytesToUseCache),
			MergeTreeMinRowsForConcurrentRead:             typesInt64FromWrapper(settings.MergeTreeMinRowsForConcurrentRead),
			MergeTreeMinBytesForConcurrentRead:            typesInt64FromWrapper(settings.MergeTreeMinBytesForConcurrentRead),
			MaxBytesBeforeExternalGroupBy:                 typesInt64FromWrapper(settings.MaxBytesBeforeExternalGroupBy),
			MaxBytesBeforeExternalSort:                    typesInt64FromWrapper(settings.MaxBytesBeforeExternalSort),
			GroupByTwoLevelThreshold:                      typesInt64FromWrapper(settings.GroupByTwoLevelThreshold),
			GroupByTwoLevelThresholdBytes:                 typesInt64FromWrapper(settings.GroupByTwoLevelThresholdBytes),
			Priority:                                      typesInt64FromWrapper(settings.Priority),
			MaxThreads:                                    typesInt64FromWrapper(settings.MaxThreads),
			MaxMemoryUsage:                                typesInt64FromWrapper(settings.MaxMemoryUsage),
			MaxMemoryUsageForUser:                         typesInt64FromWrapper(settings.MaxMemoryUsageForUser),
			MaxNetworkBandwidth:                           typesInt64FromWrapper(settings.MaxNetworkBandwidth),
			MaxNetworkBandwidthForUser:                    typesInt64FromWrapper(settings.MaxNetworkBandwidthForUser),
			MaxPartitionsPerInsertBlock:                   typesInt64FromWrapper(settings.MaxPartitionsPerInsertBlock),
			MaxConcurrentQueriesForUser:                   typesInt64FromWrapper(settings.MaxConcurrentQueriesForUser),
			ForceIndexByDate:                              typesBoolFromWrapper(settings.ForceIndexByDate),
			ForcePrimaryKey:                               typesBoolFromWrapper(settings.ForcePrimaryKey),
			MaxRowsToRead:                                 typesInt64FromWrapper(settings.MaxRowsToRead),
			MaxBytesToRead:                                typesInt64FromWrapper(settings.MaxBytesToRead),
			ReadOverflowMode:                              getOverflowModeName(settings.ReadOverflowMode),
			MaxRowsToGroupBy:                              typesInt64FromWrapper(settings.MaxRowsToGroupBy),
			GroupByOverflowMode:                           getGroupByOverflowModeName(settings.GroupByOverflowMode),
			MaxRowsToSort:                                 typesInt64FromWrapper(settings.MaxRowsToSort),
			MaxBytesToSort:                                typesInt64FromWrapper(settings.MaxBytesToSort),
			SortOverflowMode:                              getOverflowModeName(settings.SortOverflowMode),
			MaxResultRows:                                 typesInt64FromWrapper(settings.MaxResultRows),
			MaxResultBytes:                                typesInt64FromWrapper(settings.MaxResultBytes),
			ResultOverflowMode:                            getOverflowModeName(settings.ResultOverflowMode),
			MaxRowsInDistinct:                             typesInt64FromWrapper(settings.MaxRowsInDistinct),
			MaxBytesInDistinct:                            typesInt64FromWrapper(settings.MaxBytesInDistinct),
			DistinctOverflowMode:                          getOverflowModeName(settings.DistinctOverflowMode),
			MaxRowsToTransfer:                             typesInt64FromWrapper(settings.MaxRowsToTransfer),
			MaxBytesToTransfer:                            typesInt64FromWrapper(settings.MaxBytesToTransfer),
			TransferOverflowMode:                          getOverflowModeName(settings.TransferOverflowMode),
			MaxExecutionTime:                              typesInt64FromWrapper(settings.MaxExecutionTime),
			TimeoutOverflowMode:                           getOverflowModeName(settings.TimeoutOverflowMode),
			MaxRowsInSet:                                  typesInt64FromWrapper(settings.MaxRowsInSet),
			MaxBytesInSet:                                 typesInt64FromWrapper(settings.MaxBytesInSet),
			SetOverflowMode:                               getOverflowModeName(settings.SetOverflowMode),
			MaxRowsInJoin:                                 typesInt64FromWrapper(settings.MaxRowsInJoin),
			MaxBytesInJoin:                                typesInt64FromWrapper(settings.MaxBytesInJoin),
			JoinOverflowMode:                              getOverflowModeName(settings.JoinOverflowMode),
			AnyJoinDistinctRightTableKeys:                 typesBoolFromWrapper(settings.AnyJoinDistinctRightTableKeys),
			MaxColumnsToRead:                              typesInt64FromWrapper(settings.MaxColumnsToRead),
			MaxTemporaryColumns:                           typesInt64FromWrapper(settings.MaxTemporaryColumns),
			MaxTemporaryNonConstColumns:                   typesInt64FromWrapper(settings.MaxTemporaryNonConstColumns),
			MaxQuerySize:                                  typesInt64FromWrapper(settings.MaxQuerySize),
			MaxAstDepth:                                   typesInt64FromWrapper(settings.MaxAstDepth),
			MaxAstElements:                                typesInt64FromWrapper(settings.MaxAstElements),
			MaxExpandedAstElements:                        typesInt64FromWrapper(settings.MaxExpandedAstElements),
			MinExecutionSpeed:                             typesInt64FromWrapper(settings.MinExecutionSpeed),
			MinExecutionSpeedBytes:                        typesInt64FromWrapper(settings.MinExecutionSpeedBytes),
			CountDistinctImplementation:                   getCountDistinctImplementationName(settings.CountDistinctImplementation),
			InputFormatValuesInterpretExpressions:         typesBoolFromWrapper(settings.InputFormatValuesInterpretExpressions),
			InputFormatDefaultsForOmittedFields:           typesBoolFromWrapper(settings.InputFormatDefaultsForOmittedFields),
			InputFormatNullAsDefault:                      typesBoolFromWrapper(settings.InputFormatNullAsDefault),
			DateTimeInputFormat:                           getDateTimeInputFormatName(settings.DateTimeInputFormat),
			InputFormatWithNamesUseHeader:                 typesBoolFromWrapper(settings.InputFormatWithNamesUseHeader),
			OutputFormatJsonQuote_64BitIntegers:           typesBoolFromWrapper(settings.OutputFormatJsonQuote_64BitIntegers),
			OutputFormatJsonQuoteDenormals:                typesBoolFromWrapper(settings.OutputFormatJsonQuoteDenormals),
			DateTimeOutputFormat:                          getDateTimeOutputFormatName(settings.DateTimeOutputFormat),
			LowCardinalityAllowInNativeFormat:             typesBoolFromWrapper(settings.LowCardinalityAllowInNativeFormat),
			AllowSuspiciousLowCardinalityTypes:            typesBoolFromWrapper(settings.AllowSuspiciousLowCardinalityTypes),
			EmptyResultForAggregationByEmptySet:           typesBoolFromWrapper(settings.EmptyResultForAggregationByEmptySet),
			HttpConnectionTimeout:                         typesInt64FromWrapper(settings.HttpConnectionTimeout),
			HttpReceiveTimeout:                            typesInt64FromWrapper(settings.HttpReceiveTimeout),
			HttpSendTimeout:                               typesInt64FromWrapper(settings.HttpSendTimeout),
			EnableHttpCompression:                         typesBoolFromWrapper(settings.EnableHttpCompression),
			SendProgressInHttpHeaders:                     typesBoolFromWrapper(settings.SendProgressInHttpHeaders),
			HttpHeadersProgressInterval:                   typesInt64FromWrapper(settings.HttpHeadersProgressInterval),
			AddHttpCorsHeader:                             typesBoolFromWrapper(settings.AddHttpCorsHeader),
			CancelHttpReadonlyQueriesOnClientClose:        typesBoolFromWrapper(settings.CancelHttpReadonlyQueriesOnClientClose),
			MaxHttpGetRedirects:                           typesInt64FromWrapper(settings.MaxHttpGetRedirects),
			JoinedSubqueryRequiresAlias:                   typesBoolFromWrapper(settings.JoinedSubqueryRequiresAlias),
			JoinUseNulls:                                  typesBoolFromWrapper(settings.JoinUseNulls),
			TransformNullIn:                               typesBoolFromWrapper(settings.TransformNullIn),
			QuotaMode:                                     getQuotaModeName(settings.QuotaMode),
			FlattenNested:                                 typesBoolFromWrapper(settings.FlattenNested),
			FormatRegexp:                                  typesNullableString(settings.FormatRegexp),
			FormatRegexpSkipUnmatched:                     typesBoolFromWrapper(settings.FormatRegexpSkipUnmatched),
			AsyncInsert:                                   typesBoolFromWrapper(settings.AsyncInsert),
			AsyncInsertThreads:                            typesInt64FromWrapper(settings.AsyncInsertThreads),
			WaitForAsyncInsert:                            typesBoolFromWrapper(settings.WaitForAsyncInsert),
			WaitForAsyncInsertTimeout:                     typesInt64FromWrapper(settings.WaitForAsyncInsertTimeout),
			AsyncInsertMaxDataSize:                        typesInt64FromWrapper(settings.AsyncInsertMaxDataSize),
			AsyncInsertBusyTimeout:                        typesInt64FromWrapper(settings.AsyncInsertBusyTimeout),
			AsyncInsertStaleTimeout:                       typesInt64FromWrapper(settings.AsyncInsertStaleTimeout),
			MemoryProfilerStep:                            typesInt64FromWrapper(settings.MemoryProfilerStep),
			MemoryProfilerSampleProbability:               typesFloat64FromWrapper(settings.MemoryProfilerSampleProbability),
			MaxFinalThreads:                               typesInt64FromWrapper(settings.MaxFinalThreads),
			InputFormatParallelParsing:                    typesBoolFromWrapper(settings.InputFormatParallelParsing),
			InputFormatImportNestedJson:                   typesBoolFromWrapper(settings.InputFormatImportNestedJson),
			LocalFilesystemReadMethod:                     getLocalFilesystemReadMethodName(settings.LocalFilesystemReadMethod),
			MaxReadBufferSize:                             typesInt64FromWrapper(settings.MaxReadBufferSize),
			InsertKeeperMaxRetries:                        typesInt64FromWrapper(settings.InsertKeeperMaxRetries),
			DoNotMergeAcrossPartitionsSelectFinal:         typesBoolFromWrapper(settings.DoNotMergeAcrossPartitionsSelectFinal),
			MaxTemporaryDataOnDiskSizeForUser:             typesInt64FromWrapper(settings.MaxTemporaryDataOnDiskSizeForUser),
			MaxTemporaryDataOnDiskSizeForQuery:            typesInt64FromWrapper(settings.MaxTemporaryDataOnDiskSizeForQuery),
			MaxParserDepth:                                typesInt64FromWrapper(settings.MaxParserDepth),
			RemoteFilesystemReadMethod:                    getRemoteFilesystemReadMethodName(settings.RemoteFilesystemReadMethod),
			MemoryOvercommitRatioDenominator:              typesInt64FromWrapper(settings.MemoryOvercommitRatioDenominator),
			MemoryOvercommitRatioDenominatorForUser:       typesInt64FromWrapper(settings.MemoryOvercommitRatioDenominatorForUser),
			MemoryUsageOvercommitMaxWaitMicroseconds:      typesInt64FromWrapper(settings.MemoryUsageOvercommitMaxWaitMicroseconds),
			LogQueryThreads:                               typesBoolFromWrapper(settings.LogQueryThreads),
			LogQueryViews:                                 typesBoolFromWrapper(settings.LogQueryViews),
			MaxInsertThreads:                              typesInt64FromWrapper(settings.MaxInsertThreads),
			UseHedgedRequests:                             typesBoolFromWrapper(settings.UseHedgedRequests),
			IdleConnectionTimeout:                         typesInt64FromWrapper(settings.IdleConnectionTimeout),
			HedgedConnectionTimeoutMs:                     typesInt64FromWrapper(settings.HedgedConnectionTimeoutMs),
			LoadBalancing:                                 getLoadBalancingName(settings.LoadBalancing),
			PreferLocalhostReplica:                        typesBoolFromWrapper(settings.PreferLocalhostReplica),
			JoinAlgorithm:                                 joinAlgorithms,
			// FormatRegexpEscapingRule:                      (settings.)),
			FormatAvroSchemaRegistryUrl:                   typesNullableString(settings.FormatAvroSchemaRegistryUrl),
			DataTypeDefaultNullable:                       typesBoolFromWrapper(settings.DataTypeDefaultNullable),
			HttpMaxFieldNameSize:                          typesInt64FromWrapper(settings.HttpMaxFieldNameSize),
			HttpMaxFieldValueSize:                         typesInt64FromWrapper(settings.HttpMaxFieldValueSize),
			AsyncInsertUseAdaptiveBusyTimeout:             typesBoolFromWrapper(settings.AsyncInsertUseAdaptiveBusyTimeout),
			LogQueriesProbability:                         typesFloat64FromWrapper(settings.LogQueriesProbability),
			LogProcessorsProfiles:                         typesBoolFromWrapper(settings.LogProcessorsProfiles),
			UseQueryCache:                                 typesBoolFromWrapper(settings.UseQueryCache),
			EnableReadsFromQueryCache:                     typesBoolFromWrapper(settings.EnableReadsFromQueryCache),
			EnableWritesToQueryCache:                      typesBoolFromWrapper(settings.EnableWritesToQueryCache),
			QueryCacheMinQueryRuns:                        typesInt64FromWrapper(settings.QueryCacheMinQueryRuns),
			QueryCacheMinQueryDuration:                    typesInt64FromWrapper(settings.QueryCacheMinQueryDuration),
			QueryCacheTtl:                                 typesInt64FromWrapper(settings.QueryCacheTtl),
			QueryCacheMaxEntries:                          typesInt64FromWrapper(settings.QueryCacheMaxEntries),
			QueryCacheMaxSizeInBytes:                      typesInt64FromWrapper(settings.QueryCacheMaxSizeInBytes),
			QueryCacheTag:                                 typesNullableString(settings.QueryCacheTag),
			QueryCacheShareBetweenUsers:                   typesBoolFromWrapper(settings.QueryCacheShareBetweenUsers),
			QueryCacheNondeterministicFunctionHandling:    getQueryCacheNondeterministicFunctionHandlingName(settings.QueryCacheNondeterministicFunctionHandling),
			QueryCacheSystemTableHandling:                 getQueryCacheSystemTableHandlingName(settings.QueryCacheSystemTableHandling),
			IgnoreMaterializedViewsWithDroppedTargetTable: typesBoolFromWrapper(settings.IgnoreMaterializedViewsWithDroppedTargetTable),
			EnableAnalyzer:                                typesBoolFromWrapper(settings.EnableAnalyzer),
			DistributedDdlOutputMode:                      getDistributedDdlOutputModeName(settings.DistributedDdlOutputMode),
			S3UseAdaptiveTimeouts:                         typesBoolFromWrapper(settings.S3UseAdaptiveTimeouts),
		},
	)
	log.Printf("[TRACE] mdb_clickhouse_user: flatten settings to state: %+v\n", obj)
	diags.Append(d...)
	return obj
}

func flattenConnectionManager(ctx context.Context, connectionManager *clickhouse.ConnectionManager, diags *diag.Diagnostics) types.Object {
	if connectionManager == nil {
		return types.ObjectNull(connectionManagerType)
	}

	obj, d := types.ObjectValueFrom(
		ctx, connectionManagerType, ConnectionManager{
			ConnectionId: typesNullableString(connectionManager.ConnectionId),
		},
	)

	log.Printf("[TRACE] mdb_clickhouse_user: flatten connection_manager to state: %+v\n", obj)
	diags.Append(d...)
	return obj
}

func typesInt64FromWrapper(value *wrapperspb.Int64Value) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(value.GetValue())
}

func typesBoolFromWrapper(value *wrapperspb.BoolValue) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(value.Value)
}

func typesFloat64FromWrapper(value *wrapperspb.DoubleValue) types.Float64 {
	if value == nil {
		return types.Float64Null()
	}
	return types.Float64Value(value.Value)
}

func typesNullableString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
