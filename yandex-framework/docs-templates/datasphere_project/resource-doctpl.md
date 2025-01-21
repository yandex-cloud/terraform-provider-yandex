---
subcategory: "Datasphere"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Datasphere Project.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "yandex-framework/docs-templates/datasphere_project/resource-example-1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "yandex-framework/docs-templates/datasphere_project/resource-import.sh" }}
