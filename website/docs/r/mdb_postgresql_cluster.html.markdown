---
layout: "yandex"
page_title: "Yandex: yandex_mdb_postgresql_cluster"
sidebar_current: "docs-yandex-mdb-postgresql-cluster"
description: |-
  Manages a PostgreSQL cluster within Yandex.Cloud.
---

# yandex\_mdb\_postgresql\_cluster

Manages a PostgreSQL cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-postgresql/).

## Example Usage

Example of creating a Single Node PostgreSQL.

```hcl
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 12
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  database {
    name  = "db_name"
    owner = "user_name"
  }

  user {
    name     = "user_name"
    password = "your_password"
    conn_limit = 50
    permission {
      database_name = "db_name"
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.foo.id
  }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

Example of creating a High-Availability (HA) PostgreSQL Cluster.

```hcl
resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "ha"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 12
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  database {
    name  = "db_name"
    owner = "user_name"
  }

  user {
    name     = "user_name"
    password = "password"
    permission {
      database_name = "db_name"
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.foo.id
  }

  host {
    zone      = "ru-central1-b"
    subnet_id = yandex_vpc_subnet.bar.id
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
```

## Argument Reference

The following arguments are supported:

* `config` - (Required) Configuration of the PostgreSQL cluster. The structure is documented below.

* `database` - (Required) A database of the PostgreSQL cluster. The structure is documented below.

* `environment` - (Required) Deployment environment of the PostgreSQL cluster.

* `host` - (Required) A host of the PostgreSQL cluster. The structure is documented below.

* `name` - (Required) Name of the PostgreSQL cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the PostgreSQL cluster belongs.

* `user` - (Required) A user of the PostgreSQL cluster. The structure is documented below.

- - -

* `description` - (Optional) Description of the PostgreSQL cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the PostgreSQL cluster.

- - -

The `config` block supports:

* `resources` - (Required) Resources allocated to hosts of the PostgreSQL cluster. The structure is documented below.

* `version` - (Required) Version of the PostgreSQL cluster.

* `access` - (Optional) Access policy to the PostgreSQL cluster. The structure is documented below.

* `autofailover` - (Optional) Configuration setting which enables/disables autofailover in cluster.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC timezone. The structure is documented below.

* `pooler_config` - (Optional) Configuration of the connection pooler. The structure is documented below.

The `resources` block supports:

* `disk_size` - (Required) Volume of the storage available to a PostgreSQL host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of PostgreSQL hosts.

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a PostgreSQL host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/concepts/instance-types).

The `pooler_config` block supports:

* `pool_discard` - (Optional) Setting `server_reset_query_always` [parameter in PgBouncer](https://www.pgbouncer.org/config.html).

* `pooling_mode` - (Optional) Mode that the connection pooler is working in. See descriptions of all modes in the [documentation for PgBouncer](https://pgbouncer.github.io/usage).

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started (UTC).

* `minutes` - (Optional) The minute at which backup will be started (UTC).

The `access` block supports:

* `data_lens` - (Optional) Allow access for [Yandex DataLens](https://cloud.yandex.com/services/datalens).

The `user` block supports:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `grants` - (Optional) List of the user's grants.

* `login` - (Optional) User's ability to login.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

* `conn_limit` - (Optional) The maximum number of connections per user. (Default 50)

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

The `database` block supports:

* `name` - (Required) The name of the database.

* `owner` - (Required) Name of the user assigned as the owner of the database.

* `extension` - (Optional) Set of database extensions. The structure is documented below

* `lc_collate` - (Optional) POSIX locale for string sorting order. Forbidden to change in an existing database.

* `lc_type` - (Optional) POSIX locale for character classification. Forbidden to change in an existing database.

The `extension` block supports:

* `name` - (Required) Name of the database extension. For more information on available extensions see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/operations/cluster-extensions).

* `version` - (Optional) Version of the extension.

The `host` block supports:

* `zone` - (Required) The availability zone where the PostgreSQL host will be created.

* `assign_public_ip` - (Optional) Sets whether the host should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment

* `subnet_id` - (Optional) The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `fqdn` - (Computed) The fully qualified domain name of the host.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Timestamp of cluster creation.

* `health` - Aggregated health of the cluster.

* `status` - Status of the cluster.

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_postgresql_cluster.foo cluster_id
```
