---
layout: "yandex"
page_title: "Yandex: yandex_mdb_mongodb_cluster"
sidebar_current: "docs-yandex-mdb-mongodb-cluster"
description: |-
  Manages a MongoDB cluster within Yandex.Cloud.
---

# yandex\_mdb\_mongodb\_cluster

Manages a MongoDB cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-mongodb/concepts).

## Example Usage

Example of creating a Single Node MongoDB.

```hcl
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_mdb_mongodb_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.foo.id}"

  cluster_config {
    version = "4.2"
  }

  labels = {
    test_key = "test_value"
  }

  database {
    name = "testdb"
  }

  user {
    name     = "john"
    password = "password"
    permission {
      database_name = "testdb"
    }
  }

  resources {
    resource_preset_id = "b1.nano"
    disk_size          = 16
    disk_type_id       = "network-hdd"
  }

  host {
    zone_id   = "ru-central1-a"
    subnet_id = "${yandex_vpc_subnet.foo.id}"
  }
  
  maintenance_window {
    type = "ANYTIME"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the MongoDB cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the MongoDB cluster belongs.

* `environment` - (Required) Deployment environment of the MongoDB cluster. Can be either `PRESTABLE` or `PRODUCTION`.

* `cluster_config` - (Required) Configuration of the MongoDB subcluster. The structure is documented below.

* `user` - (Required) A user of the MongoDB cluster. The structure is documented below.

* `database` - (Required) A database of the MongoDB cluster. The structure is documented below.

* `host` - (Required) A host of the MongoDB cluster. The structure is documented below.

* `resources` - (Required) Resources allocated to hosts of the MongoDB cluster. The structure is documented below.

- - -

* `version` - (Optional) Version of the MongoDB server software. Can be either `4.0`, `4.2`, `4.4`, `4.4-enterprise`, `5.0` and `5.0-enterprise`.

* `description` - (Optional) Description of the MongoDB cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the MongoDB cluster.

* `access` - (Optional) Access policy to the MongoDB cluster. The structure is documented below.

* `security_group_ids` - (Optional) A set of ids of security groups assigned to hosts of the cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster.  Can be either `true` or `false`.
- - -

The `cluster_config` block supports:

* `version` - (Required) Version of MongoDB (either 5.0, 4.4, 4.2 or 4.0).

* `feature_compatibility_version` - (Optional) Feature compatibility version of MongoDB. If not provided version is taken. Can be either `5.0`, `4.4`, `4.2` and `4.0`.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC timezone. The structure is documented below.

* `access` - (Optional) Shows whether cluster has access to data lens. The structure is documented below.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started.

* `minutes` - (Optional) The minute at which backup will be started.

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a MongoDB host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/concepts).

* `disk_size` - (Required) Volume of the storage available to a MongoDB host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of MongoDB hosts.
  For more information see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts/storage).

The `user` block supports:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

* `roles` - (Optional) The roles of the user in this database. For more information see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/concepts/users-and-roles).

The `database` block supports:

* `name` - (Required) The name of the database.

The `host` block supports:

* `name` - (Computed) The fully qualified domain name of the host. Computed on server side.

* `zone_id` - (Required) The availability zone where the MongoDB host will be created.
  For more information see [the official documentation](https://cloud.yandex.com/docs/overview/concepts/geo-scope).

* `role` - (Optional) The role of the cluster (either PRIMARY or SECONDARY).

* `health` - (Computed) The health of the host.

* `subnet_id` - (Required) The ID of the subnet, to which the host belongs. The subnet must
  be a part of the network to which the cluster belongs.
  
* `assign_public_ip` -(Optional)  Should this host have assigned public IP assigned. Can be either `true` or `false`.

* `shard_name` - (Optional) The name of the shard to which the host belongs.

* `type` - (Optional) type of mongo daemon which runs on this host (mongod, mongos or monogcfg). Defaults to mongod.

The `access` block supports:

* `data_lens` - (Optional) Allow access for DataLens.

The `maintenance_window` block supports:

* `type` - (Required) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.
* `hour` - (Optional) Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.
* `day` - (Optional) Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
  For more information see `health` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
  For more information see `status` field of JSON representation in [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/api-ref/Cluster/).

* `cluster_id` - The ID of the cluster.

* `sharded` - MongoDB Cluster mode enabled/disabled.

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_mongodb_cluster.foo cluster_id
```
