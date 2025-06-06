---
subcategory: "Managed Service for Greenplum"
page_title: "Yandex: yandex_mdb_greenplum_cluster"
description: |-
  Manages a Greenplum cluster within Yandex Cloud.
---

# yandex_mdb_greenplum_cluster (Resource)

Manages a Greenplum cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-greenplum/).

Please read [Pricing for Managed Service for Greenplum](https://yandex.cloud/docs/managed-greenplum/) before using Greenplum cluster.

## Example usage

```terraform
//
// Create a new MDB Greenplum Cluster.
//
resource "yandex_mdb_greenplum_cluster" "my_cluster" {
  name               = "test"
  description        = "test greenplum cluster"
  environment        = "PRESTABLE"
  network_id         = yandex_vpc_network.foo.id
  zone_id            = "ru-central1-a"
  subnet_id          = yandex_vpc_subnet.foo.id
  assign_public_ip   = true
  version            = "6.25"
  master_host_count  = 2
  segment_host_count = 5
  segment_in_host    = 1
  master_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }
  segment_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }

  access {
    web_sql = true
  }

  greenplum_config = {
    max_connections                      = 395
    max_slot_wal_keep_size               = 1048576
    gp_workfile_limit_per_segment        = 0
    gp_workfile_limit_per_query          = 0
    gp_workfile_limit_files_per_query    = 100000
    max_prepared_transactions            = 500
    gp_workfile_compression              = "false"
    max_statement_mem                    = 2147483648
    log_statement                        = 2
    gp_add_column_inherits_table_setting = "true"
    gp_enable_global_deadlock_detector   = "true"
    gp_global_deadlock_detector_period   = 120
  }

  user_name     = "admin_user"
  user_password = "your_super_secret_password"

  security_group_ids = [yandex_vpc_security_group.test-sg-x.id]
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}

resource "yandex_vpc_security_group" "test-sg-x" {
  network_id = yandex_vpc_network.foo.id
  ingress {
    protocol       = "ANY"
    description    = "Allow incoming traffic from members of the same security group"
    from_port      = 0
    to_port        = 65535
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    protocol       = "ANY"
    description    = "Allow outgoing traffic to members of the same security group"
    from_port      = 0
    to_port        = 65535
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `assign_public_ip` (Boolean) Sets whether the master hosts should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment.
- `environment` (String) Deployment environment of the Greenplum cluster. (PRODUCTION, PRESTABLE)
- `master_host_count` (Number) Number of hosts in master subcluster (1 or 2).
- `master_subcluster` (Block List, Min: 1, Max: 1) Settings for master subcluster. (see [below for nested schema](#nestedblock--master_subcluster))
- `name` (String) The resource name.
- `network_id` (String) The `VPC Network ID` of subnets which resource attached to.
- `segment_host_count` (Number) Number of hosts in segment subcluster (from 1 to 32).
- `segment_in_host` (Number) Number of segments on segment host (not more then 1 + RAM/8).
- `segment_subcluster` (Block List, Min: 1, Max: 1) Settings for segment subcluster. (see [below for nested schema](#nestedblock--segment_subcluster))
- `subnet_id` (String) The ID of the subnet, to which the hosts belongs. The subnet must be a part of the network to which the cluster belongs.
- `user_name` (String) Greenplum cluster admin user name.
- `user_password` (String, Sensitive) Greenplum cluster admin password name.
- `version` (String) Version of the Greenplum cluster. (`6.25`)
- `zone` (String) The [availability zone](https://yandex.cloud/docs/overview/concepts/geo-scope) where resource is located. If it is not provided, the default provider zone will be used.

### Optional

- `access` (Block List, Max: 1) Access policy to the Greenplum cluster. (see [below for nested schema](#nestedblock--access))
- `background_activities` (Block List) Background activities settings. (see [below for nested schema](#nestedblock--background_activities))
- `backup_window_start` (Block List, Max: 1) Time to start the daily backup, in the UTC timezone. (see [below for nested schema](#nestedblock--backup_window_start))
- `cloud_storage` (Block List, Max: 1) Cloud Storage settings of the Greenplum cluster. (see [below for nested schema](#nestedblock--cloud_storage))
- `deletion_protection` (Boolean) The `true` value means that resource is protected from accidental deletion.
- `description` (String) The resource description.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `greenplum_config` (Map of String) Greenplum cluster config. Detail info in `Greenplum cluster settings` block.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `logging` (Block List, Max: 1) Cloud Logging settings. (see [below for nested schema](#nestedblock--logging))
- `maintenance_window` (Block List, Max: 1) Maintenance policy of the Greenplum cluster. (see [below for nested schema](#nestedblock--maintenance_window))
- `master_host_group_ids` (Set of String) A list of IDs of the host groups to place master subclusters' VMs of the cluster on.
- `pooler_config` (Block List, Max: 1) Configuration of the connection pooler. (see [below for nested schema](#nestedblock--pooler_config))
- `pxf_config` (Block List, Max: 1) Configuration of the PXF daemon. (see [below for nested schema](#nestedblock--pxf_config))
- `security_group_ids` (Set of String) The list of security groups applied to resource or their components.
- `segment_host_group_ids` (Set of String) A list of IDs of the host groups to place segment subclusters' VMs of the cluster on.
- `service_account_id` (String) ID of service account to use with Yandex Cloud resources (e.g. S3, Cloud Logging).
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `health` (String) Aggregated health of the cluster.
- `id` (String) The ID of this resource.
- `master_hosts` (List of Object) Info about hosts in master subcluster. (see [below for nested schema](#nestedatt--master_hosts))
- `segment_hosts` (List of Object) Info about hosts in segment subcluster. (see [below for nested schema](#nestedatt--segment_hosts))
- `status` (String) Status of the cluster.

<a id="nestedblock--master_subcluster"></a>
### Nested Schema for `master_subcluster`

Required:

- `resources` (Block List, Min: 1, Max: 1) Resources allocated to hosts for master subcluster of the Greenplum cluster. (see [below for nested schema](#nestedblock--master_subcluster--resources))

<a id="nestedblock--master_subcluster--resources"></a>
### Nested Schema for `master_subcluster.resources`

Required:

- `disk_size` (Number) Volume of the storage available to a host, in gigabytes.
- `disk_type_id` (String) Type of the storage of Greenplum hosts - environment default is used if missing.
- `resource_preset_id` (String) The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/ru/docs/managed-greenplum/concepts/instance-types).



<a id="nestedblock--segment_subcluster"></a>
### Nested Schema for `segment_subcluster`

Required:

- `resources` (Block List, Min: 1, Max: 1) Resources allocated to hosts for segment subcluster of the Greenplum cluster. (see [below for nested schema](#nestedblock--segment_subcluster--resources))

<a id="nestedblock--segment_subcluster--resources"></a>
### Nested Schema for `segment_subcluster.resources`

Required:

- `disk_size` (Number) Volume of the storage available to a host, in gigabytes.
- `disk_type_id` (String) Type of the storage of Greenplum hosts - environment default is used if missing.
- `resource_preset_id` (String) The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/ru/docs/managed-greenplum/concepts/instance-types).



<a id="nestedblock--access"></a>
### Nested Schema for `access`

Optional:

- `data_lens` (Boolean) Allow access for [Yandex DataLens](https://yandex.cloud/services/datalens).
- `data_transfer` (Boolean) Allow access for [DataTransfer](https://yandex.cloud/services/data-transfer)
- `web_sql` (Boolean) Allows access for [SQL queries in the management console](https://yandex.cloud/docs/managed-mysql/operations/web-sql-query).
- `yandex_query` (Boolean) Allow access for [Yandex Query](https://yandex.cloud/services/query)


<a id="nestedblock--background_activities"></a>
### Nested Schema for `background_activities`

Optional:

- `analyze_and_vacuum` (Block List) Block to configure 'ANALYZE' and 'VACUUM' daily operations. (see [below for nested schema](#nestedblock--background_activities--analyze_and_vacuum))
- `query_killer_idle` (Block List) Block to configure script that kills long running queries that are in `idle` state. (see [below for nested schema](#nestedblock--background_activities--query_killer_idle))
- `query_killer_idle_in_transaction` (Block List) Block to configure script that kills long running queries that are in `idle in transaction` state. (see [below for nested schema](#nestedblock--background_activities--query_killer_idle_in_transaction))
- `query_killer_long_running` (Block List) Block to configure script that kills long running queries (in any state). (see [below for nested schema](#nestedblock--background_activities--query_killer_long_running))

<a id="nestedblock--background_activities--analyze_and_vacuum"></a>
### Nested Schema for `background_activities.analyze_and_vacuum`

Optional:

- `analyze_timeout` (Number) Maximum duration of the `ANALYZE` operation, in seconds. The default value is `36000`. As soon as this period expires, the `ANALYZE` operation will be forced to terminate.
- `start_time` (String) Time of day in 'HH:MM' format when scripts should run.
- `vacuum_timeout` (Number) Maximum duration of the `VACUUM` operation, in seconds. The default value is `36000`. As soon as this period expires, the `VACUUM` operation will be forced to terminate.


<a id="nestedblock--background_activities--query_killer_idle"></a>
### Nested Schema for `background_activities.query_killer_idle`

Optional:

- `enable` (Boolean) Flag that indicates whether script is enabled.
- `ignore_users` (List of String) List of users to ignore when considering queries to terminate.
- `max_age` (Number) Maximum duration for this type of queries (in seconds).


<a id="nestedblock--background_activities--query_killer_idle_in_transaction"></a>
### Nested Schema for `background_activities.query_killer_idle_in_transaction`

Optional:

- `enable` (Boolean) Flag that indicates whether script is enabled.
- `ignore_users` (List of String) List of users to ignore when considering queries to terminate.
- `max_age` (Number) Maximum duration for this type of queries (in seconds).


<a id="nestedblock--background_activities--query_killer_long_running"></a>
### Nested Schema for `background_activities.query_killer_long_running`

Optional:

- `enable` (Boolean) Flag that indicates whether script is enabled.
- `ignore_users` (List of String) List of users to ignore when considering queries to terminate.
- `max_age` (Number) Maximum duration for this type of queries (in seconds).



<a id="nestedblock--backup_window_start"></a>
### Nested Schema for `backup_window_start`

Optional:

- `hours` (Number) The hour at which backup will be started (UTC).
- `minutes` (Number) The minute at which backup will be started (UTC).


<a id="nestedblock--cloud_storage"></a>
### Nested Schema for `cloud_storage`

Optional:

- `enable` (Boolean) Whether to use cloud storage or not.


<a id="nestedblock--logging"></a>
### Nested Schema for `logging`

Optional:

- `command_center_enabled` (Boolean) Deliver Yandex Command Center's logs to Cloud Logging.
- `enabled` (Boolean) Flag that indicates whether log delivery to Cloud Logging is enabled.
- `folder_id` (String) ID of folder to which deliver logs.
- `greenplum_enabled` (Boolean) Deliver Greenplum's logs to Cloud Logging.
- `log_group_id` (String) Cloud Logging group ID to send logs to.
- `pooler_enabled` (Boolean) Deliver connection pooler's logs to Cloud Logging.


<a id="nestedblock--maintenance_window"></a>
### Nested Schema for `maintenance_window`

Required:

- `type` (String) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.

Optional:

- `day` (String) Day of the week (in `DDD` format). Allowed values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.
- `hour` (Number) Hour of the day in UTC (in `HH` format). Allowed value is between 0 and 23.


<a id="nestedblock--pooler_config"></a>
### Nested Schema for `pooler_config`

Optional:

- `pool_client_idle_timeout` (Number) Value for `pool_client_idle_timeout` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_ttl-integer).
- `pool_size` (Number) Value for `pool_size` [parameter in Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool_size-integer).
- `pooling_mode` (String) Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for Odyssey](https://github.com/yandex/odyssey/blob/master/documentation/configuration.md#pool-string.


<a id="nestedblock--pxf_config"></a>
### Nested Schema for `pxf_config`

Optional:

- `connection_timeout` (Number) The Tomcat server connection timeout for read operations in seconds. Value is between 5 and 600.
- `max_threads` (Number) The maximum number of PXF tomcat threads. Value is between 1 and 1024.
- `pool_allow_core_thread_timeout` (Boolean) Identifies whether or not core streaming threads are allowed to time out.
- `pool_core_size` (Number) The number of core streaming threads. Value is between 1 and 1024.
- `pool_max_size` (Number) The maximum allowed number of core streaming threads. Value is between 1 and 1024.
- `pool_queue_capacity` (Number) The capacity of the core streaming thread pool queue. Value is positive.
- `upload_timeout` (Number) The Tomcat server connection timeout for write operations in seconds. Value is between 5 and 600.
- `xms` (Number) Maximum JVM heap size for PXF daemon. Value is between 64 and 16384.
- `xmx` (Number) Initial JVM heap size for PXF daemon. Value is between 64 and 16384.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--master_hosts"></a>
### Nested Schema for `master_hosts`

Read-Only:

- `assign_public_ip` (Boolean)
- `fqdn` (String)


<a id="nestedatt--segment_hosts"></a>
### Nested Schema for `segment_hosts`

Read-Only:

- `fqdn` (String)

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_mdb_greenplum_cluster.<resource Name> <resource Id>
terraform import yandex_mdb_greenplum_cluster.my_cluster ...
```
