---
subcategory: "Serverless Cloud Functions"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Function.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/function/r_function_1.tf" }}

{{ tffile "examples/function/r_function_2.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/function/import.sh" }}
