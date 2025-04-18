---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Encrypts given plaintext with the specified Yandex KMS key and provides access to the ciphertext.
---

# {{.Name}} ({{.Type}})

Encrypts given plaintext with the specified Yandex KMS key and provides access to the ciphertext.

~> Using this resource will allow you to conceal secret data within your resource definitions, but it does not take care of protecting that data in the logging output, plan output, or state output. Please take care to secure your secret data outside of resource definitions.

For more information, see [the official documentation](https://yandex.cloud/docs/kms/concepts/).

## Example usage

{{ tffile "examples/kms_secret_ciphertext/r_kms_secret_ciphertext_1.tf" }}

## Argument Reference

The following arguments are supported:

* `key_id` - (Required) ID of the symmetric KMS key to use for encryption.

* `aad_context` - (Optional) Additional authenticated data (AAD context), optional. If specified, this data will be required for decryption with the `SymmetricDecryptRequest`

* `plaintext` - (Required) Plaintext to be encrypted.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - an identifier for the resource with format `{key_id}/{ciphertext}`

* `ciphertext` - Resulting ciphertext, encoded with "standard" base64 alphabet as defined in RFC 4648 section 4

## Timeouts

`yandex_kms_secret_ciphertext` provides the following configuration options for [timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default 1 minute
- `delete` - Default 1 minute

## Import

~> Import for this resource is not implemented yet.

