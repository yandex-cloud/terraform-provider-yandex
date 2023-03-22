---
layout: "yandex"
page_title: "Yandex: yandex_mdb_postgresql_cluster"
sidebar_current: "docs-yandex-datasource-mdb-postgresql-cluster"
description: |-
  Get information about a Yandex Managed PostgreSQL cluster.
---

# yandex\_mdb\_postgresql\_cluster

Get information about a Yandex Managed PostgreSQL cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-postgresql/).
[How to connect to the DB](https://cloud.yandex.com/en-ru/docs/managed-postgresql/quickstart#connect). To connect, use port 6432. The port number is not configurable.

## Example Usage

```hcl
data "yandex_mdb_postgresql_cluster" "foo" {
  name = "test"
}

output "fqdn" {
  value = "${data.yandex_mdb_postgresql_cluster.foo.host.0.fqdn}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the PostgreSQL cluster.

* `name` - (Optional) The name of the PostgreSQL cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `network_id` - ID of the network, to which the PostgreSQL cluster belongs.
* `created_at` - Timestamp of cluster creation.
* `description` - Description of the PostgreSQL cluster.
* `labels` - A set of key/value label pairs to assign to the PostgreSQL cluster.
* `environment` - Deployment environment of the PostgreSQL cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `config` - Configuration of the PostgreSQL cluster. The structure is documented below.
* `host` - List of all hosts of the PostgreSQL cluster. The structure is documented below.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.
* `maintenance_window` - Maintenance window settings of the PostgreSQL cluster. The structure is documented below.
* `database` - List of all databases of the PostgreSQL cluster. The structure is documented below.
* `user` - List of all users of the PostgreSQL cluster. The structure is documented below.

The `config` block supports:

* `version` - Version of the PostgreSQL cluster.
* `autofailover` - Configuration setting which enables/disables autofailover in cluster.
* `backup_retain_period_days` - The period in days during which backups are stored.
* `resources` - Resources allocated to hosts of the PostgreSQL cluster. The structure is documented below.
* `pooler_config` - Configuration of the connection pooler. The structure is documented below.
* `backup_window_start` - Time to start the daily backup, in the UTC timezone. The structure is documented below.
* `access` - Access policy to the PostgreSQL cluster. The structure is documented below.
* `performance_diagnostics` - Cluster performance diagnostics settings. The structure is documented below. [YC Documentation](https://cloud.yandex.com/docs/managed-postgresql/api-ref/grpc/cluster_service#PerformanceDiagnostics)
* `postgresql_config` - PostgreSQL cluster config.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a PostgreSQL host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/concepts/instance-types).
* `disk_size` - Volume of the storage available to a PostgreSQL host, in gigabytes.
* `disk_type_id` - Type of the storage for PostgreSQL hosts.

The `pooler_config` block supports:

* `pooling_mode` - Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool-string.
* `pool_discard` - Value for `pool_discard` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_discard-yesno).

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `access` block supports:

* `data_lens` - Allow access for [Yandex DataLens](https://cloud.yandex.com/services/datalens).
* `web_sql` - Allow access for [SQL queries in the management console](https://cloud.yandex.com/docs/managed-postgresql/operations/web-sql-query)
* `serverless` - Allow access for [connection to managed databases from functions](https://cloud.yandex.com/docs/functions/operations/database-connection)
* `data_transfer` - (Optional) Allow access for [DataTransfer](https://cloud.yandex.com/services/data-transfer)

The `performance_diagnostics` block supports:
* `enabled` - Flag, when true, performance diagnostics is enabled
* `sessions_sampling_interval` - Interval (in seconds) for pg_stat_activity sampling Acceptable values are 1 to 86400, inclusive.
* `statements_sampling_interval` - Interval (in seconds) for pg_stat_statements sampling Acceptable values are 1 to 86400, inclusive.

The `host` block supports:

* `fqdn` - The fully qualified domain name of the host.
* `zone` - The availability zone where the PostgreSQL host will be created.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.
* `assign_public_ip` - Sets whether the host should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment.
* `role` - Role of the host in the cluster.
* `replication_source` - Host replication source (fqdn), case when replication_source is empty then host in HA group.
* `priority` - Host priority in HA group.

The `maintenance_window` block supports:

* `type` - Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`.
* `day` - Day of the week (in `DDD` format). Value is one of: "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"
* `hour` - Hour of the day in UTC (in `HH` format). Value is between 1 and 24.

The `database` block supports:

* `owner` - Name of the user assigned as the owner of the database.
* `lc_collate` - POSIX locale for string sorting order. Forbidden to change in an existing database.
* `lc_type` - POSIX locale for character classification. Forbidden to change in an existing database.
* `extension` - Set of database extensions. The structure is documented below
* `template_db` - Name of the template database.

The `extension` block supports:

* `name` - Name of the database extension. For more information on available extensions see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/operations/cluster-extensions).
* `version` - Version of the extension.

The `user` block supports:

* `permission` - Set of permissions granted to the user. The structure is documented below.
* `login` - User's ability to login.
* `grants` - List of the user's grants.
* `conn_limit` - The maximum number of connections per user.
* `settings` - Map of user settings. The structure is documented below.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.

The `settings` block supports:
Full description https://cloud.yandex.com/en-ru/docs/managed-postgresql/api-ref/grpc/user_service#UserSettings1

* `default_transaction_isolation` - defines the default isolation level to be set for all new SQL transactions. One of:
  - 0: "unspecified"
  - 1: "read uncommitted"
  - 2: "read committed"
  - 3: "repeatable read"
  - 4: "serializable"

* `lock_timeout` - The maximum time (in milliseconds) for any statement to wait for acquiring a lock on an table, index, row or other database object (default 0)

* `log_min_duration_statement` - This setting controls logging of the duration of statements. (default -1 disables logging of the duration of statements.)

* `synchronous_commit` - This setting defines whether DBMS will commit transaction in a synchronous way. One of:
  - 0: "unspecified"
  - 1: "on"
  - 2: "off"
  - 3: "local"
  - 4: "remote write"
  - 5: "remote apply"

* `temp_file_limit` - The maximum storage space size (in kilobytes) that a single process can use to create temporary files.

* `log_statement` - This setting specifies which SQL statements should be logged (on the user level). One of:
  - 0: "unspecified"
  - 1: "none"
  - 2: "ddl"
  - 3: "mod"
  - 4: "all"

* `pool_mode` - Mode that the connection pooler is working in with specified user. One of:
  - 0: "session"
  - 1: "transaction"
  - 2: "statement"

* `prepared_statements_pooling` - This setting allows user to use prepared statements with transaction pooling. Boolean.

* `catchup_timeout` - The connection pooler setting. It determines the maximum allowed replication lag (in seconds). Pooler will reject connections to the replica with a lag above this threshold. Default value is 0, which disables this feature. Integer.

* `wal_sender_timeout` - The maximum time (in milliseconds) to wait for WAL replication (can be set only for PostgreSQL 12+). Terminate replication connections that are inactive for longer than this amount of time. Integer.

* `idle_in_transaction_session_timeout` - Sets the maximum allowed idle time (in milliseconds) between queries, when in a transaction. Value of 0 (default) disables the timeout. Integer.

* `statement_timeout` - The maximum time (in milliseconds) to wait for statement. Value of 0 (default) disables the timeout. Integer