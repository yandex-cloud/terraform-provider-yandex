---
subcategory: "Kubernetes Marketplace"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of Kubernetes product installed from Yandex Cloud Marketplace.
---

# {{.Name}}

{{ .Description | trimspace }}

## Example Usage

{{ tffile "yandex-framework/docs-templates/kubernetes_marketplace_helm_release/resource-example-1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "bash"  "yandex-framework/docs-templates/kubernetes_marketplace_helm_release/resource-import.sh" }}
