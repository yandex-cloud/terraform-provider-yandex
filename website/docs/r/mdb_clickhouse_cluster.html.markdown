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

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

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

* `web_sql` - (Optional) Allow access for DataLens. Can be either `true` or `false`.

* `data_lens` - (Optional) Allow access for Web SQL. Can be either `true` or `false`.

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

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Timestamp of cluster creation.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
  For more information see `health` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
  For more information see `status` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/api-ref/Cluster/).

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_clickhouse_cluster.foo cluster_id
```
