---
subcategory: "Cloud Billing"
page_title: "Yandex: {{.Name}}"
description: |-
  Bind cloud to billing account.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/billing_cloud_binding/r_billing_cloud_binding_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash" "examples/billing_cloud_binding/import.sh" }}
