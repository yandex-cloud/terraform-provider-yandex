---
layout: "yandex"
page_title: "Yandex: yandex_mdb_mysql_cluster"
sidebar_current: "docs-yandex-mdb-mysql-cluster"
description: |-
  Manages a MySQL cluster within Yandex.Cloud.
---

# yandex\_mdb\_mysql\_cluster

Manages a MySQL cluster within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-mysql/).

## Example Usage

Example of creating a Single Node MySQL.

```hcl
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user_name"
    password = "your_password"
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

Example of creating a High-Availability(HA) MySQL Cluster.

```hcl
resource "yandex_mdb_mysql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  database {
    name = "db_name"
  }

  user {
    name     = "user_name"
    password = "your_password"
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
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}
```




## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the MySQL cluster. Provided by the client when the cluster is created.

* `network_id` - (Required) ID of the network, to which the MySQL cluster uses.

* `environment` - (Required) Deployment environment of the MySQL cluster.

* `version` - (Required) Version of the MySQL cluster.

* `resources` - (Required) Resources allocated to hosts of the MySQL cluster. The structure is documented below.

* `user` - (Required) A user of the MySQL cluster. The structure is documented below.

* `database` - (Required) A database of the MySQL cluster. The structure is documented below.

* `host` - (Required) A host of the MySQL cluster. The structure is documented below.

- - -

* `description` - (Optional) Description of the MySQL cluster.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the MySQL cluster.

* `backup_window_start` - (Optional) Time to start the daily backup, in the UTC. The structure is documented below.

- - -

The `resources` block supports:

* `resources_preset_id` - (Required) The ID of the preset for computational resources available to a MySQL host (CPU, memory etc.). 
  For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mysql/concepts/instance-types).

* `disk_size` - (Required) Volume of the storage available to a MySQL host, in gigabytes.

* `disk_type_id` - (Required) Type of the storage of MySQL hosts.

The `backup_window_start` block supports:

* `hours` - (Optional) The hour at which backup will be started.

* `minutes` - (Optional) The minute at which backup will be started.

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

* `zone` - (Required) The availability zone where the MySQL host will be created.

* `subnet_id` - (Optional) The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.

* `assign_public_ip` - (Optional) Sets whether the host should get a public IP address on creation.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the cluster.

* `health` - Aggregated health of the cluster.

* `status` - Status of the cluster.

## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_mdb_mysql_cluster.foo cluster_id
```