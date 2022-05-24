---
layout: "yandex"
page_title: "Yandex: yandex_mdb_mysql_user"
sidebar_current: "docs-yandex-datasource-mdb-mysql-user"
description: |-
  Get information about a Yandex Managed MySQL user.
---

# yandex\_mdb\_mysql\_user

Get information about a Yandex Managed MySQL user. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-mysql/).

## Example Usage

```hcl
data "yandex_mdb_mysql_user" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "permission" {
  value = "${data.yandex_mdb_mysql_user.foo.permission}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the MySQL cluster.

* `name` - (Required) The name of the MySQL user.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `name` - The name of the user.
* `password` - The password of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.
* `global_permissions` - List user's global permissions. Allowed values: `REPLICATION_CLIENT`, `REPLICATION_SLAVE`, `PROCESS` or empty list.
* `connection_limits` - User's connection limits. The structure is documented below.
* `authentication_plugin` - Authentication plugin. Allowed values: `MYSQL_NATIVE_PASSWORD`, `CACHING_SHA2_PASSWORD`, `SHA256_PASSWORD`

The `connection_limits` block supports:   
When these parameters are set to -1, backend default values will be actually used.   

* `max_questions_per_hour` - Max questions per hour.
* `max_updates_per_hour` - Max updates per hour.
* `max_connections_per_hour` - Max connections per hour.
* `max_user_connections` - Max user connections.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.
* `roles` - List user's roles in the database.
            Allowed roles: `ALL`,`ALTER`,`ALTER_ROUTINE`,`CREATE`,`CREATE_ROUTINE`,`CREATE_TEMPORARY_TABLES`,
            `CREATE_VIEW`,`DELETE`,`DROP`,`EVENT`,`EXECUTE`,`INDEX`,`INSERT`,`LOCK_TABLES`,`SELECT`,`SHOW_VIEW`,`TRIGGER`,`UPDATE`.

