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

Resource can be imported using the following syntax:

{{ codefile "shell" "examples/iam_workload_identity_oidc_federation/import.sh" }}
