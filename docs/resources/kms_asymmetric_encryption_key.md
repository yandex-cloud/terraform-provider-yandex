---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: yandex_kms_asymmetric_encryption_key"
description: |-
  Creates a Yandex KMS asymmetric encryption key that can be used for cryptographic operation.
---

# yandex_kms_asymmetric_encryption_key (Resource)

Creates a Yandex KMS asymmetric encryption key that can be used for cryptographic operation.

~> When Terraform destroys this key, any data previously encrypted with this key will be irrecoverable. For this reason, it is strongly recommended that you add lifecycle hooks to the resource to prevent accidental destruction.

For more information, see [the official documentation](https://yandex.cloud/docs/kms/concepts/).

## Example usage

```terraform
//
// Create a new KMS Assymetric Encryption Key.
//
resource "yandex_kms_asymmetric_encryption_key" "key-a" {
  name                 = "example-asymetric-encryption-key"
  description          = "description for key"
  encryption_algorithm = "RSA_2048_ENC_OAEP_SHA_256"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the key.

* `description` - (Optional) An optional description of the key.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the key.

* `encryption_algorithm` - (Optional) Encryption algorithm to be used with a new key. The default value is `RSA_2048_ENC_OAEP_SHA_256`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - The status of the key.
* `created_at` - Creation timestamp of the key.

## Timeouts

`yandex_kms_asymmetric_encryption_key` provides the following configuration options for [timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default 1 minute
- `update` - Default 1 minute
- `delete` - Default 1 minute

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_kms_asymmetric_encryption_key.<resource Name> <resource Id>
terraform import yandex_kms_asymmetric_encryption_key.key-a abj7u**********j38cd
```
