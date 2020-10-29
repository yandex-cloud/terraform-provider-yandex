---
layout: "yandex"
page_title: "Yandex: yandex_mdb_clickhouse_cluster"
sidebar_current: "docs-yandex-mdb-clickhouse-cluster"
description: |-
  Manages a ClickHouse cluster within Yandex.Cloud.
---

# yandex\_mdb\_clickhouse\_cluster

Manages a ClickHouse cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).

## Example Usage

Example of creating a Single Node ClickHouse.

```hcl
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 32
    }
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user"
    password = "your_password"
    permission {
      database_name = "db_name"
    }
  }

  host {
    type      = "CLICKHOUSE"
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

Example of creating a HA ClickHouse Cluster.

```hcl
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name        = "ha"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  zookeeper {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user"
    password = "password"
    permission {
      database_name = "db_name"
    }
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }

  host {
    type      = "CLICKHOUSE"
    zone      = "ru-central1-b"
    subnet_id = "${yandex_vpc_subnet.bar.id}"
  }

  host {
    type      = "ZOOKEEPER"
    zone      = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }

  host {
    type      = "ZOOKEEPER"
    zone      = "ru-central1-b"
    subnet_id = "${yandex_vpc_subnet.bar.id}"
  }

  host {
    type      = "ZOOKEEPER"
    zone      = "ru-central1-c"
    subnet_id = "${yandex_vpc_subnet.baz.id}"
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

Example of creating a sharded ClickHouse Cluster.

```hcl
resource "yandex_mdb_clickhouse_cluster" "foo" {
  name        = "sharded"
  environment = "PRODUCTION"
  network_id  = "${yandex_vpc_network.foo.id}"

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  zookeeper {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user"
    password = "password"
    permission {
      database_name = "db_name"
    }
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-a"
    subnet_id  = "${yandex_vpc_subnet.foo.id}"
    shard_name = "shard1"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-b"
    subnet_id  = "${yandex_vpc_subnet.bar.id}"
    shard_name = "shard1"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-b"
    subnet_id  = "${yandex_vpc_subnet.bar.id}"
    shard_name = "shard2"
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-c"
    subnet_id  = "${yandex_vpc_subnet.baz.id}"
    shard_name = "shard2"
  }

  shard_group {
    name        = "single_shard_group"
    description = "Cluster configuration that contain only shard1"
    shard_names = [
      "shard1",
    ]
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

* `name` - (Required) Name of the ClickHouse cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the ClickHouse cluster belongs.

* `environment` - (Required) Deployment environment of the ClickHouse cluster. Can be either `PRESTABLE` or `PRODUCTION`.

* `clickhouse` - (Required) Configuration of the ClickHouse subcluster. The structure is documented below.

* `user` - (Required) A user of the ClickHouse cluster. The structure is documented below.

* `database` - (Required) A database of the ClickHouse cluster. The structure is documented below.

* `host` - (Required) A host of the ClickHouse cluster. The structure is documented below.

- - -

* `version` - (Optional) Version of the ClickHouse server software.

* `description` - (Optional) Description of the ClickHouse cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the ClickHouse cluster.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC timezone. The structure is documented below.

* `access` - (Optional) Access policy to the ClickHouse cluster. The structure is documented below.

* `zookeeper` - (Optional) Configuration of the ZooKeeper subcluster. The structure is documented below.

- - -

The `clickhouse` block supports:

* `resources` - (Required) Resources allocated to hosts of the ClickHouse subcluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a ClickHouse host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).

* `disk_size` - (Required) Volume of the storage available to a ClickHouse host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of ClickHouse hosts.
  For more information see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts/storage).

The `zookeeper` block supports:

* `resources` - (Optional) Resources allocated to hosts of the ZooKeeper subcluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - (Optional) The ID of the preset for computational resources available to a ZooKeeper host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).

* `disk_size` - (Optional) Volume of the storage available to a ZooKeeper host, in gigabytes.

* `disk_type_id` - (Optional) Type of the storage of ZooKeeper hosts.
  For more information see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts/storage).

The `user` block supports:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

The `database` block supports:

* `name` - (Required) The name of the database.

The `host` block supports:

* `fqdn` - (Computed) The fully qualified domain name of the host.

* `type` - (Required) The type of the host to be deployed. Can be either `CLICKHOUSE` or `ZOOKEEPER`.

* `zone` - (Required) The availability zone where the ClickHouse host will be created.
  For more information see [the official documentation](https://cloud.yandex.com/docs/overview/concepts/geo-scope).
  
* `subnet_id` (Optional) - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `shard_name` (Optional) - The name of the shard to which the host belongs.

* `assign_public_ip` (Optional) - Sets whether the host should get a public IP address on creation. Can be either `true` or `false`.

The `shard_group` block supports:

* `name` (Required) - The name of the shard group, used as cluster name in Distributed tables.

* `description` (Optional) - Description of the shard group.

* `shard_names` (Required) -  List of shards names that belong to the shard group.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started.

* `minutes` - (Optional) The minute at which backup will be started.

The `access` block supports:

* `web_sql` - (Optional) Allow access for DataLens. Can be either `true` or `false`.

* `data_lens` - (Optional) Allow access for Web SQL. Can be either `true` or `false`.

* `metrika` - (Optional) Allow access for Yandex.Metrika. Can be either `true` or `false`.

* `serverless` - (Optional) Allow access for Serverless. Can be either `true` or `false`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Timestamp of cluster creation.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
  For more information see `health` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
  For more information see `status` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/api-ref/Cluster/).

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_clickhouse_cluster.foo cluster_id
```
