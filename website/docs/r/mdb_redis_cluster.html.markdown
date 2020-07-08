---
layout: "yandex"
page_title: "Yandex: yandex_mdb_redis_cluster"
sidebar_current: "docs-yandex-mdb-redis-cluster"
description: |-
  Manages a Redis cluster within Yandex.Cloud.
---

# yandex\_mdb\_redis\_cluster

Manages a Redis cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-redis/concepts).

## Example Usage

Example of creating a Standalone Redis.

```hcl
resource "yandex_mdb_redis_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  config {
    password = "your_password"
  }

  resources {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

Example of creating a sharded Redis Cluster.

```hcl
resource "yandex_mdb_redis_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"
  sharded     = true

  config {
    password = "your_password"
  }

  resources {
    resource_preset_id = "hm1.nano"
    disk_size          = 16
  }

  host {
    zone       = "ru-central1-a"
    subnet_id  = "${yandex_vpc_subnet.foo.id}"
    shard_name = "first"
  }

  host {
    zone       = "ru-central1-b"
    subnet_id  = "${yandex_vpc_subnet.bar.id}"
    shard_name = "second"
  }

  host {
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.baz.id}"
    shard_name = "third"
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

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.foo.id}"
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

- - -

* `description` - (Optional) Description of the Redis cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the Redis cluster.

* `sharded` - (Optional) Redis Cluster mode enabled/disabled.

- - -

The `config` block supports:

* `password` - (Required) Password for the Redis cluster.

* `timeout` - (Optional) Close the connection after a client is idle for N seconds.

* `maxmemory_policy` - (Optional) Redis key eviction policy for a dataset that reaches maximum memory.
  Can be any of the listed in [the official RedisDB documentation](https://docs.redislabs.com/latest/rs/administering/database-operations/eviction-policy/).

* `version` - (Required) Version of Redis (either 5.0 or 6.0).

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-redis/concepts).

* `disk_size` - (Required) Volume of the storage available to a host, in gigabytes.

The `host` block supports:

* `fqdn` (Computed) - The fully qualified domain name of the host.

* `zone` - (Required) The availability zone where the Redis host will be created.
  For more information see [the official documentation](https://cloud.yandex.com/docs/overview/concepts/geo-scope).
  
* `subnet_id` (Optional) - The ID of the subnet, to which the host belongs. The subnet must
  be a part of the network to which the cluster belongs.

* `shard_name` (Optional) - The name of the shard to which the host belongs.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
  For more information see `health` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-redis/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
  For more information see `status` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-redis/api-ref/Cluster/).

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_redis_cluster.foo cluster_id
```
