---
layout: "yandex"
page_title: "Yandex: yandex_mdb_postgresql_user"
sidebar_current: "docs-yandex-datasource-mdb-postgresql-user"
description: |-
  Get information about a Yandex Managed PostgreSQL user.
---

# yandex\_mdb\_postgresql\_user

Get information about a Yandex Managed PostgreSQL user. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-postgresql/).

## Example Usage

```hcl
data "yandex_mdb_postgresql_user" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "permission" {
  value = "${data.yandex_mdb_postgresql_user.foo.permission}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the PostgreSQL cluster.

* `name` - (Required) The name of the PostgreSQL user.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `password` - The password of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.
* `login` - User's ability to login.
* `grants` - List of the user's grants.
* `conn_limit` - The maximum number of connections per user.
* `settings` - Map of user settings.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.
