---
subcategory: "Managed Service for Redis"
page_title: "Yandex: yandex_mdb_redis_cluster"
description: |-
  Manages a Redis cluster within Yandex.Cloud.
---


# yandex_mdb_redis_cluster




Manages a Redis cluster within the Yandex.Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-redis/concepts).

## Example usage

```terraform
resource "yandex_mdb_redis_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    password = "your_password"
    version  = "6.2"
  }

  resources {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.foo.id
  }

  maintenance_window {
    type = "ANYTIME"
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

Example of creating a sharded Redis Cluster.

```terraform
resource "yandex_mdb_redis_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  sharded     = true

  config {
    version  = "6.2"
    password = "your_password"
  }

  resources {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  host {
    zone       = "ru-central1-a"
    subnet_id  = yandex_vpc_subnet.foo.id
    shard_name = "first"
  }

  host {
    zone       = "ru-central1-b"
    subnet_id  = yandex_vpc_subnet.bar.id
    shard_name = "second"
  }

  host {
    zone       = "ru-central1-c"
    subnet_id  = yandex_vpc_subnet.baz.id
    shard_name = "third"
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

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-c"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Redis cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the Redis cluster belongs.

* `environment` - (Required) Deployment environment of the Redis cluster. Can be either `PRESTABLE` or `PRODUCTION`.

* `config` - (Required) Configuration of the Redis cluster. The structure is documented below.

* `resources` - (Required) Resources allocated to hosts of the Redis cluster. The structure is documented below.

* `host` - (Required) A host of the Redis cluster. The structure is documented below.

---

* `description` - (Optional) Description of the Redis cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the Redis cluster.

* `access` - (Optional) Access policy to the Redis cluster. The structure is documented below.

* `sharded` - (Optional) Redis Cluster mode enabled/disabled. Enables sharding when cluster non-sharded. If cluster is sharded - disabling is not allowed.

* `tls_enabled` - (Optional) TLS support mode enabled/disabled.

* `persistence_mode` - (Optional) Persistence mode. Possible values: `ON`, `OFF`.

* `announce_hostnames` - Announce fqdn instead of ip address.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster. Can be either `true` or `false`.

---

The `access` block supports:

* `web_sql` - (Optional) Allow access for Web SQL. Can be either `true` or `false`.

* `data_lens` - (Optional) Allow access for DataLens. Can be either `true` or `false`.

The `config` block supports:

* `password` - (Required) Password for the Redis cluster.

* `timeout` - (Optional) Close the connection after a client is idle for N seconds.

* `maxmemory_policy` - (Optional) Redis key eviction policy for a dataset that reaches maximum memory. Can be any of the listed in [the official RedisDB documentation](https://docs.redislabs.com/latest/rs/administering/database-operations/eviction-policy/).

* `notify_keyspace_events` - (Optional) Select the events that Redis will notify among a set of classes.

* `slowlog_log_slower_than` - (Optional) Log slow queries below this number in microseconds.

* `slowlog_max_len` - (Optional) Slow queries log length.

* `databases` - (Optional) Number of databases (changing requires redis-server restart).

* `maxmemory_percent` - (Optional) Redis maxmemory usage in percent

* `version` - (Required) Version of Redis (6.2).

* `client_output_buffer_limit_normal` - (Optional) Normal clients output buffer limits. See [redis config file](https://github.com/redis/redis/blob/6.2/redis.conf#L1841).

* `client_output_buffer_limit_pubsub` - (Optional) Pubsub clients output buffer limits. See [redis config file](https://github.com/redis/redis/blob/6.2/redis.conf#L1843).

* `lua_time_limit` - (Optional) Maximum time in milliseconds for Lua scripts.

* `repl_backlog_size_percent` - (Optional) Replication backlog size as a percentage of flavor maxmemory.

* `cluster_require_full_coverage` - (Optional) Controls whether all hash slots must be covered by nodes.

* `cluster_allow_reads_when_down` - (Optional) Allows read operations when cluster is down.

* `cluster_allow_pubsubshard_when_down` - (Optional) Permits Pub/Sub shard operations when cluster is down.

* `lfu_decay_time` - (Optional) The time, in minutes, that must elapse in order for the key counter to be divided by two (or decremented if it has a value less <= 10).

* `lfu_log_factor` - (Optional) Determines how the frequency counter represents key hits.

* `turn_before_switchover` - (Optional) Allows to turn before switchover in RDSync.

* `allow_data_loss` - (Optional) Allows some data to be lost in favor of faster switchover/restart by RDSync.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC timezone. The structure is documented below.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started.

* `minutes` - (Optional) The minute at which backup will be started.

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-redis/concepts).

* `disk_size` - (Required) Volume of the storage available to a host, in gigabytes.

* `disk_type_id` - (Optional) Type of the storage of Redis hosts - environment default is used if missing.

The `host` block supports:

* `fqdn` (Computed) - The fully qualified domain name of the host.

* `zone` - (Required) The availability zone where the Redis host will be created. For more information see [the official documentation](https://cloud.yandex.com/docs/overview/concepts/geo-scope).

* `subnet_id` (Optional) - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `shard_name` (Optional) - The name of the shard to which the host belongs.

* `replica_priority` - (Optional) Replica priority of a current replica (usable for non-sharded only).

* `assign_public_ip` - (Optional) Sets whether the host should get a public IP address or not.

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - (Optional) Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - (Optional) Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.

The `disk_size_autoscaling` block supports:

* `disk_size_limit` - Limit of disk size after autoscaling (GiB).
* `planned_usage_threshold` - Maintenance window autoscaling disk usage (percent).
* `emergency_usage_threshold` - Immediate autoscaling disk usage (percent).

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-redis/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-redis/api-ref/Cluster/).

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_redis_cluster.foo cluster_id
```
