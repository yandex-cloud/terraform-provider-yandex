---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Disk Placement Group resource.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/compute_disk_placement_group/r_compute_disk_placement_group_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/compute_disk_placement_group/import.sh" }}
