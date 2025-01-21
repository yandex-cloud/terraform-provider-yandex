---
subcategory: "Beta Resources"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a PostgreSQL cluster within Yandex Cloud.
---

# {{.Name}}

{{ .Description | trimspace }}

## Example Usage

{{ tffile "yandex-framework/docs-templates/mdb_postgresql_cluster_beta/resource-example-1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "yandex-framework/docs-templates/mdb_postgresql_cluster_beta/resource-import.sh" }}

