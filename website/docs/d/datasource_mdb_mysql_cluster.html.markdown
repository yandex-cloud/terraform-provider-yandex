---
layout: "yandex"
page_title: "Yandex: yandex_mdb_mysql_cluster"
sidebar_current: "docs-yandex-datasource-mdb-mysql-cluster"
description: |-
  Get information about a Yandex Managed MySQL cluster.
---

# yandex\_mdb\_mysql\_cluster

Get information about a Yandex Managed MySQL cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-mysql/).

## Example Usage

```hcl
data "yandex_mdb_mysql_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = "${data.yandex_mdb_mysql_cluster.foo.network_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the MySQL cluster.

* `name` - (Optional) The name of the MySQL cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `network_id` - ID of the network, to which the MySQL cluster belongs.
* `created_at` - Creation timestamp of the key.
* `description` - Description of the MySQL cluster.
* `labels` - A set of key/value label pairs to assign to the MySQL cluster.
* `environment` - Deployment environment of the MySQL cluster.
* `version` - Version of the MySQL cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `resources` - Resources allocated to hosts of the MySQL cluster. The structure is documented below.
* `user` - A user of the MySQL cluster. The structure is documented below.
* `database` - A database of the MySQL cluster. The structure is documented below.
* `host` - A host of the MySQL cluster. The structure is documented below.
* `access` - Access policy to the MySQL cluster. The structure is documented below.
* `mysql_config` - MySQL cluster config.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.
* `maintenance_window` - Maintenance window settings of the MySQL cluster. The structure is documented below.
* `performance_diagnostics` - Cluster performance diagnostics settings. The structure is documented below. [YC Documentation](https://cloud.yandex.com/docs/managed-mysql/api-ref/grpc/cluster_service#PerformanceDiagnostics)

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a MySQL host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mysql/concepts/instance-types).
* `disk_size` - Volume of the storage available to a MySQL host, in gigabytes.
* `disk_type_id` - Type of the storage for MySQL hosts.

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `access` block supports:

* `data_lens` - (Optional) Allow access for [Yandex DataLens](https://cloud.yandex.com/services/datalens).
* `web_sql` - (Optional) Allows access for [SQL queries in the management console](https://cloud.yandex.com/docs/managed-mysql/operations/web-sql-query).
* `data_transfer` - (Optional) Allow access for [DataTransfer](https://cloud.yandex.com/services/data-transfer)

The `user` block supports:

* `name` - The name of the user.
* `password` - The password of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.
* `global_permissions` - List user's global permissions. Allowed values: `REPLICATION_CLIENT`, `REPLICATION_SLAVE`, `PROCESS` or empty list.
* `connection_limits` - User's connection limits. The structure is documented below.
* `authentication_plugin` - Authentication plugin. Allowed values: `MYSQL_NATIVE_PASSWORD`, `CACHING_SHA2_PASSWORD`, `SHA256_PASSWORD`

The `connection_limits` block supports:   
When these parameters are set to -1, backend default values will be actually used.   

* `max_questions_per_hour` - Max questions per hour.
* `max_updates_per_hour` - Max updates per hour.
* `max_connections_per_hour` - Max connections per hour.
* `max_user_connections` - Max user connections.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.
* `roles` - List user's roles in the database.
            Allowed roles: `ALL`,`ALTER`,`ALTER_ROUTINE`,`CREATE`,`CREATE_ROUTINE`,`CREATE_TEMPORARY_TABLES`,
            `CREATE_VIEW`,`DELETE`,`DROP`,`EVENT`,`EXECUTE`,`INDEX`,`INSERT`,`LOCK_TABLES`,`SELECT`,`SHOW_VIEW`,`TRIGGER`,`UPDATE`.

The `database` block supports:

* `name` - The name of the database.

The `host` block supports:

* `fqdn` - The fully qualified domain name of the host.
* `zone` - The availability zone where the MySQL host will be created.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.
* `assign_public_ip` - Sets whether the host should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment
* `replication_source` - Host replication source (fqdn), case when replication_source is empty then host in HA group.
* `priority` - Host master promotion priority. Value is between 0 and 100, default is 0. 
* `backup_priority` - Host backup priority. Value is between 0 and 100, default is 0. 

The `access` block supports:

* `data_lens` - Allow access for [Yandex DataLens](https://cloud.yandex.com/services/datalens).
* `web_sql` - Allows access for [SQL queries in the management console](https://cloud.yandex.com/docs/managed-mysql/operations/web-sql-query).

The `maintenance_window` block supports:

* `type` - Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`.
* `day` - Day of the week (in `DDD` format). Value is one of: "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"
* `hour` - Hour of the day in UTC (in `HH` format). Value is between 1 and 24.

The `performance_diagnostics` block supports:

* `enabled` - Enable performance diagnostics
* `sessions_sampling_interval` - Interval (in seconds) for my_stat_activity sampling Acceptable values are 1 to 86400, inclusive.
* `statements_sampling_interval` - Interval (in seconds) for my_stat_statements sampling Acceptable values are 1 to 86400, inclusive.
