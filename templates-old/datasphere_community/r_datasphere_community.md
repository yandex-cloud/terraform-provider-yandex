---
subcategory: "Datasphere"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Datasphere Community.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/datasphere_community/r_datasphere_community_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/datasphere_community/import.sh" }}
