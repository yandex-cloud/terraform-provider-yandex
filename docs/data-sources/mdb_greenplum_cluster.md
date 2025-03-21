---
subcategory: "Managed Service for Greenplum"
page_title: "Yandex: yandex_mdb_greenplum_cluster"
description: |-
  Get information about a Yandex Managed Greenplum cluster.
---

# yandex_mdb_greenplum_cluster (Data Source)

Get information about a Yandex Managed Greenplum cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-greenplum/).

~> Either `cluster_id` or `name` should be specified.

## Example usage

```terraform
//
// Get information about existing MDB Greenplum Cluster.
//
data "yandex_mdb_greenplum_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_greenplum_cluster.foo.network_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `background_activities` (Block List) (see [below for nested schema](#nestedblock--background_activities))
- `cluster_id` (String) The ID of the Greenplum cluster.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `greenplum_config` (Map of String)
- `master_host_group_ids` (Set of String) A list of IDs of the host groups to place master subclusters' VMs of the cluster on.
- `name` (String) The resource name.
- `pooler_config` (Block List, Max: 1) (see [below for nested schema](#nestedblock--pooler_config))
- `pxf_config` (Block List) (see [below for nested schema](#nestedblock--pxf_config))
- `segment_host_group_ids` (Set of String) A list of IDs of the host groups to place segment subclusters' VMs of the cluster on.

### Read-Only

- `access` (List of Object) (see [below for nested schema](#nestedatt--access))
- `assign_public_ip` (Boolean) Sets whether the master hosts should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment.
- `backup_window_start` (List of Object) (see [below for nested schema](#nestedatt--backup_window_start))
- `cloud_storage` (List of Object) (see [below for nested schema](#nestedatt--cloud_storage))
- `created_at` (String) The creation timestamp of the resource.
- `deletion_protection` (Boolean) The `true` value means that resource is protected from accidental deletion.
- `description` (String) The resource description.
- `environment` (String) Deployment environment of the Greenplum cluster. (PRODUCTION, PRESTABLE)
- `health` (String) Aggregated health of the cluster.
- `id` (String) The ID of this resource.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `logging` (List of Object) (see [below for nested schema](#nestedatt--logging))
- `maintenance_window` (List of Object) (see [below for nested schema](#nestedatt--maintenance_window))
- `master_host_count` (Number) Number of hosts in master subcluster (1 or 2).
- `master_hosts` (List of Object) (see [below for nested schema](#nestedatt--master_hosts))
- `master_subcluster` (List of Object) (see [below for nested schema](#nestedatt--master_subcluster))
- `network_id` (String) The `VPC Network ID` of subnets which resource attached to.
- `security_group_ids` (Set of String) The list of security groups applied to resource or their components.
- `segment_host_count` (Number) Number of hosts in segment subcluster (from 1 to 32).
- `segment_hosts` (List of Object) (see [below for nested schema](#nestedatt--segment_hosts))
- `segment_in_host` (Number) Number of segments on segment host (not more then 1 + RAM/8).
- `segment_subcluster` (List of Object) (see [below for nested schema](#nestedatt--segment_subcluster))
- `service_account_id` (String) ID of service account to use with Yandex Cloud resources (e.g. S3, Cloud Logging).
- `status` (String) Status of the cluster.
- `subnet_id` (String) The ID of the subnet, to which the hosts belongs. The subnet must be a part of the network to which the cluster belongs.
- `user_name` (String) Greenplum cluster admin user name.
- `version` (String) Version of the Greenplum cluster. (`6.25`)
- `zone` (String) The [availability zone](https://yandex.cloud/docs/overview/concepts/geo-scope) where resource is located. If it is not provided, the default provider zone will be used.

<a id="nestedblock--background_activities"></a>
### Nested Schema for `background_activities`

Optional:

- `analyze_and_vacuum` (Block List) (see [below for nested schema](#nestedblock--background_activities--analyze_and_vacuum))
- `query_killer_idle` (Block List) (see [below for nested schema](#nestedblock--background_activities--query_killer_idle))
- `query_killer_idle_in_transaction` (Block List) (see [below for nested schema](#nestedblock--background_activities--query_killer_idle_in_transaction))
- `query_killer_long_running` (Block List) (see [below for nested schema](#nestedblock--background_activities--query_killer_long_running))

<a id="nestedblock--background_activities--analyze_and_vacuum"></a>
### Nested Schema for `background_activities.analyze_and_vacuum`

Optional:

- `analyze_timeout` (Number)
- `start_time` (String)
- `vacuum_timeout` (Number)


<a id="nestedblock--background_activities--query_killer_idle"></a>
### Nested Schema for `background_activities.query_killer_idle`

Optional:

- `enable` (Boolean)
- `ignore_users` (List of String)
- `max_age` (Number)


<a id="nestedblock--background_activities--query_killer_idle_in_transaction"></a>
### Nested Schema for `background_activities.query_killer_idle_in_transaction`

Optional:

- `enable` (Boolean)
- `ignore_users` (List of String)
- `max_age` (Number)


<a id="nestedblock--background_activities--query_killer_long_running"></a>
### Nested Schema for `background_activities.query_killer_long_running`

Optional:

- `enable` (Boolean)
- `ignore_users` (List of String)
- `max_age` (Number)



<a id="nestedblock--pooler_config"></a>
### Nested Schema for `pooler_config`

Optional:

- `pool_client_idle_timeout` (Number)
- `pool_size` (Number)
- `pooling_mode` (String)


<a id="nestedblock--pxf_config"></a>
### Nested Schema for `pxf_config`

Optional:

- `connection_timeout` (Number)
- `max_threads` (Number)
- `pool_allow_core_thread_timeout` (Boolean)
- `pool_core_size` (Number)
- `pool_max_size` (Number)
- `pool_queue_capacity` (Number)
- `upload_timeout` (Number)
- `xms` (Number)
- `xmx` (Number)


<a id="nestedatt--access"></a>
### Nested Schema for `access`

Read-Only:

- `data_lens` (Boolean)
- `data_transfer` (Boolean)
- `web_sql` (Boolean)
- `yandex_query` (Boolean)


<a id="nestedatt--backup_window_start"></a>
### Nested Schema for `backup_window_start`

Read-Only:

- `hours` (Number)
- `minutes` (Number)


<a id="nestedatt--cloud_storage"></a>
### Nested Schema for `cloud_storage`

Read-Only:

- `enable` (Boolean)


<a id="nestedatt--logging"></a>
### Nested Schema for `logging`

Read-Only:

- `command_center_enabled` (Boolean)
- `enabled` (Boolean)
- `folder_id` (String)
- `greenplum_enabled` (Boolean)
- `log_group_id` (String)
- `pooler_enabled` (Boolean)


<a id="nestedatt--maintenance_window"></a>
### Nested Schema for `maintenance_window`

Read-Only:

- `day` (String)
- `hour` (Number)
- `type` (String)


<a id="nestedatt--master_hosts"></a>
### Nested Schema for `master_hosts`

Read-Only:

- `assign_public_ip` (Boolean)
- `fqdn` (String)


<a id="nestedatt--master_subcluster"></a>
### Nested Schema for `master_subcluster`

Read-Only:

- `resources` (List of Object) (see [below for nested schema](#nestedobjatt--master_subcluster--resources))

<a id="nestedobjatt--master_subcluster--resources"></a>
### Nested Schema for `master_subcluster.resources`

Read-Only:

- `disk_size` (Number)
- `disk_type_id` (String)
- `resource_preset_id` (String)



<a id="nestedatt--segment_hosts"></a>
### Nested Schema for `segment_hosts`

Read-Only:

- `fqdn` (String)


<a id="nestedatt--segment_subcluster"></a>
### Nested Schema for `segment_subcluster`

Read-Only:

- `resources` (List of Object) (see [below for nested schema](#nestedobjatt--segment_subcluster--resources))

<a id="nestedobjatt--segment_subcluster--resources"></a>
### Nested Schema for `segment_subcluster.resources`

Read-Only:

- `disk_size` (Number)
- `disk_type_id` (String)
- `resource_preset_id` (String)
