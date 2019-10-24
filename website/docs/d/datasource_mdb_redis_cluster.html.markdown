---
layout: "yandex"
page_title: "Yandex: yandex_mdb_redis_cluster"
sidebar_current: "docs-yandex-datasource-mdb-redis-cluster"
description: |-
  Get information about a Yandex Managed Redis cluster.
---

# yandex\_mdb\_redis\_cluster

Get information about a Yandex Managed Redis cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-redis/concepts).

## Example Usage

```hcl
data "yandex_mdb_redis_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = "${data.yandex_mdb_redis_cluster.foo.network_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the Redis cluster.

* `name` - (Optional) The name of the Redis cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `folder_id` - The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.
* `network_id` - ID of the network, to which the Redis cluster belongs.
* `created_at` - Creation timestamp of the key.
* `description` - Description of the Redis cluster.
* `labels` - A set of key/value label pairs to assign to the Redis cluster.
* `environment` - Deployment environment of the Redis cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `config` - Configuration of the Redis cluster. The structure is documented below.
* `resources` - Resources allocated to hosts of the Redis cluster. The structure is documented below.
* `host` - A host of the Redis cluster. The structure is documented below.
* `sharded` - Redis Cluster mode enabled/disabled.

The `config` block supports:

* `timeout` - Close the connection after a client is idle for N seconds.
* `maxmemory_policy` - Redis key eviction policy for a dataset that reaches maximum memory.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-redis/concepts/instance-types).
* `disk_size` - Volume of the storage available to a host, in gigabytes.

The `host` block supports:

* `zone` - The availability zone where the Redis host will be created.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must
  be a part of the network to which the cluster belongs.
* `shard_name` - The name of the shard to which the host belongs.
* `fqdn` - The fully qualified domain name of the host.
