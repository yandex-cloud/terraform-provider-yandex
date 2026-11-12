package usersettings

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"readonly": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"allow_ddl": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"insert_quorum": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"connect_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"receive_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"send_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"insert_quorum_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"select_sequential_consistency": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"max_replica_delay_for_distributed_queries": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"fallback_to_stale_replicas_for_distributed_queries": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"replication_alter_partitions_sync": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"distributed_product_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"distributed_aggregation_memory_efficient": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"distributed_ddl_task_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"skip_unavailable_shards": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"compile_expressions": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"min_count_to_compile_expression": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_block_size": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"min_insert_block_size_rows": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"min_insert_block_size_bytes": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_insert_block_size": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"min_bytes_to_use_direct_io": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"use_uncompressed_cache": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"merge_tree_max_rows_to_use_cache": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"merge_tree_max_bytes_to_use_cache": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"merge_tree_min_rows_for_concurrent_read": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"merge_tree_min_bytes_for_concurrent_read": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_bytes_before_external_group_by": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_bytes_before_external_sort": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"group_by_two_level_threshold": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"group_by_two_level_threshold_bytes": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"priority": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_threads": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_memory_usage": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_memory_usage_for_user": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_network_bandwidth": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_network_bandwidth_for_user": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"force_index_by_date": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"force_primary_key": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"max_rows_to_read": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_bytes_to_read": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"read_overflow_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_rows_to_group_by": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"group_by_overflow_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_rows_to_sort": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_bytes_to_sort": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"sort_overflow_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_result_rows": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_result_bytes": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"result_overflow_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_rows_in_distinct": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_bytes_in_distinct": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"distinct_overflow_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_rows_to_transfer": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_bytes_to_transfer": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"transfer_overflow_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_execution_time": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"timeout_overflow_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_rows_in_set": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_bytes_in_set": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"set_overflow_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_rows_in_join": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_bytes_in_join": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"join_overflow_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_columns_to_read": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_temporary_columns": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_temporary_non_const_columns": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_query_size": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_ast_depth": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_ast_elements": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_expanded_ast_elements": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"min_execution_speed": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"min_execution_speed_bytes": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"count_distinct_implementation": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"input_format_values_interpret_expressions": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"input_format_defaults_for_omitted_fields": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"output_format_json_quote_64bit_integers": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"output_format_json_quote_denormals": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"low_cardinality_allow_in_native_format": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"empty_result_for_aggregation_by_empty_set": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"joined_subquery_requires_alias": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"join_use_nulls": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"transform_null_in": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"http_connection_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"http_receive_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"http_send_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"enable_http_compression": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"send_progress_in_http_headers": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"http_headers_progress_interval": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"add_http_cors_header": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"quota_mode": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"max_concurrent_queries_for_user": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"memory_profiler_step": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"memory_profiler_sample_probability": schema.Float64Attribute{
			Optional: true,
			Computed: true,
		},
		"insert_null_as_default": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"allow_suspicious_low_cardinality_types": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"connect_timeout_with_failover": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"allow_introspection_functions": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"async_insert": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"async_insert_threads": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"wait_for_async_insert": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"wait_for_async_insert_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"async_insert_max_data_size": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"async_insert_busy_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"async_insert_stale_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"timeout_before_checking_execution_speed": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"cancel_http_readonly_queries_on_client_close": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"flatten_nested": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"max_http_get_redirects": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"input_format_import_nested_json": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"input_format_parallel_parsing": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"max_final_threads": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_read_buffer_size": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"local_filesystem_read_method": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"remote_filesystem_read_method": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"insert_keeper_max_retries": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"do_not_merge_across_partitions_select_final": schema.BoolAttribute{
			MarkdownDescription: "Enable or disable independent processing of partitions for **SELECT** queries with **FINAL**.",
			Optional:            true,
			Computed:            true,
		},
		"max_temporary_data_on_disk_size_for_user": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_temporary_data_on_disk_size_for_query": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_parser_depth": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"memory_overcommit_ratio_denominator": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"memory_overcommit_ratio_denominator_for_user": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"memory_usage_overcommit_max_wait_microseconds": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"log_query_threads": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"log_query_views": schema.BoolAttribute{
			MarkdownDescription: "Enables or disables query views logging to the the system.query_views_log table.",
			Optional:            true,
			Computed:            true,
		},
		"max_insert_threads": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"use_hedged_requests": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"idle_connection_timeout": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"hedged_connection_timeout_ms": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"load_balancing": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"prefer_localhost_replica": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"date_time_input_format": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"date_time_output_format": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"format_regexp": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"format_regexp_skip_unmatched": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"input_format_with_names_use_header": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"input_format_null_as_default": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"insert_quorum_parallel": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"max_partitions_per_insert_block": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"deduplicate_blocks_in_dependent_materialized_views": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"any_join_distinct_right_table_keys": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"join_algorithm": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
		},
		"format_avro_schema_registry_url": schema.StringAttribute{
			MarkdownDescription: "Avro schema registry URL.",
			Optional:            true,
			Computed:            true,
		},
		"data_type_default_nullable": schema.BoolAttribute{
			MarkdownDescription: "Allows data types without explicit modifiers NULL or NOT NULL in column definition will be Nullable.",
			Optional:            true,
			Computed:            true,
		},
		"http_max_field_name_size": schema.Int64Attribute{
			MarkdownDescription: "Maximum length of field name in HTTP header.",
			Optional:            true,
			Computed:            true,
		},
		"http_max_field_value_size": schema.Int64Attribute{
			MarkdownDescription: "Maximum length of field value in HTTP header.",
			Optional:            true,
			Computed:            true,
		},
		"async_insert_use_adaptive_busy_timeout": schema.BoolAttribute{
			MarkdownDescription: "If it is set to true, use adaptive busy timeout for asynchronous inserts.",
			Optional:            true,
			Computed:            true,
		},
		"log_queries_probability": schema.Float64Attribute{
			MarkdownDescription: "Log queries with the specified probability.",
			Optional:            true,
			Computed:            true,
		},
		"log_processors_profiles": schema.BoolAttribute{
			MarkdownDescription: "Enabled or disable logging of processors level profiling data to the the system.processors_profile_log table.",
			Optional:            true,
			Computed:            true,
		},
		"use_query_cache": schema.BoolAttribute{
			MarkdownDescription: "If turned on, SELECT queries may utilize the query cache.",
			Optional:            true,
			Computed:            true,
		},
		"enable_reads_from_query_cache": schema.BoolAttribute{
			MarkdownDescription: "If turned on, results of SELECT queries are retrieved from the query cache.",
			Optional:            true,
			Computed:            true,
		},
		"enable_writes_to_query_cache": schema.BoolAttribute{
			MarkdownDescription: "If turned on, results of SELECT queries are stored in the query cache.",
			Optional:            true,
			Computed:            true,
		},
		"query_cache_min_query_runs": schema.Int64Attribute{
			MarkdownDescription: "Minimum number of times a SELECT query must run before its result is stored in the query cache.",
			Optional:            true,
			Computed:            true,
		},
		"query_cache_min_query_duration": schema.Int64Attribute{
			MarkdownDescription: "Minimum duration in milliseconds a query needs to run for its result to be stored in the query cache.",
			Optional:            true,
			Computed:            true,
		},
		"query_cache_ttl": schema.Int64Attribute{
			MarkdownDescription: "After this time in seconds entries in the query cache become stale.",
			Optional:            true,
			Computed:            true,
		},
		"query_cache_max_entries": schema.Int64Attribute{
			MarkdownDescription: "The maximum number of query results the current user may store in the query cache. 0 means unlimited.",
			Optional:            true,
			Computed:            true,
		},
		"query_cache_max_size_in_bytes": schema.Int64Attribute{
			MarkdownDescription: "The maximum amount of memory (in bytes) the current user may allocate in the query cache. 0 means unlimited.",
			Optional:            true,
			Computed:            true,
		},
		"query_cache_tag": schema.StringAttribute{
			MarkdownDescription: "A string which acts as a label for query cache entries. The same queries with different tags are considered different by the query cache.",
			Optional:            true,
			Computed:            true,
		},
		"query_cache_share_between_users": schema.BoolAttribute{
			MarkdownDescription: "If turned on, the result of SELECT queries cached in the query cache can be read by other users. It is not recommended to enable this setting due to security reasons.",
			Optional:            true,
			Computed:            true,
		},
		"query_cache_nondeterministic_function_handling": schema.StringAttribute{
			MarkdownDescription: "Controls how the query cache handles **SELECT** queries with non-deterministic functions like rand() or now().",
			Optional:            true,
			Computed:            true,
			Validators:          UserSettings_QueryCacheNondeterministicFunctionHandling_validator,
		},
		"query_cache_system_table_handling": schema.StringAttribute{
			MarkdownDescription: "Controls how the query cache handles **SELECT** queries against system tables.",
			Optional:            true,
			Computed:            true,
			Validators:          UserSettings_QueryCacheSystemTableHandling_validator,
		},
		"ignore_materialized_views_with_dropped_target_table": schema.BoolAttribute{
			MarkdownDescription: "Ignore materialized views with dropped target table during pushing to views.",
			Optional:            true,
			Computed:            true,
		},
		"enable_analyzer": schema.BoolAttribute{
			MarkdownDescription: "Enable new query analyzer.",
			Optional:            true,
			Computed:            true,
		},
		"distributed_ddl_output_mode": schema.StringAttribute{
			MarkdownDescription: "Determines the format of distributed DDL query result.",
			Optional:            true,
			Computed:            true,
			Validators:          UserSettings_DistributedDdlOutputMode_validator,
		},
		"s3_use_adaptive_timeouts": schema.BoolAttribute{
			MarkdownDescription: "Enables or disables adaptive timeouts for S3 requests.",
			Optional:            true,
			Computed:            true,
		},
		"show_data_lake_catalogs_in_system_tables": schema.BoolAttribute{
			MarkdownDescription: "Enables or disables showing data lake catalogs in system tables.",
			Optional:            true,
			Computed:            true,
		},
		"max_bytes_ratio_before_external_group_by": schema.Float64Attribute{
			MarkdownDescription: "The ratio of available memory that is allowed for GROUP BY. Once reached, external aggregation is used.",
			Optional:            true,
			Computed:            true,
		},
		"max_bytes_ratio_before_external_sort": schema.Float64Attribute{
			MarkdownDescription: "The ratio of available memory that is allowed for ORDER BY. Once reached, external sort is used.",
			Optional:            true,
			Computed:            true,
		},
		"use_hive_partitioning": schema.BoolAttribute{
			MarkdownDescription: "Enables Hive-style partitioning when reading from file-like table engines.",
			Optional:            true,
			Computed:            true,
		},
		"max_network_bytes": schema.Int64Attribute{
			MarkdownDescription: "Limits the amount of data exchanged over the network in bytes for a query. Zero means unlimited.",
			Optional:            true,
			Computed:            true,
		},
		"connect_timeout_with_failover_secure": schema.Int64Attribute{
			MarkdownDescription: "Connection timeout in milliseconds for selecting the first healthy replica for secure connections.",
			Optional:            true,
			Computed:            true,
		},
		"connections_with_failover_max_tries": schema.Int64Attribute{
			MarkdownDescription: "The maximum number of connection attempts with each replica for the Distributed table engine.",
			Optional:            true,
			Computed:            true,
		},
		"compatibility": schema.StringAttribute{
			MarkdownDescription: "Makes ClickHouse use the default settings of the specified previous ClickHouse version.",
			Optional:            true,
			Computed:            true,
		},
		"materialize_ttl_after_modify": schema.BoolAttribute{
			MarkdownDescription: "Applies TTL to old data after an ALTER MODIFY TTL query.",
			Optional:            true,
			Computed:            true,
		},
		"max_remote_read_network_bandwidth": schema.Int64Attribute{
			MarkdownDescription: "Limits the speed of remote reads over the network in bytes per second. Zero means unlimited.",
			Optional:            true,
			Computed:            true,
		},
		"max_remote_write_network_bandwidth": schema.Int64Attribute{
			MarkdownDescription: "Limits the speed of remote writes over the network in bytes per second. Zero means unlimited.",
			Optional:            true,
			Computed:            true,
		},
		"async_socket_for_remote": schema.BoolAttribute{
			MarkdownDescription: "Enables asynchronous reads from sockets while executing remote queries.",
			Optional:            true,
			Computed:            true,
		},
		"async_query_sending_for_remote": schema.BoolAttribute{
			MarkdownDescription: "Enables asynchronous connection creation and query sending while executing remote queries.",
			Optional:            true,
			Computed:            true,
		},
		"allow_reorder_prewhere_conditions": schema.BoolAttribute{
			MarkdownDescription: "Allows reordering conditions when moving them from WHERE to PREWHERE.",
			Optional:            true,
			Computed:            true,
		},
		"database_atomic_wait_for_drop_and_detach_synchronously": schema.BoolAttribute{
			MarkdownDescription: "Makes DROP and DETACH queries wait for completion of table removal when using Atomic databases.",
			Optional:            true,
			Computed:            true,
		},
	}
}
