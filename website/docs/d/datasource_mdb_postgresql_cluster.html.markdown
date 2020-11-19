---
layout: "yandex"
page_title: "Yandex: yandex_mdb_postgresql_cluster"
sidebar_current: "docs-yandex-datasource-mdb-postgresql-cluster"
description: |-
  Get information about a Yandex Managed PostgreSQL cluster.
---

# yandex\_mdb\_postgresql\_cluster

Get information about a Yandex Managed PostgreSQL cluster. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-postgresql/).

## Example Usage

```hcl
data "yandex_mdb_postgresql_cluster" "foo" {
  name = "test"
}

output "fqdn" {
  value = "${data.yandex_mdb_postgresql_cluster.foo.host.0.fqdn}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the PostgreSQL cluster.

* `name` - (Optional) The name of the PostgreSQL cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `network_id` - ID of the network, to which the PostgreSQL cluster belongs.
* `created_at` - Timestamp of cluster creation.
* `description` - Description of the PostgreSQL cluster.
* `labels` - A set of key/value label pairs to assign to the PostgreSQL cluster.
* `environment` - Deployment environment of the PostgreSQL cluster.
* `health` - Aggregated health of the cluster.
* `status` - Status of the cluster.
* `config` - Configuration of the PostgreSQL cluster. The structure is documented below.
* `user` - A user of the PostgreSQL cluster. The structure is documented below.
* `database` - A database of the PostgreSQL cluster. The structure is documented below.
* `host` - A host of the PostgreSQL cluster. The structure is documented below.

The `config` block supports:

* `version` - Version of the PostgreSQL cluster.
* `autofailover` - Configuration setting which enables/disables autofailover in cluster.
* `resources` - Resources allocated to hosts of the PostgreSQL cluster. The structure is documented below.
* `pooler_config` - Configuration of the connection pooler. The structure is documented below.
* `backup_window_start` - Time to start the daily backup, in the UTC timezone. The structure is documented below.
* `access` - Access policy to the PostgreSQL cluster. The structure is documented below.

The `resources` block supports:

* `resources_preset_id` - The ID of the preset for computational resources available to a PostgreSQL host (CPU, memory etc.).
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/concepts/instance-types).
* `disk_size` - Volume of the storage available to a PostgreSQL host, in gigabytes.
* `disk_type_id` - Type of the storage for PostgreSQL hosts.

The `pooler_config` block supports:

* `pooling_mode` - Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for PgBouncer](https://pgbouncer.github.io/usage).
* `pool_discard` - Setting `server_reset_query_always` parameter in PgBouncer.

The `backup_window_start` block supports:

* `hours` - The hour at which backup will be started.
* `minutes` - The minute at which backup will be started.

The `access` block supports:

* `data_lens` - Allow access for [Yandex DataLens](https://cloud.yandex.com/services/datalens).

The `user` block supports:

* `name` - The name of the user.
* `password` - The password of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.
* `login` - User's ability to login.
* `grants` - List of the user's grants.
* `conn_limit` - The maximum number of connections per user.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.

The `database` block supports:

* `name` - The name of the database.
* `owner` - Name of the user assigned as the owner of the database.
* `lc_collate` - POSIX locale for string sorting order. Forbidden to change in an existing database.
* `lc_type` - POSIX locale for character classification. Forbidden to change in an existing database.
* `extension` - Set of database extensions. The structure is documented below

The `extension` block supports:

* `name` - Name of the database extension. For more information on available extensions see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/operations/cluster-extensions).
* `version` - Version of the extension.

The `host` block supports:

* `fqdn` - The fully qualified domain name of the host.
* `zone` - The availability zone where the PostgreSQL host will be created.
* `subnet_id` - The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.
* `assign_public_ip` - Sets whether the host should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment
* `role` - Role of the host in the cluster.