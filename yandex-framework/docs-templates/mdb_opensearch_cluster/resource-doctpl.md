---
subcategory: "Managed Service for OpenSearch"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a OpenSearch cluster within Yandex Cloud.
---

# {{.Name}}

{{ .Description | trimspace }}

## Example Usage

{{ tffile "yandex-framework/docs-templates/mdb_opensearch_cluster/resource-example-1.tf" }}

Example of creating a high available OpenSearch Cluster.

{{ tffile "yandex-framework/docs-templates/mdb_opensearch_cluster/resource-example-2.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "yandex-framework/docs-templates/mdb_mongodb_user/resource-import.sh" }}
