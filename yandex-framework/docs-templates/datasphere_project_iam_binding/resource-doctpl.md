---
subcategory: "Datasphere"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single IAM binding for a Yandex Datasphere Project.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "yandex-framework/docs-templates/datasphere_project_iam_binding/resource-example-1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "yandex-framework/docs-templates/datasphere_project_iam_binding/resource-import.sh" }}
