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

  database {
    name  = "db_name"
    owner = "user_name"
  }

  user {
    name       = "user_name"
    password   = "your_password"
    conn_limit = 50
    permission {
      database_name = "db_name"
    }
    settings = {
      default_transaction_isolation = "read committed"
      log_min_duration_statement    = 5000
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

  database {
    name  = "db_name"
    owner = "user_name"
  }

  user {
    name     = "user_name"
    password = "password"
    permission {
      database_name = "db_name"
    }
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
  user {
    name     = "alice"
    password = "mysecurepassword"
    permission {
      database_name = "testdb"
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

  database {
    owner = "alice"
    name  = "testdb"
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

  database {
    name  = "db_name"
    owner = "user_name"
  }

  user {
    name       = "user_name"
    password   = "your_password"
    conn_limit = 50
    permission {
      database_name = "db_name"
    }
    settings = {
      default_transaction_isolation = "read committed"
      log_min_duration_statement    = 5000
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

Example of creating a High-Availability (HA) PostgreSQL cluster with multiple databases and users.
```hcl
resource "random_password" "passwords" {
  count   = 2
  length  = 16
  special = true
}

output "db_instance_alice_password" {
  value = random_password.passwords[0].result
}

output "db_instance_bob_password" {
  value = random_password.passwords[1].result
}

resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "ha_mdu_backup"
  description = "Example of multiple databases and users"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  folder_id   = "b1g24daaaddddffma52u"

  config {
    version = "13"
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 10
      disk_type_id       = "network-ssd"
    }

    access {
      web_sql = true
    }

    postgresql_config = {
      max_connections                   = 395
      enable_parallel_hash              = true
      vacuum_cleanup_index_scale_factor = 0.2
      autovacuum_vacuum_scale_factor    = 0.32
      default_transaction_isolation     = "TRANSACTION_ISOLATION_READ_UNCOMMITTED"
      shared_preload_libraries          = "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
    }

    pooler_config {
      pool_discard = true
      pooling_mode = "SESSION"
    }
  }

  user {
    name       = "alice"
    password   = random_password.passwords[0].result
    conn_limit = 10
    permission {
      database_name = "testdb"
    }
    permission {
      database_name = "testdb1"
    }
    permission {
      database_name = "testdb2"
    }
  }

  user {
    name     = "bob"
    password = random_password.passwords[1].result
    permission {
      database_name = "testdb2"
    }
    permission {
      database_name = "testdb1"
    }
  }
  user {
    name     = "chuck"
    password = "123456789"
    permission {
      database_name = "testdb"
    }
    grants = ["bob", "alice"]
  }

  host {
    zone      = "ru-central1-b"
    subnet_id = yandex_vpc_subnet.b.id
  }
  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.a.id
  }
  host {
    zone      = "ru-central1-c"
    subnet_id = yandex_vpc_subnet.c.id
  }

  database {
    owner = "alice"
    name  = "testdb"
  }
  database {
    owner = "alice"
    name  = "testdb2"
  }
  database {
    owner = "bob"
    name  = "testdb1"
    extension {
      name = "postgis"
    }
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "a" {
  name           = "mysubnet-a"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}
resource "yandex_vpc_subnet" "b" {
  name           = "mysubnet-b"
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}
resource "yandex_vpc_subnet" "c" {
  name           = "mysubnet-c"
  zone           = "ru-central1-c"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `config` - (Required) Configuration of the PostgreSQL cluster. The structure is documented below.

* `database` - (Required) A database of the PostgreSQL cluster. The structure is documented below.

* `environment` - (Required) Deployment environment of the PostgreSQL cluster.

* `host` - (Required) A host of the PostgreSQL cluster. The structure is documented below.

* `name` - (Required) Name of the PostgreSQL cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the PostgreSQL cluster belongs.

* `user` - (Required) A user of the PostgreSQL cluster. The structure is documented below.

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

* `version` - (Required) Version of the PostgreSQL cluster. (allowed versions are: 10, 10-1c, 11, 11-1c, 12, 12-1c, 13)

* `access` - (Optional) Access policy to the PostgreSQL cluster. The structure is documented below.

* `performance_diagnostics` - (Optional) Cluster performance diagnostics settings. The structure is documented below. [YC Documentation](https://cloud.yandex.com/docs/managed-postgresql/grpc/cluster_service#PerformanceDiagnostics)

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

The `user` block supports:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `grants` - (Optional) List of the user's grants.

* `login` - (Optional) User's ability to login.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

* `conn_limit` - (Optional) The maximum number of connections per user. (Default 50)

* `settings` - (Optional) Map of user settings. List of settings is documented below.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

The `settings` block supports:
Full description https://cloud.yandex.com/docs/managed-postgresql/grpc/user_service#UserSettings  

* `default_transaction_isolation` - defines the default isolation level to be set for all new SQL transactions. 
* * 0: "unspecified"
* * 1: "read uncommitted"
* * 2: "read committed"
* * 3: "repeatable read"
* * 4: "serializable"

* `lock_timeout` - The maximum time (in milliseconds) for any statement to wait for acquiring a lock on an table, index, row or other database object (default 0)

* `log_min_duration_statement` - This setting controls logging of the duration of statements. (default -1 disables logging of the duration of statements.)

* `synchronous_commit` - This setting defines whether DBMS will commit transaction in a synchronous way.
* * 0: "unspecified"
* * 1: "on"
* * 2: "off"
* * 3: "local"
* * 4: "remote write"
* * 5: "remote apply"

* `temp_file_limit` - The maximum storage space size (in kilobytes) that a single process can use to create temporary files.

* `log_statement` - This setting specifies which SQL statements should be logged (on the user level).
* * 0: "unspecified"
* * 1: "none"
* * 2: "ddl"
* * 3: "mod"
* * 4: "all"

The `database` block supports:

* `name` - (Required) The name of the database.

* `owner` - (Required) Name of the user assigned as the owner of the database. Forbidden to change in an existing database.

* `extension` - (Optional) Set of database extensions. The structure is documented below

* `lc_collate` - (Optional) POSIX locale for string sorting order. Forbidden to change in an existing database.

* `lc_type` - (Optional) POSIX locale for character classification. Forbidden to change in an existing database.

The `extension` block supports:

* `name` - (Required) Name of the database extension. For more information on available extensions see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/operations/cluster-extensions).

* `version` - (Optional) Version of the extension.

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


## postresql config

More information about config:  
* https://www.postgresql.org/docs/current/runtime-config-connection.html
* https://www.postgresql.org/docs/current/runtime-config-resource.html
* https://www.postgresql.org/docs/current/runtime-config-wal.html
* https://www.postgresql.org/docs/current/runtime-config-query.html
* https://www.postgresql.org/docs/current/runtime-config-logging.html
* https://www.postgresql.org/docs/current/runtime-config-autovacuum.html
* https://www.postgresql.org/docs/current/runtime-config-client.html
* https://www.postgresql.org/docs/current/runtime-config-locks.html
* https://www.postgresql.org/docs/current/runtime-config-compatible.html


* `shared_preload_libraries` override if not set. One of:
* * "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
* * "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"
* * "SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
* * NO value

Other are not owweride if not set.

### Postgresql 13 config
* `archive_timeout` integer

* `array_nulls` boolean

* `auto_explain_log_analyze` boolean

* `auto_explain_log_buffers` boolean

* `auto_explain_log_min_duration` integer

* `auto_explain_log_nested_statements` boolean

* `auto_explain_log_timing` boolean

* `auto_explain_log_triggers` boolean

* `auto_explain_log_verbose` boolean

* `auto_explain_sample_rate` float

* `autovacuum_analyze_scale_factor` float

* `autovacuum_max_workers` integer

* `autovacuum_naptime` integer

* `autovacuum_vacuum_cost_delay` integer

* `autovacuum_vacuum_cost_limit` integer

* `autovacuum_vacuum_insert_scale_factor` float

* `autovacuum_vacuum_insert_threshold` integer

* `autovacuum_vacuum_scale_factor` float

* `autovacuum_work_mem` integer

* `backend_flush_after` integer

* `backslash_quote` one of:
  - 0: "BACKSLASH_QUOTE_UNSPECIFIED"
  - 1: "BACKSLASH_QUOTE"
  - 2: "BACKSLASH_QUOTE_ON"
  - 3: "BACKSLASH_QUOTE_OFF"
  - 4: "BACKSLASH_QUOTE_SAFE_ENCODING"

* `bgwriter_delay` integer

* `bgwriter_flush_after` integer

* `bgwriter_lru_maxpages` integer

* `bgwriter_lru_multiplier` float

* `bytea_output` one of:
  - 0: "BYTEA_OUTPUT_UNSPECIFIED"
  - 1: "BYTEA_OUTPUT_HEX"
  - 2: "BYTEA_OUTPUT_ESCAPED"

* `checkpoint_completion_target` float

* `checkpoint_flush_after` integer

* `checkpoint_timeout` integer

* `client_min_messages` one of:
  - 0: "LOG_LEVEL_UNSPECIFIED"
  - 1: "LOG_LEVEL_DEBUG5"
  - 2: "LOG_LEVEL_DEBUG4"
  - 3: "LOG_LEVEL_DEBUG3"
  - 4: "LOG_LEVEL_DEBUG2"
  - 5: "LOG_LEVEL_DEBUG1"
  - 6: "LOG_LEVEL_LOG"
  - 7: "LOG_LEVEL_NOTICE"
  - 8: "LOG_LEVEL_WARNING"
  - 9: "LOG_LEVEL_ERROR"
  - 10: "LOG_LEVEL_FATAL"
  - 11: "LOG_LEVEL_PANIC"

* `constraint_exclusion` one of:
  - 0: "CONSTRAINT_EXCLUSION_UNSPECIFIED"
  - 1: "CONSTRAINT_EXCLUSION_ON"
  - 2: "CONSTRAINT_EXCLUSION_OFF"
  - 3: "CONSTRAINT_EXCLUSION_PARTITION"

* `cursor_tuple_fraction` float

* `deadlock_timeout` integer

* `default_statistics_target` integer

* `default_transaction_isolation` one of:
  - 0: "TRANSACTION_ISOLATION_UNSPECIFIED"
  - 1: "TRANSACTION_ISOLATION_READ_UNCOMMITTED"
  - 2: "TRANSACTION_ISOLATION_READ_COMMITTED"
  - 3: "TRANSACTION_ISOLATION_REPEATABLE_READ"
  - 4: "TRANSACTION_ISOLATION_SERIALIZABLE"

* `default_transaction_read_only` boolean

* `default_with_oids` boolean

* `effective_cache_size` integer

* `effective_io_concurrency` integer

* `enable_bitmapscan` boolean

* `enable_hashagg` boolean

* `enable_hashjoin` boolean

* `enable_incremental_sort` boolean

* `enable_indexonlyscan` boolean

* `enable_indexscan` boolean

* `enable_material` boolean

* `enable_mergejoin` boolean

* `enable_nestloop` boolean

* `enable_parallel_append` boolean

* `enable_parallel_hash` boolean

* `enable_partition_pruning` boolean

* `enable_partitionwise_aggregate` boolean

* `enable_partitionwise_join` boolean

* `enable_seqscan` boolean

* `enable_sort` boolean

* `enable_tidscan` boolean

* `escape_string_warning` boolean

* `exit_on_error` boolean

* `force_parallel_mode` one of:
  - 0: "FORCE_PARALLEL_MODE_UNSPECIFIED"
  - 1: "FORCE_PARALLEL_MODE_ON"
  - 2: "FORCE_PARALLEL_MODE_OFF"
  - 3: "FORCE_PARALLEL_MODE_REGRESS"

* `from_collapse_limit` integer

* `gin_pending_list_limit` integer

* `hash_mem_multiplier` float

* `idle_in_transaction_session_timeout` integer

* `jit` boolean

* `join_collapse_limit` integer

* `lo_compat_privileges` boolean

* `lock_timeout` integer

* `log_checkpoints` boolean

* `log_connections` boolean

* `log_disconnections` boolean

* `log_duration` boolean

* `log_error_verbosity` one of:
  - 0: "LOG_ERROR_VERBOSITY_UNSPECIFIED"
  - 1: "LOG_ERROR_VERBOSITY_TERSE"
  - 2: "LOG_ERROR_VERBOSITY_DEFAULT"
  - 3: "LOG_ERROR_VERBOSITY_VERBOSE"

* `log_lock_waits` boolean

* `log_min_duration_sample` integer

* `log_min_duration_statement` integer

* `log_min_error_statement` one of:
  - 0: "LOG_LEVEL_UNSPECIFIED"
  - 1: "LOG_LEVEL_DEBUG5"
  - 2: "LOG_LEVEL_DEBUG4"
  - 3: "LOG_LEVEL_DEBUG3"
  - 4: "LOG_LEVEL_DEBUG2"
  - 5: "LOG_LEVEL_DEBUG1"
  - 6: "LOG_LEVEL_LOG"
  - 7: "LOG_LEVEL_NOTICE"
  - 8: "LOG_LEVEL_WARNING"
  - 9: "LOG_LEVEL_ERROR"
  - 10: "LOG_LEVEL_FATAL"
  - 11: "LOG_LEVEL_PANIC"

* `log_min_messages` one of:
  - 0: "LOG_LEVEL_UNSPECIFIED"
  - 1: "LOG_LEVEL_DEBUG5"
  - 2: "LOG_LEVEL_DEBUG4"
  - 3: "LOG_LEVEL_DEBUG3"
  - 4: "LOG_LEVEL_DEBUG2"
  - 5: "LOG_LEVEL_DEBUG1"
  - 6: "LOG_LEVEL_LOG"
  - 7: "LOG_LEVEL_NOTICE"
  - 8: "LOG_LEVEL_WARNING"
  - 9: "LOG_LEVEL_ERROR"
  - 10: "LOG_LEVEL_FATAL"
  - 11: "LOG_LEVEL_PANIC"

* `log_parameter_max_length` integer

* `log_parameter_max_length_on_error` integer

* `log_statement` one of:
  - 0: "LOG_STATEMENT_UNSPECIFIED"
  - 1: "LOG_STATEMENT_NONE"
  - 2: "LOG_STATEMENT_DDL"
  - 3: "LOG_STATEMENT_MOD"
  - 4: "LOG_STATEMENT_ALL"

* `log_statement_sample_rate` float

* `log_temp_files` integer

* `log_transaction_sample_rate` float

* `logical_decoding_work_mem` integer

* `maintenance_io_concurrency` integer

* `maintenance_work_mem` integer

* `max_connections` integer

* `max_locks_per_transaction` integer

* `max_parallel_maintenance_workers` integer

* `max_parallel_workers` integer

* `max_parallel_workers_per_gather` integer

* `max_pred_locks_per_transaction` integer

* `max_prepared_transactions` integer

* `max_slot_wal_keep_size` integer

* `max_standby_streaming_delay` integer

* `max_wal_size` integer

* `max_worker_processes` integer

* `min_wal_size` integer

* `old_snapshot_threshold` integer

* `operator_precedence_warning` boolean

* `parallel_leader_participation` boolean

* `pg_hint_plan_debug_print` one of:
  - 0: "PG_HINT_PLAN_DEBUG_PRINT_UNSPECIFIED"
  - 1: "PG_HINT_PLAN_DEBUG_PRINT_OFF"
  - 2: "PG_HINT_PLAN_DEBUG_PRINT_ON"
  - 3: "PG_HINT_PLAN_DEBUG_PRINT_DETAILED"
  - 4: "PG_HINT_PLAN_DEBUG_PRINT_VERBOSE"

* `pg_hint_plan_enable_hint` boolean

* `pg_hint_plan_enable_hint_table` boolean

* `pg_hint_plan_message_level` one of:
  - 0: "LOG_LEVEL_UNSPECIFIED"
  - 1: "LOG_LEVEL_DEBUG5"
  - 2: "LOG_LEVEL_DEBUG4"
  - 3: "LOG_LEVEL_DEBUG3"
  - 4: "LOG_LEVEL_DEBUG2"
  - 5: "LOG_LEVEL_DEBUG1"
  - 6: "LOG_LEVEL_LOG"
  - 7: "LOG_LEVEL_NOTICE"
  - 8: "LOG_LEVEL_WARNING"
  - 9: "LOG_LEVEL_ERROR"
  - 10: "LOG_LEVEL_FATAL"
  - 11: "LOG_LEVEL_PANIC"

* `plan_cache_mode` one of:
  - 0: "PLAN_CACHE_MODE_UNSPECIFIED"
  - 1: "PLAN_CACHE_MODE_AUTO"
  - 2: "PLAN_CACHE_MODE_FORCE_CUSTOM_PLAN"
  - 3: "PLAN_CACHE_MODE_FORCE_GENERIC_PLAN"

* `quote_all_identifiers` boolean

* `random_page_cost` float

* `row_security` boolean

* `search_path` text

* `seq_page_cost` float

* `shared_buffers` integer

* `standard_conforming_strings` boolean

* `statement_timeout` integer

* `synchronize_seqscans` boolean

* `synchronous_commit` one of:
  - 0: "SYNCHRONOUS_COMMIT_UNSPECIFIED"
  - 1: "SYNCHRONOUS_COMMIT_ON"
  - 2: "SYNCHRONOUS_COMMIT_OFF"
  - 3: "SYNCHRONOUS_COMMIT_LOCAL"
  - 4: "SYNCHRONOUS_COMMIT_REMOTE_WRITE"
  - 5: "SYNCHRONOUS_COMMIT_REMOTE_APPLY"

* `temp_buffers` integer

* `temp_file_limit` integer

* `timezone` text

* `track_activity_query_size` integer

* `transform_null_equals` boolean

* `vacuum_cleanup_index_scale_factor` float

* `vacuum_cost_delay` integer

* `vacuum_cost_limit` integer

* `vacuum_cost_page_dirty` integer

* `vacuum_cost_page_hit` integer

* `vacuum_cost_page_miss` integer

* `wal_keep_size` integer

* `wal_level` one of:
  - 0: "WAL_LEVEL_UNSPECIFIED"
  - 1: "WAL_LEVEL_REPLICA"
  - 2: "WAL_LEVEL_LOGICAL"

* `work_mem` integer

* `xmlbinary` one of:
  - 0: "XML_BINARY_UNSPECIFIED"
  - 1: "XML_BINARY_BASE64"
  - 2: "XML_BINARY_HEX"

* `xmloption` one of:
  - 0: "XML_OPTION_UNSPECIFIED"
  - 1: "XML_OPTION_DOCUMENT"
  - 2: "XML_OPTION_CONTENT"

### Postgresql 12 config

* `archive_timeout` integer

* `array_nulls` boolean

* `auto_explain_log_analyze` boolean

* `auto_explain_log_buffers` boolean

* `auto_explain_log_min_duration` integer

* `auto_explain_log_nested_statements` boolean

* `auto_explain_log_timing` boolean

* `auto_explain_log_triggers` boolean

* `auto_explain_log_verbose` boolean

* `auto_explain_sample_rate` float

* `autovacuum_analyze_scale_factor` float

* `autovacuum_max_workers` integer

* `autovacuum_naptime` integer

* `autovacuum_vacuum_cost_delay` integer

* `autovacuum_vacuum_cost_limit` integer

* `autovacuum_vacuum_scale_factor` float

* `autovacuum_work_mem` integer

* `backend_flush_after` integer

* `backslash_quote` one of:
* * 0: "BACKSLASH_QUOTE_UNSPECIFIED"
* * 1: "BACKSLASH_QUOTE"
* * 2: "BACKSLASH_QUOTE_ON"
* * 3: "BACKSLASH_QUOTE_OFF"
* * 4: "BACKSLASH_QUOTE_SAFE_ENCODING"


* `bgwriter_delay` integer

* `bgwriter_flush_after` integer

* `bgwriter_lru_maxpages` integer

* `bgwriter_lru_multiplier` float

* `bytea_output` one of:
* * 0: "BYTEA_OUTPUT_UNSPECIFIED"
* * 1: "BYTEA_OUTPUT_HEX"
* * 2: "BYTEA_OUTPUT_ESCAPED"


* `checkpoint_completion_target` float

* `checkpoint_flush_after` integer

* `checkpoint_timeout` integer

* `client_min_messages` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `constraint_exclusion` one of:
* * 0: "CONSTRAINT_EXCLUSION_UNSPECIFIED"
* * 1: "CONSTRAINT_EXCLUSION_ON"
* * 2: "CONSTRAINT_EXCLUSION_OFF"
* * 3: "CONSTRAINT_EXCLUSION_PARTITION"


* `cursor_tuple_fraction` float

* `deadlock_timeout` integer

* `default_statistics_target` integer

* `default_transaction_isolation` one of:
* * 0: "TRANSACTION_ISOLATION_UNSPECIFIED"
* * 1: "TRANSACTION_ISOLATION_READ_UNCOMMITTED"
* * 2: "TRANSACTION_ISOLATION_READ_COMMITTED"
* * 3: "TRANSACTION_ISOLATION_REPEATABLE_READ"
* * 4: "TRANSACTION_ISOLATION_SERIALIZABLE"


* `default_transaction_read_only` boolean

* `default_with_oids` boolean

* `effective_cache_size` integer

* `effective_io_concurrency` integer

* `enable_bitmapscan` boolean

* `enable_hashagg` boolean

* `enable_hashjoin` boolean

* `enable_indexonlyscan` boolean

* `enable_indexscan` boolean

* `enable_material` boolean

* `enable_mergejoin` boolean

* `enable_nestloop` boolean

* `enable_parallel_append` boolean

* `enable_parallel_hash` boolean

* `enable_partition_pruning` boolean

* `enable_partitionwise_aggregate` boolean

* `enable_partitionwise_join` boolean

* `enable_seqscan` boolean

* `enable_sort` boolean

* `enable_tidscan` boolean

* `escape_string_warning` boolean

* `exit_on_error` boolean

* `force_parallel_mode` one of:
* * 0: "FORCE_PARALLEL_MODE_UNSPECIFIED"
* * 1: "FORCE_PARALLEL_MODE_ON"
* * 2: "FORCE_PARALLEL_MODE_OFF"
* * 3: "FORCE_PARALLEL_MODE_REGRESS"


* `from_collapse_limit` integer

* `gin_pending_list_limit` integer

* `idle_in_transaction_session_timeout` integer

* `jit` boolean

* `join_collapse_limit` integer

* `lo_compat_privileges` boolean

* `lock_timeout` integer

* `log_checkpoints` boolean

* `log_connections` boolean

* `log_disconnections` boolean

* `log_duration` boolean

* `log_error_verbosity` one of:
* * 0: "LOG_ERROR_VERBOSITY_UNSPECIFIED"
* * 1: "LOG_ERROR_VERBOSITY_TERSE"
* * 2: "LOG_ERROR_VERBOSITY_DEFAULT"
* * 3: "LOG_ERROR_VERBOSITY_VERBOSE"


* `log_lock_waits` boolean

* `log_min_duration_statement` integer

* `log_min_error_statement` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `log_min_messages` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `log_statement` one of:
* * 0: "LOG_STATEMENT_UNSPECIFIED"
* * 1: "LOG_STATEMENT_NONE"
* * 2: "LOG_STATEMENT_DDL"
* * 3: "LOG_STATEMENT_MOD"
* * 4: "LOG_STATEMENT_ALL"


* `log_temp_files` integer

* `log_transaction_sample_rate` float

* `maintenance_work_mem` integer

* `max_connections` integer

* `max_locks_per_transaction` integer

* `max_parallel_maintenance_workers` integer

* `max_parallel_workers` integer

* `max_parallel_workers_per_gather` integer

* `max_pred_locks_per_transaction` integer

* `max_prepared_transactions` integer

* `max_standby_streaming_delay` integer

* `max_wal_size` integer

* `max_worker_processes` integer

* `min_wal_size` integer

* `old_snapshot_threshold` integer

* `operator_precedence_warning` boolean

* `parallel_leader_participation` boolean

* `pg_hint_plan_debug_print` one of:
* * 0: "PG_HINT_PLAN_DEBUG_PRINT_UNSPECIFIED"
* * 1: "PG_HINT_PLAN_DEBUG_PRINT_OFF"
* * 2: "PG_HINT_PLAN_DEBUG_PRINT_ON"
* * 3: "PG_HINT_PLAN_DEBUG_PRINT_DETAILED"
* * 4: "PG_HINT_PLAN_DEBUG_PRINT_VERBOSE"


* `pg_hint_plan_enable_hint` boolean

* `pg_hint_plan_enable_hint_table` boolean

* `pg_hint_plan_message_level` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `plan_cache_mode` one of:
* * 0: "PLAN_CACHE_MODE_UNSPECIFIED"
* * 1: "PLAN_CACHE_MODE_AUTO"
* * 2: "PLAN_CACHE_MODE_FORCE_CUSTOM_PLAN"
* * 3: "PLAN_CACHE_MODE_FORCE_GENERIC_PLAN"


* `quote_all_identifiers` boolean

* `random_page_cost` float

* `row_security` boolean

* `search_path` text

* `seq_page_cost` float

* `shared_buffers` integer

* `shared_preload_libraries` override if not set. One of:
* * "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
* * "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"
* * "SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
* * NO value


* `standard_conforming_strings` boolean

* `statement_timeout` integer

* `synchronize_seqscans` boolean

* `synchronous_commit` one of:
* * 0: "SYNCHRONOUS_COMMIT_UNSPECIFIED"
* * 1: "SYNCHRONOUS_COMMIT_ON"
* * 2: "SYNCHRONOUS_COMMIT_OFF"
* * 3: "SYNCHRONOUS_COMMIT_LOCAL"
* * 4: "SYNCHRONOUS_COMMIT_REMOTE_WRITE"
* * 5: "SYNCHRONOUS_COMMIT_REMOTE_APPLY"


* `temp_buffers` integer

* `temp_file_limit` integer

* `timezone` text

* `track_activity_query_size` integer

* `transform_null_equals` boolean

* `vacuum_cleanup_index_scale_factor` float

* `vacuum_cost_delay` integer

* `vacuum_cost_limit` integer

* `vacuum_cost_page_dirty` integer

* `vacuum_cost_page_hit` integer

* `vacuum_cost_page_miss` integer

* `wal_level` one of:
* * 0: "WAL_LEVEL_UNSPECIFIED"
* * 1: "WAL_LEVEL_REPLICA"
* * 2: "WAL_LEVEL_LOGICAL"


* `work_mem` integer

* `xmlbinary` one of:
* * 0: "XML_BINARY_UNSPECIFIED"
* * 1: "XML_BINARY_BASE64"
* * 2: "XML_BINARY_HEX"


* `xmloption` one of:
* * 0: "XML_OPTION_UNSPECIFIED"
* * 1: "XML_OPTION_DOCUMENT"
* * 2: "XML_OPTION_CONTENT"


### Postgresql 11 config

* `archive_timeout` integer

* `array_nulls` boolean

* `auto_explain_log_analyze` boolean

* `auto_explain_log_buffers` boolean

* `auto_explain_log_min_duration` integer

* `auto_explain_log_nested_statements` boolean

* `auto_explain_log_timing` boolean

* `auto_explain_log_triggers` boolean

* `auto_explain_log_verbose` boolean

* `auto_explain_sample_rate` float

* `autovacuum_analyze_scale_factor` float

* `autovacuum_max_workers` integer

* `autovacuum_naptime` integer

* `autovacuum_vacuum_cost_delay` integer

* `autovacuum_vacuum_cost_limit` integer

* `autovacuum_vacuum_scale_factor` float

* `autovacuum_work_mem` integer

* `backend_flush_after` integer

* `backslash_quote` one of:
* * 0: "BACKSLASH_QUOTE_UNSPECIFIED"
* * 1: "BACKSLASH_QUOTE"
* * 2: "BACKSLASH_QUOTE_ON"
* * 3: "BACKSLASH_QUOTE_OFF"
* * 4: "BACKSLASH_QUOTE_SAFE_ENCODING"


* `bgwriter_delay` integer

* `bgwriter_flush_after` integer

* `bgwriter_lru_maxpages` integer

* `bgwriter_lru_multiplier` float

* `bytea_output` one of:
* * 0: "BYTEA_OUTPUT_UNSPECIFIED"
* * 1: "BYTEA_OUTPUT_HEX"
* * 2: "BYTEA_OUTPUT_ESCAPED"


* `checkpoint_completion_target` float

* `checkpoint_flush_after` integer

* `checkpoint_timeout` integer

* `client_min_messages` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `constraint_exclusion` one of:
* * 0: "CONSTRAINT_EXCLUSION_UNSPECIFIED"
* * 1: "CONSTRAINT_EXCLUSION_ON"
* * 2: "CONSTRAINT_EXCLUSION_OFF"
* * 3: "CONSTRAINT_EXCLUSION_PARTITION"


* `cursor_tuple_fraction` float

* `deadlock_timeout` integer

* `default_statistics_target` integer

* `default_transaction_isolation` one of:
* * 0: "TRANSACTION_ISOLATION_UNSPECIFIED"
* * 1: "TRANSACTION_ISOLATION_READ_UNCOMMITTED"
* * 2: "TRANSACTION_ISOLATION_READ_COMMITTED"
* * 3: "TRANSACTION_ISOLATION_REPEATABLE_READ"
* * 4: "TRANSACTION_ISOLATION_SERIALIZABLE"


* `default_transaction_read_only` boolean

* `default_with_oids` boolean

* `effective_cache_size` integer

* `effective_io_concurrency` integer

* `enable_bitmapscan` boolean

* `enable_hashagg` boolean

* `enable_hashjoin` boolean

* `enable_indexonlyscan` boolean

* `enable_indexscan` boolean

* `enable_material` boolean

* `enable_mergejoin` boolean

* `enable_nestloop` boolean

* `enable_parallel_append` boolean

* `enable_parallel_hash` boolean

* `enable_partition_pruning` boolean

* `enable_partitionwise_aggregate` boolean

* `enable_partitionwise_join` boolean

* `enable_seqscan` boolean

* `enable_sort` boolean

* `enable_tidscan` boolean

* `escape_string_warning` boolean

* `exit_on_error` boolean

* `force_parallel_mode` one of:
* * 0: "FORCE_PARALLEL_MODE_UNSPECIFIED"
* * 1: "FORCE_PARALLEL_MODE_ON"
* * 2: "FORCE_PARALLEL_MODE_OFF"
* * 3: "FORCE_PARALLEL_MODE_REGRESS"


* `from_collapse_limit` integer

* `gin_pending_list_limit` integer

* `idle_in_transaction_session_timeout` integer

* `jit` boolean

* `join_collapse_limit` integer

* `lo_compat_privileges` boolean

* `lock_timeout` integer

* `log_checkpoints` boolean

* `log_connections` boolean

* `log_disconnections` boolean

* `log_duration` boolean

* `log_error_verbosity` one of:
* * 0: "LOG_ERROR_VERBOSITY_UNSPECIFIED"
* * 1: "LOG_ERROR_VERBOSITY_TERSE"
* * 2: "LOG_ERROR_VERBOSITY_DEFAULT"
* * 3: "LOG_ERROR_VERBOSITY_VERBOSE"


* `log_lock_waits` boolean

* `log_min_duration_statement` integer

* `log_min_error_statement` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `log_min_messages` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `log_statement` one of:
* * 0: "LOG_STATEMENT_UNSPECIFIED"
* * 1: "LOG_STATEMENT_NONE"
* * 2: "LOG_STATEMENT_DDL"
* * 3: "LOG_STATEMENT_MOD"
* * 4: "LOG_STATEMENT_ALL"


* `log_temp_files` integer

* `maintenance_work_mem` integer

* `max_connections` integer

* `max_locks_per_transaction` integer

* `max_parallel_maintenance_workers` integer

* `max_parallel_workers` integer

* `max_parallel_workers_per_gather` integer

* `max_pred_locks_per_transaction` integer

* `max_prepared_transactions` integer

* `max_standby_streaming_delay` integer

* `max_wal_size` integer

* `max_worker_processes` integer

* `min_wal_size` integer

* `old_snapshot_threshold` integer

* `operator_precedence_warning` boolean

* `parallel_leader_participation` boolean

* `pg_hint_plan_debug_print` one of:
* * 0: "PG_HINT_PLAN_DEBUG_PRINT_UNSPECIFIED"
* * 1: "PG_HINT_PLAN_DEBUG_PRINT_OFF"
* * 2: "PG_HINT_PLAN_DEBUG_PRINT_ON"
* * 3: "PG_HINT_PLAN_DEBUG_PRINT_DETAILED"
* * 4: "PG_HINT_PLAN_DEBUG_PRINT_VERBOSE"


* `pg_hint_plan_enable_hint` boolean

* `pg_hint_plan_enable_hint_table` boolean

* `pg_hint_plan_message_level` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `quote_all_identifiers` boolean

* `random_page_cost` float

* `row_security` boolean

* `search_path` text

* `seq_page_cost` float

* `shared_buffers` integer

* `shared_preload_libraries` override if not set. One of:
* * "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
* * "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"
* * "SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
* * NO value


* `standard_conforming_strings` boolean

* `statement_timeout` integer

* `synchronize_seqscans` boolean

* `synchronous_commit` one of:
* * 0: "SYNCHRONOUS_COMMIT_UNSPECIFIED"
* * 1: "SYNCHRONOUS_COMMIT_ON"
* * 2: "SYNCHRONOUS_COMMIT_OFF"
* * 3: "SYNCHRONOUS_COMMIT_LOCAL"
* * 4: "SYNCHRONOUS_COMMIT_REMOTE_WRITE"
* * 5: "SYNCHRONOUS_COMMIT_REMOTE_APPLY"


* `temp_buffers` integer

* `temp_file_limit` integer

* `timezone` text

* `track_activity_query_size` integer

* `transform_null_equals` boolean

* `vacuum_cleanup_index_scale_factor` float

* `vacuum_cost_delay` integer

* `vacuum_cost_limit` integer

* `vacuum_cost_page_dirty` integer

* `vacuum_cost_page_hit` integer

* `vacuum_cost_page_miss` integer

* `wal_level` one of:
* * 0: "WAL_LEVEL_UNSPECIFIED"
* * 1: "WAL_LEVEL_REPLICA"
* * 2: "WAL_LEVEL_LOGICAL"


* `work_mem` integer

* `xmlbinary` one of:
* * 0: "XML_BINARY_UNSPECIFIED"
* * 1: "XML_BINARY_BASE64"
* * 2: "XML_BINARY_HEX"


* `xmloption` one of:
* * 0: "XML_OPTION_UNSPECIFIED"
* * 1: "XML_OPTION_DOCUMENT"
* * 2: "XML_OPTION_CONTENT"

### Postgresql 10 config

* `archive_timeout` integer

* `array_nulls` boolean

* `auto_explain_log_analyze` boolean

* `auto_explain_log_buffers` boolean

* `auto_explain_log_min_duration` integer

* `auto_explain_log_nested_statements` boolean

* `auto_explain_log_timing` boolean

* `auto_explain_log_triggers` boolean

* `auto_explain_log_verbose` boolean

* `auto_explain_sample_rate` float

* `autovacuum_analyze_scale_factor` float

* `autovacuum_max_workers` integer

* `autovacuum_naptime` integer

* `autovacuum_vacuum_cost_delay` integer

* `autovacuum_vacuum_cost_limit` integer

* `autovacuum_vacuum_scale_factor` float

* `autovacuum_work_mem` integer

* `backend_flush_after` integer

* `backslash_quote` one of:
* * 0: "BACKSLASH_QUOTE_UNSPECIFIED"
* * 1: "BACKSLASH_QUOTE"
* * 2: "BACKSLASH_QUOTE_ON"
* * 3: "BACKSLASH_QUOTE_OFF"
* * 4: "BACKSLASH_QUOTE_SAFE_ENCODING"


* `bgwriter_delay` integer

* `bgwriter_flush_after` integer

* `bgwriter_lru_maxpages` integer

* `bgwriter_lru_multiplier` float

* `bytea_output` one of:
* * 0: "BYTEA_OUTPUT_UNSPECIFIED"
* * 1: "BYTEA_OUTPUT_HEX"
* * 2: "BYTEA_OUTPUT_ESCAPED"


* `checkpoint_completion_target` float

* `checkpoint_flush_after` integer

* `checkpoint_timeout` integer

* `client_min_messages` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `constraint_exclusion` one of:
* * 0: "CONSTRAINT_EXCLUSION_UNSPECIFIED"
* * 1: "CONSTRAINT_EXCLUSION_ON"
* * 2: "CONSTRAINT_EXCLUSION_OFF"
* * 3: "CONSTRAINT_EXCLUSION_PARTITION"


* `cursor_tuple_fraction` float

* `deadlock_timeout` integer

* `default_statistics_target` integer

* `default_transaction_isolation` one of:
* * 0: "TRANSACTION_ISOLATION_UNSPECIFIED"
* * 1: "TRANSACTION_ISOLATION_READ_UNCOMMITTED"
* * 2: "TRANSACTION_ISOLATION_READ_COMMITTED"
* * 3: "TRANSACTION_ISOLATION_REPEATABLE_READ"
* * 4: "TRANSACTION_ISOLATION_SERIALIZABLE"


* `default_transaction_read_only` boolean

* `default_with_oids` boolean

* `effective_cache_size` integer

* `effective_io_concurrency` integer

* `enable_bitmapscan` boolean

* `enable_hashagg` boolean

* `enable_hashjoin` boolean

* `enable_indexonlyscan` boolean

* `enable_indexscan` boolean

* `enable_material` boolean

* `enable_mergejoin` boolean

* `enable_nestloop` boolean

* `enable_seqscan` boolean

* `enable_sort` boolean

* `enable_tidscan` boolean

* `escape_string_warning` boolean

* `exit_on_error` boolean

* `force_parallel_mode` one of:
* * 0: "FORCE_PARALLEL_MODE_UNSPECIFIED"
* * 1: "FORCE_PARALLEL_MODE_ON"
* * 2: "FORCE_PARALLEL_MODE_OFF"
* * 3: "FORCE_PARALLEL_MODE_REGRESS"


* `from_collapse_limit` integer

* `gin_pending_list_limit` integer

* `idle_in_transaction_session_timeout` integer

* `join_collapse_limit` integer

* `lo_compat_privileges` boolean

* `lock_timeout` integer

* `log_checkpoints` boolean

* `log_connections` boolean

* `log_disconnections` boolean

* `log_duration` boolean

* `log_error_verbosity` one of:
* * 0: "LOG_ERROR_VERBOSITY_UNSPECIFIED"
* * 1: "LOG_ERROR_VERBOSITY_TERSE"
* * 2: "LOG_ERROR_VERBOSITY_DEFAULT"
* * 3: "LOG_ERROR_VERBOSITY_VERBOSE"


* `log_lock_waits` boolean

* `log_min_duration_statement` integer

* `log_min_error_statement` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `log_min_messages` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `log_statement` one of:
* * 0: "LOG_STATEMENT_UNSPECIFIED"
* * 1: "LOG_STATEMENT_NONE"
* * 2: "LOG_STATEMENT_DDL"
* * 3: "LOG_STATEMENT_MOD"
* * 4: "LOG_STATEMENT_ALL"


* `log_temp_files` integer

* `maintenance_work_mem` integer

* `max_connections` integer

* `max_locks_per_transaction` integer

* `max_parallel_workers` integer

* `max_parallel_workers_per_gather` integer

* `max_pred_locks_per_transaction` integer

* `max_prepared_transactions` integer

* `max_standby_streaming_delay` integer

* `max_wal_size` integer

* `max_worker_processes` integer

* `min_wal_size` integer

* `old_snapshot_threshold` integer

* `operator_precedence_warning` boolean

* `pg_hint_plan_debug_print` one of:
* * 0: "PG_HINT_PLAN_DEBUG_PRINT_UNSPECIFIED"
* * 1: "PG_HINT_PLAN_DEBUG_PRINT_OFF"
* * 2: "PG_HINT_PLAN_DEBUG_PRINT_ON"
* * 3: "PG_HINT_PLAN_DEBUG_PRINT_DETAILED"
* * 4: "PG_HINT_PLAN_DEBUG_PRINT_VERBOSE"


* `pg_hint_plan_enable_hint` boolean

* `pg_hint_plan_enable_hint_table` boolean

* `pg_hint_plan_message_level` one of:
* * 0: "LOG_LEVEL_UNSPECIFIED"
* * 1: "LOG_LEVEL_DEBUG5"
* * 2: "LOG_LEVEL_DEBUG4"
* * 3: "LOG_LEVEL_DEBUG3"
* * 4: "LOG_LEVEL_DEBUG2"
* * 5: "LOG_LEVEL_DEBUG1"
* * 6: "LOG_LEVEL_LOG"
* * 7: "LOG_LEVEL_NOTICE"
* * 8: "LOG_LEVEL_WARNING"
* * 9: "LOG_LEVEL_ERROR"
* * 10: "LOG_LEVEL_FATAL"
* * 11: "LOG_LEVEL_PANIC"


* `quote_all_identifiers` boolean

* `random_page_cost` float

* `replacement_sort_tuples` integer

* `row_security` boolean

* `search_path` text

* `seq_page_cost` float

* `shared_buffers` integer

* `shared_preload_libraries` override if not set. One of:
* * "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
* * "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"
* * "SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN"
* * NO value


* `standard_conforming_strings` boolean

* `statement_timeout` integer

* `synchronize_seqscans` boolean

* `synchronous_commit` one of:
* * 0: "SYNCHRONOUS_COMMIT_UNSPECIFIED"
* * 1: "SYNCHRONOUS_COMMIT_ON"
* * 2: "SYNCHRONOUS_COMMIT_OFF"
* * 3: "SYNCHRONOUS_COMMIT_LOCAL"
* * 4: "SYNCHRONOUS_COMMIT_REMOTE_WRITE"
* * 5: "SYNCHRONOUS_COMMIT_REMOTE_APPLY"


* `temp_buffers` integer

* `temp_file_limit` integer

* `timezone` text

* `track_activity_query_size` integer

* `transform_null_equals` boolean

* `vacuum_cost_delay` integer

* `vacuum_cost_limit` integer

* `vacuum_cost_page_dirty` integer

* `vacuum_cost_page_hit` integer

* `vacuum_cost_page_miss` integer

* `wal_level` one of:
* * 0: "WAL_LEVEL_UNSPECIFIED"
* * 1: "WAL_LEVEL_REPLICA"
* * 2: "WAL_LEVEL_LOGICAL"


* `work_mem` integer

* `xmlbinary` one of:
* * 0: "XML_BINARY_UNSPECIFIED"
* * 1: "XML_BINARY_BASE64"
* * 2: "XML_BINARY_HEX"


* `xmloption` one of:
* * 0: "XML_OPTION_UNSPECIFIED"
* * 1: "XML_OPTION_DOCUMENT"
* * 2: "XML_OPTION_CONTENT"
