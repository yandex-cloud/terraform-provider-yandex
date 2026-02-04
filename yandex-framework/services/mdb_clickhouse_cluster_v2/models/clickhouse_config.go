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

type ClickhouseConfig struct {
	LogLevel                                  types.String  `tfsdk:"log_level"`
	BackgroundPoolSize                        types.Int64   `tfsdk:"background_pool_size"`
	BackgroundMergesMutationsConcurrencyRatio types.Int64   `tfsdk:"background_merges_mutations_concurrency_ratio"`
	BackgroundSchedulePoolSize                types.Int64   `tfsdk:"background_schedule_pool_size"`
	BackgroundFetchesPoolSize                 types.Int64   `tfsdk:"background_fetches_pool_size"`
	BackgroundMovePoolSize                    types.Int64   `tfsdk:"background_move_pool_size"`
	BackgroundDistributedSchedulePoolSize     types.Int64   `tfsdk:"background_distributed_schedule_pool_size"`
	BackgroundBufferFlushSchedulePoolSize     types.Int64   `tfsdk:"background_buffer_flush_schedule_pool_size"`
	BackgroundMessageBrokerSchedulePoolSize   types.Int64   `tfsdk:"background_message_broker_schedule_pool_size"`
	BackgroundCommonPoolSize                  types.Int64   `tfsdk:"background_common_pool_size"`
	DictionariesLazyLoad                      types.Bool    `tfsdk:"dictionaries_lazy_load"`
	QueryLogRetentionSize                     types.Int64   `tfsdk:"query_log_retention_size"`
	QueryLogRetentionTime                     types.Int64   `tfsdk:"query_log_retention_time"`
	QueryThreadLogEnabled                     types.Bool    `tfsdk:"query_thread_log_enabled"`
	QueryThreadLogRetentionSize               types.Int64   `tfsdk:"query_thread_log_retention_size"`
	QueryThreadLogRetentionTime               types.Int64   `tfsdk:"query_thread_log_retention_time"`
	PartLogRetentionSize                      types.Int64   `tfsdk:"part_log_retention_size"`
	PartLogRetentionTime                      types.Int64   `tfsdk:"part_log_retention_time"`
	MetricLogEnabled                          types.Bool    `tfsdk:"metric_log_enabled"`
	MetricLogRetentionSize                    types.Int64   `tfsdk:"metric_log_retention_size"`
	MetricLogRetentionTime                    types.Int64   `tfsdk:"metric_log_retention_time"`
	TraceLogEnabled                           types.Bool    `tfsdk:"trace_log_enabled"`
	TraceLogRetentionSize                     types.Int64   `tfsdk:"trace_log_retention_size"`
	TraceLogRetentionTime                     types.Int64   `tfsdk:"trace_log_retention_time"`
	TextLogEnabled                            types.Bool    `tfsdk:"text_log_enabled"`
	TextLogRetentionSize                      types.Int64   `tfsdk:"text_log_retention_size"`
	TextLogRetentionTime                      types.Int64   `tfsdk:"text_log_retention_time"`
	TextLogLevel                              types.String  `tfsdk:"text_log_level"`
	OpentelemetrySpanLogEnabled               types.Bool    `tfsdk:"opentelemetry_span_log_enabled"`
	OpentelemetrySpanLogRetentionSize         types.Int64   `tfsdk:"opentelemetry_span_log_retention_size"`
	OpentelemetrySpanLogRetentionTime         types.Int64   `tfsdk:"opentelemetry_span_log_retention_time"`
	QueryViewsLogEnabled                      types.Bool    `tfsdk:"query_views_log_enabled"`
	QueryViewsLogRetentionSize                types.Int64   `tfsdk:"query_views_log_retention_size"`
	QueryViewsLogRetentionTime                types.Int64   `tfsdk:"query_views_log_retention_time"`
	AsynchronousMetricLogEnabled              types.Bool    `tfsdk:"asynchronous_metric_log_enabled"`
	AsynchronousMetricLogRetentionSize        types.Int64   `tfsdk:"asynchronous_metric_log_retention_size"`
	AsynchronousMetricLogRetentionTime        types.Int64   `tfsdk:"asynchronous_metric_log_retention_time"`
	SessionLogEnabled                         types.Bool    `tfsdk:"session_log_enabled"`
	SessionLogRetentionSize                   types.Int64   `tfsdk:"session_log_retention_size"`
	SessionLogRetentionTime                   types.Int64   `tfsdk:"session_log_retention_time"`
	ZookeeperLogEnabled                       types.Bool    `tfsdk:"zookeeper_log_enabled"`
	ZookeeperLogRetentionSize                 types.Int64   `tfsdk:"zookeeper_log_retention_size"`
	ZookeeperLogRetentionTime                 types.Int64   `tfsdk:"zookeeper_log_retention_time"`
	AsynchronousInsertLogEnabled              types.Bool    `tfsdk:"asynchronous_insert_log_enabled"`
	AsynchronousInsertLogRetentionSize        types.Int64   `tfsdk:"asynchronous_insert_log_retention_size"`
	AsynchronousInsertLogRetentionTime        types.Int64   `tfsdk:"asynchronous_insert_log_retention_time"`
	ProcessorsProfileLogEnabled               types.Bool    `tfsdk:"processors_profile_log_enabled"`
	ProcessorsProfileLogRetentionSize         types.Int64   `tfsdk:"processors_profile_log_retention_size"`
	ProcessorsProfileLogRetentionTime         types.Int64   `tfsdk:"processors_profile_log_retention_time"`
	ErrorLogEnabled                           types.Bool    `tfsdk:"error_log_enabled"`
	ErrorLogRetentionSize                     types.Int64   `tfsdk:"error_log_retention_size"`
	ErrorLogRetentionTime                     types.Int64   `tfsdk:"error_log_retention_time"`
	QueryMetricLogEnabled                     types.Bool    `tfsdk:"query_metric_log_enabled"`
	QueryMetricLogRetentionSize               types.Int64   `tfsdk:"query_metric_log_retention_size"`
	QueryMetricLogRetentionTime               types.Int64   `tfsdk:"query_metric_log_retention_time"`
	AccessControlImprovements                 types.Object  `tfsdk:"access_control_improvements"`
	MaxConnections                            types.Int64   `tfsdk:"max_connections"`
	MaxConcurrentQueries                      types.Int64   `tfsdk:"max_concurrent_queries"`
	MaxTableSizeToDrop                        types.Int64   `tfsdk:"max_table_size_to_drop"`
	MaxPartitionSizeToDrop                    types.Int64   `tfsdk:"max_partition_size_to_drop"`
	KeepAliveTimeout                          types.Int64   `tfsdk:"keep_alive_timeout"`
	UncompressedCacheSize                     types.Int64   `tfsdk:"uncompressed_cache_size"`
	Timezone                                  types.String  `tfsdk:"timezone"`
	GeobaseEnabled                            types.Bool    `tfsdk:"geobase_enabled"`
	GeobaseUri                                types.String  `tfsdk:"geobase_uri"`
	DefaultDatabase                           types.String  `tfsdk:"default_database"`
	TotalMemoryProfilerStep                   types.Int64   `tfsdk:"total_memory_profiler_step"`
	TotalMemoryTrackerSampleProbability       types.Float64 `tfsdk:"total_memory_tracker_sample_probability"`
	AsyncInsertThreads                        types.Int64   `tfsdk:"async_insert_threads"`
	BackupThreads                             types.Int64   `tfsdk:"backup_threads"`
	RestoreThreads                            types.Int64   `tfsdk:"restore_threads"`
	MergeTree                                 types.Object  `tfsdk:"merge_tree"`
	Compression                               types.List    `tfsdk:"compression"`
	GraphiteRollup                            types.List    `tfsdk:"graphite_rollup"`
	Kafka                                     types.Object  `tfsdk:"kafka"`
	Rabbitmq                                  types.Object  `tfsdk:"rabbitmq"`
	QueryMaskingRules                         types.List    `tfsdk:"query_masking_rules"`
	QueryCache                                types.Object  `tfsdk:"query_cache"`
	JdbcBridge                                types.Object  `tfsdk:"jdbc_bridge"`
	MySqlProtocol                             types.Bool    `tfsdk:"mysql_protocol"`
	CustomMacros                              types.List    `tfsdk:"custom_macros"`
}

var ClickhouseConfigAttrTypes = map[string]attr.Type{
	"log_level":            types.StringType,
	"background_pool_size": types.Int64Type,
	"background_merges_mutations_concurrency_ratio": types.Int64Type,
	"background_schedule_pool_size":                 types.Int64Type,
	"background_fetches_pool_size":                  types.Int64Type,
	"background_move_pool_size":                     types.Int64Type,
	"background_distributed_schedule_pool_size":     types.Int64Type,
	"background_buffer_flush_schedule_pool_size":    types.Int64Type,
	"background_message_broker_schedule_pool_size":  types.Int64Type,
	"background_common_pool_size":                   types.Int64Type,
	"dictionaries_lazy_load":                        types.BoolType,
	"query_log_retention_size":                      types.Int64Type,
	"query_log_retention_time":                      types.Int64Type,
	"query_thread_log_enabled":                      types.BoolType,
	"query_thread_log_retention_size":               types.Int64Type,
	"query_thread_log_retention_time":               types.Int64Type,
	"part_log_retention_size":                       types.Int64Type,
	"part_log_retention_time":                       types.Int64Type,
	"metric_log_enabled":                            types.BoolType,
	"metric_log_retention_size":                     types.Int64Type,
	"metric_log_retention_time":                     types.Int64Type,
	"trace_log_enabled":                             types.BoolType,
	"trace_log_retention_size":                      types.Int64Type,
	"trace_log_retention_time":                      types.Int64Type,
	"text_log_enabled":                              types.BoolType,
	"text_log_retention_size":                       types.Int64Type,
	"text_log_retention_time":                       types.Int64Type,
	"text_log_level":                                types.StringType,
	"opentelemetry_span_log_enabled":                types.BoolType,
	"opentelemetry_span_log_retention_size":         types.Int64Type,
	"opentelemetry_span_log_retention_time":         types.Int64Type,
	"query_views_log_enabled":                       types.BoolType,
	"query_views_log_retention_size":                types.Int64Type,
	"query_views_log_retention_time":                types.Int64Type,
	"asynchronous_metric_log_enabled":               types.BoolType,
	"asynchronous_metric_log_retention_size":        types.Int64Type,
	"asynchronous_metric_log_retention_time":        types.Int64Type,
	"session_log_enabled":                           types.BoolType,
	"session_log_retention_size":                    types.Int64Type,
	"session_log_retention_time":                    types.Int64Type,
	"zookeeper_log_enabled":                         types.BoolType,
	"zookeeper_log_retention_size":                  types.Int64Type,
	"zookeeper_log_retention_time":                  types.Int64Type,
	"asynchronous_insert_log_enabled":               types.BoolType,
	"asynchronous_insert_log_retention_size":        types.Int64Type,
	"asynchronous_insert_log_retention_time":        types.Int64Type,
	"processors_profile_log_enabled":                types.BoolType,
	"processors_profile_log_retention_size":         types.Int64Type,
	"processors_profile_log_retention_time":         types.Int64Type,
	"error_log_enabled":                             types.BoolType,
	"error_log_retention_size":                      types.Int64Type,
	"error_log_retention_time":                      types.Int64Type,
	"query_metric_log_enabled":                      types.BoolType,
	"query_metric_log_retention_size":               types.Int64Type,
	"query_metric_log_retention_time":               types.Int64Type,
	"access_control_improvements":                   types.ObjectType{AttrTypes: AccessControlImprovementsAttrTypes},
	"max_connections":                               types.Int64Type,
	"max_concurrent_queries":                        types.Int64Type,
	"max_table_size_to_drop":                        types.Int64Type,
	"max_partition_size_to_drop":                    types.Int64Type,
	"keep_alive_timeout":                            types.Int64Type,
	"uncompressed_cache_size":                       types.Int64Type,
	"timezone":                                      types.StringType,
	"geobase_enabled":                               types.BoolType,
	"geobase_uri":                                   types.StringType,
	"default_database":                              types.StringType,
	"total_memory_profiler_step":                    types.Int64Type,
	"total_memory_tracker_sample_probability":       types.Float64Type,
	"async_insert_threads":                          types.Int64Type,
	"backup_threads":                                types.Int64Type,
	"restore_threads":                               types.Int64Type,
	"merge_tree":                                    types.ObjectType{AttrTypes: MergeTreeConfigAttrTypes},
	"compression":                                   types.ListType{ElemType: types.ObjectType{AttrTypes: CompressionAttrTypes}},
	"graphite_rollup":                               types.ListType{ElemType: types.ObjectType{AttrTypes: GraphiteRollupAttrTypes}},
	"kafka":                                         types.ObjectType{AttrTypes: KafkaAttrTypes},
	"rabbitmq":                                      types.ObjectType{AttrTypes: RabbitmqAttrTypes},
	"query_masking_rules":                           types.ListType{ElemType: types.ObjectType{AttrTypes: QueryMaskingRuleAttrTypes}},
	"query_cache":                                   types.ObjectType{AttrTypes: QueryCacheAttrTypes},
	"jdbc_bridge":                                   types.ObjectType{AttrTypes: JdbcBridgeAttrTypes},
	"mysql_protocol":                                types.BoolType,
	"custom_macros":                                 types.ListType{ElemType: types.ObjectType{AttrTypes: MacroAttrTypes}},
}

func FlattenClickHouseConfig(ctx context.Context, state *Cluster, config *clickhouseConfig.ClickhouseConfig, diags *diag.Diagnostics) types.Object {
	if config == nil {
		return types.ObjectNull(ClickhouseConfigAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, ClickhouseConfigAttrTypes, ClickhouseConfig{
			LogLevel:           types.StringValue(config.LogLevel.Enum().String()),
			BackgroundPoolSize: mdbcommon.FlattenInt64Wrapper(ctx, config.BackgroundPoolSize, diags),
			BackgroundMergesMutationsConcurrencyRatio: mdbcommon.FlattenInt64Wrapper(ctx, config.BackgroundMergesMutationsConcurrencyRatio, diags),
			BackgroundSchedulePoolSize:                mdbcommon.FlattenInt64Wrapper(ctx, config.BackgroundSchedulePoolSize, diags),
			BackgroundFetchesPoolSize:                 mdbcommon.FlattenInt64Wrapper(ctx, config.BackgroundFetchesPoolSize, diags),
			BackgroundMovePoolSize:                    mdbcommon.FlattenInt64Wrapper(ctx, config.BackgroundMovePoolSize, diags),
			BackgroundDistributedSchedulePoolSize:     mdbcommon.FlattenInt64Wrapper(ctx, config.BackgroundDistributedSchedulePoolSize, diags),
			BackgroundBufferFlushSchedulePoolSize:     mdbcommon.FlattenInt64Wrapper(ctx, config.BackgroundBufferFlushSchedulePoolSize, diags),
			BackgroundMessageBrokerSchedulePoolSize:   mdbcommon.FlattenInt64Wrapper(ctx, config.BackgroundMessageBrokerSchedulePoolSize, diags),
			BackgroundCommonPoolSize:                  mdbcommon.FlattenInt64Wrapper(ctx, config.BackgroundCommonPoolSize, diags),
			DictionariesLazyLoad:                      mdbcommon.FlattenBoolWrapper(ctx, config.DictionariesLazyLoad, diags),
			QueryLogRetentionSize:                     mdbcommon.FlattenInt64Wrapper(ctx, config.QueryLogRetentionSize, diags),
			QueryLogRetentionTime:                     mdbcommon.FlattenInt64Wrapper(ctx, config.QueryLogRetentionTime, diags),
			QueryThreadLogEnabled:                     mdbcommon.FlattenBoolWrapper(ctx, config.QueryThreadLogEnabled, diags),
			QueryThreadLogRetentionSize:               mdbcommon.FlattenInt64Wrapper(ctx, config.QueryThreadLogRetentionSize, diags),
			QueryThreadLogRetentionTime:               mdbcommon.FlattenInt64Wrapper(ctx, config.QueryThreadLogRetentionTime, diags),
			PartLogRetentionSize:                      mdbcommon.FlattenInt64Wrapper(ctx, config.PartLogRetentionSize, diags),
			PartLogRetentionTime:                      mdbcommon.FlattenInt64Wrapper(ctx, config.PartLogRetentionTime, diags),
			MetricLogEnabled:                          mdbcommon.FlattenBoolWrapper(ctx, config.MetricLogEnabled, diags),
			MetricLogRetentionSize:                    mdbcommon.FlattenInt64Wrapper(ctx, config.MetricLogRetentionSize, diags),
			MetricLogRetentionTime:                    mdbcommon.FlattenInt64Wrapper(ctx, config.MetricLogRetentionTime, diags),
			TraceLogEnabled:                           mdbcommon.FlattenBoolWrapper(ctx, config.TraceLogEnabled, diags),
			TraceLogRetentionSize:                     mdbcommon.FlattenInt64Wrapper(ctx, config.TraceLogRetentionSize, diags),
			TraceLogRetentionTime:                     mdbcommon.FlattenInt64Wrapper(ctx, config.TraceLogRetentionTime, diags),
			TextLogEnabled:                            mdbcommon.FlattenBoolWrapper(ctx, config.TextLogEnabled, diags),
			TextLogRetentionSize:                      mdbcommon.FlattenInt64Wrapper(ctx, config.TextLogRetentionSize, diags),
			TextLogRetentionTime:                      mdbcommon.FlattenInt64Wrapper(ctx, config.TextLogRetentionTime, diags),
			TextLogLevel:                              types.StringValue(config.TextLogLevel.Enum().String()),
			OpentelemetrySpanLogEnabled:               mdbcommon.FlattenBoolWrapper(ctx, config.OpentelemetrySpanLogEnabled, diags),
			OpentelemetrySpanLogRetentionSize:         mdbcommon.FlattenInt64Wrapper(ctx, config.OpentelemetrySpanLogRetentionSize, diags),
			OpentelemetrySpanLogRetentionTime:         mdbcommon.FlattenInt64Wrapper(ctx, config.OpentelemetrySpanLogRetentionTime, diags),
			QueryViewsLogEnabled:                      mdbcommon.FlattenBoolWrapper(ctx, config.QueryViewsLogEnabled, diags),
			QueryViewsLogRetentionSize:                mdbcommon.FlattenInt64Wrapper(ctx, config.QueryViewsLogRetentionSize, diags),
			QueryViewsLogRetentionTime:                mdbcommon.FlattenInt64Wrapper(ctx, config.QueryViewsLogRetentionTime, diags),
			AsynchronousMetricLogEnabled:              mdbcommon.FlattenBoolWrapper(ctx, config.AsynchronousMetricLogEnabled, diags),
			AsynchronousMetricLogRetentionSize:        mdbcommon.FlattenInt64Wrapper(ctx, config.AsynchronousMetricLogRetentionSize, diags),
			AsynchronousMetricLogRetentionTime:        mdbcommon.FlattenInt64Wrapper(ctx, config.AsynchronousMetricLogRetentionTime, diags),
			SessionLogEnabled:                         mdbcommon.FlattenBoolWrapper(ctx, config.SessionLogEnabled, diags),
			SessionLogRetentionSize:                   mdbcommon.FlattenInt64Wrapper(ctx, config.SessionLogRetentionSize, diags),
			SessionLogRetentionTime:                   mdbcommon.FlattenInt64Wrapper(ctx, config.SessionLogRetentionTime, diags),
			ZookeeperLogEnabled:                       mdbcommon.FlattenBoolWrapper(ctx, config.ZookeeperLogEnabled, diags),
			ZookeeperLogRetentionSize:                 mdbcommon.FlattenInt64Wrapper(ctx, config.ZookeeperLogRetentionSize, diags),
			ZookeeperLogRetentionTime:                 mdbcommon.FlattenInt64Wrapper(ctx, config.ZookeeperLogRetentionTime, diags),
			AsynchronousInsertLogEnabled:              mdbcommon.FlattenBoolWrapper(ctx, config.AsynchronousInsertLogEnabled, diags),
			AsynchronousInsertLogRetentionSize:        mdbcommon.FlattenInt64Wrapper(ctx, config.AsynchronousInsertLogRetentionSize, diags),
			AsynchronousInsertLogRetentionTime:        mdbcommon.FlattenInt64Wrapper(ctx, config.AsynchronousInsertLogRetentionTime, diags),
			ProcessorsProfileLogEnabled:               mdbcommon.FlattenBoolWrapper(ctx, config.ProcessorsProfileLogEnabled, diags),
			ProcessorsProfileLogRetentionSize:         mdbcommon.FlattenInt64Wrapper(ctx, config.ProcessorsProfileLogRetentionSize, diags),
			ProcessorsProfileLogRetentionTime:         mdbcommon.FlattenInt64Wrapper(ctx, config.ProcessorsProfileLogRetentionTime, diags),
			ErrorLogEnabled:                           mdbcommon.FlattenBoolWrapper(ctx, config.ErrorLogEnabled, diags),
			ErrorLogRetentionSize:                     mdbcommon.FlattenInt64Wrapper(ctx, config.ErrorLogRetentionSize, diags),
			ErrorLogRetentionTime:                     mdbcommon.FlattenInt64Wrapper(ctx, config.ErrorLogRetentionTime, diags),
			QueryMetricLogEnabled:                     mdbcommon.FlattenBoolWrapper(ctx, config.QueryMetricLogEnabled, diags),
			QueryMetricLogRetentionSize:               mdbcommon.FlattenInt64Wrapper(ctx, config.QueryMetricLogRetentionSize, diags),
			QueryMetricLogRetentionTime:               mdbcommon.FlattenInt64Wrapper(ctx, config.QueryMetricLogRetentionTime, diags),
			AccessControlImprovements:                 flattenAccessControlImprovements(ctx, config.AccessControlImprovements, diags),
			MaxConnections:                            mdbcommon.FlattenInt64Wrapper(ctx, config.MaxConnections, diags),
			MaxConcurrentQueries:                      mdbcommon.FlattenInt64Wrapper(ctx, config.MaxConcurrentQueries, diags),
			MaxTableSizeToDrop:                        mdbcommon.FlattenInt64Wrapper(ctx, config.MaxTableSizeToDrop, diags),
			MaxPartitionSizeToDrop:                    mdbcommon.FlattenInt64Wrapper(ctx, config.MaxPartitionSizeToDrop, diags),
			KeepAliveTimeout:                          mdbcommon.FlattenInt64Wrapper(ctx, config.KeepAliveTimeout, diags),
			UncompressedCacheSize:                     mdbcommon.FlattenInt64Wrapper(ctx, config.UncompressedCacheSize, diags),
			Timezone:                                  types.StringValue(config.Timezone),
			GeobaseEnabled:                            mdbcommon.FlattenBoolWrapper(ctx, config.GeobaseEnabled, diags),
			GeobaseUri:                                types.StringValue(config.GeobaseUri),
			DefaultDatabase:                           mdbcommon.FlattenStringWrapper(ctx, config.DefaultDatabase, diags),
			TotalMemoryProfilerStep:                   mdbcommon.FlattenInt64Wrapper(ctx, config.TotalMemoryProfilerStep, diags),
			AsyncInsertThreads:                        mdbcommon.FlattenInt64Wrapper(ctx, config.AsyncInsertThreads, diags),
			BackupThreads:                             mdbcommon.FlattenInt64Wrapper(ctx, config.BackupThreads, diags),
			RestoreThreads:                            mdbcommon.FlattenInt64Wrapper(ctx, config.RestoreThreads, diags),
			TotalMemoryTrackerSampleProbability:       types.Float64Value(config.TotalMemoryTrackerSampleProbability.Value),
			MergeTree:                                 flattenMergeTree(ctx, config.MergeTree, diags),
			Compression:                               flattenListCompression(ctx, config.Compression, diags),
			GraphiteRollup:                            flattenListGraphiteRollup(ctx, config.GraphiteRollup, diags),
			Kafka:                                     flattenKafka(ctx, state, config.Kafka, diags),
			Rabbitmq:                                  flattenRabbitmq(ctx, state, config.Rabbitmq, diags),
			QueryMaskingRules:                         flattenListQueryMaskingRule(ctx, config.QueryMaskingRules, diags),
			QueryCache:                                flattenQueryCache(ctx, config.QueryCache, diags),
			JdbcBridge:                                flattenJdbcBridge(ctx, config.JdbcBridge, diags),
			MySqlProtocol:                             mdbcommon.FlattenBoolWrapper(ctx, config.MysqlProtocol, diags),
			CustomMacros:                              flattenListMacro(ctx, config.CustomMacros, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func ExpandClickHouseConfig(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var config ClickhouseConfig
	diags.Append(c.As(ctx, &config, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	logLevelValue := utils.ExpandEnum("log_level", config.LogLevel.ValueString(), clickhouseConfig.ClickhouseConfig_LogLevel_value, diags)
	if diags.HasError() {
		return nil
	}

	textLogLevelValue := utils.ExpandEnum("text_log_level", config.TextLogLevel.ValueString(), clickhouseConfig.ClickhouseConfig_LogLevel_value, diags)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig{
		BackgroundPoolSize:                        mdbcommon.ExpandInt64Wrapper(ctx, config.BackgroundPoolSize, diags),
		BackgroundMergesMutationsConcurrencyRatio: mdbcommon.ExpandInt64Wrapper(ctx, config.BackgroundMergesMutationsConcurrencyRatio, diags),
		BackgroundSchedulePoolSize:                mdbcommon.ExpandInt64Wrapper(ctx, config.BackgroundSchedulePoolSize, diags),
		BackgroundFetchesPoolSize:                 mdbcommon.ExpandInt64Wrapper(ctx, config.BackgroundFetchesPoolSize, diags),
		BackgroundMovePoolSize:                    mdbcommon.ExpandInt64Wrapper(ctx, config.BackgroundMovePoolSize, diags),
		BackgroundDistributedSchedulePoolSize:     mdbcommon.ExpandInt64Wrapper(ctx, config.BackgroundDistributedSchedulePoolSize, diags),
		BackgroundBufferFlushSchedulePoolSize:     mdbcommon.ExpandInt64Wrapper(ctx, config.BackgroundBufferFlushSchedulePoolSize, diags),
		BackgroundMessageBrokerSchedulePoolSize:   mdbcommon.ExpandInt64Wrapper(ctx, config.BackgroundMessageBrokerSchedulePoolSize, diags),
		BackgroundCommonPoolSize:                  mdbcommon.ExpandInt64Wrapper(ctx, config.BackgroundCommonPoolSize, diags),
		DictionariesLazyLoad:                      mdbcommon.ExpandBoolWrapper(ctx, config.DictionariesLazyLoad, diags),
		LogLevel:                                  clickhouseConfig.ClickhouseConfig_LogLevel(*logLevelValue),
		QueryLogRetentionSize:                     mdbcommon.ExpandInt64Wrapper(ctx, config.QueryLogRetentionSize, diags),
		QueryLogRetentionTime:                     mdbcommon.ExpandInt64Wrapper(ctx, config.QueryLogRetentionTime, diags),
		QueryThreadLogEnabled:                     mdbcommon.ExpandBoolWrapper(ctx, config.QueryThreadLogEnabled, diags),
		QueryThreadLogRetentionSize:               mdbcommon.ExpandInt64Wrapper(ctx, config.QueryThreadLogRetentionSize, diags),
		QueryThreadLogRetentionTime:               mdbcommon.ExpandInt64Wrapper(ctx, config.QueryThreadLogRetentionTime, diags),
		PartLogRetentionSize:                      mdbcommon.ExpandInt64Wrapper(ctx, config.PartLogRetentionSize, diags),
		PartLogRetentionTime:                      mdbcommon.ExpandInt64Wrapper(ctx, config.PartLogRetentionTime, diags),
		MetricLogEnabled:                          mdbcommon.ExpandBoolWrapper(ctx, config.MetricLogEnabled, diags),
		MetricLogRetentionSize:                    mdbcommon.ExpandInt64Wrapper(ctx, config.MetricLogRetentionSize, diags),
		MetricLogRetentionTime:                    mdbcommon.ExpandInt64Wrapper(ctx, config.MetricLogRetentionTime, diags),
		TraceLogEnabled:                           mdbcommon.ExpandBoolWrapper(ctx, config.TraceLogEnabled, diags),
		TraceLogRetentionSize:                     mdbcommon.ExpandInt64Wrapper(ctx, config.TraceLogRetentionSize, diags),
		TraceLogRetentionTime:                     mdbcommon.ExpandInt64Wrapper(ctx, config.TraceLogRetentionTime, diags),
		TextLogEnabled:                            mdbcommon.ExpandBoolWrapper(ctx, config.TextLogEnabled, diags),
		TextLogRetentionSize:                      mdbcommon.ExpandInt64Wrapper(ctx, config.TextLogRetentionSize, diags),
		TextLogRetentionTime:                      mdbcommon.ExpandInt64Wrapper(ctx, config.TextLogRetentionTime, diags),
		TextLogLevel:                              clickhouseConfig.ClickhouseConfig_LogLevel(*textLogLevelValue),
		OpentelemetrySpanLogEnabled:               mdbcommon.ExpandBoolWrapper(ctx, config.OpentelemetrySpanLogEnabled, diags),
		OpentelemetrySpanLogRetentionSize:         mdbcommon.ExpandInt64Wrapper(ctx, config.OpentelemetrySpanLogRetentionSize, diags),
		OpentelemetrySpanLogRetentionTime:         mdbcommon.ExpandInt64Wrapper(ctx, config.OpentelemetrySpanLogRetentionTime, diags),
		QueryViewsLogEnabled:                      mdbcommon.ExpandBoolWrapper(ctx, config.QueryViewsLogEnabled, diags),
		QueryViewsLogRetentionSize:                mdbcommon.ExpandInt64Wrapper(ctx, config.QueryViewsLogRetentionSize, diags),
		QueryViewsLogRetentionTime:                mdbcommon.ExpandInt64Wrapper(ctx, config.QueryViewsLogRetentionTime, diags),
		AsynchronousMetricLogEnabled:              mdbcommon.ExpandBoolWrapper(ctx, config.AsynchronousMetricLogEnabled, diags),
		AsynchronousMetricLogRetentionSize:        mdbcommon.ExpandInt64Wrapper(ctx, config.AsynchronousMetricLogRetentionSize, diags),
		AsynchronousMetricLogRetentionTime:        mdbcommon.ExpandInt64Wrapper(ctx, config.AsynchronousMetricLogRetentionTime, diags),
		SessionLogEnabled:                         mdbcommon.ExpandBoolWrapper(ctx, config.SessionLogEnabled, diags),
		SessionLogRetentionSize:                   mdbcommon.ExpandInt64Wrapper(ctx, config.SessionLogRetentionSize, diags),
		SessionLogRetentionTime:                   mdbcommon.ExpandInt64Wrapper(ctx, config.SessionLogRetentionTime, diags),
		ZookeeperLogEnabled:                       mdbcommon.ExpandBoolWrapper(ctx, config.ZookeeperLogEnabled, diags),
		ZookeeperLogRetentionSize:                 mdbcommon.ExpandInt64Wrapper(ctx, config.ZookeeperLogRetentionSize, diags),
		ZookeeperLogRetentionTime:                 mdbcommon.ExpandInt64Wrapper(ctx, config.ZookeeperLogRetentionTime, diags),
		AsynchronousInsertLogEnabled:              mdbcommon.ExpandBoolWrapper(ctx, config.AsynchronousInsertLogEnabled, diags),
		AsynchronousInsertLogRetentionSize:        mdbcommon.ExpandInt64Wrapper(ctx, config.AsynchronousInsertLogRetentionSize, diags),
		AsynchronousInsertLogRetentionTime:        mdbcommon.ExpandInt64Wrapper(ctx, config.AsynchronousInsertLogRetentionTime, diags),
		ProcessorsProfileLogEnabled:               mdbcommon.ExpandBoolWrapper(ctx, config.ProcessorsProfileLogEnabled, diags),
		ProcessorsProfileLogRetentionSize:         mdbcommon.ExpandInt64Wrapper(ctx, config.ProcessorsProfileLogRetentionSize, diags),
		ProcessorsProfileLogRetentionTime:         mdbcommon.ExpandInt64Wrapper(ctx, config.ProcessorsProfileLogRetentionTime, diags),
		ErrorLogEnabled:                           mdbcommon.ExpandBoolWrapper(ctx, config.ErrorLogEnabled, diags),
		ErrorLogRetentionSize:                     mdbcommon.ExpandInt64Wrapper(ctx, config.ErrorLogRetentionSize, diags),
		ErrorLogRetentionTime:                     mdbcommon.ExpandInt64Wrapper(ctx, config.ErrorLogRetentionTime, diags),
		QueryMetricLogEnabled:                     mdbcommon.ExpandBoolWrapper(ctx, config.QueryMetricLogEnabled, diags),
		QueryMetricLogRetentionSize:               mdbcommon.ExpandInt64Wrapper(ctx, config.QueryMetricLogRetentionSize, diags),
		QueryMetricLogRetentionTime:               mdbcommon.ExpandInt64Wrapper(ctx, config.QueryMetricLogRetentionTime, diags),
		AccessControlImprovements:                 expandAccessControlImprovements(ctx, config.AccessControlImprovements, diags),
		MaxConnections:                            mdbcommon.ExpandInt64Wrapper(ctx, config.MaxConnections, diags),
		MaxConcurrentQueries:                      mdbcommon.ExpandInt64Wrapper(ctx, config.MaxConcurrentQueries, diags),
		MaxTableSizeToDrop:                        mdbcommon.ExpandInt64Wrapper(ctx, config.MaxTableSizeToDrop, diags),
		MaxPartitionSizeToDrop:                    mdbcommon.ExpandInt64Wrapper(ctx, config.MaxPartitionSizeToDrop, diags),
		KeepAliveTimeout:                          mdbcommon.ExpandInt64Wrapper(ctx, config.KeepAliveTimeout, diags),
		UncompressedCacheSize:                     mdbcommon.ExpandInt64Wrapper(ctx, config.UncompressedCacheSize, diags),
		Timezone:                                  config.Timezone.ValueString(),
		GeobaseEnabled:                            mdbcommon.ExpandBoolWrapper(ctx, config.GeobaseEnabled, diags),
		GeobaseUri:                                config.GeobaseUri.ValueString(),
		DefaultDatabase:                           mdbcommon.ExpandStringWrapper(ctx, config.DefaultDatabase, diags),
		TotalMemoryProfilerStep:                   mdbcommon.ExpandInt64Wrapper(ctx, config.TotalMemoryProfilerStep, diags),
		TotalMemoryTrackerSampleProbability:       mdbcommon.ExpandDoubleWrapper(ctx, config.TotalMemoryTrackerSampleProbability, diags),
		AsyncInsertThreads:                        mdbcommon.ExpandInt64Wrapper(ctx, config.AsyncInsertThreads, diags),
		BackupThreads:                             mdbcommon.ExpandInt64Wrapper(ctx, config.BackupThreads, diags),
		RestoreThreads:                            mdbcommon.ExpandInt64Wrapper(ctx, config.RestoreThreads, diags),
		MergeTree:                                 expandMergeTree(ctx, config.MergeTree, diags),
		Compression:                               expandListCompression(ctx, config.Compression, diags),
		GraphiteRollup:                            expandListGraphiteRollup(ctx, config.GraphiteRollup, diags),
		Kafka:                                     expandKafka(ctx, config.Kafka, diags),
		Rabbitmq:                                  expandRabbitmq(ctx, config.Rabbitmq, diags),
		QueryMaskingRules:                         expandListQueryMaskingRule(ctx, config.QueryMaskingRules, diags),
		QueryCache:                                expandQueryCache(ctx, config.QueryCache, diags),
		JdbcBridge:                                expandJdbcBridge(ctx, config.JdbcBridge, diags),
		MysqlProtocol:                             mdbcommon.ExpandBoolWrapper(ctx, config.MySqlProtocol, diags),
		CustomMacros:                              expandListMacro(ctx, config.CustomMacros, diags),
	}
}
