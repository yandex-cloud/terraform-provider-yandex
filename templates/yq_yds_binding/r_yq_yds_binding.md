---
subcategory: "Yandex Query"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages Yandex DataStream binding.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/yq_yds_binding/r_yq_yds_binding_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud).

{{ codefile "shell" "examples/yq_yds_binding/import.sh" }}
