---
subcategory: "Managed Service for ClickHouse"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a ClickHouse cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a ClickHouse cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).

## Example usage

{{ tffile "examples/mdb_clickhouse_cluster/r_mdb_clickhouse_cluster_1.tf" }}

Example of creating a HA ClickHouse Cluster.

{{ tffile "examples/mdb_clickhouse_cluster/r_mdb_clickhouse_cluster_2.tf" }}

Example of creating a sharded ClickHouse Cluster.

{{ tffile "examples/mdb_clickhouse_cluster/r_mdb_clickhouse_cluster_3.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the ClickHouse cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the ClickHouse cluster belongs.

* `environment` - (Required) Deployment environment of the ClickHouse cluster. Can be either `PRESTABLE` or `PRODUCTION`.

* `clickhouse` - (Required) Configuration of the ClickHouse subcluster. The structure is documented below.

* `user` - (Required) A user of the ClickHouse cluster. The structure is documented below.

* `database` - (Required) A database of the ClickHouse cluster. The structure is documented below.

* `host` - (Required) A host of the ClickHouse cluster. The structure is documented below.

---

* `version` - (Optional) Version of the ClickHouse server software.

* `description` - (Optional) Description of the ClickHouse cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the ClickHouse cluster.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC timezone. The structure is documented below.

* `access` - (Optional) Access policy to the ClickHouse cluster. The structure is documented below.

* `zookeeper` - (Optional) Configuration of the ZooKeeper subcluster. The structure is documented below.

* `shard_group` - (Optional) A group of clickhouse shards. The structure is documented below.

* `format_schema` - (Optional) A set of protobuf or capnproto format schemas. The structure is documented below.

* `ml_model` - (Optional) A group of machine learning models. The structure is documented below

* `admin_password` - (Optional) A password used to authorize as user `admin` when `sql_user_management` enabled.

* `sql_user_management` - (Optional, ForceNew) Enables `admin` user with user management permission.

* `sql_database_management` - (Optional, ForceNew) Grants `admin` user database management permission.

* `embedded_keeper` - (Optional, ForceNew) Whether to use ClickHouse Keeper as a coordination system and place it on the same hosts with ClickHouse. If not, it's used ZooKeeper with placement on separate hosts.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `copy_schema_on_new_hosts` - (Optional) Whether to copy schema on new ClickHouse hosts.

* `service_account_id` - (Optional) ID of the service account used for access to Yandex Object Storage.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster. Can be either `true` or `false`.

* `backup_retain_period_days` - (Optional) The period in days during which backups are stored.

---

The `clickhouse` block supports:

* `resources` - (Required) Resources allocated to hosts of the ClickHouse subcluster. The structure is documented below.

* `config` - (Optional) Main ClickHouse cluster configuration.

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a ClickHouse host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).

* `disk_size` - (Required) Volume of the storage available to a ClickHouse host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of ClickHouse hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).

The `zookeeper` block supports:

* `resources` - (Optional) Resources allocated to hosts of the ZooKeeper subcluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - (Optional) The ID of the preset for computational resources available to a ZooKeeper host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).

* `disk_size` - (Optional) Volume of the storage available to a ZooKeeper host, in gigabytes.

* `disk_type_id` - (Optional) Type of the storage of ZooKeeper hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).

The `user` block supports:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

* `settings` - (Optional) Custom settings for user. The list is documented below.

* `quota` - (Optional) Set of user quotas. The structure is documented below.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

The `settings` block supports:

* `readonly` - (Optional) Restricts permissions for reading data, write data and change settings queries.

* `allow_ddl` - (Optional) Allows or denies DDL queries.

* `insert_quorum` - (Optional) Enables the quorum writes.

* `connect_timeout` - (Optional) Connect timeout in milliseconds on the socket used for communicating with the client.

* `receive_timeout` - (Optional) Receive timeout in milliseconds on the socket used for communicating with the client.

* `send_timeout` - (Optional) Send timeout in milliseconds on the socket used for communicating with the client.

* `insert_quorum_timeout` - (Optional) Write to a quorum timeout in milliseconds.

* `insert_quorum_parallel` - Enables or disables parallelism for quorum INSERT queries.

* `select_sequential_consistency` - (Optional) Enables or disables sequential consistency for SELECT queries.

* `deduplicate_blocks_in_dependent_materialized_views` - Enables or disables the deduplication check for materialized views that receive data from Replicated* tables.

* `max_replica_delay_for_distributed_queries` - (Optional) Disables lagging replicas for distributed queries.

* `fallback_to_stale_replicas_for_distributed_queries` - (Optional) Forces a query to an out-of-date replica if updated data is not available.

* `replication_alter_partitions_sync` - (Optional) For ALTER ... ATTACH|DETACH|DROP queries, you can use the replication_alter_partitions_sync setting to set up waiting.

* `distributed_product_mode` - (Optional) Changes the behaviour of distributed subqueries.

* `distributed_aggregation_memory_efficient` - (Optional) Determine the behavior of distributed subqueries.

* `distributed_ddl_task_timeout` - (Optional) Timeout for DDL queries, in milliseconds.

* `skip_unavailable_shards` - (Optional) Enables or disables silently skipping of unavailable shards.

* `compile` - (Optional) Enable compilation of queries.

* `min_count_to_compile` - (Optional) How many times to potentially use a compiled chunk of code before running compilation.

* `compile_expressions` - (Optional) Turn on expression compilation.

* `min_count_to_compile_expression` - (Optional) A query waits for expression compilation process to complete prior to continuing execution.

* `max_block_size` - (Optional) A recommendation for what size of the block (in a count of rows) to load from tables.

* `min_insert_block_size_rows` - (Optional) Sets the minimum number of rows in the block which can be inserted into a table by an INSERT query.

* `min_insert_block_size_bytes` - (Optional) Sets the minimum number of bytes in the block which can be inserted into a table by an INSERT query.

* `max_insert_block_size` - (Optional) The size of blocks (in a count of rows) to form for insertion into a table.

* `min_bytes_to_use_direct_io` - (Optional) The minimum data volume required for using direct I/O access to the storage disk.

* `use_uncompressed_cache` - (Optional) Whether to use a cache of uncompressed blocks.

* `merge_tree_max_rows_to_use_cache` - (Optional) If ClickHouse should read more than merge_tree_max_rows_to_use_cache rows in one query, it doesn’t use the cache of uncompressed blocks.

* `merge_tree_max_bytes_to_use_cache` - (Optional) If ClickHouse should read more than merge_tree_max_bytes_to_use_cache bytes in one query, it doesn’t use the cache of uncompressed blocks.

* `merge_tree_min_rows_for_concurrent_read` - (Optional) If the number of rows to be read from a file of a MergeTree table exceeds merge_tree_min_rows_for_concurrent_read then ClickHouse tries to perform a concurrent reading from this file on several threads.

* `merge_tree_min_bytes_for_concurrent_read` - (Optional) If the number of bytes to read from one file of a MergeTree-engine table exceeds merge_tree_min_bytes_for_concurrent_read, then ClickHouse tries to concurrently read from this file in several threads.

* `max_bytes_before_external_group_by` - (Optional) Limit in bytes for using memoru for GROUP BY before using swap on disk.

* `max_bytes_before_external_sort` - (Optional) This setting is equivalent of the max_bytes_before_external_group_by setting, except for it is for sort operation (ORDER BY), not aggregation.

* `group_by_two_level_threshold` - (Optional) Sets the threshold of the number of keys, after that the two-level aggregation should be used.

* `group_by_two_level_threshold_bytes` - (Optional) Sets the threshold of the number of bytes, after that the two-level aggregation should be used.

* `priority` - (Optional) Query priority.

* `max_threads` - (Optional) The maximum number of query processing threads, excluding threads for retrieving data from remote servers.

* `max_memory_usage` - (Optional) Limits the maximum memory usage (in bytes) for processing queries on a single server.

* `max_memory_usage_for_user` - (Optional) Limits the maximum memory usage (in bytes) for processing of user's queries on a single server.

* `max_network_bandwidth` - (Optional) Limits the speed of the data exchange over the network in bytes per second.

* `max_network_bandwidth_for_user` - (Optional) Limits the speed of the data exchange over the network in bytes per second.

* `force_index_by_date` - (Optional) Disables query execution if the index can’t be used by date.

* `force_primary_key` - (Optional) Disables query execution if indexing by the primary key is not possible.

* `max_rows_to_read` - (Optional) Limits the maximum number of rows that can be read from a table when running a query.

* `max_bytes_to_read` - (Optional) Limits the maximum number of bytes (uncompressed data) that can be read from a table when running a query.

* `read_overflow_mode` - (Optional) Sets behaviour on overflow while read. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.

* `max_rows_to_group_by` - (Optional) Limits the maximum number of unique keys received from aggregation function.

* `group_by_overflow_mode` - (Optional) Sets behaviour on overflow while GROUP BY operation. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.
  * `any` - perform approximate GROUP BY operation by continuing aggregation for the keys that got into the set, but don’t add new keys to the set.

* `max_rows_to_sort` - (Optional) Limits the maximum number of rows that can be read from a table for sorting.

* `max_bytes_to_sort` - (Optional) Limits the maximum number of bytes (uncompressed data) that can be read from a table for sorting.

* `sort_overflow_mode` - (Optional) Sets behaviour on overflow while sort. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.

* `max_result_rows` - (Optional) Limits the number of rows in the result.

* `max_result_bytes` - (Optional) Limits the number of bytes in the result.

* `result_overflow_mode` - (Optional) Sets behaviour on overflow in result. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.

* `max_rows_in_distinct` - (Optional) Limits the maximum number of different rows when using DISTINCT.

* `max_bytes_in_distinct` - (Optional) Limits the maximum size of a hash table in bytes (uncompressed data) when using DISTINCT.

* `distinct_overflow_mode` - (Optional) Sets behaviour on overflow when using DISTINCT. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.

* `max_rows_to_transfer` - (Optional) Limits the maximum number of rows that can be passed to a remote server or saved in a temporary table when using GLOBAL IN.

* `max_bytes_to_transfer` - (Optional) Limits the maximum number of bytes (uncompressed data) that can be passed to a remote server or saved in a temporary table when using GLOBAL IN.

* `transfer_overflow_mode` - (Optional) Sets behaviour on overflow. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.

* `max_execution_time` - (Optional) Limits the maximum query execution time in milliseconds.

* `timeout_overflow_mode` - (Optional) Sets behaviour on overflow. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.

* `max_rows_in_set` - (Optional) Limit on the number of rows in the set resulting from the execution of the IN section.

* `max_bytes_in_set` - (Optional) Limit on the number of bytes in the set resulting from the execution of the IN section.

* `set_overflow_mode` - (Optional) Sets behaviour on overflow in the set resulting. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.

* `max_rows_in_join` - (Optional) Limit on maximum size of the hash table for JOIN, in rows.

* `max_bytes_in_join` - (Optional) Limit on maximum size of the hash table for JOIN, in bytes.

* `join_overflow_mode` - (Optional) Sets behaviour on overflow in JOIN. Possible values:
  * `throw` - abort query execution, return an error.
  * `break` - stop query execution, return partial result.

* `join_algorithm` - (Optional) Specifies which JOIN algorithm is used. Possible values:
  * `hash` - hash join algorithm is used. The most generic implementation that supports all combinations of kind and strictness and multiple join keys that are combined with OR in the JOIN ON section.
  * `parallel_hash` - a variation of hash join that splits the data into buckets and builds several hashtables instead of one concurrently to speed up this process.
  * `partial_merge` - a variation of the sort-merge algorithm, where only the right table is fully sorted.
  * `direct` - this algorithm can be applied when the storage for the right table supports key-value requests.
  * `auto` - when set to auto, hash join is tried first, and the algorithm is switched on the fly to another algorithm if the memory limit is violated.
  * `full_sorting_merge` - sort-merge algorithm with full sorting joined tables before joining.
  * `prefer_partial_merge` - clickHouse always tries to use partial_merge join if possible, otherwise, it uses hash. Deprecated, same as partial_merge,hash.

* `any_join_distinct_right_table_keys` - enables legacy ClickHouse server behaviour in ANY INNER|LEFT JOIN operations.

* `max_columns_to_read` - (Optional) Limits the maximum number of columns that can be read from a table in a single query.

* `max_temporary_columns` - (Optional) Limits the maximum number of temporary columns that must be kept in RAM at the same time when running a query, including constant columns.

* `max_temporary_non_const_columns` - (Optional) Limits the maximum number of temporary columns that must be kept in RAM at the same time when running a query, excluding constant columns.

* `max_query_size` - (Optional) The maximum part of a query that can be taken to RAM for parsing with the SQL parser.

* `max_ast_depth` - (Optional) Maximum abstract syntax tree depth.

* `max_ast_elements` - (Optional) Maximum abstract syntax tree elements.

* `max_expanded_ast_elements` - (Optional) Maximum abstract syntax tree depth after after expansion of aliases.

* `min_execution_speed` - (Optional) Minimal execution speed in rows per second.

* `min_execution_speed_bytes` - (Optional) Minimal execution speed in bytes per second.

* `count_distinct_implementation` - (Optional) Specifies which of the uniq* functions should be used to perform the COUNT(DISTINCT …) construction.

* `input_format_values_interpret_expressions` - (Optional) Enables or disables the full SQL parser if the fast stream parser can’t parse the data.

* `input_format_defaults_for_omitted_fields` - (Optional) When performing INSERT queries, replace omitted input column values with default values of the respective columns.

* `input_format_null_as_default` - Enables or disables the initialization of NULL fields with default values, if data type of these fields is not nullable.

* `input_format_with_names_use_header` - Enables or disables checking the column order when inserting data.

* `output_format_json_quote_64bit_integers` - (Optional) If the value is true, integers appear in quotes when using JSON* Int64 and UInt64 formats (for compatibility with most JavaScript implementations); otherwise, integers are output without the quotes.

* `output_format_json_quote_denormals` - (Optional) Enables +nan, -nan, +inf, -inf outputs in JSON output format.

* `low_cardinality_allow_in_native_format` - (Optional) Allows or restricts using the LowCardinality data type with the Native format.

* `empty_result_for_aggregation_by_empty_set` - (Optional) Allows to retunr empty result.

* `joined_subquery_requires_alias` - (Optional) Require aliases for subselects and table functions in FROM that more than one table is present.

* `join_use_nulls` - (Optional) Sets the type of JOIN behaviour. When merging tables, empty cells may appear. ClickHouse fills them differently based on this setting.

* `transform_null_in` - (Optional) Enables equality of NULL values for IN operator.

* `http_connection_timeout` - (Optional) Timeout for HTTP connection in milliseconds.

* `http_receive_timeout` - (Optional) Timeout for HTTP connection in milliseconds.

* `http_send_timeout` - (Optional) Timeout for HTTP connection in milliseconds.

* `enable_http_compression` - (Optional) Enables or disables data compression in the response to an HTTP request.

* `send_progress_in_http_headers` - (Optional) Enables or disables X-ClickHouse-Progress HTTP response headers in clickhouse-server responses.

* `http_headers_progress_interval` - (Optional) Sets minimal interval between notifications about request process in HTTP header X-ClickHouse-Progress.

* `add_http_cors_header` - (Optional) Include CORS headers in HTTP responces.

* `quota_mode` - (Optional) Quota accounting mode.

* `max_concurrent_queries_for_user` - (Optional) The maximum number of concurrent requests per user. Default value: 0 (no limit).

* `memory_profiler_step` - (Optional) Memory profiler step (in bytes). If the next query step requires more memory than this parameter specifies, the memory profiler collects the allocating stack trace. Values lower than a few megabytes slow down query processing. Default value: 4194304 (4 MB). Zero means disabled memory profiler.

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

* `timeout_before_checking_execution_speed` - (Optional) Timeout (in seconds) between checks of execution speed. It is checked that execution speed is not less that specified in min_execution_speed parameter. Must be at least 1000.

* `cancel_http_readonly_queries_on_client_close` - (Optional) Cancels HTTP read-only queries (e.g. SELECT) when a client closes the connection without waiting for the response. Default value: false.

* `flatten_nested` - (Optional) Sets the data format of a nested columns.

* `format_regexp` - (Optional) Regular expression (for Regexp format).

* `format_regexp_skip_unmatched` - (Optional) Skip lines unmatched by regular expression.

* `max_http_get_redirects` - (Optional) Limits the maximum number of HTTP GET redirect hops for URL-engine tables.

* `input_format_import_nested_json` - (Optional) Enables or disables the insertion of JSON data with nested objects.

* `input_format_parallel_parsing` - (Optional) Enables or disables order-preserving parallel parsing of data formats. Supported only for TSV, TKSV, CSV and JSONEachRow formats.

* `max_read_buffer_size` - (Optional) The maximum size of the buffer to read from the filesystem.

* `max_final_threads` - (Optional) Sets the maximum number of parallel threads for the SELECT query data read phase with the FINAL modifier.

* `local_filesystem_read_method` - (Optional) Method of reading data from local filesystem. Possible values:
  * `read` - abort query execution, return an error.
  * `pread` - abort query execution, return an error.
  * `pread_threadpool` - stop query execution, return partial result. If the parameter is set to 0 (default), no hops is allowed.

* `remote_filesystem_read_method` - (Optional) Method of reading data from remote filesystem, one of: `read`, `threadpool`.

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

* `date_time_input_format` - (Optional) Allows choosing a parser of the text representation of date and time, one of: `best_effort`, `basic`, `best_effort_us`. Default value: `basic`. Cloud default value: `best_effort`.

* `date_time_output_format` - (Optional) Allows choosing different output formats of the text representation of date and time, one of: `simple`, `iso`, `unix_timestamp`. Default value: `simple`.

The `quota` block supports:

* `interval_duration` - (Required) Duration of interval for quota in milliseconds.

* `queries` - (Optional) The total number of queries.

* `errors` - (Optional) The number of queries that threw exception.

* `result_rows` - (Optional) The total number of rows given as the result.

* `read_rows` - (Optional) The total number of source rows read from tables for running the query, on all remote servers.

* `execution_time` - (Optional) The total query execution time, in milliseconds (wall time).

The `database` block supports:

* `name` - (Required) The name of the database.

The `host` block supports:

* `fqdn` - (Computed) The fully qualified domain name of the host.

* `type` - (Required) The type of the host to be deployed. Can be either `CLICKHOUSE` or `ZOOKEEPER`.

* `zone` - (Required) The availability zone where the ClickHouse host will be created. For more information see [the official documentation](https://yandex.cloud/docs/overview/concepts/geo-scope).

* `subnet_id` (Optional) - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `shard_name` (Optional) - The name of the shard to which the host belongs.

* `assign_public_ip` (Optional) - Sets whether the host should get a public IP address on creation. Can be either `true` or `false`.

The `shard` block supports:

* `name` - (Required) The name of shard.

* `weight` - (Optional) The weight of shard.

* `resources` - (Optional) Resources allocated to host of the shard. The resources specified for the shard takes precedence over the resources specified for the cluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).
* `disk_size` - Volume of the storage available to a host, in gigabytes.
* `disk_type_id` - Type of the storage of hosts.

The `shard_group` block supports:

* `name` (Required) - The name of the shard group, used as cluster name in Distributed tables.

* `description` (Optional) - Description of the shard group.

* `shard_names` (Required) - List of shards names that belong to the shard group.

The `format_schema` block supports:

* `name` - (Required) The name of the format schema.

* `type` - (Required) Type of the format schema.

* `uri` - (Required) Format schema file URL. You can only use format schemas stored in Yandex Object Storage.

The `ml_model` block supports:

* `name` - (Required) The name of the ml model.

* `type` - (Required) Type of the model.

* `uri` - (Required) Model file URL. You can only use models stored in Yandex Object Storage.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started.

* `minutes` - (Optional) The minute at which backup will be started.

The `access` block supports:

* `web_sql` - (Optional) Allow access for Web SQL. Can be either `true` or `false`.

* `data_lens` - (Optional) Allow access for DataLens. Can be either `true` or `false`.

* `metrika` - (Optional) Allow access for Yandex.Metrika. Can be either `true` or `false`.

* `serverless` - (Optional) Allow access for Serverless. Can be either `true` or `false`.

* `yandex_query` - (Optional) Allow access for YandexQuery. Can be either `true` or `false`.

* `data_transfer` - (Optional) Allow access for DataTransfer. Can be either `true` or `false`.

The `config` block supports:

* `log_level`, `max_connections`, `max_concurrent_queries`, `keep_alive_timeout`, `uncompressed_cache_size`, `mark_cache_size`, `max_table_size_to_drop`, `max_partition_size_to_drop`, `timezone`, `geobase_uri`, `query_log_retention_size`, `query_log_retention_time`, `query_thread_log_enabled`, `query_thread_log_retention_size`, `query_thread_log_retention_time`, `part_log_retention_size`, `part_log_retention_time`, `metric_log_enabled`, `metric_log_retention_size`, `metric_log_retention_time`, `trace_log_enabled`, `trace_log_retention_size`, `trace_log_retention_time`, `text_log_enabled`, `text_log_retention_size`, `text_log_retention_time`, `text_log_level`, `background_pool_size`, `background_schedule_pool_size`, `background_fetches_pool_size`, `background_message_broker_schedule_pool_size`,`background_merges_mutations_concurrency_ratio`, `background_move_pool_size`, `background_distributed_schedule_pool_size`, `background_common_pool_size` `default_database`, `total_memory_profiler_step`, `dictionaries_lazy_load`, `opentelemetry_span_log_enabled`, `opentelemetry_span_log_retention_size`, `opentelemetry_span_log_retention_time`, `query_views_log_enabled`, `query_views_log_retention_size`, `query_views_log_retention_time`, `asynchronous_metric_log_enabled`, `asynchronous_metric_log_retention_size`, `asynchronous_metric_log_retention_time`, `session_log_enabled`, `session_log_retention_size`, `session_log_retention_time`, `zookeeper_log_enabled`, `zookeeper_log_retention_size`, `zookeeper_log_retention_time`, `asynchronous_insert_log_enabled`, `asynchronous_insert_log_retention_size`, `asynchronous_insert_log_retention_time` - (Optional) ClickHouse server parameters. For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/settings-list).

* `merge_tree` - (Optional) MergeTree engine configuration. The structure is documented below.
* `kafka` - (Optional) Kafka connection configuration. The structure is documented below.
* `kafka_topic` - (Optional) Kafka topic connection configuration. The structure is documented below.
* `compression` - (Optional) Data compression configuration. The structure is documented below.
* `rabbitmq` - (Optional) RabbitMQ connection configuration. The structure is documented below.
* `graphite_rollup` - (Optional) Graphite rollup configuration. The structure is documented below.
* `query_masking_rules` - (Optional) Query masking rules configuration. The structure is documented below.
* `query_cache` - (Optional) Query cache configuration. The structure is documented below.
* `jdbc_bridge` - (Optional) JDBC bridge configuration. The structure is documented below.

The `merge_tree` block supports:

* `replicated_deduplication_window` - (Optional) Replicated deduplication window: Number of recent hash blocks that ZooKeeper will store (the old ones will be deleted).
* `replicated_deduplication_window_seconds` - (Optional) Replicated deduplication window seconds: Time during which ZooKeeper stores the hash blocks (the old ones wil be deleted).
* `parts_to_delay_insert` - (Optional) Parts to delay insert: Number of active data parts in a table, on exceeding which ClickHouse starts artificially reduce the rate of inserting data into the table.
* `parts_to_throw_insert` - (Optional) Parts to throw insert: Threshold value of active data parts in a table, on exceeding which ClickHouse throws the 'Too many parts ...' exception.
* `inactive_parts_to_delay_insert` - (Optional) If the number of inactive parts in a single partition in the table at least that many the inactive_parts_to_delay_insert value, an INSERT artificially slows down. It is useful when a server fails to clean up parts quickly enough.
* `inactive_parts_to_throw_insert` - (Optional) If the number of inactive parts in a single partition more than the inactive_parts_to_throw_insert value, INSERT is interrupted with the "Too many inactive parts (N). Parts cleaning are processing significantly slower than inserts" exception.
* `max_replicated_merges_in_queue` - (Optional) Max replicated merges in queue: Maximum number of merge tasks that can be in the ReplicatedMergeTree queue at the same time.
* `number_of_free_entries_in_pool_to_lower_max_size_of_merge` - (Optional) Number of free entries in pool to lower max size of merge: Threshold value of free entries in the pool. If the number of entries in the pool falls below this value, ClickHouse reduces the maximum size of a data part to merge. This helps handle small merges faster, rather than filling the pool with lengthy merges.
* `max_bytes_to_merge_at_min_space_in_pool` - (Optional) Max bytes to merge at min space in pool: Maximum total size of a data part to merge when the number of free threads in the background pool is minimum.
* `max_bytes_to_merge_at_max_space_in_pool` - (Optional) The maximum total parts size (in bytes) to be merged into one part, if there are enough resources available. max_bytes_to_merge_at_max_space_in_pool -- roughly corresponds to the maximum possible part size created by an automatic background merge.
* `min_bytes_for_wide_part` - (Optional) Minimum number of bytes in a data part that can be stored in Wide format. You can set one, both or none of these settings.
* `min_rows_for_wide_part` - (Optional) Minimum number of rows in a data part that can be stored in Wide format. You can set one, both or none of these settings.
* `ttl_only_drop_parts` - (Optional) Enables zero-copy replication when a replica is located on a remote filesystem.
* `allow_remote_fs_zero_copy_replication` - (Optional) When this setting has a value greater than zero only a single replica starts the merge immediately if merged part on shared storage and allow_remote_fs_zero_copy_replication is enabled.
* `merge_with_ttl_timeout` - (Optional) Minimum delay in seconds before repeating a merge with delete TTL. Default value: 14400 seconds (4 hours).
* `merge_with_recompression_ttl_timeout` - (Optional) Minimum delay in seconds before repeating a merge with recompression TTL. Default value: 14400 seconds (4 hours).
* `max_parts_in_total` - (Optional) Maximum number of parts in all partitions.
* `max_number_of_merges_with_ttl_in_pool` - (Optional) When there is more than specified number of merges with TTL entries in pool, do not assign new merge with TTL.
* `cleanup_delay_period` - (Optional) Minimum period to clean old queue logs, blocks hashes and parts.
* `number_of_free_entries_in_pool_to_execute_mutation` - (Optional) When there is less than specified number of free entries in pool, do not execute part mutations. This is to leave free threads for regular merges and avoid "Too many parts". Default value: 20.
* `max_avg_part_size_for_too_many_parts` - (Optional) The `too many parts` check according to `parts_to_delay_insert` and `parts_to_throw_insert` will be active only if the average part size (in the relevant partition) is not larger than the specified threshold. If it is larger than the specified threshold, the INSERTs will be neither delayed or rejected. This allows to have hundreds of terabytes in a single table on a single server if the parts are successfully merged to larger parts. This does not affect the thresholds on inactive parts or total parts.
* `min_age_to_force_merge_seconds` - (Optional) Merge parts if every part in the range is older than the value of `min_age_to_force_merge_seconds`.
* `min_age_to_force_merge_on_partition_only` - (Optional) Whether min_age_to_force_merge_seconds should be applied only on the entire partition and not on subset.
* `merge_selecting_sleep_ms` - (Optional) Sleep time for merge selecting when no part is selected. A lower setting triggers selecting tasks in background_schedule_pool frequently, which results in a large number of requests to ClickHouse Keeper in large-scale clusters.
* `merge_max_block_size` - (Optional) The number of rows that are read from the merged parts into memory. Default value: 8192.
* `check_sample_column_is_correct` - (Optional) Enables the check at table creation, that the data type of a column for sampling or sampling expression is correct. The data type must be one of unsigned integer types: UInt8, UInt16, UInt32, UInt64. Default value: true.
* `max_merge_selecting_sleep_ms` - (Optional) Maximum sleep time for merge selecting, a lower setting will trigger selecting tasks in background_schedule_pool frequently which result in large amount of requests to zookeeper in large-scale clusters. Default value: 60000 milliseconds (60 seconds).
* `max_cleanup_delay_period` - (Optional) Maximum period to clean old queue logs, blocks hashes and parts. Default value: 300 seconds.

The `kafka` block supports:

* `security_protocol` - (Optional) Security protocol used to connect to kafka server.
* `sasl_mechanism` - (Optional) SASL mechanism used in kafka authentication.
* `sasl_username` - (Optional) Username on kafka server.
* `sasl_password` - (Optional) User password on kafka server.
* `enable_ssl_certificate_verification` - (Optional) enable verification of SSL certificates.
* `max_poll_interval_ms` - (Optional) Maximum allowed time between calls to consume messages (e.g., rd_kafka_consumer_poll()) for high-level consumers. If this interval is exceeded the consumer is considered failed and the group will rebalance in order to reassign the partitions to another consumer group member.
* `session_timeout_ms` - (Optional) Client group session and failure detection timeout. The consumer sends periodic heartbeats (heartbeat.interval.ms) to indicate its liveness to the broker. If no hearts are received by the broker for a group member within the session timeout, the broker will remove the consumer from the group and trigger a rebalance.
* `debug` - (Optional) A comma-separated list of debug contexts to enable.
* `auto_offset_reset` - (Optional) Action to take when there is no initial offset in offset store or the desired offset is out of range: 'smallest','earliest' - automatically reset the offset to the smallest offset, 'largest','latest' - automatically reset the offset to the largest offset, 'error' - trigger an error (ERR__AUTO_OFFSET_RESET) which is retrieved by consuming messages and checking 'message->err'.

The `kafka_topic` block supports:

* `name` - (Required) Kafka topic name.
* `settings` - (Optional) Kafka connection settngs sanem as `kafka` block.

The `compression` block supports:

* `method` - (Optional) Method: Compression method. Two methods are available: LZ4 and zstd.
* `min_part_size` - (Optional) Min part size: Minimum size (in bytes) of a data part in a table. ClickHouse only applies the rule to tables with data parts greater than or equal to the Min part size value.
* `min_part_size_ratio` - (Optional) Min part size ratio: Minimum table part size to total table size ratio. ClickHouse only applies the rule to tables in which this ratio is greater than or equal to the Min part size ratio value.
* `level` - (Optional) Compression level for `ZSTD` method.

The `rabbitmq` block supports:

* `username` - (Optional) RabbitMQ username.
* `password` - (Optional) RabbitMQ user password.
* `vhost` - (Optional) RabbitMQ vhost. Default: '\'.

The `graphite_rollup` block supports:

* `name` - (Required) Graphite rollup configuration name.
* `pattern` - (Required) Set of thinning rules.
  * `function` - (Required) Aggregation function name.
  * `regexp` - (Optional) Regular expression that the metric name must match.
  * `retention` - Retain parameters.
    * `age` - (Required) Minimum data age in seconds.
    * `precision` - (Required) Accuracy of determining the age of the data in seconds.
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

The `jdbc_bridge` block supports:

* `host` - (Required) Host of jdbc bridge.
* `port` - (Optional) Port of jdbc bridge. Default value: 9019.

The `cloud_storage` block supports:

* `enabled` - (Required) Whether to use Yandex Object Storage for storing ClickHouse data. Can be either `true` or `false`.
* `move_factor` - Sets the minimum free space ratio in the cluster storage. If the free space is lower than this value, the data is transferred to Yandex Object Storage. Acceptable values are 0 to 1, inclusive.
* `data_cache_enabled` - Enables temporary storage in the cluster repository of data requested from the object repository.
* `data_cache_max_size` - Defines the maximum amount of memory (in bytes) allocated in the cluster storage for temporary storage of data requested from the object storage.
* `prefer_not_to_merge` - (Optional) Disables merging of data parts in `Yandex Object Storage`.

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - (Optional) Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - (Optional) Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Timestamp of cluster creation.

* `health` - Aggregated health of the cluster. Can be `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-clickhouse/api-ref/Cluster/).

* `status` - Status of the cluster. Can be `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-clickhouse/api-ref/Cluster/).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_clickhouse_cluster/import.sh" }}
