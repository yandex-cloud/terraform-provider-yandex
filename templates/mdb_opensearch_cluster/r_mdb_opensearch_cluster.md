---
subcategory: "Managed Service for OpenSearch"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a OpenSearch cluster within Yandex Cloud.
---

# {{.Name}}

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_opensearch_cluster/r_mdb_opensearch_cluster_1.tf" }}

Example of creating a high available OpenSearch Cluster.

{{ tffile "examples/mdb_opensearch_cluster/r_mdb_opensearch_cluster_2.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/mdb_mongodb_user/import.sh" }}
