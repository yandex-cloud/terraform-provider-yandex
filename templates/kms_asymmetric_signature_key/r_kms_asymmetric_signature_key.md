---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a Yandex KMS asymmetric signature key that can be used for cryptographic operation.
---

# {{.Name}} ({{.Type}})

Creates a Yandex KMS asymmetric signature key that can be used for cryptographic operation.

## Example usage

{{ tffile "examples/kms_asymmetric_signature_key/r_kms_asymmetric_signature_key_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the key.

* `description` - (Optional) An optional description of the key.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the key.

* `signature_algorithm` - (Optional) Signature algorithm to be used with a new key. The default value is `RSA_2048_SIGN_PSS_SHA_256`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - The status of the key.
* `created_at` - Creation timestamp of the key.

## Timeouts

`yandex_kms_asymmetric_signature_key` provides the following configuration options for [timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default 1 minute
- `update` - Default 1 minute
- `delete` - Default 1 minute

## Import

A KMS asymmetric signature key can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_kms_asymmetric_signature_key.top-secret kms_asymmetric_signature_key_id
```
