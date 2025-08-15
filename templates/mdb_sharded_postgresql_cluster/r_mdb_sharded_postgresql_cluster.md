---
subcategory: "Managed Service for Sharded PostgreSQL"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Sharded PostgreSQL cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_sharded_postgresql_cluster/r_mdb_sharded_postgresql_cluster_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/mdb_sharded_postgresql_cluster/import.sh" }}
