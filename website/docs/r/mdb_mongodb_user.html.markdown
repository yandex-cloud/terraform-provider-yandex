---
layout: "yandex"
page_title: "Yandex: yandex_mdb_mongodb_user"
sidebar_current: "docs-yandex-mdb-mongodb-user"
description: |-
  Manages a MongoDB user within Yandex.Cloud.
---

# yandex\_mdb\_mongodb\_user

Manages a MongoDB user within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-mongodb/).


## Example Usage

```hcl
resource "yandex_mdb_mongodb_user" "foo" {
  cluster_id = yandex_mdb_mongodb_cluster.foo.id
  name       = "alice"
  password   = "password"
}

resource "yandex_mdb_mongodb_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  cluster_config {
    version = "6.0"
  }

  host {
    zone_id      = "ru-central1-a"
    subnet_id    = yandex_vpc_subnet.foo.id
  }
  resources_mongod {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
   }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.
* `roles` - (Optional) List of strings. The roles of the user in this database. For more information see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/concepts/users-and-roles).

## Import

A MongoDB user can be imported using the following format:

```
$ terraform import yandex_mdb_mongodb_user.foo {{cluster_id}}:{{username}}
```
