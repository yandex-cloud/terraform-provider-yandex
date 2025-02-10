---
subcategory: "Managed Service for MySQL"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a MySQL database within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a MySQL database within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mysql/).

## Example usage

{{ tffile "examples/mdb_mysql_database/r_mdb_mysql_database_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the database.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_mysql_database/import.sh" }}
