---
subcategory: "Datasphere"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Datasphere Community.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/datasphere_community/r_datasphere_community_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/datasphere_community/import.sh" }}
