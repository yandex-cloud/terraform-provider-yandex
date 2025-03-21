---
subcategory: "Cloud Domain Name System (DNS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single IAM binding for a Cloud DNS Zone.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/dns_zone_iam_binding/r_dns_zone_iam_binding_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `dns_zone_id` and role.

{{ codefile "shell" "examples/dns_zone_iam_binding/import.sh" }}
