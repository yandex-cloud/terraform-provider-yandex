---
subcategory: "Managed Service for PostgreSQL"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a PostgreSQL cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_postgresql_cluster_v2/r_mdb_postgresql_cluster_v2_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/mdb_postgresql_cluster_v2/import.sh" }}
