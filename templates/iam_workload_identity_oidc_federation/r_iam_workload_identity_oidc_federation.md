---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud IAM workload identity OIDC federations.
---

# {{.Name}} ({{.Type}})

{{ .Description }}

## Example Usage

{{ tffile "examples/iam_workload_identity_oidc_federation/r_iam_workload_identity_oidc_federation_1.tf" }}

{{ .SchemaMarkdown }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/iam_workload_identity_oidc_federation/import.sh" }}
