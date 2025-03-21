---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  File storage is a virtual file system that can be attached to multiple Compute Cloud VMs in the same availability zone.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/compute_filesystem/r_compute_filesystem_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/compute_filesystem/import.sh" }}
