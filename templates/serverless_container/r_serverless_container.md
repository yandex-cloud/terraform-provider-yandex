---
subcategory: "Serverless Containers"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Serverless Container.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/serverless_container/r_serverless_container_1.tf" }}

{{ tffile "examples/serverless_container/r_serverless_container_2.tf" }}

{{ tffile "examples/serverless_container/r_serverless_container_3.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/serverless_container/import.sh" }}
