---
subcategory: "Managed Service for MySQL"
page_title: "Yandex: yandex_mdb_mysql_database"
description: |-
  Manages a MySQL database within Yandex.Cloud.
---


# yandex_mdb_mysql_database




Manages a MySQL database within the Yandex.Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mysql/).

```terraform
resource "yandex_mdb_mysql_user" "john" {
  cluster_id = yandex_mdb_mysql_cluster.foo.id
  name       = "john"
  password   = "password"

  permission {
    database_name = yandex_mdb_mysql_database.testdb.name
    roles         = ["ALL"]
  }

  permission {
    database_name = yandex_mdb_mysql_database.new_testdb.name
    roles         = ["ALL", "INSERT"]
  }

  connection_limits {
    max_questions_per_hour   = 10
    max_updates_per_hour     = 20
    max_connections_per_hour = 30
    max_user_connections     = 40
  }

  global_permissions = ["PROCESS"]

  authentication_plugin = "SHA256_PASSWORD"
}

resource "yandex_mdb_mysql_cluster" "foo" {
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

## Import

A MySQL database can be imported using the following format:

```
$ terraform import yandex_mdb_mysql_database.foo {cluster_id}:{database_name}
```
