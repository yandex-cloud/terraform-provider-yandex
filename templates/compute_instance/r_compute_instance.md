---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a VM instance resource.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/compute_instance/r_compute_instance_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/compute_instance/import.sh" }}
