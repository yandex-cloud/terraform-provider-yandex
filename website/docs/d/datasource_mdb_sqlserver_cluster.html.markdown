---
layout: "yandex"
page_title: "Yandex: yandex_mdb_sqlserver_cluster"
sidebar_current: "docs-yandex-datasource-mdb-sqlserver-cluster"
description: |-
  Get information about a Yandex Managed SQLServer cluster.
---

# yandex\_mdb\_sqlserver\_cluster

Get information about a Yandex Managed SQLServer cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-sqlserver/).

## Example Usage

```hcl
data "yandex_mdb_sqlserver_cluster" "foo" {
  name = "test"
}

output "network_id" {
  value = "${data.yandex_mdb_sqlserver_cluster.foo.network_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the SQLServer cluster.

* `name` - (Optional) The name of the SQLServer cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `network_id` - ID of the network, to which the SQLServer cluster belongs.
* `created_at` - Creation timestamp of the key.
* `description` - Description of the SQLServer cluster.
* `labels` - A set of key/value label pairs to assign to the SQLServer cluster.
* `environment` - Deployment environment of the SQLServer cluster.
* `version` - Version of the SQLServer cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `resources` - Resources allocated to hosts of the SQLServer cluster. The structure is documented below.
* `user` - A user of the SQLServer cluster. The structure is documented below.
* `database` - A database of the SQLServer cluster. The structure is documented below.
* `host` - A host of the SQLServer cluster. The structure is documented below.
* `sqlserver_config` - SQLServer cluster config.
* `security_group_ids` - A set of ids of security groups assigned to hosts of the cluster.
* `host_group_ids` - A list of IDs of the host groups hosting VMs of the cluster.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a SQLServer host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-sqlserver/concepts/instance-types).
* `disk_size` - Volume of the storage available to a SQLServer host, in gigabytes.
* `disk_type_id` - Type of the storage for SQLServer hosts.

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `user` block supports:

* `name` - The name of the user.
* `password` - The password of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.
* `roles` - List user's roles in the database.
            Allowed roles: `OWNER`, `SECURITYADMIN`, `ACCESSADMIN`, `BACKUPOPERATOR`, `DDLADMIN`, `DATAWRITER`, `DATAREADER`, `DENYDATAWRITER`, `DENYDATAREADER`.

The `database` block supports:

* `name` - The name of the database.

The `host` block supports:

* `fqdn` - The fully qualified domain name of the host.
* `zone` - The availability zone where the SQLServer host will be created.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.
* `assign_public_ip` - Sets whether the host should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment

