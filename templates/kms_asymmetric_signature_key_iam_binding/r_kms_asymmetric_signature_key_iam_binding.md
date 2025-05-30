---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single IAM binding for a Key Management Service.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/kms_asymmetric_signature_key_iam_binding/r_kms_asymmetric_signature_key_iam_binding_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

KMS Asymmetric Signature Key IAM binding resource can be imported using the `asymmetric_signature_key_id` and resource role.

{{ codefile "shell" "examples/kms_asymmetric_signature_key_iam_binding/import.sh" }}
