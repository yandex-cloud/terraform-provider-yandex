---
subcategory: "Managed Service for MySQL"
page_title: "Yandex: yandex_mdb_mysql_database"
description: |-
  Manages a MySQL database within Yandex Cloud.
---

# yandex_mdb_mysql_database (Resource)

Manages a MySQL database within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mysql/).

## Example usage

```terraform
//
// Create a new MDB MySQL Database.
//
resource "yandex_mdb_mysql_database" "my_db" {
  cluster_id = yandex_mdb_mysql_cluster.my_cluster.id
  name       = "testdb"
}

resource "yandex_mdb_mysql_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  host {
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo.id
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the database.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_mdb_mysql_database.<resource Name> "<cluster Id>:<database Name>"
terraform import yandex_mdb_mysql_database.my_db ...
```
