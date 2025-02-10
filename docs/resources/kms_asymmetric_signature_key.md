---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: yandex_kms_asymmetric_signature_key"
description: |-
  Creates a Yandex KMS asymmetric signature key that can be used for cryptographic operation.
---

# yandex_kms_asymmetric_signature_key (Resource)

Creates a Yandex KMS asymmetric signature key that can be used for cryptographic operation.

## Example usage

```terraform
//
// Create a new KMS Assymetric Signature Key.
//
resource "yandex_kms_asymmetric_signature_key" "key-a" {
  name                = "example-asymetric-signature-key"
  description         = "description for key"
  signature_algorithm = "RSA_2048_SIGN_PSS_SHA_256"
}
```

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

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_kms_asymmetric_signature_key.<resource Name> <resource Id>
terraform import yandex_kms_asymmetric_signature_key.key-a abjjf**********p3gp8
```
