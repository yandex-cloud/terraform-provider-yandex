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
* `host` - A host of the PostgreSQL cluster. The structure is documented below.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.
* `maintenance_window` - Maintenance window settings of the PostgreSQL cluster. The structure is documented below.

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


