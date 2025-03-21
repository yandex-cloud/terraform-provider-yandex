---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Encrypts given plaintext with the specified Yandex KMS key and provides access to the ciphertext.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/kms_secret_ciphertext/r_kms_secret_ciphertext_1.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

~> Import for this resource is not implemented yet.