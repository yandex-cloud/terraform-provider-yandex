---
subcategory: "Managed Service for MySQL"
page_title: "Yandex: {{.Name}}"
description: |-
  Create a MySQL database at MySQL Cluster.
---

# {{.Name}} ({{.Type}})

Create a MySQL database at MySQL Cluster.

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_mysql_database/d_mdb_mysql_database_1.tf" }}

{{ .SchemaMarkdown | trimspace }}
