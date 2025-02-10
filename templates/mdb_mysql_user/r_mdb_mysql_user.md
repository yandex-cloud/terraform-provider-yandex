---
subcategory: "Managed Service for MySQL"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a MySQL user within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a MySQL user within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mysql/).

## Example usage

{{ tffile "examples/mdb_mysql_user/r_mdb_mysql_user_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user.

* `password` - (Required) The password of the user.

* `permission` - (Optional) Set of permissions granted to the user. The structure is documented below.

* `global_permissions` - (Optional) List user's global permissions 
  Allowed permissions: `REPLICATION_CLIENT`, `REPLICATION_SLAVE`, `PROCESS` for clear list use empty list. If the attribute is not specified there will be no changes.

* `connection_limits` - (Optional) User's connection limits. The structure is documented below. If the attribute is not specified there will be no changes.

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

* `roles` - (Optional) List user's roles in the database. Allowed roles: `ALL`,`ALTER`,`ALTER_ROUTINE`,`CREATE`,`CREATE_ROUTINE`,`CREATE_TEMPORARY_TABLES`, `CREATE_VIEW`,`DELETE`,`DROP`,`EVENT`,`EXECUTE`,`INDEX`,`INSERT`,`LOCK_TABLES`,`SELECT`,`SHOW_VIEW`,`TRIGGER`,`UPDATE`.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_mysql_user/import.sh" }}
