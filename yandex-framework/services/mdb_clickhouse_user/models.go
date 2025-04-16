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

type User interface {
	SetId(id types.String)
	SetClusterId(clusterId types.String)
	SetName(name types.String)
	SetPassword(password types.String)
	SetPermissions(permissions types.Set)
	GetPermissions() types.Set
	SetSettings(settings types.Object)
	GetSettings() types.Object
	SetQuotas(quotas types.Set)
	GetQuotas() types.Set
	SetConnectionManager(connectionManager types.Object)
	GetConnectionManager() types.Object
}

type ResourceUser struct {
	Id                types.String `tfsdk:"id"`
	ClusterID         types.String `tfsdk:"cluster_id"`
	Name              types.String `tfsdk:"name"`
	Password          types.String `tfsdk:"password"`
	GeneratePassword  types.Bool   `tfsdk:"generate_password"`
	Permissions       types.Set    `tfsdk:"permission"`
	Settings          types.Object `tfsdk:"settings"`
	Quotas            types.Set    `tfsdk:"quota"`
	ConnectionManager types.Object `tfsdk:"connection_manager"`
}

func (ru *ResourceUser) SetId(id types.String) {
	ru.Id = id
}

func (ru *ResourceUser) SetClusterId(clusterId types.String) {
	ru.ClusterID = clusterId
}

func (ru *ResourceUser) SetName(name types.String) {
	ru.Name = name
}

func (ru *ResourceUser) SetPassword(password types.String) {
	ru.Password = password
}

func (ru *ResourceUser) SetPermissions(permissions types.Set) {
	ru.Permissions = permissions
}

func (ru *ResourceUser) GetPermissions() types.Set {
	return ru.Permissions
}

func (ru *ResourceUser) SetSettings(settings types.Object) {
	ru.Settings = settings
}

func (ru *ResourceUser) GetSettings() types.Object {
	return ru.Settings
}

func (ru *ResourceUser) SetQuotas(quotas types.Set) {
	ru.Quotas = quotas
}

func (ru *ResourceUser) GetQuotas() types.Set {
	return ru.Quotas
}

func (ru *ResourceUser) SetConnectionManager(connectionManager types.Object) {
	ru.ConnectionManager = connectionManager
}

func (ru *ResourceUser) GetConnectionManager() types.Object {
	return ru.ConnectionManager
}

type DatasourceUser struct {
	Id                types.String `tfsdk:"id"`
	ClusterID         types.String `tfsdk:"cluster_id"`
	Name              types.String `tfsdk:"name"`
	Password          types.String `tfsdk:"password"`
	Permissions       types.Set    `tfsdk:"permission"`
	Settings          types.Object `tfsdk:"settings"`
	Quotas            types.Set    `tfsdk:"quota"`
	ConnectionManager types.Object `tfsdk:"connection_manager"`
}

func (du *DatasourceUser) SetId(id types.String) {
	du.Id = id
}

func (du *DatasourceUser) SetClusterId(clusterId types.String) {
	du.ClusterID = clusterId
}

func (du *DatasourceUser) SetName(name types.String) {
	du.Name = name
}

func (du *DatasourceUser) SetPassword(password types.String) {
	du.Password = password
}

func (du *DatasourceUser) SetPermissions(permissions types.Set) {
	du.Permissions = permissions
}
func (du *DatasourceUser) GetPermissions() types.Set {
	return du.Permissions
}

func (du *DatasourceUser) SetSettings(settings types.Object) {
	du.Settings = settings
}

func (du *DatasourceUser) GetSettings() types.Object {
	return du.Settings
}

func (du *DatasourceUser) SetQuotas(quotas types.Set) {
	du.Quotas = quotas
}

func (du *DatasourceUser) GetQuotas() types.Set {
	return du.Quotas
}

func (du *DatasourceUser) SetConnectionManager(connectionManager types.Object) {
	du.ConnectionManager = connectionManager
}

func (du *DatasourceUser) GetConnectionManager() types.Object {
	return du.ConnectionManager
}

type Permission struct {
	DatabaseName types.String `tfsdk:"database_name"`
}

var permissionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"database_name": types.StringType,
	},
}

type Quota struct {
	IntervalDuration types.Int64 `tfsdk:"interval_duration"`
	Queries          types.Int64 `tfsdk:"queries"`
	Errors           types.Int64 `tfsdk:"errors"`
	ResultRows       types.Int64 `tfsdk:"result_rows"`
	ReadRows         types.Int64 `tfsdk:"read_rows"`
	ExecutionTime    types.Int64 `tfsdk:"execution_time"`
}

var quotaType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"interval_duration": types.Int64Type,
		"queries":           types.Int64Type,
		"errors":            types.Int64Type,
		"result_rows":       types.Int64Type,
		"read_rows":         types.Int64Type,
		"execution_time":    types.Int64Type,
	},
}

type Setting struct {
	Readonly                                      types.Int64   `tfsdk:"readonly"`
	AllowDdl                                      types.Bool    `tfsdk:"allow_ddl"`
	AllowIntrospectionFunctions                   types.Bool    `tfsdk:"allow_introspection_functions"`
	ConnectTimeout                                types.Int64   `tfsdk:"connect_timeout"`
	ConnectTimeoutWithFailover                    types.Int64   `tfsdk:"connect_timeout_with_failover"`
	ReceiveTimeout                                types.Int64   `tfsdk:"receive_timeout"`
	SendTimeout                                   types.Int64   `tfsdk:"send_timeout"`
	TimeoutBeforeCheckingExecutionSpeed           types.Int64   `tfsdk:"timeout_before_checking_execution_speed"`
	InsertQuorum                                  types.Int64   `tfsdk:"insert_quorum"`
	InsertQuorumTimeout                           types.Int64   `tfsdk:"insert_quorum_timeout"`
	InsertQuorumParallel                          types.Bool    `tfsdk:"insert_quorum_parallel"`
	InsertNullAsDefault                           types.Bool    `tfsdk:"insert_null_as_default"`
	SelectSequentialConsistency                   types.Bool    `tfsdk:"select_sequential_consistency"`
	DeduplicateBlocksInDependentMaterializedViews types.Bool    `tfsdk:"deduplicate_blocks_in_dependent_materialized_views"`
	ReplicationAlterPartitionsSync                types.Int64   `tfsdk:"replication_alter_partitions_sync"`
	MaxReplicaDelayForDistributedQueries          types.Int64   `tfsdk:"max_replica_delay_for_distributed_queries"`
	FallbackToStaleReplicasForDistributedQueries  types.Bool    `tfsdk:"fallback_to_stale_replicas_for_distributed_queries"`
	DistributedProductMode                        types.String  `tfsdk:"distributed_product_mode"`
	DistributedAggregationMemoryEfficient         types.Bool    `tfsdk:"distributed_aggregation_memory_efficient"`
	DistributedDdlTaskTimeout                     types.Int64   `tfsdk:"distributed_ddl_task_timeout"`
	SkipUnavailableShards                         types.Bool    `tfsdk:"skip_unavailable_shards"`
	CompileExpressions                            types.Bool    `tfsdk:"compile_expressions"`
	MinCountToCompileExpression                   types.Int64   `tfsdk:"min_count_to_compile_expression"`
	MaxBlockSize                                  types.Int64   `tfsdk:"max_block_size"`
	MinInsertBlockSizeRows                        types.Int64   `tfsdk:"min_insert_block_size_rows"`
	MinInsertBlockSizeBytes                       types.Int64   `tfsdk:"min_insert_block_size_bytes"`
	MaxInsertBlockSize                            types.Int64   `tfsdk:"max_insert_block_size"`
	MinBytesToUseDirectIo                         types.Int64   `tfsdk:"min_bytes_to_use_direct_io"`
	UseUncompressedCache                          types.Bool    `tfsdk:"use_uncompressed_cache"`
	MergeTreeMaxRowsToUseCache                    types.Int64   `tfsdk:"merge_tree_max_rows_to_use_cache"`
	MergeTreeMaxBytesToUseCache                   types.Int64   `tfsdk:"merge_tree_max_bytes_to_use_cache"`
	MergeTreeMinRowsForConcurrentRead             types.Int64   `tfsdk:"merge_tree_min_rows_for_concurrent_read"`
	MergeTreeMinBytesForConcurrentRead            types.Int64   `tfsdk:"merge_tree_min_bytes_for_concurrent_read"`
	MaxBytesBeforeExternalGroupBy                 types.Int64   `tfsdk:"max_bytes_before_external_group_by"`
	MaxBytesBeforeExternalSort                    types.Int64   `tfsdk:"max_bytes_before_external_sort"`
	GroupByTwoLevelThreshold                      types.Int64   `tfsdk:"group_by_two_level_threshold"`
	GroupByTwoLevelThresholdBytes                 types.Int64   `tfsdk:"group_by_two_level_threshold_bytes"`
	Priority                                      types.Int64   `tfsdk:"priority"`
	MaxThreads                                    types.Int64   `tfsdk:"max_threads"`
	MaxMemoryUsage                                types.Int64   `tfsdk:"max_memory_usage"`
	MaxMemoryUsageForUser                         types.Int64   `tfsdk:"max_memory_usage_for_user"`
	MaxNetworkBandwidth                           types.Int64   `tfsdk:"max_network_bandwidth"`
	MaxNetworkBandwidthForUser                    types.Int64   `tfsdk:"max_network_bandwidth_for_user"`
	MaxPartitionsPerInsertBlock                   types.Int64   `tfsdk:"max_partitions_per_insert_block"`
	MaxConcurrentQueriesForUser                   types.Int64   `tfsdk:"max_concurrent_queries_for_user"`
	ForceIndexByDate                              types.Bool    `tfsdk:"force_index_by_date"`
	ForcePrimaryKey                               types.Bool    `tfsdk:"force_primary_key"`
	MaxRowsToRead                                 types.Int64   `tfsdk:"max_rows_to_read"`
	MaxBytesToRead                                types.Int64   `tfsdk:"max_bytes_to_read"`
	ReadOverflowMode                              types.String  `tfsdk:"read_overflow_mode"`
	MaxRowsToGroupBy                              types.Int64   `tfsdk:"max_rows_to_group_by"`
	GroupByOverflowMode                           types.String  `tfsdk:"group_by_overflow_mode"`
	MaxRowsToSort                                 types.Int64   `tfsdk:"max_rows_to_sort"`
	MaxBytesToSort                                types.Int64   `tfsdk:"max_bytes_to_sort"`
	SortOverflowMode                              types.String  `tfsdk:"sort_overflow_mode"`
	MaxResultRows                                 types.Int64   `tfsdk:"max_result_rows"`
	MaxResultBytes                                types.Int64   `tfsdk:"max_result_bytes"`
	ResultOverflowMode                            types.String  `tfsdk:"result_overflow_mode"`
	MaxRowsInDistinct                             types.Int64   `tfsdk:"max_rows_in_distinct"`
	MaxBytesInDistinct                            types.Int64   `tfsdk:"max_bytes_in_distinct"`
	DistinctOverflowMode                          types.String  `tfsdk:"distinct_overflow_mode"`
	MaxRowsToTransfer                             types.Int64   `tfsdk:"max_rows_to_transfer"`
	MaxBytesToTransfer                            types.Int64   `tfsdk:"max_bytes_to_transfer"`
	TransferOverflowMode                          types.String  `tfsdk:"transfer_overflow_mode"`
	MaxExecutionTime                              types.Int64   `tfsdk:"max_execution_time"`
	TimeoutOverflowMode                           types.String  `tfsdk:"timeout_overflow_mode"`
	MaxRowsInSet                                  types.Int64   `tfsdk:"max_rows_in_set"`
	MaxBytesInSet                                 types.Int64   `tfsdk:"max_bytes_in_set"`
	SetOverflowMode                               types.String  `tfsdk:"set_overflow_mode"`
	MaxRowsInJoin                                 types.Int64   `tfsdk:"max_rows_in_join"`
	MaxBytesInJoin                                types.Int64   `tfsdk:"max_bytes_in_join"`
	JoinOverflowMode                              types.String  `tfsdk:"join_overflow_mode"`
	JoinAlgorithm                                 types.Set     `tfsdk:"join_algorithm"`
	AnyJoinDistinctRightTableKeys                 types.Bool    `tfsdk:"any_join_distinct_right_table_keys"`
	MaxColumnsToRead                              types.Int64   `tfsdk:"max_columns_to_read"`
	MaxTemporaryColumns                           types.Int64   `tfsdk:"max_temporary_columns"`
	MaxTemporaryNonConstColumns                   types.Int64   `tfsdk:"max_temporary_non_const_columns"`
	MaxQuerySize                                  types.Int64   `tfsdk:"max_query_size"`
	MaxAstDepth                                   types.Int64   `tfsdk:"max_ast_depth"`
	MaxAstElements                                types.Int64   `tfsdk:"max_ast_elements"`
	MaxExpandedAstElements                        types.Int64   `tfsdk:"max_expanded_ast_elements"`
	MinExecutionSpeed                             types.Int64   `tfsdk:"min_execution_speed"`
	MinExecutionSpeedBytes                        types.Int64   `tfsdk:"min_execution_speed_bytes"`
	CountDistinctImplementation                   types.String  `tfsdk:"count_distinct_implementation"`
	InputFormatValuesInterpretExpressions         types.Bool    `tfsdk:"input_format_values_interpret_expressions"`
	InputFormatDefaultsForOmittedFields           types.Bool    `tfsdk:"input_format_defaults_for_omitted_fields"`
	InputFormatNullAsDefault                      types.Bool    `tfsdk:"input_format_null_as_default"`
	DateTimeInputFormat                           types.String  `tfsdk:"date_time_input_format"`
	InputFormatWithNamesUseHeader                 types.Bool    `tfsdk:"input_format_with_names_use_header"`
	OutputFormatJsonQuote_64BitIntegers           types.Bool    `tfsdk:"output_format_json_quote_64bit_integers"`
	OutputFormatJsonQuoteDenormals                types.Bool    `tfsdk:"output_format_json_quote_denormals"`
	DateTimeOutputFormat                          types.String  `tfsdk:"date_time_output_format"`
	LowCardinalityAllowInNativeFormat             types.Bool    `tfsdk:"low_cardinality_allow_in_native_format"`
	AllowSuspiciousLowCardinalityTypes            types.Bool    `tfsdk:"allow_suspicious_low_cardinality_types"`
	EmptyResultForAggregationByEmptySet           types.Bool    `tfsdk:"empty_result_for_aggregation_by_empty_set"`
	HttpConnectionTimeout                         types.Int64   `tfsdk:"http_connection_timeout"`
	HttpReceiveTimeout                            types.Int64   `tfsdk:"http_receive_timeout"`
	HttpSendTimeout                               types.Int64   `tfsdk:"http_send_timeout"`
	EnableHttpCompression                         types.Bool    `tfsdk:"enable_http_compression"`
	SendProgressInHttpHeaders                     types.Bool    `tfsdk:"send_progress_in_http_headers"`
	HttpHeadersProgressInterval                   types.Int64   `tfsdk:"http_headers_progress_interval"`
	AddHttpCorsHeader                             types.Bool    `tfsdk:"add_http_cors_header"`
	CancelHttpReadonlyQueriesOnClientClose        types.Bool    `tfsdk:"cancel_http_readonly_queries_on_client_close"`
	MaxHttpGetRedirects                           types.Int64   `tfsdk:"max_http_get_redirects"`
	JoinedSubqueryRequiresAlias                   types.Bool    `tfsdk:"joined_subquery_requires_alias"`
	JoinUseNulls                                  types.Bool    `tfsdk:"join_use_nulls"`
	TransformNullIn                               types.Bool    `tfsdk:"transform_null_in"`
	QuotaMode                                     types.String  `tfsdk:"quota_mode"`
	FlattenNested                                 types.Bool    `tfsdk:"flatten_nested"`
	FormatRegexp                                  types.String  `tfsdk:"format_regexp"`
	FormatRegexpSkipUnmatched                     types.Bool    `tfsdk:"format_regexp_skip_unmatched"`
	AsyncInsert                                   types.Bool    `tfsdk:"async_insert"`
	AsyncInsertThreads                            types.Int64   `tfsdk:"async_insert_threads"`
	WaitForAsyncInsert                            types.Bool    `tfsdk:"wait_for_async_insert"`
	WaitForAsyncInsertTimeout                     types.Int64   `tfsdk:"wait_for_async_insert_timeout"`
	AsyncInsertMaxDataSize                        types.Int64   `tfsdk:"async_insert_max_data_size"`
	AsyncInsertBusyTimeout                        types.Int64   `tfsdk:"async_insert_busy_timeout"`
	AsyncInsertStaleTimeout                       types.Int64   `tfsdk:"async_insert_stale_timeout"`
	MemoryProfilerStep                            types.Int64   `tfsdk:"memory_profiler_step"`
	MemoryProfilerSampleProbability               types.Float64 `tfsdk:"memory_profiler_sample_probability"`
	MaxFinalThreads                               types.Int64   `tfsdk:"max_final_threads"`
	InputFormatParallelParsing                    types.Bool    `tfsdk:"input_format_parallel_parsing"`
	InputFormatImportNestedJson                   types.Bool    `tfsdk:"input_format_import_nested_json"`
	LocalFilesystemReadMethod                     types.String  `tfsdk:"local_filesystem_read_method"`
	MaxReadBufferSize                             types.Int64   `tfsdk:"max_read_buffer_size"`
	InsertKeeperMaxRetries                        types.Int64   `tfsdk:"insert_keeper_max_retries"`
	MaxTemporaryDataOnDiskSizeForUser             types.Int64   `tfsdk:"max_temporary_data_on_disk_size_for_user"`
	MaxTemporaryDataOnDiskSizeForQuery            types.Int64   `tfsdk:"max_temporary_data_on_disk_size_for_query"`
	MaxParserDepth                                types.Int64   `tfsdk:"max_parser_depth"`
	RemoteFilesystemReadMethod                    types.String  `tfsdk:"remote_filesystem_read_method"`
	MemoryOvercommitRatioDenominator              types.Int64   `tfsdk:"memory_overcommit_ratio_denominator"`
	MemoryOvercommitRatioDenominatorForUser       types.Int64   `tfsdk:"memory_overcommit_ratio_denominator_for_user"`
	MemoryUsageOvercommitMaxWaitMicroseconds      types.Int64   `tfsdk:"memory_usage_overcommit_max_wait_microseconds"`
	LogQueryThreads                               types.Bool    `tfsdk:"log_query_threads"`
	MaxInsertThreads                              types.Int64   `tfsdk:"max_insert_threads"`
	UseHedgedRequests                             types.Bool    `tfsdk:"use_hedged_requests"`
	IdleConnectionTimeout                         types.Int64   `tfsdk:"idle_connection_timeout"`
	HedgedConnectionTimeoutMs                     types.Int64   `tfsdk:"hedged_connection_timeout_ms"`
	LoadBalancing                                 types.String  `tfsdk:"load_balancing"`
	PreferLocalhostReplica                        types.Bool    `tfsdk:"prefer_localhost_replica"`
	// FormatRegexpEscapingRule                 types.String  `tfsdk:"format_regexp_escaping_rule"`
}

var settingsType = map[string]attr.Type{
	"readonly":                                types.Int64Type,
	"allow_ddl":                               types.BoolType,
	"connect_timeout":                         types.Int64Type,
	"distributed_product_mode":                types.StringType,
	"allow_introspection_functions":           types.BoolType,
	"connect_timeout_with_failover":           types.Int64Type,
	"receive_timeout":                         types.Int64Type,
	"send_timeout":                            types.Int64Type,
	"timeout_before_checking_execution_speed": types.Int64Type,
	"insert_quorum":                           types.Int64Type,
	"insert_quorum_timeout":                   types.Int64Type,
	"insert_quorum_parallel":                  types.BoolType,
	"insert_null_as_default":                  types.BoolType,
	"select_sequential_consistency":           types.BoolType,
	"deduplicate_blocks_in_dependent_materialized_views": types.BoolType,
	"replication_alter_partitions_sync":                  types.Int64Type,
	"max_replica_delay_for_distributed_queries":          types.Int64Type,
	"fallback_to_stale_replicas_for_distributed_queries": types.BoolType,
	"distributed_aggregation_memory_efficient":           types.BoolType,
	"distributed_ddl_task_timeout":                       types.Int64Type,
	"skip_unavailable_shards":                            types.BoolType,
	"compile_expressions":                                types.BoolType,
	"min_count_to_compile_expression":                    types.Int64Type,
	"max_block_size":                                     types.Int64Type,
	"min_insert_block_size_rows":                         types.Int64Type,
	"min_insert_block_size_bytes":                        types.Int64Type,
	"max_insert_block_size":                              types.Int64Type,
	"min_bytes_to_use_direct_io":                         types.Int64Type,
	"use_uncompressed_cache":                             types.BoolType,
	"merge_tree_max_rows_to_use_cache":                   types.Int64Type,
	"merge_tree_max_bytes_to_use_cache":                  types.Int64Type,
	"merge_tree_min_rows_for_concurrent_read":            types.Int64Type,
	"merge_tree_min_bytes_for_concurrent_read":           types.Int64Type,
	"max_bytes_before_external_group_by":                 types.Int64Type,
	"max_bytes_before_external_sort":                     types.Int64Type,
	"group_by_two_level_threshold":                       types.Int64Type,
	"group_by_two_level_threshold_bytes":                 types.Int64Type,
	"priority":                                           types.Int64Type,
	"max_threads":                                        types.Int64Type,
	"max_memory_usage":                                   types.Int64Type,
	"max_memory_usage_for_user":                          types.Int64Type,
	"max_network_bandwidth":                              types.Int64Type,
	"max_network_bandwidth_for_user":                     types.Int64Type,
	"max_partitions_per_insert_block":                    types.Int64Type,
	"max_concurrent_queries_for_user":                    types.Int64Type,
	"force_index_by_date":                                types.BoolType,
	"force_primary_key":                                  types.BoolType,
	"max_rows_to_read":                                   types.Int64Type,
	"max_bytes_to_read":                                  types.Int64Type,
	"read_overflow_mode":                                 types.StringType,
	"max_rows_to_group_by":                               types.Int64Type,
	"group_by_overflow_mode":                             types.StringType,
	"max_rows_to_sort":                                   types.Int64Type,
	"max_bytes_to_sort":                                  types.Int64Type,
	"sort_overflow_mode":                                 types.StringType,
	"max_result_rows":                                    types.Int64Type,
	"max_result_bytes":                                   types.Int64Type,
	"result_overflow_mode":                               types.StringType,
	"max_rows_in_distinct":                               types.Int64Type,
	"max_bytes_in_distinct":                              types.Int64Type,
	"distinct_overflow_mode":                             types.StringType,
	"max_rows_to_transfer":                               types.Int64Type,
	"max_bytes_to_transfer":                              types.Int64Type,
	"transfer_overflow_mode":                             types.StringType,
	"max_execution_time":                                 types.Int64Type,
	"timeout_overflow_mode":                              types.StringType,
	"max_rows_in_set":                                    types.Int64Type,
	"max_bytes_in_set":                                   types.Int64Type,
	"set_overflow_mode":                                  types.StringType,
	"max_rows_in_join":                                   types.Int64Type,
	"max_bytes_in_join":                                  types.Int64Type,
	"join_overflow_mode":                                 types.StringType,
	"join_algorithm":                                     types.SetType{ElemType: types.StringType},
	"any_join_distinct_right_table_keys":                 types.BoolType,
	"max_columns_to_read":                                types.Int64Type,
	"max_temporary_columns":                              types.Int64Type,
	"max_temporary_non_const_columns":                    types.Int64Type,
	"max_query_size":                                     types.Int64Type,
	"max_ast_depth":                                      types.Int64Type,
	"max_ast_elements":                                   types.Int64Type,
	"max_expanded_ast_elements":                          types.Int64Type,
	"min_execution_speed":                                types.Int64Type,
	"min_execution_speed_bytes":                          types.Int64Type,
	"count_distinct_implementation":                      types.StringType,
	"input_format_values_interpret_expressions":          types.BoolType,
	"input_format_defaults_for_omitted_fields":           types.BoolType,
	"input_format_null_as_default":                       types.BoolType,
	"date_time_input_format":                             types.StringType,
	"input_format_with_names_use_header":                 types.BoolType,
	"output_format_json_quote_64bit_integers":            types.BoolType,
	"output_format_json_quote_denormals":                 types.BoolType,
	"date_time_output_format":                            types.StringType,
	"low_cardinality_allow_in_native_format":             types.BoolType,
	"allow_suspicious_low_cardinality_types":             types.BoolType,
	"empty_result_for_aggregation_by_empty_set":          types.BoolType,
	"http_connection_timeout":                            types.Int64Type,
	"http_receive_timeout":                               types.Int64Type,
	"http_send_timeout":                                  types.Int64Type,
	"enable_http_compression":                            types.BoolType,
	"send_progress_in_http_headers":                      types.BoolType,
	"http_headers_progress_interval":                     types.Int64Type,
	"add_http_cors_header":                               types.BoolType,
	"cancel_http_readonly_queries_on_client_close":       types.BoolType,
	"max_http_get_redirects":                             types.Int64Type,
	"joined_subquery_requires_alias":                     types.BoolType,
	"join_use_nulls":                                     types.BoolType,
	"transform_null_in":                                  types.BoolType,
	"quota_mode":                                         types.StringType,
	"flatten_nested":                                     types.BoolType,
	"format_regexp":                                      types.StringType,
	"format_regexp_skip_unmatched":                       types.BoolType,
	"async_insert":                                       types.BoolType,
	"async_insert_threads":                               types.Int64Type,
	"wait_for_async_insert":                              types.BoolType,
	"wait_for_async_insert_timeout":                      types.Int64Type,
	"async_insert_max_data_size":                         types.Int64Type,
	"async_insert_busy_timeout":                          types.Int64Type,
	"async_insert_stale_timeout":                         types.Int64Type,
	"memory_profiler_step":                               types.Int64Type,
	"memory_profiler_sample_probability":                 types.Float64Type,
	"max_final_threads":                                  types.Int64Type,
	"input_format_parallel_parsing":                      types.BoolType,
	"input_format_import_nested_json":                    types.BoolType,
	"local_filesystem_read_method":                       types.StringType,
	"max_read_buffer_size":                               types.Int64Type,
	"insert_keeper_max_retries":                          types.Int64Type,
	"max_temporary_data_on_disk_size_for_user":           types.Int64Type,
	"max_temporary_data_on_disk_size_for_query":          types.Int64Type,
	"max_parser_depth":                                   types.Int64Type,
	"remote_filesystem_read_method":                      types.StringType,
	"memory_overcommit_ratio_denominator":                types.Int64Type,
	"memory_overcommit_ratio_denominator_for_user":       types.Int64Type,
	"memory_usage_overcommit_max_wait_microseconds":      types.Int64Type,
	"log_query_threads":                                  types.BoolType,
	"max_insert_threads":                                 types.Int64Type,
	"use_hedged_requests":                                types.BoolType,
	"idle_connection_timeout":                            types.Int64Type,
	"hedged_connection_timeout_ms":                       types.Int64Type,
	"load_balancing":                                     types.StringType,
	"prefer_localhost_replica":                           types.BoolType,
	// "format_regexp_escaping_rule":                   types.StringType,
}

type ConnectionManager struct {
	ConnectionId types.String `tfsdk:"connection_id"`
}

var connectionManagerType = map[string]attr.Type{
	"connection_id": types.StringType,
}

func userToState(ctx context.Context, user *clickhouse.User, state User) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[TRACE] mdb_clickhouse_user: flatten state from user: %+v\n", user)
	state.SetName(types.StringValue(user.Name))
	state.SetClusterId(types.StringValue(user.ClusterId))

	state.SetPermissions(flattenPermissions(ctx, user.Permissions, &diags))
	log.Printf("[TRACE] mdb_clickhouse_user: flattened permissions: %+v\n", state.GetPermissions())
	state.SetQuotas(flattenQuotas(ctx, user.Quotas, &diags))
	log.Printf("[TRACE] mdb_clickhouse_user: flattened quotas: %+v\n", state.GetQuotas())
	state.SetSettings(flattenSettings(ctx, user.Settings, &diags))
	log.Printf("[TRACE] mdb_clickhouse_user: flattened settings: %+v\n", state.GetSettings())
	state.SetConnectionManager(flattenConnectionManager(ctx, user.ConnectionManager, &diags))
	log.Printf("[TRACE] mdb_clickhouse_user: flattened connection_manager: %+v\n", state.GetConnectionManager())

	return diags
}

func userFromState(ctx context.Context, state *ResourceUser) (*clickhouse.UserSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	log.Printf("[TRACE] mdb_clickhouse_user: expand user from state: %+v\n", state)
	permissions := expandPermissionsFromState(ctx, state.Permissions, &diags)
	log.Printf("[TRACE] mdb_clickhouse_user: expanded quotas: %+v\n", permissions)
	quotas := expandQuotasFromState(ctx, state.Quotas, &diags)
	log.Printf("[TRACE] mdb_clickhouse_user: expanded quotas: %+v\n", quotas)
	settings := expandSettingsFromState(ctx, state.Settings, &diags)
	log.Printf("[TRACE] mdb_clickhouse_user: expanded settings: %+v\n", settings)
	return &clickhouse.UserSpec{
		Name:             state.Name.ValueString(),
		Password:         state.Password.ValueString(),
		Permissions:      permissions,
		Quotas:           quotas,
		Settings:         settings,
		GeneratePassword: wrapperspb.Bool(state.GeneratePassword.ValueBool()),
	}, diags
}
