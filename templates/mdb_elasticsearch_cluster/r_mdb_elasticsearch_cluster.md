---
subcategory: "Managed Service for Elasticsearch"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Elasticsearch cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/mdb_elasticsearch_cluster/r_mdb_elasticsearch_cluster_1.tf" }}

{{ tffile "examples/mdb_elasticsearch_cluster/r_mdb_elasticsearch_cluster_2.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/mdb_elasticsearch_cluster/import.sh" }}
