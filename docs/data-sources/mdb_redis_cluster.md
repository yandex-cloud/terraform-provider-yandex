---
subcategory: "Managed Service for Redis"
page_title: "Yandex: yandex_mdb_redis_cluster"
description: |-
  Get information about a Yandex Managed Redis cluster.
---

# yandex_mdb_redis_cluster (Data Source)

Get information about a Yandex Managed Redis cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-redis/concepts).

~> Either `cluster_id` or `name` should be specified.

## Example usage

```terraform
//
// Get information about existing MDB Redis Cluster.
//
data "yandex_mdb_redis_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = data.yandex_mdb_redis_cluster.foo.network_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `cluster_id` (String) The ID of the Redis cluster.
- `deletion_protection` (Boolean) The `true` value means that resource is protected from accidental deletion.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `name` (String) The name of the Redis cluster.

### Read-Only

- `announce_hostnames` (Boolean) Announce fqdn instead of ip address.
- `config` (List of Object) (see [below for nested schema](#nestedatt--config))
- `created_at` (String) The creation timestamp of the resource.
- `description` (String) The resource description.
- `disk_size_autoscaling` (List of Object) (see [below for nested schema](#nestedatt--disk_size_autoscaling))
- `environment` (String) Deployment environment of the Redis cluster. Can be either `PRESTABLE` or `PRODUCTION`.
- `health` (String) Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-redis/api-ref/Cluster/).
- `host` (List of Object) (see [below for nested schema](#nestedatt--host))
- `id` (String) The ID of this resource.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `maintenance_window` (List of Object) (see [below for nested schema](#nestedatt--maintenance_window))
- `network_id` (String) The `VPC Network ID` of subnets which resource attached to.
- `persistence_mode` (String) Persistence mode. Possible values: `ON`, `OFF`.
- `resources` (List of Object) (see [below for nested schema](#nestedatt--resources))
- `security_group_ids` (Set of String) The list of security groups applied to resource or their components.
- `sharded` (Boolean) Redis Cluster mode enabled/disabled. Enables sharding when cluster non-sharded. If cluster is sharded - disabling is not allowed.
- `status` (String) Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-redis/api-ref/Cluster/).
- `tls_enabled` (Boolean) TLS support mode enabled/disabled.

<a id="nestedatt--config"></a>
### Nested Schema for `config`

Read-Only:

- `allow_data_loss` (Boolean)
- `backup_window_start` (List of Object) (see [below for nested schema](#nestedobjatt--config--backup_window_start))
- `client_output_buffer_limit_normal` (String)
- `client_output_buffer_limit_pubsub` (String)
- `cluster_allow_pubsubshard_when_down` (Boolean)
- `cluster_allow_reads_when_down` (Boolean)
- `cluster_require_full_coverage` (Boolean)
- `databases` (Number)
- `io_threads_allowed` (Boolean)
- `lfu_decay_time` (Number)
- `lfu_log_factor` (Number)
- `lua_time_limit` (Number)
- `maxmemory_percent` (Number)
- `maxmemory_policy` (String)
- `notify_keyspace_events` (String)
- `repl_backlog_size_percent` (Number)
- `slowlog_log_slower_than` (Number)
- `slowlog_max_len` (Number)
- `timeout` (Number)
- `turn_before_switchover` (Boolean)
- `use_luajit` (Boolean)
- `version` (String)

<a id="nestedobjatt--config--backup_window_start"></a>
### Nested Schema for `config.backup_window_start`

Read-Only:

- `hours` (Number)
- `minutes` (Number)



<a id="nestedatt--disk_size_autoscaling"></a>
### Nested Schema for `disk_size_autoscaling`

Read-Only:

- `disk_size_limit` (Number)
- `emergency_usage_threshold` (Number)
- `planned_usage_threshold` (Number)


<a id="nestedatt--host"></a>
### Nested Schema for `host`

Read-Only:

- `assign_public_ip` (Boolean)
- `fqdn` (String)
- `replica_priority` (Number)
- `shard_name` (String)
- `subnet_id` (String)
- `zone` (String)


<a id="nestedatt--maintenance_window"></a>
### Nested Schema for `maintenance_window`

Read-Only:

- `day` (String)
- `hour` (Number)
- `type` (String)


<a id="nestedatt--resources"></a>
### Nested Schema for `resources`

Read-Only:

- `disk_size` (Number)
- `disk_type_id` (String)
- `resource_preset_id` (String)
