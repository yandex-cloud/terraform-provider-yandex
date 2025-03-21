---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a new snapshot schedule.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/compute_snapshot_schedule/r_compute_snapshot_schedule_1.tf" }}

{{ tffile "examples/compute_snapshot_schedule/r_compute_snapshot_schedule_2.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/compute_snapshot_schedule/import.sh" }}
