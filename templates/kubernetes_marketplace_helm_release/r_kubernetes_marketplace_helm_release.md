---
subcategory: "Kubernetes Marketplace"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of Kubernetes product installed from Yandex Cloud Marketplace.
---

# {{.Name}}

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/kubernetes_marketplace_helm_release/r_kubernetes_marketplace_helm_release_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash"  "examples/kubernetes_marketplace_helm_release/import.sh" }}
