package mdb_clickhouse_user

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func UserSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a ClickHouse user within the Yandex.Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the ClickHouse cluster. Provided by the client when the user is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the ClickHouse user. Provided by the client when the user is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password of the ClickHouse user. Provided by the client when the user is created.",
				Required:            true,
				Sensitive:           true,
			},
		},
		Blocks: map[string]schema.Block{
			"permission": PermissionSchema(),
			"quota":      QuotasSchema(),
			"settings":   SettingsSchema(),
		},
	}
}

func PermissionSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: "Block represents databases that are permitted to user.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"database_name": schema.StringAttribute{
					MarkdownDescription: "Name of the database that the permission grants access to.",
					Required:            true,
				},
			},
		},
	}
}

func QuotasSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: "ClickHouse quota representation. Each quota associated with an user and limits it resource usage for an interval. For more information, see [the official documentation](https://clickhouse.com/docs/en/operations/quotas)",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"interval_duration": schema.Int64Attribute{MarkdownDescription: "Duration of interval for quota in milliseconds.", Required: true},
				"queries":           schema.Int64Attribute{MarkdownDescription: "The total number of queries. 0 - unlimited.", Optional: true},
				"errors":            schema.Int64Attribute{MarkdownDescription: "The number of queries that threw exception. 0 - unlimited.", Optional: true},
				"result_rows":       schema.Int64Attribute{MarkdownDescription: "The total number of rows given as the result. 0 - unlimited.", Optional: true},
				"read_rows":         schema.Int64Attribute{MarkdownDescription: "The total number of source rows read from tables for running the query, on all remote servers. 0 - unlimited.", Optional: true},
				"execution_time":    schema.Int64Attribute{MarkdownDescription: "The total query execution time, in milliseconds (wall time). 0 - unlimited.", Optional: true},
			},
		},
	}
}

func SettingsSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "Block represents ClickHouse user settings. For more information, see [the official documentation](https://clickhouse.com/docs/ru/operations/settings/settings)",
		Attributes: map[string]schema.Attribute{
			"readonly": schema.Int64Attribute{
				MarkdownDescription: "Restricts permissions for non-DDL queries.",
				Optional:            true,
			},
			"allow_ddl": schema.BoolAttribute{
				MarkdownDescription: "Allows or denies DDL queries.",
				Optional:            true,
			},
			"insert_quorum": schema.Int64Attribute{
				MarkdownDescription: "Enables the quorum writes.",
				Optional:            true,
			},
			"connect_timeout": schema.Int64Attribute{
				MarkdownDescription: "Connection timeout in milliseconds.",
				Optional:            true,
			},
			"receive_timeout": schema.Int64Attribute{
				MarkdownDescription: "Receive timeout in milliseconds.",
				Optional:            true,
			},
			"send_timeout": schema.Int64Attribute{
				MarkdownDescription: "Send timeout in milliseconds.",
				Optional:            true,
			},
			"insert_quorum_timeout": schema.Int64Attribute{
				MarkdownDescription: "Quorum write timeout in milliseconds.",
				Optional:            true,
			},
			"select_sequential_consistency": schema.BoolAttribute{
				MarkdownDescription: "Determines the behavior of SELECT queries from replicated tables. If enabled, ClickHouse will terminate a query with error message in case the replica does not have a chunk written with the quorum and will not read the parts that have not yet been written with the quorum.",
				Optional:            true,
			},
			"max_replica_delay_for_distributed_queries": schema.Int64Attribute{
				MarkdownDescription: "Max replica delay in milliseconds. If a replica lags more than the set value,this replica is not used and becomes a stale one.",
				Optional:            true,
			},
			"fallback_to_stale_replicas_for_distributed_queries": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables query forcing to a stale replica in case the actual data is unavailable. If enabled, ClickHouse will choose the most up-to-date replica and force the query to use the data in this replica.",
				Optional:            true,
			},
			"replication_alter_partitions_sync": schema.Int64Attribute{
				MarkdownDescription: "Wait mode for asynchronous actions in ALTER queries on replicated tables.",
				Optional:            true,
			},
			"distributed_product_mode": schema.StringAttribute{
				MarkdownDescription: "Determine the behavior of distributed subqueries.",
				Optional:            true,
				Validators:          UserSettings_DistributedProductMode_validator,
			},
			"distributed_aggregation_memory_efficient": schema.BoolAttribute{
				MarkdownDescription: "Enables of disables memory saving mode when doing distributed aggregation.",
				Optional:            true,
			},
			"distributed_ddl_task_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for DDL queries, in milliseconds.",
				Optional:            true,
			},
			"skip_unavailable_shards": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables silent skipping of unavailable shards",
				Optional:            true,
			},
			"compile_expressions": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable compilation of some scalar functions and operators to native code.",
				Optional:            true,
			},
			"min_count_to_compile_expression": schema.Int64Attribute{
				MarkdownDescription: "Minimum count of executing same expression before it is get compiled.",
				Optional:            true,
			},
			"max_block_size": schema.Int64Attribute{
				MarkdownDescription: "A recommendation for what size of the block (in a count of rows) to load from tables.",
				Optional:            true,
			},
			"min_insert_block_size_rows": schema.Int64Attribute{
				MarkdownDescription: "Sets the minimum number of rows in the block which can be inserted into a table by an INSERT query.",
				Optional:            true,
			},
			"min_insert_block_size_bytes": schema.Int64Attribute{
				MarkdownDescription: "Sets the minimum number of bytes in the block which can be inserted into a table by an INSERT query.",
				Optional:            true,
			},
			"max_insert_block_size": schema.Int64Attribute{
				MarkdownDescription: "The size of blocks (in a count of rows) to form for insertion into a table.",
				Optional:            true,
			},
			"min_bytes_to_use_direct_io": schema.Int64Attribute{
				MarkdownDescription: "The minimum data volume required for using direct I/O access to the storage disk.",
				Optional:            true,
			},
			"use_uncompressed_cache": schema.BoolAttribute{
				MarkdownDescription: "Whether to use a cache of uncompressed blocks.",
				Optional:            true,
			},
			"merge_tree_max_rows_to_use_cache": schema.Int64Attribute{
				MarkdownDescription: "If ClickHouse should read more than merge_tree_max_rows_to_use_cache rows in one query, it doesn’t use the cache of uncompressed blocks.",
				Optional:            true,
			},
			"merge_tree_max_bytes_to_use_cache": schema.Int64Attribute{
				MarkdownDescription: "If ClickHouse should read more than merge_tree_max_bytes_to_use_cache bytes in one query, it doesn’t use the cache of uncompressed blocks.",
				Optional:            true,
			},
			"merge_tree_min_rows_for_concurrent_read": schema.Int64Attribute{
				MarkdownDescription: "If the number of rows to be read from a file of a MergeTree table exceeds merge_tree_min_rows_for_concurrent_read then ClickHouse tries to perform a concurrent reading from this file on several threads.",
				Optional:            true,
			},
			"merge_tree_min_bytes_for_concurrent_read": schema.Int64Attribute{
				MarkdownDescription: "If the number of bytes to read from one file of a MergeTree-engine table exceeds merge_tree_min_bytes_for_concurrent_read, then ClickHouse tries to concurrently read from this file in several threads.",
				Optional:            true,
			},
			"max_bytes_before_external_group_by": schema.Int64Attribute{
				MarkdownDescription: "Limit in bytes for using memoru for GROUP BY before using swap on disk.",
				Optional:            true,
			},
			"max_bytes_before_external_sort": schema.Int64Attribute{
				MarkdownDescription: "This setting is equivalent of the max_bytes_before_external_group_by setting, except for it is for sort operation (ORDER BY), not aggregation.",
				Optional:            true,
			},
			"group_by_two_level_threshold": schema.Int64Attribute{
				MarkdownDescription: "Sets the threshold of the number of keys, after that the two-level aggregation should be used.",
				Optional:            true,
			},
			"group_by_two_level_threshold_bytes": schema.Int64Attribute{
				MarkdownDescription: "Sets the threshold of the number of bytes, after that the two-level aggregation should be used.",
				Optional:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Priority of the query.",
				Optional:            true,
			},
			"max_threads": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of query processing threads, excluding threads for retrieving data from remote servers.",
				Optional:            true,
			},
			"max_memory_usage": schema.Int64Attribute{
				MarkdownDescription: "Maximum memory usage for processing all concurrently running queries for the user. Zero means unlimited.",
				Optional:            true,
			},
			"max_memory_usage_for_user": schema.Int64Attribute{
				MarkdownDescription: "Maximum memory usage for processing all concurrently running queries for the user. Zero means unlimited.",
				Optional:            true,
			},
			"max_network_bandwidth": schema.Int64Attribute{
				MarkdownDescription: "Limits the speed of the data exchange over the network in bytes per second.  This setting applies to every query.",
				Optional:            true,
			},
			"max_network_bandwidth_for_user": schema.Int64Attribute{
				MarkdownDescription: "Limits the speed of the data exchange over the network in bytes per second. This setting applies to all concurrently running queries performed by a single user.",
				Optional:            true,
			},
			"force_index_by_date": schema.BoolAttribute{
				MarkdownDescription: "Disable query execution if the index cannot be used by date.",
				Optional:            true,
			},
			"force_primary_key": schema.BoolAttribute{
				MarkdownDescription: "Disable query execution if indexing by the primary key is not possible.",
				Optional:            true,
			},
			"max_rows_to_read": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of rows that can be read from a table when running a query.",
				Optional:            true,
			},
			"max_bytes_to_read": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of bytes (uncompressed data) that can be read from a table when running a query.",
				Optional:            true,
			},
			"read_overflow_mode": schema.StringAttribute{
				MarkdownDescription: "Sets behaviour on overflow while read. Possible values: * throw - abort query execution, return an error.  * break - stop query execution, return partial result.",
				Optional:            true,
				Validators:          UserSettings_OverflowMode_validator,
			},
			"max_rows_to_group_by": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of unique keys received from aggregation function.",
				Optional:            true,
			},
			"group_by_overflow_mode": schema.StringAttribute{
				MarkdownDescription: "Sets behaviour on overflow while GROUP BY operation.",
				Optional:            true,
				Validators:          UserSettings_GroupByOverflowMode_validator,
			},
			"max_rows_to_sort": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of rows that can be read from a table for sorting.",
				Optional:            true,
			},
			"max_bytes_to_sort": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of bytes (uncompressed data) that can be read from a table for sorting.",
				Optional:            true,
			},
			"sort_overflow_mode": schema.StringAttribute{
				MarkdownDescription: "Sets behaviour on overflow while sort.",
				Optional:            true,
				Validators:          UserSettings_OverflowMode_validator,
			},
			"max_result_rows": schema.Int64Attribute{
				MarkdownDescription: "Limits the number of rows in the result.",
				Optional:            true,
			},
			"max_result_bytes": schema.Int64Attribute{
				MarkdownDescription: "Limits the number of bytes in the result.",
				Optional:            true,
			},
			"result_overflow_mode": schema.StringAttribute{
				MarkdownDescription: "Sets behaviour on overflow in result.",
				Optional:            true,
				Validators:          UserSettings_OverflowMode_validator,
			},
			"max_rows_in_distinct": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of different rows when using DISTINCT.",
				Optional:            true,
			},
			"max_bytes_in_distinct": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum size of a hash table in bytes (uncompressed data) when using DISTINCT.",
				Optional:            true,
			},
			"distinct_overflow_mode": schema.StringAttribute{
				MarkdownDescription: "Sets behaviour on overflow when using DISTINCT.",
				Optional:            true,
				Validators:          UserSettings_OverflowMode_validator,
			},
			"max_rows_to_transfer": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of rows that can be passed to a remote server or saved in a temporary table when using GLOBAL IN.",
				Optional:            true,
			},
			"max_bytes_to_transfer": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of bytes (uncompressed data) that can be passed to a remote server or saved in a temporary table when using GLOBAL IN.",
				Optional:            true,
			},
			"transfer_overflow_mode": schema.StringAttribute{
				MarkdownDescription: "Sets behaviour on overflow.",
				Optional:            true,
				Validators:          UserSettings_OverflowMode_validator,
			},
			"max_execution_time": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum query execution time in milliseconds.",
				Optional:            true,
			},
			"timeout_overflow_mode": schema.StringAttribute{
				MarkdownDescription: "Sets behaviour on overflow.",
				Optional:            true,
				Validators:          UserSettings_OverflowMode_validator,
			},
			"max_rows_in_set": schema.Int64Attribute{
				MarkdownDescription: "Limit on the number of rows in the set resulting from the execution of the IN section.",
				Optional:            true,
			},
			"max_bytes_in_set": schema.Int64Attribute{
				MarkdownDescription: "Limit on the number of bytes in the set resulting from the execution of the IN section.",
				Optional:            true,
			},
			"set_overflow_mode": schema.StringAttribute{
				MarkdownDescription: "Sets behaviour on overflow in the set resulting.",
				Optional:            true,
				Validators:          UserSettings_OverflowMode_validator,
			},
			"max_rows_in_join": schema.Int64Attribute{
				MarkdownDescription: "Limit on maximum size of the hash table for JOIN, in rows.",
				Optional:            true,
			},
			"max_bytes_in_join": schema.Int64Attribute{
				MarkdownDescription: "Limit on maximum size of the hash table for JOIN, in bytes.",
				Optional:            true,
			},
			"join_overflow_mode": schema.StringAttribute{
				MarkdownDescription: "Sets behaviour on overflow in JOIN.",
				Optional:            true,
				Validators:          UserSettings_OverflowMode_validator,
			},
			"max_columns_to_read": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of columns that can be read from a table in a single query.",
				Optional:            true,
			},
			"max_temporary_columns": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of temporary columns that must be kept in RAM at the same time when running a query, including constant columns.",
				Optional:            true,
			},
			"max_temporary_non_const_columns": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of temporary columns that must be kept in RAM at the same time when running a query, excluding constant columns.",
				Optional:            true,
			},
			"max_query_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum part of a query that can be taken to RAM for parsing with the SQL parser.",
				Optional:            true,
			},
			"max_ast_depth": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum depth of query syntax tree.",
				Optional:            true,
			},
			"max_ast_elements": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum size of query syntax tree in number of nodes.",
				Optional:            true,
			},
			"max_expanded_ast_elements": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum size of query syntax tree in number of nodes after expansion of aliases and the asterisk values.",
				Optional:            true,
			},
			"min_execution_speed": schema.Int64Attribute{
				MarkdownDescription: "Minimal execution speed in rows per second.",
				Optional:            true,
			},
			"min_execution_speed_bytes": schema.Int64Attribute{
				MarkdownDescription: "Minimal execution speed in bytes per second.",
				Optional:            true,
			},
			"count_distinct_implementation": schema.StringAttribute{
				MarkdownDescription: "Specifies which of the uniq* functions should be used to perform the COUNT(DISTINCT …) construction.",
				Optional:            true,
				Validators:          UserSettings_CountDistinctImplementation_validator,
			},
			"input_format_values_interpret_expressions": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables the full SQL parser if the fast stream parser can’t parse the data.",
				Optional:            true,
			},
			"input_format_defaults_for_omitted_fields": schema.BoolAttribute{
				MarkdownDescription: "When performing INSERT queries, replace omitted input column values with default values of the respective columns.",
				Optional:            true,
			},
			"output_format_json_quote_64bit_integers": schema.BoolAttribute{
				MarkdownDescription: "If the value is true, integers appear in quotes when using JSON* Int64 and UInt64 formats (for compatibility with most JavaScript implementations); otherwise, integers are output without the quotes.",
				Optional:            true,
			},
			"output_format_json_quote_denormals": schema.BoolAttribute{
				MarkdownDescription: "Enables +nan, -nan, +inf, -inf outputs in JSON output format.",
				Optional:            true,
			},
			"low_cardinality_allow_in_native_format": schema.BoolAttribute{
				MarkdownDescription: "Allows or restricts using the LowCardinality data type with the Native format.",
				Optional:            true,
			},
			"empty_result_for_aggregation_by_empty_set": schema.BoolAttribute{
				MarkdownDescription: "Allows to retunr empty result.",
				Optional:            true,
			},
			"joined_subquery_requires_alias": schema.BoolAttribute{
				MarkdownDescription: "Require aliases for subselects and table functions in FROM that more than one table is present.",
				Optional:            true,
			},
			"join_use_nulls": schema.BoolAttribute{
				MarkdownDescription: "Sets the type of JOIN behaviour. When merging tables, empty cells may appear. ClickHouse fills them differently based on this setting.",
				Optional:            true,
			},
			"transform_null_in": schema.BoolAttribute{
				MarkdownDescription: "Enables equality of NULL values for IN operator.",
				Optional:            true,
			},
			"http_connection_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for HTTP connection in milliseconds.",
				Optional:            true,
			},
			"http_receive_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for HTTP connection in milliseconds.",
				Optional:            true,
			},
			"http_send_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for HTTP connection in milliseconds.",
				Optional:            true,
			},
			"enable_http_compression": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables data compression in the response to an HTTP request.",
				Optional:            true,
			},
			"send_progress_in_http_headers": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables X-ClickHouse-Progress HTTP response headers in clickhouse-server responses.",
				Optional:            true,
			},
			"http_headers_progress_interval": schema.Int64Attribute{
				MarkdownDescription: "Sets minimal interval between notifications about request process in HTTP header X-ClickHouse-Progress.",
				Optional:            true,
			},
			"add_http_cors_header": schema.BoolAttribute{
				MarkdownDescription: "Include CORS headers in HTTP response.",
				Optional:            true,
			},
			"quota_mode": schema.StringAttribute{
				MarkdownDescription: "Quota accounting mode.",
				Optional:            true,
				Validators:          UserSettings_QuotaMode_validator,
			},
			"max_concurrent_queries_for_user": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of concurrent requests per user. Default value: 0 (no limit).",
				Optional:            true,
			},
			"memory_profiler_step": schema.Int64Attribute{
				MarkdownDescription: "Memory profiler step (in bytes). If the next query step requires more memory than this parameter specifies, the memory profiler collects the allocating stack trace. Values lower than a few megabytes slow down query processing. Default value: 4194304 (4 MB). Zero means disabled memory profiler.",
				Optional:            true,
			},
			"memory_profiler_sample_probability": schema.Float64Attribute{
				MarkdownDescription: "Collect random allocations and deallocations and write them into system.trace_log with 'MemorySample' trace_type. The probability is for every alloc/free regardless to the size of the allocation. Possible values: from 0 to 1. Default: 0.",
				Optional:            true,
			},
			"insert_null_as_default": schema.BoolAttribute{
				MarkdownDescription: "Enables the insertion of default values instead of NULL into columns with not nullable data type. Default value: true.",
				Optional:            true,
			},
			"allow_suspicious_low_cardinality_types": schema.BoolAttribute{
				MarkdownDescription: "Allows specifying LowCardinality modifier for types of small fixed size (8 or less) in CREATE TABLE statements. Enabling this may increase merge times and memory consumption.",
				Optional:            true,
			},
			"connect_timeout_with_failover": schema.Int64Attribute{
				MarkdownDescription: "The timeout in milliseconds for connecting to a remote server for a Distributed table engine.  Applies only if the cluster uses sharding and replication. If unsuccessful, several attempts are made to connect to various replicas.",
				Optional:            true,
			},
			"allow_introspection_functions": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables introspection functions for query profiling.",
				Optional:            true,
			},
			"async_insert": schema.BoolAttribute{
				MarkdownDescription: "Enables asynchronous inserts. Disabled by default.",
				Optional:            true,
			},
			"async_insert_threads": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads for background data parsing and insertion. If the parameter is set to 0, asynchronous insertions are disabled. Default value: 16.",
				Optional:            true,
			},
			"wait_for_async_insert": schema.BoolAttribute{
				MarkdownDescription: "Enables waiting for processing of asynchronous insertion. If enabled, server returns OK only after the data is inserted.",
				Optional:            true,
			},
			"wait_for_async_insert_timeout": schema.Int64Attribute{
				MarkdownDescription: "The timeout (in seconds) for waiting for processing of asynchronous insertion. Value must be at least 1000 (1 second).",
				Optional:            true,
			},
			"async_insert_max_data_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size of the unparsed data in bytes collected per query before being inserted. If the parameter is set to 0, asynchronous insertions are disabled. Default value: 100000.",
				Optional:            true,
			},
			"async_insert_busy_timeout": schema.Int64Attribute{
				MarkdownDescription: "The maximum timeout in milliseconds since the first INSERT query before inserting collected data. If the parameter is set to 0, the timeout is disabled. Default value: 200.",
				Optional:            true,
			},
			"async_insert_stale_timeout": schema.Int64Attribute{
				MarkdownDescription: "The maximum timeout in milliseconds since the last INSERT query before dumping collected data. If enabled, the settings prolongs the async_insert_busy_timeout with every INSERT query as long as async_insert_max_data_size is not exceeded.",
				Optional:            true,
			},
			"timeout_before_checking_execution_speed": schema.Int64Attribute{
				MarkdownDescription: "Timeout (in seconds) between checks of execution speed. It is checked that execution speed is not less that specified in min_execution_speed parameter. Must be at least 1000.",
				Optional:            true,
			},
			"cancel_http_readonly_queries_on_client_close": schema.BoolAttribute{
				MarkdownDescription: "Cancels HTTP read-only queries (e.g. SELECT) when a client closes the connection without waiting for the response. Default value: false.",
				Optional:            true,
			},
			"flatten_nested": schema.BoolAttribute{
				MarkdownDescription: "Sets the data format of a nested columns.",
				Optional:            true,
			},
			"max_http_get_redirects": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of HTTP GET redirect hops for URL-engine tables.",
				Optional:            true,
			},
			"input_format_import_nested_json": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables the insertion of JSON data with nested objects.",
				Optional:            true,
			},
			"input_format_parallel_parsing": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables order-preserving parallel parsing of data formats. Supported only for TSV, TKSV, CSV and JSONEachRow formats.",
				Optional:            true,
			},
			"max_final_threads": schema.Int64Attribute{
				MarkdownDescription: "Sets the maximum number of parallel threads for the SELECT query data read phase with the FINAL modifier.",
				Optional:            true,
			},
			"max_read_buffer_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size of the buffer to read from the filesystem.",
				Optional:            true,
			},
			"local_filesystem_read_method": schema.StringAttribute{
				MarkdownDescription: `Method of reading data from local filesystem. Possible values: 
* 'read' - abort query execution, return an error.  
* 'pread' - abort query execution, return an error.  
* 'pread_threadpool' - stop query execution, return partial result. 
If the parameter is set to 0 (default), no hops is allowed.`,
				Optional:   true,
				Validators: UserSettings_LocalFilesystemReadMethod_validator,
			},
			"remote_filesystem_read_method": schema.StringAttribute{
				MarkdownDescription: "Method of reading data from remote filesystem, one of: `read`, `threadpool`.",
				Optional:            true,
				Validators:          UserSettings_RemoteFilesystemReadMethod_validator,
			},
			"insert_keeper_max_retries": schema.Int64Attribute{
				MarkdownDescription: "The setting sets the maximum number of retries for ClickHouse Keeper (or ZooKeeper) requests during insert into replicated MergeTree. Only Keeper requests which failed due to network error, Keeper session timeout, or request timeout are considered for retries.",
				Optional:            true,
			},
			"max_temporary_data_on_disk_size_for_user": schema.Int64Attribute{
				MarkdownDescription: "The maximum amount of data consumed by temporary files on disk in bytes for all concurrently running user queries. Zero means unlimited.",
				Optional:            true,
			},
			"max_temporary_data_on_disk_size_for_query": schema.Int64Attribute{
				MarkdownDescription: "The maximum amount of data consumed by temporary files on disk in bytes for all concurrently running queries. Zero means unlimited.",
				Optional:            true,
			},
			"max_parser_depth": schema.Int64Attribute{
				MarkdownDescription: "Limits maximum recursion depth in the recursive descent parser. Allows controlling the stack size.",
				Optional:            true,
			},
			"memory_overcommit_ratio_denominator": schema.Int64Attribute{
				MarkdownDescription: "It represents soft memory limit in case when hard limit is reached on user level. This value is used to compute overcommit ratio for the query. Zero means skip the query.",
				Optional:            true,
			},
			"memory_overcommit_ratio_denominator_for_user": schema.Int64Attribute{
				MarkdownDescription: "It represents soft memory limit in case when hard limit is reached on global level. This value is used to compute overcommit ratio for the query. Zero means skip the query.",
				Optional:            true,
			},
			"memory_usage_overcommit_max_wait_microseconds": schema.Int64Attribute{
				MarkdownDescription: "Maximum time thread will wait for memory to be freed in the case of memory overcommit on a user level. If the timeout is reached and memory is not freed, an exception is thrown.",
				Optional:            true,
			},
			"log_query_threads": schema.BoolAttribute{
				MarkdownDescription: "Setting up query threads logging. Query threads log into the system.query_thread_log table. This setting has effect only when log_queries is true. Queries’ threads run by ClickHouse with this setup are logged according to the rules in the query_thread_log server configuration parameter. Default value: true.",
				Optional:            true,
			},
			"max_insert_threads": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads to execute the INSERT SELECT query. Default value: 0.",
				Optional:            true,
			},
			"use_hedged_requests": schema.BoolAttribute{
				MarkdownDescription: "Enables hedged requests logic for remote queries. It allows to establish many connections with different replicas for query. New connection is enabled in case existent connection(s) with replica(s) were not established within hedged_connection_timeout or no data was received within receive_data_timeout. Query uses the first connection which send non empty progress packet (or data packet, if allow_changing_replica_until_first_data_packet); other connections are cancelled. Queries with max_parallel_replicas > 1 are supported. Default value: true.",
				Optional:            true,
			},
			"idle_connection_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout to close idle TCP connections after specified number of seconds. Default value: 3600 seconds.",
				Optional:            true,
			},
			"hedged_connection_timeout_ms": schema.Int64Attribute{
				MarkdownDescription: "Connection timeout for establishing connection with replica for Hedged requests. Default value: 50 milliseconds.",
				Optional:            true,
			},
			"load_balancing": schema.StringAttribute{
				MarkdownDescription: "Specifies the algorithm of replicas selection that is used for distributed query processing, one of: random, nearest_hostname, in_order, first_or_random, round_robin. Default value: random.",
				Optional:            true,
				Validators:          UserSettings_LoadBalancing_validator,
			},
			"prefer_localhost_replica": schema.BoolAttribute{
				MarkdownDescription: "Enables/disables preferable using the localhost replica when processing distributed queries. Default value: true.",
				Optional:            true,
			},
			"date_time_input_format": schema.StringAttribute{
				MarkdownDescription: "Allows choosing a parser of the text representation of date and time, one of: `best_effort`, `basic`, `best_effort_us`. Default value: `basic`. Cloud default value: `best_effort`.",
				Optional:            true,
			},
			"date_time_output_format": schema.StringAttribute{
				MarkdownDescription: "Allows choosing different output formats of the text representation of date and time, one of: `simple`, `iso`, `unix_timestamp`. Default value: `simple`.",
				Optional:            true,
			},
			"format_regexp": schema.StringAttribute{
				MarkdownDescription: "Regular expression (for Regexp format).",
				Optional:            true,
			},
			"format_regexp_skip_unmatched": schema.BoolAttribute{
				MarkdownDescription: "Skip lines unmatched by regular expression.",
				Optional:            true,
			},
			"input_format_with_names_use_header": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables checking the column order when inserting data.",
				Optional:            true,
			},
			"input_format_null_as_default": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables the initialization of NULL fields with default values, if data type of these fields is not nullable.",
				Optional:            true,
			},
			"insert_quorum_parallel": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables parallelism for quorum INSERT queries.",
				Optional:            true,
			},
			"max_partitions_per_insert_block": schema.Int64Attribute{
				MarkdownDescription: "Limits the maximum number of partitions in a single inserted block.",
				Optional:            true,
			},
			"deduplicate_blocks_in_dependent_materialized_views": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables the deduplication check for materialized views that receive data from Replicated* tables.",
				Optional:            true,
			},
			"any_join_distinct_right_table_keys": schema.BoolAttribute{
				MarkdownDescription: "Enables legacy ClickHouse server behaviour in ANY INNER|LEFT JOIN operations.",
				Optional:            true,
			},
			"join_algorithm": schema.SetAttribute{
				MarkdownDescription: "Specifies which JOIN algorithm to use.",
				ElementType:         types.StringType,
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(UserSettings_JoinAlgorithm_validator...),
				},
			},
		},
	}

}
