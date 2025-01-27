---
subcategory: "Beta Resources"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a PostgreSQL cluster within Yandex Cloud.
---

# {{.Name}}

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_postgresql_cluster_beta/r_mdb_postgresql_cluster_beta_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/mdb_postgresql_cluster_beta/import.sh" }}
