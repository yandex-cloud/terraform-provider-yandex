---
layout: "yandex"
page_title: "Yandex: yandex_mdb_mysql_user"
sidebar_current: "docs-yandex-mdb-mysql-user"
description: |-
  Manages a MySQL user within Yandex.Cloud.
---

# yandex\_mdb\_mysql\_user

Manages a MySQL user within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-mysql/).


## Example Usage

```hcl
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

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

* `global_permissions` - (Optional) List user's global permissions     
            Allowed permissions:  `REPLICATION_CLIENT`, `REPLICATION_SLAVE`, `PROCESS` for clear list use empty list.
            If the attribute is not specified there will be no changes.

* `connection_limits` - (Optional) User's connection limits. The structure is documented below.
            If the attribute is not specified there will be no changes.

* `authentication_plugin` - (Optional) Authentication plugin. Allowed values: `MYSQL_NATIVE_PASSWORD`, `CACHING_SHA2_PASSWORD`, `SHA256_PASSWORD` (for version 5.7 `MYSQL_NATIVE_PASSWORD`, `SHA256_PASSWORD`)

The `connection_limits` block supports:  
default value is -1,   
When these parameters are set to -1, backend default values will be actually used.   

* `max_questions_per_hour` - Max questions per hour.

* `max_updates_per_hour` - Max updates per hour.

* `max_connections_per_hour` - Max connections per hour.

* `max_user_connections` - Max user connections.

The `permission` block supports:

* `database_name` - (Required) The name of the database that the permission grants access to.

* `roles` - (Optional) List user's roles in the database.
            Allowed roles: `ALL`,`ALTER`,`ALTER_ROUTINE`,`CREATE`,`CREATE_ROUTINE`,`CREATE_TEMPORARY_TABLES`,
            `CREATE_VIEW`,`DELETE`,`DROP`,`EVENT`,`EXECUTE`,`INDEX`,`INSERT`,`LOCK_TABLES`,`SELECT`,`SHOW_VIEW`,`TRIGGER`,`UPDATE`.

## Import

A MySQL user can be imported using the following format:

```
$ terraform import yandex_mdb_mysql_user.foo {{cluster_id}}:{{username}}
```
