---
layout: "yandex"
page_title: "Yandex: yandex_mdb_clickhouse_cluster"
sidebar_current: "docs-yandex-mdb-clickhouse-cluster"
description: |-
  Manages a ClickHouse cluster within Yandex.Cloud.
---

# yandex\_mdb\_clickhouse\_cluster

Manages a ClickHouse cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).

## Example Usage

Example of creating a Single Node ClickHouse.

```hcl
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 32
    }

    config {
      log_level                       = "TRACE"
      max_connections                 = 100
      max_concurrent_queries          = 50
      keep_alive_timeout              = 3000
      uncompressed_cache_size         = 8589934592
      mark_cache_size                 = 5368709120
      max_table_size_to_drop          = 53687091200
      max_partition_size_to_drop      = 53687091200
      timezone                        = "UTC"
      geobase_uri                     = ""
      query_log_retention_size        = 1073741824
      query_log_retention_time        = 2592000
      query_thread_log_enabled        = true
      query_thread_log_retention_size = 536870912
      query_thread_log_retention_time = 2592000
      part_log_retention_size         = 536870912
      part_log_retention_time         = 2592000
      metric_log_enabled              = true
      metric_log_retention_size       = 536870912
      metric_log_retention_time       = 2592000
      trace_log_enabled               = true
      trace_log_retention_size        = 536870912
      trace_log_retention_time        = 2592000
      text_log_enabled                = true
      text_log_retention_size         = 536870912
      text_log_retention_time         = 2592000
      text_log_level                  = "TRACE"
      background_pool_size            = 16
      background_schedule_pool_size   = 16

      merge_tree {
        replicated_deduplication_window                           = 100
        replicated_deduplication_window_seconds                   = 604800
        parts_to_delay_insert                                     = 150
        parts_to_throw_insert                                     = 300
        max_replicated_merges_in_queue                            = 16
        number_of_free_entries_in_pool_to_lower_max_size_of_merge = 8
        max_bytes_to_merge_at_min_space_in_pool                   = 1048576
      }

      kafka {
        security_protocol = "SECURITY_PROTOCOL_PLAINTEXT"
        sasl_mechanism    = "SASL_MECHANISM_GSSAPI"
        sasl_username     = "user1"
        sasl_password     = "pass1"
      }

      kafka_topic {
        name = "topic1"
        settings {
          security_protocol = "SECURITY_PROTOCOL_SSL"
          sasl_mechanism    = "SASL_MECHANISM_SCRAM_SHA_256"
          sasl_username     = "user2"
          sasl_password     = "pass2"
        }
      }

      kafka_topic {
        name = "topic2"
        settings {
          security_protocol = "SECURITY_PROTOCOL_SASL_PLAINTEXT"
          sasl_mechanism    = "SASL_MECHANISM_PLAIN"
        }
      }

      rabbitmq {
        username = "rabbit_user"
        password = "rabbit_pass"
      }

      compression {
        method              = "LZ4"
        min_part_size       = 1024
        min_part_size_ratio = 0.5
      }

      compression {
        method              = "ZSTD"
        min_part_size       = 2048
        min_part_size_ratio = 0.7
      }

      graphite_rollup {
        name = "rollup1"
        pattern {
          regexp   = "abc"
          function = "func1"
          retention {
            age       = 1000
            precision = 3
          }
        }
      }

      graphite_rollup {
        name = "rollup2"
        pattern {
          function = "func2"
          retention {
            age       = 2000
            precision = 5
          }
        }
      }
    }
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user"
    password = "your_password"
    permission {
      database_name = "db_name"
    }
    settings {
      max_memory_usage_for_user               = 1000000000
      read_overflow_mode                      = "throw"
      output_format_json_quote_64bit_integers = true
    }
    quota {
      interval_duration = 3600000
      queries           = 10000
      errors            = 1000
    }
    quota {
      interval_duration = 79800000
      queries           = 50000
      errors            = 5000
    }
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }

  format_schema {
    name = "test_schema"
    type = "FORMAT_SCHEMA_TYPE_CAPNPROTO"
    uri  = "https://storage.yandexcloud.net/ch-data/schema.proto"
  }

  ml_model {
    name = "test_model"
    type = "ML_MODEL_TYPE_CATBOOST"
    uri  = "https://storage.yandexcloud.net/ch-data/train.csv"
  }
  
  service_account_id = "your_service_account_id"
  
  cloud_storage {
    enabled = false
  }

  maintenance_window {
    type = "ANYTIME"
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

Example of creating a HA ClickHouse Cluster.

```hcl
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name        = "ha"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  zookeeper {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user"
    password = "password"
    permission {
      database_name = "db_name"
    }
    settings {
      max_memory_usage_for_user               = 1000000000
      read_overflow_mode                      = "throw"
      output_format_json_quote_64bit_integers = true
    }
    quota {
      interval_duration = 3600000
      queries           = 10000
      errors            = 1000
    }
    quota {
      interval_duration = 79800000
      queries           = 50000
      errors            = 5000
    }
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-b"
    subnet_id = "${yandex_vpc_subnet.bar.id}"
  }

  host {
    type      = "ZOOKEEPER"
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }

  host {
    type      = "ZOOKEEPER"
    zone      = "ru-central1-b"
    subnet_id = "${yandex_vpc_subnet.bar.id}"
  }

  host {
    type      = "ZOOKEEPER"
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.baz.id}"
  }
   
  cloud_storage {
    enabled = false
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}
```

Example of creating a sharded ClickHouse Cluster.

```hcl
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name        = "sharded"
  environment = "PRODUCTION"
  network_id  = "${yandex_vpc_network.foo.id}"

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  zookeeper {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user"
    password = "password"
    permission {
      database_name = "db_name"
    }
    settings {
      max_memory_usage_for_user               = 1000000000
      read_overflow_mode                      = "throw"
      output_format_json_quote_64bit_integers = true
    }
    quota {
      interval_duration = 3600000
      queries           = 10000
      errors            = 1000
    }
    quota {
      interval_duration = 79800000
      queries           = 50000
      errors            = 5000
    }
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-a"
    subnet_id  = "${yandex_vpc_subnet.foo.id}"
    shard_name = "shard1"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-b"
    subnet_id  = "${yandex_vpc_subnet.bar.id}"
    shard_name = "shard1"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-b"
    subnet_id  = "${yandex_vpc_subnet.bar.id}"
    shard_name = "shard2"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.baz.id}"
    shard_name = "shard2"
  }

  shard_group {
    name        = "single_shard_group"
    description = "Cluster configuration that contain only shard1"
    shard_names = [
      "shard1",
    ]
  }
  
  cloud_storage {
    enabled = false
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the ClickHouse cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the ClickHouse cluster belongs.

* `environment` - (Required) Deployment environment of the ClickHouse cluster. Can be either `PRESTABLE` or `PRODUCTION`.

* `clickhouse` - (Required) Configuration of the ClickHouse subcluster. The structure is documented below.

* `user` - (Required) A user of the ClickHouse cluster. The structure is documented below.

* `database` - (Required) A database of the ClickHouse cluster. The structure is documented below.

* `host` - (Required) A host of the ClickHouse cluster. The structure is documented below.

- - -

* `version` - (Optional) Version of the ClickHouse server software.

* `description` - (Optional) Description of the ClickHouse cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

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

* `deletion_protection` - (Optional) Inhibits deletion of the cluster.  Can be either `true` or `false`.


- - -

The `clickhouse` block supports:

* `resources` - (Required) Resources allocated to hosts of the ClickHouse subcluster. The structure is documented below.

* `config` - (Optional) Main ClickHouse cluster configuration.

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a ClickHouse host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).

* `disk_size` - (Required) Volume of the storage available to a ClickHouse host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of ClickHouse hosts.
  For more information see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts/storage).

The `zookeeper` block supports:

* `resources` - (Optional) Resources allocated to hosts of the ZooKeeper subcluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - (Optional) The ID of the preset for computational resources available to a ZooKeeper host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).

* `disk_size` - (Optional) Volume of the storage available to a ZooKeeper host, in gigabytes.

* `disk_type_id` - (Optional) Type of the storage of ZooKeeper hosts.
  For more information see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts/storage).

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

* `select_sequential_consistency` - (Optional) Enables or disables sequential consistency for SELECT queries.

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

* `zone` - (Required) The availability zone where the ClickHouse host will be created.
  For more information see [the official documentation](https://cloud.yandex.com/docs/overview/concepts/geo-scope).
  
* `subnet_id` (Optional) - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `shard_name` (Optional) - The name of the shard to which the host belongs.

* `assign_public_ip` (Optional) - Sets whether the host should get a public IP address on creation. Can be either `true` or `false`.

The `shard_group` block supports:

* `name` (Required) - The name of the shard group, used as cluster name in Distributed tables.

* `description` (Optional) - Description of the shard group.

* `shard_names` (Required) -  List of shards names that belong to the shard group.

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

The `config` block supports:

* `log_level`, `max_connections`, `max_concurrent_queries`, `keep_alive_timeout`, `uncompressed_cache_size`, `mark_cache_size`,
`max_table_size_to_drop`, `max_partition_size_to_drop`, `timezone`, `geobase_uri`, `query_log_retention_size`,
`query_log_retention_time`, `query_thread_log_enabled`, `query_thread_log_retention_size`, `query_thread_log_retention_time`,
`part_log_retention_size`, `part_log_retention_time`, `metric_log_enabled`, `metric_log_retention_size`, `metric_log_retention_time`,
`trace_log_enabled`, `trace_log_retention_size`, `trace_log_retention_time`, `text_log_enabled`, `text_log_retention_size`,
`text_log_retention_time`, `text_log_level`, `background_pool_size`, `background_schedule_pool_size` - (Optional) ClickHouse server parameters. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/operations/update#change-clickhouse-config)
and [the ClickHouse documentation](https://clickhouse.tech/docs/en/operations/server-configuration-parameters/settings/).

* `merge_tree` - (Optional) MergeTree engine configuration. The structure is documented below.
* `kafka` - (Optional) Kafka connection configuration. The structure is documented below.
* `kafka_topic` - (Optional) Kafka topic connection configuration. The structure is documented below.
* `compression` - (Optional) Data compression configuration. The structure is documented below.
* `rabbitmq` - (Optional) RabbitMQ connection configuration. The structure is documented below.
* `graphite_rollup` - (Optional) Graphite rollup configuration. The structure is documented below.

The `merge_tree` block supports:

* `replicated_deduplication_window` - (Optional) Replicated deduplication window: Number of recent hash blocks that ZooKeeper will store (the old ones will be deleted).
* `replicated_deduplication_window_seconds` - (Optional) Replicated deduplication window seconds: Time during which ZooKeeper stores the hash blocks (the old ones wil be deleted).
* `parts_to_delay_insert` - (Optional) Parts to delay insert: Number of active data parts in a table, on exceeding which ClickHouse starts artificially reduce the rate of inserting data into the table.
* `parts_to_throw_insert` - (Optional) Parts to throw insert: Threshold value of active data parts in a table, on exceeding which ClickHouse throws the 'Too many parts ...' exception.
* `max_replicated_merges_in_queue` - (Optional) Max replicated merges in queue: Maximum number of merge tasks that can be in the ReplicatedMergeTree queue at the same time.
* `number_of_free_entries_in_pool_to_lower_max_size_of_merge` - (Optional) Number of free entries in pool to lower max size of merge: Threshold value of free entries in the pool. If the number of entries in the pool falls below this value, ClickHouse reduces the maximum size of a data part to merge. This helps handle small merges faster, rather than filling the pool with lengthy merges.
* `max_bytes_to_merge_at_min_space_in_pool` - (Optional) Max bytes to merge at min space in pool: Maximum total size of a data part to merge when the number of free threads in the background pool is minimum.

The `kafka` block supports:

* `security_protocol` - (Optional) Security protocol used to connect to kafka server.
* `sasl_mechanism` - (Optional) SASL mechanism used in kafka authentication.
* `sasl_username` - (Optional) Username on kafka server.
* `sasl_password` - (Optional) User password on kafka server.

The `kafka_topic` block supports:

* `name` - (Required) Kafka topic name.
* `settings` - (Optional) Kafka connection settngs sanem as `kafka` block.

The `compression` block supports:

* `method` - (Optional) Method: Compression method. Two methods are available: LZ4 and zstd.
* `min_part_size` - (Optional) Min part size: Minimum size (in bytes) of a data part in a table. ClickHouse only applies the rule to tables with data parts greater than or equal to the Min part size value.
* `min_part_size_ratio` - (Optional) Min part size ratio: Minimum table part size to total table size ratio. ClickHouse only applies the rule to tables in which this ratio is greater than or equal to the Min part size ratio value.

The `rabbitmq` block supports:

* `username` - (Optional) RabbitMQ username.
* `password` - (Optional) RabbitMQ user password.

The `graphite_rollup` block supports:

* `name` - (Required) Graphite rollup configuration name.
* `pattern` - (Required) Set of thinning rules.
  * `function` - (Required) Aggregation function name.
  * `regexp` - (Optional) Regular expression that the metric name must match.
  * `retention` - Retain parameters.
    * `age` - (Required) Minimum data age in seconds.
    * `precision` - (Required) Accuracy of determining the age of the data in seconds.

The `cloud_storage` block supports:

* `enabled` - (Required) Whether to use Yandex Object Storage for storing ClickHouse data. Can be either `true` or `false`.

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - (Optional) Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - (Optional) Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Timestamp of cluster creation.

* `health` - Aggregated health of the cluster. Can be `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
  For more information see `health` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/api-ref/Cluster/).

* `status` - Status of the cluster. Can be `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
  For more information see `status` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/api-ref/Cluster/).

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_clickhouse_cluster.foo cluster_id
```
