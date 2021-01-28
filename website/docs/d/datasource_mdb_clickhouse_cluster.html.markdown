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

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.

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