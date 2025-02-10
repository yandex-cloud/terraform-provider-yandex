---
subcategory: "Managed Kubernetes (MK8S) Marketplace"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of Kubernetes product installed from Yandex Cloud Marketplace.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/kubernetes_marketplace_helm_release/r_kubernetes_marketplace_helm_release_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash"  "examples/kubernetes_marketplace_helm_release/import.sh" }}
