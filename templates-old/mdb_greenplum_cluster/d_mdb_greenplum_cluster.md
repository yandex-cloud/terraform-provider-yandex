---
subcategory: "Managed Service for Greenplum"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Managed Greenplum cluster.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Managed Greenplum cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-greenplum/).

## Example usage

{{ tffile "examples/mdb_greenplum_cluster/d_mdb_greenplum_cluster_1.tf" }}

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the Greenplum cluster.

* `name` - (Optional) The name of the Greenplum cluster.

~> Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `network_id` - ID of the network, to which the Greenplum cluster belongs.
* `created_at` - Timestamp of cluster creation.
* `description` - Description of the Greenplum cluster.
* `labels` - A set of key/value label pairs to assign to the Greenplum cluster.
* `environment` - Deployment environment of the Greenplum cluster.
* `version` - Version of the Greenplum cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.

* `zone` - The availability zone where the Greenplum hosts will be created.
* `subnet_id` - The ID of the subnet, to which the hosts belongs. The subnet must be a part of the network to which the cluster belongs.
* `assign_public_ip` - Sets whether the master hosts should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment.
* `master_host_count` - Number of hosts in master subcluster.
* `segment_host_count` - Number of hosts in segment subcluster.
* `segment_in_host` - Number of segments on segment host.
* `service_account_id` - (Optional) ID of service account to use with Yandex Cloud resources (e.g. S3, Cloud Logging).

* `master_subcluster` - Settings for master subcluster. The structure is documented below.
* `segment_subcluster` - Settings for segment subcluster. The structure is documented below.

* `master_hosts` - Info about hosts in master subcluster. The structure is documented below.
* `segment_hosts` - Info about hosts in segment subcluster. The structure is documented below.

* `access` - Access policy to the Greenplum cluster. The structure is documented below.
* `maintenance_window` - Maintenance window settings of the Greenplum cluster. The structure is documented below.
* `backup_window_start` - Time to start the daily backup, in the UTC timezone. The structure is documented below.

* `pooler_config` - Configuration of the connection pooler. The structure is documented below.
* `pxf_config` - Configuration of the PXF daemon. The structure is documented below.
* `greenplum_config` - Greenplum cluster config.

* `user_name` - Greenplum cluster admin user name.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.
* `deletion_protection` - Flag to protect the cluster from deletion.
* `master_host_group_ids` - (Optional) A list of IDs of the host groups to place master subclusters' VMs of the cluster on.
* `segment_host_group_ids` - (Optional) A list of IDs of the host groups to place segment subclusters' VMs of the cluster on.
* `logging` - (Optional) Block to configure log delivery to Yandex Cloud Logging .

The `master_subcluster` block supports:
* `resources` - Resources allocated to hosts for master subcluster of the Greenplum cluster. The structure is documented below.

The `segment_subcluster` block supports:
* `resources` - Resources allocated to hosts for segment subcluster of the Greenplum cluster. The structure is documented below.

The `master_hosts` block supports:
* `assign_public_ip` - Flag that indicates whether master hosts was created with a public IP.
* `fqdn` - The fully qualified domain name of the host.

The `segment_hosts` block supports:
* `fqdn` - The fully qualified domain name of the host.

The `resources` block supports:
* `resources_preset_id` - The ID of the preset for computational resources available to a Greenplum host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-greenplum/concepts/instance-types).
* `disk_size` - Volume of the storage available to a Greenplum host, in gigabytes.
* `disk_type_id` - Type of the storage for Greenplum hosts.

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `access` block supports:

* `data_lens` - (Optional) Allow access for [Yandex DataLens](https://yandex.cloud/services/datalens).
* `web_sql` - (Optional) Allows access for [SQL queries in the management console](https://yandex.cloud/docs/managed-mysql/operations/web-sql-query).
* `data_transfer` - (Optional) Allow access for [DataTransfer](https://yandex.cloud/services/data-transfer)
* `yandex_query` - (Optional) Allow access for [Yandex Query](https://yandex.cloud/services/query)

The `maintenance_window` block supports:

* `type` - Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`.
* `day` - Day of the week (in `DDD` format). Value is one of: "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"
* `hour` - Hour of the day in UTC (in `HH` format). Value is between 1 and 24.

The `pooler_config` block supports:

* `pooling_mode` - Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool-string.
* `pool_size` - Value for `pool_size` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_size-integer).
* `pool_client_idle_timeout` - Value for `pool_client_idle_timeout` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_ttl-integer).
* `pool_idle_in_transaction_timeout` - Value for `pool_idle_in_transaction_timeout` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/docs/configuration/rules.md#pool_idle_in_transaction_timeout).

The `pxf_config` block supports:

* `connection_timeout` - The Tomcat server connection timeout for read operations in seconds. Value is between 5 and 600.
* `upload_timeout` - The Tomcat server connection timeout for write operations in seconds. Value is between 5 and 600.
* `max_threads` - The maximum number of PXF tomcat threads. Value is between 1 and 1024.
* `pool_allow_core_thread_timeout` - Identifies whether or not core streaming threads are allowed to time out.
* `pool_core_size` - The number of core streaming threads. Value is between 1 and 1024.
* `pool_queue_capacity` - The capacity of the core streaming thread pool queue. Value is positive.
* `pool_max_size` - The maximum allowed number of core streaming threads. Value is between 1 and 1024.
* `xmx` - Initial JVM heap size for PXF daemon. Value is between 64 and 16384.
* `xms` - Maximum JVM heap size for PXF daemon. Value is between 64 and 16384.

The `background_activities` block supports:

* `analyze_and_vacuum` - (Optional) Block to configure 'ANALYZE' and 'VACUUM' daily operations.
  * `start_time` - Time of day in 'HH:MM' format when scripts should run.
  * `analyze_timeout` - Maximum duration of the `ANALYZE` operation, in seconds. The default value is `36000`. As soon as this period expires, the `ANALYZE` operation will be forced to terminate.
  * `vacuum_timeout` - Maximum duration of the `VACUUM` operation, in seconds. The default value is `36000`. As soon as this period expires, the `VACUUM` operation will be forced to terminate.
* `query_killer_idle` - (Optional) Block to configure script that kills long running queries that are in `idle` state.
  * `enable` - Flag that indicates whether script is enabled.
  * `max_age` - Maximum duration for this type of queries (in seconds).
  * `ignore_users` - List of users to ignore when considering queries to terminate.
* `query_killer_idle_in_transaction` - (Optional) block to configure script that kills long running queries that are in `idle in transaction` state.
  * `enable` - Flag that indicates whether script is enabled.
  * `max_age` - Maximum duration for this type of queries (in seconds).
  * `ignore_users` - List of users to ignore when considering queries to terminate.
* `query_killer_long_running` - (Optional) block to configure script that kills long running queries (in any state).
  * `enable` - Flag that indicates whether script is enabled.
  * `max_age` - Maximum duration for this type of queries (in seconds).
  * `ignore_users` - List of users to ignore when considering queries to terminate.

The `logging` block supports:
* `enabled` - Cloud Logging enable/disable switch.
* `log_group_id` - Use this log group to deliver cluster logs to.
* `folder_id` - Use this folder's default log group to deliver cluster logs to.
* `command_center_enabled` - Enable Yandex Command Center logs delivery.
* `greenplum_enabled` - Enable Greenplum logs delivery.
* `pooler_enabled` - Enable Pooler logs delivery.