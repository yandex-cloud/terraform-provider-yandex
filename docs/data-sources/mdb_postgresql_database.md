---
subcategory: "Managed Service for PostgreSQL"
page_title: "Yandex: yandex_mdb_postgresql_database"
description: |-
  Get information about a Yandex Managed PostgreSQL database.
---

# yandex_mdb_postgresql_database (Data Source)

Get information about a Yandex Managed PostgreSQL database. For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/).

## Example usage

```terraform
//
// Get information about existing MDB PostgreSQL Database.
//
data "yandex_mdb_postgresql_database" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "owner" {
  value = data.yandex_mdb_postgresql_database.foo.owner
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the PostgreSQL cluster.

* `name` - (Required) The name of the PostgreSQL cluster.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `owner` - Name of the user assigned as the owner of the database.
* `lc_collate` - POSIX locale for string sorting order. Forbidden to change in an existing database.
* `lc_type` - POSIX locale for character classification. Forbidden to change in an existing database.
* `extension` - Set of database extensions. The structure is documented below
* `template_db` - Name of the template database.
* `deletion_protection` - Inhibits deletion of the database.

The `extension` block supports:

* `name` - Name of the database extension. For more information on available extensions see [the official documentation](https://yandex.cloud/docs/managed-postgresql/operations/cluster-extensions).
* `version` - Version of the extension.
