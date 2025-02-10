---
subcategory: "Serverless Integrations"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Serverless Event Router Connector.
---

# {{.Name}} ({{.Type}})

Allows management of a Yandex Cloud Serverless Event Router Connector.

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/serverless_eventrouter_connector/r_serverless_eventrouter_connector_1.tf" }}

{{ .SchemaMarkdown | trimspace }}


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/serverless_eventrouter_connector/import.sh" }}
