---
subcategory: "Serverless Integrations"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Serverless Event Router Bus.
---

# {{.Name}} ({{.Type}})

Allows management of a Yandex Cloud Serverless Event Router Bus.

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/serverless_eventrouter_bus/r_serverless_eventrouter_bus_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/serverless_eventrouter_bus/import.sh" }}
