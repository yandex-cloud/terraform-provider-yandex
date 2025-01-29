---
subcategory: "Managed Service for MongoDB"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a MongoDB Database within Yandex Cloud.
---

# {{.Name}}

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_mongodb_database/r_mdb_mongodb_database_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/mdb_mongodb_database/import.sh" }}
