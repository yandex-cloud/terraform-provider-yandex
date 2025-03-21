---
subcategory: "Cloud Logging"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages Yandex Cloud Logging group.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/logging_group/r_logging_group_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/logging_group/import.sh" }}
