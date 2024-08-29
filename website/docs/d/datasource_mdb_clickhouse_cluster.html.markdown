---
layout: "yandex"
page_title: "Yandex: yandex_mdb_clickhouse_cluster"
sidebar_current: "docs-yandex-datasource-mdb-clickhouse-cluster"
description: |-
  Get information about a Yandex Managed ClickHouse cluster.
---

# yandex\_mdb\_clickhouse\_cluster

Get information about a Yandex Managed ClickHouse cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).

## Example Usage

```hcl
data "yandex_mdb_clickhouse_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = "${data.yandex_mdb_clickhouse_cluster.foo.network_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the ClickHouse cluster.

* `name` - (Optional) The name of the ClickHouse cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `network_id` - ID of the network, to which the ClickHouse cluster belongs.
* `created_at` - Creation timestamp of the key.
* `description` - Description of the ClickHouse cluster.
* `labels` - A set of key/value label pairs to assign to the ClickHouse cluster.
* `environment` - Deployment environment of the ClickHouse cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `clickhouse` - Configuration of the ClickHouse subcluster. The structure is documented below.
* `user` - A user of the ClickHouse cluster. The structure is documented below.
* `database` - A database of the ClickHouse cluster. The structure is documented below.
* `host` - A host of the ClickHouse cluster. The structure is documented below.
* `shard_group` - A group of clickhouse shards. The structure is documented below.
* `shard` - A shard of the ClickHouse cluster. The structure is documented below.
* `format_schema` - A set of protobuf or cap'n proto format schemas. The structure is documented below.
* `ml_model` - A group of machine learning models. The structure is documented below.
* `backup_window_start` - Time to start the daily backup, in the UTC timezone. The structure is documented below.
* `access` - Access policy to the ClickHouse cluster. The structure is documented below.
* `zookeeper` - Configuration of the ZooKeeper subcluster. The structure is documented below.
* `sql_user_management` - Enables `admin` user with user management permission.
* `sql_database_management` - Grants `admin` user database management permission.
* `embedded_keeper` - Whether to use ClickHouse Keeper as a coordination system and place it on the same hosts with ClickHouse. If not, it's used ZooKeeper with placement on separate hosts.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.

The `clickhouse` block supports:

* `resources` - Resources allocated to hosts of the ClickHouse subcluster. The structure is documented below.

* `config` - Main ClickHouse cluster configuration. The structure is documented below.

The `zookeeper` block supports:

* `resources` - Resources allocated to hosts of the ZooKeeper subcluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).
* `disk_size` - Volume of the storage available to a host, in gigabytes.
* `disk_type_id` - Type of the storage of hosts.

The `user` block supports:

* `name` - The name of the user.
* `password` - The password of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.
* `settings` - Custom settings for user. The list is documented below.
* `quota` - Set of user quotas. The structure is documented below.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.

The `settings` block supports:

* `readonly` - Restricts permissions for reading data, write data and change settings queries.
* `allow_ddl` - Allows or denies DDL queries.
* `insert_quorum` - Enables the quorum writes.
* `connect_timeout` - Connect timeout in milliseconds on the socket used for communicating with the client.
* `receive_timeout` - Receive timeout in milliseconds on the socket used for communicating with the client.
* `send_timeout` - Send timeout in milliseconds on the socket used for communicating with the client.
* `insert_quorum_timeout` - Write to a quorum timeout in milliseconds.
* `select_sequential_consistency` - Enables or disables sequential consistency for SELECT queries.
* `max_replica_delay_for_distributed_queries` - Disables lagging replicas for distributed queries.
* `fallback_to_stale_replicas_for_distributed_queries` - Forces a query to an out-of-date replica if updated data is not available.
* `replication_alter_partitions_sync` - For ALTER ... ATTACH|DETACH|DROP queries, you can use the replication_alter_partitions_sync setting to set up waiting.
* `distributed_product_mode` - Changes the behaviour of distributed subqueries.
* `distributed_aggregation_memory_efficient` - Determine the behavior of distributed subqueries.
* `distributed_ddl_task_timeout` - Timeout for DDL queries, in milliseconds.
* `skip_unavailable_shards` - Enables or disables silently skipping of unavailable shards.
* `compile` - Enable compilation of queries.
* `min_count_to_compile` - How many times to potentially use a compiled chunk of code before running compilation.
* `compile_expressions` - Turn on expression compilation.
* `min_count_to_compile_expression` - A query waits for expression compilation process to complete prior to continuing execution.
* `max_block_size` - A recommendation for what size of the block (in a count of rows) to load from tables.
* `min_insert_block_size_rows` - Sets the minimum number of rows in the block which can be inserted into a table by an INSERT query.
* `min_insert_block_size_bytes` - Sets the minimum number of bytes in the block which can be inserted into a table by an INSERT query.
* `max_insert_block_size` - The size of blocks (in a count of rows) to form for insertion into a table.
* `min_bytes_to_use_direct_io` - The minimum data volume required for using direct I/O access to the storage disk.
* `use_uncompressed_cache` - Whether to use a cache of uncompressed blocks.
* `merge_tree_max_rows_to_use_cache` - If ClickHouse should read more than merge_tree_max_rows_to_use_cache rows in one query, it doesn’t use the cache of uncompressed blocks.
* `merge_tree_max_bytes_to_use_cache` - If ClickHouse should read more than merge_tree_max_bytes_to_use_cache bytes in one query, it doesn’t use the cache of uncompressed blocks.
* `merge_tree_min_rows_for_concurrent_read` - If the number of rows to be read from a file of a MergeTree table exceeds merge_tree_min_rows_for_concurrent_read then ClickHouse tries to perform a concurrent reading from this file on several threads.
* `merge_tree_min_bytes_for_concurrent_read` - If the number of bytes to read from one file of a MergeTree-engine table exceeds merge_tree_min_bytes_for_concurrent_read, then ClickHouse tries to concurrently read from this file in several threads.
* `max_bytes_before_external_group_by` - Limit in bytes for using memoru for GROUP BY before using swap on disk.
* `max_bytes_before_external_sort` - This setting is equivalent of the max_bytes_before_external_group_by setting, except for it is for sort operation (ORDER BY), not aggregation.
* `group_by_two_level_threshold` - Sets the threshold of the number of keys, after that the two-level aggregation should be used.
* `group_by_two_level_threshold_bytes` - Sets the threshold of the number of bytes, after that the two-level aggregation should be used.
* `priority` - Query priority.
* `max_threads` - The maximum number of query processing threads, excluding threads for retrieving data from remote servers.
* `max_memory_usage` - Limits the maximum memory usage (in bytes) for processing queries on a single server.
* `max_memory_usage_for_user` - Limits the maximum memory usage (in bytes) for processing of user's queries on a single server.
* `max_network_bandwidth` - Limits the speed of the data exchange over the network in bytes per second.
* `max_network_bandwidth_for_user` - Limits the speed of the data exchange over the network in bytes per second.
* `force_index_by_date` - Disables query execution if the index can’t be used by date.
* `force_primary_key` - Disables query execution if indexing by the primary key is not possible.
* `max_rows_to_read` - Limits the maximum number of rows that can be read from a table when running a query.
* `max_bytes_to_read` - Limits the maximum number of bytes (uncompressed data) that can be read from a table when running a query.
* `read_overflow_mode` - Sets behaviour on overflow while read. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
* `max_rows_to_group_by` - Limits the maximum number of unique keys received from aggregation function.
* `group_by_overflow_mode` - Sets behaviour on overflow while GROUP BY operation. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
  * `any` - perform approximate GROUP BY operation by continuing aggregation for the keys that got into the set, but don’t add new keys to the set.
* `max_rows_to_sort` - Limits the maximum number of rows that can be read from a table for sorting.
* `max_bytes_to_sort` - Limits the maximum number of bytes (uncompressed data) that can be read from a table for sorting.
* `sort_overflow_mode` - Sets behaviour on overflow while sort. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
* `max_result_rows` - Limits the number of rows in the result.
* `max_result_bytes` - Limits the number of bytes in the result.
* `result_overflow_mode` - Sets behaviour on overflow in result. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
* `max_rows_in_distinct` - Limits the maximum number of different rows when using DISTINCT.
* `max_bytes_in_distinct` - Limits the maximum size of a hash table in bytes (uncompressed data) when using DISTINCT.
* `distinct_overflow_mode` - Sets behaviour on overflow when using DISTINCT. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
* `max_rows_to_transfer` - Limits the maximum number of rows that can be passed to a remote server or saved in a temporary table when using GLOBAL IN.
* `max_bytes_to_transfer` - Limits the maximum number of bytes (uncompressed data) that can be passed to a remote server or saved in a temporary table when using GLOBAL IN.
* `transfer_overflow_mode` - Sets behaviour on overflow. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
* `max_execution_time` - Limits the maximum query execution time in milliseconds.
* `timeout_overflow_mode` - Sets behaviour on overflow. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
* `max_rows_in_set` - Limit on the number of rows in the set resulting from the execution of the IN section.
* `max_bytes_in_set` - Limit on the number of bytes in the set resulting from the execution of the IN section.
* `set_overflow_mode` - Sets behaviour on overflow in the set resulting. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
* `max_rows_in_join` - Limit on maximum size of the hash table for JOIN, in rows.
* `max_bytes_in_join` - Limit on maximum size of the hash table for JOIN, in bytes.
* `join_overflow_mode` - Sets behaviour on overflow in JOIN. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
* `max_columns_to_read` - Limits the maximum number of columns that can be read from a table in a single query.
* `max_temporary_columns` - Limits the maximum number of temporary columns that must be kept in RAM at the same time when running a query, including constant columns.
* `max_temporary_non_const_columns` - Limits the maximum number of temporary columns that must be kept in RAM at the same time when running a query, excluding constant columns.
* `max_query_size` - The maximum part of a query that can be taken to RAM for parsing with the SQL parser.
* `max_ast_depth` - Maximum abstract syntax tree depth.
* `max_ast_elements` - Maximum abstract syntax tree elements.
* `max_expanded_ast_elements` - Maximum abstract syntax tree depth after after expansion of aliases.
* `min_execution_speed` - Minimal execution speed in rows per second.
* `min_execution_speed_bytes` - Minimal execution speed in bytes per second.
* `count_distinct_implementation` - Specifies which of the uniq* functions should be used to perform the COUNT(DISTINCT …) construction.
* `input_format_values_interpret_expressions` - Enables or disables the full SQL parser if the fast stream parser can’t parse the data.
* `input_format_defaults_for_omitted_fields` - When performing INSERT queries, replace omitted input column values with default values of the respective columns.
* `output_format_json_quote_64bit_integers` - If the value is true, integers appear in quotes when using JSON* Int64 and UInt64 formats (for compatibility with most JavaScript implementations); otherwise, integers are output without the quotes.
* `output_format_json_quote_denormals` - Enables +nan, -nan, +inf, -inf outputs in JSON output format.
* `low_cardinality_allow_in_native_format` - Allows or restricts using the LowCardinality data type with the Native format.
* `empty_result_for_aggregation_by_empty_set` - Allows to retunr empty result.
* `joined_subquery_requires_alias` - Require aliases for subselects and table functions in FROM that more than one table is present.
* `join_use_nulls` - Sets the type of JOIN behaviour. When merging tables, empty cells may appear. ClickHouse fills them differently based on this setting.
* `transform_null_in` - Enables equality of NULL values for IN operator.
* `http_connection_timeout` - Timeout for HTTP connection in milliseconds.
* `http_receive_timeout` - Timeout for HTTP connection in milliseconds.
* `http_send_timeout` - Timeout for HTTP connection in milliseconds.
* `enable_http_compression` - Enables or disables data compression in the response to an HTTP request.
* `send_progress_in_http_headers` - Enables or disables X-ClickHouse-Progress HTTP response headers in clickhouse-server responses.
* `http_headers_progress_interval` - Sets minimal interval between notifications about request process in HTTP header X-ClickHouse-Progress.
* `add_http_cors_header` - Include CORS headers in HTTP responces.
* `quota_mode` - Quota accounting mode.
* `max_concurrent_queries_for_user` - (Optional) The maximum number of concurrent requests per user. Default value: 0 (no limit).
* `memory_profiler_step` - (Optional) Memory profiler step (in bytes).  If the next query step requires more memory than this parameter specifies, the memory profiler collects the allocating stack trace. Values lower than a few megabytes slow down query processing. Default value: 4194304 (4 MB). Zero means disabled memory profiler.
* `memory_profiler_sample_probability` - (Optional) Collect random allocations and deallocations and write them into system.trace_log with 'MemorySample' trace_type. The probability is for every alloc/free regardless to the size of the allocation. Possible values: from 0 to 1. Default: 0.
* `insert_null_as_default` - (Optional) Enables the insertion of default values instead of NULL into columns with not nullable data type. Default value: true.
* `allow_suspicious_low_cardinality_types` - (Optional) Allows specifying LowCardinality modifier for types of small fixed size (8 or less) in CREATE TABLE statements. Enabling this may increase merge times and memory consumption.
* `connect_timeout_with_failover` - (Optional) The timeout in milliseconds for connecting to a remote server for a Distributed table engine, if the ‘shard’ and ‘replica’ sections are used in the cluster definition. If unsuccessful, several attempts are made to connect to various replicas. Default value: 50.
* `allow_introspection_functions` - (Optional) Enables introspections functions for query profiling.
* `async_insert` - (Optional) Enables asynchronous inserts. Disabled by default.
* `async_insert_threads` - (Optional) The maximum number of threads for background data parsing and insertion. If the parameter is set to 0, asynchronous insertions are disabled. Default value: 16.
* `wait_for_async_insert` - (Optional) Enables waiting for processing of asynchronous insertion. If enabled, server returns OK only after the data is inserted.
* `wait_for_async_insert_timeout` - (Optional) The timeout (in seconds) for waiting for processing of asynchronous insertion. Value must be at least 1000 (1 second).
* `async_insert_max_data_size` - (Optional) The maximum size of the unparsed data in bytes collected per query before being inserted. If the parameter is set to 0, asynchronous insertions are disabled. Default value: 100000.
* `async_insert_busy_timeout` - (Optional) The maximum timeout in milliseconds since the first INSERT query before inserting collected data. If the parameter is set to 0, the timeout is disabled. Default value: 200.
* `async_insert_stale_timeout` - (Optional) The maximum timeout in milliseconds since the last INSERT query before dumping collected data. If enabled, the settings prolongs the async_insert_busy_timeout with every INSERT query as long as async_insert_max_data_size is not exceeded.
* `timeout_before_checking_execution_speed` - (Optional) Timeout (in seconds) between checks of execution speed. It is checked that execution speed is not less that specified in min_execution_speed parameter.
  Must be at least 1000.
* `cancel_http_readonly_queries_on_client_close` - (Optional) Cancels HTTP read-only queries (e.g. SELECT) when a client closes the connection without waiting for the response.
  Default value: false.
* `flatten_nested` - (Optional) Sets the data format of a nested columns.
* `max_http_get_redirects` - (Optional) Limits the maximum number of HTTP GET redirect hops for URL-engine tables.
  If the parameter is set to 0 (default), no hops is allowed.
* `input_format_import_nested_json` - (Optional) Enables or disables the insertion of JSON data with nested objects.
* `input_format_parallel_parsing` - (Optional) Enables or disables order-preserving parallel parsing of data formats. Supported only for TSV, TKSV, CSV and JSONEachRow formats.
* `max_read_buffer_size` - (Optional) The maximum size of the buffer to read from the filesystem.
* `max_final_threads` - (Optional) Sets the maximum number of parallel threads for the SELECT query data read phase with the FINAL modifier.
* `local_filesystem_read_method` - (Optional) Method of reading data from local filesystem. Possible values: 
  * `read` - abort query execution, return an error.
  * `pread` - abort query execution, return an error.
  * `pread_threadpool` - stop query execution, return partial result.
* `remote_filesystem_read_method` - (Optional)  Method of reading data from remote filesystem, one of: `read`, `threadpool`.
* `max_read_buffer_size` - (Optional) The maximum size of the buffer to read from the filesystem. 
* `insert_keeper_max_retries` - (Optional) The setting sets the maximum number of retries for ClickHouse Keeper (or ZooKeeper) requests during insert into replicated MergeTree. Only Keeper requests which failed due to network error, Keeper session timeout, or request timeout are considered for retries.
* `max_temporary_data_on_disk_size_for_user` - (Optional) The maximum amount of data consumed by temporary files on disk in bytes for all concurrently running user queries. Zero means unlimited.
* `max_temporary_data_on_disk_size_for_query` - (Optional) The maximum amount of data consumed by temporary files on disk in bytes for all concurrently running queries. Zero means unlimited. 
* `max_parser_depth` - (Optional) Limits maximum recursion depth in the recursive descent parser. Allows controlling the stack size. Zero means unlimited.
* `memory_overcommit_ratio_denominator` - (Optional) It represents soft memory limit in case when hard limit is reached on user level. This value is used to compute overcommit ratio for the query. Zero means skip the query. 
* `memory_overcommit_ratio_denominator_for_user` - (Optional) It represents soft memory limit in case when hard limit is reached on global level. This value is used to compute overcommit ratio for the query. Zero means skip the query.
* `memory_usage_overcommit_max_wait_microseconds` - (Optional) Maximum time thread will wait for memory to be freed in the case of memory overcommit on a user level. If the timeout is reached and memory is not freed, an exception is thrown. 
* `log_query_threads` - (Optional) Setting up query threads logging. Query threads log into the system.query_thread_log table. This setting has effect only when log_queries is true. Queries’ threads run by ClickHouse with this setup are logged according to the rules in the query_thread_log server configuration parameter. Default value: true.
* `max_insert_threads` - (Optional) The maximum number of threads to execute the INSERT SELECT query. Default value: 0.
* `use_hedged_requests` - (Optional) Enables hedged requests logic for remote queries. It allows to establish many connections with different replicas for query. New connection is enabled in case existent connection(s) with replica(s) were not established within hedged_connection_timeout or no data was received within receive_data_timeout. Query uses the first connection which send non empty progress packet (or data packet, if allow_changing_replica_until_first_data_packet); other connections are cancelled. Queries with max_parallel_replicas > 1 are supported. Default value: true.
* `idle_connection_timeout` - (Optional) Timeout to close idle TCP connections after specified number of seconds. Default value: 3600 seconds.
* `hedged_connection_timeout_ms` - (Optional) Connection timeout for establishing connection with replica for Hedged requests. Default value: 50 milliseconds.
* `load_balancing` - (Optional) Specifies the algorithm of replicas selection that is used for distributed query processing, one of: random, nearest_hostname, in_order, first_or_random, round_robin. Default value: random.
* `prefer_localhost_replica` - (Optional) Enables/disables preferable using the localhost replica when processing distributed queries. Default value: true.

The `quota` block supports:

* `interval_duration` - Duration of interval for quota in milliseconds.
* `queries` - The total number of queries.
* `errors` - The number of queries that threw exception.
* `result_rows` - The total number of rows given as the result.
* `read_rows` - The total number of source rows read from tables for running the query, on all remote servers.
* `execution_time` - The total query execution time, in milliseconds (wall time).

The `database` block supports:

* `name` - The name of the database.

The `host` block supports:

* `fqdn` - The fully qualified domain name of the host.
* `type` - The type of the host to be deployed.
* `zone` - The availability zone where the ClickHouse host will be created.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.
* `shard_name` - The name of the shard to which the host belongs.
* `assign_public_ip` - Sets whether the host should get a public IP address on creation.

The `shard` block supports:

* `name` - The name of the shard.
* `weight` - The weight of the shard.
* `resources` - Resources allocated to hosts of the shard. The resources specified for the shard takes precedence over the resources specified for the cluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).
* `disk_size` - Volume of the storage available to a host, in gigabytes.
* `disk_type_id` - Type of the storage of hosts.

The `shard_group` block supports:

* `name` - The name of the shard group, used as cluster name in Distributed tables.
* `description` - Description of the shard group.
* `shard_names` - List of shards names that belong to the shard group.

The `format_schema` block supports:

* `name` - The name of the format schema.
* `type` - Type of the format schema.
* `uri` - Format schema file URL. You can only use format schemas stored in Yandex Object Storage.

The `ml_model` block supports:

* `name` - The name of the ml model.
* `type` - Type of the model.
* `uri` - Model file URL. You can only use models stored in Yandex Object Storage.

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `access` block supports:

* `web_sql` - Allow access for DataLens.
* `data_lens` - Allow access for Web SQL.
* `metrika` - Allow access for Yandex.Metrika.
* `serverless` - Allow access for Serverless.
* `data_transfer` - Allow access for DataTransfer
* `yandex_query` - Allow access for YandexQuery

The `config` block supports:

* `log_level`, `max_connections`, `max_concurrent_queries`, `keep_alive_timeout`, `uncompressed_cache_size`, `mark_cache_size`,
`max_table_size_to_drop`, `max_partition_size_to_drop`, `timezone`, `geobase_uri`, `query_log_retention_size`,
`query_log_retention_time`, `query_thread_log_enabled`, `query_thread_log_retention_size`, `query_thread_log_retention_time`,
`part_log_retention_size`, `part_log_retention_time`, `metric_log_enabled`, `metric_log_retention_size`, `metric_log_retention_time`,
`trace_log_enabled`, `trace_log_retention_size`, `trace_log_retention_time`, `text_log_enabled`, `text_log_retention_size`,
`text_log_retention_time`, `text_log_level`, `background_pool_size`, `background_schedule_pool_size`, `background_fetches_pool_size`, `background_message_broker_schedule_pool_size`, `background_merges_mutations_concurrency_ratio`,  `default_database`,
`total_memory_profiler_step`, `dictionaries_lazy_load` - ClickHouse server parameters. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts/settings-list).

* `merge_tree` - MergeTree engine configuration. The structure is documented below.
* `kafka` - Kafka connection configuration. The structure is documented below.
* `kafka_topic` - Kafka topic connection configuration. The structure is documented below.
* `compression` - Data compression configuration. The structure is documented below.
* `rabbitmq` - RabbitMQ connection configuration. The structure is documented below.
* `graphite_rollup` - Graphite rollup configuration. The structure is documented below.
* `query_masking_rules` - Query masking rules configuration. The structure is documented below.
* `query_cache` - Query cache configuration. The structure is documented below.

The `merge_tree` block supports:

* `replicated_deduplication_window` - Replicated deduplication window: Number of recent hash blocks that ZooKeeper will store (the old ones will be deleted).
* `replicated_deduplication_window_seconds` - Replicated deduplication window seconds: Time during which ZooKeeper stores the hash blocks (the old ones wil be deleted).
* `parts_to_delay_insert` - Parts to delay insert: Number of active data parts in a table, on exceeding which ClickHouse starts artificially reduce the rate of inserting data into the table.
* `parts_to_throw_insert` - Parts to throw insert: Threshold value of active data parts in a table, on exceeding which ClickHouse throws the 'Too many parts ...' exception.
* `max_replicated_merges_in_queue` - Max replicated merges in queue: Maximum number of merge tasks that can be in the ReplicatedMergeTree queue at the same time.
* `number_of_free_entries_in_pool_to_lower_max_size_of_merge` - Number of free entries in pool to lower max size of merge: Threshold value of free entries in the pool. If the number of entries in the pool falls below this value, ClickHouse reduces the maximum size of a data part to merge. This helps handle small merges faster, rather than filling the pool with lengthy merges.
* `max_bytes_to_merge_at_min_space_in_pool` - Max bytes to merge at min space in pool: Maximum total size of a data part to merge when the number of free threads in the background pool is minimum.
* `min_bytes_for_wide_part` - (Optional) Minimum number of bytes in a data part that can be stored in Wide format. You can set one, both or none of these settings.
* `min_rows_for_wide_part` - (Optional) Minimum number of rows in a data part that can be stored in Wide format. You can set one, both or none of these settings.
* `ttl_only_drop_parts` - (Optional) Enables or disables complete dropping of data parts where all rows are expired in MergeTree tables.
* `merge_with_ttl_timeout` - (Optional) Minimum delay in seconds before repeating a merge with delete TTL. Default value: 14400 seconds (4 hours).
* `merge_with_recompression_ttl_timeout` - (Optional) Minimum delay in seconds before repeating a merge with recompression TTL. Default value: 14400 seconds (4 hours).
* `max_parts_in_total` - (Optional) Maximum number of parts in all partitions.
* `max_number_of_merges_with_ttl_in_pool` - (Optional) When there is more than specified number of merges with TTL entries in pool, do not assign new merge with TTL. 
* `cleanup_delay_period` - (Optional) Minimum period to clean old queue logs, blocks hashes and parts.
* `number_of_free_entries_in_pool_to_execute_mutation` - (Optional) 
* `max_avg_part_size_for_too_many_parts` - (Optional) The `too many parts` check according to `parts_to_delay_insert` and `parts_to_throw_insert` will be active only if the average part size (in the relevant partition) is not larger than the specified threshold. If it is larger than the specified threshold, the INSERTs will be neither delayed or rejected. This allows to have hundreds of terabytes in a single table on a single server if the parts are successfully merged to larger parts. This does not affect the thresholds on inactive parts or total parts.
* `min_age_to_force_merge_seconds` - (Optional) Merge parts if every part in the range is older than the value of `min_age_to_force_merge_seconds`.
* `min_age_to_force_merge_on_partition_only` - (Optional) Whether min_age_to_force_merge_seconds should be applied only on the entire partition and not on subset.
* `merge_selecting_sleep_ms` - (Optional) Sleep time for merge selecting when no part is selected. A lower setting triggers selecting tasks in background_schedule_pool frequently, which results in a large number of requests to ClickHouse Keeper in large-scale clusters.
* `merge_max_block_size` - (Optional) The number of rows that are read from the merged parts into memory. Default value: 8192.
* `check_sample_column_is_correct` - (Optional) Enables the check at table creation, that the data type of a column for sampling or sampling expression is correct. The data type must be one of unsigned integer types: UInt8, UInt16, UInt32, UInt64. Default value: true.
* `max_merge_selecting_sleep_ms` - (Optional) Maximum sleep time for merge selecting, a lower setting will trigger selecting tasks in background_schedule_pool frequently which result in large amount of requests to zookeeper in large-scale clusters. Default value: 60000 milliseconds (60 seconds).
* `max_cleanup_delay_period` - (Optional) Maximum period to clean old queue logs, blocks hashes and parts. Default value: 300 seconds.

The `kafka` block supports:

* `security_protocol` - Security protocol used to connect to kafka server.
* `sasl_mechanism` - SASL mechanism used in kafka authentication.
* `sasl_username` - Username on kafka server.
* `sasl_password` - User password on kafka server.
* `enable_ssl_certificate_verification` - (Optional) enable verification of SSL certificates.
* `max_poll_interval_ms` - (Optional) Maximum allowed time between calls to consume messages (e.g., rd_kafka_consumer_poll()) for high-level consumers. If this interval is exceeded the consumer is considered failed and the group will rebalance in order to reassign the partitions to another consumer group member.
* `session_timeout_ms` - (Optional) Client group session and failure detection timeout. The consumer sends periodic heartbeats (heartbeat.interval.ms) to indicate its liveness to the broker. If no hearts are received by the broker for a group member within the session timeout, the broker will remove the consumer from the group and trigger a rebalance. 
* `debug` - (Optional) A comma-separated list of debug contexts to enable.
* `auto_offset_reset` - (Optional) Action to take when there is no initial offset in offset store or the desired offset is out of range: 'smallest','earliest' - automatically reset the offset to the smallest offset, 'largest','latest' - automatically reset the offset to the largest offset, 'error' - trigger an error (ERR__AUTO_OFFSET_RESET) which is retrieved by consuming messages and checking 'message->err'.

The `kafka_topic` block supports:

* `name` - Kafka topic name.
* `settings` - Kafka connection settngs sanem as `kafka` block.

The `compression` block supports:

* `method` - Method: Compression method. Two methods are available: LZ4 and zstd.
* `min_part_size` - Min part size: Minimum size (in bytes) of a data part in a table. ClickHouse only applies the rule to tables with data parts greater than or equal to the Min part size value.
* `min_part_size_ratio` - Min part size ratio: Minimum table part size to total table size ratio. ClickHouse only applies the rule to tables in which this ratio is greater than or equal to the Min part size ratio value.
* `level` - (Optional) Compression level for `ZSTD` method.

The `rabbitmq` block supports:

* `username` - RabbitMQ username.
* `password` - RabbitMQ user password.
* `vhost` - (Optional) RabbitMQ vhost. Default: '\'.

The `graphite_rollup` block supports:

* `name` - Graphite rollup configuration name.
* `pattern` - Set of thinning rules.
  * `function` - Aggregation function name.
  * `regexp` - Regular expression that the metric name must match.
  * `retention` - Retain parameters.
    * `age` - Minimum data age in seconds.
    * `precision` - Accuracy of determining the age of the data in seconds.
* `path_column_name` - (Optional) The name of the column storing the metric name (Graphite sensor). Default value: Path.
* `time_column_name` - (Optional) The name of the column storing the time of measuring the metric. Default value: Time.
* `value_column_name` - (Optional) The name of the column storing the value of the metric at the time set in time_column_name. Default value: Value.
* `version_column_name` - (Optional) The name of the column storing the version of the metric. Default value: Timestamp.

The `query_masking_rules` block supports:

* `name` - (Optional) Name for the rule.
* `regexp` - (Required) RE2 compatible regular expression.
* `replace` - (Optional) Substitution string for sensitive data. Default value: six asterisks.

The `query_cache` block supports:

* `max_size_in_bytes` - (Optional) The maximum cache size in bytes. 0 means the query cache is disabled. Default value: 1073741824 (1 GiB).
* `max_entries` - (Optional) The maximum number of SELECT query results stored in the cache. Default value: 1024.
* `max_entry_size_in_bytes` - (Optional) The maximum size in bytes SELECT query results may have to be saved in the cache. Default value: 1048576 (1 MiB).
* `max_entry_size_in_rows` - (Optional) The maximum number of rows SELECT query results may have to be saved in the cache. Default value: 30000000 (30 mil).

The `cloud_storage` block supports:

* `enabled` - (Required) Whether to use Yandex Object Storage for storing ClickHouse data. Can be either `true` or `false`.
* `move_factor` - Sets the minimum free space ratio in the cluster storage. If the free space is lower than this value, the data is transferred to Yandex Object Storage. Acceptable values are 0 to 1, inclusive.
* `data_cache_enabled` - Enables temporary storage in the cluster repository of data requested from the object repository.
* `data_cache_max_size` - Defines the maximum amount of memory (in bytes) allocated in the cluster storage for temporary storage of data requested from the object storage.
* `prefer_not_to_merge` - (Optional) Disables merging of data parts in `Yandex Object Storage`.

The `maintenance_window` block supports:

* `type` - Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.
