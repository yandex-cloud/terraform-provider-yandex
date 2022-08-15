---
layout: "yandex"
page_title: "Yandex: yandex_mdb_postgresql_database"
sidebar_current: "docs-yandex-mdb-postgresql-database"
description: |-
  Manages a PostgreSQL database within Yandex.Cloud.
---

# yandex\_mdb\_postgresql\_database

Manages a PostgreSQL database within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-postgresql/).


## Example Usage

```hcl
resource "yandex_mdb_postgresql_database" "foo" {
  cluster_id = yandex_mdb_postgresql_cluster.foo.id
  name       = "testdb"
  owner      = yandex_mdb_postgresql_user.alice.name
  lc_collate = "en_US.UTF-8"
  lc_type    = "en_US.UTF-8"
  extension {
    name = "uuid-ossp"
  }
  extension {
    name = "xml2"
  }
}

resource "yandex_mdb_postgresql_user" "foo" {
  cluster_id = yandex_mdb_postgresql_cluster.foo.id
  name       = "alice"
  password   = "password"
}

resource "yandex_mdb_postgresql_cluster" "foo" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 14
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
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

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the database.

* `owner` - (Required) Name of the user assigned as the owner of the database. Forbidden to change in an existing database.

* `extension` - (Optional) Set of database extensions. The structure is documented below

* `lc_collate` - (Optional) POSIX locale for string sorting order. Forbidden to change in an existing database.

* `lc_type` - (Optional) POSIX locale for character classification. Forbidden to change in an existing database.

* `template_db` - (Optional) Name of the template database.

The `extension` block supports:

* `name` - (Required) Name of the database extension. For more information on available extensions see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/operations/cluster-extensions).

* `version` - (Optional) Version of the extension.


## Import

A PostgreSQL database can be imported using the following format:

```
$ terraform import yandex_mdb_postgresql_database.foo {{cluster_id}}:{{database_name}}
```
