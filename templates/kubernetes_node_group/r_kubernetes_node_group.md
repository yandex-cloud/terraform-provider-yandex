---
subcategory: "Managed Service for Kubernetes (MK8S)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of Yandex Kubernetes Node Group.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/kubernetes_node_group/r_kubernetes_node_group_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/kubernetes_node_group/import.sh" }}

