---
subcategory: "Container Registry"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a new container repository lifecycle policy.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/container_repository_lifecycle_policy/r_container_repository_lifecycle_policy_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/container_repository_lifecycle_policy/import.sh" }}
