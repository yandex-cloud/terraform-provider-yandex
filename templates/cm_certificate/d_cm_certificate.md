---
subcategory: "Certificate Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Certificate Manager Certificate.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/cm_certificate/d_cm_certificate_1.tf" }}

{{ tffile "examples/cm_certificate/d_cm_certificate_2.tf" }}

{{ .SchemaMarkdown | trimspace }}
