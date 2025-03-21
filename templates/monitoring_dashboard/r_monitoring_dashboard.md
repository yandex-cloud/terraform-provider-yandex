---
subcategory: "Monitoring"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Monitoring Dashboard.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/monitoring_dashboard/r_monitoring_dashboard_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/monitoring_dashboard/import.sh" }}
