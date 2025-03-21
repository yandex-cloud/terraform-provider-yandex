---
subcategory: "Managed Service for Kubernetes (MK8S)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of Yandex Kubernetes Cluster.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/kubernetes_cluster/r_kubernetes_cluster_1.tf" }}

{{ tffile "examples/kubernetes_cluster/r_kubernetes_cluster_2.tf" }}

{{ tffile "examples/kubernetes_cluster/r_kubernetes_cluster_3.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/kubernetes_cluster/import.sh" }}
