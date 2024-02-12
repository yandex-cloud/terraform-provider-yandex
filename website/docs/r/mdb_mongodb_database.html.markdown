---
layout: "yandex"
page_title: "Yandex: yandex_mdb_mongodb_database"
sidebar_current: "docs-yandex-mdb-mongodb-database"
description: |-
  Manages a MongoDB database within Yandex.Cloud.
---

# yandex\_mdb\_mongodb\_database

Manages a MongoDB database within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-mongodb/).


## Example Usage

```hcl
resource "yandex_mdb_mongodb_database" "foo" {
  cluster_id = yandex_mdb_mongodb_cluster.foo.id
  name       = "testdb"
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

* `name` - (Required) The name of the database.

## Import

A MongoDB database can be imported using the following format:

```
$ terraform import yandex_mdb_mongodb_database.foo {{cluster_id}}:{{database_name}}
```
