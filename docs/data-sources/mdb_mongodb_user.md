---
subcategory: "Managed Service for MongoDB"
page_title: "Yandex: yandex_mdb_mongodb_user"
description: |-
  Get information about a Yandex Managed MongoDB user.
---


# yandex_mdb_mongodb_user




Get information about a Yandex Managed MongoDB user. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/).

## Example usage

```terraform
data "yandex_mdb_mongodb_user" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "permission" {
  value = data.yandex_mdb_mongodb_user.foo.permission
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the MongoDB cluster.

* `name` - (Required) The name of the MongoDB user.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `password` - The password of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.
* `roles` - List of strings. The roles of the user in this database. For more information see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/concepts/users-and-roles).
