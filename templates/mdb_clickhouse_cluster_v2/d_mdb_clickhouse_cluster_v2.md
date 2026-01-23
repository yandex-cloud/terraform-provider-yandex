---
subcategory: "Managed Service for ClickHouse"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Managed ClickHouse cluster.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Managed ClickHouse cluster. For more information,
see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).

## Example usage

{{ tffile "examples/mdb_clickhouse_cluster_v2/d_mdb_clickhouse_cluster_v2_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Argument Reference

One of the following arguments are required:

* `cluster_id` - The ID of the ClickHouse cluster.
* `name` - The name of the ClickHouse cluster.
