---
layout: "yandex"
page_title: "Yandex: yandex_mdb_clickhouse_cluster"
sidebar_current: "docs-yandex-datasource-mdb-clickhouse-cluster"
description: |-
  Get information about a Yandex Managed ClickHouse cluster.
---

# yandex\_mdb\_clickhouse\_cluster

Get information about a Yandex Managed ClickHouse cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).

## Example Usage

```hcl
data "yandex_mdb_clickhouse_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = "${data.yandex_mdb_clickhouse_cluster.foo.network_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the ClickHouse cluster.

* `name` - (Optional) The name of the ClickHouse cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `network_id` - ID of the network, to which the ClickHouse cluster belongs.
* `created_at` - Creation timestamp of the key.
* `description` - Description of the ClickHouse cluster.
* `labels` - A set of key/value label pairs to assign to the ClickHouse cluster.
* `environment` - Deployment environment of the ClickHouse cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `clickhouse` - Configuration of the ClickHouse subcluster. The structure is documented below.
* `user` - A user of the ClickHouse cluster. The structure is documented below.
* `database` - A database of the ClickHouse cluster. The structure is documented below.
* `host` - A host of the ClickHouse cluster. The structure is documented below.
* `backup_window_start` - Time to start the daily backup, in the UTC timezone. The structure is documented below.
* `access` - Access policy to the ClickHouse cluster. The structure is documented below.
* `zookeeper` - Configuration of the ZooKeeper subcluster. The structure is documented below.

The `clickhouse` block supports:

* `resources` - Resources allocated to hosts of the ClickHouse subcluster. The structure is documented below.

The `zookeeper` block supports:

* `resources` - Resources allocated to hosts of the ZooKeeper subcluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a ClickHouse or ZooKeeper host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).
* `disk_size` - Volume of the storage available to a ClickHouse or ZooKeeper host, in gigabytes.
* `disk_type_id` - Type of the storage of ClickHouse or ZooKeeper hosts.

The `user` block supports:

* `name` - The name of the user.
* `password` - The password of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.

The `database` block supports:

* `name` - The name of the database.

The `host` block supports:

* `fqdn` - The fully qualified domain name of the host.
* `type` - The type of the host to be deployed.
* `zone` - The availability zone where the ClickHouse host will be created.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.
* `shard_name` - The name of the shard to which the host belongs.
* `assign_public_ip` - Sets whether the host should get a public IP address on creation.

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `access` block supports:

* `web_sql` - Allow access for DataLens.
* `data_lens` - Allow access for Web SQL.
* `metrika` - Allow access for Yandex.Metrika.
