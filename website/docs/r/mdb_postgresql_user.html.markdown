---
layout: "yandex"
page_title: "Yandex: yandex_mdb_postgresql_user"
sidebar_current: "docs-yandex-mdb-postgresql-user"
description: |-
  Manages a PostgreSQL user within Yandex.Cloud.
---

# yandex\_mdb\_postgresql\_user

Manages a PostgreSQL user within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-postgresql/).


## Example Usage

```hcl
resource "yandex_mdb_postgresql_user" "foo" {
  cluster_id = yandex_mdb_postgresql_cluster.foo.id
  name       = "alice"
  password   = "password"
  conn_limit = 50
  settings = {
    default_transaction_isolation = "read committed"
    log_min_duration_statement    = 5000
  }
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

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `grants` - (Optional) List of the user's grants.

* `login` - (Optional) User's ability to login.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

* `conn_limit` - (Optional) The maximum number of connections per user. (Default 50)

* `settings` - (Optional) Map of user settings. List of settings is documented below.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

The `settings` block supports:
Full description https://cloud.yandex.com/en-ru/docs/managed-postgresql/api-ref/grpc/user_service#UserSettings1  

* `default_transaction_isolation` - defines the default isolation level to be set for all new SQL transactions. 
* * 0: "unspecified"
* * 1: "read uncommitted"
* * 2: "read committed"
* * 3: "repeatable read"
* * 4: "serializable"

* `lock_timeout` - The maximum time (in milliseconds) for any statement to wait for acquiring a lock on an table, index, row or other database object (default 0)

* `log_min_duration_statement` - This setting controls logging of the duration of statements. (default -1 disables logging of the duration of statements.)

* `synchronous_commit` - This setting defines whether DBMS will commit transaction in a synchronous way.
* * 0: "unspecified"
* * 1: "on"
* * 2: "off"
* * 3: "local"
* * 4: "remote write"
* * 5: "remote apply"

* `temp_file_limit` - The maximum storage space size (in kilobytes) that a single process can use to create temporary files.

* `log_statement` - This setting specifies which SQL statements should be logged (on the user level).
* * 0: "unspecified"
* * 1: "none"
* * 2: "ddl"
* * 3: "mod"
* * 4: "all"

## Import

A PostgreSQL user can be imported using the following format:

```
$ terraform import yandex_mdb_postgresql_user.foo {{cluster_id}}:{{username}}
```
