---
subcategory: "Managed Service for PostgreSQL"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a PostgreSQL database within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a PostgreSQL database within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/).

## Example usage

{{ tffile "examples/mdb_postgresql_database/r_mdb_postgresql_database_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the database.
* `owner` - (Required) Name of the user assigned as the owner of the database. Forbidden to change in an existing database.
* `extension` - (Optional) Set of database extensions. The structure is documented below
* `lc_collate` - (Optional) POSIX locale for string sorting order. Forbidden to change in an existing database.
* `lc_type` - (Optional) POSIX locale for character classification. Forbidden to change in an existing database.
* `template_db` - (Optional) Name of the template database.

The `extension` block supports:

* `name` - (Required) Name of the database extension. For more information on available extensions see [the official documentation](https://yandex.cloud/docs/managed-postgresql/operations/cluster-extensions).
* `version` - (Optional) Version of the extension.
* `deletion_protection` - (Optional) Inhibits deletion of the database. Can either be `true`, `false` or `unspecified`.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_postgresql_database/import.sh" }}
