---
subcategory: "Serverless Integrations"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Serverless Event Router Rule.
---

# {{.Name}} ({{.Type}})

Allows management of a Yandex Cloud Serverless Event Router Rule.

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/serverless_eventrouter_rule/r_serverless_eventrouter_rule_1.tf" }}


{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/serverless_eventrouter_rule/import.sh" }}
