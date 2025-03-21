---
subcategory: "Certificate Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  A TLS certificate signed by a certification authority confirming that it belongs to the owner of the domain name.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/cm_certificate/r_cm_certificate_1.tf" }}

{{ tffile "examples/cm_certificate/r_cm_certificate_2.tf" }}

{{ tffile "examples/cm_certificate/r_cm_certificate_3.tf" }}

{{ tffile "examples/cm_certificate/r_cm_certificate_4.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/cm_certificate/import.sh" }}
