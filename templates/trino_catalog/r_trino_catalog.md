---
subcategory: "Managed Service for Trino"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages Trino catalog within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ codefile "terraform" "examples/trino_catalog/r_trino_catalog.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/trino_catalog/import.sh" }}

