---
subcategory: "Managed Service for MySQL"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a MySQL database within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a MySQL database within the Yandex Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mysql/).

## Example usage

{{ tffile "examples/mdb_mysql_database/r_mdb_mysql_database_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the database.

## Import

A MySQL database can be imported using the following format:

```
$ terraform import yandex_mdb_mysql_database.foo {cluster_id}:{database_name}
```
