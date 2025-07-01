---
subcategory: "Yandex Query"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages Object Storage binding.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/yq_object_storage_binding/r_yq_object_storage_binding_1.tf" }}

{{ tffile "examples/yq_object_storage_binding/r_yq_object_storage_binding_2.tf" }}

{{ tffile "examples/yq_object_storage_binding/r_yq_object_storage_binding_3.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud).

{{ codefile "shell" "examples/yq_object_storage_binding/import.sh" }}
