---
subcategory: "Managed Service for OpenSearch"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a OpenSearch cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_opensearch_cluster/r_mdb_opensearch_cluster_1.tf" }}

Example of creating a high available OpenSearch Cluster.

{{ tffile "examples/mdb_opensearch_cluster/r_mdb_opensearch_cluster_2.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/mdb_opensearch_cluster/import.sh" }}
