---
subcategory: "Cloud Billing"
page_title: "Yandex: {{.Name}}"
description: |-
  Bind cloud to billing account.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "yandex-framework/docs-templates/billing_cloud_binding/resource-example-1.tf"}}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "yandex-framework/docs-templates/billing_cloud_binding/resource-import.sh" }}
