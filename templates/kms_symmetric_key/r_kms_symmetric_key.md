---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a Yandex KMS symmetric key that can be used for cryptographic operation.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/kms_symmetric_key/r_kms_symmetric_key_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/kms_symmetric_key/import.sh" }}
