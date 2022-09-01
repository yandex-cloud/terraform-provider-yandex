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

* `resources_preset_id` - The ID of the preset for computational resources available to a ClickHouse or ZooKeeper host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).
* `disk_size` - Volume of the storage available to a ClickHouse or ZooKeeper host, in gigabytes.
* `disk_type_id` - Type of the storage of ClickHouse or ZooKeeper hosts.

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
`text_log_retention_time`, `text_log_level`, `background_pool_size`, `background_schedule_pool_size` - ClickHouse server parameters. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/operations/update#change-clickhouse-config)
and [the ClickHouse documentation](https://clickhouse.tech/docs/en/operations/server-configuration-parameters/settings/).

* `merge_tree` - MergeTree engine configuration. The structure is documented below.
* `kafka` - Kafka connection configuration. The structure is documented below.
* `kafka_topic` - Kafka topic connection configuration. The structure is documented below.
* `compression` - Data compression configuration. The structure is documented below.
* `rabbitmq` - RabbitMQ connection configuration. The structure is documented below.
* `graphite_rollup` - Graphite rollup configuration. The structure is documented below.

The `merge_tree` block supports:

* `replicated_deduplication_window` - Replicated deduplication window: Number of recent hash blocks that ZooKeeper will store (the old ones will be deleted).
* `replicated_deduplication_window_seconds` - Replicated deduplication window seconds: Time during which ZooKeeper stores the hash blocks (the old ones wil be deleted).
* `parts_to_delay_insert` - Parts to delay insert: Number of active data parts in a table, on exceeding which ClickHouse starts artificially reduce the rate of inserting data into the table.
* `parts_to_throw_insert` - Parts to throw insert: Threshold value of active data parts in a table, on exceeding which ClickHouse throws the 'Too many parts ...' exception.
* `max_replicated_merges_in_queue` - Max replicated merges in queue: Maximum number of merge tasks that can be in the ReplicatedMergeTree queue at the same time.
* `number_of_free_entries_in_pool_to_lower_max_size_of_merge` - Number of free entries in pool to lower max size of merge: Threshold value of free entries in the pool. If the number of entries in the pool falls below this value, ClickHouse reduces the maximum size of a data part to merge. This helps handle small merges faster, rather than filling the pool with lengthy merges.
* `max_bytes_to_merge_at_min_space_in_pool` - Max bytes to merge at min space in pool: Maximum total size of a data part to merge when the number of free threads in the background pool is minimum.

The `kafka` block supports:

* `security_protocol` - Security protocol used to connect to kafka server.
* `sasl_mechanism` - SASL mechanism used in kafka authentication.
* `sasl_username` - Username on kafka server.
* `sasl_password` - User password on kafka server.

The `kafka_topic` block supports:

* `name` - Kafka topic name.
* `settings` - Kafka connection settngs sanem as `kafka` block.

The `compression` block supports:

* `method` - Method: Compression method. Two methods are available: LZ4 and zstd.
* `min_part_size` - Min part size: Minimum size (in bytes) of a data part in a table. ClickHouse only applies the rule to tables with data parts greater than or equal to the Min part size value.
* `min_part_size_ratio` - Min part size ratio: Minimum table part size to total table size ratio. ClickHouse only applies the rule to tables in which this ratio is greater than or equal to the Min part size ratio value.

The `rabbitmq` block supports:

* `username` - RabbitMQ username.
* `password` - RabbitMQ user password.

The `graphite_rollup` block supports:

* `name` - Graphite rollup configuration name.
* `pattern` - Set of thinning rules.
  * `function` - Aggregation function name.
  * `regexp` - Regular expression that the metric name must match.
  * `retention` - Retain parameters.
    * `age` - Minimum data age in seconds.
    * `precision` - Accuracy of determining the age of the data in seconds.

The `cloud_storage` block supports:

* `enabled` - (Required) Whether to use Yandex Object Storage for storing ClickHouse data. Can be either `true` or `false`.

The `maintenance_window` block supports:

* `type` - Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.
