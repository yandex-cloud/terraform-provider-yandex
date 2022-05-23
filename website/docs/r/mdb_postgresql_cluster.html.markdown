---
layout: "yandex"
page_title: "Yandex: yandex_mdb_postgresql_cluster"
sidebar_current: "docs-yandex-mdb-postgresql-cluster"
description: |-
  Manages a PostgreSQL cluster within Yandex.Cloud.
---

# yandex\_mdb\_postgresql\_cluster

Manages a PostgreSQL cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-postgresql/).
[How to connect to the DB](https://cloud.yandex.com/en-ru/docs/managed-postgresql/quickstart#connect). To connect, use port 6432. The port number is not configurable.

## Example Usage

Example of creating a Single Node PostgreSQL.

```hcl
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 12
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
    postgresql_config = {
      max_connections                   = 395
      enable_parallel_hash              = true
      vacuum_cleanup_index_scale_factor = 0.2
      autovacuum_vacuum_scale_factor    = 0.34
      default_transaction_isolation     = "TRANSACTION_ISOLATION_READ_COMMITTED"
      shared_preload_libraries          = "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
    }
  }

  maintenance_window {
    type = "WEEKLY"
    day  = "SAT"
    hour = 12
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.foo.id
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

Example of creating a High-Availability (HA) PostgreSQL Cluster.

```hcl
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "ha"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 12
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  maintenance_window {
    type = "ANYTIME"
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.foo.id
  }

  host {
    zone      = "ru-central1-b"
    subnet_id = yandex_vpc_subnet.bar.id
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}
```


Example of creating a High-Availability (HA) PostgreSQL Cluster with priority and set master.

```hcl

resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "test_ha"
  description = "test High-Availability (HA) PostgreSQL Cluster with priority and set master"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  host_master_name = "host_name_c_2"

  config {
    version = "12"
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
    }

  }

  host {
    zone      = "ru-central1-a"
    name      = "host_name_a"
    priority  = 2
    subnet_id = yandex_vpc_subnet.a.id
  }
  host {
    zone                    = "ru-central1-b"
    name                    = "host_name_b"
    replication_source_name = "host_name_c"
    subnet_id               = yandex_vpc_subnet.b.id
  }
  host {
    zone      = "ru-central1-c"
    name      = "host_name_c"
    subnet_id = yandex_vpc_subnet.c.id
  }
  host {
    zone      = "ru-central1-c"
    name      = "host_name_c_2"
    subnet_id = yandex_vpc_subnet.c.id
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "c" {
  zone           = "ru-central1-c"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
```


Example of creating a Single Node PostgreSQL from backup.

```hcl
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  restore {
    backup_id = "c9q99999999999999994cm:base_000000010000005F000000B4"
    time      = "2021-02-11T15:04:05"
  }

  config {
    version = 12
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
    postgresql_config = {
      max_connections                   = 395
      enable_parallel_hash              = true
      vacuum_cleanup_index_scale_factor = 0.2
      autovacuum_vacuum_scale_factor    = 0.34
      default_transaction_isolation     = "TRANSACTION_ISOLATION_READ_COMMITTED"
      shared_preload_libraries          = "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.foo.id
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the PostgreSQL cluster. Provided by the client when the cluster is created.

* `config` - (Required) Configuration of the PostgreSQL cluster. The structure is documented below.

* `environment` - (Required) Deployment environment of the PostgreSQL cluster.

* `host` - (Required) A host of the PostgreSQL cluster. The structure is documented below.

* `network_id` - (Required) ID of the network, to which the PostgreSQL cluster belongs.

- - -

* `description` - (Optional) Description of the PostgreSQL cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is unset, the default provider `folder_id` is used for create.

* `labels` - (Optional) A set of key/value label pairs to assign to the PostgreSQL cluster.

* `host_master_name` - (Optional) It sets name of master host. It works only when `host.name` is set.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster.  Can be either `true` or `false`.

- - -

* `restore` - (Optional, ForceNew) The cluster will be created from the specified backup. The structure is documented below.

- - -

* `maintenance_window` - (Optional) Maintenance policy of the PostgreSQL cluster. The structure is documented below.


The `config` block supports:

* `resources` - (Required) Resources allocated to hosts of the PostgreSQL cluster. The structure is documented below.

* `version` - (Required) Version of the PostgreSQL cluster. (allowed versions are: 10, 10-1c, 11, 11-1c, 12, 12-1c, 13, 14)

* `access` - (Optional) Access policy to the PostgreSQL cluster. The structure is documented below.

* `performance_diagnostics` - (Optional) Cluster performance diagnostics settings. The structure is documented below. [YC Documentation](https://cloud.yandex.com/en-ru/docs/managed-postgresql/api-ref/grpc/cluster_service#PerformanceDiagnostics)

* `autofailover` - (Optional) Configuration setting which enables/disables autofailover in cluster.

* `backup_retain_period_days` - (Optional) The period in days during which backups are stored.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC timezone. The structure is documented below.

* `pooler_config` - (Optional) Configuration of the connection pooler. The structure is documented below.

* `postgresql_config` - (Optional) PostgreSQL cluster config. Detail info in "postresql config" section (documented below).

The `resources` block supports:

* `disk_size` - (Required) Volume of the storage available to a PostgreSQL host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of PostgreSQL hosts.

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a PostgreSQL host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/concepts/instance-types).

The `pooler_config` block supports:

* `pool_discard` - (Optional) Setting `pool_discard` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_discard-yesno).

* `pooling_mode` - (Optional) Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool-string.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started (UTC).

* `minutes` - (Optional) The minute at which backup will be started (UTC).

The `access` block supports:

* `data_lens` - (Optional) Allow access for [Yandex DataLens](https://cloud.yandex.com/services/datalens).

* `web_sql` - Allow access for [SQL queries in the management console](https://cloud.yandex.com/docs/managed-postgresql/operations/web-sql-query)

* `serverless` - Allow access for [connection to managed databases from functions](https://cloud.yandex.com/docs/functions/operations/database-connection)

The `performance_diagnostics` block supports:

* `enabled` - Enable performance diagnostics

* `sessions_sampling_interval` - Interval (in seconds) for pg_stat_activity sampling Acceptable values are 1 to 86400, inclusive.

* `statements_sampling_interval` - Interval (in seconds) for pg_stat_statements sampling Acceptable values are 1 to 86400, inclusive.

* `user` - (Deprecated) To manage users, please switch to using a separate resource type `yandex_mdb_postgresql_user`.

* `database` - (Deprecated) To manage databases, please switch to using a separate resource type `yandex_mdb_postgresql_database`.

~> **Note:** Historically, `user` and `database` blocks of the `yandex_mdb_postgresql_cluster` resource were used to manage users and databases of the PostgreSQL cluster. However, this approach has many disadvantages. In particular, adding and removing a resource from the terraform recipe worked wrong because terraform misleads the user about the planned changes. Now, the only possible way to manage databases and users is using `yandex_mdb_postgresql_user` and  `yandex_mdb_postgresql_database` resources.

The `host` block supports:

* `zone` - (Required) The availability zone where the PostgreSQL host will be created.

* `assign_public_ip` - (Optional) Sets whether the host should get a public IP address on creation. It can be changed on the fly only when `name` is set.

* `subnet_id` - (Optional) The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `fqdn` - (Computed) The fully qualified domain name of the host.

* `name` - (Optional) Host state name. It should be set for all hosts or unset for all hosts. This field can be used by another host, to select which host will be its replication source. Please see `replication_source_name` parameter.
Also, this field is used to select which host will be selected as a master host. Please see `host_master_name` parameter.

* `replication_source` - (Computed) Host replication source (fqdn), when replication_source is empty then host is in HA group.

* `priority` - Host priority in HA group. It works only when `name` is set.

* `replication_source_name` - (Optional) Host replication source name points to host's `name` from which this host should replicate. When not set then host in HA group. It works only when `name` is set.

The `restore` block supports:

* `backup_id` - (Required, ForceNew) Backup ID. The cluster will be created from the specified backup. [How to get a list of PostgreSQL backups](https://cloud.yandex.com/docs/managed-postgresql/operations/cluster-backups). 

* `time` - (Optional, ForceNew) Timestamp of the moment to which the PostgreSQL cluster should be restored. (Format: "2006-01-02T15:04:05" - UTC). When not set, current time is used.

* `time_inclusive` - (Optional, ForceNew) Flag that indicates whether a database should be restored to the first backup point available just after the timestamp specified in the [time] field instead of just before.  
Possible values:
  - false (default) — the restore point refers to the first backup moment before [time].
  - true — the restore point refers to the first backup point after [time].

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.

* `day` - (Optional) Day of the week (in `DDD` format). Allowed values: "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"

* `hour` - (Optional) Hour of the day in UTC (in `HH` format). Allowed value is between 1 and 24.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Timestamp of cluster creation.

* `health` - Aggregated health of the cluster.

* `status` - Status of the cluster.

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_postgresql_cluster.foo cluster_id
```


## PostgreSQL cluster settings

More information about config:
* https://cloud.yandex.com/docs/managed-postgresql/concepts/settings-list
* https://www.postgresql.org/docs/current/runtime-config-connection.html
* https://www.postgresql.org/docs/current/runtime-config-resource.html
* https://www.postgresql.org/docs/current/runtime-config-wal.html
* https://www.postgresql.org/docs/current/runtime-config-query.html
* https://www.postgresql.org/docs/current/runtime-config-logging.html
* https://www.postgresql.org/docs/current/runtime-config-autovacuum.html
* https://www.postgresql.org/docs/current/runtime-config-client.html
* https://www.postgresql.org/docs/current/runtime-config-locks.html
* https://www.postgresql.org/docs/current/runtime-config-compatible.html

| Setting name and type \ PostgreSQL version | 10 | 11 | 12 | 13 | 14 |
| ------------------------------------------ | -- | -- | -- | -- | -- |
| archive_timeout : integer | supported | supported | supported | supported | supported |
| array_nulls : boolean | supported | supported | supported | supported | supported |
| auto_explain_log_analyze : boolean | supported | supported | supported | supported | supported |
| auto_explain_log_buffers : boolean | supported | supported | supported | supported | supported |
| auto_explain_log_min_duration : integer | supported | supported | supported | supported | supported |
| auto_explain_log_nested_statements : boolean | supported | supported | supported | supported | supported |
| auto_explain_log_timing : boolean | supported | supported | supported | supported | supported |
| auto_explain_log_triggers : boolean | supported | supported | supported | supported | supported |
| auto_explain_log_verbose : boolean | supported | supported | supported | supported | supported |
| auto_explain_sample_rate : float | supported | supported | supported | supported | supported |
| autovacuum_analyze_scale_factor : float | supported | supported | supported | supported | supported |
| autovacuum_max_workers : integer | supported | supported | supported | supported | supported |
| autovacuum_naptime : integer | supported | supported | supported | supported | supported |
| autovacuum_vacuum_cost_delay : integer | supported | supported | supported | supported | supported |
| autovacuum_vacuum_cost_limit : integer | supported | supported | supported | supported | supported |
| autovacuum_vacuum_insert_scale_factor : float | - | - | - | supported | supported |
| autovacuum_vacuum_insert_threshold : integer | - | - | - | supported | supported |
| autovacuum_vacuum_scale_factor : float | supported | supported | supported | supported | supported |
| autovacuum_work_mem : integer | supported | supported | supported | supported | supported |
| backend_flush_after : integer | supported | supported | supported | supported | supported |
| backslash_quote : one of<br>  - 0: "BACKSLASH_QUOTE_UNSPECIFIED"<br>  - 1: "BACKSLASH_QUOTE"<br>  - 2: "BACKSLASH_QUOTE_ON"<br>  - 3: "BACKSLASH_QUOTE_OFF"<br>  - 4: "BACKSLASH_QUOTE_SAFE_ENCODING" | supported | supported | supported | supported | supported |
| bgwriter_delay : integer | supported | supported | supported | supported | supported |
| bgwriter_flush_after : integer | supported | supported | supported | supported | supported |
| bgwriter_lru_maxpages : integer | supported | supported | supported | supported | supported |
| bgwriter_lru_multiplier : float | supported | supported | supported | supported | supported |
| bytea_output : one of<br>  - 0: "BYTEA_OUTPUT_UNSPECIFIED"<br>  - 1: "BYTEA_OUTPUT_HEX"<br>  - 2: "BYTEA_OUTPUT_ESCAPED" | supported | supported | supported | supported | supported |
| checkpoint_completion_target : float | supported | supported | supported | supported | supported |
| checkpoint_flush_after : integer | supported | supported | supported | supported | supported |
| checkpoint_timeout : integer | supported | supported | supported | supported | supported |
| client_min_messages : one of<br>  - 0: "LOG_LEVEL_UNSPECIFIED"<br>  - 1: "LOG_LEVEL_DEBUG5"<br>  - 2: "LOG_LEVEL_DEBUG4"<br>  - 3: "LOG_LEVEL_DEBUG3"<br>  - 4: "LOG_LEVEL_DEBUG2"<br>  - 5: "LOG_LEVEL_DEBUG1"<br>  - 6: "LOG_LEVEL_LOG"<br>  - 7: "LOG_LEVEL_NOTICE"<br>  - 8: "LOG_LEVEL_WARNING"<br>  - 9: "LOG_LEVEL_ERROR"<br>  - 10: "LOG_LEVEL_FATAL"<br>  - 11: "LOG_LEVEL_PANIC" | supported | supported | supported | supported | supported |
| constraint_exclusion : one of<br>  - 0: "CONSTRAINT_EXCLUSION_UNSPECIFIED"<br>  - 1: "CONSTRAINT_EXCLUSION_ON"<br>  - 2: "CONSTRAINT_EXCLUSION_OFF"<br>  - 3: "CONSTRAINT_EXCLUSION_PARTITION" | supported | supported | supported | supported | supported |
| cursor_tuple_fraction : float | supported | supported | supported | supported | supported |
| deadlock_timeout : integer | supported | supported | supported | supported | supported |
| default_statistics_target : integer | supported | supported | supported | supported | supported |
| default_transaction_isolation : one of<br>  - 0: "TRANSACTION_ISOLATION_UNSPECIFIED"<br>  - 1: "TRANSACTION_ISOLATION_READ_UNCOMMITTED"<br>  - 2: "TRANSACTION_ISOLATION_READ_COMMITTED"<br>  - 3: "TRANSACTION_ISOLATION_REPEATABLE_READ"<br>  - 4: "TRANSACTION_ISOLATION_SERIALIZABLE" | supported | supported | supported | supported | supported |
| default_transaction_read_only : boolean | supported | supported | supported | supported | supported |
| default_with_oids : boolean | supported | supported | supported | supported | supported |
| effective_cache_size : integer | supported | supported | supported | supported | supported |
| effective_io_concurrency : integer | supported | supported | supported | supported | supported |
| enable_bitmapscan : boolean | supported | supported | supported | supported | supported |
| enable_hashagg : boolean | supported | supported | supported | supported | supported |
| enable_hashjoin : boolean | supported | supported | supported | supported | supported |
| enable_incremental_sort : boolean | - | - | - | supported | supported |
| enable_indexonlyscan : boolean | supported | supported | supported | supported | supported |
| enable_indexscan : boolean | supported | supported | supported | supported | supported |
| enable_material : boolean | supported | supported | supported | supported | supported |
| enable_mergejoin : boolean | supported | supported | supported | supported | supported |
| enable_nestloop : boolean | supported | supported | supported | supported | supported |
| enable_parallel_append : boolean | - | supported | supported | supported | supported |
| enable_parallel_hash : boolean | - | supported | supported | supported | supported |
| enable_partition_pruning : boolean | - | supported | supported | supported | supported |
| enable_partitionwise_aggregate : boolean | - | supported | supported | supported | supported |
| enable_partitionwise_join : boolean | - | supported | supported | supported | supported |
| enable_seqscan : boolean | supported | supported | supported | supported | supported |
| enable_sort : boolean | supported | supported | supported | supported | supported |
| enable_tidscan : boolean | supported | supported | supported | supported | supported |
| escape_string_warning : boolean | supported | supported | supported | supported | supported |
| exit_on_error : boolean | supported | supported | supported | supported | supported |
| force_parallel_mode : one of<br>  - 0: "FORCE_PARALLEL_MODE_UNSPECIFIED"<br>  - 1: "FORCE_PARALLEL_MODE_ON"<br>  - 2: "FORCE_PARALLEL_MODE_OFF"<br>  - 3: "FORCE_PARALLEL_MODE_REGRESS" | supported | supported | supported | supported | supported |
| from_collapse_limit : integer | supported | supported | supported | supported | supported |
| gin_pending_list_limit : integer | supported | supported | supported | supported | supported |
| hash_mem_multiplier : float | - | - | - | supported | supported |
| idle_in_transaction_session_timeout : integer | supported | supported | supported | supported | supported |
| jit : boolean | - | supported | supported | supported | supported |
| join_collapse_limit : integer | supported | supported | supported | supported | supported |
| lo_compat_privileges : boolean | supported | supported | supported | supported | supported |
| lock_timeout : integer | supported | supported | supported | supported | supported |
| log_checkpoints : boolean | supported | supported | supported | supported | supported |
| log_connections : boolean | supported | supported | supported | supported | supported |
| log_disconnections : boolean | supported | supported | supported | supported | supported |
| log_duration : boolean | supported | supported | supported | supported | supported |
| log_error_verbosity : one of<br>  - 0: "LOG_ERROR_VERBOSITY_UNSPECIFIED"<br>  - 1: "LOG_ERROR_VERBOSITY_TERSE"<br>  - 2: "LOG_ERROR_VERBOSITY_DEFAULT"<br>  - 3: "LOG_ERROR_VERBOSITY_VERBOSE" | supported | supported | supported | supported | supported |
| log_lock_waits : boolean | supported | supported | supported | supported | supported |
| log_min_duration_sample : integer | - | - | - | supported | supported |
| log_min_duration_statement : integer | supported | supported | supported | supported | supported |
| log_min_error_statement : one of<br>  - 0: "LOG_LEVEL_UNSPECIFIED"<br>  - 1: "LOG_LEVEL_DEBUG5"<br>  - 2: "LOG_LEVEL_DEBUG4"<br>  - 3: "LOG_LEVEL_DEBUG3"<br>  - 4: "LOG_LEVEL_DEBUG2"<br>  - 5: "LOG_LEVEL_DEBUG1"<br>  - 6: "LOG_LEVEL_LOG"<br>  - 7: "LOG_LEVEL_NOTICE"<br>  - 8: "LOG_LEVEL_WARNING"<br>  - 9: "LOG_LEVEL_ERROR"<br>  - 10: "LOG_LEVEL_FATAL"<br>  - 11: "LOG_LEVEL_PANIC" | supported | supported | supported | supported | supported |
| log_min_messages : one of<br>  - 0: "LOG_LEVEL_UNSPECIFIED"<br>  - 1: "LOG_LEVEL_DEBUG5"<br>  - 2: "LOG_LEVEL_DEBUG4"<br>  - 3: "LOG_LEVEL_DEBUG3"<br>  - 4: "LOG_LEVEL_DEBUG2"<br>  - 5: "LOG_LEVEL_DEBUG1"<br>  - 6: "LOG_LEVEL_LOG"<br>  - 7: "LOG_LEVEL_NOTICE"<br>  - 8: "LOG_LEVEL_WARNING"<br>  - 9: "LOG_LEVEL_ERROR"<br>  - 10: "LOG_LEVEL_FATAL"<br>  - 11: "LOG_LEVEL_PANIC" | supported | supported | supported | supported | supported |
| log_parameter_max_length : integer | - | - | - | supported | supported |
| log_parameter_max_length_on_error : integer | - | - | - | supported | supported |
| log_statement : one of<br>  - 0: "LOG_STATEMENT_UNSPECIFIED"<br>  - 1: "LOG_STATEMENT_NONE"<br>  - 2: "LOG_STATEMENT_DDL"<br>  - 3: "LOG_STATEMENT_MOD"<br>  - 4: "LOG_STATEMENT_ALL" | supported | supported | supported | supported | supported |
| log_statement_sample_rate : float | - | - | - | supported | supported |
| log_temp_files : integer | supported | supported | supported | supported | supported |
| log_transaction_sample_rate : float | - | - | supported | supported | supported |
| logical_decoding_work_mem : integer | - | - | - | supported | supported |
| maintenance_io_concurrency : integer | - | - | - | supported | supported |
| maintenance_work_mem : integer | supported | supported | supported | supported | supported |
| max_connections : integer | supported | supported | supported | supported | supported |
| max_locks_per_transaction : integer | supported | supported | supported | supported | supported |
| max_parallel_maintenance_workers : integer | - | supported | supported | supported | supported |
| max_parallel_workers : integer | supported | supported | supported | supported | supported |
| max_parallel_workers_per_gather : integer | supported | supported | supported | supported | supported |
| max_pred_locks_per_transaction : integer | supported | supported | supported | supported | supported |
| max_prepared_transactions : integer | supported | supported | supported | supported | supported |
| max_slot_wal_keep_size : integer | - | - | - | supported | supported |
| max_standby_streaming_delay : integer | supported | supported | supported | supported | supported |
| max_wal_size : integer | supported | supported | supported | supported | supported |
| max_worker_processes : integer | supported | supported | supported | supported | supported |
| min_wal_size : integer | supported | supported | supported | supported | supported |
| old_snapshot_threshold : integer | supported | supported | supported | supported | supported |
| operator_precedence_warning : boolean | supported | supported | supported | supported | supported |
| parallel_leader_participation : boolean | - | supported | supported | supported | supported |
| pg_hint_plan_debug_print : one of<br>  - 0: "PG_HINT_PLAN_DEBUG_PRINT_UNSPECIFIED"<br>  - 1: "PG_HINT_PLAN_DEBUG_PRINT_OFF"<br>  - 2: "PG_HINT_PLAN_DEBUG_PRINT_ON"<br>  - 3: "PG_HINT_PLAN_DEBUG_PRINT_DETAILED"<br>  - 4: "PG_HINT_PLAN_DEBUG_PRINT_VERBOSE" | supported | supported | supported | supported | supported |
| pg_hint_plan_enable_hint : boolean | supported | supported | supported | supported | supported |
| pg_hint_plan_enable_hint_table : boolean | supported | supported | supported | supported | supported |
| pg_hint_plan_message_level : one of<br>  - 0: "LOG_LEVEL_UNSPECIFIED"<br>  - 1: "LOG_LEVEL_DEBUG5"<br>  - 2: "LOG_LEVEL_DEBUG4"<br>  - 3: "LOG_LEVEL_DEBUG3"<br>  - 4: "LOG_LEVEL_DEBUG2"<br>  - 5: "LOG_LEVEL_DEBUG1"<br>  - 6: "LOG_LEVEL_LOG"<br>  - 7: "LOG_LEVEL_NOTICE"<br>  - 8: "LOG_LEVEL_WARNING"<br>  - 9: "LOG_LEVEL_ERROR"<br>  - 10: "LOG_LEVEL_FATAL"<br>  - 11: "LOG_LEVEL_PANIC" | supported | supported | supported | supported | supported |
| plan_cache_mode : one of<br>  - 0: "PLAN_CACHE_MODE_UNSPECIFIED"<br>  - 1: "PLAN_CACHE_MODE_AUTO"<br>  - 2: "PLAN_CACHE_MODE_FORCE_CUSTOM_PLAN"<br>  - 3: "PLAN_CACHE_MODE_FORCE_GENERIC_PLAN" | - | - | supported | supported | supported |
| quote_all_identifiers : boolean | supported | supported | supported | supported | supported |
| random_page_cost : float | supported | supported | supported | supported | supported |
| replacement_sort_tuples : integer | supported | - | - | - | - |
| row_security : boolean | supported | supported | supported | supported | supported |
| search_path : text | supported | supported | supported | supported | supported |
| seq_page_cost : float | supported | supported | supported | supported | supported |
| shared_buffers : integer | supported | supported | supported | supported | supported |
| shared_preload_libraries : override if not set. one of<br> - "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN<br> - "SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"<br> - "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"<br> - "SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"<br> - NO value | supported | supported | supported | supported | supported |
| standard_conforming_strings : boolean | supported | supported | supported | supported | supported |
| statement_timeout : integer | supported | supported | supported | supported | supported |
| synchronize_seqscans : boolean | supported | supported | supported | supported | supported |
| synchronous_commit : one of<br>  - 0: "SYNCHRONOUS_COMMIT_UNSPECIFIED"<br>  - 1: "SYNCHRONOUS_COMMIT_ON"<br>  - 2: "SYNCHRONOUS_COMMIT_OFF"<br>  - 3: "SYNCHRONOUS_COMMIT_LOCAL"<br>  - 4: "SYNCHRONOUS_COMMIT_REMOTE_WRITE"<br>  - 5: "SYNCHRONOUS_COMMIT_REMOTE_APPLY" | supported | supported | supported | supported | supported |
| temp_buffers : integer | supported | supported | supported | supported | supported |
| temp_file_limit : integer | supported | supported | supported | supported | supported |
| timezone : text | supported | supported | supported | supported | supported |
| track_activity_query_size : integer | supported | supported | supported | supported | supported |
| transform_null_equals : boolean | supported | supported | supported | supported | supported |
| vacuum_cleanup_index_scale_factor : float | - | supported | supported | supported | - |
| vacuum_cost_delay : integer | supported | supported | supported | supported | supported |
| vacuum_cost_limit : integer | supported | supported | supported | supported | supported |
| vacuum_cost_page_dirty : integer | supported | supported | supported | supported | supported |
| vacuum_cost_page_hit : integer | supported | supported | supported | supported | supported |
| vacuum_cost_page_miss : integer | supported | supported | supported | supported | supported |
| wal_keep_size : integer | - | - | - | supported | supported |
| wal_level : one of<br>  - 0: "WAL_LEVEL_UNSPECIFIED"<br>  - 1: "WAL_LEVEL_REPLICA"<br>  - 2: "WAL_LEVEL_LOGICAL" | supported | supported | supported | supported | supported |
| work_mem : integer | supported | supported | supported | supported | supported |
| xmlbinary : one of<br>  - 0: "XML_BINARY_UNSPECIFIED"<br>  - 1: "XML_BINARY_BASE64"<br>  - 2: "XML_BINARY_HEX" | supported | supported | supported | supported | supported |
| xmloption : one of<br>  - 0: "XML_OPTION_UNSPECIFIED"<br>  - 1: "XML_OPTION_DOCUMENT"<br>  - 2: "XML_OPTION_CONTENT" | supported | supported | supported | supported | supported |