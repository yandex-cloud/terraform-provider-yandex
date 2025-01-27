---
subcategory: "Datasphere"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Datasphere Project.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/datasphere_project/r_datasphere_project_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/datasphere_project/import.sh" }}
