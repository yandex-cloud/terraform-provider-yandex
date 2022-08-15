---
layout: "yandex"
page_title: "Yandex: yandex_mdb_postgresql_database"
sidebar_current: "docs-yandex-datasource-mdb-postgresql-database"
description: |-
  Get information about a Yandex Managed PostgreSQL database.
---

# yandex\_mdb\_postgresql\_database

Get information about a Yandex Managed PostgreSQL database. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-postgresql/).

## Example Usage

```hcl
data "yandex_mdb_postgresql_database" "foo" {
  cluster_id = "some_cluster_id"
  name = "test"
}

output "owner" {
  value = "${data.yandex_mdb_postgresql_database.foo.owner}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the PostgreSQL cluster.

* `name` - (Required) The name of the PostgreSQL cluster.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `owner` - Name of the user assigned as the owner of the database.
* `lc_collate` - POSIX locale for string sorting order. Forbidden to change in an existing database.
* `lc_type` - POSIX locale for character classification. Forbidden to change in an existing database.
* `extension` - Set of database extensions. The structure is documented below
* `template_db` - Name of the template database.

The `extension` block supports:

* `name` - Name of the database extension. For more information on available extensions see [the official documentation](https://cloud.yandex.com/docs/managed-postgresql/operations/cluster-extensions).
* `version` - Version of the extension.

