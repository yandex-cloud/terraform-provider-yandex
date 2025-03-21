---
subcategory: "Managed Service for PostgreSQL"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a PostgreSQL cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/mdb_postgresql_cluster/r_mdb_postgresql_cluster_1.tf" }}

{{ tffile "examples/mdb_postgresql_cluster/r_mdb_postgresql_cluster_2.tf" }}

{{ tffile "examples/mdb_postgresql_cluster/r_mdb_postgresql_cluster_3.tf" }}

{{ tffile "examples/mdb_postgresql_cluster/r_mdb_postgresql_cluster_4.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_postgresql_cluster/import.sh" }}
