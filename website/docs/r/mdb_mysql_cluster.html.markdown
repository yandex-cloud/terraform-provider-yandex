---
layout: "yandex"
page_title: "Yandex: yandex_mdb_mysql_cluster"
sidebar_current: "docs-yandex-mdb-mysql-cluster"
description: |-
  Manages a MySQL cluster within Yandex.Cloud.
---

# yandex\_mdb\_mysql\_cluster

Manages a MySQL cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-mysql/).

## Example Usage

Example of creating a Single Node MySQL.

```hcl
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  mysql_config = {
    sql_mode                      = "ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"
    max_connections               = 100
    default_authentication_plugin = "MYSQL_NATIVE_PASSWORD"
    innodb_print_all_deadlocks    = true

  }

  access {
    web_sql = true
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user_name"
    password = "your_password"
    permission {
      database_name = "db_name"
      roles         = ["ALL"]
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

Example of creating a High-Availability(HA) MySQL Cluster.

```hcl
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  database {
    name = "db_name"
  }

  maintenance_window {
    type = "WEEKLY"
    day  = "SAT"
    hour = 12
  }

  user {
    name     = "user_name"
    password = "your_password"
    permission {
      database_name = "db_name"
      roles         = ["ALL"]
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
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}
```

Example of creating a MySQL Cluster with cascade replicas: HA-group consist of 'na-1' and 'na-2', cascade replicas form a chain 'na-1' -> 'nb-1' -> 'nb-2'

```hcl
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  database {
    name = "db_name"
  }

  maintenance_window {
    type = "WEEKLY"
    day  = "SAT"
    hour = 12
  }

  user {
    name     = "user_name"
    password = "your_password"
    permission {
      database_name = "db_name"
      roles         = ["ALL"]
    }
  }

  host {
    zone      = "ru-central1-a"
    name      = "na-1"
    subnet_id = yandex_vpc_subnet.foo.id
  }
  host {
    zone      = "ru-central1-a"
    name      = "na-2"
    subnet_id = yandex_vpc_subnet.foo.id
  }
  host {
    zone                    = "ru-central1-b"
    name                    = "nb-1"
    replication_source_name = "na-1"
    subnet_id               = yandex_vpc_subnet.bar.id
  }
  host {
    zone                    = "ru-central1-b"
    name                    = "nb-2"
    replication_source_name = "nb-1"
    subnet_id               = yandex_vpc_subnet.bar.id
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
```


Example of creating a MySQL Cluster with different backup priorities. Backup will be created from nb-2, if it's not master. na-2 will be used as a backup source as a last resort.

```hcl
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  database {
    name = "db_name"
  }

  maintenance_window {
    type = "WEEKLY"
    day  = "SAT"
    hour = 12
  }

  user {
    name     = "user_name"
    password = "your_password"
    permission {
      database_name = "db_name"
      roles         = ["ALL"]
    }
  }

  host {
    zone      = "ru-central1-a"
    name      = "na-1"
    subnet_id = yandex_vpc_subnet.foo.id
  }
  host {
    zone            = "ru-central1-b"
    name            = "nb-1"
    backup_priority = 5
    subnet_id       = yandex_vpc_subnet.bar.id
  }
  host {
    zone            = "ru-central1-b"
    name            = "nb-2"
    backup_priority = 10
    subnet_id       = yandex_vpc_subnet.bar.id
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
```

Example of creating a MySQL Cluster with different host priorities. During failover master will be set to nb-2

```hcl
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  database {
    name = "db_name"
  }

  maintenance_window {
    type = "WEEKLY"
    day  = "SAT"
    hour = 12
  }

  user {
    name     = "user_name"
    password = "your_password"
    permission {
      database_name = "db_name"
      roles         = ["ALL"]
    }
  }

  host {
    zone      = "ru-central1-a"
    name      = "na-1"
    subnet_id = yandex_vpc_subnet.foo.id
  }
  host {
    zone            = "ru-central1-b"
    name            = "nb-1"
    priority        = 5
    subnet_id       = yandex_vpc_subnet.bar.id
  }
  host {
    zone            = "ru-central1-b"
    name            = "nb-2"
    priority        = 10
    subnet_id       = yandex_vpc_subnet.bar.id
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
```

Example of creating a Single Node MySQL with user params.

```hcl
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  database {
    name = "db_name"
  }

  maintenance_window {
    type = "ANYTIME"
  }

  user {
    name     = "user_name"
    password = "your_password"
    permission {
      database_name = "db_name"
      roles         = ["ALL"]
    }

    connection_limits {
      max_questions_per_hour = 10
    }

    global_permissions = ["REPLICATION_SLAVE", "PROCESS"]

    authentication_plugin = "CACHING_SHA2_PASSWORD"
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

Example of restoring MySQL cluster.

```hcl
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  restore {
    backup_id   = "c9qj2tns23432471d9qha:stream_20210122T141717Z"
    time = "2021-01-23T15:04:05"
  }

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user_name"
    password = "your_password"
    permission {
      database_name = "db_name"
      roles         = ["ALL"]
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

* `name` - (Required) Name of the MySQL cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the MySQL cluster uses.

* `environment` - (Required) Deployment environment of the MySQL cluster.

* `version` - (Required) Version of the MySQL cluster. (allowed versions are: 5.7, 8.0)

* `resources` - (Required) Resources allocated to hosts of the MySQL cluster. The structure is documented below.

* `user` - (Required) A user of the MySQL cluster. The structure is documented below.

* `database` - (Required) A database of the MySQL cluster. The structure is documented below.

* `host` - (Required) A host of the MySQL cluster. The structure is documented below.

* `access` - (Optional) Access policy to the MySQL cluster. The structure is documented below.

* `mysql_config` - (Optional) MySQL cluster config. Detail info in "MySQL config" section (documented below).

* `restore` - (Optional, ForceNew) The cluster will be created from the specified backup. The structure is documented below.

* `maintenance_window` - (Optional) Maintenance policy of the MySQL cluster. The structure is documented below.

- - -

* `description` - (Optional) Description of the MySQL cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the MySQL cluster.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC. The structure is documented below.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster.  Can be either `true` or `false`.

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a MySQL host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mysql/concepts/instance-types).

* `disk_size` - (Required) Volume of the storage available to a MySQL host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of MySQL hosts.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started.

* `minutes` - (Optional) The minute at which backup will be started.

The `user` block supports:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

* `global_permissions` - (Optional) List user's global permissions     
            Allowed permissions:  `REPLICATION_CLIENT`, `REPLICATION_SLAVE`, `PROCESS` for clear list use empty list.
            If the attribute is not specified there will be no changes.

* `connection_limits` - (Optional) User's connection limits. The structure is documented below.
            If the attribute is not specified there will be no changes.

* `authentication_plugin` - (Optional) Authentication plugin. Allowed values: `MYSQL_NATIVE_PASSWORD`, `CACHING_SHA2_PASSWORD`, `SHA256_PASSWORD` (for version 5.7 `MYSQL_NATIVE_PASSWORD`, `SHA256_PASSWORD`)

The `connection_limits` block supports:  
default value is -1,   
When these parameters are set to -1, backend default values will be actually used.   

* `max_questions_per_hour` - Max questions per hour.

* `max_updates_per_hour` - Max updates per hour.

* `max_connections_per_hour` - Max connections per hour.

* `max_user_connections` - Max user connections.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

* `roles` - (Optional) List user's roles in the database.
            Allowed roles: `ALL`,`ALTER`,`ALTER_ROUTINE`,`CREATE`,`CREATE_ROUTINE`,`CREATE_TEMPORARY_TABLES`,
            `CREATE_VIEW`,`DELETE`,`DROP`,`EVENT`,`EXECUTE`,`INDEX`,`INSERT`,`LOCK_TABLES`,`SELECT`,`SHOW_VIEW`,`TRIGGER`,`UPDATE`.

The `database` block supports:

* `name` - (Required) The name of the database.

The `host` block supports:

* `zone` - (Required) The availability zone where the MySQL host will be created.

* `fqdn` - (Computed) The fully qualified domain name of the host.

* `subnet_id` - (Optional) The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `assign_public_ip` - (Optional) Sets whether the host should get a public IP address. It can be changed on the fly only when `name` is set.

* `name` - (Optional) Host state name. It should be set for all hosts or unset for all hosts. This field can be used by another host, to select which host will be its replication source. Please refer to `replication_source_name` parameter.

* `replication_source` - (Computed) Host replication source (fqdn), when replication_source is empty then host is in HA group.

* `replication_source_name` - (Optional) Host replication source name points to host's `name` from which this host should replicate. When not set then host in HA group. It works only when `name` is set.

* `priority` - (Optional) Host master promotion priority. Value is between 0 and 100, default is 0.

* `backup_priority` - (Optional) Host backup priority. Value is between 0 and 100, default is 0.

The `access` block supports:
If not specified then does not make any changes.  

* `data_lens` - (Optional) Allow access for [Yandex DataLens](https://cloud.yandex.com/services/datalens).

* `web_sql` - Allows access for [SQL queries in the management console](https://cloud.yandex.com/docs/managed-mysql/operations/web-sql-query).

The `restore` block supports:

* `backup_id` - (Required, ForceNew) Backup ID. The cluster will be created from the specified backup. [How to get a list of MySQL backups](https://cloud.yandex.com/docs/managed-mysql/operations/cluster-backups). 

* `time` - (Optional, ForceNew) Timestamp of the moment to which the MySQL cluster should be restored. (Format: "2006-01-02T15:04:05" - UTC). When not set, current time is used.

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.

* `day` - (Optional) Day of the week (in `DDD` format). Allowed values: "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"

* `hour` - (Optional) Hour of the day in UTC (in `HH` format). Allowed value is between 0 and 23.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the cluster.

* `health` - Aggregated health of the cluster.

* `status` - Status of the cluster.

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_mysql_cluster.foo cluster_id
```

## MySQL config
If not specified `mysql_config` then does not make any changes.  

* `sql_mode` default value: `ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION`  

some of:  
	-	1: "ALLOW_INVALID_DATES"
	-	2: "ANSI_QUOTES"
	-	3: "ERROR_FOR_DIVISION_BY_ZERO"
	-	4: "HIGH_NOT_PRECEDENCE"
	-	5: "IGNORE_SPACE"
	-	6: "NO_AUTO_VALUE_ON_ZERO"
	-	7: "NO_BACKSLASH_ESCAPES"
	-	8: "NO_ENGINE_SUBSTITUTION"
	-	9: "NO_UNSIGNED_SUBTRACTION"
	-	10: "NO_ZERO_DATE"
	-	11: "NO_ZERO_IN_DATE"
	-	15: "ONLY_FULL_GROUP_BY"
	-	16: "PAD_CHAR_TO_FULL_LENGTH"
	-	17: "PIPES_AS_CONCAT"
	-	18: "REAL_AS_FLOAT"
	-	19: "STRICT_ALL_TABLES"
	-	20: "STRICT_TRANS_TABLES"
	-	21: "TIME_TRUNCATE_FRACTIONAL"
	-	22: "ANSI"
	-	23: "TRADITIONAL"
	-	24: "NO_DIR_IN_CREATE"
or:  
  - 0: "SQLMODE_UNSPECIFIED"

### MysqlConfig 8.0
* `audit_log` boolean

* `auto_increment_increment` integer

* `auto_increment_offset` integer

* `binlog_cache_size` integer

* `binlog_group_commit_sync_delay` integer

* `binlog_row_image` one of:
  - 0: "BINLOG_ROW_IMAGE_UNSPECIFIED"
  - 1: "FULL"
  - 2: "MINIMAL"
  - 3: "NOBLOB"

* `binlog_rows_query_log_events` boolean

* `character_set_server` text

* `collation_server` text

* `default_authentication_plugin` one of:
  - 0: "AUTH_PLUGIN_UNSPECIFIED"
  - 1: "MYSQL_NATIVE_PASSWORD"
  - 2: "CACHING_SHA2_PASSWORD"
  - 3: "SHA256_PASSWORD"

* `default_time_zone` text

* `explicit_defaults_for_timestamp` boolean

* `general_log` boolean

* `group_concat_max_len` integer

* `innodb_adaptive_hash_index` boolean

* `innodb_buffer_pool_size` integer

* `innodb_flush_log_at_trx_commit` integer

* `innodb_io_capacity` integer

* `innodb_io_capacity_max` integer

* `innodb_lock_wait_timeout` integer

* `innodb_log_buffer_size` integer

* `innodb_log_file_size` integer

* `innodb_numa_interleave` boolean

* `innodb_print_all_deadlocks` boolean

* `innodb_purge_threads` integer

* `innodb_read_io_threads` integer

* `innodb_temp_data_file_max_size` integer

* `innodb_thread_concurrency` integer

* `innodb_write_io_threads` integer

* `join_buffer_size` integer

* `long_query_time` float

* `max_allowed_packet` integer

* `max_connections` integer

* `max_heap_table_size` integer

* `net_read_timeout` integer

* `net_write_timeout` integer

* `regexp_time_limit` integer

* `rpl_semi_sync_master_wait_for_slave_count` integer

* `slave_parallel_type` one of:
  - 0: "SLAVE_PARALLEL_TYPE_UNSPECIFIED"
  - 1: "DATABASE"
  - 2: "LOGICAL_CLOCK"

* `slave_parallel_workers` integer

* `sort_buffer_size` integer

* `sync_binlog` integer

* `table_definition_cache` integer

* `table_open_cache` integer

* `table_open_cache_instances` integer

* `thread_cache_size` integer

* `thread_stack` integer

* `tmp_table_size` integer

* `transaction_isolation` one of:
  - 0: "TRANSACTION_ISOLATION_UNSPECIFIED"
  - 1: "READ_COMMITTED"
  - 2: "REPEATABLE_READ"
  - 3: "SERIALIZABLE"

### MysqlConfig 5.7
* `audit_log` boolean

* `auto_increment_increment` integer

* `auto_increment_offset` integer

* `binlog_cache_size` integer

* `binlog_group_commit_sync_delay` integer

* `binlog_row_image` one of:
  - 0: "BINLOG_ROW_IMAGE_UNSPECIFIED"
  - 1: "FULL"
  - 2: "MINIMAL"
  - 3: "NOBLOB"

* `binlog_rows_query_log_events` boolean

* `character_set_server` text

* `collation_server` text

* `default_authentication_plugin` one of:
  - 0: "AUTH_PLUGIN_UNSPECIFIED"
  - 1: "MYSQL_NATIVE_PASSWORD"
  - 2: "CACHING_SHA2_PASSWORD"
  - 3: "SHA256_PASSWORD"

* `default_time_zone` text

* `explicit_defaults_for_timestamp` boolean

* `general_log` boolean

* `group_concat_max_len` integer

* `innodb_adaptive_hash_index` boolean

* `innodb_buffer_pool_size` integer

* `innodb_flush_log_at_trx_commit` integer

* `innodb_io_capacity` integer

* `innodb_io_capacity_max` integer

* `innodb_lock_wait_timeout` integer

* `innodb_log_buffer_size` integer

* `innodb_log_file_size` integer

* `innodb_numa_interleave` boolean

* `innodb_print_all_deadlocks` boolean

* `innodb_purge_threads` integer

* `innodb_read_io_threads` integer

* `innodb_temp_data_file_max_size` integer

* `innodb_thread_concurrency` integer

* `innodb_write_io_threads` integer

* `join_buffer_size` integer

* `long_query_time` float

* `max_allowed_packet` integer

* `max_connections` integer

* `max_heap_table_size` integer

* `net_read_timeout` integer

* `net_write_timeout` integer

* `rpl_semi_sync_master_wait_for_slave_count` integer

* `slave_parallel_type` one of:
  - 0: "SLAVE_PARALLEL_TYPE_UNSPECIFIED"
  - 1: "DATABASE"
  - 2: "LOGICAL_CLOCK"

* `slave_parallel_workers` integer

* `sort_buffer_size` integer

* `sync_binlog` integer

* `table_definition_cache` integer

* `table_open_cache` integer

* `table_open_cache_instances` integer

* `thread_cache_size` integer

* `thread_stack` integer

* `tmp_table_size` integer

* `transaction_isolation` one of:
  - 0: "TRANSACTION_ISOLATION_UNSPECIFIED"
  - 1: "READ_COMMITTED"
  - 2: "REPEATABLE_READ"
  - 3: "SERIALIZABLE"
