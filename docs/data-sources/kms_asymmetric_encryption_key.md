---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: yandex_kms_asymmetric_encryption_key"
description: |-
  Get data from Yandex KMS asymmetric encryption key.
---

# yandex_kms_asymmetric_encryption_key (Data Source)

Get data from Yandex KMS asymmetric encryption key.

## Example Usage

```terraform
//
// TBD
//
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `asymmetric_encryption_key_id` (String) Asymmetric encryption key ID.

### Optional

- `deletion_protection` (Boolean) The `true` value means that resource is protected from accidental deletion.
- `description` (String) The resource description.
- `encryption_algorithm` (String) Encryption algorithm to be used with a new key. The default value is `RSA_2048_ENC_OAEP_SHA_256`.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `name` (String) The resource name.

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `id` (String) The ID of this resource.
- `status` (String) The status of the key.
